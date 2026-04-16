# Testing Specification (Delta)

## Purpose

Define unit test requirements for MCP Policy Proxy to ensure reliability and security hardening correctness.

## ADDED Requirements

### Requirement: Proxy HTTP Handler Tests

The system SHALL include unit tests for all HTTP handlers using `net/http/httptest`.

#### Test File: `proxy_test.go`

**Scenario: Health endpoint returns healthy status**

- GIVEN Lakera client is configured and reachable
- WHEN GET `/health` is called via httptest
- THEN response status SHALL be 200
- AND response body SHALL contain `"status":"healthy"`

**Scenario: Health endpoint returns degraded when Lakera unavailable**

- GIVEN Lakera client returns error on health check
- WHEN GET `/health` is called via httptest
- THEN response status SHALL be 200
- AND response body SHALL contain `"status":"degraded"`

**Scenario: Metrics endpoint returns correct JSON**

- GIVEN a proxy with recorded metrics (10 total, 2 blocked, 8 allowed)
- WHEN GET `/metrics` is called via httptest
- THEN response status SHALL be 200
- AND response body SHALL contain `"total_requests":10`
- AND response body SHALL contain `"blocked_requests":2`
- AND response body SHALL contain `"allowed_requests":8`

**Scenario: Readiness endpoint returns ready**

- GIVEN a running proxy instance
- WHEN GET `/ready` is called via httptest
- THEN response status SHALL be 200
- AND response body SHALL contain `"status":"ready"`

**Scenario: Protected endpoint returns 401 without auth**

- GIVEN auth is enabled in configuration
- WHEN POST `/mcp/tools/call` is called without Authorization header
- THEN response status SHALL be 401
- AND response body SHALL contain error code 401

**Scenario: Protected endpoint returns 200 with valid auth**

- GIVEN auth is enabled with valid JWT secret
- AND valid JWT token is provided
- WHEN POST `/mcp/tools/call` is called
- THEN request SHALL pass authentication
- AND Lakera check SHALL be performed

### Requirement: Rate Limiter Tests

The system SHALL include unit tests for rate limiter token bucket logic.

#### Test File: `rate_limiter_test.go`

**Scenario: First request is allowed**

- GIVEN a rate limiter initialized with 10 tokens
- WHEN `Allow()` is called
- THEN it SHALL return `true`
- AND remaining tokens SHALL be 9

**Scenario: Requests exhaust tokens**

- GIVEN a rate limiter initialized with 2 tokens
- WHEN `Allow()` is called 3 times consecutively
- THEN first 2 calls SHALL return `true`
- AND third call SHALL return `false`

**Scenario: Tokens refill after duration**

- GIVEN a rate limiter initialized with 1 token that refills every minute
- WHEN `Allow()` is called, exhausting tokens
- WHEN 60+ seconds elapse
- WHEN `Allow()` is called again
- THEN it SHALL return `true`

**Scenario: Concurrent requests are thread-safe**

- GIVEN a rate limiter initialized with 100 tokens
- WHEN 50 goroutines each call `Allow()` simultaneously
- THEN exactly 50 calls SHALL return `true`
- AND exactly 50 calls SHALL return `false`

### Requirement: Metrics Tests

The system SHALL include unit tests for metrics recording and retrieval.

#### Test File: `metrics_test.go`

**Scenario: RecordRequest increments counters correctly**

- GIVEN a new Metrics instance
- WHEN `RecordRequest(true, 100ms, 200)` is called
- THEN `GetMetrics()` SHALL return total=1, blocked=0, allowed=1
- AND average latency SHALL be approximately 100ms

**Scenario: Blocked requests increment LakeraBlockCount**

- GIVEN a new Metrics instance
- WHEN `RecordRequest(false, 0, 403)` is called (blocked request)
- THEN `GetMetrics()` SHALL return blocked=1
- AND `LakeraBlockCount` SHALL be 1

**Scenario: Status codes are tracked**

- GIVEN a new Metrics instance
- WHEN multiple requests return various status codes (200, 200, 403, 500)
- THEN `GetMetrics()` status codes SHALL contain `{200: 2, 403: 1, 500: 1}`

**Scenario: Average latency calculation**

- GIVEN a Metrics instance that recorded 3 requests with latencies 100ms, 200ms, 300ms
- WHEN `GetMetrics()` is called
- THEN avgLatency SHALL be 200ms (sum / count / ms conversion)

### Requirement: Lakera Client Tests

The system SHALL include unit tests for Lakera client using httptest.

#### Test File: `lakera_test.go`

**Scenario: CheckToolCall returns allowed on successful response**

- GIVEN a test server returning `{"score": 30, "verdict": "safe", "reasons": []}`
- WHEN `LakeraClient.CheckToolCall(ctx, "test_tool", "{}")` is called
- THEN it SHALL return `(true, 30, "", nil)`

**Scenario: CheckToolCall blocks high score**

- GIVEN a test server returning `{"score": 85, "verdict": "blocked", "reasons": ["high risk"]}`
- WHEN `LakeraClient.CheckToolCall(ctx, "test_tool", "{}")` is called
- THEN it SHALL return `(false, 85, "high risk", nil)`

**Scenario: CheckToolCall returns error on server failure**

- GIVEN a test server returning HTTP 500
- WHEN `LakeraClient.CheckToolCall(ctx, "test_tool", "{}")` is called
- THEN it SHALL return error
- AND based on fail mode, may return `(false, 0, "", error)` or `(true, 0, "", error)`

**Scenario: CheckToolCall returns error on timeout**

- GIVEN a test server that hangs (no response)
- WHEN `LakeraClient.CheckToolCall()` is called with short timeout
- THEN it SHALL return context deadline exceeded error

**Scenario: Empty API key allows all requests**

- GIVEN LakeraClient initialized without API key
- WHEN `CheckToolCall(ctx, "test", "{}")` is called
- THEN it SHALL return `(true, 0, "API key not configured", nil)`
- AND no HTTP request SHALL be made

### Requirement: JSON-RPC Parsing Tests

The system SHALL include unit tests for JSON-RPC request parsing.

#### Test File: `jsonrpc_test.go`

**Scenario: Parse valid tool call request**

- GIVEN JSON `{"jsonrpc":"2.0","method":"tools/call","params":{"name":"test","arguments":{}},"id":1}`
- WHEN `ParseJSONRPC()` is called
- THEN returned ParsedRequest SHALL have Method="tools/call", ToolName="test"

**Scenario: Parse batch request**

- GIVEN JSON array of valid JSON-RPC requests
- WHEN `ParseJSONRPC()` is called
- THEN returned ParsedRequest SHALL have `IsBatch=true`
- AND `BatchReqs` SHALL contain all valid requests

**Scenario: Reject invalid JSON**

- GIVEN invalid JSON string
- WHEN `ParseJSONRPC()` is called
- THEN it SHALL return error with ParseErrorCode

**Scenario: Reject wrong JSON-RPC version**

- GIVEN JSON with `"jsonrpc":"1.0"`
- WHEN `ParseJSONRPC()` is called
- THEN it SHALL return error with InvalidRequestCode

### Requirement: Test Coverage

The system SHOULD achieve minimum 70% code coverage for:
- `proxy.go`: HTTP handlers, middleware chain
- `rate_limiter.go`: Token bucket logic
- `metrics.go`: Recording and aggregation
- `lakera.go`: API calls and error handling
- `jsonrpc.go`: Request parsing

Tests SHALL run in CI pipeline and MUST pass before merge.
