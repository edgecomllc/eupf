/**
 * Copyright 2026 Edgecom LLC
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

/* LPM trie key: prefix length in bits followed by the IPv4 prefix in network byte order */
struct ue_rebalance_lpm_key {
    __u32 prefix_len;
    __u8  data[4];
};

/* UE prefixes eligible for downlink rebalance */
struct
{
    __uint(type, BPF_MAP_TYPE_LPM_TRIE);
    __type(key, struct ue_rebalance_lpm_key);
    __type(value, __u8);
    __uint(max_entries, 16);
    __uint(map_flags, BPF_F_NO_PREALLOC);
} ue_rebalance_lpm_map SEC(".maps");
