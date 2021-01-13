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
	"io"
)

type ram []byte

// Open creates a type that can use a Reader/Writer interface to the
// underlying byte array.
func (base ram) Open() *ramIO {
	return &ramIO{Data: base, max: cap(base)}
}

// ramIO implements various io interfaces, using an underlying byte array.
type ramIO struct {
	Data    []byte
	current int
	max     int
}

// Write copies the byte slice into the RAM array
func (r *ramIO) Write(p []byte) (int, error) {
	n := copy(r.Data[r.current:], p)
	r.current += n
	if n != len(p) {
		return n, io.EOF
	}
	return n, nil
}

// WriteAt copies the byte slice into the RAM array at the offset specified
func (r *ramIO) WriteAt(p []byte, offs int64) (int, error) {
	if int(offs) >= r.max {
		return 0, io.EOF
	}
	r.current = int(offs)
	return r.Write(p)
}

func (r *ramIO) WriteByte(b byte) error {
	if r.current >= r.max {
		return io.EOF
	}
	r.Data[r.current] = b
	r.current++
	return nil
}

// Seek moves the offset
func (r *ramIO) Seek(offs int64, whence int) (int64, error) {
	n := int(offs)
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		n += r.current
	case io.SeekEnd:
		n = r.max - n
	default:
		return 0, fmt.Errorf("unknown whence")
	}
	if n < 0 {
		return 0, fmt.Errorf("negative offset")
	}
	r.current = n
	return int64(r.current), nil
}

func (r *ramIO) ReadByte() (byte, error) {
	if r.current >= r.max {
		return 0, io.EOF
	}
	b := r.Data[r.current]
	r.current++
	return b, nil
}

func (r *ramIO) Read(p []byte) (int, error) {
	n := copy(p, r.Data[r.current:])
	r.current += n
	if n != len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (r *ramIO) ReadAt(p []byte, offs int64) (int, error) {
	if int(offs) >= r.max {
		return 0, io.EOF
	}
	r.current = int(offs)
	return r.Read(p)
}
