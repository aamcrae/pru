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

//go:generate pasm -b prucounter.p

package main

import (
	"log"
	"time"

	"github.com/aamcrae/pru"
)

func main() {
	p, err := pru.Open(pru.DefaultConfig)
	if err != nil {
		log.Fatalf("%s", err)
	}
	u := p.Unit(0)
	p.Order.PutUint32(u.Ram[0:], 0x00022000) // PRU 0 control registers
	p.Order.PutUint32(u.Ram[4:], 0) // Saved cycles value
	p.Order.PutUint32(u.Ram[8:], 0) // Saved stalls value
	err = u.LoadAndRunFile("prucounter.bin")
	if err != nil {
		log.Fatalf("%s", err)
	}
	time.Sleep(time.Second)
	log.Printf("Cycles are %d (202 expected), stalls = %d", p.Order.Uint32(u.Ram[4:]), p.Order.Uint32(u.Ram[8:]))
	p.Close()
}
