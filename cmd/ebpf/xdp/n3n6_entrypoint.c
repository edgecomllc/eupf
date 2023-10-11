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
#include "xdp/pdr.h"

#include "xdp/utils/common.h"
#include "xdp/utils/trace.h"
#include "xdp/utils/packet_context.h"
#include "xdp/utils/parsers.h"
#include "xdp/utils/csum.h"
#include "xdp/utils/gtp_utils.h"
#include "xdp/utils/routing.h"
#include "xdp/utils/icmp.h"


#define DEFAULT_XDP_ACTION XDP_PASS


static __always_inline enum xdp_action send_to_gtp_tunnel(struct packet_context *ctx, int srcip, int dstip, __u8 tos, int teid) {
    if (-1 == add_gtp_over_ip4_headers(ctx, srcip, dstip, tos, teid))
        return XDP_ABORTED;
    upf_printk("upf: send gtp pdu %pI4 -> %pI4", &ctx->ip4->saddr, &ctx->ip4->daddr);
    increment_counter(ctx->n3_n6_counter, tx_n3);
    return route_ipv4(ctx->xdp_ctx, ctx->eth, ctx->ip4);
}

static __always_inline __u8 get_sdf_protocol(__u8 ip_protocol) {
    switch(ip_protocol)
    {
        case IPPROTO_ICMP: return 1;
        case IPPROTO_TCP: return 2;
        case IPPROTO_UDP: return 3;
        default: return 0;
    }
}

static __always_inline __u8 match_sdf_filter_ipv4(struct packet_context *ctx, struct sdf_filter *sdf) {
    // const struct iphdr *ip4 = ctx->ip4;
    struct iphdr* ip4 = parse_ip4_protocol(ctx);   
    if(!ip4)
        return 0;
    __u8 packet_protocol = get_sdf_protocol(ip4->protocol); //TODO: convert protocol in golang part
    __u32 sdf_src_ip = bpf_htonl(sdf->src_addr.ip);
    __u32 sdf_dst_ip = bpf_htonl(sdf->dst_addr.ip);

    __u16 packet_src_port = 0;
    __u16 packet_dst_port = 0;
    switch (ip4->protocol) {
        case IPPROTO_UDP:
            upf_printk("The packet is adhering to the UDP standards");
            struct udphdr *udp_header = parse_udp_src_dst(ctx); 
            if(!udp_header) {
                return 0;
            }
            packet_src_port = bpf_ntohs(udp_header->source); //TODO: convert port in golang part
            packet_dst_port = bpf_ntohs(udp_header->dest); //TODO: convert port in golang part
            break;
        case IPPROTO_TCP:
            upf_printk("The packet is adhering to the TCP standards");
            struct tcphdr *tcp_header = parse_tcp_src_dst(ctx); //FIXME: use ctx
            if(!tcp_header) {
                return 0;
            }
            packet_src_port = bpf_ntohs(tcp_header->source); //TODO: convert port in golang part
            packet_dst_port = bpf_ntohs(tcp_header->dest); //TODO: convert port in golang part
            break;
        default:
            upf_printk("The packet is not adhering to TCP or UDP");
    }

    upf_printk("SDF: filter protocol: %u", sdf->protocol);
    upf_printk("SDF: filter source ip: %pI4, destination ip: %pI4",  &sdf_src_ip,  &sdf_dst_ip);
    upf_printk("SDF: filter source port lower bound: %u, source port upper bound: %u", sdf->src_port.lower_bound, sdf->src_port.upper_bound);
    upf_printk("SDF: filter destination port lower bound: %u, destination port upper bound: %u", sdf->dst_port.lower_bound, sdf->dst_port.upper_bound);
    upf_printk("SDF: filter source address mask: %pI4, destination address mask: %pI4", &sdf->dst_addr.mask, &sdf->dst_addr.mask);

    upf_printk("SDF: packet protocol: %u", ip4->protocol);
    upf_printk("SDF: packet source ip: %pI4, destination ip: %pI4",  &ip4->saddr,  &ip4->daddr);
    upf_printk("SDF: packet source port: %u, destination port: %u", packet_src_port, packet_dst_port);
    
    
    if (sdf->protocol != packet_protocol 
        || ((ip4->saddr & sdf->src_addr.mask) != sdf_src_ip)  
        || ((ip4->daddr & sdf->dst_addr.mask) != sdf_dst_ip) 
        || (packet_src_port < sdf->src_port.lower_bound || packet_src_port > sdf->src_port.upper_bound) 
        || (packet_dst_port < sdf->dst_port.lower_bound || packet_dst_port > sdf->dst_port.upper_bound)) 
    {
        return 0;
    }
    
    upf_printk("Packet with source ip:%pI4, destination ip:%pI4 matches SDF filter",
               &ip4->saddr, &ip4->daddr);

    return 1;
}

static __always_inline __u8 match_sdf_filter_ipv6(struct packet_context *ctx, struct sdf_filter *sdf) {
    // const struct ipv6hdr *ipv6 = ctx->ip6;  
    struct ipv6hdr *ipv6 = parse_ip6_protocol(ctx);  
    if(!ipv6)
        return 0;           
    __u8 packet_protocol = get_sdf_protocol(ipv6->nexthdr);
    struct in6_addr packet_src_ip = ipv6->saddr;
    struct in6_addr packet_dst_ip = ipv6->daddr;

    __uint128_t packet_src_ip_128 = *((__uint128_t*)packet_src_ip.s6_addr);
    __uint128_t packet_dst_ip_128 = *((__uint128_t*)packet_dst_ip.s6_addr);

    __u16 packet_src_port = 0;
    __u16 packet_dst_port = 0;

    __uint128_t sdf_src_ip = bpf_htonl(sdf->src_addr.ip);
    __uint128_t sdf_dst_ip = bpf_htonl(sdf->dst_addr.ip);

    switch (ipv6->nexthdr) {
        case IPPROTO_UDP:
            upf_printk("The packet is adhering to the UDP standards");
            struct udphdr *udp_header = parse_udp_src_dst(ctx);
            if(!udp_header) {
                return 0;
            }
            packet_src_port = bpf_ntohs(udp_header->source);
            packet_dst_port = bpf_ntohs(udp_header->dest);
            break;
        case IPPROTO_TCP:
            upf_printk("The packet is adhering to the TCP standards");
            struct tcphdr *tcp_header = parse_tcp_src_dst(ctx);
            if(!tcp_header) {
                return 0;
            }
            packet_src_port = bpf_ntohs(tcp_header->source);
            packet_dst_port = bpf_ntohs(tcp_header->dest);
            break;
        default:
            upf_printk("The packet is not adhering to TCP or UDP");
    }

    upf_printk("SDF: filter protocol: %u", sdf->protocol);
    upf_printk("SDF: filter source ip: %pI6c, destination ip: %pI6c",  &sdf_src_ip,  &sdf_dst_ip);
    upf_printk("SDF: filter source port lower bound: %u, source port upper bound: %u", sdf->src_port.lower_bound, sdf->src_port.upper_bound);
    upf_printk("SDF: filter destination port lower bound: %u, destination port upper bound: %u", sdf->dst_port.lower_bound, sdf->dst_port.upper_bound);
    upf_printk("SDF: filter source address mask: %pI4, destination address mask: %pI4", &sdf->dst_addr.mask, &sdf->dst_addr.mask);

    upf_printk("SDF: packet protocol: %u", packet_protocol);
    upf_printk("SDF: packet source ip: %pI6c, destination ip: %pI6c",  &packet_src_ip_128,  &packet_dst_ip_128);
    upf_printk("SDF: packet source port: %u, destination port: %u", packet_src_port, packet_dst_port);

    if (sdf->protocol != packet_protocol ||
        (packet_src_ip_128 & sdf->src_addr.mask) != sdf_src_ip || 
        (packet_dst_ip_128 & sdf->dst_addr.mask) != sdf_dst_ip ||
         packet_src_port < sdf->src_port.lower_bound || packet_src_port > sdf->src_port.upper_bound ||
         packet_dst_port < sdf->dst_port.lower_bound || packet_dst_port > sdf->dst_port.upper_bound) {
        return 0;
    }

    upf_printk("SDF: packet with source ip:%pI6c, destination ip:%pI6c matches SDF filter",
               &packet_src_ip_128, &packet_dst_ip_128);

    return 1;
}

static __always_inline __u16 handle_n6_packet_ipv4(struct packet_context *ctx) {
    const struct iphdr *ip4 = ctx->ip4;
    struct pdr_info *pdr = bpf_map_lookup_elem(&pdr_map_downlink_ip4, &ip4->daddr);
    if (!pdr) {
        upf_printk("upf: no downlink session for ip:%pI4", &ip4->daddr);
        return DEFAULT_XDP_ACTION;
    }

    struct sdf_filter *sdf = &pdr->sdf_rules.sdf_filter;
    __u32 far_id = pdr->far_id;
    __u32 qer_id = pdr->qer_id;
    __u8 outer_header_removal = pdr->outer_header_removal;
    if (sdf->protocol >= 0 && sdf->protocol <= 3) {
        __u8 packet_matched = match_sdf_filter_ipv4(ctx, sdf);
        if(packet_matched) {
            upf_printk("Packet with source ip:%pI4 and destination ip:%pI4 matches SDF filter", &ip4->saddr, &ip4->daddr);
            far_id = pdr->sdf_rules.far_id;
            qer_id = pdr->sdf_rules.qer_id;
            outer_header_removal = pdr->sdf_rules.outer_header_removal;
        } 
    }

    struct far_info *far = bpf_map_lookup_elem(&far_map, &far_id);
    if (!far) {
        upf_printk("upf: no downlink session far for ip:%pI4 far:%d", &ip4->daddr, far_id);
        return XDP_DROP;
    }

    upf_printk("upf: downlink session for ip:%pI4  far:%d action:%d", &ip4->daddr, far_id, far->action);

    // Only forwarding action is supported at the moment
    if (!(far->action & FAR_FORW))
        return XDP_DROP;

    // Only outer header GTP/UDP/IPv4 is supported at the moment
    if (!(far->outer_header_creation & OHC_GTP_U_UDP_IPv4))
        return XDP_DROP;

    struct qer_info *qer = bpf_map_lookup_elem(&qer_map, &qer_id);  
    if (!qer) {
        upf_printk("upf: no downlink session qer for ip:%pI4 qer:%d", &ip4->daddr, qer_id);
        return XDP_DROP;
    }

    upf_printk("upf: qer:%d gate_status:%d mbr:%d", qer_id, qer->dl_gate_status, qer->dl_maximum_bitrate);

    if (qer->dl_gate_status != GATE_STATUS_OPEN)
        return XDP_DROP;

    const __u64 packet_size = ctx->xdp_ctx->data_end - ctx->xdp_ctx->data;
    if (XDP_DROP == limit_rate_sliding_window(packet_size, &qer->dl_start, qer->dl_maximum_bitrate))
        return XDP_DROP;

    __u8 tos = far->transport_level_marking >> 8;

    upf_printk("upf: use mapping %pI4 -> TEID:%d", &ip4->daddr, far->teid);
    return send_to_gtp_tunnel(ctx, far->localip, far->remoteip, tos, far->teid);
}

static __always_inline enum xdp_action handle_n6_packet_ipv6(struct packet_context *ctx) {
    const struct ipv6hdr *ip6 = ctx->ip6;
    struct pdr_info *pdr = bpf_map_lookup_elem(&pdr_map_downlink_ip6, &ip6->daddr);
    if (!pdr) {
        upf_printk("upf: no downlink session for ip:%pI6c", &ip6->daddr);
        return DEFAULT_XDP_ACTION;
    }

    struct sdf_filter *sdf = &pdr->sdf_rules.sdf_filter;
    __u32 far_id = pdr->far_id;
    __u32 qer_id = pdr->qer_id;
    __u8 outer_header_removal = pdr->outer_header_removal;
    if (sdf && (sdf->protocol >= 0 && sdf->protocol <= 3)) {
        __u8 packet_matched = match_sdf_filter_ipv6(ctx, sdf);
        if(packet_matched) {
            upf_printk("Packet with source ip:%pI6c and destination ip:%pI6c matches SDF filter", &ip6->saddr, &ip6->daddr);
            far_id = pdr->sdf_rules.far_id;
            qer_id = pdr->sdf_rules.qer_id;
            outer_header_removal = pdr->sdf_rules.outer_header_removal;
        } 
    }

    struct far_info *far = bpf_map_lookup_elem(&far_map, &far_id);
    if (!far) {
        upf_printk("upf: no downlink session far for ip:%pI6c far:%d", &ip6->daddr, far_id);
        return XDP_DROP;
    }

    upf_printk("upf: downlink session for ip:%pI6c far:%d action:%d", &ip6->daddr, far_id, far->action);

    // Only forwarding action supported at the moment
    if (!(far->action & FAR_FORW))
        return XDP_DROP;

    // Only outer header GTP/UDP/IPv4 is supported at the moment
    if (!(far->outer_header_creation & OHC_GTP_U_UDP_IPv4))
        return XDP_DROP;

    struct qer_info *qer = bpf_map_lookup_elem(&qer_map, &qer_id);  
    if (!qer) {
        upf_printk("upf: no downlink session qer for ip:%pI6c qer:%d", &ip6->daddr, qer_id);
        return XDP_DROP;
    }

    upf_printk("upf: qer:%d gate_status:%d mbr:%d", qer_id, qer->dl_gate_status, qer->dl_maximum_bitrate);

    if (qer->dl_gate_status != GATE_STATUS_OPEN)
        return XDP_DROP;

    const __u64 packet_size = ctx->xdp_ctx->data_end - ctx->xdp_ctx->data;
    if (XDP_DROP == limit_rate_sliding_window(packet_size, &qer->dl_start, qer->dl_maximum_bitrate))
        return XDP_DROP;

    __u8 tos = far->transport_level_marking >> 8;

    upf_printk("upf: use mapping %pI6c -> TEID:%d", &ip6->daddr, far->teid);
    return send_to_gtp_tunnel(ctx, far->localip, far->remoteip, tos, far->teid);
}

static __always_inline enum xdp_action handle_gtp_packet(struct packet_context *ctx) {
    if (!ctx->gtp) {
        upf_printk("upf: unexpected packet context. no gtp header");
        return DEFAULT_XDP_ACTION;
    }

    /*
     *   Step 1: search for PDR and apply PDR instructions
     */
    __u32 teid = bpf_htonl(ctx->gtp->teid);
    struct pdr_info *pdr = bpf_map_lookup_elem(&pdr_map_uplink_ip4, &teid);
    if (!pdr) {
        upf_printk("upf: no session for teid:%d", teid);
        return DEFAULT_XDP_ACTION;
    }

    struct sdf_filter *sdf = &pdr->sdf_rules.sdf_filter;
    __u32 far_id = pdr->far_id;
    __u32 qer_id = pdr->qer_id;
    __u8 outer_header_removal = pdr->outer_header_removal;
    if (sdf && (sdf->protocol >= 0 && sdf->protocol <= 3)) {

        struct upf_statistic fake_stat;
        struct packet_context inner_context = {
            .data = (char *)(long)ctx->data,
            .data_end = (const char *)(long)ctx->data_end,
            .xdp_ctx = ctx->xdp_ctx,
            .counters = &fake_stat,
            .n3_n6_counter = &fake_stat
        };

        //1. check ip4 or ip6 (see get_eth_protocol)
        //2. parse ip4/ip6 (with keeping header in context)
        //3. parse tcp/udp (with keeping header in context)

        __u8 packet_matched = match_sdf_filter_ipv4(&inner_context, sdf);
        if(packet_matched) {
            upf_printk("upf: sdf filter matched for teid:%d", teid);
            far_id = pdr->sdf_rules.far_id;
            qer_id = pdr->sdf_rules.qer_id;
            outer_header_removal = pdr->sdf_rules.outer_header_removal;
        } else {
            upf_printk("upf: sdf filter did not match for teid:%d", teid);
        }
    } else {
        upf_printk("upf: no sdf for teid:%d", teid);
    }

    /*
     *   Step 2: search for FAR and apply FAR instructions
     */
    struct far_info *far = bpf_map_lookup_elem(&far_map, &far_id);
    if (!far) {
        upf_printk("upf: no session far for teid:%d far:%d", teid, far_id);
        return XDP_DROP;
    }

    upf_printk("upf: far:%d action:%d outer_header_creation:%d", far_id, far->action, far->outer_header_creation);

    // Only forwarding action supported at the moment
    if (!(far->action & FAR_FORW))
        return XDP_DROP;

    /*
     *   Step 3: search for QER and apply QER instructions
     */
    struct qer_info *qer = bpf_map_lookup_elem(&qer_map, &qer_id);
    if (!qer) {
        upf_printk("upf: no session qer for teid:%d qer:%d", teid, qer_id);
        return XDP_DROP;
    }

    upf_printk("upf: qer:%d gate_status:%d mbr:%d", qer_id, qer->ul_gate_status, qer->ul_maximum_bitrate);

    if (qer->ul_gate_status != GATE_STATUS_OPEN)
        return XDP_DROP;

    const __u64 packet_size = ctx->xdp_ctx->data_end - ctx->xdp_ctx->data;
    if (XDP_DROP == limit_rate_sliding_window(packet_size, &qer->ul_start, qer->ul_maximum_bitrate))
        return XDP_DROP;

    upf_printk("upf: session for teid:%d far:%d outer_header_removal:%d", teid, pdr->far_id, outer_header_removal);

    // N9: Only outer header GTP/UDP/IPv4 is supported at the moment
    if (far->outer_header_creation & OHC_GTP_U_UDP_IPv4)
    {
        upf_printk("upf: session for teid:%d -> %d remote:%pI4", teid, far->teid, &far->remoteip);
        update_gtp_tunnel(ctx, far->localip, far->remoteip, 0, far->teid);
    } else if (outer_header_removal == OHR_GTP_U_UDP_IPv4) {
        long result = remove_gtp_header(ctx);
        if (result) {
            upf_printk("upf: handle_gtp_packet: can't remove gtp header: %d", result);
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
        upf_printk("upf: prepare icmp ping reply to request %pI4 -> %pI4", &ctx->ip4->saddr, &ctx->ip4->daddr);
        if (-1 == prepare_icmp_echo_reply(ctx, far->localip, ctx->ip4->saddr))
            return XDP_ABORTED;

        upf_printk("upf: send icmp ping reply %pI4 -> %pI4", &ctx->ip4->saddr, &ctx->ip4->daddr);
        return handle_n6_packet_ipv4(ctx);
    }

    /*
     *   Step 4: Route packet finally
     */
    if (ctx->ip4) {
        increment_counter(ctx->n3_n6_counter, tx_n6);
        return route_ipv4(ctx->xdp_ctx, ctx->eth, ctx->ip4);
    } else if (ctx->ip6) {
        increment_counter(ctx->n3_n6_counter, tx_n6);
        return route_ipv6(ctx->xdp_ctx, ctx->eth, ctx->ip6);
    } else {
        return XDP_ABORTED;
    }
        
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
            if (GTP_UDP_PORT == parse_udp(ctx)) {
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

    return action;
}

char _license[] SEC("license") = "GPL";