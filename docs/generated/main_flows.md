# Execution Flows

## Request Processing Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    REQUEST LIFECYCLE                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. REQUEST RECEIVED                                        │
│     │                                                       │
│     ▼                                                       │
│  2. SECURITY HEADERS                                        │
│     - X-Content-Type-Options: nosniff                       │
│     - X-Frame-Options: DENY                                 │
│     - Content-Security-Policy                                │
│     │                                                       │
│     ▼                                                       │
│  3. RATE LIMIT CHECK                                       │
│     - Per-client token bucket                              │
│     - Max 60 requests/minute                                │
│     │                                                       │
│     ▼                                                       │
│  4. AUTHENTICATION                                         │
│     - JWT Bearer token validation                           │
│     - HS256/384/512 only                                  │
│     │                                                       │
│     ▼                                                       │
│  5. SEMANTIC CHECK (Lakera)                               │
│     - Prompt injection detection                         │
│     - Tool call validation                                │
│     │                                                       │
│     ▼                                                       │
│  6. MCP BACKEND PROXY                                      │
│     - Forward to MCP server                               │
│     │                                                       │
│     ▼                                                       │
│  7. RESPONSE                                              │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Middleware Chain

The proxy uses a middleware chain pattern:

```
Request
    │
    ▼
┌─────────────────────────────────┐
│ CORS Middleware                   │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Security Headers Middleware     │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Logging Middleware             │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Rate Limit Middleware           │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Auth Middleware                 │
└─────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────┐
│ Semantic Check Middleware       │
└─────────────────────────────────┘
    │
    ▼
Response
```

---

*Generated from code analysis*
