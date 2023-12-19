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

#include <bpf/bpf_helpers.h>
#include <linux/bpf.h>

#include <stdint.h>

struct
{
    __uint(type, BPF_MAP_TYPE_PROG_ARRAY);
    __type(key, uint32_t);
    __type(value, uint32_t);
    __uint(max_entries, 16);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} upf_pipeline SEC(".maps");
