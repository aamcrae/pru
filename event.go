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
	"fmt"
	"os"
)

const nEvents = 8

type Event struct {
	file *os.File

	// Exported fields
}

func event(p *PRU, id int) (* Event, error) {
	if id < 0 || id >= nEvents {
		return nil, fmt.Errorf("Invalid event number")
	}
	e := p.events[id]
	if e == nil {
		e = new(Event)
		f, err := os.OpenFile(fmt.Sprintf(drvUioBase, id), os.O_RDWR|os.O_SYNC, 0660)
		if err != nil {
			return nil, err
		}
		e.file = f
		p.events[id] = e
	}
	return e, nil
}
