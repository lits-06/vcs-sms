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
COPY ./auth_service ./auth_service

COPY ./pkg ./pkg

COPY ./proto ./proto

# Build the application
RUN go build -o auth ./auth_service/cmd/main.go

# Final stage
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/auth ./auth

# Copy config file
COPY --from=builder /app/auth_service/config/config.yaml ./auth_service/config/

# Expose HTTP port
EXPOSE 8003

# Run the application
CMD ["./auth"]
