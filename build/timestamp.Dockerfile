FROM golang:1.21-alpine AS builder
WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Cache and download modules
COPY go.mod go.sum ./
RUN apk add --no-cache git && go mod download

# Copy source and build static binary
COPY . .
ARG BUILD_TARGET=./cmd/timestamp
RUN go build -trimpath -ldflags="-s -w" -o /timestamp ${BUILD_TARGET}

# Minimal runtime image
FROM scratch
# If your binary needs CA certs or other files, replace scratch with alpine and install ca-certificates.
COPY --from=builder /timestamp /timestamp

EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/timestamp"]