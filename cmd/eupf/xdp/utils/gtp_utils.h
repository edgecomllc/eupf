
#pragma once

#include <linux/types.h>
#include <linux/bpf.h>
#include <linux/in.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/udp.h>

#include "xdp/utils/packet_context.h"
#include "xdp/utils/gtpu.h"

static __always_inline __u32 parse_gtp(struct packet_context *ctx)
{
    void *data = ctx->data;
    void *data_end = ctx->data_end;

    struct gtpuhdr *gtp = data;
    const int hdrsize = sizeof(*gtp);

    if (data + hdrsize > data_end)
        return -1;

    ctx->data += hdrsize;
    ctx->gtp = gtp;
    return gtp->message_type;
}

static __always_inline __u32 handle_echo_request(struct packet_context *ctx)
{
    struct ethhdr   *eth = ctx->eth;
    struct iphdr    *iph = ctx->ip4;
    struct udphdr   *udp = ctx->udp;
    struct gtpuhdr  *gtp = ctx->gtp;

    if(!eth || !iph || !udp || !gtp)
        return XDP_ABORTED;

    gtp->message_type = GTPU_ECHO_RESPONSE;

    __u32 tmp_ip = iph->daddr;
    iph->daddr = iph->saddr;
    iph->saddr = tmp_ip;
    iph->check = 0;
    __u64 cs = 0;
    ipv4_csum(iph, sizeof(*iph), &cs);
    iph->check = cs;

    __u16 tmp = udp->dest;
    udp->dest = udp->source;
    udp->source = tmp;
    // Update UDP checksum
    udp->check = 0;
    //cs = 0;
    //ipv4_l4_csum(udp, sizeof(*udp), &cs, iph);
    //udp->check = cs;

    __u8 mac[6];
    __builtin_memcpy(mac, eth->h_source, sizeof(mac));
    __builtin_memcpy(eth->h_source, eth->h_dest, sizeof(eth->h_source));
    __builtin_memcpy(eth->h_dest, mac, sizeof(eth->h_dest));

    bpf_printk("upf: send gtp echo response [ %pI4 -> %pI4 ]", &iph->saddr, &iph->daddr);
    return XDP_TX;
}