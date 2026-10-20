# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/order-service ./internal/cmd/order

FROM alpine:3.20
COPY --from=builder /out/order-service /order-service
ENTRYPOINT ["/order-service"]
