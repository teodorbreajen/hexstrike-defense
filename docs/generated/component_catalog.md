# Component Catalog

## Source Components

| Component | Type | Language | Lines | Purpose |
|-----------|------|----------|-------|----------|
| proxy.go | source | Go | 1420 | Main logic |
| main.go | source | Go | 372 | Main logic |
| jsonrpc.go | source | Go | 271 | Main logic |
| dlq.go | source | Go | 259 | Main logic |
| retry_client.go | source | Go | 225 | Main logic |
| logger.go | source | Go | 212 | Main logic |
| lakera.go | source | Go | 193 | Main logic |
| prometheus.go | source | Go | 177 | Main logic |
| config.go | source | Go | 136 | Main logic |
| cleanup.go | source | Go | 95 | Main logic |
| cors.go | source | Go | 71 | Main logic |

## Configuration Files

| File | Size | Purpose |
|------|------|---------|
| Makefile | 4.4KB | Configuration |

## Automation Scripts

| Script | Purpose |
|---------|---------|
| test-attacks.sh | Automation |
| validate.sh | Automation |
| deploy.sh | Automation |
| check-updates.sh | Automation |

## Exported Functions (API)

| Module | Function | Signature | Exported | Handler |
|--------|----------|-----------|----------|---------|
| security_comprehensive_test | TestIsInternalURLComprehensive | (t) | ✓ |  |
| security_comprehensive_test | TestIsInternalURLDecimalIP | (t) | ✓ |  |
| security_comprehensive_test | TestIsInternalURLEmptyAndInvalid | (t) | ✓ |  |
| security_comprehensive_test | TestIsTrustedProxyCIDR | (t) | ✓ |  |
| security_comprehensive_test | TestValidateBackendURLSSRF | (t) | ✓ |  |
| security_comprehensive_test | TestSanitizeToolInputComprehensive | (t) | ✓ |  |
| security_comprehensive_test | TestGetToolInfoNilProtection | (t) | ✓ |  |
| security_comprehensive_test | TestParseBatchRequestSizeLimit | (t) | ✓ |  |
| security_comprehensive_test | TestExtractToolsListParamsValidation | (t) | ✓ |  |
| security_comprehensive_test | TestValidateAndExtractRequestWhitelist | (t) | ✓ |  |
| race_test | TestClientRateLimiterRace | (t) | ✓ |  |
| race_test | TestRateLimiterConcurrentClients | (t) | ✓ |  |
| race_test | TestMetricsRace | (t) | ✓ |  |
| race_test | TestCircuitBreakerRace | (t) | ✓ |  |
| race_test | TestProxyHandlerRace | (t) | ✓ | ✓ |
| race_test | TestGetClientIPRace | (t) | ✓ |  |
| race_test | TestIsTrustedProxyRace | (t) | ✓ |  |
| race_test | TestIsInternalURLRace | (t) | ✓ |  |
| race_test | TestConcurrentTokenBucketRefill | (t) | ✓ |  |
| race_test | TestConcurrentMetricsGetSet | (t) | ✓ |  |
| prometheus_test | TestPrometheusFormat | (t) | ✓ |  |
| prometheus_test | TestPrometheusRequestsTotal | (t) | ✓ |  |
| prometheus_test | TestPrometheusStatusCodes | (t) | ✓ |  |
| prometheus_test | TestPrometheusHistogram | (t) | ✓ |  |
| prometheus_test | TestPrometheusCircuitBreaker | (t) | ✓ |  |
| prometheus_test | TestPrometheusActiveRequests | (t) | ✓ |  |
| prometheus_test | TestPrometheusGauges | (t) | ✓ |  |
| prometheus_test | TestPrometheusLabels | (t) | ✓ |  |
| prometheus_test | TestPrometheusContentType | (t) | ✓ |  |
| prometheus_test | TestPrometheusConcurrentAccess | (t) | ✓ |  |
| prometheus_test | TestPrometheusMultipleEndpointMetrics | (t) | ✓ |  |
| metrics_test | TestMetrics_RecordRequestIncrementsCountersCorrectly | (t) | ✓ |  |
| metrics_test | TestMetrics_GetMetricsReturnsCorrectValues | (t) | ✓ |  |
| metrics_test | TestMetrics_StatusCodesAreTracked | (t) | ✓ |  |
| metrics_test | TestMetrics_ConcurrentAccess | (t) | ✓ |  |
| metrics_test | TestMetrics_LatencyTracking | (t) | ✓ |  |
| metrics_test | TestMetrics_EmptyMetrics | (t) | ✓ |  |
| metrics_test | TestMetrics_NewMetrics | (t) | ✓ |  |
| fuzz_test | FuzzSanitizeToolInput | (f) | ✓ |  |
| fuzz_test | FuzzIsInternalURL | (f) | ✓ |  |
| fuzz_test | FuzzParseJSONRPC | (f) | ✓ |  |
| fuzz_test | FuzzValidateBackendURL | (f) | ✓ |  |
| fuzz_test | FuzzTokenBucket | (f) | ✓ |  |
| fuzz_test | FuzzCircuitBreaker | (f) | ✓ |  |
| fuzz_test | FuzzMetrics | (f) | ✓ |  |
| jsonrpc | ParseJSONRPC | (data) | ✓ |  |
| jsonrpc | SerializeResponse | (resp) | ✓ |  |
| jsonrpc | SerializeBatchResponse | (resps) | ✓ |  |
| jsonrpc | GetToolInfo | (parsed) | ✓ |  |
| logger_test | TestLogger_CreatesStructuredJSON | (t) | ✓ |  |
| logger_test | TestLogger_LevelFiltering | (t) | ✓ |  |
| logger_test | TestLogger_WithError | (t) | ✓ |  |
| logger_test | TestLogger_WithExtra | (t) | ✓ |  |
| logger_test | TestLogger_WithLatency | (t) | ✓ |  |
| logger_test | TestGenerateCorrelationID | (t) | ✓ |  |
| logger_test | TestLogger_ComponentDefault | (t) | ✓ |  |
| logger_test | TestLogger_AllLevels | (t) | ✓ |  |
| logger_test | TestLogEntry_AllFields | (t) | ✓ |  |
| logger_test | TestGetCorrelationID | (t) | ✓ |  |
| proxy_handler_test | TestProxyHandler_HealthEndpointReturns200 | (t) | ✓ | ✓ |
| proxy_handler_test | TestProxyHandler_AuthMiddlewareBlocksWithoutTokenOnMCPEndpoint | (t) | ✓ | ✓ |
| proxy_handler_test | TestProxyHandler_AuthMiddlewareAllowsValidJWT | (t) | ✓ | ✓ |
| proxy_handler_test | TestProxyHandler_BodySizeLimitReturns413ForOversized | (t) | ✓ | ✓ |
| proxy_handler_test | TestProxyHandler_RateLimitingReturns429WhenExhausted | (t) | ✓ | ✓ |
| proxy_handler_test | TestProxyHandler_HealthEndpointReturnsJSON | (t) | ✓ | ✓ |
| proxy_handler_test | TestProxyHandler_ReadyEndpoint | (t) | ✓ | ✓ |
| proxy_handler_test | TestProxyHandler_UnprotectedEndpointsDontRequireAuth | (t) | ✓ | ✓ |
| proxy_handler_test | TestProxyHandler_CorrectErrorResponseFormat | (t) | ✓ | ✓ |
| rate_limiter_test | TestRateLimiter_TokenRefillAfterTime | (t) | ✓ |  |
| rate_limiter_test | TestRateLimiter_AllowReturnsFalseWhenExhausted | (t) | ✓ |  |
| rate_limiter_test | TestRateLimiter_AllowReturnsTrueWhenTokensAvailable | (t) | ✓ |  |
| rate_limiter_test | TestRateLimiter_NewRateLimiter | (t) | ✓ |  |
| rate_limiter_test | TestRateLimiter_DecrementBehavior | (t) | ✓ |  |
| config | LoadConfigFile | (path) | ✓ |  |
| cors_test | TestCORSMiddleware_NoOriginHeader | (t) | ✓ |  |
| cors_test | TestCORSMiddleware_AllowedOrigin | (t) | ✓ |  |
| cors_test | TestCORSMiddleware_DeniedOrigin | (t) | ✓ |  |
| cors_test | TestCORSMiddleware_PreflightOPTIONS | (t) | ✓ |  |
| cors_test | TestCORSMiddleware_PreflightWithHeaders | (t) | ✓ |  |
| cors_test | TestCORSMiddleware_NoOriginsConfigured | (t) | ✓ |  |
| cors_test | TestCORSMiddleware_CredentialsDisabled | (t) | ✓ |  |
| cors_test | TestCORSMiddleware_AllHTTPMethodsAllowed | (t) | ✓ |  |
| cors_test | TestNewCORSMiddleware_CaseInsensitiveStorage | (t) | ✓ |  |
| cleanup | CleanupNow | (dlq) | ✓ |  |
| cleanup | CleanupWithTTL | (dlq, ttlHours) | ✓ |  |
| dlq_test | TestDLQ_Enqueue_SavesFile | (t) | ✓ |  |
| dlq_test | TestDLQ_Replay_ExecutesHandler | (t) | ✓ | ✓ |
| dlq_test | TestDLQ_Replay_FIFOOrder | (t) | ✓ |  |
| dlq_test | TestDLQ_Size_CountsFiles | (t) | ✓ |  |
| dlq_test | TestDLQ_Cleanup_EliminatesExpiredMessages | (t) | ✓ |  |
| dlq_test | TestDLQ_CleanupWithTTL | (t) | ✓ |  |
| dlq_test | TestDLQ_Enqueue_GeneratesID | (t) | ✓ |  |
| dlq_test | TestDLQ_Enqueue_SetsTimestamp | (t) | ✓ |  |
| dlq_test | TestDLQ_Remove_DeletesMessage | (t) | ✓ |  |
| dlq_test | TestDLQ_Remove_NotFound | (t) | ✓ |  |
| dlq_test | TestDLQ_Peek_ReturnsOldestMessage | (t) | ✓ |  |
| dlq_test | TestDLQ_Peek_EmptyQueue | (t) | ✓ |  |
| dlq_test | TestDLQ_GetMessages_ReturnsAllMessages | (t) | ✓ |  |
| dlq_test | TestDLQ_Enqueue_NilRequest | (t) | ✓ |  |
| dlq_test | TestDLQ_Replay_NilHandler | (t) | ✓ | ✓ |

## Types and Data Structures

| Module | Type | Kind | Exported | Fields |
|--------|------|------|----------|-------|
| jsonrpc | JSONRPCError | struct | ✓ | 3 |
| jsonrpc | JSONRPCRequest | struct | ✓ | 4 |
| jsonrpc | JSONRPCResponse | struct | ✓ | 4 |
| jsonrpc | ToolCallParams | struct | ✓ | 2 |
| jsonrpc | ToolsListParams | struct | ✓ | 2 |
| jsonrpc | ParsedRequest | struct | ✓ | 7 |
| cors | CORSMiddleware | struct | ✓ | 2 |
| config | ConfigFile | struct | ✓ | 16 |
| cleanup | CleanupConfig | struct | ✓ | 2 |