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
	"syscall"
	"time"
)

// Config holds all environment-based configuration
type Config struct {
	// Server
	ListenAddr string `json:"listen_addr"`

	// MCP Backend
	MCPBackendURL string `json:"mcp_backend_url"`

	// Lakera
	LakeraAPIKey  string `json:"lakera_api_key"`
	LakeraURL     string `json:"lakera_url"`
	LakeraTimeout int    `json:"lakera_timeout"` // seconds

	// Rate Limiting
	RateLimitPerMinute int `json:"rate_limit_per_minute"`

	// Proxy
	ProxyTimeout int `json:"proxy_timeout"` // seconds
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	config := &Config{
		ListenAddr:         getEnv("LISTEN_ADDR", "0.0.0.0:8080"),
		MCPBackendURL:      getEnv("MCP_BACKEND_URL", "http://localhost:9090"),
		LakeraAPIKey:       getEnv("LAKERA_API_KEY", ""),
		LakeraURL:          getEnv("LAKERA_URL", "https://api.lakera.ai"),
		LakeraTimeout:      5,
		RateLimitPerMinute: getEnvInt("RATE_LIMIT_PER_MINUTE", 60),
		ProxyTimeout:       30,
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

// healthResponse represents the health check response
type healthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// HealthHandler handles health check requests
func HealthHandler(lakeraClient *LakeraClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checks := make(map[string]string)

		// Check Lakera connectivity (if configured)
		if lakeraClient != nil && lakeraClient.GetConfig().APIKey != "" {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			if err := lakeraClient.HealthCheck(ctx); err != nil {
				checks["lakera"] = fmt.Sprintf("unavailable: %v", err)
			} else {
				checks["lakera"] = "ok"
			}
		} else {
			checks["lakera"] = "not_configured"
		}

		// Determine overall status
		status := "healthy"
		for _, v := range checks {
			if v != "ok" && v != "not_configured" {
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

// readinessHandler handles readiness check requests
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
	flag.Parse()

	if *showVersion {
		fmt.Println("MCP Policy Proxy v1.0.0")
		fmt.Println("Semantic firewall for MCP tool calls")
		return
	}

	// Load configuration
	var config *Config
	if *configFile != "" {
		config = loadConfigFromFile(*configFile)
	} else {
		config = loadConfig()
	}

	log.Printf("Starting MCP Policy Proxy...")
	log.Printf("Listen: %s", config.ListenAddr)
	log.Printf("MCP Backend: %s", config.MCPBackendURL)

	// Create Lakera client
	lakeraConfig := &LakeraConfig{
		APIKey:    config.LakeraAPIKey,
		Threshold: 70, // Default, could be configurable
		Timeout:   time.Duration(config.LakeraTimeout) * time.Second,
		BaseURL:   config.LakeraURL,
	}
	lakeraClient := NewLakeraClient(lakeraConfig)

	// Create proxy configuration
	proxyConfig := &ProxyConfig{
		ListenAddr:         config.ListenAddr,
		MCPBackendURL:      config.MCPBackendURL,
		RateLimitPerMinute: config.RateLimitPerMinute,
		Timeout:            time.Duration(config.ProxyTimeout) * time.Second,
	}

	// Create proxy
	proxy := NewProxy(proxyConfig, lakeraClient)

	// Create router
	mux := http.NewServeMux()

	// Health endpoints
	mux.HandleFunc("/health", HealthHandler(lakeraClient))
	mux.HandleFunc("/ready", ReadinessHandler())

	// Metrics endpoint
	mux.HandleFunc("/metrics", proxy.GetMetricsHandler())

	// MCP proxy endpoint (catch-all)
	mux.Handle("/", proxy.Handler())

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
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

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
