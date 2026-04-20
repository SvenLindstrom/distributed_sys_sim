FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./internal/misc ./internal/misc
COPY ./internal/network ./internal/network
COPY ./internal/task ./internal/task
COPY ./cmd/worker ./cmd/worker
COPY ./internal/worker ./internal/worker

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/worker

FROM alpine
RUN apk add --no-cache iproute2

WORKDIR /app

COPY --from=builder /app/app .

CMD ["./app"]
