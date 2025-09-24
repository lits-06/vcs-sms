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
COPY ./report_service ./report_service

COPY ./pkg ./pkg

# Build the application
RUN go build -o report ./report_service/cmd/main.go

# Final stage
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/report ./report

# Expose HTTP port
EXPOSE 8002

# Run the application
CMD ["./report"]
