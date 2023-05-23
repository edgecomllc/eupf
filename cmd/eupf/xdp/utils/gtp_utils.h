
#pragma once

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/types.h>
#include <linux/udp.h>

#include "xdp/utils/gtpu.h"
#include "xdp/utils/packet_context.h"

static __always_inline __u32 parse_gtp(struct packet_context *ctx) {
    struct gtpuhdr *gtp = (struct gtpuhdr *)ctx->data;
    if ((void *)(gtp + 1) > ctx->data_end)
        return -1;

    ctx->data += sizeof(*gtp);
    ctx->gtp = gtp;
    return gtp->message_type;
}

static __always_inline __u32 handle_echo_request(struct packet_context *ctx) {
    struct ethhdr *eth = ctx->eth;
    struct iphdr *iph = ctx->ip4;
    struct udphdr *udp = ctx->udp;
    struct gtpuhdr *gtp = ctx->gtp;

    if (!eth || !iph || !udp || !gtp)
        return XDP_ABORTED;

    gtp->message_type = GTPU_ECHO_RESPONSE;

    // TODO: add support GTP over IPv6

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
    // cs = 0;
    // ipv4_l4_csum(udp, sizeof(*udp), &cs, iph);
    // udp->check = cs;

    __u8 mac[6];
    __builtin_memcpy(mac, eth->h_source, sizeof(mac));
    __builtin_memcpy(eth->h_source, eth->h_dest, sizeof(eth->h_source));
    __builtin_memcpy(eth->h_dest, mac, sizeof(eth->h_dest));

    bpf_printk("upf: send gtp echo response [ %pI4 -> %pI4 ]", &iph->saddr, &iph->daddr);
    return XDP_TX;
}

static __always_inline long remove_gtp_header(struct packet_context *ctx) {
    if (!ctx->gtp) {
        bpf_printk("upf: remove_gtp_header: not a gtp packet");
        return -1;
    }

    int ext_gtp_header_size = 0;
    struct gtpuhdr *gtp = ctx->gtp;
    if (gtp->e || gtp->s || gtp->pn)
        ext_gtp_header_size += sizeof(struct gtp_hdr_ext) + 4;

    const int GTP_ENCAPSULATED_SIZE = sizeof(struct iphdr) + sizeof(struct udphdr) + sizeof(struct gtpuhdr) + ext_gtp_header_size;

    void *data = (void *)(long)ctx->xdp_ctx->data;
    void *data_end = (void *)(long)ctx->xdp_ctx->data_end;
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end) {
        bpf_printk("upf: remove_gtp_header: can't parse eth");
        return -1;
    }

    struct ethhdr *new_eth = data + GTP_ENCAPSULATED_SIZE;
    if ((void *)(new_eth + 1) > data_end) {
        bpf_printk("upf: remove_gtp_header: can't set new eth");
        return -1;
    }
    __builtin_memcpy(new_eth, eth, sizeof(*eth));

    long result = bpf_xdp_adjust_head(ctx->xdp_ctx, GTP_ENCAPSULATED_SIZE);
    if (result)
        return result;

    // update packet pointers
    ctx->data = (void *)(long)ctx->xdp_ctx->data;
    ctx->data_end = (void *)(long)ctx->xdp_ctx->data_end;
    ctx->eth = 0;
    ctx->ip4 = 0;
    ctx->ip6 = 0;
    ctx->udp = 0;
    ctx->gtp = 0;

    // Have to implicitly inline `update_packet_context` here. Thanks to verifier.
    // if(-1 == update_packet_context(ctx))
    //     return XDP_ABORTED;
    __u16 l3_protocol = parse_ethernet(ctx);
    switch (l3_protocol) {
        case ETH_P_IPV6: {
            if (-1 == parse_ip6(ctx)) {
                bpf_printk("upf: can't parse ip6 after gtp header removal");
                return -1;
            }
            break;
        }
        case ETH_P_IP: {
            if (-1 == parse_ip4(ctx)) {
                bpf_printk("upf: can't parse ip4 after gtp header removal");
                return -1;
            }
            break;
        }
        default:
            // do nothing with non-ip packets
            bpf_printk("upf: can't process not an ip packet after gtp header removal: %d", l3_protocol);
            return -1;
    }

    return 0;
}

static __always_inline long add_gtp_header(struct packet_context *ctx, int saddr, int daddr, int teid) {
    static const int GTP_ENCAPSULATED_SIZE = sizeof(struct iphdr) + sizeof(struct udphdr) + sizeof(struct gtpuhdr);
    bpf_xdp_adjust_head(ctx->xdp_ctx, (__s32)-GTP_ENCAPSULATED_SIZE);

    void *data = (void *)(long)ctx->xdp_ctx->data;
    void *data_end = (void *)(long)ctx->xdp_ctx->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return -1;

    struct ethhdr *orig_eth = data + GTP_ENCAPSULATED_SIZE;
    if ((void *)(orig_eth + 1) > data_end)
        return -1;

    __builtin_memcpy(eth, orig_eth, sizeof(*eth));  // FIXME

    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return -1;

    struct iphdr *inner_ip = (void *)ip + GTP_ENCAPSULATED_SIZE;
    if ((void *)(inner_ip + 1) > data_end)
        return -1;

    // Add the outer IP header
    ip->version = 4;
    ip->ihl = 5;  // No options
    ip->tos = 0;
    ip->tot_len = bpf_htons(bpf_ntohs(inner_ip->tot_len) + GTP_ENCAPSULATED_SIZE);
    ip->id = 0;             // No fragmentation
    ip->frag_off = 0x0040;  // Don't fragment; Fragment offset = 0
    ip->ttl = 64;
    ip->protocol = IPPROTO_UDP;
    ip->check = 0;
    ip->saddr = saddr;
    ip->daddr = daddr;

    // Add the UDP header
    struct udphdr *udp = (void *)(ip + 1);
    if ((void *)(udp + 1) > data_end)
        return -1;

    udp->source = bpf_htons(GTP_UDP_PORT);
    udp->dest = bpf_htons(GTP_UDP_PORT);
    udp->len = bpf_htons(bpf_ntohs(inner_ip->tot_len) + sizeof(*udp) + sizeof(struct gtpuhdr));
    udp->check = 0;

    // Add the GTP header
    struct gtpuhdr *gtp = (void *)(udp + 1);
    if ((void *)(gtp + 1) > data_end)
        return -1;

    __u8 flags = GTP_FLAGS;  // FIXME
    __builtin_memcpy(gtp, &flags, sizeof(__u8));
    gtp->message_type = GTPU_G_PDU;
    gtp->message_length = inner_ip->tot_len;
    gtp->teid = bpf_htonl(teid);

    __u64 cs = 0;
    ipv4_csum(ip, sizeof(*ip), &cs);
    ip->check = cs;

    // Fuck ebpf verifier. I give up
    // cs = 0;
    // const void* udp_start = (void*)udp;
    // const __u16 udp_len = bpf_htons(udp->len);
    // ipv4_l4_csum(udp, udp_len, &cs, ip);
    // udp->check = cs;

    // update packet pointers
    ctx->data = data;
    ctx->data_end = data_end;
    ctx->eth = eth;
    ctx->ip4 = ip;
    ctx->ip6 = 0;
    ctx->udp = udp;
    ctx->gtp = gtp;
    return 0;
}