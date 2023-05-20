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
#include <sys/socket.h>

#include "xdp/program_array.h"
#include "xdp/statistics.h"
#include "xdp/qer.h"
#include "xdp/pdr.h"

#include "xdp/utils/common.h"
#include "xdp/utils/packet_context.h"
#include "xdp/utils/parsers.h"
#include "xdp/utils/csum.h"
#include "xdp/utils/gtp_utils.h"
#include "xdp/utils/routing.h"


#undef bpf_printk
//#define bpf_printk(fmt, ...)
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


static __always_inline __u32 handle_n6_packet_ipv4(struct xdp_md *ctx, const struct iphdr *ip4)
{
    struct pdr_info* pdr = bpf_map_lookup_elem(&pdr_map_downlink_ip4, &ip4->daddr);
    if(!pdr) {
            bpf_printk("upf: no downlink session for ip:%pI4", &ip4->daddr);
            return DEFAULT_XDP_ACTION;
    }

    struct far_info* far = bpf_map_lookup_elem(&far_map, &pdr->far_id);
    if(!far) {
        bpf_printk("upf: no downlink session far for ip:%pI4 far:%d", &ip4->daddr, pdr->far_id);
            return XDP_DROP;
    }

    bpf_printk("upf: downlink session for ip:%pI4  far:%d action:%d", &ip4->daddr, pdr->far_id, far->action);

    //Only forwarding action supported at the moment
    if(!(far->action & FAR_FORW))
        return XDP_DROP;

    struct qer_info* qer = bpf_map_lookup_elem(&qer_map, &pdr->qer_id);
    if(!qer) {
        bpf_printk("upf: no downlink session qer for ip:%pI4 qer:%d", &ip4->daddr, pdr->qer_id);
            return XDP_DROP;
    }

    bpf_printk("upf: qer:%d gate_status:%d mbr:%d", pdr->qer_id, qer->dl_gate_status, qer->dl_maximum_bitrate);

    if(qer->dl_gate_status != GATE_STATUS_OPEN)
        return XDP_DROP;

    if(XDP_DROP == limit_rate_sliding_window(ctx, &qer->dl_start, qer->dl_maximum_bitrate))
        return XDP_DROP;

    bpf_printk("upf: use mapping %pI4 -> TEID:%d", &ip4->daddr, far->teid);

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
    ip->saddr = far->localip;
    ip->daddr = far->remoteip;

    // Add the UDP header
    struct udphdr *udp = (void *)(ip + 1);
    if ((void *)(udp + 1) > data_end)
    {
        return XDP_DROP;
    }

    udp->source = bpf_htons(GTP_UDP_PORT);
    udp->dest = bpf_htons(GTP_UDP_PORT);
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
    gtp->teid = bpf_htonl(far->teid);

    __u64 cs = 0;
    ipv4_csum(ip, sizeof(*ip), &cs);
    ip->check = cs;

    //Fuck ebpf verifier. I give up
    //cs = 0;
    //const void* udp_start = (void*)udp;
    //const __u16 udp_len = bpf_htons(udp->len);
    //ipv4_l4_csum(udp, udp_len, &cs, ip);
    //udp->check = cs;

    bpf_printk("upf: send gtp pdu %pI4 -> %pI4", &ip->saddr, &ip->daddr);
    return route_ipv4(ctx, eth, ip);
}

static __always_inline __u32 handle_n6_packet_ipv6(struct xdp_md *ctx, struct ipv6hdr *ip6)
{
    struct pdr_info* pdr = bpf_map_lookup_elem(&pdr_map_downlink_ip6, &ip6->daddr);
    if(!pdr) {
            bpf_printk("upf: no downlink session for ip:%pI6c", &ip6->daddr);
            return DEFAULT_XDP_ACTION;
    }

    struct far_info* far = bpf_map_lookup_elem(&far_map, &pdr->far_id);
    if(!far) {
        bpf_printk("upf: no downlink session far for ip:%pI6c far:%d", &ip6->daddr, pdr->far_id);
            return XDP_DROP;
    }

    bpf_printk("upf: downlink session for ip:%pI6c far:%d action:%d", &ip6->daddr, pdr->far_id, far->action);

    //Only forwarding action supported at the moment
    if(!(far->action & FAR_FORW))
        return XDP_DROP;

    bpf_printk("upf: use mapping %pI6c -> TEID:%d", &ip6->daddr, far->teid);

    //TODO: incapsulate & apply routing
    return XDP_PASS;
}

static __always_inline __u32 handle_n3_packet(struct packet_context *ctx)
{
    if(!ctx->gtp)
    {
        bpf_printk("upf: unexpected packet context. no gtp header");
        return DEFAULT_XDP_ACTION;
    }

    __u32 teid = bpf_htonl(ctx->gtp->teid);

    struct pdr_info* pdr = bpf_map_lookup_elem(&pdr_map_uplink_ip4, &teid);
    if(!pdr) {
        bpf_printk("upf: no uplink session for teid:%d", teid);
        return DEFAULT_XDP_ACTION;
    }

    bpf_printk("upf: uplink session for teid:%d far:%d headrm:%d", teid, pdr->far_id, pdr->outer_header_removal);

    struct far_info* far = bpf_map_lookup_elem(&far_map, &pdr->far_id);
    if(!far) {
        bpf_printk("upf: no uplink session far for teid:%d far:%d", teid, pdr->far_id);
        return XDP_DROP;
    }

    bpf_printk("upf: far:%d action:%d outer_header_creation:%d", pdr->far_id, far->action, far->outer_header_creation);

    //Only forwarding action supported at the moment
    if(!(far->action & FAR_FORW))
        return XDP_DROP;

    struct qer_info* qer = bpf_map_lookup_elem(&qer_map, &pdr->qer_id);
    if(!qer) {
        bpf_printk("upf: no uplink session qer for teid:%d qer:%d", teid, pdr->qer_id);
        return XDP_DROP;
    }

    bpf_printk("upf: qer:%d gate_status:%d mbr:%d", pdr->qer_id, qer->ul_gate_status, qer->ul_maximum_bitrate);

    if(qer->ul_gate_status != GATE_STATUS_OPEN)
        return XDP_DROP;

    if(XDP_DROP == limit_rate_sliding_window(ctx->ctx, &qer->ul_start, qer->ul_maximum_bitrate))
        return XDP_DROP;

    if(pdr->outer_header_removal == OHR_GTP_U_UDP_IPv4) 
    {
        void *data = (void *)(long)ctx->ctx->data;
        void *data_end = (void *)(long)ctx->ctx->data_end;

        int ext_gtp_header_size = 0;
        if(ctx->gtp) {
            struct gtpuhdr *gtp = ctx->gtp;
            if (gtp->e || gtp->s || gtp->pn)
                ext_gtp_header_size += sizeof(struct gtp_hdr_ext) + 4;
        }

        const int GTP_ENCAPSULATED_SIZE = sizeof(struct iphdr) + sizeof(struct udphdr) + sizeof(struct gtpuhdr) + ext_gtp_header_size;
        struct ethhdr *eth = data;
        if ((void *)(eth + 1) > data_end)
        {
            return XDP_DROP;
        }

        struct ethhdr *new_eth = data + GTP_ENCAPSULATED_SIZE;
        if ((void *)(new_eth + 1) > data_end)
        {
            return XDP_DROP;
        }
        __builtin_memcpy(new_eth, eth, sizeof(*eth));

        bpf_xdp_adjust_head(ctx->ctx, GTP_ENCAPSULATED_SIZE);
    }

    void *data = (void *)(long)ctx->ctx->data;
    void *data_end = (void *)(long)ctx->ctx->data_end;
    struct packet_context context = {.data = data, .data_end = data_end, .ctx = ctx->ctx};
    __u16 l3_protocol = parse_ethernet(&context);
    switch (l3_protocol)
    {
    case ETH_P_IPV6:
    {
        return DEFAULT_XDP_ACTION;
    }
    case ETH_P_IP:
    {
        int l4_protocol = parse_ip4(&context);
        if(l4_protocol != -1)
        {
            return route_ipv4(context.ctx, context.eth, context.ip4);
        }
    }
    default:
        return DEFAULT_XDP_ACTION;
    }

    return XDP_PASS; // Now lets kernel takes care

    // bpf_printk("Access packet [ teid:%d sessionid:%d ]", teid, session_id);
    // bpf_tail_call(ctx, &upf_pipeline, UPF_PROG_TYPE_MAIN);
    // bpf_printk("tail call to UPF_PROG_TYPE_MAIN key failed");
    // return DEFAULT_XDP_ACTION;
}

static __always_inline __u32 handle_gtpu(struct packet_context *ctx)
{
    int pdu_type = parse_gtp(ctx);
    switch (pdu_type)
    {
    case GTPU_G_PDU:
        increment_counter(ctx->counters, rx_gtp_pdu);
        //if (ctx->ip4)
        //{
        //    bpf_printk("upf: gtp pdu [ %pI4 -> %pI4 ]", &ctx->ip4->saddr, &ctx->ip4->daddr);
        //}
        return handle_n3_packet(ctx);
    case GTPU_ECHO_REQUEST:
        increment_counter(ctx->counters, rx_gtp_echo);
        //bpf_printk("upf: gtp header [ version=%d, pt=%d, e=%d]", gtp->version, gtp->pt, gtp->e);
        //bpf_printk("upf: gtp echo request [ type=%d ]", pdu_type);
        if (ctx->ip4)
        {
            bpf_printk("upf: gtp echo request [ %pI4 -> %pI4 ]", &ctx->ip4->saddr, &ctx->ip4->daddr);
        }
        return handle_echo_request(ctx);
    case GTPU_ECHO_RESPONSE:
    case GTPU_ERROR_INDICATION:
    case GTPU_SUPPORTED_EXTENSION_HEADERS_NOTIFICATION:
    case GTPU_END_MARKER:
        increment_counter(ctx->counters, rx_gtp_other);
        return DEFAULT_XDP_ACTION;
    default:
        increment_counter(ctx->counters, rx_gtp_unexp);
        bpf_printk("upf: unexpected gtp message: type=%d", pdu_type);
        return DEFAULT_XDP_ACTION;
    }
}

static __always_inline __u32 handle_ip4(struct packet_context *ctx)
{
    int l4_protocol = parse_ip4(ctx);
    switch (l4_protocol)
    {
    case IPPROTO_ICMP: {
        increment_counter(ctx->counters, rx_icmp);
        break;
    }
    case IPPROTO_UDP:
        increment_counter(ctx->counters, rx_udp);
        if (GTP_UDP_PORT == parse_udp(ctx))
        {
            bpf_printk("upf: gtp-u received");
            return handle_gtpu(ctx);
        }
        break;
    case IPPROTO_TCP:
        increment_counter(ctx->counters, rx_tcp);
        break;
    default:
        increment_counter(ctx->counters, rx_other);
        return DEFAULT_XDP_ACTION;
    }

    return handle_n6_packet_ipv4(ctx->ctx, ctx->ip4);
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
        increment_counter(ctx->counters, rx_icmp6);
        return XDP_PASS;
    case IPPROTO_UDP:
        increment_counter(ctx->counters, rx_udp);
        if (GTP_UDP_PORT == parse_udp(ctx))
        {
            bpf_printk("upf: gtp-u received");
            return handle_gtpu(ctx);
        }
        break;
    case IPPROTO_TCP:
        increment_counter(ctx->counters, rx_tcp);
        break;
    default:
        increment_counter(ctx->counters, rx_other);
        return DEFAULT_XDP_ACTION;
    }

    return handle_n6_packet_ipv6(ctx->ctx, ip6);
}

static __always_inline __u32 process_packet(struct packet_context *ctx)
{
    __u16 l3_protocol = parse_ethernet(ctx);
    switch (l3_protocol)
    {
    case ETH_P_IPV6:
        increment_counter(ctx->counters, rx_ip6);
        return handle_ip6(ctx);
    case ETH_P_IP:
        increment_counter(ctx->counters, rx_ip4);
        return handle_ip4(ctx);
    case ETH_P_ARP: // Let kernel stack takes care
    {
        increment_counter(ctx->counters, rx_arp);
        bpf_printk("upf: arp received. passing to kernel");
        return XDP_PASS;
    }
    }

    return DEFAULT_XDP_ACTION;
}

SEC("xdp/upf_ip_entrypoint")
int upf_ip_entrypoint_func(struct xdp_md *ctx)
{
    //bpf_printk("upf_ip_entrypoint start");
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;

    __u32 cpu_ip = 0;
    struct upf_counters *upf_counters = bpf_map_lookup_elem(&upf_ext_stat, &cpu_ip);

    struct upf_statistic *statistic = bpf_map_lookup_elem(&upf_ext_stat2, &cpu_ip);

    /* These keep track of the next header type and iterator pointer */
    struct packet_context context = {.data = data, .data_end = data_end, .ctx = ctx, .counters = upf_counters};

    __u32 xdp_action = process_packet(&context);

    // TODO: move xdp action statistic to upf_ext_stat
    __u64 *counter = bpf_map_lookup_elem(&upf_xdp_statistic, &xdp_action);
    if (counter)
    {
        __sync_fetch_and_add(counter, 1);
    }

    if(xdp_action < EUPF_MAX_XDP_ACTION)
    {
        __sync_fetch_and_add(&statistic->xdp_actions[xdp_action], 1);   
    }

    return xdp_action;
}

char _license[] SEC("license") = "GPL";