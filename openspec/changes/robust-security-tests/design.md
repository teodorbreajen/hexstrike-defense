# Design: Robust Security Tests for mcp-policy-proxy

## Technical Approach

Transform unit tests into integration tests by adding mock helpers that control external dependencies (Lakera client, JWT keys), then verify actual handler behavior through `proxy.Handler()` with `httptest.NewRecorder`. This closes 8 identified test gaps without changing production code.

## Architecture Decisions

### Decision: mockLakeraClient Implementation

**Choice**: Internal package struct with controlled response methods
**Alternatives considered**: 
- Interface-based mock (requires changing production interface)
- External package mock (breaks encapsulation)
**Rationale**: Control allows both error and blocking scenarios while keeping production `LakeraClient` interface unchanged. Fail-closed tests require returning error from `CheckToolCall()`.

```go
// Internal to security_test.go package
type mockLakeraClient struct {
    shouldError  bool
    shouldBlock bool
    blockReason string
    blockScore  int
}

func (m *mockLakeraClient) CheckToolCall(ctx context.Context, tool, args string) (bool, int, string, error) {
    if m.shouldError {
        return false, 0, "", fmt.Errorf("mock error")
    }
    return !m.shouldBlock, m.blockScore, m.blockReason, nil
}

func (m *mockLakeraClient) HealthCheck(ctx context.Context) error {
    if m.shouldError {
        return fmt.Errorf("mock health check failed")
    }
    return nil
}

// Control methods
func (m *mockLakeraClient) SetError(err bool) { m.shouldError = err }
func (m *mockLakeraClient) SetBlock(allowed bool, reason string) {
    m.shouldBlock = !allowed
    m.blockReason = reason
}
```

### Decision: RSA/EC Keys for JWT Algorithm Confusion Tests

**Choice**: Generate once in `TestMain()`, store as package variables
**Alternatives considered**: 
- Hardcoded PEM constants (bloats file, 2048-line RSA key)
- Generate per test (wasteful)
**Rationale**: Algorithm confusion tests (RS256, ES256) need real key objects. Generate in `TestMain` to run once.

```go
var (
    testRSAPrivateKey  *rsa.PrivateKey
    testECDSAPrivateKey *ecdsa.PrivateKey
    testRSAPublicKey   *rsa.PublicKey
    testECDSAPublicKey *ecdsa.PublicKey
)

func TestMain(m *testing.M) {
    var err error
    testRSAPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        panic("failed to generate RSA key")
    }
    testRSAPublicKey = &testRSAPrivateKey.PublicKey
    
    testECDSAPrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        panic("failed to generate ECDSA key")
    }
    testECDSAPublicKey = &testECDSAPrivateKey.PublicKey
    
    os.Exit(m.Run())
}
```

### Decision: CircuitBreaker Timeout for Tests

**Choice**: Use `time.Sleep` with explicit timeout in test, not configurable
**Alternatives considered**:
- Option A: Make timeout configurable via `NewCircuitBreakerConfig` (adds new constructor)
- Option B: Inject clock interface (too invasive)
- Option C: Wait in test with documented reason
**Rationale**: Production code already works. Test waits 1 second (documented in spec). This is the simplest approach that doesn't require production code changes. Keep existing `NewCircuitBreaker(threshold, timeout)` signature unchanged.

### Decision: RateLimiter Cleanup Testing

**Choice**: Expose `CleanupTTL()` setter for testing, trigger cleanup via time check in `Allow()`
**Alternatives considered**:
- Force cleanup method (leaks test-only API)
- Short TTL in test config (affects production defaults)
**Rationale**: Test creates rate limiter with 1-minute TTL, waits 90 seconds, verifies cleanup runs. No API changes needed — cleanup already triggers on `now.Sub(lastCleanup) > cleanupTTL`.

### Decision: Test Router with Lakera Mock

**Choice**: Create `createTestRouterWithLakera(proxy, mockLakeraClient)` variant
**Alternatives considered**: Modify existing `createTestRouter` (breaks existing tests)
**Rationale**: New test scenarios need mock Lakera. Existing tests use `nil` Lakera. Keep both variants for backward compatibility.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `src/mcp-policy-proxy/security_test.go` | Modify | Add mockLakeraClient, RSA/EC keys in TestMain, new test cases |
| `src/mcp-policy-proxy/proxy_handler_test.go` | Modify | Add createTestRouterWithLakera, integration tests for SSRF/FailClosed/BatchSize |

## New Test Scenarios

| Test | Gap Addressed | Approach |
|------|---------------|-----------|
| JWT RS256/ES256 rejection | Algorithm confusion | Use generated key objects from TestMain |
| Fail-closed with Lakera error | Missing integration | mockLakeraClient.SetError(true), verify 503 |
| SSRF via handler | Config constant only | External backend URL + proxy.Handler() |
| BatchSize via handler | Config constant only | External backend + proxy.Handler() |
| Sanitization injection vectors | Limited patterns | Table-driven with SQLi, cmd injection, XSS |
| Rate limiter TTL cleanup | Not tested | 90-second wait, verify client count |

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | sanitizeToolInput patterns | Direct function calls, no HTTP |
| Unit | JWT validation logic | Direct `validateJWT()` calls |
| Integration | Fail-closed mode | proxy.Handler() + mockLakeraClient |
| Integration | SSRF protection | proxy.Handler() + internal URL |
| Integration | Batch size limit | proxy.Handler() + oversized batch |
| Integration | Rate limiter cleanup | Create client, wait, verify count |

## Open Questions

- [ ] **CI Timing**: 90-second wait for cleanup test may slow CI. Consider marking as `t.Skip` in short mode.
- [ ] **Key Generation Failure**: If `TestMain` fails to generate keys, tests can't run. Panic is appropriate?
- [ ] **Lakera Interface**: Current `LakeraClient` is a concrete type, not interface. mockLakeraClient works because tests use nil. Consider interface for true dependency injection.

## Next Step

Ready for tasks (sdd-tasks).