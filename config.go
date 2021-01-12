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
	nSysEvents = 64 // Number of system events
	nChan      = 10 // Number of interrupt channels
	nHostInts  = 10 // Number of host interrupts
	nSignals   = nHostInts - 2 // Number of host interrupts routed to CPU
)

// Config contains the configuration mappings for the PRU.
// A configuration is initialised through config methods on this structure e.g:
//   ic := NewConfig()
//   ic.Channel2Interrupt(2, 2).Event2Channel(16, 2)
//   p := pru.Open(ic)
type Config struct {
	ev2chan    map[byte]byte
	chan2hint  map[byte]byte
}

// The default interrupt config.
// The default configuration is to map all the channels to the corresponding
// host interrupts as 1:1, and map the first 10 of the PRU R31 interrupts
// to the corresponding channels. The first 2 of the PRU R31 are PRU-PRU
// interrupts, and the remaining 8 map to the 8 event devices.
//
// Before the PRU is opened, this may be modified
// to overwrite the default configuration e.g
// DefaultIntConfig.Clear().Event2Channel(16, 4).Channel2Interrupt(4, 4)
var DefaultIntConfig *IntConfig

func init() {
	// Init the default interrupt controller config.
	// The default is to map the same channels to host interrupts, and
	// map the first 10 of the PRU R31 interrupts to the channels.
	DefaultIntConfig = NewIntConfig()
	var i uint
	for i = 0; i < nChan; i++ {
		DefaultIntConfig.Channel2Interrupt(i, i)
		DefaultIntConfig.Event2Channel(i+16, i)
	}
}

// NewIntConfig creates an empty IntConfig.
func NewIntConfig() *IntConfig {
	ic := new(IntConfig)
	ic.Clear()
	return ic
}

// Clear resets the configuration
func (ic *IntConfig) Clear() *IntConfig {
	ic.ev2chan = make(map[byte]byte)
	ic.chan2hint = make(map[byte]byte)
	return ic
}

// Event2Channel maps the system event to one of the 10
// interrupt channels. Multiple system events may be mapped to a single channel,
// but the same system events should not be mapped to multiple channels.
// Adding the mapping will enable the system event.
func (ic *IntConfig) Event2Channel(s, c uint) *IntConfig {
	ic.ev2chan[byte(s % nSysEvents)] = byte(c % nChan)
	return ic
}

// Channel2Interrupt maps the channel to a host interrupt.
// It is recommended to map the channels to interrupts as 1:1 i.e
// channel 1 mapped to host interrupt 1 etc.
// Multiple channels should not be mapped to a single host interrupt.
// A channel to host interrupt mapping must be present for the host interrupt to
// be enabled.
func (ic *IntConfig) Channel2Interrupt(c, h uint) *IntConfig {
	ic.chan2hint[byte(c % nChan)] = byte(h % nHostInts)
	return ic
}
