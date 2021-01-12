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

// Event handles waiting on or receiving system events.
type Event struct {
	handlerRegistered bool
	evChan            chan bool
	stopChan          chan bool
}

// newEvent creates and initialises an Event structure.
func newEvent() (*Event) {
	ev := new(Event)
	ev.evChan = make(chan bool, 50)
	return ev
}

// SetHandler installs an asynch handler that is invoked when events are
// read from the host interrupt device.
func (e *Event) SetHandler(f func()) {
	if e.handlerRegistered {
		e.ClearHandler()
	}
	e.handlerRegistered = true
	e.stopChan = make(chan chan bool)
	go e.dispatcher(f)
}

// ClearHandler removes any currently installed handler for this event
func (e *Signal) ClearHandler() {
	if e.handlerRegistered {
		// Create a channel to be used to signal when the handler has exited.
		c := make(chan bool)
		e.stopChan <- c
		// Once the handler receives the stop channel, a value is signalled back
		// to indicate that the handler has exited.
		<-c
		close(e.stopChan)
		s.handlerRegistered = false
	}
}

// Wait reads the event channel and returns the value once available.
// This cannot be used if a handler has been installed on this event.
func (e *Event) Wait() error {
	if e.handlerRegistered {
		return 0, fmt.Errorf("Handler registered, cannot use Wait")
	}
	<-e.evChan
	return nil
}

// WaitTimeout reads the event, returning if the timeout expires.
// This cannot be used if a handler has been installed on this device e.g
//  ok, err := e.WaitTimeout(time.Second)
//  if ok {
//      // Event received
//  else {
//      // Timed out
//  }
func (e *Event) WaitTimeout(tout time.Duration) (bool, error) {
	if s.handlerRegistered {
		return false, fmt.Errorf("Handler registered, cannot use WaitTimeout")
	}
	ticker := time.NewTicker(tout)
	defer ticker.Stop()
	select {
	case <-e.evChan:
		return true, nil
	case <-ticker.C:
		return false, nil
	}
}

// dispatcher is a shim between the channel and the
// external handler that will be invoked when an event is received.
// A stop channel is used to indicate when the handler should terminate.
func (e *Event) dispatcher(f func()) {
	for {
		select {
		case c := <-e.stopChan:
			// Send a value back to signal that the handler has terminated.
			c <- true
			return
		case v := <-e.evChan:
			// Reading a closed channel will return false
			if v {
				f()
			} else {
				return
			}
		}
	}
}
