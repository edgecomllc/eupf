//
// Created by pirog-spb on 14.12.2022.
//

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#include "xdp/program_array.h"

// N3 only entrypoint. Attach to N3 interfaces only
SEC("xdp/upf_n3_entrypoint")
int upf_n3_entrypoint_func(struct xdp_md *ctx)
{
    bpf_printk("upf n3 entrypoint start\n");
    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";