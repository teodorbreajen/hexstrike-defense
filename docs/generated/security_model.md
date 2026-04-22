# Security Model

## Authentication & Access Control

- **JWT Authentication**: Bearer token validation required
- **Algorithm Restriction**: HS256/384/512 only
- **Algorithm Confusion Protection**: Blocks alg:none attacks

## Input Validation

- **Fail-Closed**: Block when Lakera unavailable
- **Body Size Limit**: 1MB max (configurable)
- **Input Sanitization**: SSRF, SQL injection, command injection detection
- **Path Traversal Protection**: Blocks ../ variants

## Rate Limiting & DoS Protection

- **Per-Client Rate Limiting**: Token bucket per client IP
- **Concurrent Request Limiting**: Max 100 concurrent
- **Batch Request Limits**: Max 10 requests per batch

## Resilience

- **Circuit Breaker**: Prevents cascade failures
- **Retry with Exponential Backoff**: 1s, 2s, 4s strategy
- **Connection Pooling**: Reusable HTTP connections
- **Dead Letter Queue**: Failed requests stored for replay

## Security Headers

| Header | Value |
|--------|-------|
| X-Content-Type-Options | nosniff |
| X-Frame-Options | DENY |
| Strict-Transport-Security | max-age=31536000 |
| Content-Security-Policy | default-src 'none' |

## Protected Endpoints

| Endpoint | Auth Required |
|----------|--------------|
| `/health` | No |
| `/metrics` | No |
| `/ready` | No |
| `/mcp/*` | Yes (Bearer JWT) |

---

*Generated from security code*
