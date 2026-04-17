# Security Spec — Delta

## ADDED Requirements

### Requirement: JWT_SECRET Startup Validation

The system MUST validate that `JWT_SECRET` environment variable is configured at application startup. If not set, the application MUST fail immediately with a clear error message.

**Behavior:**
- On startup, check if `JWT_SECRET` is non-empty
- If empty, log fatal error: "JWT_SECRET environment variable is required but not set"
- Exit with exit code 1
- This prevents the application from running in an insecure state

**Environment Variable:**
- `JWT_SECRET` (required when `AUTH_ENABLED=true`, default behavior)

#### Scenario: JWT_SECRET is set at startup

- GIVEN `JWT_SECRET` environment variable is set to a non-empty value
- WHEN application starts
- THEN startup SHALL proceed normally
- AND JWT validation middleware SHALL be initialized

#### Scenario: JWT_SECRET is missing at startup

- GIVEN `JWT_SECRET` environment variable is not set or empty
- WHEN application starts
- THEN application SHALL log fatal error: "JWT_SECRET environment variable is required but not set"
- AND application SHALL exit with code 1
- AND no HTTP server SHALL start

### Requirement: CORS Configurable Allowlist

The system MUST support configurable CORS origins via environment variable. Only explicitly allowed origins SHALL receive CORS headers.

**Environment Variable:**
- `CORS_ORIGINS` — comma-separated list of allowed origins (e.g., `https://app.example.com,https://admin.example.com`)

**Behavior:**
- If `CORS_ORIGINS` is empty, CORS SHALL be disabled (no headers returned)
- Preflight requests (OPTIONS) SHALL be rejected from non-allowed origins
- Credentials support SHALL be configurable via `CORS_ALLOW_CREDENTIALS`

#### Scenario: Valid origin receives CORS headers

- GIVEN `CORS_ORIGINS=https://app.example.com`
- WHEN a request arrives with `Origin: https://app.example.com`
- THEN response SHALL include headers:
  - `Access-Control-Allow-Origin: https://app.example.com`
  - `Access-Control-Allow-Methods: GET, POST, OPTIONS`
  - `Access-Control-Allow-Headers: Content-Type, Authorization`

#### Scenario: Invalid origin is rejected

- GIVEN `CORS_ORIGINS=https://app.example.com`
- WHEN a request arrives with `Origin: https://evil.com`
- THEN response SHALL NOT include `Access-Control-Allow-Origin`
- AND preflight OPTIONS request SHALL return 403

#### Scenario: CORS disabled when no origins configured

- GIVEN `CORS_ORIGINS` is not set or empty
- WHEN any request arrives
- THEN no CORS headers SHALL be included in response
- AND preflight requests SHALL be passed through without headers

#### Scenario: Preflight OPTIONS request from allowed origin

- GIVEN `CORS_ORIGINS=https://app.example.com`
- WHEN OPTIONS request arrives with `Origin: https://app.example.com`
- THEN response SHALL be HTTP 204
- AND include CORS headers without forwarding to backend
