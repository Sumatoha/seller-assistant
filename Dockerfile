# Multi-stage build for smaller image size
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build binaries
RUN go build -o /api ./cmd/api
RUN go build -o /worker ./cmd/worker

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /api .
COPY --from=builder /worker .

EXPOSE 8080

CMD ["./api"]
