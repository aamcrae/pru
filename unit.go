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
)

const (
	// Control registers offset
	c_CONTROL   = 0x00
	c_STATUS    = 0x04
	c_WAKEUP_EN = 0x08
	c_CYCLE     = 0x0C
	c_STALL     = 0x10
	c_CTBIR0    = 0x20
	c_CTBIR1    = 0x24
	c_CTPPR0    = 0x28
	c_CTPPR1    = 0x2C
)
const (
	ctl_RESET       = 0x0001
	ctl_ENABLE      = 0x0002
	ctl_SLEEPING    = 0x0004
	ctl_COUNTER_EN  = 0x0008
	ctl_SINGLE_STEP = 0x0100
	ctl_RUNSTATE    = 0x8000
)

// Unit represents one PRU (core) of the PRU-ICSS subsystem
type Unit struct {
	pru     *PRU
	iram    uintptr
	ctlBase uintptr

	Ram          ram  // PRU unit data ram
}

// newUnit initialises the unit's fields
func newUnit(p *PRU, ram, iram, ctl uintptr) *Unit {
	u := new(Unit)
	u.pru = p
	u.ctlBase = ctl
	u.Ram = p.mem[ram : ram+am3xxRamSize]
	u.iram = iram
	u.Reset()
	return u
}

// Reset resets the PRU unit
func (u *Unit) Reset() {
	u.pru.wr(u.ctlBase+c_CONTROL, 0)
}

// Disable disables this PRU unit
func (u *Unit) Disable() {
	u.pru.wr(u.ctlBase+c_CONTROL, ctl_RESET)
}

// IsRunning returns true if the PRU is enabled and running.
func (u *Unit) IsRunning() bool {
	return (u.pru.rd(u.ctlBase+c_CONTROL) & ctl_RUNSTATE) != 0
}

// Run enables the PRU core to run at address 0.
func (u *Unit) Run() error {
	return u.RunAt(0)
}

// RunAt enables the PRU core to begin execution at the specified byte address (which
// must be 32 bit aligned so that it points to the start of an instruction).
func (u *Unit) RunAt(addr uint) error {
	if (addr % 4) != 0 {
		return fmt.Errorf("start address is not 32 bit aligned")
	}
	if addr >= am3xxIRamSize {
		return fmt.Errorf("start address out of range")
	}
	u.Disable()
	// Upper 16 bits is instruction word address.
	u.pru.wr(u.ctlBase+c_CONTROL, (uint32(addr) << (16-2)) | ctl_ENABLE)
	return nil
}
// Load the program from a file to instruction address 0.
func (u *Unit) LoadFile(s string) error {
	return u.LoadFileAt(s, 0)
}

// Load and execute the program from the file specified.
func (u *Unit) LoadAndRunFile(s string) error {
	return u.LoadAndRunFileAt(s, 0)
}

// Load and execute the program 
func (u *Unit) LoadAndRun(code []uint32) error {
	return u.LoadAndRunAt(code, 0)
}

// Load and execute the program at the address specified.
func (u *Unit) LoadAndRunAt(code []uint32, addr uint) error {
	err := u.LoadAt(code, addr)
	if err != nil {
		return err
	}
	return u.RunAt(addr)
}

// Load and execute the program from a file at the address specified.
func (u *Unit) LoadAndRunFileAt(s string, addr uint) error {
	err := u.LoadFileAt(s, addr)
	if err != nil {
		return err
	}
	return u.RunAt(addr)
}

// Load the program from a file to the address specified.
func (u *Unit) LoadFileAt(s string, addr uint) error {
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
	code := make([]uint32, fi.Size()/4)
	err = binary.Read(f, u.pru.Order, code)
	if err != nil {
		return err
	}
	return u.LoadAt(code, addr)
}

// LoadAt loads the PRU code into the IRAM at the specified byte address.
func (u *Unit) LoadAt(code []uint32, addr uint) error {
	if uint(len(code) * 4) + addr > am3xxIRamSize {
		return fmt.Errorf("Program too large")
	}
	// Ensure unit is not running before writing IRAM.
	u.Disable()
	// Copy to IRAM.
	u.pru.write(code, u.iram + uintptr(addr))
	return nil
}
