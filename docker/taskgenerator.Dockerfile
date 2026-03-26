FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod ./

COPY ./internal/misc ./internal/misc
COPY ./internal/network ./internal/network
COPY ./internal/task ./internal/task
COPY ./internal/generator ./internal/generator

COPY ./cmd/taskgenerator ./cmd/taskgenerator

RUN go build -o generator ./cmd/taskgenerator

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/generator .

CMD ["./generator"]
