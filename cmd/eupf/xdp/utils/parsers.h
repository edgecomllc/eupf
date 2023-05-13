#pragma once

#include <linux/bpf.h>
#include <linux/types.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/udp.h>

#include <bpf/bpf_endian.h>

#include "xdp/utils/packet_context.h"

static __always_inline __u16 parse_ethernet(struct packet_context *ctx, struct ethhdr **ethhdr)
{
    void *data = ctx->data;
    void *data_end = ctx->data_end;

    struct ethhdr *eth = data;
    const int hdrsize = sizeof(*eth);

    if (data + hdrsize > data_end)
        return -1;

    ctx->data += hdrsize;
    *ethhdr = eth;

    return bpf_htons(eth->h_proto); /* network-byte-order */
}

/* 0x3FFF mask to check for fragment offset field */
#define IP_FRAGMENTED 65343

static __always_inline int parse_ip4(struct packet_context *ctx, struct iphdr **ip4hdr)
{
    void *data = ctx->data;
    void *data_end = ctx->data_end;

    struct iphdr *ip4 = data;
    const int hdrsize = sizeof(*ip4);

    if (data + hdrsize > data_end)
        return -1;

    /* do not support fragmented packets as L4 headers may be missing */
    // if (ip4->frag_off & IP_FRAGMENTED)
    //	return -1;

    ctx->data += hdrsize;
    *ip4hdr = ip4;
    // tuple5->proto = ip4->protocol;
    // tuple5->dst_ip.ip4.addr = ip4->daddr;
    // tuple5->src_ip.ip4.addr = ip4->saddr;
    return ip4->protocol; /* network-byte-order */
}

static __always_inline int parse_ip6(struct packet_context *ctx, struct ipv6hdr **ip6hdr)
{
    void *data = ctx->data;
    void *data_end = ctx->data_end;

    struct ipv6hdr *ip6 = data;
    const int hdrsize = sizeof(*ip6);

    if (data + hdrsize > data_end)
        return -1;

    ctx->data += hdrsize;
    *ip6hdr = ip6;
    // tuple5->proto = ip6->nexthdr;
    // tuple5->dst_ip.ip6 = ip6->daddr;
    // tuple5->src_ip.ip6 = ip6->saddr;

    return ip6->nexthdr; /* network-byte-order */
}

static __always_inline __u16 parse_udp(struct packet_context *ctx)
{
    void *data = ctx->data;
    void *data_end = ctx->data_end;

    struct udphdr *udp = data;
    const int hdrsize = sizeof(*udp);

    if (data + hdrsize > data_end)
        return -1;

    ctx->udp = udp;
    ctx->data += hdrsize;
    // tuple5->src_port = udp->source;
    // tuple5->dst_port = udp->dest;
    return bpf_htons(udp->dest);
}