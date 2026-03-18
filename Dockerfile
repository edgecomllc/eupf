FROM golang:1.24 AS builder

WORKDIR /app

# Install dependencies for ebpf compilation
RUN apt update \
    && apt install --no-install-recommends -y clang llvm gcc-multilib libbpf-dev \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12

COPY go.mod go.sum ./
# Pre-download dependencies to optimize build caching
RUN go mod download -x
COPY cmd cmd

ARG BPF_ENABLE_LOG "0"
ARG BPF_ENABLE_ROUTE_CACHE "0"
ENV BPF_CFLAGS=""
RUN if [ "$BPF_ENABLE_LOG" = "1" ]; then \
        echo "Enabling BPF logging"; \
        export BPF_CFLAGS="$BPF_CFLAGS -DENABLE_LOG"; \
    fi \
    && if [ "$BPF_ENABLE_ROUTE_CACHE" = "1" ]; then \
        echo "Enabling route cache"; \
        export BPF_CFLAGS="$BPF_CFLAGS -DENABLE_ROUTE_CACHE"; \
    fi \
    && echo "Final BPF_CFLAGS: $BPF_CFLAGS" \
    && go generate -v ./cmd/...

RUN CGO_ENABLED=0 go build -v -o bin/eupf ./cmd/

FROM alpine:3.22.0 AS runtime
LABEL org.opencontainers.image.source="https://github.com/edgecomllc/eupf"

COPY --from=builder /app/bin/ /app/bin/
COPY --from=builder /app/cmd/docs/swagger.* /app/
COPY --from=builder /app/cmd/ebpf/zeroentrypoint_bpf.o /app/
COPY ./entrypoint.sh /app/bin/entrypoint.sh

RUN apk add iproute2 --no-cache

# CMD is overridden if arguments are passed.
ENTRYPOINT [ "sh", "/app/bin/entrypoint.sh" ]
