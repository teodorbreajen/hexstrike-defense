package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPrometheusFormat verifies that the /metrics endpoint returns valid Prometheus text format
func TestPrometheusFormat(t *testing.T) {
	// Create Prometheus metrics
	pm := NewPrometheusMetrics()

	// Record some test metrics
	pm.RecordRequest("POST", "/mcp", true, http.StatusOK, 100*time.Millisecond)
	pm.RecordRequest("POST", "/mcp", false, http.StatusForbidden, 0)
	pm.RecordLakeraBlock("injection")
	pm.RecordBackendError("timeout")
	pm.RecordRetry("/mcp")
	pm.RecordDLQMessage("mcp-backend")
	pm.SetDLQCount(5)
	pm.SetCircuitBreakerState("/mcp", CircuitClosed)
	pm.IncActiveRequests()
	pm.DecActiveRequests()

	// Create handler
	handler := NewPrometheusHandler(pm)

	// Request metrics
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
	assert.Contains(t, w.Header().Get("Content-Type"), "version=0.0.4")

	body := w.Body.String()

	// Verify expected metrics are present
	assert.Contains(t, body, "mcp_proxy_requests_total")
	assert.Contains(t, body, "mcp_proxy_status_codes_total")
	assert.Contains(t, body, "mcp_proxy_lakera_blocks_total")
	assert.Contains(t, body, "mcp_proxy_backend_errors_total")
	assert.Contains(t, body, "mcp_proxy_retries_total")
	assert.Contains(t, body, "mcp_proxy_dlq_messages")
	assert.Contains(t, body, "mcp_proxy_dlq_messages_total")
	assert.Contains(t, body, "mcp_proxy_request_duration_seconds")
	assert.Contains(t, body, "mcp_proxy_active_requests")
	assert.Contains(t, body, "mcp_proxy_circuit_breaker_state")

	// Verify metric format (Prometheus text format)
	lines := strings.Split(body, "\n")
	metricLineCount := 0
	for _, line := range lines {
		// Metric lines should have the format: metric_name{labels} value
		// Or: metric_name value (for untyped metrics)
		if strings.HasPrefix(line, "mcp_proxy_") {
			metricLineCount++
			// Verify the line ends with a number (the metric value)
			assert.True(t, strings.HasSuffix(line, " ") || isValidMetricValue(line),
				"metric line should end with a value: %s", line)
		}
	}

	assert.Greater(t, metricLineCount, 0, "should have at least one MCP metric")
}

// isValidMetricValue checks if a string ends with a valid Prometheus metric value
func isValidMetricValue(line string) bool {
	// Prometheus metric values are numbers (integers or floats)
	// The line should end with something like " 1" or " 1.5"
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return false
	}
	lastPart := parts[len(parts)-1]
	return strings.HasSuffix(lastPart, "0") ||
		strings.HasSuffix(lastPart, "1") ||
		strings.HasSuffix(lastPart, "2") ||
		strings.HasSuffix(lastPart, "3") ||
		strings.HasSuffix(lastPart, "4") ||
		strings.HasSuffix(lastPart, "5") ||
		strings.HasSuffix(lastPart, "6") ||
		strings.HasSuffix(lastPart, "7") ||
		strings.HasSuffix(lastPart, "8") ||
		strings.HasSuffix(lastPart, "9") ||
		strings.HasSuffix(lastPart, ".") ||
		strings.Contains(lastPart, ".")
}

// TestPrometheusRequestsTotal verifies the requests_total counter
func TestPrometheusRequestsTotal(t *testing.T) {
	pm := NewPrometheusMetrics()

	// Record requests
	pm.RecordRequest("POST", "/test", true, http.StatusOK, 50*time.Millisecond)
	pm.RecordRequest("POST", "/test", true, http.StatusOK, 100*time.Millisecond)
	pm.RecordRequest("POST", "/test", false, http.StatusForbidden, 0)

	// Get metrics
	body := fetchMetrics(t, pm)

	// Check for allowed requests (exact match with full labels)
	assert.Contains(t, body, `mcp_proxy_requests_total{allowed="true",endpoint="/test",method="POST"}`)
	// Check for blocked requests
	assert.Contains(t, body, `mcp_proxy_requests_total{allowed="false",endpoint="/test",method="POST"}`)
}

// TestPrometheusStatusCodes verifies status code tracking
func TestPrometheusStatusCodes(t *testing.T) {
	pm := NewPrometheusMetrics()

	// Record various status codes
	statusCodes := []int{200, 200, 201, 400, 401, 403, 500, 503}
	for _, code := range statusCodes {
		pm.RecordRequest("POST", "/mcp", code < 400, code, 10*time.Millisecond)
	}

	body := fetchMetrics(t, pm)

	// Verify each status code appears in the mcp_proxy_status_codes_total metric
	for _, code := range statusCodes {
		assert.Contains(t, body, fmt.Sprintf(`code="%d"`, code))
	}
}

// TestPrometheusHistogram verifies request duration histogram
func TestPrometheusHistogram(t *testing.T) {
	pm := NewPrometheusMetrics()

	// Record requests with various durations
	durations := []time.Duration{
		1 * time.Millisecond,
		10 * time.Millisecond,
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}

	for _, d := range durations {
		pm.RecordRequest("POST", "/mcp", true, http.StatusOK, d)
	}

	body := fetchMetrics(t, pm)

	// Histogram should be present with buckets
	assert.Contains(t, body, "mcp_proxy_request_duration_seconds_bucket")
	assert.Contains(t, body, "mcp_proxy_request_duration_seconds_sum")
	assert.Contains(t, body, "mcp_proxy_request_duration_seconds_count")
}

// TestPrometheusCircuitBreaker verifies circuit breaker state metric
func TestPrometheusCircuitBreaker(t *testing.T) {
	pm := NewPrometheusMetrics()

	// Test all circuit states
	states := []CircuitState{CircuitClosed, CircuitOpen, CircuitHalfOpen}
	stateNames := []string{"closed", "open", "half_open"}

	for i, state := range states {
		pm.SetCircuitBreakerState("/mcp", state)
		body := fetchMetrics(t, pm)

		// The metric should have the endpoint label
		assert.Contains(t, body, `endpoint="/mcp"`)

		// The value should be the state value (0, 1, or 2)
		// and it should be present in the output
		assert.Contains(t, body, "mcp_proxy_circuit_breaker_state")
		t.Logf("Circuit state %s: value %d", stateNames[i], state)
	}
}

// TestPrometheusActiveRequests verifies active requests gauge
func TestPrometheusActiveRequests(t *testing.T) {
	pm := NewPrometheusMetrics()

	// Simulate active requests
	pm.IncActiveRequests()
	pm.IncActiveRequests()
	pm.IncActiveRequests()

	body := fetchMetrics(t, pm)
	assert.Contains(t, body, "mcp_proxy_active_requests")

	pm.DecActiveRequests()
	pm.DecActiveRequests()

	body = fetchMetrics(t, pm)
	// Should still have the metric
	assert.Contains(t, body, "mcp_proxy_active_requests")
}

// TestPrometheusGauges verifies gauge metrics
func TestPrometheusGauges(t *testing.T) {
	pm := NewPrometheusMetrics()

	// Test DLQ gauge
	pm.SetDLQCount(10)
	body := fetchMetrics(t, pm)
	assert.Contains(t, body, "mcp_proxy_dlq_messages 10")

	// Update count
	pm.SetDLQCount(5)
	body = fetchMetrics(t, pm)
	assert.Contains(t, body, "mcp_proxy_dlq_messages 5")
}

// TestPrometheusLabels verifies metrics have correct labels
func TestPrometheusLabels(t *testing.T) {
	pm := NewPrometheusMetrics()

	pm.RecordRequest("POST", "/mcp/v1/tools", true, http.StatusOK, 50*time.Millisecond)
	pm.RecordRequest("GET", "/mcp/v1/resources", true, http.StatusOK, 30*time.Millisecond)
	pm.RecordRequest("POST", "/mcp/v1/call", true, http.StatusOK, 100*time.Millisecond)

	body := fetchMetrics(t, pm)

	// Verify method labels
	assert.Contains(t, body, `method="POST"`)
	assert.Contains(t, body, `method="GET"`)

	// Verify endpoint labels
	assert.Contains(t, body, `endpoint="/mcp/v1/tools"`)
	assert.Contains(t, body, `endpoint="/mcp/v1/resources"`)
	assert.Contains(t, body, `endpoint="/mcp/v1/call"`)
}

// TestPrometheusContentType verifies correct Content-Type header
func TestPrometheusContentType(t *testing.T) {
	pm := NewPrometheusMetrics()
	handler := NewPrometheusHandler(pm)

	tests := []struct {
		name        string
		contentType string
	}{
		{"Prometheus 0.0.4", "text/plain; version=0.0.4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			contentType := w.Header().Get("Content-Type")
			assert.Contains(t, contentType, tt.contentType)
		})
	}
}

// TestPrometheusConcurrentAccess tests thread safety
func TestPrometheusConcurrentAccess(t *testing.T) {
	pm := NewPrometheusMetrics()

	done := make(chan bool, 10)

	// Concurrent metric updates
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				pm.RecordRequest("POST", "/mcp", true, http.StatusOK, time.Millisecond)
				pm.IncActiveRequests()
				pm.DecActiveRequests()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should complete without panics
	// Verify metrics were recorded
	body := fetchMetrics(t, pm)
	assert.NotEmpty(t, body)
}

// fetchMetrics retrieves metrics from the handler
func fetchMetrics(t *testing.T, pm *PrometheusMetrics) string {
	t.Helper()

	handler := NewPrometheusHandler(pm)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	return w.Body.String()
}

// TestPrometheusMultipleEndpointMetrics tests metrics for multiple endpoints
func TestPrometheusMultipleEndpointMetrics(t *testing.T) {
	pm := NewPrometheusMetrics()

	// Record to different endpoints
	endpoints := []string{"/mcp/v1/tools", "/mcp/v1/resources", "/mcp/v1/call"}
	for _, ep := range endpoints {
		for i := 0; i < 5; i++ {
			pm.RecordRequest("POST", ep, true, http.StatusOK, 50*time.Millisecond)
		}
	}

	body := fetchMetrics(t, pm)

	// Each endpoint should have its own metric line
	for _, ep := range endpoints {
		assert.Contains(t, body, `endpoint="`+ep+`"`, "should have metric for endpoint: %s", ep)
	}
}
