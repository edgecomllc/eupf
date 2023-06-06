FROM golang:1.20-bullseye as builder

WORKDIR /app

EXPOSE 2345

#Install dependencies for ebpf compilation
RUN apt update \
    && apt install --no-install-recommends -y clang llvm gcc-multilib libbpf-dev \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12
RUN go install github.com/go-delve/delve/cmd/dlv@latest

COPY go.mod go.sum ./
COPY cmd/eupf cmd/eupf

RUN go generate -v ./cmd/eupf
RUN CGO_ENABLED=0 go build -gcflags "all=-N -l" -v -o bin/eupf ./cmd/eupf

RUN dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec bin/eupf
#FROM alpine:3.18 AS runtime
#LABEL org.opencontainers.image.source="https://github.com/edgecomllc/eupf"

#COPY --from=builder /app/bin/ /app/bin/
#COPY ./entrypoint.sh /app/bin/entrypoint.sh

# CMD is overridden if arguments are passed.
#ENTRYPOINT [ "sh", "/app/bin/entrypoint.sh" ]
