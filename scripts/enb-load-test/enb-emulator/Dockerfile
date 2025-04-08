FROM golang:1.24

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -v -o /app/app main.go

CMD ["/app/app"]
