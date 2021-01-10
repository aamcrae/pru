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
Start:
  MOV r31.b0, (0x20|2)
  MOV r31.b0, (0x20|3)
  MOV r31.b0, (0x20|4)
  MOV r31.b0, (0x20|5)
  MOV r31.b0, (0x20|6)
  MOV r31.b0, (0x20|8)
  MOV r31.b0, (0x20|9)
  HALT
