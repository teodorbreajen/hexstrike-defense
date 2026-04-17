package main

import (
	"net/http"
	"strings"
)

// CORSMiddleware handles CORS (Cross-Origin Resource Sharing) for the proxy
type CORSMiddleware struct {
	allowedOrigins   map[string]bool
	allowCredentials bool
}

// NewCORSMiddleware creates a new CORS middleware with the given allowed origins
func NewCORSMiddleware(allowedOrigins []string, allowCredentials bool) *CORSMiddleware {
	// Build a map for O(1) lookup
	originMap := make(map[string]bool)
	for _, origin := range allowedOrigins {
		originMap[strings.ToLower(origin)] = true
	}

	return &CORSMiddleware{
		allowedOrigins:   originMap,
		allowCredentials: allowCredentials,
	}
}

// Handle implements the Middleware interface for CORS
func (c *CORSMiddleware) Handle(next http.Handler) http.Handler {
	// If no origins configured, skip CORS middleware entirely
	if len(c.allowedOrigins) == 0 {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Origin header from the request
		origin := r.Header.Get("Origin")

		// If no Origin header, continue without CORS headers
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Check if the origin is allowed (case-insensitive lookup)
		originLower := strings.ToLower(origin)
		if !c.allowedOrigins[originLower] {
			// Origin not in allowlist - deny CORS request
			next.ServeHTTP(w, r)
			return
		}

		// This is a CORS request from an allowed origin - add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", origin)

		// Add Access-Control-Allow-Credentials if configured
		if c.allowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Add allowed methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

		// Add allowed headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Correlation-ID")

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			// Check for required preflight headers
			accessControlRequestMethod := r.Header.Get("Access-Control-Request-Method")
			if accessControlRequestMethod != "" {
				// This is a valid preflight request
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		// Continue to the next handler
		next.ServeHTTP(w, r)
	})
}

// Middleware returns the CORS middleware for integration into the middleware chain
func (c *CORSMiddleware) Middleware() Middleware {
	return c.Handle
}
