#include <linux/bpf.h>
#include <linux/if_ether.h>

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define ETH_P_IP 0x0800 /* Internet Protocol packet */
#define ETH_P_NSH	0x894F /* Network Service Header */

#include "xdp/utils/nsh.h"
#include "xdp/utils/packet_context.h"
#include "xdp/utils/trace.h"

// TODO: support metadata
static __always_inline __u32 add_nsh_over_ip4_headers(struct packet_context *ctx, __u32 path_hdr) {

    static const size_t nsh_encap_size =  NSH_BASE_HDR_LEN; // Without metadata


    int result = bpf_xdp_adjust_head(ctx->xdp_ctx, (__s32)-nsh_encap_size);
    if (result)
        return -1;

    char *data = (char *)(long)ctx->xdp_ctx->data;
    char *data_end = (char *)(long)ctx->xdp_ctx->data_end;

    struct ethhdr *orig_eth = (struct ethhdr *)(data + nsh_encap_size);
    if ((const char *)(orig_eth + 1) > data_end)
        return -1;

    struct ethhdr *eth = (struct ethhdr *)data;
    __builtin_memcpy(eth, orig_eth, sizeof(*eth));
    eth->h_proto = bpf_htons(ETH_P_NSH);

    // /* Add the NSH header */
    struct nshhdr *nsh = (struct nshhdr *)(eth + 1);
    if ((const char *)(nsh + 1) > data_end)
        return -1;

    nsh->ver_flags_ttl_len = 0;
    nsh->mdtype          = NSH_M_TYPE2;
    nsh->np              = 0x01;
    nsh->path_hdr             = bpf_htonl(path_hdr);
    nsh_set_flags_ttl_len(nsh, 0x0, 0x3F, nsh_encap_size);

    upf_printk("upf: added nsh encap");

    struct iphdr *ip = (struct iphdr *)((char *)nsh + nsh_encap_size);
    if ((const char *)(ip + 1) > data_end)
        return -1;

    context_set_ip4(ctx, (char *)(long)ctx->xdp_ctx->data, (const char *)(long)ctx->xdp_ctx->data_end, eth, ip, 0, 0, nsh);
    return 0;
}

static __always_inline long remove_nsh_header(struct packet_context *ctx) {
    return 0;
}

