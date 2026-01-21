# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version injection
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# Build the binary with version info
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X 'github.com/<your-username>/jump/internal/version.Version=${VERSION}' -X 'github.com/<your-username>/jump/internal/version.Commit=${COMMIT}' -X 'github.com/<your-username>/jump/internal/version.BuildDate=${BUILD_DATE}'" \
    -o jump ./cmd/jump

# Final stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /build/jump /usr/local/bin/jump

USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/jump"]