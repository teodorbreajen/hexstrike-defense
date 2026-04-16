package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Context key for storing correlation ID
type contextKey string

const correlationIDKey contextKey = "correlation_id"

// ProxyConfig holds all proxy configuration
type ProxyConfig struct {
	ListenAddr         string
	MCPBackendURL      string
	RateLimitPerMinute int
	Timeout            time.Duration
	AllowedOrigins     []string
	FailMode           string // "closed" or "open" - fail-closed blocks on Lakera errors
	MaxBodySize        int64  // Maximum request body size in bytes (default: 1MB)
	JWTSecret          string // JWT secret for auth validation
}

// Middleware defines the proxy middleware function signature
type Middleware func(http.Handler) http.Handler

// Proxy holds the proxy server state
type Proxy struct {
	config          *ProxyConfig
	lakeraClient    *LakeraClient
	rateLimiter     *RateLimiter
	metrics         *Metrics
	middlewareChain []Middleware
	logger          *Logger
}

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	mu         sync.Mutex
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(perMinute int) *RateLimiter {
	return &RateLimiter{
		tokens:     perMinute,
		maxTokens:  perMinute,
		refillRate: time.Minute,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed under rate limits
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastRefill)

	// Refill tokens based on elapsed time
	if elapsed >= r.refillRate {
		r.tokens = r.maxTokens
		r.lastRefill = now
	}

	if r.tokens > 0 {
		r.tokens--
		return true
	}

	return false
}

// Metrics holds proxy metrics
type Metrics struct {
	mu               sync.RWMutex
	TotalRequests    int64
	BlockedRequests  int64
	AllowedRequests  int64
	TotalLatency     time.Duration
	LakeraBlockCount int64
	StatusCodes      map[int]int64
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		StatusCodes: make(map[int]int64),
	}
}

// RecordRequest records a request in metrics
func (m *Metrics) RecordRequest(allowed bool, latency time.Duration, statusCode int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	m.TotalLatency += latency

	if !allowed {
		m.BlockedRequests++
		m.LakeraBlockCount++
	} else {
		m.AllowedRequests++
	}

	m.StatusCodes[statusCode]++
}

// GetMetrics returns current metrics snapshot
func (m *Metrics) GetMetrics() (total, blocked, allowed int64, avgLatency float64, statusCodes map[int]int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total = m.TotalRequests
	blocked = m.BlockedRequests
	allowed = m.AllowedRequests
	if total > 0 {
		avgLatency = float64(m.TotalLatency) / float64(total) / float64(time.Millisecond)
	}

	// Copy status codes
	statusCodes = make(map[int]int64, len(m.StatusCodes))
	for k, v := range m.StatusCodes {
		statusCodes[k] = v
	}

	return
}

// NewProxy creates a new proxy instance
func NewProxy(config *ProxyConfig, lakeraClient *LakeraClient) *Proxy {
	proxy := &Proxy{
		config:       config,
		lakeraClient: lakeraClient,
		rateLimiter:  NewRateLimiter(config.RateLimitPerMinute),
		metrics:      NewMetrics(),
		logger:       NewLogger("proxy"),
	}

	// Build middleware chain
	proxy.middlewareChain = []Middleware{
		proxy.loggingMiddleware,
		proxy.rateLimitMiddleware,
		proxy.authMiddleware,
		proxy.semanticCheckMiddleware,
	}

	return proxy
}

// Middleware chain execution
func (p *Proxy) Handler() http.Handler {
	var handler http.Handler = http.HandlerFunc(p.forwardToMCP)

	// Apply middleware in reverse order (so first in list is outermost)
	for i := len(p.middlewareChain) - 1; i >= 0; i-- {
		handler = p.middlewareChain[i](handler)
	}

	return handler
}

// getCorrelationID extracts correlation ID from context
func getCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return ""
}

// loggingMiddleware logs incoming requests with correlation ID
func (p *Proxy) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate correlation ID for this request
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = GenerateCorrelationID()
		}

		// Store in request context
		ctx := context.WithValue(r.Context(), correlationIDKey, correlationID)
		r = r.WithContext(ctx)

		// Add correlation ID to response headers
		w.Header().Set("X-Correlation-ID", correlationID)

		wrapped := &statusWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapped, r)

		p.logger.Info("request completed",
			WithCorrelationID(correlationID),
			WithMethod(r.Method),
			WithPath(r.URL.Path),
			WithStatusCode(wrapped.statusCode),
			WithLatency(time.Since(start)),
		)
	})
}

// statusWriter wraps http.ResponseWriter to capture status code
type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.statusCode = code
	sw.ResponseWriter.WriteHeader(code)
}

// rateLimitMiddleware enforces rate limiting
func (p *Proxy) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !p.rateLimiter.Allow() {
			p.metrics.RecordRequest(false, 0, http.StatusTooManyRequests)
			p.logger.Warn("rate limit exceeded",
				WithCorrelationID(getCorrelationID(r.Context())),
				WithMethod(r.Method),
				WithPath(r.URL.Path),
				WithStatusCode(http.StatusTooManyRequests),
			)
			p.sendErrorResponse(w, r, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authMiddleware validates JWT Bearer token authentication
func (p *Proxy) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Unprotected endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/metrics" || r.URL.Path == "/ready" {
			next.ServeHTTP(w, r)
			return
		}

		correlationID := getCorrelationID(r.Context())

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			p.logger.Warn("missing authorization header",
				WithCorrelationID(correlationID),
				WithPath(r.URL.Path),
				WithStatusCode(http.StatusUnauthorized),
			)
			p.sendErrorResponse(w, r, http.StatusUnauthorized, "Missing Authorization header")
			return
		}

		// Validate Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			p.logger.Warn("invalid authorization format",
				WithCorrelationID(correlationID),
				WithPath(r.URL.Path),
				WithStatusCode(http.StatusUnauthorized),
			)
			p.sendErrorResponse(w, r, http.StatusUnauthorized, "Invalid Authorization format - expected Bearer token")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate JWT
		if err := p.validateJWT(tokenString); err != nil {
			p.logger.Warn("jwt validation failed",
				WithCorrelationID(correlationID),
				WithPath(r.URL.Path),
				WithStatusCode(http.StatusUnauthorized),
				WithError(err),
			)
			p.sendErrorResponse(w, r, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		p.logger.Debug("jwt validated",
			WithCorrelationID(correlationID),
			WithPath(r.URL.Path),
		)
		next.ServeHTTP(w, r)
	})
}

// validateJWT validates a JWT token and returns an error if invalid
func (p *Proxy) validateJWT(tokenString string) error {
	// If no JWT secret configured, skip validation (backward compatibility)
	if p.config.JWTSecret == "" {
		return nil
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(p.config.JWTSecret), nil
	})

	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

// semanticCheckMiddleware validates tool calls using Lakera
func (p *Proxy) semanticCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := getCorrelationID(r.Context())

		// Only check JSON-RPC requests (POST with JSON content)
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		// Check content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "" && contentType != "application/json" {
			next.ServeHTTP(w, r)
			return
		}

		// Check body size limit
		if r.ContentLength > p.config.MaxBodySize {
			p.logger.Warn("request body size exceeds limit",
				WithCorrelationID(correlationID),
				WithExtra("content_length", r.ContentLength),
				WithExtra("max_size", p.config.MaxBodySize),
				WithStatusCode(http.StatusRequestEntityTooLarge),
			)
			p.metrics.RecordRequest(false, 0, http.StatusRequestEntityTooLarge)
			p.sendErrorResponse(w, r, http.StatusRequestEntityTooLarge,
				fmt.Sprintf("Request body exceeds maximum size of %d bytes", p.config.MaxBodySize))
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			p.logger.Error("failed to read request body",
				WithCorrelationID(correlationID),
				WithError(err),
			)
			p.sendErrorResponse(w, r, http.StatusBadRequest, "Failed to read request body")
			return
		}

		// Parse JSON-RPC request
		parsed, err := ParseJSONRPC(body)
		if err != nil {
			// If it's not a valid JSON-RPC, forward anyway (might be raw MCP)
			p.logger.Debug("failed to parse JSON-RPC, forwarding anyway",
				WithCorrelationID(correlationID),
				WithError(err),
			)
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

		// If it's a batch request, check each request
		if parsed.IsBatch {
			for _, batchReq := range parsed.BatchReqs {
				if allowed, reason := p.checkBatchRequest(&batchReq); !allowed {
					p.logger.Warn("batch request blocked",
						WithCorrelationID(correlationID),
						WithExtra("reason", reason),
						WithStatusCode(http.StatusForbidden),
					)
					p.metrics.RecordRequest(false, 0, http.StatusForbidden)
					p.sendErrorResponse(w, r, http.StatusForbidden,
						fmt.Sprintf("Blocked: %s", reason))
					return
				}
			}
			// All batch items allowed, forward the original request
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

		// Get tool info for semantic check
		toolName, args, ok := GetToolInfo(parsed)
		if !ok {
			// Not a tool call, forward without check
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

		// Check with Lakera
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		allowed, score, reason, err := p.lakeraClient.CheckToolCall(ctx, toolName, args)
		if err != nil {
			p.logger.Error("lakera check error",
				WithCorrelationID(correlationID),
				WithExtra("tool", toolName),
				WithError(err),
			)
			// Fail-closed: block request if Lakera fails
			if p.config.FailMode == "closed" {
				p.logger.Warn("fail-closed mode: blocking request due to Lakera error",
					WithCorrelationID(correlationID),
					WithExtra("tool", toolName),
					WithStatusCode(http.StatusServiceUnavailable),
				)
				p.metrics.RecordRequest(false, 0, http.StatusServiceUnavailable)
				p.sendErrorResponse(w, r, http.StatusServiceUnavailable,
					"Semantic check unavailable - request blocked for security")
				return
			}
			// Fail-open: allow request (backward compatible)
			p.logger.Warn("fail-open mode: allowing request despite Lakera error",
				WithCorrelationID(correlationID),
				WithExtra("tool", toolName),
			)
		}

		if !allowed {
			p.logger.Warn("tool blocked by semantic firewall",
				WithCorrelationID(correlationID),
				WithExtra("tool", toolName),
				WithExtra("score", score),
				WithExtra("reason", reason),
				WithStatusCode(http.StatusForbidden),
			)
			p.metrics.RecordRequest(false, 0, http.StatusForbidden)
			p.sendErrorResponse(w, r, http.StatusForbidden,
				fmt.Sprintf("Tool '%s' blocked by semantic firewall: %s", toolName, reason))
			return
		}

		p.logger.Debug("tool call allowed by semantic check",
			WithCorrelationID(correlationID),
			WithExtra("tool", toolName),
		)

		// Request allowed, forward to MCP backend
		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}

// checkBatchRequest checks a single request in a batch
func (p *Proxy) checkBatchRequest(req *ParsedRequest) (bool, string) {
	toolName, args, ok := GetToolInfo(req)
	if !ok {
		return true, "" // Not a tool call
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	allowed, score, reason, _ := p.lakeraClient.CheckToolCall(ctx, toolName, args)
	if !allowed {
		return false, fmt.Sprintf("tool '%s' (score: %d): %s", toolName, score, reason)
	}

	return true, ""
}

// forwardToMCP forwards the request to the MCP backend
func (p *Proxy) forwardToMCP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	correlationID := getCorrelationID(r.Context())

	// Create backend URL
	url := p.config.MCPBackendURL + r.URL.Path

	// Create proxy request with context containing correlation ID
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, url, r.Body)
	if err != nil {
		p.logger.Error("failed to create proxy request",
			WithCorrelationID(correlationID),
			WithError(err),
		)
		p.sendErrorResponse(w, r, http.StatusInternalServerError, "Failed to create proxy request")
		return
	}

	// Copy headers (except Host)
	for k, v := range r.Header {
		if k != "Host" {
			proxyReq.Header[k] = v
		}
	}

	// Forward correlation ID to MCP backend if not already present
	if proxyReq.Header.Get("X-Correlation-ID") == "" {
		proxyReq.Header.Set("X-Correlation-ID", correlationID)
	}

	// Use the same transport with timeout
	client := &http.Client{
		Timeout: p.config.Timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   p.config.Timeout,
				KeepAlive: p.config.Timeout,
			}).DialContext,
		},
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		p.logger.Error("backend error",
			WithCorrelationID(correlationID),
			WithExtra("backend_url", p.config.MCPBackendURL),
			WithError(err),
		)
		p.sendErrorResponse(w, r, http.StatusBadGateway, "MCP backend unavailable")
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	// Copy response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.Error("failed to read response body",
			WithCorrelationID(correlationID),
			WithError(err),
		)
		p.sendErrorResponse(w, r, http.StatusInternalServerError, "Failed to read response")
		return
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)

	// Record metrics
	latency := time.Since(start)
	p.metrics.RecordRequest(true, latency, resp.StatusCode)

	p.logger.Debug("response forwarded from backend",
		WithCorrelationID(correlationID),
		WithStatusCode(resp.StatusCode),
		WithLatency(latency),
	)
}

// sendErrorResponse sends a JSON-RPC error response
func (p *Proxy) sendErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	correlationID := getCorrelationID(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)

	// If original request was a JSON-RPC request, return JSON-RPC error
	// Otherwise return plain HTTP error
	resp := CreateErrorResponse(nil, statusCode, message)
	body, _ := SerializeResponse(resp)

	w.WriteHeader(statusCode)
	w.Write(body)
}

// GetMetricsHandler returns the metrics as JSON
func (p *Proxy) GetMetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		correlationID := getCorrelationID(r.Context())

		total, blocked, allowed, avgLatency, statusCodes := p.metrics.GetMetrics()

		metrics := map[string]interface{}{
			"total_requests":   total,
			"blocked_requests": blocked,
			"allowed_requests": allowed,
			"avg_latency_ms":   avgLatency,
			"lakera_blocks":    blocked,
			"status_codes":     statusCodes,
		}

		p.logger.Debug("metrics requested",
			WithCorrelationID(correlationID),
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	}
}
