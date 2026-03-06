FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod ./

COPY ./internal/worker ./internal/worker
COPY ./internal/misc ./internal/misc
COPY ./internal/network ./internal/network
COPY ./internal/job ./internal/job

COPY ./cmd/worker ./cmd/worker

RUN go build -o worker ./cmd/worker

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/worker .

CMD ["./worker"]