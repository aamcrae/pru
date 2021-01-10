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
	if err != nil {
		log.Fatalf("%s", err)
	}
	u, err := p.Unit(0)
	if err != nil {
		log.Fatalf("%s", err)
	}
	for i := 0; i < 8; i++ {
		e, err := p.Event(i)
		if err != nil {
			log.Fatalf("%s", err)
		}
		e.SetHandler(func(id int) func(int) {
			return func(v int) {
				log.Printf("Handler for id %d, val = %d", id, v)
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
