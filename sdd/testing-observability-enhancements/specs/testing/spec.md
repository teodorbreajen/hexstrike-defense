# Delta for testing/

## ADDED Requirements

### Requirement: Fuzz Testing in CI Pipeline

The CI pipeline SHALL execute fuzzing tests using go-fuzz on all security-critical functions to detect input handling edge cases and prevent regressions.

The fuzzing tests MUST cover:
- `sanitizeToolInput()` - input sanitization bypass attempts
- `isInternalURL()` - SSRF bypass detection
- `ParseJSONRPC()` - malformed JSON-RPC handling
- `validateBackendURL()` - URL validation edge cases

Fuzzing SHALL run with a corpus of at least 10 known edge cases per function and MUST execute for a minimum of 60 seconds per fuzz target.

#### Scenario: Fuzzing detects regression in sanitization

- GIVEN a fuzz test running against `sanitizeToolInput()`
- WHEN fuzzing generates a malicious input that bypasses current sanitization
- THEN the fuzz test MUST fail with a panic or unexpected error
- AND the CI pipeline SHALL block the merge

#### Scenario: Fuzzing runs on schedule

- GIVEN the CI pipeline triggers on push to main or PR
- WHEN fuzz targets are compiled with `go-fuzz-build`
- THEN the fuzzing process MUST run for at least 60 seconds
- AND a corpus MUST be generated and persisted for future runs

---

### Requirement: Integration Tests with Mock MCP Server

The project SHALL include integration tests that exercise the full proxy request flow using a real HTTP mock server instead of mocked interfaces.

The integration tests MUST verify:
- End-to-end request routing from client → proxy → mock MCP backend
- JSON-RPC request/response handling through the full stack
- Metrics collection accuracy across the full request lifecycle
- Error propagation from backend through proxy to client

#### Scenario: Integration test with mock backend

- GIVEN a mock MCP server running on `localhost:9090`
- WHEN the proxy receives a valid JSON-RPC tools/list request
- THEN the proxy MUST forward to mock server and return its response
- AND metrics MUST increment `total_requests` and `allowed_requests`

#### Scenario: Integration test validates metrics accuracy

- GIVEN the proxy and mock server are running
- WHEN 5 requests are made through the proxy
- THEN `/metrics` endpoint MUST report `total_requests: 5`
- AND status code counters MUST match actual responses

## MODIFIED Requirements

*(None)*

## REMOVED Requirements

*(None)*
