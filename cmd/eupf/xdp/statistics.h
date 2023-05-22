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

#ifdef __RELEASE
struct bpf_map_def SEC("maps") upf_xdp_statistic = {
    .type = BPF_MAP_TYPE_ARRAY,
    .key_size = sizeof(__u32), // xdp_action
    .value_size = sizeof(__u64),
    .max_entries = 5,
};
#else
struct
{
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32); // xdp_action
    __type(value, __u64);
    __uint(max_entries, 5);
} upf_xdp_statistic SEC(".maps");
#endif


struct upf_counters
{
    __u64 rx_arp;
    __u64 rx_icmp;
    __u64 rx_icmp6;
    __u64 rx_ip4;
    __u64 rx_ip6;
    __u64 rx_tcp;
    __u64 rx_udp;
    __u64 rx_other;
    __u64 rx_gtp_echo;
    __u64 rx_gtp_pdu;
    __u64 rx_gtp_other;
    __u64 rx_gtp_unexp;
};

#ifdef __RELEASE
struct bpf_map_def SEC("maps") upf_ext_stat = {
    .type = BPF_MAP_TYPE_ARRAY,
    .key_size = sizeof(__u32), // cpu
    .value_size = sizeof(struct upf_counters),
    .max_entries = 1,
};

#else

// Using BTF for fun and for more comfortable debugging:
// Example:
// > bpftool map dump name upf_ext_stat
// [{
//         "key": 0,
//         "value": {
//             "rx_total": 0,
//             "rx_arp": 1,
//             "rx_icmp": 0,
//             "rx_icmp6": 8,
//             "rx_ip4": 0,
//             "rx_ip6": 28,
//             "rx_tcp": 0,
//             "rx_udp": 15,
//             "rx_other": 5,
//             "rx_gtp_echo": 0,
//             "rx_gtp_pdu": 0,
//             "rx_gtp_other": 0,
//             "rx_gtp_unexp": 0
//         }
//     }
// ]

struct
{
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32); // cpu
    __type(value, struct upf_counters);
    __uint(max_entries, 1);
} upf_ext_stat SEC(".maps");
#endif


#define EUPF_MAX_XDP_ACTION 5

struct upf_statistic {
    struct upf_counters upf_counters;
    __u32 xdp_actions[EUPF_MAX_XDP_ACTION];
};

struct
{
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32); // cpu
    __type(value, struct upf_statistic);
    __uint(max_entries, 1);
} upf_ext_stat2 SEC(".maps");