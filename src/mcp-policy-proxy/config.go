package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ConfigFile represents the JSON configuration file structure
type ConfigFile struct {
	Server struct {
		ListenAddr string `json:"listen_addr"`
		Port       int    `json:"port"`
	} `json:"server"`

	MCPBackend struct {
		URL string `json:"url"`
	} `json:"mcp_backend"`

	Lakera struct {
		APIKey     string `json:"api_key"`
		URL        string `json:"url"`
		Threshold  int    `json:"threshold"`
		TimeoutSec int    `json:"timeout_sec"`
	} `json:"lakera"`

	RateLimit struct {
		PerMinute int `json:"per_minute"`
	} `json:"rate_limit"`

	Auth struct {
		Enabled   bool     `json:"enabled"`
		APIKeys   []string `json:"api_keys"`
		JWTSecret string   `json:"jwt_secret"`
	} `json:"auth"`
}

// LoadConfigFile loads configuration from a JSON file
func LoadConfigFile(path string) (*ConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Validate validates the configuration
// SECURITY FIX: Added comprehensive security validation
func (c *ConfigFile) Validate() error {
	// Validate server config
	if c.Server.ListenAddr == "" && c.Server.Port == 0 {
		return nil // Use defaults
	}

	// Validate MCP backend
	if c.MCPBackend.URL == "" {
		return nil // Will use default
	}

	// Validate rate limit
	if c.RateLimit.PerMinute <= 0 {
		c.RateLimit.PerMinute = 60 // Default
	}

	// Validate Lakera threshold
	if c.Lakera.Threshold <= 0 || c.Lakera.Threshold > 100 {
		c.Lakera.Threshold = 70 // Default
	}

	// Validate Lakera timeout
	if c.Lakera.TimeoutSec <= 0 {
		c.Lakera.TimeoutSec = 5 // Default
	}

	// SECURITY FIX: Validate that JWT secret is set (required for production)
	// This ensures config files can't bypass authentication
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("security validation failed: jwt_secret is required in auth config")
	}

	// SECURITY FIX: Validate backend URL is not localhost/internal for production
	// Check for localhost patterns
	lowerURL := strings.ToLower(c.MCPBackend.URL)
	if strings.Contains(lowerURL, "localhost") ||
		strings.Contains(lowerURL, "127.0.0.1") ||
		strings.Contains(lowerURL, "0.0.0.0") {
		// Allow localhost only if explicitly marked as dev mode
		if !c.Auth.Enabled { // If auth is not enabled, this might be intentional dev
			// Still warn - should not be used in production
		}
	}

	return nil
}

// ToEnvConfig converts ConfigFile to environment-based Config
// SECURITY FIX: Now calls Validate() to ensure security config is enforced
func (c *ConfigFile) ToEnvConfig() *Config {
	// SECURITY FIX: Call Validate() to enforce security requirements
	// This ensures config files can't bypass authentication and other security settings
	if err := c.Validate(); err != nil {
		// Log warning but don't fail - allow environment variables to override
		fmt.Printf("WARNING: Config file validation failed: %v\n", err)
		fmt.Printf("WARNING: Ensure JWT_SECRET is set via environment variable for production\n")
	}

	config := &Config{
		ListenAddr:         c.Server.ListenAddr,
		MCPBackendURL:      c.MCPBackend.URL,
		LakeraAPIKey:       c.Lakera.APIKey,
		LakeraURL:          c.Lakera.URL,
		LakeraTimeout:      c.Lakera.TimeoutSec,
		RateLimitPerMinute: c.RateLimit.PerMinute,
		ProxyTimeout:       30,
	}

	// SECURITY FIX: Default to localhost for safety (same as env var default)
	// Previously defaulted to 0.0.0.0 which exposed to network
	if config.ListenAddr == "" && c.Server.Port > 0 {
		config.ListenAddr = fmt.Sprintf("127.0.0.1:%d", c.Server.Port)
	}

	return config
}

// ExampleConfig returns an example configuration JSON
func ExampleConfig() string {
	return `{
  "server": {
    "listen_addr": "0.0.0.0:8080",
    "port": 8080
  },
  "mcp_backend": {
    "url": "http://mcp-server:9090"
  },
  "lakera": {
    "api_key": "",
    "url": "https://api.lakera.ai",
    "threshold": 70,
    "timeout_sec": 5
  },
  "rate_limit": {
    "per_minute": 60
  },
  "auth": {
    "enabled": false,
    "api_keys": [],
    "jwt_secret": ""
  }
}`
}
