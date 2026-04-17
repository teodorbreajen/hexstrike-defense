# Proposal: Testing & Observability Enhancements

## Intent

Add missing testing capabilities (fuzz testing, integration tests with real backend, Prometheus metrics) to achieve production-grade CI/CD pipeline and observability standards.

## Scope

### In Scope
- **T-1**: Fuzz testing in CI — Integrate go-fuzz/libFuzzer into test pipeline
- **T-2**: Integration tests with real backend — E2E tests using httptest server (not just unit mocks)
- **O-1**: Prometheus metrics export — Expose `/metrics` endpoint in Prometheus text format

### Out of Scope
- OpenTelemetry tracing (deferred to futureobservability phase)
- Grafana dashboards (deferred to future devops phase)

## Approach

1. **Fuzz Testing**: Add `FuzzParseJSONRPC` corpus and integrate `go test -fuzz` in Makefile CI target
2. **Integration Tests**: Create `integration_test.go` that spins up real httptest server with controlled responses, testing full HTTP lifecycle
3. **Prometheus Metrics**: Add `/metrics` handler using prometheus/client_golang TextParser output

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `src/mcp-policy-proxy/fuzz_test.go` | Modified | Add fuzz corpus, integrate in CI |
| `src/mcp-policy-proxy/integration_test.go` | New | Real backend E2E tests |
| `src/mcp-policy-proxy/proxy.go` | Modified | Add `/metrics` Prometheus endpoint |
| `Makefile` | Modified | Add fuzz test and prometheus test targets |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Fuzz corpus explosion | Low | Limit iterations, seed with valid/invalid samples |
| Integration test flakiness | Low | Strict timeouts, controlled mock responses |
| Prometheus export overhead | Low | Lazy initialization, minimal allocation |

## Rollback Plan

- Revert `proxy.go` to remove `/metrics` handler
- Delete `integration_test.go`
- Remove fuzz targets from Makefile
- All changes are additive only

## Dependencies

- `prometheus/client_golang` (add to go.mod)

## Success Criteria

- [ ] Fuzz tests run in CI and discover 0 new bugs
- [ ] Integration tests pass against real httptest server
- [ ] `/metrics` returns Prometheus text format with all defined metrics
- [ ] Coverage maintained above 70%