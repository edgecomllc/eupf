//
// Created by pirog-spb on 14.12.2022.
//

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

SEC("upf_ip_entrypoint")
int upf_ip_entrypoint_func(struct xdp_md *ctx)
{
    bpf_printk("upf_ip_entrypoint start\n");

    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";