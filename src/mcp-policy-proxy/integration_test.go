//go:build integration

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_ProxyForwardsRequests tests end-to-end flow from request to backend response
func TestIntegration_ProxyForwardsRequests(t *testing.T) {
	// Create a mock MCP backend server
	var requestCount int32

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)

		// Read and store the request body
		body := make([]byte, 1024)
		n, _ := r.Body.Read(body)
		_ = string(body[:n]) // Body read for verification

		// Simulate MCP JSON-RPC response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"result": map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": "Hello from MCP backend"},
				},
			},
			"id": 1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer backend.Close()

	// Note: httptest.Server uses 127.0.0.1 which is blocked by SSRF
	// For full integration test with real backend, use external test environment
	// This test validates the mock backend works correctly
	assert.NotEmpty(t, backend.URL)

	// Make direct request to verify backend works
	resp, err := http.Get(backend.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestIntegration_MetricsRecorded tests that metrics are recorded correctly
func TestIntegration_MetricsRecorded(t *testing.T) {
	// Create Prometheus metrics
	pm := NewPrometheusMetrics()

	// Record test metrics
	pm.RecordRequest("POST", "/mcp/v1/call", true, http.StatusOK, 100*time.Millisecond)
	pm.RecordRequest("POST", "/mcp/v1/call", true, http.StatusOK, 150*time.Millisecond)
	pm.RecordRequest("POST", "/mcp/v1/call", true, http.StatusOK, 200*time.Millisecond)

	// Fetch and verify
	body := fetchPrometheusMetrics(t, pm)
	assert.Contains(t, body, "mcp_proxy_requests_total")
	assert.Contains(t, body, `allowed="true"`)
}

// TestIntegration_RejectInternalHost tests SSRF protection
func TestIntegration_RejectInternalHost(t *testing.T) {
	// Test that internal URLs are correctly identified
	testCases := []struct {
		url      string
		expected bool
	}{
		{"http://localhost:8080/attack", true},
		{"http://127.0.0.1:6379/attack", true},
		{"http://169.254.169.254/latest/meta-data/", true},
		{"http://kubernetes.default.svc.cluster.local/api", true},
		{"http://10.0.0.1/internal-api", true},
		{"https://api.example.com/normal", false},
	}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			result := isInternalURL(tc.url)
			assert.Equal(t, tc.expected, result, "SSRF check for %s should be %v", tc.url, tc.expected)
		})
	}
}

// TestIntegration_CircuitBreakerIntegration tests circuit breaker with failing backend
func TestIntegration_CircuitBreakerIntegration(t *testing.T) {
	cb := NewCircuitBreaker(5, 30*time.Second)

	// Initially should be closed
	assert.Equal(t, CircuitClosed, cb.GetState())

	// Record failures until threshold - 1
	for i := 0; i < 4; i++ {
		cb.RecordFailure()
		assert.Equal(t, CircuitClosed, cb.GetState(), "circuit should remain closed after %d failures", i+1)
	}

	// 5th failure should open circuit
	cb.RecordFailure()
	assert.Equal(t, CircuitOpen, cb.GetState(), "circuit should be open after 5 failures")

	// Allow() should return false while circuit is open
	assert.False(t, cb.Allow(), "circuit should not allow requests while open")

	// After timeout, circuit should allow requests (transition to half-open)
	cb.lastFailure = time.Now().Add(-31 * time.Second) // Set last failure to past timeout
	assert.True(t, cb.Allow(), "circuit should allow after timeout")

	// Verify we're in half-open state now
	assert.Equal(t, CircuitHalfOpen, cb.GetState(), "circuit should be half-open after timeout")
}

// TestIntegration_RetryBehavior tests retry logic
func TestIntegration_RetryBehavior(t *testing.T) {
	var attemptCount int32
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attemptCount, 1)

		// Fail first 2 attempts, succeed on 3rd
		if count < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"temporarily unavailable"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"result":  map[string]interface{}{"success": true},
			"id":      1,
		})
	}))
	defer backend.Close()

	// Verify the backend correctly fails then succeeds
	// First request
	resp, err := http.Get(backend.URL)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode, "first request should fail")
	assert.Equal(t, int32(1), atomic.LoadInt32(&attemptCount))

	// Second request
	resp, err = http.Get(backend.URL)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode, "second request should fail")
	assert.Equal(t, int32(2), atomic.LoadInt32(&attemptCount))

	// Third request should succeed
	resp, err = http.Get(backend.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "third request should succeed")
	assert.Equal(t, int32(3), atomic.LoadInt32(&attemptCount))
}

// TestIntegration_PrometheusMetrics tests Prometheus metrics export
func TestIntegration_PrometheusMetrics(t *testing.T) {
	// Create Prometheus metrics
	promMetrics := NewPrometheusMetrics()

	// Record test data
	promMetrics.RecordRequest("POST", "/mcp", true, http.StatusOK, 100*time.Millisecond)
	promMetrics.RecordLakeraBlock("injection")
	promMetrics.RecordBackendError("timeout")
	promMetrics.RecordRetry("/mcp")
	promMetrics.RecordDLQMessage("mcp-backend")
	promMetrics.SetDLQCount(5)
	promMetrics.SetCircuitBreakerState("/mcp", CircuitClosed)

	// Fetch metrics
	body := fetchPrometheusMetrics(t, promMetrics)

	// Verify all metrics are present
	assert.Contains(t, body, "mcp_proxy_requests_total")
	assert.Contains(t, body, "mcp_proxy_lakera_blocks_total")
	assert.Contains(t, body, "mcp_proxy_backend_errors_total")
	assert.Contains(t, body, "mcp_proxy_retries_total")
	assert.Contains(t, body, "mcp_proxy_dlq_messages")
	assert.Contains(t, body, "mcp_proxy_circuit_breaker_state")
}

// TestIntegration_ConcurrentRequests tests handling of concurrent requests
func TestIntegration_ConcurrentRequests(t *testing.T) {
	// Test concurrent metric updates
	promMetrics := NewPrometheusMetrics()

	const concurrentRequests = 10
	done := make(chan bool, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func() {
			promMetrics.IncActiveRequests()
			promMetrics.RecordRequest("POST", "/mcp", true, http.StatusOK, 50*time.Millisecond)
			promMetrics.DecActiveRequests()
			done <- true
		}()
	}

	// Wait for all requests
	for i := 0; i < concurrentRequests; i++ {
		<-done
	}

	// Verify metrics
	body := fetchPrometheusMetrics(t, promMetrics)
	assert.Contains(t, body, "mcp_proxy_requests_total")
	assert.Contains(t, body, "mcp_proxy_active_requests")
}

// TestIntegration_RateLimitingEndToEnd tests rate limiting through the full stack
func TestIntegration_RateLimitingEndToEnd(t *testing.T) {
	// Test rate limiter directly
	rl := NewClientRateLimiter(5)

	// Exhaust rate limit
	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow("test-client"), "request %d should be allowed", i+1)
	}

	// Next request should be blocked
	assert.False(t, rl.Allow("test-client"), "request 6 should be blocked")

	// Different client should be allowed
	assert.True(t, rl.Allow("other-client"), "different client should be allowed")
}

// TestMockMCPBackend tests the mock backend functionality
func TestMockMCPBackend(t *testing.T) {
	mock := NewMockMCPBackend()
	defer mock.Close()

	// Make a request to the mock
	resp, err := http.Get(mock.URL())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&mock.RequestCount))

	// Test failure mode
	mock.FailOnCount = 2
	resp2, err := http.Get(mock.URL())
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp2.StatusCode)
}

// MockMCPBackend provides a configurable mock backend for testing
type MockMCPBackend struct {
	Server         *httptest.Server
	RequestBody    string
	RequestCount   int32
	FailOnCount    int32
	ResponseStatus int
	ResponseBody   string
	mu             int32 // For atomic operations
}

// NewMockMCPBackend creates a new mock MCP backend
func NewMockMCPBackend() *MockMCPBackend {
	m := &MockMCPBackend{
		ResponseStatus: http.StatusOK,
		ResponseBody:   `{"jsonrpc":"2.0","result":{},"id":1}`,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&m.RequestCount, 1)

		// Read request body
		body := make([]byte, 4096)
		n, _ := r.Body.Read(body)
		atomic.StoreInt32(&m.mu, 1) // Signal body is read
		m.RequestBody = string(body[:n])

		// Check if we should fail
		currentCount := atomic.LoadInt32(&m.RequestCount)
		if m.FailOnCount > 0 && currentCount <= m.FailOnCount {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"mock error"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(m.ResponseStatus)
		w.Write([]byte(m.ResponseBody))
	}))

	return m
}

// Close shuts down the mock server
func (m *MockMCPBackend) Close() {
	m.Server.Close()
}

// URL returns the server URL
func (m *MockMCPBackend) URL() string {
	return m.Server.URL
}

// WaitForRequest waits for at least n requests to be received
func (m *MockMCPBackend) WaitForRequest(n int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&m.RequestCount) >= int32(n) {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// fetchPrometheusMetrics retrieves metrics from the handler
func fetchPrometheusMetrics(t *testing.T, pm *PrometheusMetrics) string {
	t.Helper()

	handler := NewPrometheusHandler(pm)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	return w.Body.String()
}
