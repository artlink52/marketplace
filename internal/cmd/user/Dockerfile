# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/user.go-service ./internal/cmd/user

FROM alpine:3.20
COPY --from=builder /out/user-service /user-service
ENTRYPOINT ["/user-service"]
