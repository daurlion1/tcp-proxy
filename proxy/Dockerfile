# Dockerfile for proxy package
FROM golang:1.17-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o proxy .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/proxy .

CMD ["./proxy"]