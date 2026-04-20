FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod ./

COPY ./internal/misc ./internal/misc
COPY ./internal/network ./internal/network
COPY ./internal/task ./internal/task
COPY ./internal/generator ./internal/generator

COPY ./cmd/taskgenerator ./cmd/taskgenerator

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o app ./cmd/taskgenerator

FROM alpine
RUN apk add --no-cache iproute2

WORKDIR /app

COPY --from=builder /app/app .

CMD ["./app"]
