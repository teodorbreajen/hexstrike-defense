package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/hexstrike/mcp-policy-proxy/dlq"
)

// Security constants
const (
	MaxToolNameLength  = 256     // Maximum tool name length to prevent overflow
	MaxArgumentsLength = 65536   // Maximum arguments size (64KB)
	MaxPathDepth       = 10      // Maximum URL path depth
	MaxRequestBodySize = 1048576 // 1MB max body (DoS protection)
	MaxConcurrentReqs  = 100     // Max concurrent requests per client
	MaxBatchSize       = 10      // Max batch requests to prevent batch attacks
)

// isInternalURL checks if a URL points to internal services (SSRF protection)
// SECURITY: Comprehensive SSRF protection with multiple detection layers
func isInternalURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return true // Block unparseable URLs
	}

	host := strings.ToLower(parsed.Hostname())

	// SECURITY FIX: Check localhost variants including IPv6
	localhostVariants := []string{
		"localhost",
		"127.0.0.1",
		"::1",
		"::ffff:127.0.0.1", // IPv4-mapped IPv6
		"0.0.0.0",
		"::", // All IPv6 interfaces
	}
	for _, variant := range localhostVariants {
		if host == variant {
			return true
		}
	}

	// SECURITY FIX: Parse and check IP directly for accurate CIDR validation
	if ip := net.ParseIP(host); ip != nil {
		// Check IPv4 private ranges (RFC 1918)
		if ip.IsLoopback() {
			return true
		}
		if ip.IsPrivate() {
			return true
		}
		if ip.IsLinkLocalUnicast() {
			return true
		}
		if ip.IsUnspecified() {
			return true
		}
		// Check IPv6 private ranges (fc00::/7, fe80::/10)
		if ip.IsLoopback() || isIPv6Private(ip) {
			return true
		}
	}

	// SECURITY FIX: Check private IP ranges - COMPLETE RFC 1918 ranges
	// 10.0.0.0/8 - 10.x.x.x
	if strings.HasPrefix(host, "10.") {
		return true
	}
	// 172.16.0.0/12 - 172.16.x.x through 172.31.x.x (including 172.15.x.x!)
	if strings.HasPrefix(host, "172.") {
		// Extract second octet to validate 16-31 range
		if len(host) >= 5 {
			secondOctet := host[4:]
			for _, prefix := range []string{
				"16.", "17.", "18.", "19.", "20.", "21.", "22.", "23.",
				"24.", "25.", "26.", "27.", "28.", "29.", "30.", "31.",
				"15.", // SECURITY FIX: Include 172.15.x.x (was missing!)
			} {
				if strings.HasPrefix(secondOctet, prefix) {
					return true
				}
			}
		}
	}
	// 192.168.0.0/16
	if strings.HasPrefix(host, "192.168.") {
		return true
	}
	// 169.254.0.0/16 (link-local)
	if strings.HasPrefix(host, "169.254.") {
		return true
	}

	// SECURITY FIX: Check for cloud metadata service IPs
	metadataIPs := []string{
		"169.254.169.254",          // AWS, Azure, GCP, DigitalOcean
		"169.254.169.253",          // Oracle Cloud
		"100.100.100.200",          // Alibaba Cloud
		"metadata.google.internal", // GCP hostname
	}
	for _, metaIP := range metadataIPs {
		if host == strings.ToLower(metaIP) {
			return true
		}
	}

	// SECURITY FIX: Check for kubernetes internal service
	lowerHost := strings.ToLower(host)
	kubernetesIdentifiers := []string{
		"kubernetes",
		"kubernetes.default",
		"kubernetes.default.svc",
		"kubernetes.default.svc.cluster.local",
	}
	for _, identifier := range kubernetesIdentifiers {
		if strings.Contains(lowerHost, identifier) {
			return true
		}
	}

	return false
}

// isIPv6Private checks if an IPv6 IP is in private range (fc00::/7)
func isIPv6Private(ip net.IP) bool {
	// fc00::/7 = 1111110x in binary
	// x=1 for local scope (fc00::), x=0 for global scope (fd00::)
	// fe80::/10 = link-local
	ip6 := ip.To16()
	if ip6 == nil {
		return false
	}
	firstByte := ip6[0]
	// fc00::/7 starts with 1111110
	if firstByte&0xfe == 0xfc {
		return true
	}
	// fe80::/10 starts with 1111111010
	if firstByte&0xff == 0xfe {
		return true
	}
	return false
}

// sanitizeToolInput sanitizes tool name and arguments to prevent injection
// SECURITY: Multiple layers of validation
func sanitizeToolInput(toolName, args string) (string, string, error) {
	// LAYER 1: Length limits
	if len(toolName) > MaxToolNameLength {
		toolName = toolName[:MaxToolNameLength]
	}

	// LAYER 2: Null byte detection
	if strings.Contains(toolName, "\x00") || strings.Contains(args, "\x00") {
		return "", "", fmt.Errorf("null byte detected")
	}

	// LAYER 3: Empty/whitespace only validation
	if strings.TrimSpace(toolName) == "" {
		return "", "", fmt.Errorf("tool name empty or whitespace only")
	}

	// LAYER 4: Character allowlist for tool name
	for _, c := range toolName {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '_' || c == '-' || c == '/' || c == '.') {
			return "", "", fmt.Errorf("invalid character '%c' in tool name", c)
		}
	}

	// LAYER 5: Arguments size limit
	if len(args) > MaxArgumentsLength {
		args = args[:MaxArgumentsLength]
	}

	// LAYER 6: Path traversal detection (arguments)
	lowerArgs := strings.ToLower(args)
	dangerousPatterns := []string{
		"../", "..\\", "%2e%2e", "%252e%252e", "..%5c", "%252e%255c",
		"&&", "||", ";", "`", "$(", "${", "|", ">", "<",
	}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerArgs, pattern) {
			return "", "", fmt.Errorf("suspicious pattern '%s' in arguments", pattern)
		}
	}

	// LAYER 6b: SQL injection detection (arguments)
	sqlPatterns := []string{
		"' or ", "' or'", "or 1=1", "or true", "or '1'='1",
		"union select", "union all select", "union from",
		"drop table", "drop table ", "delete from", "insert into",
		"update ", "set ", "exec(", "execute(", "xp_",
		"--", "/*", "*/", "@@", "@variable",
	}
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerArgs, pattern) {
			return "", "", fmt.Errorf("sql injection pattern '%s' in arguments", pattern)
		}
	}

	// LAYER 7: Command injection patterns
	commandPatterns := []string{
		"rm -rf", "del /f", "format c:", "dd if=", ":(){:|:&};:",
		"wget ", "curl ", "nc -", "ncat ",
	}
	for _, pattern := range commandPatterns {
		if strings.Contains(strings.ToLower(args), pattern) {
			return "", "", fmt.Errorf("dangerous command pattern in arguments")
		}
	}

	return toolName, args, nil
}

// validateBackendURL validates the backend URL to prevent SSRF attacks
func validateBackendURL(backendURL, path string) (string, error) {
	// Parse and validate the full URL
	fullURL := backendURL + path

	parsed, err := url.Parse(fullURL)
	if err != nil {
		return "", fmt.Errorf("invalid backend URL: %w", err)
	}

	// Block internal URLs (SSRF protection)
	if isInternalURL(backendURL) {
		return "", fmt.Errorf("internal URLs not allowed")
	}

	// Only allow http/https schemes
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("only http/https schemes allowed")
	}

	// Validate path depth to prevent path traversal
	pathParts := strings.Split(parsed.Path, "/")
	if len(pathParts) > MaxPathDepth {
		return "", fmt.Errorf("path too deep (max %d)", MaxPathDepth)
	}

	// Check for suspicious path patterns - use lowercase for comparison
	lowerPath := strings.ToLower(parsed.Path)
	suspiciousPatterns := []string{"../", "..\\", "%2e%2e", "%252e%252e"}
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerPath, pattern) {
			return "", fmt.Errorf("suspicious path pattern detected")
		}
	}

	// Check for null bytes in path
	if strings.Contains(parsed.Path, "\x00") {
		return "", fmt.Errorf("invalid path: contains null byte")
	}

	return fullURL, nil
}

// Context key for storing correlation ID
type contextKey string

const correlationIDKey contextKey = "correlation_id"

// ProxyConfig holds all proxy configuration
type ProxyConfig struct {
	ListenAddr           string
	MCPBackendURL        string
	RateLimitPerMinute   int
	Timeout              time.Duration
	AllowedOrigins       []string
	FailMode             string // "closed" or "open" - fail-closed blocks on Lakera errors
	MaxBodySize          int64  // Maximum request body size in bytes (default: 1MB)
	JWTSecret            string // JWT secret for auth validation
	TrustedProxies       string // Comma-separated list of trusted proxy IPs for X-Forwarded-For validation
	CORSAllowedOrigins   []string
	CORSAllowCredentials bool
	DLQPath              string // Path for DLQ storage
	DLQTTLHours          int    // TTL for DLQ messages in hours
}

// Middleware defines the proxy middleware function signature
type Middleware func(http.Handler) http.Handler

// Proxy holds the proxy server state
type Proxy struct {
	config          *ProxyConfig
	lakeraClient    LakeraChecker      // SECURITY: Uses interface for testability - can be nil
	rateLimiter     *RateLimiter       // Legacy (DEPRECATED, kept for backward compat)
	clientRL        *ClientRateLimiter // SECURITY: Per-client rate limiter ACTIVE
	metrics         *Metrics
	middlewareChain []Middleware
	logger          *Logger
	circuitBreaker  *CircuitBreaker // SECURITY: Circuit breaker for backend resilience
	semaphore       chan struct{}   // FIX #2: For concurrent request limiting
	backendClient   *http.Client    // FIX #4: Reusable client with connection pooling
	retryClient     *RetryClient    // Retry client with exponential backoff
	lakeraBreaker   *CircuitBreaker // Circuit breaker specifically for Lakera
	corsMiddleware  *CORSMiddleware // CORS middleware for cross-origin requests
	dlq             *dlq.DLQ        // Dead Letter Queue for failed requests
	dlqStop         func()          // Stop function for DLQ cleanup goroutine
}

// ClientRateLimiter implements per-client rate limiting to prevent DoS attacks
// Each client gets their own token bucket, preventing one attacker from blocking all requests
type ClientRateLimiter struct {
	mu          sync.Mutex
	clients     map[string]*clientBucket
	maxTokens   int
	refillRate  time.Duration
	cleanupTTL  time.Duration // Cleanup inactive clients after this duration
	lastCleanup time.Time
	maxClients  int  // Maximum number of distinct clients to track
	enabled     bool // Whether per-client rate limiting is active
}

// clientBucket represents rate limit state for a single client
type clientBucket struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
}

// NewClientRateLimiter creates a new per-client rate limiter
func NewClientRateLimiter(perMinute int) *ClientRateLimiter {
	return &ClientRateLimiter{
		clients:     make(map[string]*clientBucket),
		maxTokens:   perMinute,
		refillRate:  time.Minute,
		cleanupTTL:  5 * time.Minute, // Cleanup clients inactive for 5 minutes
		lastCleanup: time.Now(),
		maxClients:  10000, // Limit to prevent memory exhaustion
		enabled:     true,  // Per-client rate limiting ACTIVE
	}
}

// getClientID (DEPRECATED - kept for backward compatibility)
// SECURITY: Use getClientIP() + maskIP() instead for rate limiting to avoid logging sensitive data
// Note: This function is no longer used in rate limit middleware
func getClientID(r *http.Request) string {
	// DEPRECATED: We now use IP-based identification only
	// to avoid exposing JWT tokens in logs
	_ = r // Suppress unused warning
	return "deprecated"
}

// isTrustedProxy checks if the remote addr is a trusted proxy
// SECURITY: Prevents X-Forwarded-For spoofing by only trusting headers from known proxies
// SECURITY FIX: Added CIDR support for trusted proxy ranges
func isTrustedProxy(remoteAddr string, trustedProxies string) bool {
	if trustedProxies == "" {
		return false // No trusted proxies configured
	}

	// Parse remoteAddr to get IP
	remoteIP, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		remoteIP = remoteAddr
	}

	// Parse the remote IP
	remoteAddrParsed := net.ParseIP(remoteIP)
	if remoteAddrParsed == nil {
		return false // Invalid IP
	}

	// Check if remote IP is in trusted list (supports both IPs and CIDR ranges)
	trustedList := strings.Split(trustedProxies, ",")
	for _, trusted := range trustedList {
		trusted = strings.TrimSpace(trusted)
		if trusted == "" {
			continue
		}

		// Check if it contains "/" (CIDR notation)
		if strings.Contains(trusted, "/") {
			// Parse CIDR range
			_, ipNet, err := net.ParseCIDR(trusted)
			if err == nil {
				// Check if remote IP is in the CIDR range
				if ipNet.Contains(remoteAddrParsed) {
					return true
				}
			}
		} else {
			// Exact IP match
			if remoteIP == trusted {
				return true
			}
			// Also try parsing as IP to handle different formats
			if trustedParsed := net.ParseIP(trusted); trustedParsed != nil {
				if remoteAddrParsed.Equal(trustedParsed) {
					return true
				}
			}
		}
	}

	return false
}

// getClientIP extracts the real client IP, considering proxies
// SECURITY: Only trusts X-Forwarded-For if request comes from trusted proxy
// X-Real-IP is always trusted as it's typically set by the reverse proxy directly
func getClientIP(r *http.Request, trustedProxies string) string {
	// Check X-Real-IP first (typically set by reverse proxy, safer than X-Forwarded-For)
	xri := r.Header.Get("X-Real-IP")
	if xri != "" && isValidIP(xri) {
		return strings.TrimSpace(xri)
	}

	// SECURITY: Only trust X-Forwarded-For if request comes from trusted proxy
	isTrusted := isTrustedProxy(r.RemoteAddr, trustedProxies)

	if isTrusted {
		// Check X-Forwarded-For (common in Kubernetes/proxy setups)
		xff := r.Header.Get("X-Forwarded-For")
		if xff != "" {
			// Take the first IP in the chain
			if idx := strings.Index(xff, ","); idx > 0 {
				xff = strings.TrimSpace(xff[:idx])
			}
			// SECURITY: Validate the extracted IP is actually valid format
			if isValidIP(xff) {
				return xff
			}
		}
	}

	// Fall back to RemoteAddr (most reliable - always trusted)
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return r.RemoteAddr
}

// isValidIP validates that a string is a valid IP address
// SECURITY: Prevents IP spoofing via malformed headers
// Uses only net.ParseIP to prevent malformed IPs like 999.999.999.999 from being accepted
func isValidIP(ipStr string) bool {
	return net.ParseIP(ipStr) != nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Allow checks if a request is allowed under rate limits for this specific client
func (r *ClientRateLimiter) Allow(clientID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Periodic cleanup of stale clients
	if now.Sub(r.lastCleanup) > r.cleanupTTL {
		r.cleanupStaleClients(now)
	}

	// Get or create client bucket
	bucket, exists := r.clients[clientID]
	if !exists {
		// Check if we've hit max clients
		if len(r.clients) >= r.maxClients {
			// Try to cleanup first
			r.cleanupStaleClients(now)
			if len(r.clients) >= r.maxClients {
				// Force cleanup oldest if still at capacity
				r.forceCleanupOldest(now)
			}
		}

		// Create new bucket for this client
		bucket = &clientBucket{
			tokens:     r.maxTokens - 1, // Reserve one for this request
			maxTokens:  r.maxTokens,
			refillRate: r.refillRate,
			lastRefill: now,
		}
		r.clients[clientID] = bucket
		return true
	}

	// SECURITY FIX: Token bucket refill - properly refill tokens based on elapsed time
	// Previous bug: Reset all tokens, now: refill proportionally
	elapsed := now.Sub(bucket.lastRefill)
	if elapsed >= bucket.refillRate {
		// Calculate how many full refill periods have passed
		periods := int(elapsed / bucket.refillRate)
		// Add tokens for each period (capped at max)
		newTokens := min(bucket.maxTokens, bucket.tokens+periods)
		if newTokens > bucket.tokens {
			bucket.tokens = newTokens
		}
		bucket.lastRefill = bucket.lastRefill.Add(time.Duration(periods) * bucket.refillRate)
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// cleanupStaleClients removes clients that have been inactive
func (r *ClientRateLimiter) cleanupStaleClients(now time.Time) {
	r.lastCleanup = now

	for id, bucket := range r.clients {
		if now.Sub(bucket.lastRefill) > r.cleanupTTL {
			delete(r.clients, id)
		}
	}
}

// forceCleanupOldest removes the oldest client to make room for new ones
func (r *ClientRateLimiter) forceCleanupOldest(now time.Time) {
	var oldestID string
	var oldestTime time.Time

	for id, bucket := range r.clients {
		if oldestTime.IsZero() || bucket.lastRefill.Before(oldestTime) {
			oldestTime = bucket.lastRefill
			oldestID = id
		}
	}

	if oldestID != "" {
		delete(r.clients, oldestID)
	}
}

// GetClientCount returns the number of tracked clients
func (r *ClientRateLimiter) GetClientCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.clients)
}

// DEPRECATED: Use NewClientRateLimiter instead
// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	mu         sync.Mutex
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter (DEPRECATED - use NewClientRateLimiter)
func NewRateLimiter(perMinute int) *RateLimiter {
	return &RateLimiter{
		tokens:     perMinute,
		maxTokens:  perMinute,
		refillRate: time.Minute,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed under rate limits (DEPRECATED)
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

// CircuitBreaker implements a simple circuit breaker pattern for backend protection
type CircuitBreaker struct {
	mu                  sync.Mutex
	state               CircuitState
	failureCount        int
	successCount        int
	lastFailure         time.Time
	threshold           int           // Failures before opening circuit
	timeout             time.Duration // How long circuit stays open
	halfOpenMaxRequests int           // Max requests during half-open state
}

// CircuitState represents the circuit breaker state
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // Normal operation
	CircuitOpen                         // Failing, reject requests
	CircuitHalfOpen                     // Testing if backend recovered
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 5 // Default: open after 5 failures
	}
	if timeout <= 0 {
		timeout = 30 * time.Second // Default: try again after 30 seconds
	}

	return &CircuitBreaker{
		state:               CircuitClosed,
		threshold:           threshold,
		timeout:             timeout,
		halfOpenMaxRequests: 3,
	}
}

// Allow returns true if requests are allowed through the circuit
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true

	case CircuitOpen:
		// Check if timeout has passed to try half-open
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = CircuitHalfOpen
			cb.successCount = 0
			return true
		}
		return false

	case CircuitHalfOpen:
		// Allow limited requests to test recovery
		return cb.successCount < cb.halfOpenMaxRequests
	}

	return false
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.halfOpenMaxRequests {
			// Recovery successful, close the circuit
			cb.state = CircuitClosed
			cb.failureCount = 0
		}

	case CircuitClosed:
		// Reset failure count on success
		cb.failureCount = 0
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailure = time.Now()

	switch cb.state {
	case CircuitHalfOpen:
		// Failure during half-open: go back to open
		cb.state = CircuitOpen

	case CircuitClosed:
		if cb.failureCount >= cb.threshold {
			// Too many failures, open the circuit
			cb.state = CircuitOpen
		}
	}
}

// GetState returns the current circuit state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
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

// NewProxy creates a new proxy instance with security enhancements
func NewProxy(config *ProxyConfig, lakeraClient *LakeraClient) *Proxy {
	proxy := &Proxy{
		config:         config,
		lakeraClient:   lakeraClient,                                    // Can be nil - nil checks handle it
		rateLimiter:    NewRateLimiter(config.RateLimitPerMinute),       // Legacy
		clientRL:       NewClientRateLimiter(config.RateLimitPerMinute), // SECURITY: Per-client ACTIVE
		metrics:        NewMetrics(),
		logger:         NewLogger("proxy"),
		circuitBreaker: NewCircuitBreaker(5, 30*time.Second), // SECURITY: Open after 5 failures, retry after 30s
	}

	// FIX #1: Log warning if JWT auth is disabled
	if config.JWTSecret == "" {
		proxy.logger.Warn("JWT authentication is DISABLED - JWTSecret not configured. ALL requests will be allowed without authentication.")
	}

	// FIX #2: Initialize semaphore for concurrent request limiting
	proxy.semaphore = make(chan struct{}, MaxConcurrentReqs)

	// FIX #4: Initialize backend client with connection pooling
	proxy.backendClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: config.Timeout,
	}

	// Initialize retry client with exponential backoff (1s, 2s, 4s)
	proxy.retryClient = NewRetryClient(proxy.backendClient, &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		Logger:     proxy.logger,
	})

	// Initialize CORS middleware (disabled if no origins configured)
	proxy.corsMiddleware = NewCORSMiddleware(config.CORSAllowedOrigins, config.CORSAllowCredentials)

	// Initialize DLQ (Dead Letter Queue) for failed requests
	if config.DLQPath != "" {
		dlqInstance, err := dlq.NewDLQ(&dlq.DLQConfig{
			Path:     config.DLQPath,
			TTLHours: config.DLQTTLHours,
		})
		if err != nil {
			proxy.logger.Error("failed to initialize DLQ", WithError(err))
		} else {
			dlqInstance.SetLogger(func(format string, args ...interface{}) {
				proxy.logger.Debug(fmt.Sprintf(format, args...))
			})
			proxy.dlq = dlqInstance

			// Start background cleanup goroutine (runs every hour by default)
			proxy.dlqStop = dlq.StartCleanupGoroutine(dlqInstance, &dlq.CleanupConfig{
				Interval: 1 * time.Hour,
				TTLHours: config.DLQTTLHours,
			})
			proxy.logger.Info("DLQ initialized",
				WithExtra("path", config.DLQPath),
				WithExtra("ttl_hours", config.DLQTTLHours),
			)
		}
	}

	// Build middleware chain with security layers
	proxy.middlewareChain = []Middleware{
		proxy.corsMiddleware.Handle,     // CORS: Handle cross-origin requests
		proxy.securityHeadersMiddleware, // SECURITY: Add security headers first
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

// securityHeadersMiddleware adds security headers to all responses
// Implements defense-in-depth by adding standard security headers
func (p *Proxy) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// SECURITY HEADERS:
		// X-Content-Type-Options: Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// X-XSS-Protection: Legacy browser XSS protection (modern browsers use CSP)
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Strict-Transport-Security: Enforce HTTPS
		// SECURITY: Enable HSTS to prevent protocol downgrade attacks
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content-Security-Policy: Prevent XSS and data injection
		// Using 'default-src none' to be restrictive
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; form-action 'self'; connect-src 'self'")

		// Referrer-Policy: Control information leakage
		w.Header().Set("Referrer-Policy", "strict-origin-when-origin")

		// Permissions-Policy: Control browser features
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		next.ServeHTTP(w, r)
	})
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

		// SECURITY: Correlation ID only in logs, not exposed in response headers
		// (Was previously exposed which is a potential information disclosure)
		// w.Header().Set("X-Correlation-ID", correlationID) // REMOVED

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

// rateLimitMiddleware enforces rate limiting per-client to prevent DoS attacks
// SECURITY: Uses per-client rate limiting instead of global to prevent one attacker from blocking all users
func (p *Proxy) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// FIX #2: Concurrent request limiting using semaphore
		// SECURITY FIX: Added panic recovery to prevent semaphore leak
		select {
		case p.semaphore <- struct{}{}:
			defer func() {
				<-p.semaphore
				// Recover from any panic in the handler
				if rec := recover(); rec != nil {
					p.logger.Error("panic in request handler", WithError(fmt.Errorf("%v", rec)))
				}
			}()
		default:
			p.metrics.RecordRequest(false, 0, http.StatusServiceUnavailable)
			p.logger.Warn("concurrent request limit exceeded",
				WithCorrelationID(getCorrelationID(r.Context())),
				WithMethod(r.Method),
				WithPath(r.URL.Path),
				WithStatusCode(http.StatusServiceUnavailable),
			)
			p.sendErrorResponse(w, r, http.StatusServiceUnavailable, "Too many concurrent requests")
			return
		}

		// Get unique client identifier
		// SECURITY: Use IP-based only for rate limiting to avoid logging sensitive JWT data
		// Only trust X-Forwarded-For from trusted proxies
		clientIP := getClientIP(r, p.config.TrustedProxies)
		if clientIP == "" {
			clientIP = "unknown"
		}
		clientID := "ip:" + clientIP

		// SECURITY: Use per-client rate limiter (not global!)
		// This prevents one attacker from blocking all users
		if !p.clientRL.Allow(clientID) {
			p.metrics.RecordRequest(false, 0, http.StatusTooManyRequests)
			p.logger.Warn("rate limit exceeded",
				WithCorrelationID(getCorrelationID(r.Context())),
				WithMethod(r.Method),
				WithPath(r.URL.Path),
				WithStatusCode(http.StatusTooManyRequests),
				WithExtra("client_ip", maskIP(clientIP)), // SECURITY: Don't log full IP
			)
			p.sendErrorResponse(w, r, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// maskIP partially masks an IP for privacy in logs (e.g., 192.168.1.100 -> 192.168.1.x)
func maskIP(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) >= 4 {
		// IPv4: mask last octet
		return parts[0] + "." + parts[1] + "." + parts[2] + ".x"
	}
	// For shorter IPs or other formats, just show first part
	if len(ip) > 4 {
		return ip[:4] + "..."
	}
	return "***"
}

// authMiddleware validates JWT Bearer token authentication
func (p *Proxy) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// SECURITY FIX V3: Only /health and /ready are unprotected
		// /metrics now requires authentication to protect sensitive metrics
		if r.URL.Path == "/health" || r.URL.Path == "/ready" {
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

// FIX #1: If no JWT secret configured, fail hard in production but this is logged at startup
func (p *Proxy) validateJWT(tokenString string) error {
	// SECURITY FIX V1: Reject ALL requests if JWT is not configured in production
	// This is a critical security control - no auth = no requests
	if p.config.JWTSecret == "" {
		return fmt.Errorf("JWT authentication not configured - requests rejected")
	}

	// CRITICAL SECURITY FIX: Parse with explicit algorithm validation
	// This prevents algorithm confusion attacks where an attacker could try to use "none" or asymmetric keys
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// CRITICAL: Validate the signing algorithm is EXACTLY HMAC (HS256/HS384/HS512)
		// This is the FIRST line of defense against algorithm confusion attacks
		switch token.Method.Alg() {
		case jwt.SigningMethodHS256.Alg(), jwt.SigningMethodHS384.Alg(), jwt.SigningMethodHS512.Alg():
			// Valid HMAC algorithms - proceed
		default:
			// BLOCK ANY OTHER ALGORITHM - this prevents:
			// - "alg:none" attacks
			// - RSA/HMAC confusion
			// - Key confusion with asymmetric keys
			return nil, fmt.Errorf("invalid signing algorithm: %s (only HS256/HS384/HS512 allowed)", token.Method.Alg())
		}

		// CRITICAL: Verify the "alg" header hasn't been tampered with directly
		// Attackers might try to modify the header after signing
		if algHeader, ok := token.Header["alg"].(string); !ok || algHeader == "" {
			return nil, fmt.Errorf("missing algorithm header")
		}

		// Explicitly validate HMAC signing method to prevent bypass
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("signing method %s not allowed", token.Method.Alg())
		}

		return []byte(p.config.JWTSecret), nil
	})

	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return fmt.Errorf("invalid token claims")
	}

	// Additional validation: Ensure we have valid claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid token format")
	}

	// SECURITY FIX V7: Validate exp claim (REQUIRED for replay protection)
	if exp, exists := claims["exp"]; exists {
		if expFloat, ok := exp.(float64); ok {
			expTime := int64(expFloat)
			now := time.Now().Unix()
			// Reject tokens that expired more than 5 minutes ago (grace period for clock skew)
			if expTime < now-300 {
				return fmt.Errorf("token expired")
			}
			// Reject tokens with future exp more than 24 hours (prevent farming future tokens)
			if expTime > now+86400 {
				return fmt.Errorf("token exp too far in future")
			}
		}
	} else {
		// REJECT tokens without exp claim - critical for replay protection
		return fmt.Errorf("token missing exp claim - required for security")
	}

	// SECURITY FIX V7: Validate iat claim (issued at) to prevent replay
	if iat, exists := claims["iat"]; exists {
		if iatFloat, ok := iat.(float64); ok {
			iatTime := int64(iatFloat)
			now := time.Now().Unix()
			// Reject tokens issued in the future
			if iatTime > now+60 {
				return fmt.Errorf("token issued in the future")
			}
			// Reject tokens issued more than 24 hours ago
			if iatTime < now-86400 {
				return fmt.Errorf("token too old")
			}
		}
	} else {
		// RECOMMEND: Reject tokens without iat - helps with replay detection
		p.logger.Debug("token missing iat claim - recommending rejection for better security")
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

		// SECURITY FIX: Read the request body with size limit to prevent memory exhaustion
		// Use LimitReader to enforce max body size before reading
		limitedBody := io.LimitReader(r.Body, p.config.MaxBodySize+1)
		body, err := io.ReadAll(limitedBody)
		if err != nil {
			// Check if body exceeded limit
			if len(body) > int(p.config.MaxBodySize) {
				p.logger.Warn("request body size exceeds limit",
					WithCorrelationID(correlationID),
					WithExtra("body_size", len(body)),
					WithExtra("max_size", p.config.MaxBodySize),
					WithStatusCode(http.StatusRequestEntityTooLarge),
				)
				p.metrics.RecordRequest(false, 0, http.StatusRequestEntityTooLarge)
				p.sendErrorResponse(w, r, http.StatusRequestEntityTooLarge,
					fmt.Sprintf("Request body exceeds maximum size of %d bytes", p.config.MaxBodySize))
				return
			}
			p.logger.Error("failed to read request body",
				WithCorrelationID(correlationID),
				WithError(err),
			)
			p.sendErrorResponse(w, r, http.StatusBadRequest, "Failed to read request body")
			return
		}

		// Double-check size after reading (defense in depth)
		if len(body) > int(p.config.MaxBodySize) {
			p.logger.Warn("request body size exceeds limit after read",
				WithCorrelationID(correlationID),
				WithExtra("body_size", len(body)),
				WithExtra("max_size", p.config.MaxBodySize),
				WithStatusCode(http.StatusRequestEntityTooLarge),
			)
			p.metrics.RecordRequest(false, 0, http.StatusRequestEntityTooLarge)
			p.sendErrorResponse(w, r, http.StatusRequestEntityTooLarge,
				fmt.Sprintf("Request body exceeds maximum size of %d bytes", p.config.MaxBodySize))
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
			// SECURITY: Limit batch size to prevent batch attack vectors
			if len(parsed.BatchReqs) > MaxBatchSize {
				p.logger.Warn("batch size exceeds limit",
					WithCorrelationID(correlationID),
					WithExtra("batch_size", len(parsed.BatchReqs)),
					WithExtra("max_batch_size", MaxBatchSize),
					WithStatusCode(http.StatusBadRequest),
				)
				p.sendErrorResponse(w, r, http.StatusBadRequest,
					fmt.Sprintf("Batch size %d exceeds maximum %d", len(parsed.BatchReqs), MaxBatchSize))
				return
			}

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

		// SECURITY: Sanitize tool inputs before processing
		sanitizedToolName, sanitizedArgs, err := sanitizeToolInput(toolName, args)
		if err != nil {
			p.logger.Warn("tool input validation failed",
				WithCorrelationID(correlationID),
				WithExtra("tool", toolName),
				WithError(err),
				WithStatusCode(http.StatusBadRequest),
			)
			p.metrics.RecordRequest(false, 0, http.StatusBadRequest)
			p.sendErrorResponse(w, r, http.StatusBadRequest,
				fmt.Sprintf("Invalid tool input: %s", err.Error()))
			return
		}

		// Use sanitized values
		toolName = sanitizedToolName
		args = sanitizedArgs

		// Check with Lakera (if configured)
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// SECURITY: Handle nil lakeraClient gracefully
		// Check both interface nil and nil *LakeraClient pointer case
		clientNil := p.lakeraClient == nil
		if !clientNil {
			// If it's a *LakeraClient specifically, check if pointer is nil
			if lc, ok := p.lakeraClient.(*LakeraClient); ok && lc == nil {
				clientNil = true
			}
		}

		if clientNil {
			p.logger.Debug("lakera client not configured, using fail mode",
				WithCorrelationID(correlationID),
				WithExtra("fail_mode", p.config.FailMode),
			)
			// With nil lakera, apply fail mode
			if p.config.FailMode == "closed" {
				p.metrics.RecordRequest(false, 0, http.StatusServiceUnavailable)
				p.sendErrorResponse(w, r, http.StatusServiceUnavailable,
					"Semantic check unavailable - request blocked for security")
				return
			}
			// Fail-open: allow request to proceed
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

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

// checkBatchRequest checks a single request in a batch with input sanitization
func (p *Proxy) checkBatchRequest(req *ParsedRequest) (bool, string) {
	toolName, args, ok := GetToolInfo(req)
	if !ok {
		return true, "" // Not a tool call
	}

	// SECURITY: Sanitize inputs in batch requests too
	sanitizedToolName, sanitizedArgs, err := sanitizeToolInput(toolName, args)
	if err != nil {
		return false, fmt.Sprintf("input validation failed for tool '%s': %s", toolName, err.Error())
	}

	// Use sanitized values
	toolName = sanitizedToolName
	args = sanitizedArgs

	// SECURITY: Handle nil lakeraClient in batch check
	// Check both interface nil and nil *LakeraClient pointer case
	clientNil := p.lakeraClient == nil
	if !clientNil {
		// If it's a *LakeraClient specifically, check if pointer is nil
		if lc, ok := p.lakeraClient.(*LakeraClient); ok && lc == nil {
			clientNil = true
		}
	}

	if clientNil {
		// Apply fail mode for batch requests too
		if p.config.FailMode == "closed" {
			return false, "blocked: lakera unavailable (fail-closed)"
		}
		// Fail-open: allow
		return true, ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	allowed, score, reason, _ := p.lakeraClient.CheckToolCall(ctx, toolName, args)
	if !allowed {
		return false, fmt.Sprintf("tool '%s' (score: %d): %s", toolName, score, reason)
	}

	return true, ""
}

// forwardToMCP forwards the request to the MCP backend with SSRF protection
func (p *Proxy) forwardToMCP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	correlationID := getCorrelationID(r.Context())

	// SECURITY: Validate backend URL before making the request (SSRF protection)
	validatedURL, err := validateBackendURL(p.config.MCPBackendURL, r.URL.Path)
	if err != nil {
		p.logger.Error("backend URL validation failed",
			WithCorrelationID(correlationID),
			WithError(err),
			WithExtra("backend_url", p.config.MCPBackendURL),
			WithExtra("path", r.URL.Path),
		)
		p.sendErrorResponse(w, r, http.StatusInternalServerError, "Invalid backend configuration")
		return
	}

	// Create proxy request with context containing correlation ID
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, validatedURL, r.Body)
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

	// FIX #4: Use pooled backend client instead of creating new client per request
	// SECURITY: Use retry client with exponential backoff for resilience
	// Retry happens BEFORE circuit breaker consideration (retry first, then circuit open)
	resp, err := p.retryClient.Do(r.Context(), proxyReq)
	if err != nil {
		// SECURITY: Record failure in circuit breaker
		p.circuitBreaker.RecordFailure()

		// SECURITY: Enqueue to DLQ if fail-closed mode is active
		if p.config.FailMode == "closed" && p.dlq != nil {
			// Read body for DLQ storage
			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}

			// Build headers map
			headers := make(map[string]string)
			for k, v := range r.Header {
				if len(v) > 0 {
					headers[k] = v[0]
				}
			}

			failedReq := &dlq.FailedRequest{
				Timestamp:  time.Now(),
				Method:     r.Method,
				URL:        validatedURL,
				Headers:    headers,
				Body:       bodyBytes,
				Error:      err.Error(),
				RetryCount: 3, // Max retries from RetryClient
				Source:     "mcp-backend",
			}

			if err := p.dlq.Enqueue(r.Context(), failedReq); err != nil {
				p.logger.Error("failed to enqueue to DLQ",
					WithCorrelationID(correlationID),
					WithError(err),
				)
			} else {
				p.logger.Info("request enqueued to DLQ for later replay",
					WithCorrelationID(correlationID),
					WithError(err),
				)
			}
		}

		p.logger.Error("backend error",
			WithCorrelationID(correlationID),
			WithExtra("backend_url", p.config.MCPBackendURL),
			WithExtra("circuit_state", p.circuitBreaker.GetState()),
			WithExtra("dlq_enqueued", p.config.FailMode == "closed" && p.dlq != nil),
			WithError(err),
		)
		p.sendErrorResponse(w, r, http.StatusBadGateway, "MCP backend unavailable")
		return
	}
	defer resp.Body.Close()

	// SECURITY: Record success in circuit breaker
	p.circuitBreaker.RecordSuccess()

	// SECURITY: Whitelisted headers from backend - only pass safe headers
	whitelistedHeaders := map[string]bool{
		"content-type":     true,
		"content-length":   true,
		"content-encoding": true,
		"cache-control":    true,
		"etag":             true,
		"expires":          true,
		"last-modified":    true,
		"vary":             true,
	}

	// Copy whitelisted response headers
	for k, v := range resp.Header {
		if whitelistedHeaders[strings.ToLower(k)] {
			w.Header()[k] = v
		}
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
	_ = getCorrelationID(r.Context()) // Log correlation ID for debugging

	w.Header().Set("Content-Type", "application/json")
	// SECURITY: Correlation ID removed from error response headers
	// Only exposed in structured logs for debugging, not in HTTP headers

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
