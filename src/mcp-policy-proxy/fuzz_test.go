//go:build fuzz

package main

import (
	"testing"
	"time"
)

// Fuzz sanitization functions to find bypasses

func FuzzSanitizeToolInput(f *testing.F) {
	// Add test cases from corpus
	testCases := []struct {
		toolName string
		args     string
	}{
		{"simple_tool", "simple argument"},
		{"tool_with_underscore", "argument with spaces"},
		{"list", "limit=10"},
		{"read", `{"path": "/tmp/file"}`},
	}

	for _, tc := range testCases {
		f.Add(tc.toolName, tc.args)
	}

	f.Fuzz(func(t *testing.T, toolName, args string) {
		// Should never panic
		_, _, _ = sanitizeToolInput(toolName, args)
	})
}

func FuzzIsInternalURL(f *testing.F) {
	// Test cases for SSRF bypass attempts
	testCases := []string{
		"http://localhost/",
		"http://127.0.0.1/",
		"http://10.0.0.1/",
		"http://172.16.0.1/",
		"http://192.168.0.1/",
		"https://example.com/",
		"https://api.openai.com/",
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, rawURL string) {
		// Should never panic
		_ = isInternalURL(rawURL)
	})
}

func FuzzParseJSONRPC(f *testing.F) {
	// Valid JSON-RPC requests
	testCases := []string{
		`{"jsonrpc": "2.0", "method": "tools/list", "id": 1}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "test", "arguments": "{}"}, "id": 1}`,
		`[{"jsonrpc": "2.0", "method": "tools/list", "id": 1}]`,
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data string) {
		// Should never panic
		_, _ = ParseJSONRPC([]byte(data))
	})
}

func FuzzValidateBackendURL(f *testing.F) {
	testCases := []struct {
		backend string
		path    string
	}{
		{"http://backend:8080", "/mcp"},
		{"http://localhost:9090", "/api"},
	}

	for _, tc := range testCases {
		f.Add(tc.backend, tc.path)
	}

	f.Fuzz(func(t *testing.T, backend, path string) {
		// Should never panic
		_, _ = validateBackendURL(backend, path)
	})
}

// Fuzz token bucket rate limiter
func FuzzTokenBucket(f *testing.F) {
	f.Fuzz(func(t *testing.T, clientID string, numRequests int) {
		rl := NewClientRateLimiter(60)

		// Clamp numRequests to reasonable range
		if numRequests < 0 {
			numRequests = 0
		}
		if numRequests > 1000 {
			numRequests = 1000
		}

		// Make multiple requests - should never panic
		for i := 0; i < numRequests; i++ {
			_ = rl.Allow(clientID)
		}
	})
}

// Fuzz circuit breaker
func FuzzCircuitBreaker(f *testing.F) {
	f.Fuzz(func(t *testing.T, numFailures, numSuccesses int) {
		cb := NewCircuitBreaker(5, 100*time.Millisecond)

		// Clamp values
		if numFailures < 0 {
			numFailures = 0
		}
		if numFailures > 20 {
			numFailures = 20
		}
		if numSuccesses < 0 {
			numSuccesses = 0
		}
		if numSuccesses > 20 {
			numSuccesses = 20
		}

		// Record failures
		for i := 0; i < numFailures; i++ {
			cb.RecordFailure()
		}

		// Record successes
		for i := 0; i < numSuccesses; i++ {
			cb.RecordSuccess()
		}

		// Check state is consistent
		state := cb.GetState()
		if state < CircuitClosed || state > CircuitHalfOpen {
			t.Errorf("Invalid circuit breaker state: %d", state)
		}
	})
}

// Fuzz metrics
func FuzzMetrics(f *testing.F) {
	f.Fuzz(func(t *testing.T, allowed bool, statusCode int, latencyMs int) {
		m := NewMetrics()

		// Clamp values
		if latencyMs < 0 {
			latencyMs = 0
		}
		if latencyMs > 60000 {
			latencyMs = 60000
		}
		if statusCode < 0 {
			statusCode = 200
		}
		if statusCode > 599 {
			statusCode = 200
		}

		latency := time.Duration(latencyMs) * time.Millisecond

		// Should never panic
		m.RecordRequest(allowed, latency, statusCode)

		// GetMetrics should work
		_, _, _, _, _ = m.GetMetrics()
	})
}
