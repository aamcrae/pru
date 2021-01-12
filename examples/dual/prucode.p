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
; Generate the in-program data by:
;   pasm -m prucode.p prucode
;   ../../util/img2go.sh prucode main

.origin 0
.entrypoint Start

; u32 event - r0
; u32 count - r1
; u32 src   - r2
; u32 dest  - r3

;
Start:
  MOV r8, 0
; Load first 4 32 bit values
  LBBO r0, r8, 0, 16
MovLoop:
  LBBO r4, r2, 0, 4
  SBBO r4, r3, 0, 4
  ADD r2, r2, 4
  ADD r3, r3, 4
  SUB r1, r1, 1
  QBNE  MovLoop, r1, 0
; Send completion signal
  OR r31.b0, r0, 0x20
  HALT
