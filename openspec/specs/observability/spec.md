# Observability Specification (Delta)

## Purpose

Define structured logging requirements for SIEM integration and incident response.

## ADDED Requirements

### Requirement: JSON Structured Logging

All log output SHALL be in JSON format for SIEM consumption.

**Log Format:**
```json
{
  "timestamp": "2026-04-16T10:30:00.000Z",
  "level": "INFO|WARN|ERROR",
  "correlation_id": "uuid-v4",
  "component": "proxy|lakera|ratelimit",
  "message": "human readable message",
  "request_id": "uuid-v4",
  "method": "POST",
  "path": "/mcp/tools/call",
  "status_code": 200,
  "latency_ms": 45,
  "error": "error message if present"
}
```

**Required Fields:**
- `timestamp`: ISO 8601 UTC format
- `level`: INFO, WARN, ERROR
- `correlation_id`: Request-scoped UUID
- `component`: Which subsystem generated the log

**Optional Fields:**
- `request_id`: Per-request identifier
- `method`: HTTP method
- `path`: Request path
- `status_code`: HTTP response code
- `latency_ms`: Request duration
- `error`: Error message or stack trace

### Requirement: Correlation ID Generation and Propagation

The system SHALL generate and propagate correlation IDs for request tracing.

**Requirements:**
- Generate UUID v4 at request start
- Store in request context
- Include in all logs for that request
- Return in `X-Correlation-ID` response header
- Forward to MCP backend in `X-Correlation-ID` header

#### Scenario: Correlation ID in logs

- GIVEN an incoming HTTP request
- WHEN request is received
- THEN correlation_id SHALL be generated
- AND included in all log entries for that request
- AND returned in `X-Correlation-ID` response header

#### Scenario: Correlation ID forwarded to backend

- GIVEN an incoming request with correlation ID `abc-123`
- WHEN request is forwarded to MCP backend
- THEN backend request SHALL include header `X-Correlation-ID: abc-123`

### Requirement: Security Event Logging

The system SHALL log security-relevant events with elevated verbosity.

**Security Events to Log:**
| Event | Level | Required Fields |
|-------|-------|------------------|
| Request blocked by Lakera | WARN | tool_name, score, reason, correlation_id |
| Lakera error (fail-closed) | ERROR | error, correlation_id |
| Authentication failure | WARN | path, error, ip_address, correlation_id |
| Rate limit exceeded | WARN | ip_address, correlation_id |
| Oversized request body | WARN | size, max_size, correlation_id |
| Malformed JSON-RPC | INFO | error, correlation_id |

### Requirement: Metrics Export Format

Metrics endpoint SHALL return Prometheus-compatible JSON.

**Response Format:**
```json
{
  "total_requests": 1000,
  "blocked_requests": 50,
  "allowed_requests": 950,
  "avg_latency_ms": 45.5,
  "lakera_blocks": 50,
  "status_codes": {"200": 900, "403": 50, "503": 50},
  "timestamp": "2026-04-16T10:30:00Z"
}
```

### Requirement: Health Check with Detailed Status

Health endpoint SHALL provide component-level health status.

**Response Format:**
```json
{
  "status": "healthy|degraded|unhealthy",
  "timestamp": "2026-04-16T10:30:00Z",
  "checks": {
    "lakera": "ok|unavailable:reason",
    "mcp_backend": "ok|unavailable:reason"
  }
}
```

**Health Criteria:**
- `healthy`: All checks pass
- `degraded`: Non-critical checks fail (e.g., Lakera unavailable with fail-open)
- `unhealthy`: Critical checks fail (e.g., MCP backend unavailable)

#### Scenario: Health check shows degraded on Lakera failure

- GIVEN Lakera is unreachable
- WHEN GET `/health` is called
- THEN status SHALL be `"degraded"`
- AND checks.lakera SHALL contain `"unavailable: connection refused"`

### Requirement: Log Level Configuration

The system SHALL support configurable log levels via environment variable.

**Environment Variable:** `LOG_LEVEL`
- Values: `debug`, `info`, `warn`, `error`
- Default: `info`

Production deployments SHOULD use `warn` or `error` to reduce log volume.
