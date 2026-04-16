# Proposal: fase-3-hardening

## Intent

Implement critical security hardening fixes identified in the security audit to make the MCP Policy Proxy production-ready. Current proxy has fail-open behavior on Lakera failure, no authentication, and no input validation—making it unsuitable for production deployment.

## Scope

### In Scope (CRITICAL - MUST fix before production)

| Priority | Item | File | Current Issue |
|----------|------|------|---------------|
| CRITICAL | Fail-Closed Lakera | `proxy.go:284-287` | Allows all requests on Lakera error |
| CRITICAL | Real Authentication | `proxy.go:207-218` | authMiddleware commented out, allows ALL |
| CRITICAL | Input Validation | `proxy.go:238` | No max body size limit |

### In Scope (HIGH PRIORITY)

| Priority | Item | File | Current Issue |
|----------|------|------|---------------|
| HIGH | Unit Tests | `src/mcp-policy-proxy/` | No Go unit tests exist |
| HIGH | Docker Build | `src/mcp-policy-proxy/Dockerfile` | Exists but may need hardening |
| HIGH | Makefile | Project root | No standard build automation |

### In Scope (MEDIUM PRIORITY)

| Priority | Item | File | Current Issue |
|----------|------|------|---------------|
| MEDIUM | go.mod/go.sum | `src/mcp-policy-proxy/` | Missing go.sum |
| MEDIUM | SIEM Integration | `proxy.go` | No structured JSON logging |
| MEDIUM | Basic Runbook | `docs/` | Missing incident procedures |

### Out of Scope
- Modifying hexstrike-ai source code
- Kubernetes cluster changes
- Falco rule tuning
- Production deployment automation

## Approach

1. **Fix CRITICAL security issues first** (Priority 1)
   - Implement fail-closed behavior: block requests when Lakera is unavailable
   - Add JWT/API key authentication validation
   - Add max body size limit (1MB default)

2. **Add unit tests** (Priority 2)
   - Use `httptest` for HTTP handler tests
   - Test proxy.go, rate_limiter.go, metrics.go, lakera_client.go

3. **Add DevOps automation** (Priority 3)
   - Verify/toughen Dockerfile for non-root
   - Create Makefile with build, test, docker-build, docker-run targets

4. **Add observability and docs** (Priority 4)
   - JSON structured logging with correlation IDs
   - Document incident response runbook

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `src/mcp-policy-proxy/proxy.go` | Modified | Auth, fail-closed, body limits |
| `src/mcp-policy-proxy/Dockerfile` | Modified | Non-root user, security hardening |
| `src/mcp-policy-proxy/Makefile` | New | Build automation |
| `src/mcp-policy-proxy/*_test.go` | New | Unit tests |
| `docs/runbook.md` | New | Incident procedures |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|-------------|
| Blocking valid traffic on Lakera outage | High | Configurable fail-closed vs fail-open via env var |
| Breaking existing E2E tests | Medium | Add tests first, verify before deploy |
| Auth blocking internal tools | Medium | Whitelist internal paths in config |

## Rollback Plan

- Keep backup of original `proxy.go` before changes
- Use feature flags in config for auth/validation toggles
- If Docker build fails: use existing `mcp-policy-proxy.exe`
- If tests fail: CI must fail before merge

## Dependencies

- Go 1.21+ for unit tests
- golangci-lint for linting (optional)
- Docker for containerization

## Success Criteria

- [ ] Lakera failures result in blocked requests (fail-closed)
- [ ] Requests without valid auth return 401/403
- [ ] Requests >1MB body rejected
- [ ] Unit tests pass for proxy, rate limiter, metrics, lakera client
- [ ] Dockerfile builds successfully with non-root user
- [ ] Makefile targets work: build, test, docker-build, docker-run
- [ ] JSON structured logs with correlation IDs
- [ ] Runbook documents Lakera-down and Falco-alert procedures