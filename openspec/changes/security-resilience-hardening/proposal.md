# Proposal: Security & Resilience Hardening

## Intent

Elevar el security score de 78.7/100 a 90+ abordando vulnerabilidades críticas y de alta prioridad identificadas en la auditoría de seguridad.

## Scope

### In Scope
- **S-1**: JWT_SECRET validación en startup (required, no runtime-only)
- **S-2**: CORS configurable vía entorno (whitelist-based)
- **R-1**: Retry logic con exponential backoff (HTTP client, DB ops)
- **R-2**: Dead letter queue para requests fallidos (async queue)

### Out of Scope
- IP masking improvement (futuro)
- Error response sanitization (partial, existente)
- Metrics persistencia (backend stability priorizada)
- Fuzz tests en CI (coverage primero)
- Integration tests con backend real (test infrastructure)

## Approach

1. **Config validation layer**: Agregar startup validation para JWT_SECRET, CORS origins
2. **Retry middleware**: Wrapper HTTP/DB con exponential backoff (max 3 retries, 1s-2s-4s)
3. **DLQ implementation**:cola separada para messages que fallan después de max retries

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/server/main.go` | Modified | Startup config validation |
| `internal/config/` | Modified | JWT_SECRET, CORS config |
| `internal/client/http.go` | New | Retry wrapper |
| `internal/queue/` | Modified | DLQ implementation |
| `.env.example` | Modified | New vars: JWT_SECRET, CORS_ORIGINS |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|-------------|
| Breaking change: JWT_SECRET required | High | Backward compat con warning日志; release note clara |
| DLQ full queue | Low | TTL en mensajes, cleanup job |

## Rollback Plan

- Revertir cambios en config validation ( permissive mode)
- Disable retry middleware via env flag
- Disable DLQ, messages se pierden (aceptable para rollback)

## Dependencies

- None external required

## Success Criteria

- [ ] Security score ≥ 90 en próxima auditoría
- [ ] JWT_SECRET fail fast en startup si no está seteado
- [ ] CORS working con CORS_ORIGINS env var
- [ ] Retry logic visible en logs (attempt N, success/fail)
- [ ] DLQ receives failed messages after max retries