/**
 * Copyright 2023-2025 Edgecom LLC
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

#include "xdp/utils/trace.h"
#include "xdp/sizing.h"

struct urr_info {
    __u64 ul;
    __u64 dl;
};


/* URR ID -> URR */
struct
{
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32);
    __type(value, struct urr_info);
    __uint(max_entries, URR_MAP_SIZE);
} urr_map SEC(".maps");


static __always_inline void update_urr(__u32 urr_id, __u64 uplink_bytes, __u64 downlink_bytes)
{
    struct urr_info *urr = bpf_map_lookup_elem(&urr_map, &urr_id);  
    if (urr) {
        urr->ul += uplink_bytes;
        urr->dl += downlink_bytes;
        upf_printk("upf: urr:%u uplink:%u downlink:%u", urr_id, urr->ul, urr->dl);    
    }
}
