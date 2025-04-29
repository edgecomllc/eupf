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
#include <linux/tcp.h>
#include <linux/udp.h>
#include <sys/socket.h>

#include "xdp/program_array.h"
#include "xdp/statistics.h"
#include "xdp/qer.h"
#include "xdp/urr.h"
#include "xdp/pdr.h"
#include "xdp/sdf_filter.h"

#include "xdp/utils/common.h"
#include "xdp/utils/trace.h"
#include "xdp/utils/packet_context.h"
#include "xdp/utils/packet_trace.h"
#include "xdp/utils/parsers.h"
#include "xdp/utils/csum.h"
#include "xdp/utils/gtp_utils.h"
#include "xdp/utils/routing.h"
#include "xdp/utils/icmp.h"


#define DEFAULT_XDP_ACTION XDP_PASS

struct dataplane_config {
    __u32 n3_ipv4_address;
    __u32 n9_ipv4_address;  
} global_config;

static __always_inline enum xdp_action send_to_gtp_tunnel(struct packet_context *ctx, int srcip, int dstip, __u8 tos, __u8 qfi, int teid) {
    if (-1 == add_gtp_over_ip4_headers(ctx, srcip, dstip, tos, qfi, teid))
        return XDP_ABORTED;
    upf_printk("upf: send gtp pdu %pI4 -> %pI4", &ctx->ip4->saddr, &ctx->ip4->daddr);
    increment_counter(ctx->n3_n6_counter, tx_n3);
    return route_ipv4(ctx->xdp_ctx, ctx->eth, ctx->ip4);
}

static __always_inline const struct pdr* check_sdf_filters_ipv4(struct packet_context *ctx, /*const*/ struct pdr_info* session, __u32 ref)
{
    const int sdf_filters_num = sizeof(session->dedicated_pdrs)/sizeof(session->dedicated_pdrs[0]);
    for (int i = 0; i < sdf_filters_num; i++) {
        /*const*/ struct sdf_filter *sdf = &session->dedicated_pdrs[i].sdf_filter;
        if(sdf->protocol && match_sdf_filter_ipv4(ctx, sdf)) {
            /*const*/ struct sdf_rule* rule = &session->dedicated_pdrs[i];
            if(rule->notify == 1) { //FIXME: need more appropriate way to make notifications
                rule->notify += 1;
                send_sdf_notify_ip4(ctx, ref);
            }
            return &rule->pdr;
        }
    }

    return 0;
}

static __always_inline const struct pdr* check_sdf_filters_ipv6(struct packet_context *ctx, const struct pdr_info* session, __u32 ref)
{
    const int sdf_filters_num = sizeof(session->dedicated_pdrs)/sizeof(session->dedicated_pdrs[0]);
    for (int i = 0; i < sdf_filters_num; i++) {
        const struct sdf_filter *sdf = &session->dedicated_pdrs[i].sdf_filter;
        if(sdf->protocol && match_sdf_filter_ipv6(ctx, sdf)){
             const struct sdf_rule* rule = &session->dedicated_pdrs[i];
            if(rule->notify)
                send_sdf_notify_ip6(ctx, ref);
            return &rule->pdr;
        }
    }

    return 0;
}

static __always_inline const struct pdr* check_sdf_filters_gtp(struct packet_context *ctx, /*const*/ struct pdr_info* session, __u32 ref)
{
    struct packet_context inner_context = {
        .xdp_ctx = ctx->xdp_ctx,
        .data = (char *)(long)ctx->data,
        .data_end = (const char *)(long)ctx->data_end,
    };

    if (inner_context.data + 1 > inner_context.data_end)
        return 0;

    int eth_protocol = guess_eth_protocol(inner_context.data);
    switch (eth_protocol) {
        case ETH_P_IP_BE:
        {
            int ip_protocol = parse_ip4(&inner_context);
            if (-1 == ip_protocol) {
                upf_printk("upf: [n3] unable to parse IPv4 header");
                return 0;
            }

            if( -1 == parse_l4(ip_protocol, &inner_context)) {
                upf_printk("upf: [n3] unable to parse L4 header");
                return 0;
            }

            const int sdf_filters_num = sizeof(session->dedicated_pdrs)/sizeof(session->dedicated_pdrs[0]);
            for (int i = 0; i < sdf_filters_num; i++) {
                /*const*/ struct sdf_filter *sdf = &session->dedicated_pdrs[i].sdf_filter;
                if(sdf->protocol && match_sdf_filter_ipv4(&inner_context, sdf)){
                    /*const*/ struct sdf_rule* rule = &session->dedicated_pdrs[i];
                    if(rule->notify == 1) {
                        rule->notify += 1;
                        send_sdf_notify_ip4(&inner_context, ref);
                    }
                    return &rule->pdr;
                }
            }
            break;
        }
        case ETH_P_IPV6_BE:
        {
            int ip_protocol = parse_ip6(&inner_context);
            if (ip_protocol == -1) {
                upf_printk("upf: [n3] unable to parse IPv6 header");
                return 0;
            }

            if( -1 == parse_l4(ip_protocol, &inner_context)) {
                upf_printk("upf: [n3] unable to parse L4 header");
                return 0;
            }

            const int sdf_filters_num = sizeof(session->dedicated_pdrs)/sizeof(session->dedicated_pdrs[0]);
            for (int i = 0; i < sdf_filters_num; i++) {
                const struct sdf_filter *sdf = &session->dedicated_pdrs[i].sdf_filter;
                if(sdf->protocol && match_sdf_filter_ipv6(&inner_context, sdf)){
                    const struct sdf_rule* rule = &session->dedicated_pdrs[i];
                    if(rule->notify)
                        send_sdf_notify_ip6(&inner_context, ref);
                    return &rule->pdr;
                }
            }
            break;
        }
        default:
            upf_printk("upf: [n3] unsupported inner ethernet protocol: %d", eth_protocol);
            break;
    }

    return 0;
}

static __always_inline enum xdp_action handle_n6_packet_ipv4(struct packet_context *ctx) {
    const struct iphdr *ip4 = ctx->ip4;
    struct pdr_info *session = bpf_map_lookup_elem(&pdr_map_downlink_ip4, &ip4->daddr);
    if (!session) {
        upf_printk("upf: [n6] no downlink session for ip:%pI4", &ip4->daddr);
        return DEFAULT_XDP_ACTION;
    }

    upf_printk("upf: [n6] downlink session for ip:%pI4 trace:%d", &ip4->daddr, session->trace_flag);
    if(session->trace_flag)
        trace_packet(ctx, PACKET_DIRECTION_IN);

    // Set defaults
    const struct pdr *pdr = &session->default_pdr;
    if (session->sdf_mode) {
        const struct pdr *pdr_sdf = check_sdf_filters_ipv4(ctx, session, ip4->daddr);
        if(pdr_sdf) {
            upf_printk(" [n6] Packet with source ip:%pI4 and destination ip:%pI4 matches SDF filter1", &ip4->saddr, &ip4->daddr);
            pdr = pdr_sdf;
        }
    }

    struct far_info *far = bpf_map_lookup_elem(&far_map, &pdr->far_id);
    if (!far) {
        upf_printk("upf: [n6] no downlink session far for ip:%pI4 far:%d", &ip4->daddr, pdr->far_id);
        return XDP_DROP;
    }

    upf_printk("upf: [n6] downlink session for ip:%pI4  far:%d action:%d", &ip4->daddr, pdr->far_id, far->action);

    if ((far->action & FAR_NOCP) && far->trigger == 0) {
          far->trigger = 1;
    }

    // Only forwarding action is supported at the moment
    if (!(far->action & FAR_FORW))
        return XDP_DROP;

    // Only outer header GTP/UDP/IPv4 is supported at the moment
    if (!(far->outer_header_creation & OHC_GTP_U_UDP_IPv4))
        return XDP_DROP;

    struct qer_info *qer = bpf_map_lookup_elem(&qer_map, &pdr->qer_id);
    if (!qer) {
        upf_printk("upf: [n6] no downlink session qer for ip:%pI4 qer:%d", &ip4->daddr, pdr->qer_id);
        return XDP_DROP;
    }

    upf_printk("upf: [n6] qer:%d gate_status:%d mbr:%u", pdr->qer_id, qer->dl_gate_status, qer->dl_maximum_bitrate);

    if (qer->dl_gate_status != GATE_STATUS_OPEN)
        return XDP_DROP;

    const __u64 packet_size = ctx->xdp_ctx->data_end - ctx->xdp_ctx->data;
    if (XDP_DROP == limit_rate_sliding_window(packet_size, &qer->dl_start, qer->dl_maximum_bitrate))
        return XDP_DROP;

    __u8 tos = qer->dscp ? qer->dscp : far->transport_level_marking >> 8;

    const int urr_size = sizeof(pdr->urr_id)/sizeof(pdr->urr_id[0]);
    for (int i = 0; i < urr_size; i++) {
        update_urr(pdr->urr_id[i], 0, packet_size);
    }

    upf_printk("upf: [n6] use mapping %pI4 -> teid:%u", &ip4->daddr, far->teid);
    enum xdp_action action = send_to_gtp_tunnel(ctx, far->localip, far->remoteip, tos, qer->qfi, far->teid);

    if(session->trace_flag && (action == XDP_TX || action == XDP_REDIRECT))
        trace_packet(ctx, PACKET_DIRECTION_OUT);

    return action;
}

static __always_inline enum xdp_action handle_n6_packet_ipv6(struct packet_context *ctx) {
    const struct ipv6hdr *ip6 = ctx->ip6;
    struct pdr_info *session = bpf_map_lookup_elem(&pdr_map_downlink_ip6, &ip6->daddr);
    if (!session) {
        upf_printk("upf: [n6] no downlink session for ip:%pI6c", &ip6->daddr);
        return DEFAULT_XDP_ACTION;
    }
    
    upf_printk("upf: [n6] downlink session for ip:%pI6c trace:%d", &ip6->daddr, session->trace_flag);
    if(session->trace_flag)
        trace_packet(ctx, PACKET_DIRECTION_IN);

    // Set defaults
    const struct pdr *pdr = &session->default_pdr;
    if (session->sdf_mode) {
        const struct pdr *pdr_sdf = check_sdf_filters_ipv6(ctx, session, 0);
        if(pdr_sdf) {
            upf_printk(" [n6] Packet with source ip:%pI6c and destination ip:%pI6c matches SDF filter", &ip6->saddr, &ip6->daddr);
            pdr = pdr_sdf;
        }
    }

    struct far_info *far = bpf_map_lookup_elem(&far_map, &pdr->far_id);
    if (!far) {
        upf_printk("upf: [n6] no downlink session far for ip:%pI6c far:%d", &ip6->daddr, pdr->far_id);
        return XDP_DROP;
    }

     if ((far->action & FAR_NOCP) && far->trigger == 0) {
        far->trigger = 1;
     }

    upf_printk("upf: [n6] downlink session for ip:%pI6c far:%d action:%d", &ip6->daddr, pdr->far_id, far->action);

    // Only forwarding action supported at the moment
    if (!(far->action & FAR_FORW))
        return XDP_DROP;

    // Only outer header GTP/UDP/IPv4 is supported at the moment
    if (!(far->outer_header_creation & OHC_GTP_U_UDP_IPv4))
        return XDP_DROP;

    struct qer_info *qer = bpf_map_lookup_elem(&qer_map, &pdr->qer_id);
    if (!qer) {
        upf_printk("upf: [n6] no downlink session qer for ip:%pI6c qer:%d", &ip6->daddr, pdr->qer_id);
        return XDP_DROP;
    }

    upf_printk("upf: [n6] qer:%d gate_status:%d mbr:%u", pdr->qer_id, qer->dl_gate_status, qer->dl_maximum_bitrate);

    if (qer->dl_gate_status != GATE_STATUS_OPEN)
        return XDP_DROP;

    const __u64 packet_size = ctx->xdp_ctx->data_end - ctx->xdp_ctx->data;
    if (XDP_DROP == limit_rate_sliding_window(packet_size, &qer->dl_start, qer->dl_maximum_bitrate))
        return XDP_DROP;

    __u8 tos = qer->dscp ? qer->dscp : far->transport_level_marking >> 8;

    const int urr_size = sizeof(pdr->urr_id)/sizeof(pdr->urr_id[0]);
    for (int i = 0; i < urr_size; i++) {
        update_urr(pdr->urr_id[i], 0, packet_size);
    }

    upf_printk("upf: [n6] use mapping %pI6c -> teid:%u", &ip6->daddr, far->teid);
    enum xdp_action action = send_to_gtp_tunnel(ctx, far->localip, far->remoteip, tos, qer->qfi, far->teid);

    if(session->trace_flag && (action == XDP_TX || action == XDP_REDIRECT))
        trace_packet(ctx, PACKET_DIRECTION_OUT);

    return action;
}


static __always_inline enum xdp_action handle_gtp_packet(struct packet_context *ctx) {
    if (!ctx->gtp) {
        upf_printk("upf: [n3] unexpected packet context. no gtp header");
        return DEFAULT_XDP_ACTION;
    }

    /*
     *   Step 1: search for PDR and apply PDR instructions
     */
    __u32 teid = bpf_htonl(ctx->gtp->teid);
    struct pdr_info *session = bpf_map_lookup_elem(&pdr_map_uplink_ip4, &teid);
    if (!session) {
        upf_printk("upf: [n3] no session for teid:%u", teid);
        return DEFAULT_XDP_ACTION;
    }

    upf_printk("upf: [n3] teid:%u trace:%d", teid, session->trace_flag);
    if(session->trace_flag)
        trace_packet(ctx, PACKET_DIRECTION_IN);

    // Set defaults
    const struct pdr *pdr = &session->default_pdr;
    if (session->sdf_mode) {
        const struct pdr *pdr_sdf = check_sdf_filters_gtp(ctx, session, teid);
        if(pdr_sdf) {
            upf_printk("upf: [n3] sdf filter matches teid:%u", teid);
            pdr = pdr_sdf;
        }
    }

    /*
     *   Step 2: search for FAR and apply FAR instructions
     */
    struct far_info *far = bpf_map_lookup_elem(&far_map, &pdr->far_id);
    if (!far) {
        upf_printk("upf: [n3] no session far for teid:%u far:%d", teid, pdr->far_id);
        return XDP_DROP;
    }

    upf_printk("upf: [n3] far:%d action:%d outer_header_creation:%d", pdr->far_id, far->action, far->outer_header_creation);

    // Only forwarding action supported at the moment
    if (!(far->action & FAR_FORW))
        return XDP_DROP;

    /*
     *   Step 3: search for QER and apply QER instructions
     */
    struct qer_info *qer = bpf_map_lookup_elem(&qer_map, &pdr->qer_id);
    if (!qer) {
        upf_printk("upf: [n3] no session qer for teid:%u qer:%d", teid, pdr->qer_id);
        return XDP_DROP;
    }

    upf_printk("upf: [n3] qer:%d gate_status:%d mbr:%u", pdr->qer_id, qer->ul_gate_status, qer->ul_maximum_bitrate);

    if (qer->ul_gate_status != GATE_STATUS_OPEN)
        return XDP_DROP;

    const __u64 packet_size = bpf_ntohs(ctx->gtp->message_length);
    //const __u64 packet_size = ctx->xdp_ctx->data_end - ctx->xdp_ctx->data;
    if (XDP_DROP == limit_rate_sliding_window(packet_size, &qer->ul_start, qer->ul_maximum_bitrate))
        return XDP_DROP;

    const int urr_size = sizeof(pdr->urr_id)/sizeof(pdr->urr_id[0]);
    for (int i = 0; i < urr_size; i++) {
        update_urr(pdr->urr_id[i], packet_size, 0);
    }

    upf_printk("upf: [n3] session for teid:%u far:%d outer_header_removal:%d", teid, pdr->far_id, pdr->outer_header_removal);

    __u8 tos = qer->dscp ? qer->dscp : far->transport_level_marking >> 8;
    // N9: Only outer header GTP/UDP/IPv4 is supported at the moment
    if (far->outer_header_creation & OHC_GTP_U_UDP_IPv4)
    {
        upf_printk("upf: [n3] session for teid:%u -> %u remote:%pI4", teid, far->teid, &far->remoteip);
        update_gtp_tunnel(ctx, far->localip, far->remoteip, 0, far->teid);
    } else if (pdr->outer_header_removal == OHR_GTP_U_UDP_IPv4) {
        long result = remove_gtp_header(ctx);
        if (result) {
            upf_printk("upf: [n3] handle_gtp_packet: can't remove gtp header: %d", result);
            return XDP_ABORTED;
        }
    }

    /*
     * Decrement IP TTL and reply TTL exeeded message (debug purspose only)
     */
    // if(ctx->ip4 && ctx->ip4->ttl < 2)
    // {
    //     if (-1 == add_icmp_over_ip4_headers(ctx, far->localip, ctx->ip4->saddr))
    //         return XDP_ABORTED;

    //     upf_printk("upf: send icmp ttl exeeded %pI4 -> %pI4", &ctx->ip4->saddr, &ctx->ip4->daddr);
    //     return handle_n6_packet_ipv4(ctx);
    // }

    /*
     * Reply to ping requests (debug purspose only)
     */
    if(ctx->ip4 && ctx->ip4->daddr == far->localip && ctx->ip4->protocol == IPPROTO_ICMP)
    {
        upf_printk("upf: [n3] prepare icmp ping reply to request %pI4 -> %pI4", &ctx->ip4->saddr, &ctx->ip4->daddr);
        if (-1 == prepare_icmp_echo_reply(ctx, far->localip, ctx->ip4->saddr))
            return XDP_ABORTED;

        upf_printk("upf: [n3] send icmp ping reply %pI4 -> %pI4", &ctx->ip4->saddr, &ctx->ip4->daddr);
        return handle_n6_packet_ipv4(ctx);
    }

    /*
     *   Step 4: Route packet finally
     */
     enum xdp_action action = XDP_ABORTED;
    if (ctx->ip4) {
        increment_counter(ctx->n3_n6_counter, tx_n6);
        update_tos_ipv4(ctx->ip4, tos);
        action = route_ipv4(ctx->xdp_ctx, ctx->eth, ctx->ip4);
    } else if (ctx->ip6) {
        increment_counter(ctx->n3_n6_counter, tx_n6);
        update_tos_ipv6(ctx->ip6, tos);
        action = route_ipv6(ctx->xdp_ctx, ctx->eth, ctx->ip6);
    }

    if(session->trace_flag && (action == XDP_TX || action == XDP_REDIRECT))
        trace_packet(ctx, PACKET_DIRECTION_OUT);

    return action;

}

static __always_inline enum xdp_action handle_gtpu(struct packet_context *ctx) {
    int pdu_type = parse_gtp(ctx);
    switch (pdu_type) {
        case GTPU_G_PDU:
            increment_counter(ctx->counters, rx_gtp_pdu);
            return handle_gtp_packet(ctx);
        case GTPU_ECHO_REQUEST:
            increment_counter(ctx->counters, rx_gtp_echo);
            // upf_printk("upf: gtp header [ version=%d, pt=%d, e=%d]", gtp->version, gtp->pt, gtp->e);
            // upf_printk("upf: gtp echo request [ type=%d ]", pdu_type);
            upf_printk("upf: gtp echo request [ %pI4 -> %pI4 ]", &ctx->ip4->saddr, &ctx->ip4->daddr);
            return handle_echo_request(ctx);
        case GTPU_ECHO_RESPONSE:
            return XDP_PASS; //Pass echo response to userspace program
        case GTPU_ERROR_INDICATION:
        case GTPU_SUPPORTED_EXTENSION_HEADERS_NOTIFICATION:
        case GTPU_END_MARKER:
            increment_counter(ctx->counters, rx_gtp_other);
            return DEFAULT_XDP_ACTION;
        default:
            increment_counter(ctx->counters, rx_gtp_unexp);
            upf_printk("upf: unexpected gtp message: type=%d", pdu_type);
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
            if (GTP_UDP_PORT == parse_udp(ctx)  
                && (ctx->ip4->daddr == global_config.n3_ipv4_address 
                    || ctx->ip4->daddr == global_config.n9_ipv4_address)) {
                upf_printk("upf: gtp-u received");
                increment_counter(ctx->n3_n6_counter, rx_n3);
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

    increment_counter(ctx->n3_n6_counter, rx_n6);
    return handle_n6_packet_ipv4(ctx);
}

static __always_inline enum xdp_action handle_ip6(struct packet_context *ctx) {
    int l4_protocol = parse_ip6(ctx);
    switch (l4_protocol) {
        case IPPROTO_ICMPV6:  // Let kernel stack take care
            upf_printk("upf: icmp received. passing to kernel");
            increment_counter(ctx->counters, rx_icmp6);
            return XDP_PASS;
        case IPPROTO_UDP:
            increment_counter(ctx->counters, rx_udp);
            // Don't expect GTP over IPv6 at the moment
            // if (GTP_UDP_PORT == parse_udp(ctx))
            // {
            //     upf_printk("upf: gtp-u received");
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
    increment_counter(ctx->n3_n6_counter, rx_n6);
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
            upf_printk("upf: arp received. passing to kernel");
            return XDP_PASS;
        }
    }

    return DEFAULT_XDP_ACTION;
}

// Combined N3 & N6 entrypoint. Use for "on-a-stick" interfaces
SEC("xdp/upf_ip_entrypoint")
int upf_ip_entrypoint_func(struct xdp_md *ctx) {
    // upf_printk("upf n3 & n6 combined entrypoint start");
    upf_printk("upf: n3ip:%pI4 n9ip:%pI4", &global_config.n3_ipv4_address, &global_config.n9_ipv4_address);

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
        .counters = &statistic->upf_counters,
        .n3_n6_counter = &statistic->upf_n3_n6_counter};

    enum xdp_action action = process_packet(&context);
    statistic->xdp_actions[action & EUPF_MAX_XDP_ACTION_MASK] += 1;

//#define PACKET_TRACE
#ifdef PACKET_TRACE
    if(action != XDP_TX && action != XDP_REDIRECT) // write all packets
        trace_packet(&context, PACKET_DIRECTION_BLOCKED);
#endif

    return action;
}

char _license[] SEC("license") = "GPL";