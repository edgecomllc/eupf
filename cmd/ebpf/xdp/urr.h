// SPDX-License-Identifier: Apache-2.0
// Copyright 2023-2026 Edgecom LLC
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
