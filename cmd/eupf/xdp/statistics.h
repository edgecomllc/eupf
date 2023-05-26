#pragma once

#pragma clang diagnostic ignored "-Wlanguage-extension-token"
#include <bpf/bpf_helpers.h>
#pragma clang diagnostic warning "-Wlanguage-extension-token"
#include <linux/bpf.h>

struct upf_counters {
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

#define EUPF_MAX_XDP_ACTION 8

struct upf_statistic {
    struct upf_counters upf_counters;
    __u64 xdp_actions[EUPF_MAX_XDP_ACTION];
};

struct
{
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, __u32);
    __type(value, struct upf_statistic);
    __uint(max_entries, 1);
} upf_ext_stat SEC(".maps");
