#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

SEC("xdp/upf_far_program")
int upf_far_program_func(struct xdp_md *ctx)
{
    bpf_printk("upf_far_program start\n");

    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";