# Fase 3 Hardening - Specification Index

## Specs Overview

| Domain | File | Priority | Requirements | Scenarios |
|--------|------|----------|-------------|-----------|
| Security | `security/spec.md` | CRITICAL | 4 | 15 |
| Testing | `testing/spec.md` | HIGH | 5 | 20 |
| DevOps | `devops/spec.md` | HIGH | 4 | 8 |
| Observability | `observability/spec.md` | MEDIUM | 5 | 7 |
| Documentation | `documentation/spec.md` | MEDIUM | 4 | 11 |

## Files Modified/Created

### CRITICAL (Must Complete)
- `src/mcp-policy-proxy/proxy.go` - Fail-closed, auth, body limits
- `src/mcp-policy-proxy/lakera.go` - Fail-closed behavior
- `src/mcp-policy-proxy/config.go` - New config options

### HIGH (Should Complete)
- `src/mcp-policy-proxy/*_test.go` - Unit tests (5 files)
- `src/mcp-policy-proxy/Dockerfile` - Non-root, multi-stage
- `Makefile` - Build automation

### MEDIUM (Nice to Have)
- `docs/runbook.md` - Incident procedures

## Verification Checklist

- [ ] Lakera error → HTTP 503 (fail-closed)
- [ ] Missing auth header → HTTP 401
- [ ] Body >1MB → HTTP 413
- [ ] Unit tests pass for all modules
- [ ] Dockerfile runs as non-root
- [ ] Make targets work: build, test, docker-build, docker-run
- [ ] JSON logs include correlation_id
- [ ] Runbook documents incident procedures
