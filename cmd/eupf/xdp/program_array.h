#pragma once

#include <bpf/bpf_helpers.h>
#include <linux/bpf.h>

#include <stdint.h>

enum upf_program_type {
    UPF_PROG_TYPE_MAIN = 0,
    UPF_PROG_TYPE_FAR = 1,
    UPF_PROG_TYPE_QER = 2,
};

struct
{
    __uint(type, BPF_MAP_TYPE_PROG_ARRAY);
    __type(key, uint32_t);
    __type(value, uint32_t);
    __uint(max_entries, 16);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} upf_pipeline SEC(".maps");
