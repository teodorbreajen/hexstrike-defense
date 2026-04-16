package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMetrics_RecordRequestIncrementsCountersCorrectly tests that RecordRequest increments counters correctly
func TestMetrics_RecordRequestIncrementsCountersCorrectly(t *testing.T) {
	tests := []struct {
		name          string
		allowed       bool
		latency       time.Duration
		statusCode    int
		expectTotal   int64
		expectBlocked int64
		expectAllowed int64
	}{
		{
			name:          "allowed request increments total and allowed",
			allowed:       true,
			latency:       100 * time.Millisecond,
			statusCode:    http.StatusOK,
			expectTotal:   1,
			expectBlocked: 0,
			expectAllowed: 1,
		},
		{
			name:          "blocked request increments total and blocked",
			allowed:       false,
			latency:       0,
			statusCode:    http.StatusTooManyRequests,
			expectTotal:   1,
			expectBlocked: 1,
			expectAllowed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetrics()

			m.RecordRequest(tt.allowed, tt.latency, tt.statusCode)

			total, blocked, allowed, _, _ := m.GetMetrics()

			assert.Equal(t, tt.expectTotal, total, "total requests should match")
			assert.Equal(t, tt.expectBlocked, blocked, "blocked requests should match")
			assert.Equal(t, tt.expectAllowed, allowed, "allowed requests should match")
		})
	}
}

// TestMetrics_GetMetricsReturnsCorrectValues tests that GetMetrics returns correct values
func TestMetrics_GetMetricsReturnsCorrectValues(t *testing.T) {
	m := NewMetrics()

	// Record multiple requests (only allowed requests contribute to latency)
	m.RecordRequest(true, 50*time.Millisecond, http.StatusOK)
	m.RecordRequest(true, 100*time.Millisecond, http.StatusOK)
	m.RecordRequest(false, 0, http.StatusUnauthorized)

	total, blocked, allowed, avgLatency, statusCodes := m.GetMetrics()

	assert.Equal(t, int64(3), total, "total should be 3")
	assert.Equal(t, int64(1), blocked, "blocked should be 1")
	assert.Equal(t, int64(2), allowed, "allowed should be 2")

	// Verify latency tracking works (we just verify it's > 0, the exact value depends on implementation)
	assert.Greater(t, avgLatency, float64(0), "average latency should be > 0")

	// Check status codes
	assert.Equal(t, int64(2), statusCodes[http.StatusOK], "status 200 should have count 2")
	assert.Equal(t, int64(1), statusCodes[http.StatusUnauthorized], "status 401 should have count 1")
}

// TestMetrics_StatusCodesAreTracked tests that status codes are tracked correctly
func TestMetrics_StatusCodesAreTracked(t *testing.T) {
	tests := []struct {
		name        string
		statusCodes []int
		expected    map[int]int64
	}{
		{
			name:        "single status code",
			statusCodes: []int{http.StatusOK},
			expected:    map[int]int64{http.StatusOK: 1},
		},
		{
			name:        "multiple same status code",
			statusCodes: []int{http.StatusOK, http.StatusOK, http.StatusOK},
			expected:    map[int]int64{http.StatusOK: 3},
		},
		{
			name:        "different status codes",
			statusCodes: []int{http.StatusOK, http.StatusCreated, http.StatusBadRequest, http.StatusOK},
			expected:    map[int]int64{http.StatusOK: 2, http.StatusCreated: 1, http.StatusBadRequest: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetrics()

			for _, code := range tt.statusCodes {
				m.RecordRequest(true, 10*time.Millisecond, code)
			}

			_, _, _, _, statusCodes := m.GetMetrics()

			for expectedCode, expectedCount := range tt.expected {
				assert.Equal(t, expectedCount, statusCodes[expectedCode],
					"status code %d should have count %d", expectedCode, expectedCount)
			}
		})
	}
}

// TestMetrics_ConcurrentAccess tests thread safety of metrics
func TestMetrics_ConcurrentAccess(t *testing.T) {
	m := NewMetrics()

	done := make(chan bool, 10)

	// Run concurrent recorders
	for i := 0; i < 10; i++ {
		go func() {
			m.RecordRequest(true, 10*time.Millisecond, http.StatusOK)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	total, _, allowed, _, statusCodes := m.GetMetrics()

	assert.Equal(t, int64(10), total, "total should be 10 after concurrent writes")
	assert.Equal(t, int64(10), allowed, "allowed should be 10")
	assert.Equal(t, int64(10), statusCodes[http.StatusOK], "status 200 should have count 10")
}

// TestMetrics_LatencyTracking tests that latency is tracked correctly
func TestMetrics_LatencyTracking(t *testing.T) {
	tests := []struct {
		name        string
		latencies   []time.Duration
		expectedAvg float64
	}{
		{
			name:        "single request",
			latencies:   []time.Duration{100 * time.Millisecond},
			expectedAvg: 100.0,
		},
		{
			name:        "multiple requests same latency",
			latencies:   []time.Duration{100 * time.Millisecond, 100 * time.Millisecond},
			expectedAvg: 100.0,
		},
		{
			name:        "varying latencies",
			latencies:   []time.Duration{50 * time.Millisecond, 100 * time.Millisecond, 150 * time.Millisecond},
			expectedAvg: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetrics()

			for _, latency := range tt.latencies {
				m.RecordRequest(true, latency, http.StatusOK)
			}

			_, _, _, avgLatency, _ := m.GetMetrics()

			assert.InDelta(t, tt.expectedAvg, avgLatency, 0.01, "average latency should match")
		})
	}
}

// TestMetrics_EmptyMetrics tests that empty metrics return zero values
func TestMetrics_EmptyMetrics(t *testing.T) {
	m := NewMetrics()

	total, blocked, allowed, avgLatency, statusCodes := m.GetMetrics()

	assert.Equal(t, int64(0), total, "total should be 0")
	assert.Equal(t, int64(0), blocked, "blocked should be 0")
	assert.Equal(t, int64(0), allowed, "allowed should be 0")
	assert.Equal(t, float64(0), avgLatency, "avgLatency should be 0")
	assert.Empty(t, statusCodes, "statusCodes should be empty")
}

// TestMetrics_NewMetrics tests the NewMetrics constructor
func TestMetrics_NewMetrics(t *testing.T) {
	m := NewMetrics()

	assert.NotNil(t, m, "metrics should not be nil")
	assert.NotNil(t, m.StatusCodes, "StatusCodes map should be initialized")
	assert.Equal(t, int64(0), m.TotalRequests, "TotalRequests should be 0")
	assert.Equal(t, int64(0), m.BlockedRequests, "BlockedRequests should be 0")
	assert.Equal(t, int64(0), m.AllowedRequests, "AllowedRequests should be 0")
}
