// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pru

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Device paths.
const (
	drvMemBase = "/sys/class/uio/uio0/maps/map0/addr"
	drvMemSize = "/sys/class/uio/uio0/maps/map0/size"
	drvUIO = "/dev/uio0"
)

// versions
const (
	am18xx = iota
	am33xx = iota
)

const (
    // AM3xx
	// Memory offsets
	am3xxPru0Ram   = 0x00000000
	am3xxPru1Ram   = 0x00002000
	am3xxSharedRam = 0x00010000
	am3xxIntc       = 0x00020000
	am3xxPru0Ctl   = 0x00022000
	am3xxPru0Dbg   = 0x00022400
	am3xxPru1Ctl   = 0x00024000
	am3xxPru1Dbg   = 0x00024400
	am3xxPru0Iram  = 0x00034000
	am3xxPru1Iram  = 0x00038000
	// Memory sizes
	am3xxRamSize = 8 * 1024
	am3xxSharedRamSize = 12 * 1024
	am3xxIRamSize = 8 * 1024
	am3xxICount = am3xxIRamSize / 4
)

const nUnits = 2

type Unit struct {
	iram []uint32
	ctlReg *uint32

	// Exported fields
	Ram []byte
}

type PRU struct {
	mmapFile *os.File
	memBase int
	memSize int
	mem []byte
	version int
	units [nUnits]*Unit

	// Exported fields
	SharedRam []byte
	Order binary.ByteOrder
}

var pru PRU
var units [nUnits]Unit

// Open initialises the PRU device
func Open() (*PRU, error) {
	if pru.mmapFile == nil {
		var err error
		pru.memBase, err = readDriverValue(drvMemBase)
		if err != nil {
			return nil, err
		}
		pru.memSize, err = readDriverValue(drvMemSize)
		if err != nil {
			return nil, err
		}
		f, err := os.OpenFile(drvUIO, os.O_RDWR|os.O_SYNC, 0660)
		if err != nil {
			return nil, err
		}
		pru.mem, err = unix.Mmap(int(f.Fd()), 0, pru.memSize, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("%s: %v", drvUIO, err)
		}
		// Determine PRU version (AM18xx or AM33xx)
		vers := atomic.LoadUint32((* uint32)(unsafe.Pointer(&pru.mem[am3xxIntc])))
		switch vers {
		case 0x00a9824e:
			pru.Order = binary.BigEndian
			pru.version = am33xx
		case 0x4E82A900:
			pru.Order = binary.LittleEndian
			pru.version = am33xx
		default:
			f.Close()
			return nil, fmt.Errorf("Unknown PRU version: 0x%08x", vers)
		}
		pru.SharedRam = pru.mem[am3xxSharedRam:am3xxSharedRam + am3xxSharedRamSize]
		pru.units[0] = newUnit(am3xxPru0Ram, am3xxPru0Iram, am3xxPru0Ctl)
		pru.units[1] = newUnit(am3xxPru1Ram, am3xxPru1Iram, am3xxPru1Ctl)
		pru.mmapFile = f
	}
	return &pru, nil
}

func newUnit(ram, iram, ctl uintptr) (* Unit) {
	u := new(Unit)
	u.ctlReg = (* uint32)(unsafe.Pointer(&pru.mem[ctl]))
	u.Ram =  pru.mem[ram:ram + am3xxRamSize]
	u.iram = (*(*[am3xxICount]uint32)(unsafe.Pointer(&pru.mem[iram])))[:]
	return u
}

func (p *PRU) Unit(u int) (* Unit, error) {
	if u < 0 || u >= nUnits {
		return nil, fmt.Errorf("Invalid unit number")
	}
	return p.units[u], nil
}

func (p *PRU) Close() {
	if p.mmapFile != nil {
		unix.Munmap(p.mem)
		p.mmapFile.Close()
		p.mmapFile = nil
	}
}

func (p *PRU) Description() string {
	var s strings.Builder
	fmt.Fprint(&s, "PRU")
	if p.version == am33xx {
		fmt.Fprint(&s, " AM33xx")
	} else {
		fmt.Fprint(&s, " AM18xx")
	}
	if p.Order == binary.LittleEndian {
		fmt.Fprint(&s, " Little endian")
	} else {
		fmt.Fprint(&s, " Big endian")
	}
	return s.String()
}

func readDriverValue(s string) (int, error) {
	var val int
	f, err := os.Open(s)
	if err != nil {
		return -1, err
	}
	defer f.Close()
	n, err := fmt.Fscanf(f, "%v", &val)
	if err != nil {
		return -1, fmt.Errorf("%s: %v", s, err)
	}
	if n != 1 {
		return -1, fmt.Errorf("%s: no value found", s)
	}
	return val, nil
}

func (u *Unit) Reset() {
	atomic.StoreUint32(u.ctlReg, 0)
}

func (u *Unit) Disable() {
	atomic.StoreUint32(u.ctlReg, 1)
}

func (u *Unit) Enable() {
	u.EnableAt(0)
}

func (u *Unit) EnableAt(addr uint) {
	atomic.StoreUint32(u.ctlReg, (uint32(addr) << 16) | 2)
}

func (u *Unit) IsRunning() bool {
	return (atomic.LoadUint32(u.ctlReg) & (1 << 15)) != 0
}

func (u *Unit) RunFile(s string) error {
	return u.RunFileAt(s, 0)
}

func (u *Unit) RunFileAt(s string, addr uint) error {
	f, err := os.Open(s)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if (fi.Size() % 4) != 0 {
		return fmt.Errorf("length is not 32 bit aligned")
	}
	code := make([]uint32, fi.Size() / 4)
	err = binary.Read(f, pru.Order, code)
	if err != nil {
		return err
	}
	return u.RunAt(code, addr)
}

func (u *Unit) Run(code []uint32) error {
	return u.RunAt(code, 0)
}

func (u *Unit) RunAt(code []uint32, addr uint) error {
	if len(code) > am3xxICount {
		return fmt.Errorf("Program too large")
	}
	u.copy(code, u.iram)
	u.EnableAt(addr)
	return nil
}

func (u *Unit) copy(data []uint32, dest []uint32) {
	u.Disable()
	// Copy to memory.
	for i, c := range data {
		atomic.StoreUint32(&dest[i], c)
	}
}
