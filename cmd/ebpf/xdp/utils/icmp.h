// Copyright 2023 Edgecom LLC
// 
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#pragma once

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/types.h>
#include <linux/icmp.h>
#include <linux/icmpv6.h>

#include "xdp/utils/csum.h"
#include "xdp/utils/parsers.h"
#include "xdp/utils/packet_context.h"

static __always_inline void fill_icmp_header(struct icmphdr *icmp) {
    icmp->type = ICMP_TIME_EXCEEDED;
    icmp->code = ICMP_EXC_TTL;
    icmp->un.gateway = 0;
    icmp->checksum = 0;
}

// static __always_inline __u32 add_icmp_over_ip4_headers(struct packet_context *ctx, int saddr, int daddr) {
//     static const size_t icmp_encap_size = sizeof(struct iphdr) + sizeof(struct icmphdr);

//     if (!ctx->ip4)
//         return -1;

//     const __u32 ip_packet_len = bpf_ntohs(ctx->ip4->tot_len);

//     int result = bpf_xdp_adjust_head(ctx->xdp_ctx, (__s32)-icmp_encap_size);
//     if (result)
//         return -1;

//     char *data = (char *)(long)ctx->xdp_ctx->data;
//     const char *data_end = (const char *)(long)ctx->xdp_ctx->data_end;

//     struct ethhdr *orig_eth = (struct ethhdr *)(data + icmp_encap_size);
//     if ((const char *)(orig_eth + 1) > data_end)
//         return -1;

//     struct ethhdr *eth = (struct ethhdr *)data;
//     __builtin_memcpy(eth, orig_eth, sizeof(*eth));
//     eth->h_proto = bpf_htons(ETH_P_IP);

//     struct iphdr *ip = (struct iphdr *)(eth + 1);
//     if ((const char *)(ip + 1) > data_end)
//         return -1;

//     /* Add the outer IP header */
//     fill_ip_header(ip, saddr, daddr, 0, ip_packet_len + icmp_encap_size);
//     ip->protocol = IPPROTO_ICMP;
//     ip->check = ipv4_csum(ip, sizeof(*ip));

//     /* Add the ICMP header */
//     struct icmphdr *icmp = (struct icmphdr *)(ip + 1);
//     if ((const char *)(icmp + 1) > data_end)
//         return -1;

//     fill_icmp_header(icmp);
//     const __s8 icmp_payload_size = data_end - (const char *)icmp;
//     icmp->checksum = ipv4_csum(icmp, icmp_payload_size);

//     /* Update packet pointers */
//     context_set_ip4(ctx, (char *)(long)ctx->xdp_ctx->data, (const char *)(long)ctx->xdp_ctx->data_end, eth, ip, 0, 0);
//     return 0;
// }

#define ICMPV6_ROUTER_SOLICITATION    	133
#define ICMPV6_ROUTER_ADVERTISEMENT    	134

struct icmp6hdr_ra_ {
    __u8	icmp6_type;
    __u8	icmp6_code;
    __sum16	icmp6_cksum;

    union {
        __be32	un_data32[1];
        __be16	un_data16[2];
        __u8	un_data8[4];

        struct icmpv6_nd_ra_ {
        __u8    hop_limit;
        #if defined(__LITTLE_ENDIAN_BITFIELD)
        __u8	reserved:3,
                router_pref:2,
                home_agent:1,
                other:1,
                managed:1;

#elif defined(__BIG_ENDIAN_BITFIELD)
        __u8	managed:1,
                other:1,
                home_agent:1,
                router_pref:2,
                reserved:3;
#else
#error	"Please fix <asm/byteorder.h>"
#endif
        __be16	rt_lifetime;
        __be32	rt_reachabletime;
        __be32	rt_retranstimer;
        } u_nd_ra;
    } icmp6_dataun;
};

struct icmpv6_option_prefix {
    __u8    type;
    __u8    length;
    __u8    prefix_length;
    __u8    reserved:5,
            router_addr_flag:1,
            auto_flag:1,
            onlink_flag:1;
    __be32 valid_lifetime;
    __be32 pref_lifetime;
    __be32 _reserved;
    struct in6_addr	prefix;
};

static __always_inline __u32 prepare_icmp_echo_reply(struct packet_context *ctx, int saddr, int daddr) {
    if (!ctx->ip4)
        return -1;

    struct ethhdr *eth = ctx->eth;
    swap_mac(eth);

    const char *data_end = (const char *)(long)ctx->xdp_ctx->data_end;
    struct iphdr *ip = ctx->ip4;
    if ((const char *)(ip + 1) > data_end)
        return -1;

    swap_ip(ip);

    struct icmphdr *icmp = (struct icmphdr *)(ip + 1);
    if ((const char *)(icmp + 1) > data_end)
        return -1;

    if(icmp->type != ICMP_ECHO)
        return -1;

    __u16 old = *(__u16*)&icmp->type;
    icmp->type = ICMP_ECHOREPLY;
    icmp->code = 0;
    
    ipv4_csum_replace(&icmp->checksum, old, *(__u16*)&icmp->type);

    return 0;
}

static __always_inline __u32 prepare_icmp6_ra(struct packet_context *ctx, const struct in6_addr *prefix, __u8 prefix_length) {
    if (!ctx->ip6)
        return -1;

    char *data = (char *)(long)ctx->xdp_ctx->data;
    const char *data_end = (const char *)(long)ctx->xdp_ctx->data_end;
    const size_t current = data_end - data;
    const size_t requred = sizeof(struct ethhdr) 
                        + sizeof(struct ipv6hdr) 
                        + sizeof(struct icmp6hdr_ra_) 
                        + sizeof(struct icmpv6_option_prefix);
    if(current < requred) {
        long result = bpf_xdp_adjust_tail(ctx->xdp_ctx, requred-current);
        if (result)
            return -1;
    }

    data = (char *)(long)ctx->xdp_ctx->data;
    data_end = (const char *)(long)ctx->xdp_ctx->data_end;
    if( context_reinit(ctx, data, data_end))
        return -1;
    
    struct ethhdr *eth = ctx->eth;
    swap_mac(eth);

    data_end = (const char *)(long)ctx->xdp_ctx->data_end;
    if (!ctx->ip6)
        return -1;

    struct ipv6hdr *ip6 = ctx->ip6;
    if ((const char *)(ip6 + 1) > data_end)
        return -1;
    
    __builtin_memcpy(ip6->daddr.in6_u.u6_addr8, ip6->saddr.in6_u.u6_addr8, sizeof(ip6->saddr.in6_u.u6_addr8));
    ip6->daddr.in6_u.u6_addr32[1] = prefix->in6_u.u6_addr32[1]; //0x0905a029;
    ip6->daddr.in6_u.u6_addr32[0] = prefix->in6_u.u6_addr32[0]; //0x00d0032a;

    ip6->saddr.in6_u.u6_addr32[3] = 0x01000000;
    ip6->saddr.in6_u.u6_addr32[2] = 0;
    ip6->saddr.in6_u.u6_addr32[1] = 0;
    ip6->saddr.in6_u.u6_addr32[0] = 0x000080fe;
    
    ip6->payload_len = bpf_ntohs(sizeof(struct icmp6hdr_ra_) + sizeof(struct icmpv6_option_prefix));

    struct icmp6hdr_ra_ *icmp6 = (struct icmp6hdr_ra_ *)(ip6 + 1);
    if ((const char *)(icmp6 + 1) > data_end)
        return -1;

    icmp6->icmp6_type   = ICMPV6_ROUTER_ADVERTISEMENT;
    icmp6->icmp6_code   = 0;
    icmp6->icmp6_cksum  = 0;

    icmp6->icmp6_dataun.u_nd_ra.hop_limit           = 64;
    icmp6->icmp6_dataun.u_nd_ra.rt_lifetime         = bpf_ntohs(64800);
    icmp6->icmp6_dataun.u_nd_ra.rt_reachabletime    = 0;
    icmp6->icmp6_dataun.u_nd_ra.rt_retranstimer     = 0;

    struct icmpv6_option_prefix *option_prefix = (struct icmpv6_option_prefix *)(icmp6 + 1);
    if ((const char *)(option_prefix + 1) > data_end)
        return -1;
    
    option_prefix->type             = 3;
    option_prefix->length           = 4;
    option_prefix->prefix_length    = prefix_length;
    option_prefix->onlink_flag      = 1;
    option_prefix->auto_flag        = 1;
    option_prefix->router_addr_flag = 0;
    option_prefix->valid_lifetime   = 0xffffffff;
    option_prefix->pref_lifetime    = 0xffffffff;
    option_prefix->prefix.in6_u.u6_addr32[0] = prefix->in6_u.u6_addr32[0];
    option_prefix->prefix.in6_u.u6_addr32[1] = prefix->in6_u.u6_addr32[1];
    
    icmp6->icmp6_cksum = icmp6_csum(ip6, icmp6->icmp6_code, icmp6->icmp6_type, icmp6, sizeof(struct icmp6hdr_ra_) + sizeof(struct icmpv6_option_prefix));
    return 0;
}