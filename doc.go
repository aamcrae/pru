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

/*

Package pru manages access and control to the Programmable Real-time Units (PRU)
of the TI AM335x (https://www.ti.com/processors/sitara-arm/applications/industrial-communications.html)
The commonly available product with this part is the Beaglebone Black (https://beagleboard.org/black)

The PRU subsystem includes 2 cores that can be separately controlled, and an event/interrupt
subsystem that allows the PRU cores and the host CPU to send and receive interrupt driven events.

This package is based on https://github.com/beagleboard/am335x_pru_package
The repo mentioned above is required in order to build and install the PRU assembler, if
custom PRU programs are to be loaded and executed. This repo also contains a number of useful
reference documents.

Complete documentation is available via https://github.com/aamcrae/pru, and through godoc at
https://godoc.org/github.com/aamcrae/pru

*/
package pru
