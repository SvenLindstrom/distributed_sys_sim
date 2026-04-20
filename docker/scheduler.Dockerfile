FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./internal/network ./internal/network
COPY ./internal/task ./internal/task
COPY ./internal/fault ./internal/fault
COPY ./internal/misc ./internal/misc
COPY ./cmd/scheduler ./cmd/scheduler
COPY ./internal/scheduler ./internal/scheduler


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o app ./cmd/scheduler

FROM alpine
RUN apk add --no-cache iproute2

WORKDIR /app

COPY --from=builder /app/app .

CMD ["./app"]
