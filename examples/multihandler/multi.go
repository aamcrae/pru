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

//go:generate pasm -b pruhand.p

package main

import (
	"log"
	"time"

	"github.com/aamcrae/pru"
)

const hostInt = 18

func main() {
	// Use the default config which maps events 18 - 25 to
	// channels 2-9 and channels 2-9 to host interrupts 2-9.
	p, err := pru.Open(pru.DefaultConfig)
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
	err = u.LoadAndRunFile("pruhand.bin")
	if err != nil {
		log.Fatalf("%s", err)
	}
	time.Sleep(time.Second * 2)
	p.Close()
}
