# Execution Flows

## Request Processing Lifecycle

```
┌─────────────────────────────────────────────────────────────┐
│                   REQUEST LIFECYCLE                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. REQUEST RECEIVED (HTTP/WS)                            │
│     │                                                       │
│     ▼                                                       │
│  2. SECURITY HEADERS                                       │
│     - X-Content-Type-Options: nosniff                       │
│     - X-Frame-Options: DENY                                │
│     - Content-Security-Policy                               │
│     │                                                       │
│     ▼                                                       │
│  3. RATE LIMIT CHECK                                      │
│     - Token bucket algorithm                               │
│     - Per-client limits                                  │
│     │                                                       │
│     ▼                                                       │
│  4. AUTHENTICATION                                       │
│     - JWT Bearer token (HS256/384/512)                   │
│     - Token expiry validation                            │
│     │                                                       │
│     ▼                                                       │
│  5. SEMANTIC SECURITY CHECK                             │
│     - Prompt injection detection                        │
│     - Tool call validation                           │
│     - Content classification                        │
│     │                                                       │
│     ▼                                                       │
│  6. MCP BACKEND PROXY                                    │
│     - Request transformation                         │
│     - Response handling                              │
│     │                                                       │
│     ▼                                                       │
│  7. RESPONSE                                          │
│     - JSON serialization                            │
│     - Security headers                             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Middleware Chain

The proxy implements a middleware chain pattern:

```
Request
    │
    ▼
┌─────────────────────────────────┐
│ CORS Middleware                   │
│ - Origin validation             │
│ - Method checking             │
└───────────────────────���─���───────┘
    │
    ▼
┌─────────────────────────────────┐
│ Security Headers Middleware     │
│ - CSP headers                 │
│ - HSTS                        │
│ - X-Frame-Options             │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Logging Middleware             │
│ - Request ID                  │
│ - Access logs                │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Rate Limit Middleware           │
│ - Token bucket                │
│ - Per-client tracking        │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Auth Middleware                │
│ - JWT validation             │
│ - Claims extraction         │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Semantic Check Middleware       │
│ - Lakera integration         │
│ - Content filtering         │
└─────────────────────────────────┘
    │
    ▼
Response
```

## Error Handling Flow

```
Error Occurs
    │
    ▼
┌─────────────────────────────────┐
│ Error Classification           │
│ - Validation Error          │
│ - Authentication Error      │
│ - Rate Limit Error          │
│ - Semantic Error           │
│ - Backend Error            │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Error Response                │
│ - Appropriate HTTP code     │
│ - Safe error message      │
│ - Request ID for debug   │
└─────────────────────────────────┘
    │
    ▼
Logging + Alerting
```

---

*Generated from code analysis*
