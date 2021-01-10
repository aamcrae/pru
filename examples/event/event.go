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
	"time"

	"github.com/aamcrae/pru"
)

func main() {
	p, err := pru.Open()
	if err != nil {
		log.Fatalf("%s", err)
	}
	// Set up just one system event (pr1_pru_mst_intr[0]_intr_req) and map it to channel 2.
	// Map channel 2 to host interrupt 2 (which appears on event device 0)
	ic := pru.NewIntConfig()
	ic.EnableSysEvent(16).Channel2Interrupt(2, 2).SysEvent2Channel(16, 2)
	p.IntConfigure(ic)
	if err != nil {
		log.Fatalf("%s", err)
	}
	u := p.Unit(0)
	e, err := p.Event(0)
	if err != nil {
		log.Fatalf("%s", err)
	}
	err = u.RunFile("pruevent.bin")
	if err != nil {
		log.Fatalf("%s", err)
	}
	v, ok, err := e.WaitTimeout(time.Second)
	if err != nil {
		log.Fatalf("%s", err)
	}
	if !ok {
		log.Printf("Event timed out!")
	} else {
		log.Printf("Event received: %d", v)
	}
	p.Close()
}
