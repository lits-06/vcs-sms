# Build stage
FROM golang:1.24.1-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY ./healthcheck_service ./healthcheck_service

COPY ./pkg ./pkg

# Build the application
RUN go build -o healthcheck ./healthcheck_service/cmd/main.go

# Final stage
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/healthcheck ./healthcheck

# Copy config file
COPY --from=builder /app/healthcheck_service/config/config.yaml ./healthcheck_service/config/

# Run the application
CMD ["./healthcheck"]
