//
// Created by pirog-spb on 14.12.2022.
//

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#include "xdp/program_array.h"

SEC("upf_gtp_entrypoint")
int upf_gtp_entrypoint_func(struct xdp_md *ctx)
{
    bpf_printk("upf_gtp_entrypoint start\n");

    bpf_printk("tail call to UPF_PROG_TYPE_MAIN key\n");
    bpf_tail_call(ctx, &upf_pipeline, UPF_PROG_TYPE_MAIN);
    bpf_printk("tail call to UPF_PROG_TYPE_MAIN key failed\n");
    return XDP_ABORTED;
}

char _license[] SEC("license") = "GPL";