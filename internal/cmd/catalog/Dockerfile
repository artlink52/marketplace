# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/catalog-service ./internal/cmd/catalog

FROM alpine:3.20
COPY --from=builder /out/catalog-service /catalog-service
ENTRYPOINT ["/catalog-service"]
