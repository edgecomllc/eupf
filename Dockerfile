FROM golang:1.20-bullseye as builder

WORKDIR /app

# Install dependencies for ebpf compilation
RUN apt update \
    && apt install --no-install-recommends -y clang llvm gcc-multilib libbpf-dev \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12

COPY go.mod go.sum ./
COPY cmd cmd

ARG BPF_ENABLE_LOG "0"
ARG BPF_ENABLE_ROUTE_CACHE "0"
RUN BPF_CFLAGS="" \
    && if [ "$BPF_ENABLE_LOG" = "1" ]; then BPF_CFLAGS="$BPF_CFLAGS -DENABLE_LOG"; fi \
    && if [ "$BPF_ENABLE_ROUTE_CACHE" = "1" ]; then BPF_CFLAGS="$BPF_CFLAGS -DENABLE_ROUTE_CACHE"; fi \
    && BPF_CFLAGS=$BPF_CFLAGS go generate -v ./cmd/...
RUN CGO_ENABLED=0 go build -v -o bin/eupf ./cmd/

FROM alpine:3.18.5 AS runtime
LABEL org.opencontainers.image.source="https://github.com/edgecomllc/eupf"

COPY --from=builder /app/bin/ /app/bin/
COPY --from=builder /app/cmd/docs/swagger.* /app/
COPY --from=builder /app/cmd/ebpf/zeroentrypoint_bpf.o /app/
COPY ./entrypoint.sh /app/bin/entrypoint.sh

# CMD is overridden if arguments are passed.
ENTRYPOINT [ "sh", "/app/bin/entrypoint.sh" ]
