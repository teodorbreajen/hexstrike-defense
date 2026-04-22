# Observability & Logging

## Structured Logging

- **Format**: JSON for SIEM integration
- **Correlation IDs**: UUID v4 for request tracing
- **Log Levels**: DEBUG, INFO, WARN, ERROR

## Metrics

- **Prometheus Metrics**: Endpoint at `/metrics`
- **Metrics Tracked**:
  - Total requests
  - Blocked requests
  - Allowed requests
  - Average latency
  - Status codes

## Key Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `total_requests` | Counter | Total requests processed |
| `blocked_requests` | Counter | Requests blocked by Lakera |
| `allowed_requests` | Counter | Requests allowed |
| `avg_latency_ms` | Gauge | Average latency in ms |

---

*Generated from metrics code*
