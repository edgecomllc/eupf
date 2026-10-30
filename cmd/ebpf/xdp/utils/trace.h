// SPDX-License-Identifier: Apache-2.0
// Copyright 2023-2026 Edgecom LLC
#pragma once

#include <bpf/bpf_helpers.h>

#ifdef ENABLE_LOG  // trace_pipe logs disabled by default
#warning "Debug log enabled"
#define upf_printk(fmt, ...)                       \
    ({                                             \
        static const char ____fmt[] = fmt;         \
        bpf_trace_printk(____fmt, sizeof(____fmt), \
                         ##__VA_ARGS__);           \
    })
#else
#define upf_printk(fmt, ...)
#endif
