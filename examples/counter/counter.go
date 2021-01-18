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
	p, err := pru.Open(pru.DefaultConfig)
	if err != nil {
		log.Fatalf("%s", err)
	}
	u := p.Unit(0)
	u.CycleCounter = true
	err = u.LoadAndRunFile("prucounter.bin")
	if err != nil {
		log.Fatalf("%s", err)
	}
	time.Sleep(time.Second)
	log.Printf("Counter is %d (201 expected)", u.Counter())
	p.Close()
}
