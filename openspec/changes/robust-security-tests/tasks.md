# Tasks: Robust Security Tests for mcp-policy-proxy

## Phase 1: Test Infrastructure (Mock Helpers & Keys)

- [ ] 1.1 Add `mockLakeraClient` struct in `security_test.go` with `shouldError`, `shouldBlock`, `blockReason`, `blockScore` fields
- [ ] 1.2 Implement `CheckToolCall()` method returning configurable response
- [ ] 1.3 Implement `HealthCheck()` method that returns error if `shouldError=true`
- [ ] 1.4 Add control methods: `SetError()`, `SetBlock()`, `SetAllow()`
- [ ] 1.5 Add `TestMain()` in `security_test.go` that generates RSA/EC keys once
- [ ] 1.6 Add package variables: `testRSAPrivateKey`, `testRSAPublicKey`, `testECDSAPrivateKey`, `testECDSAPublicKey`

## Phase 2: JWT Algorithm Validation Tests

- [ ] 2.1 Add test case "RS256 with real RSA key rejected" in `TestSecurity_JWTAlgorithmValidation`
- [ ] 2.2 Add test case "ES256 with real EC key rejected" in `TestSecurity_JWTAlgorithmValidation`
- [ ] 2.3 Add test case "alg:none token rejected" in `TestSecurity_JWTAlgorithmValidation`
- [ ] 2.4 Update test to use `testRSAPrivateKey` and `testECDSAPrivateKey` from TestMain

## Phase 3: Integration Tests - FailClosed & SSRF

- [ ] 3.1 Add `createTestRouterWithLakera()` in `proxy_handler_test.go` accepting `mockLakeraClient`
- [ ] 3.2 Create `TestSecurity_FailClosedMode` with mock that returns error, verify 503 response
- [ ] 3.3 Create `TestSecurity_SSRFProtection` with proxy using localhost backend, verify 500 response
- [ ] 3.4 Test SSRF with `127.0.0.1` and `10.0.0.1` backends
- [ ] 3.5 Add `TestSecurity_LakeraNilFallback` verifying no crash with `nil` lakeraClient

## Phase 4: Batch Size & Input Sanitization Tests

- [ ] 4.1 Add `TestSecurity_BatchSizeLimit` with actual POST `/mcp/v1/call` sending 11 requests, verify 400
- [ ] 4.2 Add SQL injection cases: `' OR 1=1 --`, `'; DROP TABLE --`, `UNION SELECT`
- [ ] 4.3 Add command injection cases: `$(whoami)`, `` `id` ``, `${HOME}`, `; cat /etc/passwd`
- [ ] 4.4 Add XSS case: `<script>alert(1)</script>`
- [ ] 4.5 Add Unicode/emoji case: `🏴󠁧󠁢󠁥󠁮󠁧󠁿`, `🇦🇷bash`
- [ ] 4.6 Add case variations: `Bash`, `BASH`, `BaSh`

## Phase 5: Rate Limiter Cleanup Tests

- [ ] 5.1 Add `TestSecurity_RateLimiterPerClient` cleanup test with configurable TTL
- [ ] 5.2 Add `TestSecurity_RateLimiterMaxClients` verifying limit enforcement
- [ ] 5.3 Mark long cleanup test with `t.Skip()` for short mode if >60s

## Files to Modify

| File | Changes |
|------|---------|
| `src/mcp-policy-proxy/security_test.go` | mockLakeraClient, TestMain, expanded JWT/sanitization tests |
| `src/mcp-policy-proxy/proxy_handler_test.go` | createTestRouterWithLakera, integration tests |
