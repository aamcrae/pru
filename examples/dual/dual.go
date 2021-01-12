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

package main

import (
	"log"

	"github.com/aamcrae/pru"
)

const sram = 0x10000

func main() {
	p, err := pru.Open(pru.DefaultConfig)
	if err != nil {
		log.Fatalf("%s", err)
	}
	defer p.Close()
	u0 := p.Unit(0)
	u1 := p.Unit(1)
	e0 := p.Event(18)
	e1 := p.Event(19)
	s0 := "Unit 0, test string, should appear at offs 0 in shared RAM"
	s1 := "Another string, copied from unit 1, at offset 0x100"
	runUnit(p, u0, 2, sram, s0)
	runUnit(p, u1, 3, sram+0x100, s1)
	e0.Wait()
	e1.Wait()
	log.Printf("string at     0: %s", p.SharedRam[0:len(s0)])
	log.Printf("string at 0x100: %s", p.SharedRam[0x100:len(s1)+0x100])
}

func runUnit(p *pru.PRU, u *pru.Unit, ev uint32, offs int, txt string) {
	slen := (len(txt) + 7) / 4 // Length in words
	p.Order.PutUint32(u.Ram[0:], ev)
	p.Order.PutUint32(u.Ram[4:], uint32(slen))
	p.Order.PutUint32(u.Ram[8:], 16)
	p.Order.PutUint32(u.Ram[12:], uint32(offs))
	copy(u.Ram[16:], txt)
	err := u.RunFile("prucode.bin")
	if err != nil {
		log.Fatalf("%s", err)
	}
}
