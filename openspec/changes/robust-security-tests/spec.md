# Delta Spec: Robust Security Tests for mcp-policy-proxy

## Purpose

Transform unit tests into integration tests that verify actual security enforcement behavior, closing gaps in attack vector coverage (SQL injection, command injection, SSRF) and removing fragile timing dependencies.

## ADDED Requirements

### Requirement: JWT Algorithm Confusion Prevention

The system SHALL reject JWT tokens signed with asymmetric algorithms (RS256, ES256, PS256, EdDSA) even if signed with the configured HMAC secret.

#### Scenario: RS256 token rejected

- GIVEN a proxy configured with JWTSecret="test-secret" and FailMode="open"
- WHEN a request is made to `/mcp/v1/tools` with Authorization: Bearer {RS256_TOKEN}
- THEN the response SHALL be 401 Unauthorized with message "invalid signing algorithm"
- AND no request SHALL be forwarded to the MCP backend

#### Scenario: ES256 token rejected

- GIVEN a proxy configured with JWTSecret="test-secret" and FailMode="open"
- WHEN a request is made to `/mcp/v1/tools` with Authorization: Bearer {ES256_TOKEN}
- THEN the response SHALL be 401 Unauthorized with message "invalid signing algorithm"
- AND no request SHALL be forwarded to the MCP backend

#### Scenario: "alg: none" token rejected

- GIVEN a proxy configured with JWTSecret="test-secret" and FailMode="open"
- WHEN a request is made to `/mcp/v1/tools` with Authorization: Bearer {ALG_NONE_TOKEN}
- THEN the response SHALL be 401 Unauthorized
- AND no request SHALL be forwarded to the MCP backend

#### Scenario: HS256 with wrong secret rejected

- GIVEN a proxy configured with JWTSecret="correct-secret" and FailMode="open"
- WHEN a request is made to `/mcp/v1/tools` with Authorization: Bearer {HS256_WRONG_SECRET_TOKEN}
- THEN the response SHALL be 401 Unauthorized with message containing "signature"
- AND no request SHALL be forwarded to the MCP backend

---

### Requirement: Fail-Closed Mode with Lakera Error

When FailMode is "closed" and the Lakera client returns an error, the system SHALL block the request with 503 Service Unavailable.

#### Scenario: Fail-closed blocks when Lakera returns error

- GIVEN a proxy configured with FailMode="closed", lakeraClient returning error
- WHEN POST /mcp/v1/call is made with valid JSON-RPC tool call body
- THEN the response SHALL be 503 Service Unavailable
- AND response body SHALL contain "Semantic check unavailable" or "blocked for security"
- AND the request SHALL NOT be forwarded to the MCP backend

#### Scenario: Fail-open allows when Lakera returns error

- GIVEN a proxy configured with FailMode="open", lakeraClient returning error
- WHEN POST /mcp/v1/call is made with valid JSON-RPC tool call body
- THEN the response SHALL be 502 Bad Gateway (backend unavailable)
- OR 200 OK (if mock backend responds)
- AND the request SHALL be forwarded to the MCP backend

---

### Requirement: SSRF Protection via Handler Integration

The system SHALL validate backend URLs at handler level, blocking internal URLs before any network request is made.

#### Scenario: SSRF blocked at handler level - localhost

- GIVEN a proxy configured with MCPBackendURL="http://localhost:9090" and valid JWT
- WHEN GET /mcp/v1/tools is made with valid Authorization header
- THEN the response SHALL be 500 Internal Server Error
- AND response body SHALL contain "Invalid backend configuration" or "internal URLs not allowed"
- AND NO actual HTTP request SHALL be made to localhost:9090

#### Scenario: SSRF blocked at handler level - 127.0.0.1

- GIVEN a proxy configured with MCPBackendURL="http://127.0.0.1:9090" and valid JWT
- WHEN GET /mcp/v1/tools is made with valid Authorization header
- THEN the response SHALL be 500 Internal Server Error
- AND NO actual HTTP request SHALL be made to 127.0.0.1:9090

#### Scenario: SSRF blocked at handler level - private range

- GIVEN a proxy configured with MCPBackendURL="http://10.0.0.1:9090" and valid JWT
- WHEN GET /mcp/v1/tools is made with valid Authorization header
- THEN the response SHALL be 500 Internal Server Error
- AND NO actual HTTP request SHALL be made to 10.0.0.1:9090

---

### Requirement: Batch Size Limit via Handler

The system SHALL enforce MaxBatchSize limit at handler level, returning 400 for oversized batches.

#### Scenario: Batch size 11 rejected with 400

- GIVEN a proxy configured with MaxBatchSize=10, valid JWT, external MCPBackendURL
- WHEN POST /mcp/v1/call is made with body containing array of 11 JSON-RPC requests
- THEN the response SHALL be 400 Bad Request
- AND response body SHALL contain "exceeds maximum" or the batch size limit value
- AND the request SHALL NOT be forwarded to the MCP backend

#### Scenario: Batch size exactly at limit accepted

- GIVEN a proxy configured with MaxBatchSize=10, valid JWT, external MCPBackendURL
- WHEN POST /mcp/v1/call is made with body containing array of 10 JSON-RPC requests
- THEN the response SHALL be 200 OK or 502 Bad Gateway (backend unavailable)
- AND the request SHALL be forwarded to the MCP backend

---

### Requirement: Input Sanitization - Injection Vectors

The system SHALL reject tool inputs containing SQL injection, command injection, XSS, Unicode bypass, and case variation attack patterns.

#### Scenario: SQL injection patterns rejected

- GIVEN a proxy configured with valid JWT, lakeraClient returning allowed=true
- WHEN POST /mcp/v1/call is made with tool name "bash" and args containing "'; DROP TABLE users; --"
- THEN the response SHALL be 400 Bad Request
- AND response body SHALL contain "suspicious" or "invalid" or "command injection"

#### Scenario: Command injection - $() syntax rejected

- GIVEN a proxy configured with valid JWT, lakeraClient returning allowed=true
- WHEN POST /mcp/v1/call is made with tool name "bash" and args containing "$(whoami)"
- THEN the response SHALL be 400 Bad Request
- AND response body SHALL contain "suspicious" or "command injection"

#### Scenario: Command injection - backticks rejected

- GIVEN a proxy configured with valid JWT, lakeraClient returning allowed=true
- WHEN POST /mcp/v1/call is made with tool name "bash" and args containing "`id`"
- THEN the response SHALL be 400 Bad Request
- AND response body SHALL contain "suspicious" or "command injection"

#### Scenario: Command injection - ${} syntax rejected

- GIVEN a proxy configured with valid JWT, lakeraClient returning allowed=true
- WHEN POST /mcp/v1/call is made with tool name "bash" and args containing "${HOME}"
- THEN the response SHALL be 400 Bad Request
- AND response body SHALL contain "suspicious" or "command injection"

#### Scenario: XSS payload in args rejected

- GIVEN a proxy configured with valid JWT, lakeraClient returning allowed=true
- WHEN POST /mcp/v1/call is made with tool name "bash" and args containing "<script>alert(1)</script>"
- THEN the response SHALL be 400 Bad Request
- AND response body SHALL contain appropriate validation error

#### Scenario: Unicode emoji in tool name rejected

- GIVEN a proxy configured with valid JWT, lakeraClient returning allowed=true
- WHEN POST /mcp/v1/call is made with tool name containing emoji flags (e.g., "🇦🇷bash")
- THEN the response SHALL be 400 Bad Request
- AND response body SHALL contain "invalid character"

#### Scenario: Case variations of dangerous commands blocked

- GIVEN a proxy configured with valid JWT, lakeraClient returning allowed=true
- WHEN POST /mcp/v1/call is made with tool name "Bash", "BASH", "BaSh" (case variations)
- THEN the system SHALL handle case-insensitively OR block based on allowlist

#### Scenario: Null byte variations rejected

- GIVEN a proxy configured with valid JWT, lakeraClient returning allowed=true
- WHEN POST /mcp/v1/call is made with args containing "bash\x00" or "bash%00" or "bash\u0000"
- THEN the response SHALL be 400 Bad Request
- AND response body SHALL contain "null byte" or "invalid"

---

### Requirement: Lakera Nil Fallback Behavior

When lakeraClient is nil, the system SHALL use the configured FailMode without crashing.

#### Scenario: Nil lakeraClient with fail-open allows request

- GIVEN a proxy created with NewProxy(config, nil) where FailMode="open"
- WHEN POST /mcp/v1/call is made with valid JSON-RPC tool call
- THEN the system SHALL NOT panic
- AND the request SHALL be forwarded to the MCP backend (fail-open behavior)

#### Scenario: Nil lakeraClient with fail-closed blocks request

- GIVEN a proxy created with NewProxy(config, nil) where FailMode="closed"
- WHEN POST /mcp/v1/call is made with valid JSON-RPC tool call
- THEN the system SHALL NOT panic
- AND the response SHALL be 503 Service Unavailable

---

### Requirement: CircuitBreaker Timeout Configurable

The CircuitBreaker SHALL accept configurable timeout for testing purposes without using time.Sleep.

#### Scenario: CircuitBreaker with zero timeout for instant testing

- GIVEN a CircuitBreaker created with NewCircuitBreaker(3, 0) or WithTimeout(0)
- WHEN RecordFailure() is called 3 times
- THEN the circuit SHALL immediately transition to Open state
- AND Allow() SHALL return false immediately

#### Scenario: CircuitBreaker recovery with configurable timeout

- GIVEN a CircuitBreaker created with NewCircuitBreakerConfig(threshold=3, timeout=100ms)
- WHEN RecordFailure() is called 3 times
- THEN after timeout duration, Allow() SHALL return true
- AND state SHALL be HalfOpen

---

### Requirement: RateLimiter TTL/Cleanup Behavior

The ClientRateLimiter SHALL automatically cleanup stale clients after TTL expires and enforce maxClients limit.

#### Scenario: Client cleanup after TTL expires

- GIVEN a ClientRateLimiter with cleanupTTL=1 minute
- WHEN client "test-client" makes a request
- AND then no requests are made for TTL + cleanupInterval
- THEN GetClientCount() SHALL return 0 or decreased count
- AND the stale client SHALL be removed from tracking

#### Scenario: MaxClients limit enforced

- GIVEN a ClientRateLimiter with maxClients=100
- WHEN more than 100 unique clients make requests
- THEN the rate limiter SHALL either reject new clients or cleanup oldest
- AND GetClientCount() SHALL NOT exceed maxClients significantly

#### Scenario: Cleanup interval processes stale clients

- GIVEN a ClientRateLimiter with cleanupTTL=1 minute, cleanupInterval=30 seconds
- WHEN client "stale-client" makes a request
- AND 90 seconds pass with no requests from that client
- THEN the next Allow() call SHALL trigger cleanup
- AND "stale-client" SHALL be removed from tracking

---

## MODIFIED Requirements

### Requirement: Testable Security Constants

The MaxBatchSize, MaxToolNameLength, and other security constants SHALL be tested via integration tests that verify handler behavior, not just constant values.

(Previously: Tests only checked `MaxBatchSize > batchSize` without calling actual handler)

#### Scenario: Batch size validation via handler

- GIVEN a proxy with valid JWT and external MCPBackendURL
- WHEN POST /mcp/v1/call is made with oversized batch
- THEN the handler SHALL return 400 BEFORE parsing individual requests
- AND the response SHALL mention the limit

---

## Edge Cases Covered

| Test | Edge Case | Expected Behavior |
|------|-----------|-------------------|
| JWT RS256 | Token signed with RSA private key | 401 - algorithm mismatch |
| JWT ES256 | Token signed with EC private key | 401 - algorithm mismatch |
| FailClosed | Lakera timeout (5s) + fail-closed | 503 after timeout |
| FailClosed | Lakera returns nil allowed | Uses FailMode |
| SSRF | IPv6 localhost (::1) | Blocked |
| SSRF | Hex-encoded IP (0x7f000001) | Blocked |
| SSRF | DNS rebinding to internal | Blocked at URL validation |
| BatchSize | Empty batch [] | Accepted |
| BatchSize | Single item [req] | Accepted |
| Sanitization | Empty tool name | 400 - empty/whitespace |
| Sanitization | Unicode normalization (café vs café) | Depends on validation |
| Sanitization | Right-to-left override (U+202E) | Should be blocked |
| RateLimiter | Concurrent requests from same IP | Shared bucket |
| RateLimiter | IP with port (X-Forwarded-For chain) | First IP used |

---

## Test Implementation Notes

### mockLakeraClient Helper

```go
type mockLakeraClient struct {
    allowed bool
    score   int
    reason  string
    err     error
}

func (m *mockLakeraClient) CheckToolCall(ctx context.Context, tool, args string) (bool, int, string, error) {
    return m.allowed, m.score, m.reason, m.err
}
```

### RSA/ECDSA Key Generation in TestMain

```go
var (
    testRSAPrivateKey *rsa.PrivateKey
    testECDSAPrivateKey *ecdsa.PrivateKey
)

func TestMain(m *testing.M) {
    // Generate once, reuse for all tests
    testRSAPrivateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
    testECDSAPrivateKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    os.Exit(m.Run())
}
```

### No time.Sleep in Security Tests

All timing-dependent tests SHALL use:
- Configurable timeouts (e.g., NewCircuitBreakerConfig with 0 timeout)
- Mock time or test clock (if available)
- Explicit state verification instead of waiting

## Success Criteria Met

- [x] All 8 identified test gaps specified with scenarios
- [x] No time.Sleep in spec (configurable timeouts used)
- [x] JWT tests use real RSA/EC key generation
- [x] SSRF/FailClosed/BatchSize tests verify handler behavior
- [x] Sanitization tests cover: SQL injection, command injection, XSS, Unicode, case variations
- [x] Rate limiter tests verify cleanup behavior after TTL
