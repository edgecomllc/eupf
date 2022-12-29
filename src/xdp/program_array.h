#pragma once

#include <bpf/bpf_helpers.h>
#include <linux/bpf.h>
//#include <bpf/types.h>

#include <stdint.h>

enum upf_program_type {
    UPF_PROG_TYPE_MAIN,
    UPF_PROG_TYPE_FAR,
    UPF_PROG_TYPE_QER,
};

struct bpf_map_def SEC("maps") upf_pipeline = {
    .type = BPF_MAP_TYPE_PROG_ARRAY,
    .key_size = sizeof(uint32_t),
    .value_size = sizeof(uint32_t),
    .max_entries = 16,
};