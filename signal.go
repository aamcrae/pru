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
	"time"
	"unsafe"
)

const nSignals = 8

// Signal represents a host interrupt device.
type Signal struct {
	id                int
	file              *os.File
	handlerRegistered bool
	sigChan           chan int
	hStop             chan chan int
}

// newSignal initialises the Signal structure for the interrupt device unit id.
func newSignal(p *PRU, id int) (*Signal, error) {
	if id < 0 || id >= nSignals {
		return nil, fmt.Errorf("Invalid host interrupt number")
	}
	s := p.signals[id]
	if s == nil {
		s = new(Signal)
		f, err := os.OpenFile(fmt.Sprintf(drvUioBase, id), os.O_RDWR|os.O_SYNC, 0660)
		if err != nil {
			return nil, err
		}
		s.sigChan = make(chan int, 50)
		s.file = f
		s.id = id
		p.signals[id] = s
		go s.signalReader()
	}
	return s, nil
}

// Close frees the resources associated with this signal.
func (s *Signal) Close() {
	s.ClearHandler()
	close(s.sigChan)
	s.file.Close()
	pru.signals[s.id] = nil
}

// SetHandler installs an asynch handler that is invoked when signals are
// read from the host interrupt device. The argument to the handler is the
// running count of how many signals have been sent on this device.
func (s *Signal) SetHandler(f func(int)) {
	if s.handlerRegistered {
		s.ClearHandler()
	}
	s.handlerRegistered = true
	s.hStop = make(chan chan int)
	go s.hDispatcher(f)
}

// ClearHandler removes any currently installed handler for this interrupt device.
func (s *Signal) ClearHandler() {
	if s.handlerRegistered {
		// Create a channel to be used to signal when the handler has exited.
		c := make(chan int)
		s.hStop <- c
		// Once the handler receives the stop channel, a value is signalled back
		// to indicate that the handler has exited.
		<-c
		close(s.hStop)
		s.handlerRegistered = false
	}
}

// Wait reads the interrupt device and returns the value once available.
// This cannot be used if a handler has been installed on this device.
func (s *Signal) Wait() (int, error) {
	if s.handlerRegistered {
		return 0, fmt.Errorf("Handler registered, cannot use Wait")
	}
	return <-s.sigChan, nil
}

// WaitTimeout reads the interrupt device channel, returning if the timeout expires.
// This cannot be used if a handler has been installed on this device e.g
//  v, ok, err := s.WaitTimeout(time.Second)
//  if ok {
//      // Signal received
//  else {
//      // Timed out
//  }
func (s *Signal) WaitTimeout(tout time.Duration) (int, bool, error) {
	if s.handlerRegistered {
		return 0, false, fmt.Errorf("Handler registered, cannot use WaitTimeout")
	}
	ticker := time.NewTicker(tout)
	defer ticker.Stop()
	select {
	case v := <-s.sigChan:
		return v, true, nil
	case <-ticker.C:
		return -1, false, nil
	}
}

// hDispatcher is a shim between the device and the
// external handler that will be invoked when a signal is received.
// A stop channel is used to indicate when the handler should terminate.
func (s *Signal) hDispatcher(f func(int)) {
	for {
		select {
		case c := <-s.hStop:
			// Send a value back to signal that the handler has terminated.
			c <- 0
			return
		case v := <-s.sigChan:
			f(v)
		}
	}
}

// signalReader polls the device and sends the 32 bit value
// read from the device to the channel.
func (s *Signal) signalReader() {
	b := make([]byte, 4)
	for {
		n, err := s.file.Read(b)
		if err != nil {
			return
		}
		if n == 4 {
			val := *(*int32)(unsafe.Pointer(&b[0]))
			select {
			case s.sigChan <- int(val):
			default:
				// Unable to send, maybe the channel is closed.
			}
		}
	}
}
