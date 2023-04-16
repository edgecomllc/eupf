#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#include "xdp/program_array.h"

//#define DROP_HORIZON 1000000000ULL // 1 секунда
//#define BURST 5000000ULL		   // 5 мс

// static __always_inline int edt_sched_departure(struct __ctx_buff *ctx)
// {
// 	__u64 delay, now, t, t_next;
// 	struct edt_id aggregate;
// 	struct edt_info *info;
// 	__u16 proto;

// 	//if (!validate_ethertype(ctx, &proto))
// 	//	return CTX_ACT_OK;
// 	//if (proto != bpf_htons(ETH_P_IP) &&
// 	//    proto != bpf_htons(ETH_P_IPV6))
// 	//	return CTX_ACT_OK;

// 	//aggregate.id = edt_get_aggregate(ctx);
// 	//if (!aggregate.id)
// 	//	return CTX_ACT_OK;

// 	//info = map_lookup_elem(&THROTTLE_MAP, &aggregate);
// 	//if (!info)
// 	//	return CTX_ACT_OK;

// 	now = ktime_get_ns();
// 	t = ctx->tstamp;
// 	if (t < now)
// 		t = now;
// 	delay = ((__u64)ctx_wire_len(ctx)) * NSEC_PER_SEC / info->bps;
// 	t_next = READ_ONCE(info->t_last) + delay;
// 	if (t_next <= t) {
// 		WRITE_ONCE(info->t_last, t);
// 		return CTX_ACT_OK;
// 	}
// 	/* FQ implements a drop horizon, see also 39d010504e6b ("net_sched:
// 	 * sch_fq: add horizon attribute"). However, we explicitly need the
// 	 * drop horizon here to i) avoid having t_last messed up and ii) to
// 	 * potentially allow for per aggregate control.
// 	 */
// 	if (t_next - now >= info->t_horizon_drop)
// 		return CTX_ACT_DROP;
// 	WRITE_ONCE(info->t_last, t_next);
// 	ctx->tstamp = t_next;
// 	return CTX_ACT_OK;
// }


struct bucket {
	volatile __u64 t_next;
	__u64 upper_limit_bps;
};

static __always_inline int limit_rate_ok(struct xdp_md *ctx, struct bucket *bucket)
{
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
int upf_qer_program_func(struct xdp_md *ctx)
{
	bpf_printk("upf_qer_program start\n");

	bpf_printk("tail call to UPF_PROG_TYPE_FAR key\n");
	bpf_tail_call(ctx, &upf_pipeline, UPF_PROG_TYPE_FAR);
	bpf_printk("tail call to UPF_PROG_TYPE_FAR key failed\n");
	return XDP_ABORTED;
}

char _license[] SEC("license") = "GPL";