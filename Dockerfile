FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o certchecker ./cmd/certchecker

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for SSL verification
RUN apk add --no-cache ca-certificates tzdata curl

# Create necessary directories
RUN mkdir -p /root/.certchecker/config /root/.certchecker/logs /root/.certchecker/data

# Copy the binary from builder
COPY --from=builder /app/certchecker /app/certchecker

# Set environment variables
ENV PATH="/app:${PATH}" \
    LISTEN_ADDRESS="0.0.0.0:8081"

# Add health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8081/ || exit 1

# Expose HTTP port (default: 8081)
EXPOSE 8081

# Default command (can be overridden)
CMD ["./certchecker", "-webui"] 