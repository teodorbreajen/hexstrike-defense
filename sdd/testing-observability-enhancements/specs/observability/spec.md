# Delta for observability/

## ADDED Requirements

### Requirement: Prometheus Metrics Endpoint

The system SHALL expose a `/metrics` endpoint that outputs metrics in Prometheus text format (RFC-compliant) for scraping by Prometheus or compatible collectors.

The `/metrics` endpoint MUST:
- Return Content-Type: `text/plain; version=0.0.4; charset=utf-8`
- Output metrics in Prometheus text format (see https://prometheus.io/docs/instrumenting/exposition_formats/)
- Include help text and type annotations for each metric
- Support metric names with underscores and labels

The following metrics MUST be exposed:

| Metric Name | Type | Labels | Description |
|------------|------|--------|-------------|
| `mcp_proxy_requests_total` | Counter | `status`, `allowed` | Total requests processed |
| `mcp_proxy_requests_blocked_total` | Counter | `reason` | Requests blocked by policy |
| `mcp_proxy_request_duration_seconds` | Histogram | `status` | Request latency distribution |
| `mcp_proxy_rate_limit_hits_total` | Counter | - | Rate limit rejections |
| `mcp_proxy_circuit_breaker_state` | Gauge | - | Circuit breaker state (0=closed, 1=open, 2=half-open) |
| `mcp_proxy_lakera_blocks_total` | Counter | - | Requests blocked by Lakera |
| `mcp_proxy_active_connections` | Gauge | - | Current concurrent requests |

#### Scenario: Prometheus scrapes metrics endpoint

- GIVEN a Prometheus server configured to scrape `localhost:8080/metrics`
- WHEN Prometheus fetches the `/metrics` endpoint
- THEN the response MUST be valid Prometheus text format
- AND all defined metrics MUST be present with correct types

#### Scenario: Metrics reflect current state

- GIVEN the proxy has processed requests
- WHEN `/metrics` is queried
- THEN `mcp_proxy_requests_total` value MUST equal total requests processed
- AND `mcp_proxy_requests_total{allowed="true"}` MUST equal allowed requests
- AND `mcp_proxy_requests_total{allowed="false"}` MUST equal blocked requests

#### Scenario: Histogram buckets are present

- GIVEN Prometheus format requirement
- WHEN `/metrics` is queried
- THEN `mcp_proxy_request_duration_seconds` MUST include bucket labels (le="0.01", le="0.05", le="0.1", le="0.5", le="1", le="+Inf")
- AND the histogram MUST include `_sum` and `_count` suffixes

## MODIFIED Requirements

*(None)*

## REMOVED Requirements

*(None)*
