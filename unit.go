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
	"sync/atomic"
	"unsafe"
)

const nUnits = 2

type Unit struct {
	iram []uint32
	ctlReg *uint32

	// Exported fields
	Ram []byte
}

func newUnit(ram, iram, ctl uintptr) (* Unit) {
	u := new(Unit)
	u.ctlReg = (* uint32)(unsafe.Pointer(&pru.mem[ctl]))
	u.Ram =  pru.mem[ram:ram + am3xxRamSize]
	u.iram = (*(*[am3xxICount]uint32)(unsafe.Pointer(&pru.mem[iram])))[:]
	return u
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
