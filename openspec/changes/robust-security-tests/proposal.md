# Proposal: Robust Security Tests for mcp-policy-proxy

## Intent

Current security tests have critical gaps: they test configuration constants rather than real behavior, miss attack vectors (SQL injection, command injection, SSRF), and use fragile timing (1-second sleeps). This change transforms unit tests into integration tests that verify actual security enforcement.

## Scope

### In Scope
- Add `mockLakeraClient` helper for controlling Lakera responses in tests
- Create RSA/EC key pairs for JWT algorithm confusion tests
- Expand sanitization test cases (SQLi, command injection, XSS, Unicode bypass, case variations)
- Add lakera nil fallback behavior test
- Convert SSRF/BatchSize/FailClosed tests from unit to integration (test `proxy.Handler()`)
- Make CircuitBreaker timeout configurable for faster tests
- Add rate limiter cleanup/TTL/maxClients tests

### Out of Scope
- New security features (only test hardening)
- Performance testing
- Load testing

## Approach

1. **Add mock helper** (`mockLakeraClient`) to control `allowed`, `score`, `reason`, `error` responses
2. **RSA/ECDSA keys** generated once in `init()` or `TestMain()` for JWT tests
3. **Integration tests** use `proxy.Handler()` with `httptest.NewRecorder`
4. **Configurable timeouts** via `NewCircuitBreakerConfig(testing.T)` or env var
5. **Table-driven cases** for sanitization with clear expected outcomes

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `src/mcp-policy-proxy/security_test.go` | Modified | Expand test cases + fix tautological tests |
| `src/mcp-policy-proxy/proxy_handler_test.go` | Modified | Add security integration tests |
| `src/mcp-policy-proxy/proxy.go` | Modified | Add `WithTimeout()` option to CircuitBreaker |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Flaky tests from timing | Medium | Use configurable timeouts; avoid `time.Sleep` in tests |
| Test maintains mock state | Low | Reset mocks per test case |

## Rollback Plan

Revert changes to `security_test.go`, `proxy_handler_test.go`, and `proxy.go` to previous commit. No schema/data changes.

## Dependencies

- `github.com/stretchr/testify/assert` (already in use)
- RSA/EC key generation (stdlib `crypto/rsa`, `crypto/ecdsa`)

## Success Criteria

- [ ] All 8 identified test gaps are closed
- [ ] No `time.Sleep` in security tests (except for actual timeout behavior)
- [ ] JWT tests use real RSA/EC keys, not `[]byte` key type
- [ ] SSRF/FailClosed/BatchSize tests verify handler behavior, not just config
- [ ] `TestSecurity_FailClosedMode` mocks Lakera to return error and verifies 503 response
- [ ] Sanitization tests cover: SQL injection, command injection, XSS, Unicode, case variations
- [ ] Rate limiter tests verify cleanup behavior after TTL
