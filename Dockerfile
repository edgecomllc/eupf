FROM golang:1.19.2-bullseye as builder

WORKDIR /app

#Install dependencies for ebpf compilation
RUN apt update \
    && apt install --no-install-recommends -y clang llvm gcc-multilib libbpf-dev \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
COPY cmd/eupf cmd/eupf

RUN go generate -v cmd/eupf/ebpf_objects.go
RUN CGO_ENABLED=0 go build -v -o bin/eupf ./cmd/eupf

FROM alpine:3.17 AS runtime
LABEL org.opencontainers.image.source="https://github.com/edgecomllc/eupf"

COPY --from=builder /app/bin/ /app/bin/
COPY ./entrypoint.sh /app/bin/entrypoint.sh

# CMD is overridden if arguments are passed.
ENTRYPOINT [ "sh", "/app/bin/entrypoint.sh" ]