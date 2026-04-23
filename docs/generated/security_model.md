# Security Model

## Authentication & Access Control

- **JWT Authentication**: Bearer token validation required
- **Algorithm Restriction**: HS256/384/512 only
- **Algorithm Confusion Protection**: Blocks alg:none attacks
- **Token Expiry**: Required, configurable max age

## Input Validation

- **Fail-Closed**: Block when validation service unavailable
- **Body Size Limit**: 1MB max (configurable)
- **Input Sanitization**: SSRF, SQL injection, command injection detection
- **Path Traversal Protection**: Blocks ../ variants

## Rate Limiting & DoS Protection

- **Per-Client Rate Limiting**: Token bucket per client IP
- **Concurrent Request Limiting**: Max 100 concurrent requests
- **Batch Request Limits**: Max 10 requests per batch

## Security Patterns Detected

| Severity | Count | Patterns |
|----------|-------|----------|
| Critical | 0 | Hardcoded secrets, unsafe deserialization |
| High | 1 | SQL/Command injection, XSS |
| Medium | 52 | Weak crypto, path traversal |

## Security Headers

| Header | Value |
|--------|-------|
| X-Content-Type-Options | nosniff |
| X-Frame-Options | DENY |
| Strict-Transport-Security | max-age=31536000 |
| Content-Security-Policy | default-src 'none' |

## Endpoint Security

| Endpoint | Auth Required | Rate Limited |
|----------|--------------|------------|
| `/health` | No | No |
| `/ready` | No | No |
| `/metrics` | No | No |
| `/mcp/*` | Yes (Bearer JWT) | Yes |

## Security Best Practices

1. Always use HTTPS in production
2. Rotate JWT secrets regularly
3. Enable fail-closed mode
4. Monitor security events
5. Keep dependencies updated

---

*Generated from security code*
