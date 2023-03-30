#pragma once

#include <bpf/bpf_helpers.h>
#include <linux/bpf.h>

enum gate_status {
    OPEN        = 0,
    CLOSED      = 1,
    RESERVED1   = 2,
    RESERVED2   = 3,
};

struct qer_info
{
    __u8    ul_gate_status;
    __u8    dl_gate_status;
    __u8    qfi;
    __u32   ul_maximum_bitrate;
    __u32   dl_maximum_bitrate;
};

#ifdef __RELEASE
struct bpf_map_def SEC("maps") qer_map = {
    .type = BPF_MAP_TYPE_ARRAY,
    .key_size = sizeof(__u32), // QER ID
    .value_size = sizeof(struct qer_info),
    .max_entries = 1024, // FIXME
};
#else
struct
{
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32); // qer id
    __type(value, struct qer_info);
    __uint(max_entries, 1024);
} qer_map SEC(".maps");
#endif