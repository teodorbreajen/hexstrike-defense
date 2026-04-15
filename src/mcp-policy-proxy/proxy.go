package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// ProxyConfig holds all proxy configuration
type ProxyConfig struct {
	ListenAddr         string
	MCPBackendURL      string
	RateLimitPerMinute int
	Timeout            time.Duration
	AllowedOrigins     []string
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
	}

	// Build middleware chain
	proxy.middlewareChain = []Middleware{
		loggingMiddleware,
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

// loggingMiddleware logs incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &statusWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapped, r)

		log.Printf("[HTTP] %s %s - %d - %v",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			time.Since(start))
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
			p.sendErrorResponse(w, r, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authMiddleware validates authentication (placeholder - implement based on needs)
func (p *Proxy) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For now, allow all requests
		// In production, implement JWT/API key validation
		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" && r.URL.Path != "/health" && r.URL.Path != "/metrics" {
			// Allow health/metrics without auth, require for others
			// log.Printf("[Auth] No authorization header for %s", r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}

// semanticCheckMiddleware validates tool calls using Lakera
func (p *Proxy) semanticCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			p.sendErrorResponse(w, r, http.StatusBadRequest, "Failed to read request body")
			return
		}

		// Parse JSON-RPC request
		parsed, err := ParseJSONRPC(body)
		if err != nil {
			// If it's not a valid JSON-RPC, forward anyway (might be raw MCP)
			log.Printf("[Semantic] Failed to parse JSON-RPC: %v - forwarding anyway", err)
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

		// If it's a batch request, check each request
		if parsed.IsBatch {
			for _, batchReq := range parsed.BatchReqs {
				if allowed, reason := p.checkBatchRequest(&batchReq); !allowed {
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
			log.Printf("[Semantic] Lakera check error: %v", err)
			// On error, allow (graceful degradation)
		}

		if !allowed {
			log.Printf("[Semantic] Blocked tool '%s' - score: %d, reason: %s",
				toolName, score, reason)
			p.metrics.RecordRequest(false, 0, http.StatusForbidden)
			p.sendErrorResponse(w, r, http.StatusForbidden,
				fmt.Sprintf("Tool '%s' blocked by semantic firewall: %s", toolName, reason))
			return
		}

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

	// Create backend URL
	url := p.config.MCPBackendURL + r.URL.Path

	// Create proxy request
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, url, r.Body)
	if err != nil {
		p.sendErrorResponse(w, r, http.StatusInternalServerError, "Failed to create proxy request")
		return
	}

	// Copy headers (except Host)
	for k, v := range r.Header {
		if k != "Host" {
			proxyReq.Header[k] = v
		}
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
		log.Printf("[Proxy] Backend error: %v", err)
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
		log.Printf("[Proxy] Failed to read response body: %v", err)
		p.sendErrorResponse(w, r, http.StatusInternalServerError, "Failed to read response")
		return
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)

	// Record metrics
	latency := time.Since(start)
	p.metrics.RecordRequest(true, latency, resp.StatusCode)
}

// sendErrorResponse sends a JSON-RPC error response
func (p *Proxy) sendErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")

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
		total, blocked, allowed, avgLatency, statusCodes := p.metrics.GetMetrics()

		metrics := map[string]interface{}{
			"total_requests":   total,
			"blocked_requests": blocked,
			"allowed_requests": allowed,
			"avg_latency_ms":   avgLatency,
			"lakera_blocks":    blocked,
			"status_codes":     statusCodes,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	}
}
