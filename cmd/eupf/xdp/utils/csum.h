#pragma once

#include <linux/bpf.h>
#include <linux/types.h>

#include <bpf/bpf_helpers.h>

/* calculate ip header checksum */
// next_iph_u16 = (__u16 *)&iph_tnl;
// #pragma clang loop unroll(full)
// for (int i = 0; i < (int)sizeof(*iph) >> 1; i++)
//	csum += *next_iph_u16++;
// iph_tnl.check = ~((csum & 0xffff) + (csum >> 16));

static __always_inline __u16 csum_fold_helper(__u64 csum)
{
    int i;
#pragma unroll
    for (i = 0; i < 4; i++)
    {
        if (csum >> 16)
            csum = (csum & 0xffff) + (csum >> 16);
    }
    return ~csum;
}

static __always_inline void ipv4_csum(void *data_start, int data_size, __u64 *csum)
{
    *csum = bpf_csum_diff(0, 0, data_start, data_size, *csum);
    *csum = csum_fold_helper(*csum);
}

// static __always_inline void ipv4_l4_csum(void* data_start, int data_size, __u64* csum, struct iphdr* iph) {
//   __u32 tmp = 0;
//   *csum = bpf_csum_diff(0, 0, &iph->saddr, sizeof(__be32), *csum);
//   *csum = bpf_csum_diff(0, 0, &iph->daddr, sizeof(__be32), *csum);
//   tmp = __builtin_bswap32((__u32)(iph->protocol));
//   *csum = bpf_csum_diff(0, 0, &tmp, sizeof(__u32), *csum);
//   tmp = __builtin_bswap32((__u32)(data_size));
//   *csum = bpf_csum_diff(0, 0, &tmp, sizeof(__u32), *csum);
//   *csum = bpf_csum_diff(0, 0, data_start, data_size, *csum);
//   *csum = csum_fold_helper(*csum);
// }
