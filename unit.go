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

const nUnits = 2

const (
	// Control registers offset
	c_CONTROL = 0x00
	c_STATUS = 0x04
	c_WAKEUP_EN = 0x08
	c_CYCLE = 0x0C
	c_STALL = 0x10
	c_CTBIR0 = 0x20
	c_CTBIR1 = 0x24
	c_CTPPR0 = 0x28
	c_CTPPR1 = 0x2C
)
const (
	ctl_RESET = 0x0001
	ctl_ENABLE = 0x0002
	ctl_SLEEPING = 0x0004
	ctl_COUNTER_EN = 0x0008
	ctl_SINGLE_STEP = 0x0100
	ctl_RUNSTATE = 0x8000
)

// Unit represents one PRU (core) of the PRU-ICSS subsystem
type Unit struct {
	iram   uintptr
	ctlBase uintptr

	// Exported fields
	Ram []byte
	CycleCounter bool // Enable cycle counter
}

// newUnit initialises the unit's fields
func newUnit(ram, iram, ctl uintptr) *Unit {
	u := new(Unit)
	u.ctlBase = ctl
	u.Ram = pru.mem[ram : ram+am3xxRamSize]
	u.iram = iram
	return u
}

// Reset resets the PRU
func (u *Unit) Reset() {
	pru.wr(u.ctlBase + c_CONTROL, 0)
}

// Disable disables the PRU
func (u *Unit) Disable() {
	pru.wr(u.ctlBase + c_CONTROL, ctl_RESET)
}

// Enable enables the PRU
func (u *Unit) Enable() {
	u.EnableAt(0)
}

// EnableAt enables the PRU and sets the starting execution address.
// The address is specified as the instruction word, not the byte offset i.e a value
// of 10 will begin execution at the 10th instruction word (byte offset of 40 in the IRAM).
func (u *Unit) EnableAt(addr uint) {
	c := (uint32(addr)<<16)|ctl_ENABLE
	if u.CycleCounter {
		c |= ctl_COUNTER_EN
		// Clear cycle counter
		pru.wr(u.ctlBase + c_CYCLE, 0)
	}
	pru.wr(u.ctlBase + c_CONTROL, c)
}

// Counter returns the current cycle counter.
func (u *Unit) Counter() uint32 {
	return pru.rd(u.ctlBase + c_CYCLE)
}

// IsRunning returns true if the PRU is enabled and running.
func (u *Unit) IsRunning() bool {
	return (pru.rd(u.ctlBase + c_CONTROL) & ctl_RUNSTATE) != 0
}

// Load and execute the program from the file specified.
func (u *Unit) RunFile(s string) error {
	return u.RunFileAt(s, 0)
}

// Load and execute the program from the file specified and begin
// execution at the address specified.
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
	code := make([]uint32, fi.Size()/4)
	err = binary.Read(f, pru.Order, code)
	if err != nil {
		return err
	}
	return u.RunAt(code, addr)
}

// Run loads the PRU code into the IRAM and enables the PRU.
func (u *Unit) Run(code []uint32) error {
	return u.RunAt(code, 0)
}

// Run loads the PRU code into the IRAM and enables the PRU to
// begin execution at the address indicated.
func (u *Unit) RunAt(code []uint32, addr uint) error {
	if len(code) > am3xxICount {
		return fmt.Errorf("Program too large")
	}
	u.Disable()
	// Copy to IRAM.
	pru.copy(code, u.iram)
	u.EnableAt(addr)
	return nil
}
