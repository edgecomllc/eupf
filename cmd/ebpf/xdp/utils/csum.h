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

#include <bpf/bpf_helpers.h>
#include <linux/bpf.h>
#include <linux/types.h>

#ifndef __packed
# define __packed		__attribute__((packed))
#endif

/* The IPv6 pseudo-header */
struct ipv6_pseudo_header {
    union {
        struct header {
            struct in6_addr src_ip;
            struct in6_addr dst_ip;
            __be32 length;
            __u8 zero[3];
            __u8 next_header;
        } __packed fields;
        __u16 words[20];
    };
};

static __always_inline __u16 csum_fold_helper(__u64 csum) {
	#pragma unroll
    for (int i = 0; i < 4; i++) {
            csum = (csum & 0xffff) + (csum >> 16);
    }

    return ~csum;
}

static __always_inline __u64 ipv4_csum(void *data_start, __u32 data_size) {
    __u64 csum = bpf_csum_diff(0, 0, data_start, data_size, 0);
    return csum_fold_helper(csum);
}

static __always_inline __u64 icmp6_csum(struct ipv6hdr *ip6, __u8 icmp6_code, __u8 icmp6_type, void *data_start, __u32 data_size) {
    /* Calculate the unfolded checksum of the ICMPv6 sample */
	__u64 csum = bpf_csum_diff(0, 0, data_start, data_size, 0);

	/* Fill pseudo header */
    struct ipv6_pseudo_header pseudo_header = {0};
	__builtin_memcpy(&pseudo_header.fields.src_ip, &ip6->saddr, sizeof(struct in6_addr));
	__builtin_memcpy(&pseudo_header.fields.dst_ip, &ip6->daddr, sizeof(struct in6_addr));
	pseudo_header.fields.length = bpf_htonl(data_size);
	pseudo_header.fields.next_header = IPPROTO_ICMPV6;

	#pragma unroll
	for (int i = 0; i < (int)(sizeof(pseudo_header.words) / sizeof(__u16)); i++)
		csum += pseudo_header.words[i];

    return csum_fold_helper(csum);
}

static __always_inline void ipv4_csum_replace(__u16 *sum, __u16 old, __u16 new)
{
	__u16 csum = ~*sum;
	csum += ~old;
	csum += csum < (__u16)~old;
	csum += new;
	csum += csum < (__u16)new;
	*sum = ~csum;
}