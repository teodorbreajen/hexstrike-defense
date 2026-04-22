# Component Catalog

## Core Components

| Component | Type | Language | Lines | Purpose |
|-----------|------|----------|-------|----------|
| config.go | source | Go | 136 | Proxy logic |
| cors.go | source | Go | 71 | Proxy logic |
| jsonrpc.go | source | Go | 271 | Proxy logic |
| lakera.go | source | Go | 193 | Proxy logic |
| logger.go | source | Go | 212 | Proxy logic |
| main.go | source | Go | 372 | Proxy logic |
| prometheus.go | source | Go | 177 | Proxy logic |
| proxy.go | source | Go | 1420 | Proxy logic |
| retry_client.go | source | Go | 225 | Proxy logic |
| cleanup.go | source | Go | 95 | Proxy logic |
| dlq.go | source | Go | 259 | Proxy logic |
| test_cilium_policies.go | source | Go | 329 | Proxy logic |
| test_falco_detection.go | source | Go | 280 | Proxy logic |
| test_semantic_firewall.go | source | Go | 249 | Proxy logic |
| cluster.go | source | Go | 193 | Proxy logic |
| utils.go | source | Go | 194 | Proxy logic |

## Configuration Files

| File | Purpose |
|------|---------|
| Makefile | Configuration |

## Scripts

| File | Purpose |
|------|---------|
| scripts\check-updates.sh | Automation script |
| scripts\deploy.sh | Automation script |
| scripts\test-attacks.sh | Automation script |
| scripts\validate.sh | Automation script |

## Exported Functions

| Module | Function | Handler | Purpose |
|--------|----------|---------|----------|
| config | LoadConfigFile |  | *ConfigFile, error |
| cors_test | TestCORSMiddleware_NoOriginHeader |  | N/A |
| cors_test | TestCORSMiddleware_AllowedOrigin |  | N/A |
| cors_test | TestCORSMiddleware_DeniedOrigin |  | N/A |
| cors_test | TestCORSMiddleware_PreflightOPTIONS |  | N/A |
| cors_test | TestCORSMiddleware_PreflightWithHeaders |  | N/A |
| cors_test | TestCORSMiddleware_NoOriginsConfigured |  | N/A |
| cors_test | TestCORSMiddleware_CredentialsDisabled |  | N/A |
| cors_test | TestCORSMiddleware_AllHTTPMethodsAllowed |  | N/A |
| cors_test | TestNewCORSMiddleware_CaseInsensitiveStorage |  | N/A |
| fuzz_test | FuzzSanitizeToolInput |  | N/A |
| fuzz_test | FuzzIsInternalURL |  | N/A |
| fuzz_test | FuzzParseJSONRPC |  | N/A |
| fuzz_test | FuzzValidateBackendURL |  | N/A |
| fuzz_test | FuzzTokenBucket |  | N/A |
| fuzz_test | FuzzCircuitBreaker |  | N/A |
| fuzz_test | FuzzMetrics |  | N/A |
| integration_test | Close |  | N/A |
| integration_test | TestIntegration_ProxyForwardsRequests |  | N/A |
| integration_test | TestIntegration_MetricsRecorded |  | N/A |
| integration_test | TestIntegration_RejectInternalHost |  | N/A |
| integration_test | TestIntegration_CircuitBreakerIntegration |  | N/A |
| integration_test | TestIntegration_RetryBehavior |  | N/A |
| integration_test | TestIntegration_PrometheusMetrics |  | N/A |
| integration_test | TestIntegration_ConcurrentRequests |  | N/A |
| integration_test | TestIntegration_RateLimitingEndToEnd |  | N/A |
| integration_test | TestMockMCPBackend |  | N/A |
| jsonrpc | ParseJSONRPC |  | *ParsedRequest, error |
| jsonrpc | SerializeResponse |  | []byte, error |
| jsonrpc | SerializeBatchResponse |  | []byte, error |
| jsonrpc | GetToolInfo |  | toolName string, args string, ok bool |
| lakera | CheckToolCall |  | bool, int, string, error |
| logger | SetWriter |  | N/A |
| logger | SetMinLevel |  | N/A |
| logger_test | TestLogger_CreatesStructuredJSON |  | N/A |
| logger_test | TestLogger_LevelFiltering |  | N/A |
| logger_test | TestLogger_WithError |  | N/A |
| logger_test | TestLogger_WithExtra |  | N/A |
| logger_test | TestLogger_WithLatency |  | N/A |
| logger_test | TestGenerateCorrelationID |  | N/A |
| logger_test | TestLogger_ComponentDefault |  | N/A |
| logger_test | TestLogger_AllLevels |  | N/A |
| logger_test | TestLogEntry_AllFields |  | N/A |
| logger_test | TestGetCorrelationID |  | N/A |
| main | ServeHTTP |  | N/A |
| metrics_test | TestMetrics_RecordRequestIncrementsCountersCorrectly |  | N/A |
| metrics_test | TestMetrics_GetMetricsReturnsCorrectValues |  | N/A |
| metrics_test | TestMetrics_StatusCodesAreTracked |  | N/A |
| metrics_test | TestMetrics_ConcurrentAccess |  | N/A |
| metrics_test | TestMetrics_LatencyTracking |  | N/A |
| metrics_test | TestMetrics_EmptyMetrics |  | N/A |
| metrics_test | TestMetrics_NewMetrics |  | N/A |
| prometheus | RecordRequest |  | N/A |
| prometheus | RecordLakeraBlock |  | N/A |
| prometheus | RecordBackendError |  | N/A |
| prometheus | RecordRetry |  | N/A |
| prometheus | RecordDLQMessage |  | N/A |
| prometheus | SetDLQCount |  | N/A |
| prometheus | SetCircuitBreakerState |  | N/A |
| prometheus | IncActiveRequests |  | N/A |
| prometheus | DecActiveRequests |  | N/A |
| prometheus | Gather |  | []*dto.MetricFamily, error |
| prometheus_test | TestPrometheusFormat |  | N/A |
| prometheus_test | TestPrometheusRequestsTotal |  | N/A |
| prometheus_test | TestPrometheusStatusCodes |  | N/A |
| prometheus_test | TestPrometheusHistogram |  | N/A |
| prometheus_test | TestPrometheusCircuitBreaker |  | N/A |
| prometheus_test | TestPrometheusActiveRequests |  | N/A |
| prometheus_test | TestPrometheusGauges |  | N/A |
| prometheus_test | TestPrometheusLabels |  | N/A |
| prometheus_test | TestPrometheusContentType |  | N/A |
| prometheus_test | TestPrometheusConcurrentAccess |  | N/A |
| prometheus_test | TestPrometheusMultipleEndpointMetrics |  | N/A |
| proxy | RecordSuccess |  | N/A |
| proxy | RecordFailure |  | N/A |
| proxy | RecordRequest |  | N/A |
| proxy | GetMetrics |  | total, blocked, allowed int64, avgLatency float64, statusCodes map[int]int64 |
| proxy | WriteHeader |  | N/A |
| proxy_handler_test | TestProxyHandler_HealthEndpointReturns200 | ✓ | N/A |
| proxy_handler_test | TestProxyHandler_AuthMiddlewareBlocksWithoutTokenOnMCPEndpoint | ✓ | N/A |
| proxy_handler_test | TestProxyHandler_AuthMiddlewareAllowsValidJWT | ✓ | N/A |
| proxy_handler_test | TestProxyHandler_BodySizeLimitReturns413ForOversized | ✓ | N/A |
| proxy_handler_test | TestProxyHandler_RateLimitingReturns429WhenExhausted | ✓ | N/A |
| proxy_handler_test | TestProxyHandler_HealthEndpointReturnsJSON | ✓ | N/A |
| proxy_handler_test | TestProxyHandler_ReadyEndpoint | ✓ | N/A |
| proxy_handler_test | TestProxyHandler_UnprotectedEndpointsDontRequireAuth | ✓ | N/A |
| proxy_handler_test | TestProxyHandler_CorrectErrorResponseFormat | ✓ | N/A |
| race_test | TestClientRateLimiterRace |  | N/A |
| race_test | TestRateLimiterConcurrentClients |  | N/A |
| race_test | TestMetricsRace |  | N/A |
| race_test | TestCircuitBreakerRace |  | N/A |
| race_test | TestProxyHandlerRace | ✓ | N/A |
| race_test | TestGetClientIPRace |  | N/A |
| race_test | TestIsTrustedProxyRace |  | N/A |
| race_test | TestIsInternalURLRace |  | N/A |
| race_test | TestConcurrentTokenBucketRefill |  | N/A |
| race_test | TestConcurrentMetricsGetSet |  | N/A |
| rate_limiter_test | TestRateLimiter_TokenRefillAfterTime |  | N/A |
| rate_limiter_test | TestRateLimiter_AllowReturnsFalseWhenExhausted |  | N/A |
| rate_limiter_test | TestRateLimiter_AllowReturnsTrueWhenTokensAvailable |  | N/A |
| rate_limiter_test | TestRateLimiter_NewRateLimiter |  | N/A |
| rate_limiter_test | TestRateLimiter_DecrementBehavior |  | N/A |
| retry_client | Do |  | *http.Response, error |
| retry_client | DoWithRetry |  | *http.Response, error |
| retry_client_test | Read |  | n int, err error |
| retry_client_test | TestRetryClient_SuccessAtFirstAttempt |  | N/A |
| retry_client_test | TestRetryClient_RetryOn5xxEventuallySuccess |  | N/A |
| retry_client_test | TestRetryClient_MaxRetriesExhausted |  | N/A |
| retry_client_test | TestRetryClient_4xxNotRetryable |  | N/A |
| retry_client_test | TestRetryClient_429IsRetryable |  | N/A |
| retry_client_test | TestIsRetryableStatusCode |  | N/A |
| retry_client_test | TestCalculateBackoff |  | N/A |
| retry_client_test | TestIsRetryableError |  | N/A |
| retry_client_test | TestRetryClient_ContextCancellation |  | N/A |
| retry_client_test | TestRetryClient_RequestBodyPreserved |  | N/A |
| retry_client_test | TestRetryClient_WithCorrelationID |  | N/A |
| retry_client_test | TestNewRetryClient_DefaultConfig |  | N/A |
| retry_client_test | TestDefaultRetryConfig |  | N/A |
| security_comprehensive_test | TestIsInternalURLComprehensive |  | N/A |
| security_comprehensive_test | TestIsInternalURLDecimalIP |  | N/A |
| security_comprehensive_test | TestIsInternalURLEmptyAndInvalid |  | N/A |
| security_comprehensive_test | TestIsTrustedProxyCIDR |  | N/A |
| security_comprehensive_test | TestValidateBackendURLSSRF |  | N/A |
| security_comprehensive_test | TestSanitizeToolInputComprehensive |  | N/A |
| security_comprehensive_test | TestGetToolInfoNilProtection |  | N/A |
| security_comprehensive_test | TestParseBatchRequestSizeLimit |  | N/A |
| security_comprehensive_test | TestExtractToolsListParamsValidation |  | N/A |
| security_comprehensive_test | TestValidateAndExtractRequestWhitelist |  | N/A |
| security_test | CheckToolCall |  | bool, int, string, error |
| security_test | SetError |  | N/A |
| security_test | SetBlock |  | N/A |
| security_test | TestMain |  | N/A |
| security_test | TestSecurity_JWTAlgorithmValidation |  | N/A |
| security_test | TestSecurity_JWTAlgorithmConfusionHandler | ✓ | N/A |
| security_test | TestSecurity_JWTExpiredToken |  | N/A |
| security_test | TestSecurity_SSRFProtection |  | N/A |
| security_test | TestSecurity_SSRFProtectionHandler | ✓ | N/A |
| security_test | TestSecurity_InputSanitization |  | N/A |
| security_test | TestSecurity_SecurityHeaders |  | N/A |
| security_test | TestSecurity_CircuitBreaker |  | N/A |
| security_test | TestSecurity_RateLimiterPerClient |  | N/A |
| security_test | TestSecurity_GetClientIP |  | N/A |
| security_test | TestSecurity_PathTraversalDetection |  | N/A |
| security_test | TestSecurity_BatchSizeLimit |  | N/A |
| security_test | TestSecurity_MaxConcurrentRequests |  | N/A |
| security_test | TestSecurity_MaxBodySize |  | N/A |
| security_test | TestSecurity_TimeoutConfiguration |  | N/A |
| security_test | TestSecurity_EmptyJWTSecretBlocksRequests |  | N/A |
| security_test | TestSecurity_CorrelationIDGeneration |  | N/A |
| security_test | TestSecurity_IPMasking |  | N/A |
| security_test | TestSecurity_UnsafeIPValidation |  | N/A |
| security_test | TestSecurity_LakeraNilFallback |  | N/A |
| security_test | TestSecurity_LakeraMockError |  | N/A |
| security_test | TestSecurity_LakeraMockBlock |  | N/A |
| security_test | TestSecurity_RateLimiterCleanup |  | N/A |
| security_test | TestSecurity_RateLimiterMaxClients |  | N/A |
| cleanup | CleanupNow |  | int, error |
| cleanup | CleanupWithTTL |  | int, error |
| dlq | Size |  | int, error |
| dlq | Cleanup |  | int, error |
| dlq | GetMessages |  | []DLQMessage, error |
| dlq | Peek |  | *DLQMessage, error |
| dlq | NewDLQ |  | *DLQ, error |
| dlq_test | TestDLQ_Enqueue_SavesFile |  | N/A |
| dlq_test | TestDLQ_Replay_ExecutesHandler | ✓ | N/A |
| dlq_test | TestDLQ_Replay_FIFOOrder |  | N/A |
| dlq_test | TestDLQ_Size_CountsFiles |  | N/A |
| dlq_test | TestDLQ_Cleanup_EliminatesExpiredMessages |  | N/A |
| dlq_test | TestDLQ_CleanupWithTTL |  | N/A |
| dlq_test | TestDLQ_Enqueue_GeneratesID |  | N/A |
| dlq_test | TestDLQ_Enqueue_SetsTimestamp |  | N/A |
| dlq_test | TestDLQ_Remove_DeletesMessage |  | N/A |
| dlq_test | TestDLQ_Remove_NotFound |  | N/A |
| dlq_test | TestDLQ_Peek_ReturnsOldestMessage |  | N/A |
| dlq_test | TestDLQ_Peek_EmptyQueue |  | N/A |
| dlq_test | TestDLQ_GetMessages_ReturnsAllMessages |  | N/A |
| dlq_test | TestDLQ_Enqueue_NilRequest |  | N/A |
| dlq_test | TestDLQ_Replay_NilHandler | ✓ | N/A |
| dlq_test | TestDLQ_NewDLQ_CreatesDirectory |  | N/A |
| dlq_test | TestDLQ_Replay_ContextCancellation |  | N/A |
| dlq_test | TestDLQ_SetLogger |  | N/A |
| dlq_test | TestStartCleanupGoroutine |  | N/A |
| test_cilium_policies | TestCiliumPolicies_DefaultDeny |  | N/A |
| test_cilium_policies | TestCiliumPolicies_DNSWhitelist |  | N/A |
| test_cilium_policies | TestCiliumPolicies_LLMEndpoints |  | N/A |
| test_cilium_policies | TestCiliumPolicies_TargetDomains |  | N/A |
| test_cilium_policies | TestCiliumPolicies_HubbleLogs |  | N/A |
| test_cilium_policies | TestCiliumPolicies_PolicyEnforcement |  | N/A |
| test_cilium_policies | TestCiliumPolicies_MeshCommunication |  | N/A |
| test_falco_detection | TestFalcoDetection_ShellSpawn |  | N/A |
| test_falco_detection | TestFalcoDetection_EtcWrite |  | N/A |
| test_falco_detection | TestFalcoDetection_TalonResponse |  | N/A |
| test_falco_detection | TestFalcoDetection_FalsePositives |  | N/A |
| test_falco_detection | TestFalcoDetection_FalcoRulesValidation |  | N/A |
| test_semantic_firewall | TestSemanticFirewall_ValidJSONRPC |  | N/A |
| test_semantic_firewall | TestSemanticFirewall_MalformedJSONRPC |  | N/A |
| test_semantic_firewall | TestSemanticFirewall_MaliciousPromptInjection |  | N/A |
| test_semantic_firewall | TestSemanticFirewall_RateLimiting |  | N/A |
| test_semantic_firewall | TestSemanticFirewall_LakeraTimeoutHandling |  | N/A |
| test_semantic_firewall | TestSemanticFirewall_HealthEndpoints |  | N/A |
| test_semantic_firewall | TestSemanticFirewall_AuthValidation |  | N/A |
| governance_test | TestSpecImmutability |  | N/A |
| governance_test | TestContractEnforcement |  | N/A |
| governance_test | TestNoScopeExpansion |  | N/A |
| governance_test | TestGovernance_SDDWorkflow |  | N/A |
| governance_test | TestGovernance_ChangeLifecycle |  | N/A |
| governance_test | TestGovernance_ValidationEnforcement |  | N/A |
| network_security_test | TestEgress_BlockedNonWhitelistedDomain |  | N/A |
| network_security_test | TestEgress_Allowed_api_anthropic_com |  | N/A |
| network_security_test | TestDNS_Only_CoreDNS |  | N/A |
| network_security_test | TestIngress_DefaultDeny |  | N/A |
| network_security_test | TestL7Protocol_DROP_on_C2 |  | N/A |
| network_security_test | TestNetworkSecurity_CiliumPolicyValidation |  | N/A |
| runtime_security_test | TestReverseShell_bash_i |  | N/A |
| runtime_security_test | TestReverseShell_netcat |  | N/A |
| runtime_security_test | TestReverseShell_python |  | N/A |
| runtime_security_test | TestReverseShell_curl_pipe_bash |  | N/A |
| runtime_security_test | TestFileWrite_etc_passwd |  | N/A |
| runtime_security_test | TestExec_from_unusual_directory |  | N/A |
| runtime_security_test | TestRuntimeSecurity_FalcoAlertValidation |  | N/A |
| semantic_proxy_test | TestPromptInjection_Direct |  | N/A |
| semantic_proxy_test | TestPromptInjection_Base64Encoded |  | N/A |
| semantic_proxy_test | TestPromptInjection_UnicodeObfuscation |  | N/A |
| semantic_proxy_test | TestJailbreak_IgnorePreviousInstructions |  | N/A |
| semantic_proxy_test | TestJailbreak_CharacterRolePlay |  | N/A |
| semantic_proxy_test | TestContextExhaustion_TokenPadding |  | N/A |
| semantic_proxy_test | TestRateLimiting_Enforced |  | N/A |
| cluster | GetPodLogs |  | string, error |
| cluster | ListPods |  | []corev1.Pod, error |
| cluster | PodExists |  | *corev1.Pod, bool, error |
| cluster | ExecInPod |  | string, string, error |
| cluster | NewClient |  | *Client, error |
| utils | SetAuth |  | N/A |
| utils | SendJSONRPC |  | *JSONRPCResponse, error |
| utils | SendRawJSON |  | int, string, error |
| utils | GetHealth |  | map[string]interface{}, error |
| utils | GetMetrics |  | map[string]interface{}, error |

## Types and Structs

| Module | Type | Kind | Exported |
|--------|------|------|----------|
| config | ConfigFile | struct | ✓|
| cors | CORSMiddleware | struct | ✓|
| integration_test | MockMCPBackend | struct | ✓|
| jsonrpc | JSONRPCError | struct | ✓|
| jsonrpc | JSONRPCRequest | struct | ✓|
| jsonrpc | JSONRPCResponse | struct | ✓|
| jsonrpc | ToolCallParams | struct | ✓|
| jsonrpc | ToolsListParams | struct | ✓|
| jsonrpc | ParsedRequest | struct | ✓|
| lakera | LakeraConfig | struct | ✓|
| lakera | LakeraClient | struct | ✓|
| lakera | LakeraResponse | struct | ✓|
| lakera | LakeraChecker | interface | ✓|
| logger | LogEntry | struct | ✓|
| logger | Logger | struct | ✓|
| main | Config | struct | ✓|
| main | healthResponse | struct | |
| main | securityHeaderHandler | struct | |
| prometheus | PrometheusMetrics | struct | ✓|
| proxy | ProxyConfig | struct | ✓|
| proxy | Proxy | struct | ✓|
| proxy | ClientRateLimiter | struct | ✓|
| proxy | clientBucket | struct | |
| proxy | RateLimiter | struct | ✓|
| proxy | Metrics | struct | ✓|
| proxy | CircuitBreaker | struct | ✓|
| proxy | statusWriter | struct | |
| retry_client | RetryableError | struct | ✓|
| retry_client | RetryClient | struct | ✓|
| retry_client | RetryConfig | struct | ✓|
| retry_client_test | testReadCloser | struct | |
| security_test | mockLakeraClient | struct | |
| cleanup | CleanupConfig | struct | ✓|
| dlq | FailedRequest | struct | ✓|
| dlq | DLQMessage | struct | ✓|
| dlq | DLQ | struct | ✓|
| dlq | DLQConfig | struct | ✓|
| cluster | ClusterConfig | struct | ✓|
| cluster | Client | struct | ✓|
| utils | HTTPClient | struct | ✓|
| utils | JSONRPCRequest | struct | ✓|
| utils | JSONRPCResponse | struct | ✓|
| utils | JSONRPCError | struct | ✓|
| utils | ToolCallParams | struct | ✓|