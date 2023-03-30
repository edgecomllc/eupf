//
// Created by pirog-spb on 14.12.2022.
//

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#include <linux/in.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/udp.h>

#include "xdp/gtpu.h"
#include "xdp/program_array.h"
#include "xdp/statistics.h"
#include "xdp/qer.h"

#undef bpf_printk
#define bpf_printk(fmt, ...)                       \
    ({                                             \
        static const char ____fmt[] = fmt;         \
        bpf_trace_printk(____fmt, sizeof(____fmt), \
                         ##__VA_ARGS__);           \
    })

#ifndef NULL
#define NULL 0
#endif

enum default_action
{
    DEFAULT_XDP_ACTION = XDP_PASS,
};

/* Header cursor to keep track of current parsing position */
struct packet_context
{
    void *data;
    void *data_end;
    struct upf_counters *counters;
    struct xdp_md *ctx;
    struct ethhdr *eth;
    struct iphdr *ip4;
    struct ipv6hdr *ip6;
    struct udphdr *udp;
};

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

static __always_inline __u32 parse_gtp(struct packet_context *ctx, struct gtpuhdr **gtphdr)
{
    void *data = ctx->data;
    void *data_end = ctx->data_end;

    struct gtpuhdr *gtp = data;
    const int hdrsize = sizeof(*gtp);

    if (data + hdrsize > data_end)
        return -1;

    ctx->data += hdrsize;
    *gtphdr = gtp;

    return gtp->message_type;
}

/* calculate ip header checksum */
// next_iph_u16 = (__u16 *)&iph_tnl;
// #pragma clang loop unroll(full)
// for (int i = 0; i < (int)sizeof(*iph) >> 1; i++)
//	csum += *next_iph_u16++;
// iph_tnl.check = ~((csum & 0xffff) + (csum >> 16));

static __always_inline __u16 csum_fold_helper(__u64 csum)
{
    int i;
#pragma unroll
    for (i = 0; i < 4; i++)
    {
        if (csum >> 16)
            csum = (csum & 0xffff) + (csum >> 16);
    }
    return ~csum;
}

static __always_inline void ipv4_csum(void *data_start, int data_size, __u64 *csum)
{
    *csum = bpf_csum_diff(0, 0, data_start, data_size, *csum);
    *csum = csum_fold_helper(*csum);
}

static __always_inline void ipv4_l4_csum(void *data_start, __u32 data_size,
                                         __u64 *csum, struct iphdr *iph)
{
    __u32 tmp = 0;
    *csum = bpf_csum_diff(0, 0, &iph->saddr, sizeof(__be32), *csum);
    *csum = bpf_csum_diff(0, 0, &iph->daddr, sizeof(__be32), *csum);
    tmp = __builtin_bswap32((__u32)(iph->protocol));
    *csum = bpf_csum_diff(0, 0, &tmp, sizeof(__u32), *csum);
    tmp = __builtin_bswap32((__u32)(data_size));
    *csum = bpf_csum_diff(0, 0, &tmp, sizeof(__u32), *csum);
    *csum = bpf_csum_diff(0, 0, data_start, data_size, *csum);
    *csum = csum_fold_helper(*csum);
}

static __always_inline __u32 handle_echo_request(struct packet_context *ctx, struct gtpuhdr *gtpu)
{
    struct ethhdr *eth = ctx->eth;
    struct iphdr *iph = ctx->ip4;
    struct udphdr *udp = ctx->udp;

    gtpu->message_type = GTPU_ECHO_RESPONSE;

    if (iph == NULL)
        return XDP_ABORTED;
    __u32 tmp_ip = iph->daddr;
    iph->daddr = iph->saddr;
    iph->saddr = tmp_ip;
    iph->check = 0;
    __u64 cs = 0;
    ipv4_csum(iph, sizeof(*iph), &cs);
    iph->check = cs;

    if (udp == NULL)
        return XDP_ABORTED;
    __u16 tmp = udp->dest;
    udp->dest = udp->source;
    udp->source = tmp;
    // Update UDP checksum
    udp->check = 0;
    cs = 0;
    ipv4_l4_csum(udp, sizeof(*udp), &cs, iph);
    udp->check = cs;

    __u8 mac[6];
    __builtin_memcpy(mac, eth->h_source, sizeof(mac));
    __builtin_memcpy(eth->h_source, eth->h_dest, sizeof(eth->h_source));
    __builtin_memcpy(eth->h_dest, mac, sizeof(eth->h_dest));

    bpf_printk("upf: send gtp echo response [ %pI4 -> %pI4 ]", &iph->saddr, &iph->daddr);
    return XDP_TX;
}

struct pdr_info
{
    __u8 outer_header_removal;
    __u16 far_id;
};

struct bpf_map_def SEC("maps") pdr_map_uplink_ip4 = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u32), // IPv4
    .value_size = sizeof(struct pdr_info),
    .max_entries = 1024, // FIXME
};

struct bpf_map_def SEC("maps") pdr_map_downlink_ip4 = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u32), // TEID
    .value_size = sizeof(struct pdr_info),
    .max_entries = 1024, // FIXME
};

struct far_info
{
    __u8 action;
    __u8 outer_header_creation;
    __u32 teid;
    __u32 srcip;
};

struct bpf_map_def SEC("maps") far_map = {
    .type = BPF_MAP_TYPE_ARRAY,
    .key_size = sizeof(__u32), // FAR ID
    .value_size = sizeof(struct far_info),
    .max_entries = 1024, // FIXME
};

struct gtp_tunnel_info
{
    __u32 teid;
    __u32 srcip;
    __u32 dstip;
    __u16 dstport;
};

struct bpf_map_def SEC("maps") context_map_ip4 = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u32), // IPv4
    .value_size = sizeof(struct gtp_tunnel_info),
    .max_entries = 10, // FIXME
};

struct bpf_map_def SEC("maps") context_map_ip6 = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct in6_addr), // IPv6
    .value_size = sizeof(__u32),         // TEID
    .max_entries = 10,                   // FIXME
};

// struct bpf_map_def SEC("maps") context_map_teid = {
//     .type = BPF_MAP_TYPE_HASH,
//     .key_size = sizeof(__u32),      // TEID
//     .value_size = sizeof(__u32),    // SessionID
//     .max_entries = 10,              // FIXME
// };

static __always_inline __u32 handle_core_packet_ipv4(struct xdp_md *ctx, const struct iphdr *ip4)
{

    const struct gtp_tunnel_info *tunnel = bpf_map_lookup_elem(&context_map_ip4, &ip4->daddr);
    if (tunnel == NULL)
    {
        bpf_printk("upf: no mapping for dest %pI4", &ip4->daddr);
        return DEFAULT_XDP_ACTION;
    }

    bpf_printk("upf: use mapping %pI4 -> TEID:%d", &ip4->daddr, tunnel->teid);

    static const int GTP_ENCAPSULATED_SIZE = sizeof(struct iphdr) + sizeof(struct udphdr) + sizeof(struct gtpuhdr);
    bpf_xdp_adjust_head(ctx, (__s32)-GTP_ENCAPSULATED_SIZE);

    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
    {
        return XDP_DROP;
    }

    struct ethhdr *orig_eth = data + GTP_ENCAPSULATED_SIZE;
    if ((void *)(orig_eth + 1) > data_end)
    {
        return XDP_DROP;
    }

    __builtin_memcpy(eth, orig_eth, sizeof(*eth)); // FIXME

    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
    {
        return XDP_DROP;
    }

    struct iphdr *inner_ip = (void *)ip + GTP_ENCAPSULATED_SIZE;
    if ((void *)(inner_ip + 1) > data_end)
    {
        return XDP_DROP;
    }

    // Add the outer IP header
    ip->version = 4;
    ip->ihl = 5; // No options
    ip->tos = 0;
    ip->tot_len = bpf_htons(bpf_ntohs(inner_ip->tot_len) + GTP_ENCAPSULATED_SIZE);
    ip->id = 0;            // No fragmentation
    ip->frag_off = 0x0040; // Don't fragment; Fragment offset = 0
    ip->ttl = 64;
    ip->protocol = IPPROTO_UDP;
    ip->check = 0;
    ip->saddr = tunnel->srcip;
    ip->daddr = tunnel->dstip; // p_far->forwarding_parameters.outer_header_creation.ipv4_address.s_addr;

    // Add the UDP header
    struct udphdr *udp = (void *)(ip + 1);
    if ((void *)(udp + 1) > data_end)
    {
        return XDP_DROP;
    }

    udp->source = bpf_htons(GTP_UDP_PORT);
    udp->dest = tunnel->dstport; // bpf_htons(p_far->forwarding_parameters.outer_header_creation.port_number);
    udp->len = bpf_htons(bpf_ntohs(inner_ip->tot_len) + sizeof(*udp) + sizeof(struct gtpuhdr));
    udp->check = 0;

    // Add the GTP header
    struct gtpuhdr *gtp = (void *)(udp + 1);
    if ((void *)(gtp + 1) > data_end)
    {
        return XDP_DROP;
    }

    __u8 flags = GTP_FLAGS;
    __builtin_memcpy(gtp, &flags, sizeof(__u8));
    gtp->message_type = GTPU_G_PDU;
    gtp->message_length = inner_ip->tot_len;
    gtp->teid = tunnel->teid; // p_far->forwarding_parameters.outer_header_creation.teid;

    __u64 cs = 0;
    ipv4_csum(ip, sizeof(*ip), &cs);
    ip->check = cs;

    cs = 0;
    ipv4_l4_csum(udp, sizeof(*udp), &cs, ip);
    udp->check = cs;

    bpf_printk("upf: send gtp pdu %pI4 -> %pI4", &ip->saddr, &ip->daddr);
    return XDP_PASS; // Let's kernel takes care

    //__builtin_memcpy(eth->h_dest, orig_eth->h_source, sizeof(orig_eth->h_source));
    //__builtin_memcpy(eth->h_source, orig_eth->h_dest, sizeof(orig_eth->h_dest));
    // eth->h_proto = orig_eth->h_proto;
    // return XDP_TX;

    // // Compute l3 checksum
    // __wsum l3sum = pcn_csum_diff(0, 0, (__be32 *)p_ip, sizeof(*p_ip), 0);
    // pcn_l3_csum_replace(p_ctx, IP_CSUM_OFFSET, 0, l3sum, 0);

    // bpf_printk("tail call to UPF_PROG_TYPE_MAIN key");
    // bpf_tail_call(ctx, &upf_pipeline, UPF_PROG_TYPE_MAIN);
    // bpf_printk("tail call to UPF_PROG_TYPE_MAIN key failed");
    // return DEFAULT_XDP_ACTION;
}

static __always_inline __u32 handle_core_packet_ipv6(struct xdp_md *ctx, struct ipv6hdr *ip6)
{
    return XDP_DROP;
}

static __always_inline __u32 handle_access_packet(struct packet_context *ctx, __u32 teid)
{
    if (ctx->ip4 && ctx->udp)
    {

        struct iphdr *inner_ip;
        if (-1 == parse_ip4(ctx, &inner_ip))
            return DEFAULT_XDP_ACTION;

        bpf_printk("upf: update mapping %pI4 -> TEID:%d", &inner_ip->saddr, teid);
        struct gtp_tunnel_info tunnel;
        __builtin_memset(&tunnel, 0, sizeof(tunnel));
        tunnel.teid = teid;
        tunnel.srcip = ctx->ip4->daddr;
        tunnel.dstip = ctx->ip4->saddr;
        tunnel.dstport = ctx->udp->source;
        bpf_map_update_elem(&context_map_ip4, &inner_ip->saddr, &tunnel, BPF_ANY);

        // __u32* session_teid = bpf_map_lookup_elem(&context_map_ip4, &ctx->ip4->saddr);

        // if(!session_teid) {
        //     bpf_printk("upf: no session for %pI4", &ctx->ip4->saddr);
        //     return DEFAULT_XDP_ACTION;
        // }
    }

    void *data = (void *)(long)ctx->ctx->data;
    void *data_end = (void *)(long)ctx->ctx->data_end;
    static const int GTP_ENCAPSULATED_SIZE = sizeof(struct iphdr) + sizeof(struct udphdr) + sizeof(struct gtpuhdr);
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
    {
        return DEFAULT_XDP_ACTION;
    }

    struct ethhdr *new_eth = data + GTP_ENCAPSULATED_SIZE;
    if ((void *)(new_eth + 1) > data_end)
    {
        return DEFAULT_XDP_ACTION;
    }
    __builtin_memcpy(new_eth, eth, sizeof(*eth));

    bpf_xdp_adjust_head(ctx->ctx, GTP_ENCAPSULATED_SIZE);

    return XDP_PASS; // Now lets kernel takes care

    // bpf_printk("Access packet [ teid:%d sessionid:%d ]", teid, session_id);
    // bpf_tail_call(ctx, &upf_pipeline, UPF_PROG_TYPE_MAIN);
    // bpf_printk("tail call to UPF_PROG_TYPE_MAIN key failed");
    // return DEFAULT_XDP_ACTION;
}

static __always_inline __u32 handle_gtpu(struct packet_context *ctx)
{
    struct gtpuhdr *gtp;
    int pdu_type = parse_gtp(ctx, &gtp);
    switch (pdu_type)
    {
    case GTPU_G_PDU:
        if (ctx->ip4)
        {
            bpf_printk("upf: gtp pdu [ %pI4 -> %pI4 ]", &ctx->ip4->saddr, &ctx->ip4->daddr);
        }
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_gtp_pdu, 1);
        return handle_access_packet(ctx, bpf_htonl(gtp->teid));
    case GTPU_ECHO_REQUEST:
        bpf_printk("upf: gtp header [ version=%d, pt=%d, e=%d]", gtp->version, gtp->pt, gtp->e);
        bpf_printk("upf: gtp echo request [ type=%d ]", pdu_type);
        if (ctx->ip4)
        {
            bpf_printk("upf: gtp echo request [ %pI4 -> %pI4 ]", &ctx->ip4->saddr, &ctx->ip4->daddr);
        }
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_gtp_echo, 1);
        return handle_echo_request(ctx, gtp);
    case GTPU_ECHO_RESPONSE:
    case GTPU_ERROR_INDICATION:
    case GTPU_SUPPORTED_EXTENSION_HEADERS_NOTIFICATION:
    case GTPU_END_MARKER:
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_gtp_other, 1);
        return DEFAULT_XDP_ACTION;
    default:
        bpf_printk("upf: unexpected gtp message: type=%d", pdu_type);
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_gtp_unexp, 1);
        return DEFAULT_XDP_ACTION;
    }
}

static __always_inline __u32 handle_ip4(struct packet_context *ctx)
{
    struct iphdr *ip4;
    int l4_protocol = parse_ip4(ctx, &ip4);
    ctx->ip4 = ip4; // fixme

    switch (l4_protocol)
    {
    case IPPROTO_ICMP: // Let kernel stack takes care
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_icmp, 1);
        bpf_printk("upf: icmp received. passing to kernel");
        return XDP_PASS;
    case IPPROTO_UDP:
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_udp, 1);
        if (GTP_UDP_PORT == parse_udp(ctx))
        {
            bpf_printk("upf: gtp-u received");
            return handle_gtpu(ctx);
        }
        break;
    case IPPROTO_TCP:
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_tcp, 1);
        break;
    default:
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_other, 1);
        return DEFAULT_XDP_ACTION;
    }

    return handle_core_packet_ipv4(ctx->ctx, ip4);
}

static __always_inline __u32 handle_ip6(struct packet_context *ctx)
{
    struct ipv6hdr *ip6;
    int l4_protocol = parse_ip6(ctx, &ip6);
    ctx->ip6 = ip6; // fixme

    switch (l4_protocol)
    {
    case IPPROTO_ICMPV6: // Let kernel stack takes care
        bpf_printk("upf: icmp received. passing to kernel");
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_icmp6, 1);
        return XDP_PASS;
    case IPPROTO_UDP:
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_udp, 1);
        if (GTP_UDP_PORT == parse_udp(ctx))
        {
            bpf_printk("upf: gtp-u received");
            return handle_gtpu(ctx);
        }
        break;
    case IPPROTO_TCP:
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_tcp, 1);
        break;
    default:
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_other, 1);
        return DEFAULT_XDP_ACTION;
    }

    return handle_core_packet_ipv6(ctx->ctx, ip6);
}

static __always_inline __u32 process_packet(struct packet_context *ctx)
{
    struct ethhdr *eth;
    __u16 l3_protocol = parse_ethernet(ctx, &eth);
    ctx->eth = eth; // fixme
    switch (l3_protocol)
    {
    case ETH_P_IPV6:
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_ip6, 1);
        return handle_ip6(ctx);
    case ETH_P_IP:
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_ip4, 1);
        return handle_ip4(ctx);
    case ETH_P_ARP: // Let kernel stack takes care
    {
        if (ctx->counters)
            __sync_fetch_and_add(&ctx->counters->rx_arp, 1);
        bpf_printk("upf: arp received. passing to kernel");
        return XDP_PASS;
    }
    }

    return DEFAULT_XDP_ACTION;
}

SEC("xdp/upf_ip_entrypoint")
int upf_ip_entrypoint_func(struct xdp_md *ctx)
{
    bpf_printk("upf_ip_entrypoint start");
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;

    __u32 cpu_ip = 0;
    struct upf_counters *upf_counters = bpf_map_lookup_elem(&upf_ext_stat, &cpu_ip);

    /* These keep track of the next header type and iterator pointer */
    struct packet_context context = {.data = data, .data_end = data_end, .ctx = ctx, .counters = upf_counters};

    __u32 xdp_action = process_packet(&context);

    // TODO: move xdp action statistic to upf_ext_stat
    __u64 *counter = bpf_map_lookup_elem(&upf_xdp_statistic, &xdp_action);
    if (counter)
    {
        __sync_fetch_and_add(counter, 1);
    }

    return xdp_action;
}

char _license[] SEC("license") = "GPL";