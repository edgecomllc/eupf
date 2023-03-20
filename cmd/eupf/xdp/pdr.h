#pragma once

#include <bpf/bpf_helpers.h>
#include <linux/bpf.h>

struct pdr_info
{
    __u8 outer_header_removal;
    __u16 far_id;
};

#ifdef __RELEASE
struct bpf_map_def SEC("maps") pdr_map_uplink_ip4 = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u32), // IPv4
    .value_size = sizeof(struct pdr_info),
    .max_entries = 1024, // FIXME
};

struct bpf_map_def SEC("maps") pdr_map_downlink_ip4 = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u32), // TEID
    .value_size = sizeof(struct pdr_info),
    .max_entries = 1024, // FIXME
};
#else
struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u32); // ipv4
    __type(value, struct pdr_info);
    __uint(max_entries, 1024);
} pdr_map_uplink_ip4 SEC(".maps");

struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u32); // teid
    __type(value, struct pdr_info);
    __uint(max_entries, 1024);
} pdr_map_downlink_ip4 SEC(".maps");
#endif

enum far_action_mask {
    FAR_DROP = 0x00,
    FAR_FORW = 0x01,
    FAR_BUFF = 0x02,
    FAR_NOCP = 0x04,
    FAR_DUPL = 0x08,
    FAR_IPMA = 0x10,
    FAR_IPMD = 0x20,
    FAR_DFRT = 0x40,
};

struct far_info
{
    __u8 action;
    __u8 outer_header_creation;
    __u32 teid;
    __u32 srcip;
};

#ifdef __RELEASE
struct bpf_map_def SEC("maps") far_map = {
    .type = BPF_MAP_TYPE_ARRAY,
    .key_size = sizeof(__u32), // FAR ID
    .value_size = sizeof(struct far_info),
    .max_entries = 1024, // FIXME
};
#else
struct
{
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32); // cpu
    __type(value, struct far_info);
    __uint(max_entries, 1024);
} far_map SEC(".maps");
#endif