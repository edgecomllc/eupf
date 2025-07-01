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

#pragma once

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#include "xdp/utils/packet_context.h"

enum packet_direction {
    PACKET_DIRECTION_IN = 0,
    PACKET_DIRECTION_OUT = 1,
    PACKET_DIRECTION_BLOCKED = 2,
};

#define min(x, y) ((x) < (y) ? (x) : (y))
//#define MAX_CPUS 4096
#define SAMPLE_SIZE 1024ul
/* Metadata will be in the perf event before the packet data. */
struct packet_trace_metadata {
	__u16 cookie;
	__u16 pkt_len;
    __u32 iface;
    //__u8 packet[1024];
    //__u16 direction;
} __attribute__((packed));

struct {
	__uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
	__uint(key_size, sizeof(int));
	__uint(value_size, sizeof(__u32));
	//__uint(max_entries, MAX_CPUS);
} trace_map SEC(".maps");

// struct {
// 	__uint(type, BPF_MAP_TYPE_RINGBUF);
// 	__uint(max_entries, 256 * 1024);
// } trace_map SEC(".maps");

static __always_inline void trace_packet(struct packet_context *packet_ctx, __u16 direction) 
{
    struct xdp_md *ctx = packet_ctx->xdp_ctx;
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;
    __u16 sample_size = (__u16)(data_end - data);

    // struct packet_trace_metadata* meta = bpf_ringbuf_reserve(&trace_map, sizeof(struct packet_trace_metadata), 0);
	// if (!meta)
 	// 	return;

    // meta->cookie = 0xdead;
    // meta->pkt_len = min(sample_size, SAMPLE_SIZE);
    // meta->iface = direction;

    // if ((const char*)data + 64 > (const char *)data_end) {
    //     bpf_ringbuf_discard(meta, 0);
    //     return;
    // }

    //__builtin_memcpy(meta->packet, data, 64);

    struct packet_trace_metadata meta = {
        .cookie = 0xdead,
        .pkt_len = min(sample_size, SAMPLE_SIZE),
        .iface = direction,
    };
    
    __u64 flags = BPF_F_CURRENT_CPU;
    flags |= (__u64)sample_size << 32;
    int ret = bpf_perf_event_output(ctx, &trace_map, flags, &meta, sizeof(meta));
    //int ret = bpf_ringbuf_output(&trace_map, &meta, sizeof(meta), 0);
    //bpf_ringbuf_submit(meta, 0);
    if (ret)
        bpf_printk("perf_event_output failed: %d\n", ret);

}