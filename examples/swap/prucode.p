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

;
Start:
  MOV r8, 0
; Load first 2 32 bit values
  LBBO r0, r8, 0, 8
; switch values
  MOV r2,r1
  MOV r3,r0
; store switched values
  SBBO r2, r8, 0, 8
  HALT
