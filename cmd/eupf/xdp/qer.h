#pragma once

#include <bpf/bpf_helpers.h>
#include <linux/bpf.h>

enum gate_status {
    GATE_STATUS_OPEN        = 0,
    GATE_STATUS_CLOSED      = 1,
    GATE_STATUS_RESERVED1   = 2,
    GATE_STATUS_RESERVED2   = 3,
};

struct qer_info
{
    __u8    ul_gate_status;
    __u8    dl_gate_status;
    __u8    qfi;
    __u32   ul_maximum_bitrate;
    __u32   dl_maximum_bitrate;
    __u64   ul_start;
    __u64   dl_start;
};

#ifdef __RELEASE
struct bpf_map_def SEC("maps") qer_map = {
    .type = BPF_MAP_TYPE_ARRAY,
    .key_size = sizeof(__u32), // QER ID
    .value_size = sizeof(struct qer_info),
    .max_entries = 1024, // FIXME
};
#else
struct
{
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32); // qer id
    __type(value, struct qer_info);
    __uint(max_entries, 1024);
} qer_map SEC(".maps");
#endif

static __always_inline __u32 limit_rate_sliding_window(struct xdp_md *ctx, __u64 *windows_start, const __u64 rate)
{
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    static const __u64 NSEC_PER_SEC = 1000000000ULL;
    static const __u64 window_size = 5000000ULL;
    __u64 tx_time = (data_end - data) * 8 * NSEC_PER_SEC / rate;
    __u64 now = bpf_ktime_get_ns();

    __u64 start = *(volatile __u64 *)windows_start;
    if (start + tx_time > now)
        return XDP_DROP;

    if (start + window_size < now)
    {
        *(volatile __u64 *)&windows_start = now - window_size + tx_time;
        return XDP_PASS;
    }

    *(volatile __u64 *)&windows_start = start + tx_time;
    //__sync_fetch_and_add(&window->start, tx_time);
    return XDP_PASS;
}