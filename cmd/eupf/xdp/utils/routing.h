#pragma once

#include <linux/bpf.h>
#include <linux/types.h>
#include <linux/in.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <sys/socket.h>

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

static __always_inline __u32 route_ipv4(struct xdp_md *ctx, struct ethhdr *eth, const struct iphdr *ip4)
{
    struct bpf_fib_lookup fib_params = {};
    fib_params.family = AF_INET;
    fib_params.tos = ip4->tos;
    fib_params.l4_protocol = ip4->protocol;
    fib_params.sport = 0;
    fib_params.dport = 0;
    fib_params.tot_len = bpf_ntohs(ip4->tot_len);
    fib_params.ipv4_src = ip4->saddr;
    fib_params.ipv4_dst = ip4->daddr;
    fib_params.ifindex = ctx->ingress_ifindex;

    int rc = bpf_fib_lookup(ctx, &fib_params, sizeof(fib_params), 0 /*BPF_FIB_LOOKUP_OUTPUT*/);
    switch (rc)
    {
    case BPF_FIB_LKUP_RET_SUCCESS:
        bpf_printk("upf: bpf_fib_lookup %pI4 -> %pI4: nexthop: %pI4", &ip4->saddr, &ip4->daddr, &fib_params.ipv4_dst);
        //_decr_ttl(ether_proto, l3hdr);
        __builtin_memcpy(eth->h_dest, fib_params.dmac, ETH_ALEN);
        __builtin_memcpy(eth->h_source, fib_params.smac, ETH_ALEN);
        bpf_printk("upf: bpf_redirect: if=%d %lu -> %lu", fib_params.ifindex, fib_params.smac, fib_params.dmac);
        return bpf_redirect(fib_params.ifindex, 0);
        // return XDP_TX;
        // return bpf_redirect_map(&if_redirect, fib_params.ifindex, 0);
    case BPF_FIB_LKUP_RET_BLACKHOLE:
    case BPF_FIB_LKUP_RET_UNREACHABLE:
    case BPF_FIB_LKUP_RET_PROHIBIT:
        bpf_printk("upf: bpf_fib_lookup %pI4 -> %pI4: %d", &ip4->saddr, &ip4->daddr, rc);
        return XDP_DROP;
    case BPF_FIB_LKUP_RET_NOT_FWDED:
    case BPF_FIB_LKUP_RET_FWD_DISABLED:
    case BPF_FIB_LKUP_RET_UNSUPP_LWT:
    case BPF_FIB_LKUP_RET_NO_NEIGH:
    case BPF_FIB_LKUP_RET_FRAG_NEEDED:
    default:
        bpf_printk("upf: bpf_fib_lookup %pI4 -> %pI4: %d", &ip4->saddr, &ip4->daddr, rc);
        return XDP_PASS; // Let's kernel takes care
    }
}

static __always_inline __u32 route_ipv6(struct xdp_md *ctx, struct ethhdr *eth, const struct ipv6hdr *ip6)
{
    struct bpf_fib_lookup fib_params = {};
    fib_params.family = AF_INET;
    // fib_params.tos = ip6->flow_lbl;
    fib_params.l4_protocol = ip6->nexthdr;
    fib_params.sport = 0;
    fib_params.dport = 0;
    fib_params.tot_len = bpf_ntohs(ip6->payload_len);
    __builtin_memcpy(fib_params.ipv6_src, &ip6->saddr, sizeof(ip6->saddr));
    __builtin_memcpy(fib_params.ipv6_dst, &ip6->daddr, sizeof(ip6->daddr));
    fib_params.ifindex = ctx->ingress_ifindex;

    int rc = bpf_fib_lookup(ctx, &fib_params, sizeof(fib_params), 0 /*BPF_FIB_LOOKUP_OUTPUT*/);
    switch (rc)
    {
    case BPF_FIB_LKUP_RET_SUCCESS:
        bpf_printk("upf: bpf_fib_lookup %pI6c -> %pI6c: nexthop: %pI4", &ip6->saddr, &ip6->daddr, &fib_params.ipv4_dst);
        //_decr_ttl(ether_proto, l3hdr);
        __builtin_memcpy(eth->h_dest, fib_params.dmac, ETH_ALEN);
        __builtin_memcpy(eth->h_source, fib_params.smac, ETH_ALEN);
        bpf_printk("upf: bpf_redirect: if=%d %lu -> %lu", fib_params.ifindex, fib_params.smac, fib_params.dmac);
        return bpf_redirect(fib_params.ifindex, 0);
        // return XDP_TX;
        // return bpf_redirect_map(&if_redirect, fib_params.ifindex, 0);
    case BPF_FIB_LKUP_RET_BLACKHOLE:
    case BPF_FIB_LKUP_RET_UNREACHABLE:
    case BPF_FIB_LKUP_RET_PROHIBIT:
        bpf_printk("upf: bpf_fib_lookup %pI6c -> %pI6c: %d", &ip6->saddr, &ip6->daddr, rc);
        return XDP_DROP;
    case BPF_FIB_LKUP_RET_NOT_FWDED:
    case BPF_FIB_LKUP_RET_FWD_DISABLED:
    case BPF_FIB_LKUP_RET_UNSUPP_LWT:
    case BPF_FIB_LKUP_RET_NO_NEIGH:
    case BPF_FIB_LKUP_RET_FRAG_NEEDED:
    default:
        bpf_printk("upf: bpf_fib_lookup %pI6c -> %pI6c: %d", &ip6->saddr, &ip6->daddr, rc);
        return XDP_PASS; // Let's kernel takes care
    }
}