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
struct hdr_cursor {
    void *pos;
};

static __always_inline __u16 parse_ethernet(struct hdr_cursor *nh, void *data_end, struct ethhdr **ethhdr)
{
    struct ethhdr *eth = nh->pos;
    int hdrsize = sizeof(*eth);

    if (nh->pos + hdrsize > data_end)
        return -1;

    nh->pos += hdrsize;
    *ethhdr = eth;

    return bpf_htons(eth->h_proto); /* network-byte-order */
}

static __always_inline int parse_ip4(struct hdr_cursor *nh, void *data_end, struct iphdr **ip4hdr)
{
    struct iphdr *ip4 = nh->pos;
    int hdrsize = sizeof(*ip4);

    if (nh->pos + hdrsize > data_end)
        return -1;

    nh->pos += hdrsize;
    *ip4hdr = ip4;
    // tuple5->proto = ip4->protocol;
    // tuple5->dst_ip.ip4.addr = ip4->daddr;
    // tuple5->src_ip.ip4.addr = ip4->saddr;
    return ip4->protocol; /* network-byte-order */
}

static __always_inline int parse_ip6(struct hdr_cursor *nh, void *data_end, struct ipv6hdr **ip6hdr)
{
    struct ipv6hdr *ip6 = nh->pos;
    int hdrsize = sizeof(*ip6);

    if (nh->pos + hdrsize > data_end)
        return -1;

    nh->pos += hdrsize;
    *ip6hdr = ip6;
    // tuple5->proto = ip6->nexthdr;
    // tuple5->dst_ip.ip6 = ip6->daddr;
    // tuple5->src_ip.ip6 = ip6->saddr;

    return ip6->nexthdr; /* network-byte-order */
}

static __always_inline __u16 parse_udp(struct hdr_cursor *nh, void *data_end)
{
    struct udphdr *udp = nh->pos;
    int hdrsize = sizeof(*udp);

    if (nh->pos + hdrsize > data_end)
        return -1;

    nh->pos += hdrsize;
    // tuple5->src_port = udp->source;
    // tuple5->dst_port = udp->dest;
    return udp->dest;
}

static __always_inline __u32 parse_gtp(struct hdr_cursor *nh, void *data_end, struct gtpuhdr **gtphdr)
{
    struct gtpuhdr *gtp = nh->pos;
    int hdrsize = sizeof(*gtp);

    /* Byte-count bounds check; check if current pointer + size of header
     * is after data_end.
     */
    if (nh->pos + hdrsize > data_end)
        return -1;

    nh->pos += hdrsize;
    *gtphdr = gtp;

    return gtp->message_type;
}

static __always_inline __u32 handle_echo_request(struct xdp_md *ctx, struct gtpuhdr *gtpu)
{
    return XDP_DROP;
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

static __always_inline __u32 handle_core_packet_ipv4(struct xdp_md *ctx, struct iphdr *ip4)
{
    __u32* session_id = bpf_map_lookup_elem(&context_map_ip4, &ip4->daddr);

    if(!session_id)
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

    bpf_printk("Access packet > teid:%d sessionid:%d\n", teid, session_id);
    bpf_tail_call(ctx, &upf_pipeline, UPF_PROG_TYPE_MAIN);
    bpf_printk("tail call to UPF_PROG_TYPE_MAIN key failed\n");
    return DEFAULT_XDP_ACTION;
}

SEC("xdp/upf_ip_entrypoint")
int upf_ip_entrypoint_func(struct xdp_md *ctx)
{
    bpf_printk("upf_ip_entrypoint start\n");

    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;

    /* These keep track of the next header type and iterator pointer */
    struct hdr_cursor cursor = { .pos = data };

    struct ethhdr *eth;


    __u16 l3_protocol = parse_ethernet(&cursor, data_end, &eth);
    int l4_protocol = 0;
    switch (l3_protocol) {
        case ETH_P_IPV6: 
        {
            struct ipv6hdr *ip6;
            l4_protocol = parse_ip6(&cursor, data_end, &ip6);
            return handle_core_packet_ipv6(ctx, ip6);
        }
        case ETH_P_IP:
        {
            struct iphdr *ip4;
            l4_protocol = parse_ip4(&cursor, data_end, &ip4);
            return handle_core_packet_ipv4(ctx, ip4);
        }
        case ETH_P_ARP: //Let kernel stack takes care
            return XDP_PASS;
        default:
            return DEFAULT_XDP_ACTION;
    }

    switch(l4_protocol)
    {
        case IPPROTO_ICMP: //Let kernel stack takes care
            return XDP_PASS;
        case IPPROTO_UDP:
        {
            __u16 dest_port = parse_udp(&cursor, data_end);
            if(dest_port != GTP_UDP_PORT)
                return DEFAULT_XDP_ACTION;
            break;
        }
        default:
            return DEFAULT_XDP_ACTION;
    }
    

    struct gtpuhdr *gtp;
    int pdu_type = parse_gtp(&cursor, data_end, &gtp);
    switch(pdu_type) {
        case GTPU_G_PDU:
            return handle_access_packet(ctx, bpf_htonl(gtp->teid));
        case GTPU_ECHO_REQUEST:
            return handle_echo_request(ctx, gtp);
        case GTPU_ECHO_RESPONSE:
        case GTPU_ERROR_INDICATION:
        case GTPU_SUPPORTED_EXTENSION_HEADERS_NOTIFICATION:
        case GTPU_END_MARKER:
        default:
            return DEFAULT_XDP_ACTION;

    }

    return DEFAULT_XDP_ACTION;
}

char _license[] SEC("license") = "GPL";