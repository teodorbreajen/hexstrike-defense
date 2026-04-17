# Tasks: Security & Resilience Hardening

## Phase 1: Config & Security Foundation

- [x] 1.1 Add `CORS_ALLOWED_ORIGINS` and `CORS_ALLOW_CREDENTIALS` env vars to `Config` struct in `main.go`
- [x] 1.2 Add `CORSAllowedOrigins []string` field to `ProxyConfig` in `proxy.go`
- [x] 1.3 Verify JWT fail-hard already implemented in `loadConfig()` (lines 98-103) — + GIN_MODE=release support
- [x] 1.4 Pass CORS config from main.go to ProxyConfig when creating proxy

## Phase 2: CORS Middleware

- [x] 2.1 Create `src/mcp-policy-proxy/cors.go` with `CORSMiddleware` struct and `map[string]bool` for allowed origins
- [x] 2.2 Implement `Handle()` method with Origin header allowlist matching
- [x] 2.3 Add CORS response headers: `Access-Control-Allow-Origin`, `Access-Control-Allow-Credentials`, `Access-Control-Allow-Methods`
- [x] 2.4 Handle preflight OPTIONS requests with 204 No Content
- [x] 2.5 Integrate CORS middleware into proxy middleware chain in `proxy.go`

## Phase 3: Retry Client

- [x] 3.1 Create `src/mcp-policy-proxy/retry_client.go` with `RetryClient` struct wrapping `http.Client`
- [x] 3.2 Implement `DoWithRetry()` with exponential backoff: 1s, 2s, 4s delays
- [x] 3.3 Define retryable errors: network timeout, HTTP 5xx, connection reset
- [x] 3.4 Define non-retryable: HTTP 4xx (except 429), success responses
- [x] 3.5 Add structured logging for retry attempts with fields: `retry_attempt`, `max_retries`, `delay_ms`, `will_retry`
- [x] 3.6 Integrate `RetryClient` into `Proxy.backendClient` in `proxy.go`

## Phase 4: Dead Letter Queue

- [x] 4.1 Create `src/mcp-policy-proxy/dlq/` directory
- [x] 4.2 Create `dlq.go` with `DLQ` struct, `Enqueue()`, `Dequeue()`, `Replay()` methods
- [x] 4.3 Implement file-based storage: `data/dlq/{uuid}.json` format
- [x] 4.4 Add `DLQMessage` struct with fields: `id`, `timestamp`, `payload`, `error`, `retry_count`, `source`
- [x] 4.5 Create `cleanup.go` with background goroutine for 24h TTL cleanup
- [x] 4.6 Add `DLQ_PATH` and `DLQ_TTL_HOURS` config options to main.go
- [x] 4.7 Wire DLQ into retry client to enqueue on final failure

## Phase 5: Testing

- [x] 5.1 Create `cors_test.go` with table-driven tests for origin matching (allowed, denied, wildcard)
- [x] 5.2 Create `retry_client_test.go` with mocked HTTP server for retry scenarios
- [x] 5.3 Test exponential backoff timing and max retries exhaustion
- [x] 5.4 Test non-retryable 4xx errors fail immediately
- [x] 5.5 Create `dlq_test.go` with temp directory for file-based storage
- [x] 5.6 Test Enqueue/Dequeue operations and TTL expiration logic
- [x] 5.7 Test DLQ Replay in FIFO order

## Phase 6: Verification

- [ ] 6.1 Run `go test ./...` to verify all tests pass
- [ ] 6.2 Run `go vet ./...` for static analysis
- [ ] 6.3 Verify coverage improvement with `go test -cover`
- [ ] 6.4 Manual test CORS headers with curl or browser
- [ ] 6.5 Verify JWT fail-hard behavior with empty JWT_SECRET and DEV_MODE=false
