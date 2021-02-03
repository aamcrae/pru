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

//go:generate pasm -b prugpio.p

package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"

	"github.com/aamcrae/pru"
)

const event = 16
const intr1 = 17
const intr2 = 18
const intrBit = 31

var in = flag.Int("in", 15, "Input bit for GPIO")    // P8_15 input
var out = flag.Int("out", 15, "Output bit for GPIO") // P8_11 output
var unit = flag.Int("pru", 0, "PRU unit to execute on")

func main() {
	flag.Parse()
	// Set up just one system event (pr1_pru_mst_intr[0]_intr_req) and map it to channel 2.
	// Map channel 2 to host interrupt 2 (which appears on event device 0)
	pc := pru.NewConfig()
	pc.EnableUnit(*unit)
	pc.Event2Channel(intr2, 0).Channel2Interrupt(0, 0)
	pc.Event2Channel(intr1, 1).Channel2Interrupt(1, 1)
	pc.Event2Channel(event, 2).Channel2Interrupt(2, 2)
	p, err := pru.Open(pc)
	if err != nil {
		log.Fatalf("%s", err)
	}
	defer p.Close()
	u := p.Unit(*unit)
	e := p.Event(event)
	r := u.Ram.Open()
	params := []interface{}{
		uint32(event - 16),
		uint32(intrBit),
		uint32(1),
		uint32(0xDEADBEEF),
		uint32(*in),
		uint32(*out),
	}
	for _, v := range params {
		binary.Write(r, p.Order, v)
	}
	log.Printf("Running PRU")
	err = u.LoadAndRunFile("prugpio.bin")
	if err != nil {
		log.Fatalf("%s", err)
	}
	fmt.Printf("Press Enter key to terminate\n")
	fmt.Scanln()
	p.SendEvent(intr1)
	e.Wait()
	p.ClearEvent(intr1)
}
