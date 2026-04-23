# Observability & Logging

## Structured Logging

- **Format**: JSON for SIEM integration
- **Correlation IDs**: UUID v4 for request tracing
- **Log Levels**: DEBUG, INFO, WARN, ERROR, FATAL

## Log Format

```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "level": "INFO",
  "message": "Request processed",
  "request_id": "uuid-v4",
  "client_ip": "192.168.1.1",
  "path": "/mcp/proxy",
  "method": "POST",
  "status": 200,
  "latency_ms": 45
}
```

## Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `total_requests` | Counter | Total requests processed |
| `blocked_requests` | Counter | Requests blocked |
| `allowed_requests` | Counter | Requests allowed |
| `avg_latency_ms` | Gauge | Average latency in ms |
| `rate_limit_hits` | Counter | Rate limit rejections |

## Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'mcp-proxy'
    static_configs:
      - targets: ['mcp-proxy:8080']
```

## Tracing

- **Distributed Tracing**: OpenTelemetry compatible
- **Span Attributes**: request_id, user_id, path, method
- **Sample Rate**: 100% for errors, 10% for success

---

*Generated from observability code*
