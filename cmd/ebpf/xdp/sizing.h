// SPDX-License-Identifier: Apache-2.0
// Copyright 2023-2026 Edgecom LLC
 #pragma once

#define MAX_SESSIONS 65535

#define PDR_MAP_SIZE MAX_SESSIONS * 2   //  2 PDR per session
#define FAR_MAP_SIZE PDR_MAP_SIZE       //  1 FAR per PDR
#define QER_MAP_SIZE MAX_SESSIONS       //  1 QWR per session
#define URR_LIST_SIZE 3                 //  3 URR per session
#define URR_MAP_SIZE MAX_SESSIONS * URR_LIST_SIZE
#define SDF_LIST_SIZE 5

#define XSTR(x) STR(x)
#define STR(x) #x
#pragma message "Max configured sessions:   " XSTR(MAX_SESSIONS)
#pragma message "Max configured PDRs:       " XSTR(PDR_MAP_SIZE)
#pragma message "Max configured FARs:       " XSTR(FAR_MAP_SIZE)
#pragma message "Max configured QERs:       " XSTR(QER_MAP_SIZE)
#pragma message "Max configured URRs:       " XSTR(URR_MAP_SIZE)
#pragma message "Max configured SDF per PDR: " XSTR(SDF_LIST_SIZE)