#pragma once

#include <linux/types.h>

#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/udp.h>
#include <linux/ipv6.h>
#include "xdp/utils/gtpu.h"

/* Header cursor to keep track of current parsing position */
struct packet_context
{
    void *data;
    const void *data_end;
    struct upf_counters *counters;
    struct xdp_md *xdp_ctx;
    struct ethhdr *eth;
    struct iphdr *ip4;
    struct ipv6hdr *ip6;
    struct udphdr *udp;
    struct gtpuhdr *gtp;
};