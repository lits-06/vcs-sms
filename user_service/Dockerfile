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
COPY ./user_service ./user_service

COPY ./pkg ./pkg

COPY ./proto ./proto

# Build the application
RUN go build -o user ./user_service/cmd/main.go

# Final stage
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/user ./user

# Expose ports - HTTP and gRPC
EXPOSE 8001 50001

# Run the application
CMD ["./user"]
