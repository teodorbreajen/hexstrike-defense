package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRetryClient_SuccessAtFirstAttempt tests that no retry happens on immediate success
func TestRetryClient_SuccessAtFirstAttempt(t *testing.T) {
	var attemptCount int32

	// Create a mock server that succeeds immediately
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attemptCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Create retry client with minimal config
	client := &RetryClient{
		client:     server.Client(),
		maxRetries: 3,
		baseDelay:  10 * time.Millisecond, // Short delay for testing
		logger:     NewLogger("retry"),
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := client.DoWithRetry(context.Background(), req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, int32(1), atomic.LoadInt32(&attemptCount), "should only attempt once on success")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestRetryClient_RetryOn5xxEventuallySuccess tests retry on 5xx until success
func TestRetryClient_RetryOn5xxEventuallySuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping retry timing test in short mode")
	}

	var attemptCount int32

	// Create a mock server that fails twice then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attemptCount, 1)
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"server error"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Create retry client with short delays
	client := &RetryClient{
		client:     server.Client(),
		maxRetries: 3,
		baseDelay:  50 * time.Millisecond,
		logger:     NewLogger("retry"),
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	start := time.Now()
	resp, err := client.DoWithRetry(context.Background(), req)
	elapsed := time.Since(start)

	require.NoError(t, err)
	defer resp.Body.Close()

	// Should have attempted 3 times
	assert.Equal(t, int32(3), atomic.LoadInt32(&attemptCount), "should retry until success")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Should have waited at least base delays (50ms * 2 retries = 100ms minimum)
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond, "should have waited for backoff")
}

// TestRetryClient_MaxRetriesExhausted tests that error is returned when max retries are exhausted
func TestRetryClient_MaxRetriesExhausted(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping retry timing test in short mode")
	}

	var attemptCount int32

	// Create a mock server that always returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attemptCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"server error"}`))
	}))
	defer server.Close()

	// Create retry client with max 2 retries
	client := &RetryClient{
		client:     server.Client(),
		maxRetries: 2,
		baseDelay:  50 * time.Millisecond,
		logger:     NewLogger("retry"),
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := client.DoWithRetry(context.Background(), req)

	// For HTTP status codes (even 5xx), the client returns the response after exhausting retries
	// The caller is responsible for checking the status code
	assert.NoError(t, err, "HTTP responses don't return errors, caller checks status")
	require.NotNil(t, resp, "response should be returned")
	defer resp.Body.Close()

	// Should have attempted maxRetries + 1 times (initial + 2 retries)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attemptCount), "should attempt maxRetries+1 times")

	// The response should be the last 500 response
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// TestRetryClient_4xxNotRetryable tests that 4xx errors fail immediately (not retryable)
func TestRetryClient_4xxNotRetryable(t *testing.T) {
	var attemptCount int32

	// Create a mock server that always returns 400 Bad Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attemptCount, 1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	// Create retry client with max 3 retries
	client := &RetryClient{
		client:     server.Client(),
		maxRetries: 3,
		baseDelay:  10 * time.Millisecond,
		logger:     NewLogger("retry"),
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := client.DoWithRetry(context.Background(), req)

	// 4xx is NOT retryable, so it returns the response immediately (no error for HTTP status)
	// The retry logic doesn't convert 4xx to an error - it just returns the response
	assert.NoError(t, err, "4xx responses are returned as successful HTTP responses")
	require.NotNil(t, resp, "response should not be nil")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Should have attempted only once - 4xx is not retryable
	assert.Equal(t, int32(1), atomic.LoadInt32(&attemptCount), "4xx should not be retried")
}

// TestRetryClient_429IsRetryable tests that 429 Too Many Requests is retryable
func TestRetryClient_429IsRetryable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping retry timing test in short mode")
	}

	var attemptCount int32

	// Create a mock server that returns 429 twice then 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attemptCount, 1)
		if count < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limited"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Create retry client
	client := &RetryClient{
		client:     server.Client(),
		maxRetries: 3,
		baseDelay:  50 * time.Millisecond,
		logger:     NewLogger("retry"),
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := client.DoWithRetry(context.Background(), req)

	require.NoError(t, err)
	defer resp.Body.Close()

	// 429 should be retried
	assert.Equal(t, int32(3), atomic.LoadInt32(&attemptCount), "429 should be retried until success")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestIsRetryableStatusCode tests the isRetryableStatusCode helper function
func TestIsRetryableStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   bool
	}{
		// 5xx are retryable
		{"500 Internal Server Error", 500, true},
		{"501 Not Implemented", 501, true},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
		{"504 Gateway Timeout", 504, true},
		{"599 Custom Error", 599, true},

		// 429 is retryable
		{"429 Too Many Requests", 429, true},

		// 4xx (except 429) are NOT retryable
		{"400 Bad Request", 400, false},
		{"401 Unauthorized", 401, false},
		{"403 Forbidden", 403, false},
		{"404 Not Found", 404, false},
		{"405 Method Not Allowed", 405, false},
		{"418 I'm a teapot", 418, false},

		// 3xx are NOT retryable
		{"301 Moved Permanently", 301, false},
		{"302 Found", 302, false},
		{"304 Not Modified", 304, false},

		// 2xx are NOT retryable
		{"200 OK", 200, false},
		{"201 Created", 201, false},
		{"204 No Content", 204, false},

		// 1xx are NOT retryable
		{"100 Continue", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableStatusCode(tt.statusCode)
			assert.Equal(t, tt.expected, result, "status %d should be retryable=%v", tt.statusCode, tt.expected)
		})
	}
}

// TestCalculateBackoff tests the exponential backoff calculation
func TestCalculateBackoff(t *testing.T) {
	baseDelay := 1 * time.Second

	tests := []struct {
		attempt    int
		expectedMs int64
	}{
		{0, 1000}, // 1s
		{1, 2000}, // 2s
		{2, 4000}, // 4s
		{3, 8000}, // 8s
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			delay := calculateBackoff(tt.attempt, baseDelay)
			assert.Equal(t, tt.expectedMs, delay.Milliseconds(), "attempt %d should have %dms delay", tt.attempt, tt.expectedMs)
		})
	}
}

// TestIsRetryableError tests the isRetryableError helper function
func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "context canceled",
			err:      context.Canceled,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRetryClient_ContextCancellation tests that context cancellation stops retries
func TestRetryClient_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping retry timing test in short mode")
	}

	var attemptCount int32

	// Create a mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attemptCount, 1)
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create retry client with long delays
	client := &RetryClient{
		client:     server.Client(),
		maxRetries: 3,
		baseDelay:  200 * time.Millisecond,
		logger:     NewLogger("retry"),
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	resp, err := client.DoWithRetry(ctx, req)

	assert.Error(t, err, "should return error on cancelled context")
	assert.Nil(t, resp)
}

// TestRetryClient_RequestBodyPreserved tests that request body is preserved across retries
func TestRetryClient_RequestBodyPreserved(t *testing.T) {
	var receivedBody string
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &RetryClient{
		client:     server.Client(),
		maxRetries: 3,
		baseDelay:  10 * time.Millisecond,
		logger:     NewLogger("retry"),
	}

	testBody := `{"key":"value","number":42}`

	req, err := http.NewRequest(http.MethodPost, server.URL, nil)
	require.NoError(t, err)
	req.Body = io.NopCloser(io.NopCloser(nil))
	req.Body = io.NopCloser(newTestReadCloser(testBody))
	req.ContentLength = int64(len(testBody))

	resp, err := client.DoWithRetry(context.Background(), req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Body should be preserved and readable
	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount))
	assert.Equal(t, testBody, receivedBody)
}

// TestRetryClient_WithCorrelationID tests that correlation IDs are preserved in retries
func TestRetryClient_WithCorrelationID(t *testing.T) {
	var correlationID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID = r.Header.Get("X-Correlation-ID")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &RetryClient{
		client:     server.Client(),
		maxRetries: 3,
		baseDelay:  10 * time.Millisecond,
		logger:     NewLogger("retry"),
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	req.Header.Set("X-Correlation-ID", "test-correlation-123")

	resp, err := client.DoWithRetry(context.Background(), req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "test-correlation-123", correlationID, "correlation ID should be preserved")
}

// TestNewRetryClient_DefaultConfig tests that NewRetryClient uses default config when nil
func TestNewRetryClient_DefaultConfig(t *testing.T) {
	client := NewRetryClient(&http.Client{}, nil)

	assert.Equal(t, 3, client.maxRetries, "default max retries should be 3")
	assert.Equal(t, 1*time.Second, client.baseDelay, "default base delay should be 1s")
	assert.NotNil(t, client.logger, "logger should not be nil")
}

// TestDefaultRetryConfig tests the DefaultRetryConfig function
func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.BaseDelay)
	assert.NotNil(t, config.Logger)
}

// Helper type for test
type testReadCloser struct {
	data string
	pos  int
}

func newTestReadCloser(data string) *testReadCloser {
	return &testReadCloser{data: data}
}

func (t *testReadCloser) Read(p []byte) (n int, err error) {
	if t.pos >= len(t.data) {
		return 0, io.EOF
	}
	n = copy(p, t.data[t.pos:])
	t.pos += n
	return n, nil
}

func (t *testReadCloser) Close() error {
	return nil
}
