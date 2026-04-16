# =============================================================================
# Builder Stage - Compile Go binary
# =============================================================================
FROM docker.io/golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /build

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . ./

# Build arguments
ARG VERSION=1.0.0
ARG COMMIT_SHA=unknown

# Build the binary with version info embedded
RUN CGO_ENABLED=1 go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT_SHA}" \
    -o mcp-policy-proxy .

# =============================================================================
# Runtime Stage - Lightweight Alpine container
# =============================================================================
FROM docker.io/alpine:3.19 AS runtime

# Security: Install CA certificates for HTTPS connections
RUN apk add --no-cache ca-certificates

# Create non-root user (UID 1000 for security)
RUN adduser -D -u 1000 -G app appuser

# Labels for container metadata
LABEL maintainer="HexStrike Team <dev@hexstrike.io>"
LABEL description="MCP Policy Proxy - Semantic firewall for MCP tool calls"
LABEL version="1.0.0"
LABEL architecture="amd64"

# Expose port
EXPOSE 8080

# Create data directory
RUN mkdir -p /data /tmp /var/log && \
    chown -R appuser:app /data /tmp /var/log

# Copy binary from builder stage
COPY --from=builder /build/mcp-policy-proxy /usr/local/bin/

# Switch to non-root user
USER appuser

WORKDIR /data

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/mcp-policy-proxy"]

# =============================================================================
# Health Check Configuration
# =============================================================================
# Uses the /health endpoint for container health checks
# - Interval: 30s between checks
# - Timeout: 10s per check  
# - Start period: 5s for initial startup
# - Retries: 3 consecutive failures before marking unhealthy
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1