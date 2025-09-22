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
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o report_service ./report_service/cmd/main.go

# Final stage
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/report_service .

# Copy config file
COPY --from=builder /app/report_service/config/config.yaml ./config/

# Copy HTML templates
COPY --from=builder /app/report_service/internal/templates/ ./templates/

# Expose HTTP port
EXPOSE 8002

# Run the application
CMD ["./report_service"]
