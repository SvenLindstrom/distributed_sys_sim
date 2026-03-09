FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod ./

COPY ./internal/misc ./internal/misc
COPY ./internal/network ./internal/network
COPY ./internal/job ./internal/job
COPY ./internal/generator ./internal/generator

COPY ./cmd/jobgenerator ./cmd/jobgenerator

RUN go build -o generator ./cmd/jobgenerator

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/generator .

CMD ["./generator"]
