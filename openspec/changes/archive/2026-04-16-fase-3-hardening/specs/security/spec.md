# Security Specification (Delta)

## MODIFIED Requirements

### Requirement: Lakera Fail-Closed Behavior

**Previous**: When Lakera API returns an error, the proxy logs the error and allows the request (fail-open).  
**Current**: The system MUST block requests when Lakera is unavailable or returns an error.

The proxy SHALL return HTTP 503 (Service Unavailable) with message "Security service unavailable - request blocked" when:
- Lakera API returns any error (network failure, timeout, HTTP 5xx)
- Lakera API returns non-200 status code
- Lakera client `handleError()` is invoked

The system SHALL log the failure with ERROR level for investigation.

#### Scenario: Lakera API returns error on tool call

- GIVEN a JSON-RPC tool call request to `/mcp/*`
- WHEN `LakeraClient.CheckToolCall()` returns an error
- THEN the request SHALL be blocked with HTTP 503
- AND response body SHALL contain `{"error": {"code": 503, "message": "Security service unavailable - request blocked"}}`
- AND the error SHALL be logged at ERROR level with correlation ID

#### Scenario: Lakera API timeout

- GIVEN a JSON-RPC tool call request to `/mcp/*`
- WHEN the request to Lakera exceeds the configured timeout
- THEN the request SHALL be blocked with HTTP 503
- AND log entry SHALL include "Lakera timeout"

#### Scenario: Lakera returns non-200 status

- GIVEN a JSON-RPC tool call request to `/mcp/*`
- WHEN Lakera API returns HTTP status 500, 502, 503, or 504
- THEN the request SHALL be blocked with HTTP 503
- AND the HTTP status code SHALL be logged

### Requirement: Authentication Middleware

**Previous**: `authMiddleware` exists but passes all requests through without validation.  
**Current**: The system MUST validate Authorization header for protected endpoints.

The system SHALL validate the `Authorization` header for all `/mcp/*` endpoints:
- The header MUST be present and non-empty
- The header MUST follow `Bearer <token>` format
- The token MUST be a valid JWT signed with the configured secret

Protected endpoints:
- All paths matching `/mcp/*` pattern

Unprotected endpoints (no auth required):
- `/health`
- `/metrics`
- `/ready`

#### Scenario: Valid JWT token provided

- GIVEN a request to `/mcp/tools/call` with valid JWT Bearer token
- WHEN `Authorization: Bearer <valid_jwt>` header is present
- THEN the request SHALL pass authentication
- AND request SHALL continue to Lakera semantic check

#### Scenario: Missing Authorization header on protected endpoint

- GIVEN a request to `/mcp/tools/call` without Authorization header
- THEN the request SHALL be rejected with HTTP 401
- AND response body SHALL contain `{"error": {"code": 401, "message": "Authorization required"}}`

#### Scenario: Invalid JWT token format

- GIVEN a request to `/mcp/tools/call` with malformed token
- WHEN `Authorization` header is present but not in `Bearer <token>` format
- THEN the request SHALL be rejected with HTTP 401
- AND response body SHALL contain `{"error": {"code": 401, "message": "Invalid authorization format"}}`

#### Scenario: Expired or invalid JWT

- GIVEN a request to `/mcp/tools/call` with expired JWT
- WHEN JWT validation fails (signature mismatch, expired, malformed)
- THEN the request SHALL be rejected with HTTP 401
- AND response body SHALL contain `{"error": {"code": 401, "message": "Invalid or expired token"}}`

#### Scenario: Health endpoint without auth

- GIVEN a request to `/health` without Authorization header
- THEN the request SHALL be allowed without authentication
- AND health check SHALL complete normally

### Requirement: Input Validation - Body Size Limit

**Previous**: `io.ReadAll(r.Body)` reads entire body without limit, vulnerable to DoS.  
**Current**: The system MUST reject requests with body exceeding 1MB.

The system SHALL:
- Reject requests with `Content-Length` header exceeding 1MB (1,048,576 bytes)
- Return HTTP 413 (Request Entity Too Large) if body exceeds limit
- Add `MaxBodySize` configuration option (default: 1MB)
- Check `Content-Length` before reading body to avoid memory exhaustion

#### Scenario: Request body within size limit

- GIVEN a POST request to `/mcp/tools/call` with body size 500KB
- WHEN body is read and processed
- THEN the request SHALL proceed normally
- AND Lakera semantic check SHALL be performed

#### Scenario: Request body exceeds 1MB

- GIVEN a POST request to `/mcp/tools/call` with body size exceeding 1MB
- WHEN the request body is being read or `Content-Length` header is checked
- THEN the request SHALL be rejected with HTTP 413
- AND response body SHALL contain `{"error": {"code": 413, "message": "Request body exceeds maximum size of 1MB"}}`

#### Scenario: Content-Length header indicates oversized body

- GIVEN a POST request with `Content-Length: 2097152` (2MB)
- WHEN middleware checks `Content-Length` before reading body
- THEN the request SHALL be rejected with HTTP 413
- AND body SHALL NOT be read into memory

### Requirement: Configurable Fail Mode

The system SHALL support environment variable `LAKERA_FAIL_MODE`:
- Value `closed` (default): Block on Lakera error
- Value `open`: Allow on Lakera error (backward compatible)

This allows operators to choose fail behavior based on security vs availability requirements.

#### Scenario: Fail mode set to open

- GIVEN environment variable `LAKERA_FAIL_MODE=open`
- WHEN Lakera API returns an error
- THEN the request SHALL be allowed (graceful degradation)
- AND error SHALL be logged at WARN level

#### Scenario: Fail mode set to closed (default)

- GIVEN environment variable `LAKERA_FAIL_MODE=closed` or not set
- WHEN Lakera API returns an error
- THEN the request SHALL be blocked with HTTP 503
- AND error SHALL be logged at ERROR level
