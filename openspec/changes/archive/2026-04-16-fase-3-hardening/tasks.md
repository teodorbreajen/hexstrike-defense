# Tasks: Phase 3 Hardening

## Phase 1: CRITICAL Security Fixes (MUST do first) ✅ COMPLETE

Based on: `specs/security/spec.md`

### Task 1.1: Add LAKERA_FAIL_MODE config to ProxyConfig ✅

- [x] COMPLETE - Added FailMode, MaxBodySize, JWTSecret to ProxyConfig in proxy.go

### Task 1.2: Modify lakeraClient.CheckToolCall() error handling to block when fail mode is closed ✅

- [x] COMPLETE - Implemented fail-closed logic with HTTP 503 response

### Task 1.3: Add JWT validation to authMiddleware ✅

- [x] COMPLETE - JWT Bearer token validation with 401 responses

### Task 1.4: Add body size limit (1MB max) in semanticCheckMiddleware ✅

- [x] COMPLETE - Content-Length check returns 413 if exceeded

---

## Phase 2: Unit Tests ✅ COMPLETE

Based on: `specs/testing/spec.md`

### Task 2.1: Create unit tests for rate_limiter.go ✅

- [x] COMPLETE - Created rate_limiter_test.go with 6 tests

### Task 2.2: Create unit tests for metrics.go ✅

- [x] COMPLETE - Created metrics_test.go with 8 tests

### Task 2.3: Create unit tests for proxy handlers (use httptest) ✅

- [x] COMPLETE - Created proxy_handler_test.go with 10 tests

---

## Phase 3: DevOps ✅ COMPLETE

Based on: `specs/devops/spec.md`

### Task 3.1: Create src/mcp-policy-proxy/Dockerfile ✅

- [x] COMPLETE - Multi-stage Dockerfile with non-root user

### Task 3.2: Create Makefile in project root with build, test, docker-build targets ✅

- [x] COMPLETE - Created Makefile with all required targets

---

## Phase 4: Observability ✅ COMPLETE

Based on: `specs/observability/spec.md`

### Task 4.1: Add JSON structured logging with correlation IDs ✅

- [x] COMPLETE - Created logger.go with JSON structured logging

### Task 4.2: Add correlation ID to log entries ✅

- [x] COMPLETE - Added correlation ID middleware and header propagation

---

## Phase 5: Documentation ✅ COMPLETE

Based on: `specs/documentation/spec.md`

### Task 5.1: Create RUNBOOK.md with incident procedures ✅

- [x] COMPLETE - Created docs/runbook.md with all incident procedures