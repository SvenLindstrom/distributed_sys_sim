FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod  ./

COPY ./internal/scheduler ./internal/scheduler
COPY ./internal/misc ./internal/misc
COPY ./internal/network ./internal/network
COPY ./internal/job ./internal/job

COPY ./cmd/scheduler ./cmd/scheduler

RUN go build -o scheduler ./cmd/scheduler

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/scheduler .

CMD ["./scheduler"]
