# Multi-stage build for smaller image size
FROM golang:1.21-alpine AS builder

# Install build dependencies including musl for CGO
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build all binaries with CGO enabled for proper TLS/SSL support
RUN CGO_ENABLED=1 GOOS=linux go build -o /worker ./cmd/worker
RUN CGO_ENABLED=1 GOOS=linux go build -o /api ./cmd/api

# Final stage
FROM alpine:latest

# Install ca-certificates and libc for CGO binaries
RUN apk --no-cache add ca-certificates tzdata libc6-compat

WORKDIR /root/

# Copy binaries from builder
COPY --from=builder /worker .
COPY --from=builder /api .

# Expose port (optional, for health checks)
EXPOSE 8080

# Run API server by default
CMD ["/root/api"]
