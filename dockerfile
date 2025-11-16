# Multi-stage Dockerfile for corroboros
# Builder stage: compile the Go binary
FROM golang:1.21-alpine AS builder
WORKDIR /src

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN apk add --no-cache git build-base

# Download deps first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the sources and build
COPY . .
RUN go build -o /out/corroboros ./cmd

# Runtime stage: minimal Alpine image with CA certs
FROM alpine:3.18
RUN apk add --no-cache ca-certificates

COPY --from=builder /out/corroboros /usr/local/bin/corroboros

EXPOSE 8080

# Run as non-root user
RUN addgroup -S app && adduser -S app -G app
USER app

ENTRYPOINT ["/usr/local/bin/corroboros"]

