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
	nEvents   = 64            // Number of system events
	nChannels = 10            // Number of interrupt channels
	nHostInts = 10            // Number of host interrupts
	nUnits    = 2             // Number of PRU cores
	nSignals  = nHostInts - 2 // Number of host interrupts routed to CPU
)

// Config contains the configuration mappings for the PRU.
// A configuration is initialised through config methods on this structure e.g:
//   ic := NewConfig()
//   ic.EnableUnit(1)
//   ic.Channel2Interrupt(2, 2).Event2Channel(16, 2)
//   p := pru.Open(ic)
type Config struct {
	umask     int
	ev2chan   map[byte]byte
	chan2hint map[byte]byte
}

// The default config.
// The default configuration is to enable both PRU cores, map all the channels
// to the corresponding host interrupts as 1:1, and map the first 10 of the
// PRU R31 interrupts to the corresponding channels.
// The first 2 of the PRU R31 are PRU-PRU interrupts, and the remaining 8 map
// to the 8 event devices.
//
// Before the PRU is opened, this may be modified
// to overwrite the default configuration e.g
// DefaultConfig.Clear().EnableUnit(0).Event2Channel(16, 4).Channel2Interrupt(4, 4)
var DefaultConfig *Config

func init() {
	// Init the default interrupt controller config.
	// The default is to map the same channels to host interrupts, and
	// map the first 10 of the PRU R31 interrupts to the channels.
	DefaultConfig = NewConfig()
	DefaultConfig.EnableUnit(0).EnableUnit(1)
	for i := 0; i < nChannels; i++ {
		DefaultConfig.Channel2Interrupt(i, i)
		DefaultConfig.Event2Channel(i+16, i)
	}
}

// NewConfig creates an empty Config.
func NewConfig() *Config {
	ic := new(Config)
	ic.Clear()
	return ic
}

// Clear resets the configuration
func (ic *Config) Clear() *Config {
	ic.umask = 0
	ic.ev2chan = make(map[byte]byte)
	ic.chan2hint = make(map[byte]byte)
	return ic
}

// EnableUnit enables the use of a PRU core unit in this process.
func (ic *Config) EnableUnit(u int) *Config {
	ic.umask |= 1 << uint(u % nUnits)
	return ic
}

// Event2Channel maps the system event to one of the 10
// interrupt channels. Multiple system events may be mapped to a single channel,
// but the same system events should not be mapped to multiple channels.
// Adding the mapping will enable the system event.
func (ic *Config) Event2Channel(s, c int) *Config {
	ic.ev2chan[byte(s % nEvents)] = byte(c % nChannels)
	return ic
}

// Channel2Interrupt maps the channel to a host interrupt.
// It is recommended to map the channels to interrupts as 1:1 i.e
// channel 1 mapped to host interrupt 1 etc.
// Multiple channels should not be mapped to a single host interrupt.
// A channel to host interrupt mapping must be present for the host interrupt to
// be enabled.
func (ic *Config) Channel2Interrupt(c, h int) *Config {
	ic.chan2hint[byte(c % nChannels)] = byte(h % nHostInts)
	return ic
}
