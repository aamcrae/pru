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

//go:generate pasm -m prucode.p prucode
//go:generate img2go.sh prucode main

package main

import (
	"log"
	"time"

	"github.com/aamcrae/pru"
)

const W1 = 0xDEADBEEF
const W2 = 0xCAFEF00D

func main() {
	p, err := pru.Open(pru.DefaultConfig)
	if err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("Description: %s", p.Description())
	u := p.Unit(0)
	log.Printf("RAM %p, size is %d, shared RAM size = %d", &u.Ram[0], len(u.Ram), len(p.SharedRam))
	p.Order.PutUint32(u.Ram[0:], W1)
	p.Order.PutUint32(u.Ram[4:], W2)
	err = u.LoadAt(prucode_img, 0x200)
	if err != nil {
		log.Fatalf("%s", err)
	}
	err = u.RunAt(0x200)
	if err != nil {
		log.Fatalf("%s", err)
	}
	for i := 1; i < 1000; i++ {
		time.Sleep(time.Millisecond)
		if !u.IsRunning() {
			log.Printf("Program has completed (time %d ms)", i)
			log.Printf("Wrote 0x%x and 0x%x", uint32(W1), uint32(W2))
			log.Printf("Read  0x%x and 0x%x", p.Order.Uint32(u.Ram[0:]), p.Order.Uint32(u.Ram[4:]))
			break
		}
	}
	p.Close()
}
