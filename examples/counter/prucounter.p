; Copyright 2021 Google LLC
;
; Licensed under the Apache License, Version 2.0 (the "License");
; you may not use this file except in compliance with the License.
; You may obtain a copy of the License at
;
;     https://www.apache.org/licenses/LICENSE-2.0
;
; Unless required by applicable law or agreed to in writing, software
; distributed under the License is distributed on an "AS IS" BASIS,
; WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
; See the License for the specific language governing permissions and
; limitations under the License.
;

.origin 0
.entrypoint Start

/*

u32 r0      Address of control registers
u32         Saved cycle count
u32         Saved stall count

*/
#define PCONTROL 0
#define PCYCLE   0x0C

;
Start:
    MOV     r0, 0
    MOV r8, 100                 ; Counter value
    LBBO    r2, r0, 0, 4        ; Get ptr to control registers
;
; Disable the counter, clear it, and then re-enable it
;
    LBBO    r3, r2, PCONTROL, 4 ; Read CONTROL register
    CLR     r3, 3               ; Disable counter
    SBBO    r3, r2, PCONTROL, 4 ; Write CONTROL register
    SBBO    r0, r2, PCYCLE, 4   ; cycle counter = 0
    SET     r3, 3               ; Enable counter
    SBBO    r3, r2, PCONTROL, 4 ; Write CONTROL register
CLoop:
    SUB r8, r8, 1
    QBNE CLoop, r8, 0           ; Counter loop
;
    LBBO    r4, r2, PCYCLE, 8   ; Get cycle and stall counter
    SBBO    r4, r0, 4, 8        ; Store in saved count memory
; Should have executed 202 cycles
; 200 in loop, 2 for read/write of values
    HALT
