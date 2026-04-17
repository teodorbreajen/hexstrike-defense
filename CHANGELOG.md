# Changelog - HexStrike Defense

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [2.0.0] - 2026-04-17

### Breaking Changes

- **Minimum Go version**: Now requires Go 1.21+
- **JWT Authentication**: Authentication is now enforced by default in production (GIN_MODE=release)
- **Fail-Closed by Default**: Lakera API errors now block requests by default (HTTP 503) instead of allowing them

### Added

#### Security Hardening (Phase 3)

- **Fail-Closed Behavior**: Lakera API errors now block requests by default (HTTP 503)
  - Configurable via `LAKERA_FAIL_MODE` environment variable
  - Default: `closed` (secure by default)
  - Set to `open` for backward compatibility

- **JWT Authentication**: Bearer token validation for protected endpoints
  - Validates `Authorization: Bearer <token>` header
  - Protected endpoints: `/mcp/*`
  - Unprotected: `/health`, `/metrics`, `/ready`
  - Returns HTTP 401 for invalid/missing tokens

- **Request Body Size Limit**: Maximum 1MB per request
  - Configurable via `MAX_BODY_SIZE` environment variable
  - Returns HTTP 413 if exceeded
  - Checks Content-Length before reading body

- **JWT Secret Configuration**: 
  - Configurable via `JWT_SECRET` environment variable
  - Required in production (GIN_MODE=release)
  - Logs warning if empty (development mode)

#### Resilience & Observability

- **Circuit Breaker Pattern**: Prevents cascade failures when backend is unhealthy
- **Dead Letter Queue (DLQ)**: Failed requests stored for later replay
- **Retry with Exponential Backoff**: Automatic retries with backoff (1s, 2s, 4s)
- **Connection Pooling**: Reusable HTTP connections for better performance
- **Concurrent Request Limiting**: Max 100 concurrent requests per instance
- **Per-Client Rate Limiting**: Token bucket per client IP

- **Structured JSON Logging** (`logger.go`)
  - SIEM-compatible JSON format
  - Fields: timestamp, level, correlation_id, component, message
  - Optional: request_id, method, path, status_code, latency_ms, error

- **Correlation ID Support**
  - UUID v4 per request
  - Propagated via X-Correlation-ID header
  - Stored in request context for logging

#### Comprehensive Testing

- **Unit Tests for Rate Limiter** (`rate_limiter_test.go`)
- **Unit Tests for Metrics** (`metrics_test.go`)
- **Unit Tests for Proxy Handlers** (`proxy_handler_test.go`)
- **Unit Tests for Logger** (`logger_test.go`)
- **Unit Tests for CORS** (`cors_test.go`)
- **Integration Tests** (`integration_test.go`)
- **Security Tests** (`security_test.go`)
- **Race Condition Tests** (`race_test.go`)
- **Fuzz Tests** (`fuzz_test.go`)
- **Prometheus Metrics Tests** (`prometheus_test.go`)
- **Comprehensive Security Tests** (`security_comprehensive_test.go`)
- **Total: 39+ tests, all passing**

#### DevOps

- **Dockerfile** (multi-stage build)
  - Builder: golang:1.21-alpine
  - Runtime: alpine:3.19
  - Non-root user (UID 1000)
  - Health check endpoint

- **Makefile**
  - `make build`: Compile binary
  - `make test`: Run unit tests
  - `make docker-build`: Build Docker image
  - `make docker-run`: Run container
  - `make lint`: Run go vet
  - `make clean`: Remove artifacts

#### Documentation

- **Incident Runbook** (`docs/runbook.md`)
- **Phase 3 Technical Report** (`docs/PHASE3-HARDENING-REPORT.md`)
- **Test Results** (`docs/TEST-RESULTS.md`)

### Changed

- **Security defaults**: Changed from fail-open to fail-closed
- **Authentication**: Added enforcement (was stub/disabled)
- **Logging**: Changed from plain text to JSON structured
- **Listening address**: Default changed from `0.0.0.0:8080` to `127.0.0.1:8080` for security

### Dependencies

```
github.com/golang-jwt/jwt/v5    v5.2.1    - JWT validation
github.com/stretchr/testify     v1.9.0    - Testing assertions
github.com/google/uuid         v1.6.0    - UUID generation
```

---

## [1.3.0] - 2026-04-16

### Added

#### Security Hardening (Phase 3)

- **Fail-Closed Behavior**: Lakera API errors now block requests by default (HTTP 503)
  - Configurable via `LAKERA_FAIL_MODE` environment variable
  - Default: `closed` (secure by default)
  - Set to `open` for backward compatibility

- **JWT Authentication**: Bearer token validation for protected endpoints
  - Validates `Authorization: Bearer <token>` header
  - Protected endpoints: `/mcp/*`
  - Unprotected: `/health`, `/metrics`, `/ready`
  - Returns HTTP 401 for invalid/missing tokens

- **Request Body Size Limit**: Maximum 1MB per request
  - Configurable via `MAX_BODY_SIZE` environment variable
  - Returns HTTP 413 if exceeded
  - Checks Content-Length before reading body

- **JWT Secret Configuration**: 
  - Configurable via `JWT_SECRET` environment variable
  - Skip validation if empty (development mode)

#### Testing

- **Unit Tests for Rate Limiter** (`rate_limiter_test.go`)
  - Token bucket algorithm tests
  - Concurrent access tests
  - Refill behavior tests

- **Unit Tests for Metrics** (`metrics_test.go`)
  - Counter increment tests
  - Latency calculation tests
  - Status code tracking tests

- **Unit Tests for Proxy Handlers** (`proxy_handler_test.go`)
  - Health endpoint tests
  - Authentication tests
  - Body size limit tests
  - Rate limiting tests

- **Logger Tests** (`logger_test.go`)
  - JSON structured logging tests
  - Correlation ID tests
  - Level filtering tests

- **Total: 39 tests, all passing**

#### DevOps

- **Dockerfile** (multi-stage build)
  - Builder: golang:1.21-alpine
  - Runtime: alpine:3.19
  - Non-root user (UID 1000)
  - Health check endpoint
  - Multi-stage for minimal image size

- **Makefile**
  - `make build`: Compile binary
  - `make test`: Run unit tests
  - `make docker-build`: Build Docker image
  - `make docker-run`: Run container
  - `make lint`: Run go vet
  - `make clean`: Remove artifacts

#### Observability

- **Structured JSON Logging** (`logger.go`)
  - SIEM-compatible JSON format
  - Fields: timestamp, level, correlation_id, component, message
  - Optional: request_id, method, path, status_code, latency_ms, error

- **Correlation ID Support**
  - UUID v4 per request
  - Propagated via X-Correlation-ID header
  - Stored in request context for logging

#### Documentation

- **Incident Runbook** (`docs/runbook.md`)
  - Lakera outage response procedures
  - Falco alert response procedures
  - Rate limit exceeded response
  - General troubleshooting guide
  - Configuration reference

- **Phase 3 Technical Report** (`docs/PHASE3-HARDENING-REPORT.md`)
  - Detailed security fixes documentation
  - Configuration reference
  - Verification results

- **Test Results** (`docs/TEST-RESULTS.md`)
  - Complete test execution report
  - Security test scenarios
  - Coverage analysis

---

### Changed

- **Security defaults**: Changed from fail-open to fail-closed
- **Authentication**: Added enforcement (was stub/disabled)
- **Logging**: Changed from plain text to JSON structured

---

### Dependencies Added

```
github.com/golang-jwt/jwt/v5    v5.2.1    - JWT validation
github.com/stretchr/testify    v1.8.4    - Testing assertions
github.com/google/uuid        v1.6.0    - UUID generation
```

---

## [1.2.0] - 2026-04-15

### Added

- Initial architecture documentation
- Kubernetes manifests (Cilium, Falco, Talon)
- MCP Policy Proxy implementation (Phase 1)
- E2E tests for Kubernetes integration

---

## [1.1.0] - 2026-04-14

### Added

- Initial project structure
- Security governance framework

---

## [1.0.0] - 2026-04-13

### Added

- Project initialization