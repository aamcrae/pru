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

const (
	nSysEvents = 64
	nHostInts = 10
	nChan = 10
)

// IntConfig contains the configuration mappings for the interrupt controller.
type IntConfig struct {
	sysevEnabled uint64
	sysev2chan map[byte]byte
	chan2hint map[byte]byte
}

var DefaultIntConfig *IntConfig

func init() {
	// Init the default interrupt controller config.
	// The default is to map the same channels to host interrupts, and
	// map the first 10 of the PRU R31 interrupts to the channels.
	ic := NewIntConfig()
	var i uint
	for i = 0; i < nChan; i++ {
		ic.Channel2Interrupt(i, i)
	}
	for i = 0; i < 10; i++ {
		ic.EnableSysEvent(i + 16)
		ic.SysEvent2Channel(i + 16, i)
	}
	DefaultIntConfig = ic
}

func NewIntConfig() (* IntConfig) {
	ic := new(IntConfig)
	ic.sysev2chan = make(map[byte]byte)
	ic.chan2hint = make(map[byte]byte)
	return ic
}

func (ic* IntConfig) EnableSysEvent(n uint) (*IntConfig) {
	ic.sysevEnabled |= 1 << uint(n % nSysEvents)
	return ic
}

func (ic* IntConfig) SysEvent2Channel(s, c uint) (*IntConfig) {
	ic.sysev2chan[byte(s % nSysEvents)] = byte(c % nChan)
	return ic
}

func (ic* IntConfig) Channel2Interrupt(c, h uint) (*IntConfig) {
	ic.chan2hint[byte(c % nChan)] = byte(h % nHostInts)
	return ic
}
