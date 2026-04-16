package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// TestProxyHandler_HealthEndpointReturns200 tests that the health endpoint returns 200
func TestProxyHandler_HealthEndpointReturns200(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "health endpoint",
			path:           "/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ready endpoint",
			path:           "/ready",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "metrics endpoint",
			path:           "/metrics",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := createTestProxy()
			mux := createTestRouter(proxy)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "health endpoint should return correct status")
		})
	}
}

// TestProxyHandler_AuthMiddlewareBlocksWithoutTokenOnMCPEndpoint tests that auth middleware blocks without token on /mcp/*
func TestProxyHandler_AuthMiddlewareBlocksWithoutTokenOnMCPEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		authHeader     string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "no auth header",
			path:           "/mcp/v1/tools",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Missing Authorization header",
		},
		{
			name:           "invalid auth header format",
			path:           "/mcp/v1/tools",
			authHeader:     "Basic abc123",
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Invalid Authorization format - expected Bearer token",
		},
		{
			name:           "invalid token",
			path:           "/mcp/v1/tools",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "Invalid or expired token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := createTestProxyWithJWT()
			mux := createTestRouter(proxy)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "status should match")

			// Check error message if expected
			if tt.expectedMsg != "" {
				body := w.Body.String()
				assert.True(t, strings.Contains(body, tt.expectedMsg),
					"response should contain expected message: %s", tt.expectedMsg)
			}
		})
	}
}

// TestProxyHandler_AuthMiddlewareAllowsValidJWT tests that auth middleware allows valid JWT token
func TestProxyHandler_AuthMiddlewareAllowsValidJWT(t *testing.T) {
	proxy := createTestProxyWithJWT()
	mux := createTestRouter(proxy)

	// Create a valid JWT token
	token := createTestJWT(t, proxy.config.JWTSecret)

	req := httptest.NewRequest(http.MethodGet, "/mcp/v1/tools", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	// Should either get 200 (forwarded) or 502 (backend unavailable is OK - we're testing auth)
	// The key is that it's NOT a 401
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadGateway,
		"valid JWT should not be rejected with 401")
}

// TestProxyHandler_BodySizeLimitReturns413ForOversized tests that body size limit returns 413 for oversized
func TestProxyHandler_BodySizeLimitReturns413ForOversized(t *testing.T) {
	tests := []struct {
		name           string
		bodySize       int64
		maxBodySize    int64
		expectedStatus int
	}{
		{
			name:           "body above limit returns 413",
			bodySize:       4096,
			maxBodySize:    2048,
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := createTestProxyWithConfig(&ProxyConfig{
				MaxBodySize:        tt.maxBodySize,
				RateLimitPerMinute: 60,
				MCPBackendURL:      "http://localhost:9090",
				Timeout:            30 * time.Second,
				JWTSecret:          "test-secret",
			})
			mux := createTestRouter(proxy)

			// Create oversized body
			body := strings.Repeat("a", int(tt.bodySize))

			req := httptest.NewRequest(http.MethodPost, "/mcp/v1/call", strings.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+createTestJWT(t, proxy.config.JWTSecret))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "status should match body size limit")

			// Verify error message
			bodyStr := w.Body.String()
			assert.True(t, strings.Contains(bodyStr, "exceeds maximum size"),
				"response should mention body size limit")
		})
	}
}

// TestProxyHandler_RateLimitingReturns429WhenExhausted tests that rate limiting returns 429 when exhausted
func TestProxyHandler_RateLimitingReturns429WhenExhausted(t *testing.T) {
	// Create a rate limiter with very low limit
	proxy := createTestProxyWithConfig(&ProxyConfig{
		RateLimitPerMinute: 2, // Very low limit for testing
		MCPBackendURL:      "http://localhost:9090",
		Timeout:            30 * time.Second,
		JWTSecret:          "test-secret",
	})
	mux := createTestRouter(proxy)

	// Exhaust the rate limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/mcp/v1/tools", nil)
		req.Header.Set("Authorization", "Bearer "+createTestJWT(t, proxy.config.JWTSecret))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
	}

	// Next request should get 429
	req := httptest.NewRequest(http.MethodGet, "/mcp/v1/tools", nil)
	req.Header.Set("Authorization", "Bearer "+createTestJWT(t, proxy.config.JWTSecret))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "rate limited request should return 429")

	// Verify the response contains "Rate limit exceeded"
	body := w.Body.String()
	assert.True(t, strings.Contains(body, "Rate limit exceeded"),
		"response should contain rate limit message")
}

// TestProxyHandler_HealthEndpointReturnsJSON tests that health endpoint returns proper JSON
func TestProxyHandler_HealthEndpointReturnsJSON(t *testing.T) {
	proxy := createTestProxy()
	mux := createTestRouter(proxy)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"),
		"health endpoint should return JSON content type")

	// Try to parse as JSON
	var resp healthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "response should be valid JSON")
	assert.NotEmpty(t, resp.Status, "status should not be empty")
}

// TestProxyHandler_ReadyEndpoint tests the /ready endpoint
func TestProxyHandler_ReadyEndpoint(t *testing.T) {
	proxy := createTestProxy()
	mux := createTestRouter(proxy)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "ready endpoint should return 200")

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "response should be valid JSON")
	assert.Equal(t, "ready", resp["status"], "status should be ready")
}

// TestProxyHandler_UnprotectedEndpointsDontRequireAuth tests unprotected endpoints don't require auth
func TestProxyHandler_UnprotectedEndpointsDontRequireAuth(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "health endpoint",
			path: "/health",
		},
		{
			name: "ready endpoint",
			path: "/ready",
		},
		{
			name: "metrics endpoint",
			path: "/metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := createTestProxyWithJWT() // JWT configured
			mux := createTestRouter(proxy)

			// No auth header
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			// Should NOT return 401 - these are unprotected
			assert.NotEqual(t, http.StatusUnauthorized, w.Code,
				"unprotected endpoint should not require auth")
		})
	}
}

// TestProxyHandler_CorrectErrorResponseFormat tests error responses are in correct format
func TestProxyHandler_CorrectErrorResponseFormat(t *testing.T) {
	proxy := createTestProxyWithJWT()
	mux := createTestRouter(proxy)

	// Request to /mcp without auth should return error
	req := httptest.NewRequest(http.MethodGet, "/mcp/v1/tools", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	// Should be JSON error response
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"),
		"error response should be JSON")

	// Should be parseable as JSON-RPC error
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err, "error response should be valid JSON")
}

// Helper functions for testing

// createTestProxy creates a proxy with default test configuration
func createTestProxy() *Proxy {
	return createTestProxyWithConfig(&ProxyConfig{
		RateLimitPerMinute: 60,
		MCPBackendURL:      "http://localhost:9090",
		Timeout:            30 * time.Second,
		MaxBodySize:        1048576, // 1MB
		JWTSecret:          "",
	})
}

// createTestProxyWithJWT creates a proxy with JWT secret configured
func createTestProxyWithJWT() *Proxy {
	return createTestProxyWithConfig(&ProxyConfig{
		RateLimitPerMinute: 60,
		MCPBackendURL:      "http://localhost:9090",
		Timeout:            30 * time.Second,
		MaxBodySize:        1048576,
		JWTSecret:          "test-secret-key-for-testing",
	})
}

// createTestProxyWithConfig creates a proxy with custom configuration
func createTestProxyWithConfig(config *ProxyConfig) *Proxy {
	// Use nil lakeraClient for testing (Lakera checks won't run)
	proxy := NewProxy(config, nil)
	return proxy
}

// createTestRouter creates an HTTP test router
func createTestRouter(proxy *Proxy) *http.ServeMux {
	mux := http.NewServeMux()

	// Health endpoints
	mux.HandleFunc("/health", HealthHandler(nil))
	mux.HandleFunc("/ready", ReadinessHandler())

	// Metrics endpoint
	mux.HandleFunc("/metrics", proxy.GetMetricsHandler())

	// MCP proxy endpoint (catch-all)
	mux.Handle("/", proxy.Handler())

	return mux
}

// createTestJWT creates a valid JWT token for testing
func createTestJWT(t *testing.T, secret string) string {
	t.Helper()

	if secret == "" {
		return ""
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to create test JWT: %v", err)
	}

	return tokenString
}
