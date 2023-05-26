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

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#include "xdp/program_array.h"

struct bucket {
    volatile __u64 t_next;
    __u64 upper_limit_bps;
};

static __always_inline int limit_rate_ok(struct xdp_md *ctx, struct bucket *bucket) {
    static const __u64 DROP_HORIZON = 1000000000ULL;
    static const __u64 BURST = 5000000ULL;
    static const __u64 NSEC_PER_SEC = 1000000000ULL;

    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    __u64 now = bpf_ktime_get_ns();
    __u64 t_next = bucket->t_next;
    __u64 upper_limit_bps = bucket->upper_limit_bps;
    // skb->tstamp = max(now - BURST, t_next);
    __u64 ts = max(now - BURST, t_next);

    if (t_next - now > DROP_HORIZON)
        return XDP_DROP;
    // t_next = skb->tstamp + skb->wire_len * NSEC_PER_SEC / upper_limit_bps;
    t_next = ts + (data_end - data) * NSEC_PER_SEC / upper_limit_bps;
    return XDP_PASS;
}

SEC("xdp/upf_qer_program")
int upf_qer_program_func(struct xdp_md *ctx) {
    bpf_printk("upf_qer_program start\n");

    bpf_printk("tail call to UPF_PROG_TYPE_FAR key\n");
    bpf_tail_call(ctx, &upf_pipeline, UPF_PROG_TYPE_FAR);
    bpf_printk("tail call to UPF_PROG_TYPE_FAR key failed\n");
    return XDP_ABORTED;
}

char _license[] SEC("license") = "GPL";