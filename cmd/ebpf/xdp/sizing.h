/**
 * Copyright 2023 Edgecom LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#pragma once

#define MAX_SESSIONS 65535

#define PDR_MAP_SIZE MAX_SESSIONS * 2 //  2 PDR per session
#define FAR_MAP_SIZE PDR_MAP_SIZE     //  1 FAR per PDR
#define QER_MAP_SIZE MAX_SESSIONS     //  1 QWR per session
#define URR_LIST_SIZE 2               //  2 URR per session
#define URR_MAP_SIZE MAX_SESSIONS *URR_LIST_SIZE
#define SDF_LIST_SIZE 5

#define XSTR(x) STR(x)
#define STR(x) #x
#pragma message "Max configured sessions:   " XSTR(MAX_SESSIONS)
#pragma message "Max configured PDRs:       " XSTR(PDR_MAP_SIZE)
#pragma message "Max configured FARs:       " XSTR(FAR_MAP_SIZE)
#pragma message "Max configured QERs:       " XSTR(QER_MAP_SIZE)
#pragma message "Max configured URRs:       " XSTR(URR_MAP_SIZE)
#pragma message "Max configured SDF per PDR: " XSTR(SDF_LIST_SIZE)