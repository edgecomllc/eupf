#pragma once

#include <linux/bpf.h>
#include <linux/types.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/udp.h>

#include <bpf/bpf_endian.h>

#include "xdp/utils/packet_context.h"

static __always_inline __u16 parse_ethernet(struct packet_context *ctx)
{
    struct ethhdr *eth = (struct ethhdr *)ctx->data;
    if ((void*)(eth + 1) > ctx->data_end)
        return -1;

    ctx->data += sizeof(*eth);
    ctx->eth = eth; 
    return bpf_htons(eth->h_proto);
}

/* 0x3FFF mask to check for fragment offset field */
#define IP_FRAGMENTED 65343

static __always_inline int parse_ip4(struct packet_context *ctx)
{
    struct iphdr *ip4 = (struct iphdr *)ctx->data;
    if ((void*)(ip4 + 1) > ctx->data_end)
        return -1;

    /* do not support fragmented packets as L4 headers may be missing */
    // if (ip4->frag_off & IP_FRAGMENTED)
    //	return -1;

    ctx->data += sizeof(*ip4);
    ctx->ip4 = ip4;
    return ip4->protocol; /* network-byte-order */
}

static __always_inline int parse_ip6(struct packet_context *ctx)
{
    struct ipv6hdr *ip6 = ctx->data;
    if ((void*)(ip6 + 1) > ctx->data_end)
        return -1;

    ctx->data += sizeof(*ip6);
    ctx->ip6 = ip6;
    return ip6->nexthdr; /* network-byte-order */
}

static __always_inline __u16 parse_udp(struct packet_context *ctx)
{
    struct udphdr *udp = (struct udphdr *)ctx->data;
    if ((void*)(udp + 1) > ctx->data_end)
        return -1;

    ctx->data += sizeof(*udp);
    ctx->udp = udp;
    return bpf_htons(udp->dest);
}