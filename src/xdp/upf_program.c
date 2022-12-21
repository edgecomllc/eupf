#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#include "xdp/program_array.h"

SEC("upf_program")
int upf_program_func(struct xdp_md *ctx)
{
    bpf_printk("upf_program start\n");

    bpf_printk("tail call to UPF_PROG_TYPE_QER key\n");
    bpf_tail_call(ctx, &upf_program_array, UPF_PROG_TYPE_QER);
    bpf_printk("tail call to UPF_PROG_TYPE_QER key failed\n");
    return XDP_ABORTED;
}

char _license[] SEC("license") = "GPL";