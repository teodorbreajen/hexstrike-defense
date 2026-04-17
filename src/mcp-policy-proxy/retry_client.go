package main

import (
	"context"
	"errors"
	"io"
	"math"
	"net"
	"net/http"
	"strings"
	"time"
)

// RetryableError wraps an error with retry metadata
type RetryableError struct {
	Err       error
	Retryable bool
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// RetryClient wraps http.Client with exponential backoff retry logic
type RetryClient struct {
	client     *http.Client
	maxRetries int
	baseDelay  time.Duration
	logger     *Logger
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	Logger     *Logger
}

// DefaultRetryConfig returns sensible defaults for retry behavior
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		Logger:     NewLogger("retry"),
	}
}

// NewRetryClient creates a new RetryClient wrapping the provided http.Client
func NewRetryClient(client *http.Client, config *RetryConfig) *RetryClient {
	if config == nil {
		config = DefaultRetryConfig()
	}
	if config.Logger == nil {
		config.Logger = NewLogger("retry")
	}
	return &RetryClient{
		client:     client,
		maxRetries: config.MaxRetries,
		baseDelay:  config.BaseDelay,
		logger:     config.Logger,
	}
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Network errors are always retryable
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// Context deadline exceeded (timeout) is retryable
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Context canceled is NOT retryable (user cancelled)
	if errors.Is(err, context.Canceled) {
		return false
	}

	return false
}

// isRetryableStatusCode determines if an HTTP status code should trigger a retry
func isRetryableStatusCode(statusCode int) bool {
	// 5xx server errors are retryable
	if statusCode >= 500 && statusCode < 600 {
		return true
	}

	// 429 Too Many Requests is retryable (rate limited)
	if statusCode == http.StatusTooManyRequests {
		return true
	}

	// 4xx client errors are NOT retryable (except 429 handled above)
	if statusCode >= 400 && statusCode < 500 {
		return false
	}

	// 3xx redirects, 2xx success - not retryable
	return false
}

// calculateBackoff returns the delay for a given attempt number
// Uses exponential backoff: 1s, 2s, 4s for attempts 0, 1, 2 respectively
func calculateBackoff(attempt int, baseDelay time.Duration) time.Duration {
	delay := float64(baseDelay) * math.Pow(2, float64(attempt))
	return time.Duration(delay)
}

// Do performs an HTTP request with automatic retry on retryable errors
// It uses exponential backoff with delays of 1s, 2s, 4s for attempts 1, 2, 3
func (r *RetryClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return r.DoWithRetry(ctx, req)
}

// DoWithRetry performs an HTTP request with automatic retry on retryable errors
// It uses exponential backoff with delays of 1s, 2s, 4s for attempts 1, 2, 3
func (r *RetryClient) DoWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error
	correlationID := ""

	// Extract correlation ID for logging if available
	if id := req.Header.Get("X-Correlation-ID"); id != "" {
		correlationID = id
	}

	// Clone the request body so we can retry with the same body
	// We need to read and restore the body for each attempt
	var originalBody []byte
	if req.Body != nil {
		originalBody, _ = io.ReadAll(req.Body)
		req.Body.Close()
	}

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		// Restore body for this attempt
		if originalBody != nil {
			req.Body = io.NopCloser(strings.NewReader(string(originalBody)))
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader(string(originalBody))), nil
			}
			req.ContentLength = int64(len(originalBody))
		}

		// Create a context with timeout for this attempt
		attemptCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		req = req.WithContext(attemptCtx)

		// Log retry attempt
		if attempt > 0 {
			delay := calculateBackoff(attempt-1, r.baseDelay)
			r.logger.Info("retry_attempt",
				WithCorrelationID(correlationID),
				WithExtra("retry_attempt", attempt),
				WithExtra("max_retries", r.maxRetries),
				WithExtra("delay_ms", delay.Milliseconds()),
				WithExtra("will_retry", attempt < r.maxRetries),
				WithExtra("url", req.URL.String()),
			)

			// Wait before retrying with exponential backoff
			select {
			case <-attemptCtx.Done():
				cancel()
				return nil, context.DeadlineExceeded
			case <-time.After(delay):
			}
		}

		// Execute request
		start := time.Now()
		resp, err := r.client.Do(req)
		cancel()

		latency := time.Since(start)

		// Check for context cancellation
		if err != nil && attempt < r.maxRetries {
			if ctx.Err() != nil {
				// Original context was cancelled, don't retry
				return nil, ctx.Err()
			}
		}

		// Handle response
		if err != nil {
			lastErr = err

			// Log the error
			r.logger.Warn("request_error",
				WithCorrelationID(correlationID),
				WithExtra("attempt", attempt+1),
				WithExtra("error", err.Error()),
				WithExtra("is_retryable", isRetryableError(err)),
				WithExtra("latency_ms", latency.Milliseconds()),
			)

			// Check if we should retry
			if isRetryableError(err) && attempt < r.maxRetries {
				continue
			}

			// Not retryable or max retries reached
			return nil, err
		}

		// Check status code
		if isRetryableStatusCode(resp.StatusCode) {
			lastErr = &RetryableError{
				Err:       errors.New("retryable status code: " + http.StatusText(resp.StatusCode)),
				Retryable: true,
			}

			r.logger.Warn("retryable_status",
				WithCorrelationID(correlationID),
				WithExtra("attempt", attempt+1),
				WithExtra("status_code", resp.StatusCode),
				WithExtra("will_retry", attempt < r.maxRetries),
				WithLatency(latency),
			)

			// Close body before retry
			resp.Body.Close()

			// Retry if we haven't exhausted retries
			if attempt < r.maxRetries {
				continue
			}
		}

		// Log success
		if attempt > 0 {
			r.logger.Info("retry_success",
				WithCorrelationID(correlationID),
				WithExtra("attempt", attempt+1),
				WithExtra("status_code", resp.StatusCode),
				WithLatency(latency),
			)
		}

		return resp, nil
	}

	// Max retries exhausted
	r.logger.Error("max_retries_exhausted",
		WithCorrelationID(correlationID),
		WithExtra("max_retries", r.maxRetries),
		WithError(lastErr),
	)

	return nil, &RetryableError{
		Err:       errors.New("max retries exceeded: " + lastErr.Error()),
		Retryable: false,
	}
}
