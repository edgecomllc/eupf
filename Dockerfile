FROM golang:1.26.10-bookworm AS builder

WORKDIR /app

# Install dependencies for ebpf compilation. gcc-multilib exists only on
# amd64, where it pulls in the /usr/include/asm symlink the eBPF compilation
# relies on; on arm64 (and other arches) the BPF_CFLAGS -I flag below points
# clang at the arch-specific headers instead.
RUN apt update \
    && apt install --no-install-recommends -y clang llvm gcc libbpf-dev \
    && if [ "$(dpkg --print-architecture)" = "amd64" ]; then apt install --no-install-recommends -y gcc-multilib; fi \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12

COPY go.mod go.sum ./
COPY cmd cmd

ARG BPF_ENABLE_LOG "0"
ARG BPF_ENABLE_ROUTE_CACHE "0"
RUN BPF_CFLAGS="" \
    && if [ "$BPF_ENABLE_LOG" = "1" ]; then BPF_CFLAGS="$BPF_CFLAGS -DENABLE_LOG"; fi \
    && if [ "$BPF_ENABLE_ROUTE_CACHE" = "1" ]; then BPF_CFLAGS="$BPF_CFLAGS -DENABLE_ROUTE_CACHE"; fi \
    && BPF_CFLAGS="-I/usr/include/$(gcc -dumpmachine) $BPF_CFLAGS" go generate -v ./cmd/...
RUN CGO_ENABLED=0 go build -v -o bin/eupf ./cmd/

FROM alpine:3.22 AS runtime
LABEL org.opencontainers.image.source="https://github.com/edgecomllc/eupf"

COPY --from=builder /app/bin/ /app/bin/
COPY --from=builder /app/cmd/docs/swagger.* /app/
COPY --from=builder /app/cmd/ebpf/zeroentrypoint_bpf.o /app/
COPY ./entrypoint.sh /app/bin/entrypoint.sh

# File capabilities, so eupf gets what it needs without running as root:
# BPF and PERFMON load the eBPF programs (kernel 5.8+, the chart requires
# 5.15), NET_ADMIN attaches XDP and sets routes, SYS_RESOURCE raises
# RLIMIT_MEMLOCK, and DAC_OVERRIDE creates the pin directory under
# /sys/fs/bpf, which is owned by root. Whoever runs the image, as root or
# not, must grant all of them (the chart does; docker needs --privileged or
# --cap-add): exec of a binary with the effective bit set fails with EPERM
# when one is missing from the bounding set.
RUN apk add --no-cache libcap \
    && setcap cap_net_admin,cap_sys_resource,cap_bpf,cap_perfmon,cap_dac_override=ep /app/bin/eupf \
    && apk del libcap

# CMD is overridden if arguments are passed.
ENTRYPOINT [ "sh", "/app/bin/entrypoint.sh" ]
