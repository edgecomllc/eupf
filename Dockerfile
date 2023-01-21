FROM golang:1.19.2-bullseye as builder
 
WORKDIR /app
 
# Effectively tracks changes within your go.mod file
COPY go.mod go.sum ./
 
RUN go mod download && go mod verify
RUN go generate
RUN go build -v -o bin/eupf ./...

FROM golang:1.19.2-alpine AS runtime

WORKDIR /app
COPY --from=builder /app/bin/ /app/bin/
ENTRYPOINT ["./bin/eupf"]
