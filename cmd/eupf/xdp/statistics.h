#pragma once

#include <bpf/bpf_helpers.h>
#include <linux/bpf.h>

// enum xdp_action {
// 	XDP_ABORTED = 0,
// 	XDP_DROP,
// 	XDP_PASS,
// 	XDP_TX,
// 	XDP_REDIRECT,
// };

struct bpf_map_def SEC("maps") upf_xdp_statistic = {
    .type = BPF_MAP_TYPE_ARRAY,
    .key_size = sizeof(__u32),      // xdp_action
    .value_size = sizeof(__u64), 
    .max_entries = 5,
};