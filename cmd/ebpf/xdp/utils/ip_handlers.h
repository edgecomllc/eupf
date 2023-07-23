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

#pragma once

#include <linux/bpf.h>

#include "xdp/n6.h"
#include "xdp/statistics.h"

#include "xdp/utils/common.h"
#include "xdp/utils/parsers.h"

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
