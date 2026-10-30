// SPDX-License-Identifier: Apache-2.0
// Copyright 2023-2026 Edgecom LLC
#pragma once

#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/types.h>
#include <linux/udp.h>
#include <linux/tcp.h>
#include "xdp/utils/gtpu.h"

/* Header cursor to keep track of current parsing position */
struct packet_context {
    char *data;
    const char *data_end;
    struct upf_counters *counters;
    struct n3_n6_counters *n3_n6_counter;
    struct xdp_md *xdp_ctx;
    struct ethhdr *eth;
    struct iphdr *ip4;
    struct ipv6hdr *ip6;
    struct udphdr *udp;
    struct tcphdr *tcp;
    struct gtpuhdr *gtp;
    __u8 trace;
};
