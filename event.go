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

const nEvents = 8

// Event represents an event device or event channel.
type Event struct {
	id                int
	file              *os.File
	handlerRegistered bool
	evChan            chan int
	hStop             chan chan int
}

// newEvent initialises the Event structure for the event device unit id.
func newEvent(p *PRU, id int) (*Event, error) {
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
		e.evChan = make(chan int, 50)
		e.file = f
		e.id = id
		p.events[id] = e
		go e.eventReader()
	}
	return e, nil
}

// Close frees the resources associated with this event.
func (e *Event) Close() {
	e.ClearHandler()
	close(e.evChan)
	e.file.Close()
	pru.events[e.id] = nil
}

// SetHandler installs an asynch handler that is invoked when events are
// read from the event device. The argument to the handler is the
// running count of how many events have been sent on this device.
func (e *Event) SetHandler(f func(int)) {
	if e.handlerRegistered {
		e.ClearHandler()
	}
	e.handlerRegistered = true
	e.hStop = make(chan chan int)
	go e.hDispatcher(f)
}

// ClearHandler removes any currently installed handler for this event device.
func (e *Event) ClearHandler() {
	if e.handlerRegistered {
		// Create a channel to be used to signal when the handler has exited.
		c := make(chan int)
		e.hStop <- c
		// Once the handler receives the stop channel, a value is signalled back
		// to indicate that the handler has exited.
		<-c
		close(e.hStop)
		e.handlerRegistered = false
	}
}

// Wait reads the event device channel and returns the value once available.
// This cannot be used if a handler has been installed on this event device.
func (e *Event) Wait() (int, error) {
	if e.handlerRegistered {
		return 0, fmt.Errorf("Handler registered, cannot use Wait")
	}
	return <-e.evChan, nil
}

// WaitTimeout reads the event device channel, returning if the timeout expires.
// This cannot be used if a handler has been installed on this event device e.g
//  v, ok, err := ev.WaitTimeout(time.Second)
//  if ok {
//      // Event received
//  else {
//      // Timed out
//  }
func (e *Event) WaitTimeout(tout time.Duration) (int, bool, error) {
	if e.handlerRegistered {
		return 0, false, fmt.Errorf("Handler registered, cannot use WaitTimeout")
	}
	ticker := time.NewTicker(tout)
	defer ticker.Stop()
	select {
	case v := <-e.evChan:
		return v, true, nil
	case <-ticker.C:
		return -1, false, nil
	}
}

// hDispatcher is a shim between the event channel and the
// external handler that will be invoked when an event is received.
// A stop channel is used to signal when the handler should terminate.
func (e *Event) hDispatcher(f func(int)) {
	for {
		select {
		case c := <-e.hStop:
			// Send a value back to signal that the handler has terminated.
			c <- 0
			return
		case v := <-e.evChan:
			f(v)
		}
	}
}

// eventReader polls the event device and sends the 32 bit value
// read from the device to the event channel.
func (e *Event) eventReader() {
	b := make([]byte, 4)
	for {
		n, err := e.file.Read(b)
		if err != nil {
			return
		}
		if n == 4 {
			val := *(*int32)(unsafe.Pointer(&b[0]))
			select {
			case e.evChan <- int(val):
			default:
				// Unable to send, maybe the channel is closed.
			}
		}
	}
}
