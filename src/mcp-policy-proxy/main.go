package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Config holds all environment-based configuration
type Config struct {
	// Server
	ListenAddr string `json:"listen_addr"`
	// SECURITY FIX: TLS configuration
	TLSEnabled  bool   `json:"tls_enabled"` // Enable TLS
	TLSCertFile string `json:"tls_cert"`    // TLS certificate file path
	TLSKeyFile  string `json:"tls_key"`     // TLS key file path

	// MCP Backend
	MCPBackendURL string `json:"mcp_backend_url"`

	// Lakera
	LakeraAPIKey    string `json:"lakera_api_key"`
	LakeraURL       string `json:"lakera_url"`
	LakeraTimeout   int    `json:"lakera_timeout"`   // seconds
	LakeraThreshold int    `json:"lakera_threshold"` // SECURITY: configurable threshold (default 70)

	// Rate Limiting
	RateLimitPerMinute int `json:"rate_limit_per_minute"`

	// Proxy
	ProxyTimeout int `json:"proxy_timeout"` // seconds

	// Security
	FailMode       string `json:"fail_mode"`       // "closed" or "open" - fail-closed blocks on Lakera errors
	MaxBodySize    int64  `json:"max_body_size"`   // Maximum request body size in bytes
	JWTSecret      string `json:"jwt_secret"`      // JWT secret for auth validation (REQUIRED in production)
	ExposeHealth   bool   `json:"expose_health"`   // SECURITY: expose health details (default false for security)
	TrustedProxies string `json:"trusted_proxies"` // Comma-separated list of trusted proxy IPs/CIDR for X-Forwarded-For
	DevMode        bool   `json:"dev_mode"`        // FIX #3: Allow localhost in development
	// CORS
	CORSAllowedOrigins   []string `json:"cors_allowed_origins"`   // Comma-separated list of allowed origins
	CORSAllowCredentials bool     `json:"cors_allow_credentials"` // Allow credentials in CORS responses
	// DLQ
	DLQPath     string `json:"dlq_path"`      // Path for DLQ storage (default: data/dlq)
	DLQTTLHours int    `json:"dlq_ttl_hours"` // TTL for DLQ messages in hours (default: 24)
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	exposeHealth := getEnv("EXPOSE_HEALTH", "false") == "true"
	devMode := getEnv("DEV_MODE", "false") == "true" || getEnv("DEV_MODE", "false") == "1"

	// JWT fail-hard: check GIN_MODE for production detection
	ginMode := getEnv("GIN_MODE", "debug")
	isProduction := ginMode == "release"

	// Parse CORS allowed origins from comma-separated string
	corsOriginsStr := getEnv("CORS_ALLOWED_ORIGINS", "")
	var corsOrigins []string
	if corsOriginsStr != "" && corsOriginsStr != "disabled" {
		origins := strings.Split(corsOriginsStr, ",")
		for _, o := range origins {
			origin := strings.TrimSpace(o)
			if origin != "" {
				corsOrigins = append(corsOrigins, origin)
			}
		}
	}

	config := &Config{
		ListenAddr:           getEnv("LISTEN_ADDR", "127.0.0.1:8080"),
		TLSEnabled:           getEnv("TLS_ENABLED", "false") == "true",
		TLSCertFile:          getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:           getEnv("TLS_KEY_FILE", ""),
		MCPBackendURL:        getEnv("MCP_BACKEND_URL", "http://localhost:9090"),
		LakeraAPIKey:         getEnv("LAKERA_API_KEY", ""),
		LakeraURL:            getEnv("LAKERA_URL", "https://api.lakera.ai"),
		LakeraTimeout:        5,
		LakeraThreshold:      getEnvInt("LAKERA_THRESHOLD", 70), // SECURITY: configurable threshold (default 70)
		RateLimitPerMinute:   getEnvInt("RATE_LIMIT_PER_MINUTE", 60),
		ProxyTimeout:         30,
		FailMode:             getEnv("LAKERA_FAIL_MODE", "closed"),  // Fail-closed by default for security
		MaxBodySize:          getEnvInt64("MAX_BODY_SIZE", 1048576), // 1MB default
		JWTSecret:            getEnv("JWT_SECRET", ""),
		ExposeHealth:         exposeHealth,
		TrustedProxies:       getEnv("TRUSTED_PROXIES", ""), // SECURITY: Only trust these IPs/CIDR for X-Forwarded-For
		DevMode:              devMode,
		CORSAllowedOrigins:   corsOrigins,
		CORSAllowCredentials: getEnv("CORS_ALLOW_CREDENTIALS", "false") == "true",
		DLQPath:              getEnv("DLQ_PATH", "data/dlq"),
		DLQTTLHours:          getEnvInt("DLQ_TTL_HOURS", 24),
	}

	// SECURITY FIX: Warn if TLS enabled but no cert/key provided
	if config.TLSEnabled {
		if config.TLSCertFile == "" || config.TLSKeyFile == "" {
			log.Printf("WARNING: TLS_ENABLED but TLS_CERT_FILE or TLS_KEY_FILE not set")
		}
	}

	// FIX #3: Validate backend URL - block localhost in production
	backendURL := config.MCPBackendURL
	isLocalhost := strings.HasPrefix(backendURL, "http://localhost") ||
		strings.HasPrefix(backendURL, "http://127.0.0.1") ||
		strings.HasPrefix(backendURL, "http://10.") ||
		strings.HasPrefix(backendURL, "http://192.168") ||
		strings.HasPrefix(backendURL, "http://172.1") // 172.16-31.x

	if isLocalhost {
		if !devMode {
			log.Fatalf("FATAL: MCP_BACKEND_URL cannot be localhost/internal IP in production. Got: %s. Set DEV_MODE=true to override.", backendURL)
		}
		log.Printf("WARNING: Using localhost/internal backend - NOT for production! Set DEV_MODE=false for production.")
	}

	// SECURITY FIX V1: JWT mandatory in production - fail hard if not configured
	// Uses GIN_MODE=release to detect production (not DEV_MODE)
	if config.JWTSecret == "" {
		if isProduction {
			log.Fatalf("FATAL: JWT_SECRET is REQUIRED in production (GIN_MODE=release). Set JWT_SECRET for security.")
		}
		log.Printf("WARNING: JWT_SECRET not set - authentication DISABLED. DEV MODE ONLY!")
	}

	// SECURITY: Warn about X-Forwarded-For without trusted proxies
	if config.TrustedProxies == "" && config.ListenAddr != "127.0.0.1:8080" && config.ListenAddr != "localhost:8080" {
		log.Printf("WARNING: TRUSTED_PROXIES not set - X-Forwarded-For will not be trusted in production")
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		var intVal int64
		if _, err := fmt.Sscanf(value, "%d", &intVal); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// loadConfigFromFile loads configuration from a JSON file
func loadConfigFromFile(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Failed to read config file: %v", err)
		log.Printf("Using environment variables instead")
		return loadConfig()
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Failed to parse config file: %v", err)
		log.Printf("Using environment variables instead")
		return loadConfig()
	}

	return &config
}

// healthResponse represents the health check response
type healthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// securityHeaderHandler wraps http.Handler to add security headers to all responses
type securityHeaderHandler struct {
	h http.Handler
}

func (sh securityHeaderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// SECURITY HEADERS:
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
	w.Header().Set("Referrer-Policy", "strict-origin-when-origin")
	w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	sh.h.ServeHTTP(w, r)
}

// HealthHandler handles health check requests
func HealthHandler(lakeraClient *LakeraClient, exposeHealth bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checks := make(map[string]string)

		// Check Lakera connectivity (if configured)
		// SECURITY: Only expose Lakera status if exposeHealth is true
		if exposeHealth {
			if lakeraClient != nil && lakeraClient.GetConfig().APIKey != "" {
				ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
				defer cancel()

				if err := lakeraClient.HealthCheck(ctx); err != nil {
					checks["lakera"] = fmt.Sprintf("unavailable: %v", err)
				} else {
					checks["lakera"] = "ok"
				}
			} else {
				checks["lakera"] = "not configured"
			}
		} else {
			// SECURITY: When exposeHealth=false, never reveal whether Lakera is configured
			// Always return "ok" to avoid information disclosure
			checks["lakera"] = "ok"
		}

		// Determine overall status
		status := "healthy"
		for _, v := range checks {
			if v != "ok" && v != "not configured" {
				status = "degraded"
				break
			}
		}

		resp := healthResponse{
			Status:    status,
			Timestamp: time.Now(),
			Checks:    checks,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// ReadinessHandler handles readiness check requests
func ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ready",
		})
		w.WriteHeader(http.StatusOK)
	}
}

// main is the entry point for the MCP Policy Proxy
func main() {
	// Parse command-line flags
	configFile := flag.String("config", "", "Path to config file (JSON)")
	showVersion := flag.Bool("version", false, "Show version")
	insecureDev := flag.Bool("insecure-dev", false, "FIX #3: Allow localhost backend (NOT for production)")
	flag.Parse()

	if *showVersion {
		fmt.Println("HexStrike Defense v2.0.0")
		fmt.Println("Security-hardened semantic firewall for MCP tool calls")
		fmt.Println("")
		fmt.Println("Security Features:")
		fmt.Println("  - JWT Bearer authentication (HS256/384/512)")
		fmt.Println("  - SSRF protection with IP allowlist")
		fmt.Println("  - Input sanitization & injection prevention")
		fmt.Println("  - Fail-closed mode by default")
		fmt.Println("  - Circuit breaker & retry with backoff")
		fmt.Println("  - Per-client rate limiting")
		fmt.Println("  - Structured JSON logging")
		fmt.Println("")
		fmt.Println("For more info: https://github.com/teodorbreajen/hexstrike-defense")
		return
	}

	// Load configuration
	var config *Config
	if *configFile != "" {
		config = loadConfigFromFile(*configFile)
	} else {
		config = loadConfig()
	}

	// FIX #3: Override dev mode from command line
	if *insecureDev {
		config.DevMode = true
	}

	log.Printf("Starting HexStrike Defense v2.0.0")
	log.Printf("Listen: %s", config.ListenAddr)
	if config.TLSEnabled {
		log.Printf("TLS: ENABLED (cert: %s)", config.TLSCertFile)
	} else {
		log.Printf("TLS: DISABLED (set TLS_ENABLED=true for production)")
	}
	log.Printf("MCP Backend: %s", config.MCPBackendURL)
	log.Printf("Lakera Threshold: %d (configurable via LAKERA_THRESHOLD)", config.LakeraThreshold)

	// Create Lakera client with configurable threshold
	lakeraConfig := &LakeraConfig{
		APIKey:    config.LakeraAPIKey,
		Threshold: config.LakeraThreshold, // SECURITY: Now configurable via LAKERA_THRESHOLD
		Timeout:   time.Duration(config.LakeraTimeout) * time.Second,
		BaseURL:   config.LakeraURL,
	}
	lakeraClient := NewLakeraClient(lakeraConfig)

	// Create proxy configuration
	proxyConfig := &ProxyConfig{
		ListenAddr:           config.ListenAddr,
		MCPBackendURL:        config.MCPBackendURL,
		RateLimitPerMinute:   config.RateLimitPerMinute,
		Timeout:              time.Duration(config.ProxyTimeout) * time.Second,
		FailMode:             config.FailMode,
		MaxBodySize:          config.MaxBodySize,
		JWTSecret:            config.JWTSecret,
		TrustedProxies:       config.TrustedProxies,
		CORSAllowedOrigins:   config.CORSAllowedOrigins,
		CORSAllowCredentials: config.CORSAllowCredentials,
		DLQPath:              config.DLQPath,
		DLQTTLHours:          config.DLQTTLHours,
	}

	// Create proxy
	proxy := NewProxy(proxyConfig, lakeraClient)

	// Initialize Prometheus metrics
	promMetrics := NewPrometheusMetrics()

	// Create router using createRouter helper
	mux := createRouter(proxy, lakeraClient, config.ExposeHealth, promMetrics)

	// Create server
	server := &http.Server{
		Addr:         config.ListenAddr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on %s", config.ListenAddr)

		var err error
		if config.TLSEnabled && config.TLSCertFile != "" && config.TLSKeyFile != "" {
			// SECURITY FIX: Use TLS for production
			log.Printf("Starting TLS server with cert: %s", config.TLSCertFile)
			err = server.ListenAndServeTLS(config.TLSCertFile, config.TLSKeyFile)
		} else {
			// Development mode - no TLS
			if config.ListenAddr != "127.0.0.1:8080" && config.ListenAddr != "localhost:8080" {
				log.Printf("WARNING: Running without TLS on non-localhost address!")
			}
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// SECURITY FIX: Improved graceful shutdown with proper drain timeout
	// Wait for active connections to finish with longer timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First shutdown - stop accepting new connections
	if err := server.Shutdown(ctx); err != nil {
		// If graceful shutdown fails, force close
		log.Printf("Graceful shutdown failed, forcing close: %v", err)
		if err := server.Close(); err != nil {
			log.Printf("Force close error: %v", err)
		}
	}

	log.Println("Server stopped gracefully")
}

// createRouter sets up the HTTP routes with security middleware
func createRouter(proxy *Proxy, lakeraClient *LakeraClient, exposeHealth bool, promMetrics *PrometheusMetrics) *http.ServeMux {
	mux := http.NewServeMux()

	// Health endpoint with security headers - configurable exposure
	healthHandler := securityHeaderHandler{HealthHandler(lakeraClient, exposeHealth)}
	mux.Handle("/health", healthHandler)

	// Ready endpoint with security headers
	readyHandler := securityHeaderHandler{ReadinessHandler()}
	mux.Handle("/ready", readyHandler)

	// Prometheus metrics endpoint (format: text/plain; version=0.0.4)
	if promMetrics != nil {
		promHandler := securityHeaderHandler{NewPrometheusHandler(promMetrics)}
		mux.Handle("/metrics", promHandler)
	} else {
		// Fall back to JSON metrics if Prometheus not configured
		metricsHandler := securityHeaderHandler{proxy.GetMetricsHandler()}
		mux.Handle("/metrics", metricsHandler)
	}

	// MCP proxy endpoint (catch-all with full middleware chain)
	mux.Handle("/", proxy.Handler())

	return mux
}
