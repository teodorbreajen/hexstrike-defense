FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o mcp-policy-proxy -ldflags="-s -w" .

# Final stage
FROM alpine:3.19

# Install ca-certificates for HTTPS connections
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mcp-policy-proxy .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./mcp-policy-proxy"]
