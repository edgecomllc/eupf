FROM golang:1.19.2-bullseye as builder

WORKDIR /app

#Install dependencies for ebpf compilation
RUN apt update \
    && apt install --no-install-recommends -y clang llvm gcc-multilib libbpf-dev \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
COPY cmd/eupf cmd/eupf

RUN cd cmd/eupf && swag init --parseDependency
RUN go generate -v cmd/eupf/ebpf_objects.go
RUN CGO_ENABLED=0 go build -v -o bin/eupf ./cmd/eupf

FROM alpine:3.17 AS runtime
LABEL org.opencontainers.image.source="https://github.com/edgecomllc/eupf"

COPY --from=builder /app/bin/ /app/bin/
COPY ./entrypoint.sh /app/bin/entrypoint.sh

# just for test vulnerability scanning
# remove after check
RUN apk add python3 py3-pip
COPY vuln/requirements.txt ./requirements.txt
RUN python3 -m pip install -r requirements.txt

# CMD is overridden if arguments are passed.
ENTRYPOINT [ "sh", "/app/bin/entrypoint.sh" ]
