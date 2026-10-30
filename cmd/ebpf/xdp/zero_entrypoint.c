// SPDX-License-Identifier: Apache-2.0
// Copyright 2023-2026 Edgecom LLC
#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

/* Zero entrypoint purposed for enabling native attach mode for veth */
SEC("xdp/upf_zero_entrypoint")
int upf_n3_entrypoint_func(struct xdp_md *ctx) {
    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";
