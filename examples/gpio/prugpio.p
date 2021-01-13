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

.origin 0
.entrypoint Start
;
; u32 event - r0
; u32 interrupt - r1
; u32 count - r2  count of pairs of GPIOs
; u32 current state
; u32 in    - r10
; u32 out   - r11
; ...
;
Start:
    MOV     r8,0
; Load first 3 32 bit values
    LBBO    r0, r8, 0, 12
IOLoop:
    MOV     r5,16           ; r5 = start of GPIO pairs
    MOV     r6,r2           ; r6 = pairs count
    SBBO    r31, r8, 12, 4  ; Store R31
Next:
    QBEQ    CheckInt,r6,0
    LBBO    r10, r5, 0, 8   ; r10 = in, r11 = out
    QBBS    SetOn,r31,r10   ; Test state of IN GPIO
    CLR     r30,r11         ; Clear OUT GPIO
    QBA     Continue
SetOn:
    SET     r30,r11         ; Set OUT GPIO
Continue:
    ADD     r5,r5,8         ; Increment address
    SUB     r6,r6,1         ; Decrement count
CheckInt:
    QBBC    IOLoop,r31,r1   ; If PRU interrupt set, exit
    OR      r31.b0,r0,0x20  ; Send completion signal
    HALT
