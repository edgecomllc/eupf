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


typedef unsigned __int128 __u128;

static __always_inline __u128 __bswap128(__u128 in) {
    union _128_as_64 {
        __u128 v;
        __u64 q[2];
    } u1;

    u1.q[1] = bpf_be64_to_cpu(((union _128_as_64)in).q[0]);
    u1.q[0] = bpf_be64_to_cpu(((union _128_as_64)in).q[1]);
    return u1.v;
}

#define ___bpf_swab128(x) ((__u128)(			\
    ___bpf_mvb(x, 128, 0,  15) |	\
    ___bpf_mvb(x, 128, 1,  14) |	\
    ___bpf_mvb(x, 128, 2,  13) |	\
    ___bpf_mvb(x, 128, 3,  12) |	\
    ___bpf_mvb(x, 128, 4,  11) |	\
    ___bpf_mvb(x, 128, 5,  10) |	\
    ___bpf_mvb(x, 128, 6,  9)  |	\
    ___bpf_mvb(x, 128, 7,  8)  |    \
    ___bpf_mvb(x, 128, 8,  7)  |	\
    ___bpf_mvb(x, 128, 9,  6)  |	\
    ___bpf_mvb(x, 128, 10, 5)  |	\
    ___bpf_mvb(x, 128, 11, 4)  |	\
    ___bpf_mvb(x, 128, 12, 3)  |	\
    ___bpf_mvb(x, 128, 13, 2)  |	\
    ___bpf_mvb(x, 128, 14, 1)  |	\
    ___bpf_mvb(x, 128, 15, 0)))

    #if __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
    # define __bpf_ntohlll(x)			__bswap128(x)
    # define __bpf_htonlll(x)			__bswap128(x)
    # define __bpf_constant_ntohlll(x)	___bpf_swab128(x)
    # define __bpf_constant_htonlll(x)	___bpf_swab128(x)
    #elif __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
    # define __bpf_ntohlll(x)			(x)
    # define __bpf_htonlll(x)			(x)
    # define __bpf_constant_ntohlll(x)	(x)
    # define __bpf_constant_htonlll(x)	(x)
    #else
    # error "Fix your compiler's __BYTE_ORDER__?!"
    #endif

#define bpf_htonlll(x)				\
	(__builtin_constant_p(x) ?		\
	 __bpf_constant_htonlll(x) : __bpf_htonlll(x))
#define bpf_ntohlll(x)				\
	(__builtin_constant_p(x) ?		\
	 __bpf_constant_ntohlll(x) : __bpf_ntohlll(x))