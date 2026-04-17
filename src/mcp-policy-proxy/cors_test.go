package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCORSMiddleware_NoOriginHeader tests that no CORS headers are added when Origin is empty
func TestCORSMiddleware_NoOriginHeader(t *testing.T) {
	middleware := NewCORSMiddleware([]string{"https://example.com"}, true)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify no CORS headers were set by middleware
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), "no Origin header should not set CORS headers")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	// No Origin header set
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestCORSMiddleware_AllowedOrigin tests that CORS headers are correctly set for allowed origin
func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	tests := []struct {
		name               string
		allowedOrigins     []string
		requestOrigin      string
		expectAllowOrigin  bool
		expectAllowCreds   bool
		expectAllowMethods bool
		expectAllowHeaders bool
	}{
		{
			name:               "single allowed origin exact match",
			allowedOrigins:     []string{"https://example.com"},
			requestOrigin:      "https://example.com",
			expectAllowOrigin:  true,
			expectAllowCreds:   true,
			expectAllowMethods: true,
			expectAllowHeaders: true,
		},
		{
			name:               "multiple allowed origins - first match",
			allowedOrigins:     []string{"https://example.com", "https://app.example.org"},
			requestOrigin:      "https://example.com",
			expectAllowOrigin:  true,
			expectAllowCreds:   true,
			expectAllowMethods: true,
			expectAllowHeaders: true,
		},
		{
			name:               "multiple allowed origins - second match",
			allowedOrigins:     []string{"https://example.com", "https://app.example.org"},
			requestOrigin:      "https://app.example.org",
			expectAllowOrigin:  true,
			expectAllowCreds:   true,
			expectAllowMethods: true,
			expectAllowHeaders: true,
		},
		{
			name:               "case insensitive origin matching",
			allowedOrigins:     []string{"https://EXAMPLE.COM"},
			requestOrigin:      "https://example.com",
			expectAllowOrigin:  true,
			expectAllowCreds:   true,
			expectAllowMethods: true,
			expectAllowHeaders: true,
		},
		{
			name:               "case insensitive - request has uppercase",
			allowedOrigins:     []string{"https://example.com"},
			requestOrigin:      "HTTPS://EXAMPLE.COM",
			expectAllowOrigin:  true,
			expectAllowCreds:   true,
			expectAllowMethods: true,
			expectAllowHeaders: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewCORSMiddleware(tt.allowedOrigins, true)

			var capturedOrigin string
			handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedOrigin = w.Header().Get("Access-Control-Allow-Origin")
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.Header.Set("Origin", tt.requestOrigin)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if tt.expectAllowOrigin {
				assert.Equal(t, tt.requestOrigin, capturedOrigin, "Access-Control-Allow-Origin should match request origin")
				assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"), "Credentials should be allowed")
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET", "Should contain GET method")
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type", "Should contain Content-Type header")
			}
		})
	}
}

// TestCORSMiddleware_DeniedOrigin tests that denied origins get no CORS headers
func TestCORSMiddleware_DeniedOrigin(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
	}{
		{
			name:           "origin not in allowlist",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://evil.com",
		},
		{
			name:           "partial match not allowed",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://example.com.evil.com",
		},
		{
			name:           "different subdomain not allowed",
			allowedOrigins: []string{"https://app.example.com"},
			requestOrigin:  "https://example.com",
		},
		{
			name:           "http vs https difference",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "http://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewCORSMiddleware(tt.allowedOrigins, true)

			handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// No CORS headers should be set for denied origin
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.Header.Set("Origin", tt.requestOrigin)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

// TestCORSMiddleware_PreflightOPTIONS tests preflight OPTIONS requests return 204
func TestCORSMiddleware_PreflightOPTIONS(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		expectStatus   int
	}{
		{
			name:           "valid preflight with Access-Control-Request-Method",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://example.com",
			expectStatus:   http.StatusNoContent,
		},
		{
			name:           "preflight without Access-Control-Request-Method from denied origin",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://evil.com",
			expectStatus:   http.StatusOK, // Denied origin passes through to handler
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewCORSMiddleware(tt.allowedOrigins, true)

			handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodOptions, "/api/test", nil)
			req.Header.Set("Origin", tt.requestOrigin)
			req.Header.Set("Access-Control-Request-Method", "POST")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectStatus, w.Code)
		})
	}
}

// TestCORSMiddleware_PreflightWithHeaders tests preflight with Access-Control-Request-Headers
func TestCORSMiddleware_PreflightWithHeaders(t *testing.T) {
	middleware := NewCORSMiddleware([]string{"https://example.com"}, true)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

// TestCORSMiddleware_NoOriginsConfigured tests that middleware passes through when no origins configured
func TestCORSMiddleware_NoOriginsConfigured(t *testing.T) {
	middleware := NewCORSMiddleware([]string{}, true)

	var called bool
	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, called, "handler should be called")
	assert.Equal(t, http.StatusOK, w.Code)
	// No CORS headers should be set when no origins configured
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

// TestCORSMiddleware_CredentialsDisabled tests behavior when credentials are disabled
func TestCORSMiddleware_CredentialsDisabled(t *testing.T) {
	middleware := NewCORSMiddleware([]string{"https://example.com"}, false)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"), "Credentials header should not be set when disabled")
}

// TestCORSMiddleware_AllHTTPMethodsAllowed tests that all standard methods are allowed
func TestCORSMiddleware_AllHTTPMethodsAllowed(t *testing.T) {
	middleware := NewCORSMiddleware([]string{"https://example.com"}, true)

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/test", nil)
			req.Header.Set("Origin", "https://example.com")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			allowedMethods := w.Header().Get("Access-Control-Allow-Methods")
			assert.Contains(t, allowedMethods, method, "Method %s should be allowed", method)
		})
	}
}

// TestNewCORSMiddleware_CaseInsensitiveStorage tests that origins are stored case-insensitively
func TestNewCORSMiddleware_CaseInsensitiveStorage(t *testing.T) {
	// Test that the constructor stores origins in lowercase
	middleware := NewCORSMiddleware([]string{"HTTPS://EXAMPLE.COM", "https://APP.EXAMPLE.ORG"}, true)

	// The middleware should match case-insensitively
	assert.NotNil(t, middleware)

	// Verify through the Handle method
	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test various case combinations
	testOrigins := []string{
		"https://example.com",
		"HTTPS://EXAMPLE.COM",
		"Https://Example.Com",
		"https://app.example.org",
		"HTTPS://APP.EXAMPLE.ORG",
	}

	for _, origin := range testOrigins {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Origin", origin)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Origin %s should be allowed", origin)
		assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"))
	}
}
