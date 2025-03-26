# Build stage
FROM golang:1.24.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates curl bash

# Set working directory
WORKDIR /app

# Copy binary from builder and install it in PATH
COPY --from=builder /app/bin/llmscript /usr/local/bin/

# Copy examples directory
COPY --from=builder /app/examples ./examples

# Set the default command
CMD ["llmscript"]