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

#include <bpf/bpf_endian.h>
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/types.h>
#include <linux/udp.h>

#include "xdp/utils/packet_context.h"

static __always_inline int parse_ethernet(struct packet_context *ctx) {
    struct ethhdr *eth = (struct ethhdr *)ctx->data;
    if ((void *)(eth + 1) > ctx->data_end)
        return -1;

    /* TODO: Add vlan support */

    ctx->data += sizeof(*eth);
    ctx->eth = eth;
    return bpf_ntohs(eth->h_proto);
}

/* 0x3FFF mask to check for fragment offset field */
#define IP_FRAGMENTED 65343

static __always_inline int parse_ip4(struct packet_context *ctx) {
    struct iphdr *ip4 = (struct iphdr *)ctx->data;
    if ((void *)(ip4 + 1) > ctx->data_end)
        return -1;

    /* do not support fragmented packets as L4 headers may be missing */
    // if (ip4->frag_off & IP_FRAGMENTED)
    //	return -1;

    ctx->data += ip4->ihl*4; /* header + options */
    ctx->ip4 = ip4;
    return ip4->protocol;
}

static __always_inline int parse_ip6(struct packet_context *ctx) {
    struct ipv6hdr *ip6 = (struct ipv6hdr *)ctx->data;
    if ((void *)(ip6 + 1) > ctx->data_end)
        return -1;

    /* TODO: Add extention headers support */

    ctx->data += sizeof(*ip6);
    ctx->ip6 = ip6;
    return ip6->nexthdr;
}

static __always_inline int parse_udp(struct packet_context *ctx) {
    struct udphdr *udp = (struct udphdr *)ctx->data;
    if ((void *)(udp + 1) > ctx->data_end)
        return -1;

    ctx->data += sizeof(*udp);
    ctx->udp = udp;
    return bpf_ntohs(udp->dest);
}
