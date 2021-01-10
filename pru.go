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
	drvUioBase = "/dev/uio%d"
	drvUIO0    = "/dev/uio0"
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
	am3xxIntc      = 0x00020000
	am3xxPru0Ctl   = 0x00022000
	am3xxPru0Dbg   = 0x00022400
	am3xxPru1Ctl   = 0x00024000
	am3xxPru1Dbg   = 0x00024400
	am3xxPru0Iram  = 0x00034000
	am3xxPru1Iram  = 0x00038000
	// Memory sizes
	am3xxRamSize       = 8 * 1024
	am3xxSharedRamSize = 12 * 1024
	am3xxIRamSize      = 8 * 1024
	am3xxICount        = am3xxIRamSize / 4

	// Interrupt controller register offsets
	rREVID = 0x20000
	rCR = 0x20004
	rGER = 0x20010
	rGNLR = 0x2001C
	rSISR = 0x20020
	rSICR = 0x20024
	rEISR = 0x20028
	rEICR = 0x2002C
	rHIEISR = 0x20034
	rHIDISR = 0x20038
	rGPIR = 0x20080
)

type PRU struct {
	mmapFile *os.File
	memBase  int
	memSize  int
	mem      []byte
	version  int
	units    [nUnits]*Unit
	events   [nEvents]*Event

	// Exported fields
	SharedRam []byte
	Order     binary.ByteOrder
}

// IntConfig contains the configuration mappings for the interrupt controller.
type IntConfig struct {
}

var DefaultIntConfig IntConfig

var pru PRU

func init() {
  // Init the default interrupt controller config.
}

// Open initialises the PRU device
func Open() (*PRU, error) {
	p := &pru
	if p.mmapFile == nil {
		var err error
		p.memBase, err = readDriverValue(drvMemBase)
		if err != nil {
			return nil, err
		}
		p.memSize, err = readDriverValue(drvMemSize)
		if err != nil {
			return nil, err
		}
		f, err := os.OpenFile(drvUIO0, os.O_RDWR|os.O_SYNC, 0660)
		if err != nil {
			return nil, err
		}
		p.mem, err = unix.Mmap(int(f.Fd()), 0, p.memSize, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("%s: %v", drvUIO0, err)
		}
		// Determine PRU version (AM18xx or AM33xx)
		vers := p.rd(rREVID)
		switch vers {
		case 0x00a9824e:
			p.Order = binary.BigEndian
			p.version = am33xx
		case 0x4E82A900:
			p.Order = binary.LittleEndian
			p.version = am33xx
		default:
			f.Close()
			return nil, fmt.Errorf("Unknown PRU version: 0x%08x", vers)
		}
		p.SharedRam = p.mem[am3xxSharedRam : am3xxSharedRam+am3xxSharedRamSize]
		p.units[0] = newUnit(am3xxPru0Ram, am3xxPru0Iram, am3xxPru0Ctl)
		p.units[1] = newUnit(am3xxPru1Ram, am3xxPru1Iram, am3xxPru1Ctl)
		p.mmapFile = f
		// Assume that the default configuration will not return an error.
		p.IntConfigure(&DefaultIntConfig)
	}
	return p, nil
}

func (p *PRU) Unit(u int) (*Unit, error) {
	if u < 0 || u >= nUnits {
		return nil, fmt.Errorf("Invalid unit number")
	}
	return p.units[u], nil
}

func (p *PRU) Event(id int) (*Event, error) {
	return newEvent(p, id)
}

func (p *PRU) IntConfigure(ic *IntConfig) error {
	return nil
}

func (p *PRU) Close() {
	if p.mmapFile != nil {
		unix.Munmap(p.mem)
		p.mmapFile.Close()
		p.mmapFile = nil
		p.units[0] = nil
		p.units[1] = nil
		for i, e := range p.events {
			if e != nil {
				e.Close()
				p.events[i] = nil
			}
		}
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

func (p *PRU) rd(offs uintptr) uint32 {
	return atomic.LoadUint32((*uint32)(unsafe.Pointer(&p.mem[offs])))
}

func (p *PRU) wr(offs uintptr, v uint32) {
	atomic.StoreUint32((*uint32)(unsafe.Pointer(&p.mem[offs])), v)
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
