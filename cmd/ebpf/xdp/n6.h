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
#include <bpf/bpf_helpers.h>

#include "xdp/pdr.h"
#include "xdp/qer.h"

#include "xdp/utils/common.h"
#include "xdp/utils/gtp_utils.h"
#include "xdp/utils/packet_context.h"

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

    __u8 tos = far->transport_level_marking >> 8;

    bpf_printk("upf: use mapping %pI4 -> TEID:%d", &ip4->daddr, far->teid);
    return send_to_gtp_tunnel(ctx, far->localip, far->remoteip, tos, far->teid);
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

    __u8 tos = far->transport_level_marking >> 8;

    bpf_printk("upf: use mapping %pI6c -> TEID:%d", &ip6->daddr, far->teid);
    return send_to_gtp_tunnel(ctx, far->localip, far->remoteip, tos, far->teid);
}
