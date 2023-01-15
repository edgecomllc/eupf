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

enum default_action {
    DEFAULT_XDP_ACTION = XDP_PASS,
};

/* Header cursor to keep track of current parsing position */
struct packet_context {
    void            *data;
    void            *data_end;
    struct xdp_md   *ctx;
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

static __always_inline int parse_ip4(struct packet_context *ctx, struct iphdr **ip4hdr)
{
    void *data = ctx->data;
    void *data_end = ctx->data_end;

    struct iphdr *ip4 = data;
    const int hdrsize = sizeof(*ip4);

    if (data + hdrsize > data_end)
        return -1;

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

static __always_inline __u32 handle_echo_request(struct xdp_md *ctx, struct gtpuhdr *gtpu)
{
    return XDP_TX;
}

struct bpf_map_def SEC("maps") context_map_ip4 = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u32),      // IPv4
    .value_size = sizeof(__u32),    // SessionID
    .max_entries = 10,              // FIXME
};

struct bpf_map_def SEC("maps") context_map_teid = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u32),      // TEID
    .value_size = sizeof(__u32),    // SessionID
    .max_entries = 10,              // FIXME
};

static __always_inline __u32 handle_core_packet_ipv4(struct xdp_md *ctx, const struct iphdr *ip4)
{
    const __u32* session_id = bpf_map_lookup_elem(&context_map_ip4, &ip4->daddr);
     if(session_id == NULL)
         return DEFAULT_XDP_ACTION;

    bpf_printk("tail call to UPF_PROG_TYPE_MAIN key\n");
    bpf_tail_call(ctx, &upf_pipeline, UPF_PROG_TYPE_MAIN);
    bpf_printk("tail call to UPF_PROG_TYPE_MAIN key failed\n");
    return DEFAULT_XDP_ACTION;
}

static __always_inline __u32 handle_core_packet_ipv6(struct xdp_md *ctx, struct ipv6hdr *ip6)
{
    return XDP_DROP;
}

static __always_inline __u32 handle_access_packet(struct xdp_md *ctx, __u32 teid)
{
    __u32* session_id = bpf_map_lookup_elem(&context_map_teid, &teid);

    if(!session_id) {
        bpf_printk("No session for teid %d\n", teid);
        return DEFAULT_XDP_ACTION;
    }

    bpf_printk("Access packet [ teid:%d sessionid:%d ]\n", teid, session_id);
    bpf_tail_call(ctx, &upf_pipeline, UPF_PROG_TYPE_MAIN);
    bpf_printk("tail call to UPF_PROG_TYPE_MAIN key failed\n");
    return DEFAULT_XDP_ACTION;
}

static __always_inline __u32 handle_gtpu(struct packet_context *ctx)
{
    struct gtpuhdr *gtp;
    int pdu_type = parse_gtp(ctx, &gtp);
    switch(pdu_type) {
        case GTPU_G_PDU:
            
            return handle_access_packet(ctx->ctx, bpf_htonl(gtp->teid));
        case GTPU_ECHO_REQUEST:
            bpf_printk("upf: gtp header [ version=%d, pt=%d, e=%d]\n", gtp->version, gtp->pt, gtp->e);
            bpf_printk("upf: gtp echo request [ type=%d ]\n", pdu_type);
            return handle_echo_request(ctx->ctx, gtp);
        case GTPU_ECHO_RESPONSE:
        case GTPU_ERROR_INDICATION:
        case GTPU_SUPPORTED_EXTENSION_HEADERS_NOTIFICATION:
        case GTPU_END_MARKER:
            return DEFAULT_XDP_ACTION;
        default:
            bpf_printk("upf: unexpected gtp message: type=%d\n", pdu_type);
            return DEFAULT_XDP_ACTION;

    }
}

SEC("xdp/upf_ip_entrypoint")
int upf_ip_entrypoint_func(struct xdp_md *ctx)
{
    bpf_printk("upf_ip_entrypoint start\n");

    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;

    /* These keep track of the next header type and iterator pointer */
    struct packet_context context = { .data = data, .data_end = data_end, .ctx = ctx };

    struct ethhdr   *eth;
    struct iphdr    *ip4;
    struct ipv6hdr  *ip6;

    __u16 l3_protocol = parse_ethernet(&context, &eth);

    int l4_protocol = 0;
    switch (l3_protocol) {
        case ETH_P_IPV6: 
            l4_protocol = parse_ip6(&context, &ip6);
            break;
        case ETH_P_IP:
            l4_protocol = parse_ip4(&context, &ip4);
            break;
        case ETH_P_ARP: //Let kernel stack takes care
            bpf_printk("upf: arp received. passing to kernel\n");
            return XDP_PASS;
        default:
            return DEFAULT_XDP_ACTION;
    }

    switch(l4_protocol)
    {
        case IPPROTO_ICMP: //Let kernel stack takes care
            bpf_printk("upf: icmp received. passing to kernel\n");
            return XDP_PASS;
        case IPPROTO_UDP:
            if(GTP_UDP_PORT == parse_udp(&context)) {
                bpf_printk("upf: gtp-u received\n");
                return handle_gtpu(&context);
            }
            break;
        case IPPROTO_TCP:
            break;
        default:
            return DEFAULT_XDP_ACTION;
    }

    switch (l3_protocol) {
        case ETH_P_IPV6: 
            return handle_core_packet_ipv6(ctx, ip6);
        case ETH_P_IP:
            return handle_core_packet_ipv4(ctx, ip4);
        default:
            return DEFAULT_XDP_ACTION;
    }

    return DEFAULT_XDP_ACTION;
}

char _license[] SEC("license") = "GPL";