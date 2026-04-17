# Design: Security & Resilience Hardening

## Technical Approach

Implement security and resilience hardening across 4 deliverables: JWT mandatory validation, CORS allowlist, retry with exponential backoff, and dead letter queue. Follows existing middleware chain pattern in `proxy.go` and environment-based configuration from `main.go`.

## Architecture Decisions

### Decision 1: JWT Startup Validation (Fail-Hard)

**Choice**: Add `log.Fatalf` in `loadConfig()` when `JWT_SECRET` is empty and `DEV_MODE=false`.
**Alternatives considered**: Runtime-only validation (rejected — insecure window), config file enforcement (rejected — env vars are source of truth per existing pattern).
**Rationale**: Maintains existing pattern where `loadConfig()` already logs warnings but doesn't fail. Fail-hard only in production mode aligns with spec requirement.

### Decision 2: CORS Middleware Implementation

**Choice**: Custom middleware in `cors.go` using Origin header allowlist matching.
**Alternatives considered**: Gin CORS library (rejected — adds dependency), gorilla/handlers (rejected — overkill for simple allowlist).
**Rationale**: No existing CORS dependency. Custom implementation matches codebase's minimal-dependency philosophy and existing `AllowedOrigins` field in `ProxyConfig`.

### Decision 3: Retry Client Pattern

**Choice**: New `retry_client.go` wraps `http.Client` with `DoWithRetry()` method.
**Alternatives considered**: Aspect-oriented wrapping (rejected — complexity), HTTP transport middleware (rejected — too invasive).
**Rationale**: Clean separation, testable in isolation, uses existing `backendClient` field pattern from `proxy.go`.

### Decision 4: DLQ Storage

**Choice**: File-based JSON in `data/dlq/` directory, one file per message.
**Alternatives considered**: Redis/queue broker (rejected — external dependency), in-memory (rejected — doesn't survive restarts).
**Rationale**: Matches spec's disk persistence requirement, survives restarts, simple to inspect/debug. TTL cleanup via background goroutine.

## Data Flow

```
Request → Middleware Chain → RetryClient.DoWithRetry() ─┬─→ MCP Backend
                            │                            │
                            │ (on 5xx/network error)    │
                            ↓                            │
                       DLQ.Enqueue() ──→ data/dlq/{uuid}.json
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `src/mcp-policy-proxy/main.go` | Modify | Add CORS env vars (`CORS_ALLOWED_ORIGINS`, `CORS_ALLOW_CREDENTIALS`), update JWT validation to fail-hard in production |
| `src/mcp-policy-proxy/proxy.go` | Modify | Add `CORSAllowedOrigins` to `ProxyConfig`, integrate CORS middleware |
| `src/mcp-policy-proxy/cors.go` | Create | CORS middleware with Origin allowlist matching |
| `src/mcp-policy-proxy/retry_client.go` | Create | HTTP client wrapper with exponential backoff (1s, 2s, 4s) |
| `src/mcp-policy-proxy/dlq/dlq.go` | Create | DLQ struct with Enqueue, Dequeue, Replay methods |
| `src/mcp-policy-proxy/dlq/cleanup.go` | Create | Background cleanup goroutine for 24h TTL |

## Interfaces / Contracts

```go
// src/mcp-policy-proxy/cors.go
type CORSMiddleware struct {
    allowedOrigins map[string]bool
    allowCreds     bool
}

func (c *CORSMiddleware) Handle(next http.Handler) http.Handler

// src/mcp-policy-proxy/retry_client.go
type RetryClient struct {
    client     *http.Client
    maxRetries int
    baseDelay  time.Duration
}

func (r *RetryClient) DoWithRetry(ctx context.Context, req *http.Request) (*http.Response, error)

// src/mcp-policy-proxy/dlq/dlq.go
type DLQ struct {
    path      string
    ttlHours  int
}

type DLQMessage struct {
    ID        string                 `json:"id"`
    Timestamp time.Time              `json:"timestamp"`
    Payload   json.RawMessage        `json:"payload"`
    Error     string                 `json:"error"`
    RetryCount int                   `json:"retry_count"`
    Source    string                 `json:"source"`
}

func (d *DLQ) Enqueue(msg *DLQMessage) error
func (d *DLQ) Replay() (int, error)  // Returns count of processed messages
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | CORS origin matching, retry backoff timing, DLQ TTL | Table-driven tests in `*_test.go` |
| Integration | Retry actually retries, DLQ persists to disk | `testdata/dlq/` for temp files |
| E2E | Full request with CORS headers | Existing `*_test.go` files |

## Migration / Rollout

No migration required. New configuration fields have safe defaults:
- `CORS_ALLOWED_ORIGINS=""` → CORS disabled (current behavior)
- `DLQ_PATH=data/dlq/` → Auto-created if missing
- JWT fail-hard is additive safety (existing code already warns)

## Open Questions

None — all requirements are fully specified in `specs/security/spec.md` and `specs/resilience/spec.md`.
