# Design: Testing & Observability Enhancements

## Technical Approach

Enhance the MCP Policy Proxy with production-grade testing and observability capabilities: fuzz testing corpus management, integration tests with real HTTP servers, and Prometheus metrics export.

## Architecture Decisions

### Decision: Fuzz Corpus Storage Location

**Choice**: `testdata/fuzz/` directory using Go's standard fuzzing corpus format
**Alternatives considered**: External S3 bucket, in-memory generated corpus
**Rationale**: Go 1.18+ native fuzzing expects corpus files in `testdata/fuzz/<fuzz-test>/`. Enables `go test -fuzz` without additional tooling. Simple file-based versioning via git.

### Decision: Prometheus Metrics Format

**Choice**: Prometheus text-based exposition format (Content-Type: `text/plain; version=0.0.4`)
**Alternatives considered**: JSON metrics endpoint (already exists), OpenTelemetry
**Rationale**: Standard Prometheus scrape target format. Supports existing Prometheus ecosystem without Grafana dependency. Content negotiation preserves backward compatibility for JSON consumers.

### Decision: Integration Test Architecture

**Choice**: Dedicated `integration_test.go` with `httptest.Server` and controlled backends
**Alternatives considered**: Shared test helpers in existing test files, external test containers
**Rationale**: Keeps integration tests isolated from unit tests. `httptest.Server` provides real HTTP behavior without Docker overhead. Clear separation follows project conventions.

## Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    CI Pipeline Flow                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Unit Tests ──→ Fuzz Tests ──→ Integration Tests ──→ Deploy    │
│       │              │                 │                        │
│       ▼              ▼                 ▼                        │
│  go test -short  go test -fuzz    integration_test.go          │
│                      │                 │                        │
│                      ▼                 ▼                        │
│              testdata/fuzz/      httptest.Server × 2             │
│              (corpus files)      (proxy + backend)              │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                 Prometheus Metrics Flow                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Request ──→ Proxy ──→ Record Metrics ──→ /metrics ──→ Prometheus│
│                     │                      │                     │
│                     ▼                      ▼                     │
│              In-memory Metrics      Text Format Output            │
│              (existing mutex)       (prometheus lib)             │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `testdata/fuzz/FuzzSanitizeToolInput/corpus.json` | Create | Seed corpus for tool input fuzzing |
| `testdata/fuzz/FuzzIsInternalURL/corpus.txt` | Create | Seed corpus for SSRF detection |
| `testdata/fuzz/FuzzParseJSONRPC/corpus.json` | Create | Seed corpus for JSON-RPC parsing |
| `src/mcp-policy-proxy/integration_test.go` | Create | Integration tests with httptest.Server |
| `src/mcp-policy-proxy/prometheus.go` | Create | Prometheus metrics handler |
| `src/mcp-policy-proxy/proxy.go` | Modify | Wire Prometheus handler to /metrics |
| `src/mcp-policy-proxy/Makefile` | Modify | Add fuzz and prometheus CI targets |
| `src/mcp-policy-proxy/go.mod` | Modify | Add prometheus/client_golang dependency |

## Interfaces / Contracts

### Prometheus Metrics Handler

```go
// NewPrometheusHandler creates a handler that exports metrics in Prometheus format
func NewPrometheusHandler(m *Metrics) http.HandlerFunc

// Required metrics to export:
// - mcp_proxy_requests_total{allowed="true|false"} - Counter
// - mcp_proxy_request_duration_seconds - Histogram
// - mcp_proxy_status_codes_total{code="200"} - Counter
// - mcp_proxy_lakera_blocks_total - Counter
```

### Integration Test Server Setup

```go
// TestMain pattern for integration tests:
// 1. Create backend httptest.Server with controlled responses
// 2. Create proxy httptest.Server pointing to backend
// 3. Execute test scenarios
// 4. Verify metrics and responses
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Individual fuzz targets | `go test -v -fuzz=FuzzParseJSONRPC` |
| Fuzz | Corpus coverage + edge cases | `go test -fuzztime=30s -fuzz ./...` |
| Integration | Full HTTP lifecycle | `go test -tags=integration -run TestIntegration` |
| Metrics | Prometheus format output | `curl /metrics` + parse validation |

## Migration / Rollout

No migration required. All changes are additive:
- Fuzz corpus is new files
- Integration tests use build tags (`integration`)
- Prometheus handler adds `/metrics` alongside existing JSON endpoint (content-type differentiated)
- Makefile targets are new CI jobs

## Open Questions

- [ ] Should fuzz corpus be git-lfs tracked for binary inputs?
- [ ] Integration test CI time budget (recommend 60s max)?
