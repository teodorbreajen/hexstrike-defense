# Tasks: Testing & Observability Enhancements

## Phase 1: Fuzz Testing Foundation

- [x] 1.1 Create `testdata/fuzz/FuzzSanitizeToolInput/corpus.json` with valid/invalid JSON inputs
- [x] 1.2 Create `testdata/fuzz/FuzzIsInternalURL/corpus.txt` with safe/unsafe URL samples
- [x] 1.3 Create `testdata/fuzz/FuzzParseJSONRPC/corpus.json` with valid/malformed JSON-RPC requests
- [x] 1.4 Add `fuzz` target to `src/mcp-policy-proxy/Makefile`: `go test -fuzztime=30s -fuzz ./...`
- [x] 1.5 Add `ci-fuzz` target to Makefile for GitHub Actions (1 min timeout, conservative iterations)
- [x] 1.6 Add fuzz run instructions to `src/mcp-policy-proxy/fuzz/README.md` (local + CI)

## Phase 2: Integration Tests

- [x] 2.1 Create `src/mcp-policy-proxy/integration_test.go` with `//go:build integration` tag
- [x] 2.2 Implement `TestIntegration_ProxyForwardsRequests` using `httptest.Server` for backend
- [x] 2.3 Implement `TestIntegration_MetricsRecorded` verifying request counters increment
- [x] 2.4 Implement `TestIntegration_RejectInternalHost` testing SSRF blocking end-to-end
- [x] 2.5 Add `integration` target to Makefile: `go test -tags=integration -run TestIntegration -v`
- [x] 2.6 Create `testdata/integration/mock_backend_responses.go` helper with controlled JSON-RPC responses

## Phase 3: Prometheus Metrics

- [x] 3.1 Add `github.com/prometheus/client_golang/prometheus` to `src/mcp-policy-proxy/go.mod`
- [x] 3.2 Create `src/mcp-policy-proxy/prometheus.go` with `NewPrometheusHandler(m *Metrics) http.HandlerFunc`
- [x] 3.3 Export `mcp_proxy_requests_total{allowed="true|false"}` counter
- [x] 3.4 Export `mcp_proxy_request_duration_seconds` histogram
- [x] 3.5 Export `mcp_proxy_status_codes_total{code="200"}` counter
- [x] 3.6 Export `mcp_proxy_lakera_blocks_total` counter
- [x] 3.7 Wire `/metrics` route in `src/mcp-policy-proxy/proxy.go` mux setup
- [x] 3.8 Add `prometheus-test` target to Makefile: curl /metrics and validate format
- [x] 3.9 Write `TestPrometheusFormat` verifying text output parses correctly

## Phase 4: CI Integration

- [x] 4.1 Create `.github/workflows/ci.yml` with fuzz test job (60s min)
- [x] 4.2 Create `.github/workflows/ci.yml` with integration test job (`//go:build integration`)
- [x] 4.3 Create `.github/workflows/ci.yml` with prometheus metrics validation job
- [x] 4.4 Verified all tests pass locally and `go mod tidy` successful
