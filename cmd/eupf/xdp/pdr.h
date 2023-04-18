#pragma once

#include <bpf/bpf_helpers.h>
#include <linux/bpf.h>
#include <linux/ipv6.h>


enum outer_header_removal_values {
    OHR_GTP_U_UDP_IPv4  = 0,
    OHR_GTP_U_UDP_IPv6  = 1,
    OHR_UDP_IPv4        = 2,
    OHR_UDP_IPv6        = 3,
    OHR_IPv4            = 4,
    OHR_IPv6            = 5,
    OHR_GTP_U_UDP_IP    = 6,
    OHR_VLAN_S_TAG      = 7,
    OHR_S_TAG_C_TAG     = 8,
};

struct pdr_info
{
    __u8 outer_header_removal;
    __u32 far_id;
    __u32 qer_id;
};

#ifdef __RELEASE
struct bpf_map_def SEC("maps") pdr_map_downlink_ip4 = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u32), // IPv4
    .value_size = sizeof(struct pdr_info),
    .max_entries = 1024, // FIXME
};

struct bpf_map_def SEC("maps") pdr_map_downlink_ip6 = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(struct in6_addr), // IPv6
    .value_size = sizeof(struct pdr_info),
    .max_entries = 1024, // FIXME
};

struct bpf_map_def SEC("maps") pdr_map_uplink = {
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
} pdr_map_downlink_ip4 SEC(".maps");

struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, struct in6_addr); // ipv6
    __type(value, struct pdr_info);
    __uint(max_entries, 1024);
} pdr_map_downlink_ip6 SEC(".maps");

struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u32); // teid
    __type(value, struct pdr_info);
    __uint(max_entries, 1024);
} pdr_map_uplink_ip4 SEC(".maps");
#endif

enum far_action_mask {
    FAR_DROP = 0x01,
    FAR_FORW = 0x02,
    FAR_BUFF = 0x04,
    FAR_NOCP = 0x08,
    FAR_DUPL = 0x10,
    FAR_IPMA = 0x20,
    FAR_IPMD = 0x40,
    FAR_DFRT = 0x80,
};

struct far_info
{
    __u8 action;
    __u8 outer_header_creation;
    __u32 teid;
    __u32 remoteip;
    __u32 localip;
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