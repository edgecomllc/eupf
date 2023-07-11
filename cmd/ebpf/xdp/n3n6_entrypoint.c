/**
 * Copyright 2023 Edgecom LLC
 * 
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * 
 *     http://www.apache.org/licenses/LICENSE-2.0
 * 
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
#ifdef ENABLE_LOG  // trace_pipe logs disabled by default
#warning "Debug log enabled"
#define bpf_printk(fmt, ...)                       \
    ({                                             \
        static const char ____fmt[] = fmt;         \
        bpf_trace_printk(____fmt, sizeof(____fmt), \
                         ##__VA_ARGS__);           \
    })
#else
#define bpf_printk(fmt, ...)
#endif

#define DEFAULT_XDP_ACTION XDP_PASS

static __always_inline enum xdp_action send_to_gtp_tunnel(struct packet_context *ctx, int srcip, int dstip, int teid) {
    if (-1 == add_gtp_over_ip4_headers(ctx, srcip, dstip, teid))
        return XDP_ABORTED;

    bpf_printk("upf: send gtp pdu %pI4 -> %pI4", &ctx->ip4->saddr, &ctx->ip4->daddr);
    return route_ipv4(ctx->xdp_ctx, ctx->eth, ctx->ip4);
}

static __always_inline enum xdp_action handle_n6_packet_ipv4(struct packet_context *ctx) {
    const struct iphdr *ip4 = ctx->ip4;
    struct pdr_info *pdr = bpf_map_lookup_elem(&pdr_map_downlink_ip4, &ip4->daddr);
    if (!pdr) {
        bpf_printk("upf: no downlink session for ip:%pI4", &ip4->daddr);
        return DEFAULT_XDP_ACTION;
    }

    struct far_info *far = bpf_map_lookup_elem(&far_map, &pdr->far_id);
    if (!far) {
        bpf_printk("upf: no downlink session far for ip:%pI4 far:%d", &ip4->daddr, pdr->far_id);
        return XDP_DROP;
    }

    bpf_printk("upf: downlink session for ip:%pI4  far:%d action:%d", &ip4->daddr, pdr->far_id, far->action);

    // Only forwarding action is supported at the moment
    if (!(far->action & FAR_FORW))
        return XDP_DROP;

    // Only outer header GTP/UDP/IPv4 is supported at the moment
    if (!(far->outer_header_creation & OHC_GTP_U_UDP_IPv4))
        return XDP_DROP;

    struct qer_info *qer = bpf_map_lookup_elem(&qer_map, &pdr->qer_id);
    if (!qer) {
        bpf_printk("upf: no downlink session qer for ip:%pI4 qer:%d", &ip4->daddr, pdr->qer_id);
        return XDP_DROP;
    }

    bpf_printk("upf: qer:%d gate_status:%d mbr:%d", pdr->qer_id, qer->dl_gate_status, qer->dl_maximum_bitrate);

    if (qer->dl_gate_status != GATE_STATUS_OPEN)
        return XDP_DROP;

    const __u64 packet_size = ctx->xdp_ctx->data_end - ctx->xdp_ctx->data;
    if (XDP_DROP == limit_rate_sliding_window(packet_size, &qer->dl_start, qer->dl_maximum_bitrate))
        return XDP_DROP;

    bpf_printk("upf: use mapping %pI4 -> TEID:%d", &ip4->daddr, far->teid);
    return send_to_gtp_tunnel(ctx, far->localip, far->remoteip, far->teid);
}

static __always_inline enum xdp_action handle_n6_packet_ipv6(struct packet_context *ctx) {
    const struct ipv6hdr *ip6 = ctx->ip6;
    struct pdr_info *pdr = bpf_map_lookup_elem(&pdr_map_downlink_ip6, &ip6->daddr);
    if (!pdr) {
        bpf_printk("upf: no downlink session for ip:%pI6c", &ip6->daddr);
        return DEFAULT_XDP_ACTION;
    }

    struct far_info *far = bpf_map_lookup_elem(&far_map, &pdr->far_id);
    if (!far) {
        bpf_printk("upf: no downlink session far for ip:%pI6c far:%d", &ip6->daddr, pdr->far_id);
        return XDP_DROP;
    }

    bpf_printk("upf: downlink session for ip:%pI6c far:%d action:%d", &ip6->daddr, pdr->far_id, far->action);

    // Only forwarding action supported at the moment
    if (!(far->action & FAR_FORW))
        return XDP_DROP;

    // Only outer header GTP/UDP/IPv4 is supported at the moment
    if (!(far->outer_header_creation & OHC_GTP_U_UDP_IPv4))
        return XDP_DROP;

    struct qer_info *qer = bpf_map_lookup_elem(&qer_map, &pdr->qer_id);
    if (!qer) {
        bpf_printk("upf: no downlink session qer for ip:%pI6c qer:%d", &ip6->daddr, pdr->qer_id);
        return XDP_DROP;
    }

    bpf_printk("upf: qer:%d gate_status:%d mbr:%d", pdr->qer_id, qer->dl_gate_status, qer->dl_maximum_bitrate);

    if (qer->dl_gate_status != GATE_STATUS_OPEN)
        return XDP_DROP;

    const __u64 packet_size = ctx->xdp_ctx->data_end - ctx->xdp_ctx->data;
    if (XDP_DROP == limit_rate_sliding_window(packet_size, &qer->dl_start, qer->dl_maximum_bitrate))
        return XDP_DROP;

    bpf_printk("upf: use mapping %pI6c -> TEID:%d", &ip6->daddr, far->teid);
    return send_to_gtp_tunnel(ctx, far->localip, far->remoteip, far->teid);
}

static __always_inline enum xdp_action handle_n3_packet(struct packet_context *ctx) {
    if (!ctx->gtp) {
        bpf_printk("upf: unexpected packet context. no gtp header");
        return DEFAULT_XDP_ACTION;
    }

    /*
     *   Step 1: search for PDR and apply PDR instructions
     */
    __u32 teid = bpf_htonl(ctx->gtp->teid);
    struct pdr_info *pdr = bpf_map_lookup_elem(&pdr_map_uplink_ip4, &teid);
    if (!pdr) {
        bpf_printk("upf: no uplink session for teid:%d", teid);
        return DEFAULT_XDP_ACTION;
    }

    /*
     *   Step 2: search for FAR and apply FAR instructions
     */
    struct far_info *far = bpf_map_lookup_elem(&far_map, &pdr->far_id);
    if (!far) {
        bpf_printk("upf: no uplink session far for teid:%d far:%d", teid, pdr->far_id);
        return XDP_DROP;
    }

    bpf_printk("upf: far:%d action:%d outer_header_creation:%d", pdr->far_id, far->action, far->outer_header_creation);

    // Only forwarding action supported at the moment
    if (!(far->action & FAR_FORW))
        return XDP_DROP;

    /*
     *   Step 3: search for QER and apply QER instructions
     */
    struct qer_info *qer = bpf_map_lookup_elem(&qer_map, &pdr->qer_id);
    if (!qer) {
        bpf_printk("upf: no uplink session qer for teid:%d qer:%d", teid, pdr->qer_id);
        return XDP_DROP;
    }

    bpf_printk("upf: qer:%d gate_status:%d mbr:%d", pdr->qer_id, qer->ul_gate_status, qer->ul_maximum_bitrate);

    if (qer->ul_gate_status != GATE_STATUS_OPEN)
        return XDP_DROP;

    const __u64 packet_size = ctx->xdp_ctx->data_end - ctx->xdp_ctx->data;
    if (XDP_DROP == limit_rate_sliding_window(packet_size, &qer->ul_start, qer->ul_maximum_bitrate))
        return XDP_DROP;

    bpf_printk("upf: uplink session for teid:%d far:%d outer_header_removal:%d", teid, pdr->far_id, pdr->outer_header_removal);
    if (pdr->outer_header_removal == OHR_GTP_U_UDP_IPv4) {
        long result = remove_gtp_header(ctx);
        if (result) {
            bpf_printk("upf: handle_n3_packet: can't remove gtp header: %d", result);
            return XDP_ABORTED;
        }
    }

    /*
     *   Step 4: Route packet finally
     */
    if (ctx->ip4)
        return route_ipv4(ctx->xdp_ctx, ctx->eth, ctx->ip4);
    else if (ctx->ip6)
        return route_ipv6(ctx->xdp_ctx, ctx->eth, ctx->ip6);
    else
        return XDP_ABORTED;
}

static __always_inline enum xdp_action handle_gtpu(struct packet_context *ctx) {
    int pdu_type = parse_gtp(ctx);
    switch (pdu_type) {
        case GTPU_G_PDU:
            increment_counter(ctx->counters, rx_gtp_pdu);
            return handle_n3_packet(ctx);
        case GTPU_ECHO_REQUEST:
            increment_counter(ctx->counters, rx_gtp_echo);
            // bpf_printk("upf: gtp header [ version=%d, pt=%d, e=%d]", gtp->version, gtp->pt, gtp->e);
            // bpf_printk("upf: gtp echo request [ type=%d ]", pdu_type);
            bpf_printk("upf: gtp echo request [ %pI4 -> %pI4 ]", &ctx->ip4->saddr, &ctx->ip4->daddr);
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

static __always_inline enum xdp_action handle_ip4(struct packet_context *ctx) {
    int l4_protocol = parse_ip4(ctx);
    switch (l4_protocol) {
        case IPPROTO_ICMP: {
            increment_counter(ctx->counters, rx_icmp);
            break;
        }
        case IPPROTO_UDP:
            increment_counter(ctx->counters, rx_udp);
            if (GTP_UDP_PORT == parse_udp(ctx)) {
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

    return handle_n6_packet_ipv4(ctx);
}

static __always_inline enum xdp_action handle_ip6(struct packet_context *ctx) {
    int l4_protocol = parse_ip6(ctx);
    switch (l4_protocol) {
        case IPPROTO_ICMPV6:  // Let kernel stack takes care
            bpf_printk("upf: icmp received. passing to kernel");
            increment_counter(ctx->counters, rx_icmp6);
            return XDP_PASS;
        case IPPROTO_UDP:
            increment_counter(ctx->counters, rx_udp);
            // Don't expect GTP over IPv6 at the moment
            // if (GTP_UDP_PORT == parse_udp(ctx))
            // {
            //     bpf_printk("upf: gtp-u received");
            //     return handle_gtpu(ctx);
            // }
            break;
        case IPPROTO_TCP:
            increment_counter(ctx->counters, rx_tcp);
            break;
        default:
            increment_counter(ctx->counters, rx_other);
            return DEFAULT_XDP_ACTION;
    }

    return handle_n6_packet_ipv6(ctx);
}

static __always_inline enum xdp_action process_packet(struct packet_context *ctx) {
    __u16 l3_protocol = parse_ethernet(ctx);
    switch (l3_protocol) {
        case ETH_P_IPV6:
            increment_counter(ctx->counters, rx_ip6);
            return handle_ip6(ctx);
        case ETH_P_IP:
            increment_counter(ctx->counters, rx_ip4);
            return handle_ip4(ctx);
        case ETH_P_ARP:  // Let kernel stack takes care
        {
            increment_counter(ctx->counters, rx_arp);
            bpf_printk("upf: arp received. passing to kernel");
            return XDP_PASS;
        }
    }

    return DEFAULT_XDP_ACTION;
}

// Combined N3 & N6 entrypoint. Use for "on-a-stick" interfaces
SEC("xdp/upf_ip_entrypoint")
int upf_ip_entrypoint_func(struct xdp_md *ctx) {
    // bpf_printk("upf n3 & n6 combined entrypoint start");
    const __u32 key = 0;
    struct upf_statistic *statistic = bpf_map_lookup_elem(&upf_ext_stat, &key);
    if (!statistic) {
        const struct upf_statistic initval = {};
        bpf_map_update_elem(&upf_ext_stat, &key, &initval, BPF_ANY);
        statistic = bpf_map_lookup_elem(&upf_ext_stat, &key);
        if(!statistic)
            return XDP_ABORTED;
    }

    /* These keep track of the packet pointers and statistic */
    struct packet_context context = {
        .data = (char *)(long)ctx->data,
        .data_end = (const char *)(long)ctx->data_end,
        .xdp_ctx = ctx,
        .counters = &statistic->upf_counters};

    enum xdp_action action = process_packet(&context);
    statistic->xdp_actions[action & EUPF_MAX_XDP_ACTION_MASK] += 1;

    return action;
}

char _license[] SEC("license") = "GPL";