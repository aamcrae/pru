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

const hostInt = 18

func main() {
	pc := pru.NewConfig()
	// Map all 8 system events to the same interrupt channel.
	for i := hostInt; i < hostInt+8; i++ {
		pc.Event2Channel(i, 0)
	}
	// Map channel 0 to host interrupt 2
	pc.Channel2Interrupt(0, 2)
	// Enable unit 0
	pc.EnableUnit(0)
	p, err := pru.Open(pc)
	if err != nil {
		log.Fatalf("%s", err)
	}
	if err != nil {
		log.Fatalf("%s", err)
	}
	u := p.Unit(0)
	for i := hostInt; i < hostInt+8; i++ {
		e := p.Event(i)
		if err != nil {
			log.Fatalf("%s", err)
		}
		e.SetHandler(func(id int) func() {
			return func() {
				log.Printf("Handler for event id %d", id)
			}
		}(i))
	}
	err = u.RunFile("pruhand.bin")
	if err != nil {
		log.Fatalf("%s", err)
	}
	time.Sleep(time.Second * 2)
	p.Close()
}
