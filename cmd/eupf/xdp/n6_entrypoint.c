//
// Created by pirog-spb on 14.12.2022.
//

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#include "xdp/program_array.h"

// N6 only entrypoint. Attach to N6 interfaces only
SEC("xdp/upf_n6_entrypoint")
int upf_n6_entrypoint_func(struct xdp_md *ctx) {
    bpf_printk("upf n6 entrypoint start\n");
    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";