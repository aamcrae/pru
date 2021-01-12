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
	rREVID   = 0x20000
	rCR      = 0x20004
	rGER     = 0x20010
	rGNLR    = 0x2001C
	rSISR    = 0x20020
	rSICR    = 0x20024
	rEISR    = 0x20028
	rEICR    = 0x2002C
	rHIEISR  = 0x20034
	rHIDISR  = 0x20038
	rGPIR    = 0x20080
	rSRSR0   = 0x20200
	rSRSR1   = 0x20204
	rSECR0   = 0x20280
	rSECR1   = 0x20284
	rESR0    = 0x20300
	rESR1    = 0x20304
	rCMRBase = 0x20400
	rHMRBase = 0x20800
	rSIPR0   = 0x20D00
	rSIPR1   = 0x20D04
	rSITR0   = 0x20D80
	rSITR1   = 0x20D84
)

type PRU struct {
	mmapFile *os.File
	memBase  int
	memSize  int
	mem      []byte
	version  int
	units    [nUnits]*Unit
	signals  []*os.File
	events   [nEvents]*Event
	sigMask  [nSignals]uint64 // System event mask for each signal
    evMask   uint64 // Global mask for system events

	// Exported fields
	SharedRam []byte
	Order     binary.ByteOrder
}

// Single instance of PRU.
var pru *PRU

// Open initialises the PRU subsystem using the configuration provided.
func Open(pc *Config) (*PRU, error) {
	if pru != nil {
		return nil, fmt.Errorf("Device already open; must close it first")
	}
	p = new(PRU)
	// Validate config.
	// All events must map to a host interrupt.
	var cmr [nSysEvents / 4]uint32
	for se, c := range pc.ev2chan {
		cmr[se/4] |= uint32(c) << ((se % 4) * 8)
		hi, ok := pc.chan2hint[c]
		if !ok {
			return fmt.Errorf("chan %d not mapped to host interrupt", c)
		}
		if hi >= 2 {
			p.sigMask[hostInt2Signal(hi)] |= 1 << se
		}
		p.evMask |= 1 << se
	}
	// Channels must map to only one host interrupt
	var hiMapped [nHostInts]bool
	var hmr [(nHostInts + 3) / 4]uint32
	for c, hi := range pc.chan2hint {
		if hiMapped[c] {
			return fmt.Errorf("host interrupt %d has multiple channels assigned", hi)
		}
		hmr[c/4] |= uint32(hi) << ((c % 4) * 8)
		hiMapped[c] = true
	}
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
	// Open signal devices for each enabled host interrupt (the first 2 are skipped).
	for i := 0; i < nSignals; i++ {
		if p.sigMask[i] != 0 {
			f, err := os.OpenFile(fmt.Sprintf(drvUioBase, i), os.O_RDWR|os.O_SYNC, 0660)
			if err != nil {
				p.Close()
				return nil, err
			}
			p.signals = append(p.signals, f)
			go p.signalReader(f)
		}
	}
	// For
	for i := 0; i < nEvents; i++ {
		if (p.evMask & (1 << i)) != 0 {
			p.events[i] := newEvent()
		}
	}
	// Start setting up hardware
	// Disable global interrupts
	p.wr(rGER, 0)
	p.wr64(rSIPR0, 0xFFFFFFFFFFFFFFFF)
	// Init the CMR (Channel Map Registers)
	p.copy(cmr[:], rCMRBase)
	// Init the HMR (Host Interrupt Map Registers)
	p.copy(hmr[:], rHMRBase)
	p.wr64(rSITR0, 0)
	// Enable the system events that are used.
	p.wr64(rESR0, ic.evMask)
	p.wr64(rSECR0, ic.evMask)
	for i, hset := range hiMapped {
		if hset {
			p.wr(rHIEISR, uint32(i))
		}
	}
	// Re-enable interrupts globally.
	p.wr(rGER, 1)
	pru = p
	return p, nil
}

// Unit returns a structure pointer representing a single PRU Core
func (p *PRU) Unit(u int) *Unit {
	return p.units[u]
}

// Signal returns the host interrupt device identified by id.
func (p *PRU) Signal(id int) (*Signal, error) {
	return newSignal(p, id)
}

// SendEvent triggers the system event
func (p *PRU) SendEvent(se uint) {
	p.wr64(rSRSR0, 1 << se)
}

// ClearEvent resets the system event, and re-enables the associated host interrupt.
func (p *PRU) ClearEvent(se uint) {
	p.wr64(rSECR0, 1 << se)
	// Re-enable the host interrupt.
	ch, ok := p.ic.ev2chan[byte(se)]
	if ok {
		hi, ok := p.ic.chan2hint[byte(ch)]
		if ok {
			p.wr(rHIEISR, uint32(hi))
		}
	}
}

// IntConfigure applies the interrupt controller configuration to the PRU
func (p *PRU) IntConfigure(ic *IntConfig) error {
	return nil
}

// Close deactivates the PRU subsystem, releasing all the resources associated with it.
func (p *PRU) Close() {
	pru = nil
	for _, s := range p.signals {
		s.Close()
	}
	for i, _ := range p.events
	unix.Munmap(p.mem)
	p.mmapFile.Close()
}

// signalReader polls the device and sends the 32 bit value
// read from the device to the channel.
func (p *PRU) signalReader(f *os.File) {
	b := make([]byte, 4)
	for {
		n, err := f.Read(b)
		if err != nil {
			return
		}
		if n == 4 {
			val := *(*int32)(unsafe.Pointer(&b[0]))
			select {
			case s.sigChan <- int(val):
			default:
				// Unable to send, maybe the channel is closed.
			}
		}
	}
}

// Description returns a human readable string describing the PRU
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

// rd reads one 32 bit word from the shared memory area
func (p *PRU) rd(offs uintptr) uint32 {
	return atomic.LoadUint32((*uint32)(unsafe.Pointer(&p.mem[offs])))
}

// wr writes one 32 bit word to the shared memory area
func (p *PRU) wr(offs uintptr, v uint32) {
	atomic.StoreUint32((*uint32)(unsafe.Pointer(&p.mem[offs])), v)
}

// rd64 reads 2 32 bits words from successive addresses and combines them to a 64 bit value
// The lower 32 bit of the 64 bit word is read from the first address
func (p *PRU) rd64(offs uintptr) uint64 {
	v := uint64(atomic.LoadUint32((*uint32)(unsafe.Pointer(&p.mem[offs]))))
	v |= uint64(atomic.LoadUint32((*uint32)(unsafe.Pointer(&p.mem[offs + 4])))) << 32
	return v
}

// rd64 writes a 64 bit value to 2 successive addresses
// The lower 32 bits of the 64 bit word is written to the first address
func (p *PRU) wr64(offs uintptr, v uint64) {
	atomic.StoreUint32((*uint32)(unsafe.Pointer(&p.mem[offs])), uint32(v))
	atomic.StoreUint32((*uint32)(unsafe.Pointer(&p.mem[offs + 4])), uint32(v >> 32))
}

// copy copies the 32 bit data to the shared memory area
func (p *PRU) copy(src []uint32, dst uintptr) {
	for _, c := range src {
		pru.wr(dst, c)
		dst += 4
	}
}

// hostInt2Signal - convert host interrupt index to signal index
func hostInt2Signal(hi int) int {
	return hi - 2
}

// readDriverValue opens and reads a string from a device file and decodes
// the string as an integer. This is used to retrieve device specific
// parameters from the PRU kernel device driver.
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
