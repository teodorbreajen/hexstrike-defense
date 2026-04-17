package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// LakeraConfig holds configuration for the Lakera client
type LakeraConfig struct {
	APIKey    string
	Threshold int
	Timeout   time.Duration
	BaseURL   string
}

// LakeraChecker defines the interface for Lakera security checks
// This allows for mock implementations in tests
type LakeraChecker interface {
	CheckToolCall(ctx context.Context, toolName string, arguments string) (bool, int, string, error)
	HealthCheck(ctx context.Context) error
	GetConfig() *LakeraConfig
}

// LakeraClient handles communication with the Lakera Guard API
type LakeraClient struct {
	config     *LakeraConfig
	httpClient *http.Client
}

// Ensure LakeraClient implements LakeraChecker
var _ LakeraChecker = (*LakeraClient)(nil)

// LakeraResponse represents the response from Lakera Guard API
type LakeraResponse struct {
	Score     int       `json:"score"`
	Verdict   string    `json:"verdict"`
	Reasons   []string  `json:"reasons,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// NewLakeraClient creates a new Lakera client with the given configuration
func NewLakeraClient(config *LakeraConfig) *LakeraClient {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.lakera.ai"
	}
	if config.Threshold == 0 {
		config.Threshold = 70 // Default threshold
	}

	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	return &LakeraClient{
		config:     config,
		httpClient: client,
	}
}

// CheckToolCall sends a tool call to Lakera for semantic analysis
// Returns: isAllowed (bool), score (int), reason (string), error
func (c *LakeraClient) CheckToolCall(ctx context.Context, toolName string, arguments string) (bool, int, string, error) {
	// SECURITY FIX: Fail-closed when no API key is configured
	// Previously allowed all requests silently, now require explicit configuration
	if c.config.APIKey == "" {
		// Log at ERROR level to ensure operators notice
		log.Println("[Lakera] CRITICAL: No API key configured - security is DISABLED")
		log.Println("[Lakera] WARNING: Set LAKERA_API_KEY environment variable for production")
		// Return error so the proxy can apply fail mode based on its configuration
		return c.handleError(fmt.Errorf("Lakera API key not configured"))
	}

	// Build the request payload
	payload := map[string]interface{}{
		"tool_name": toolName,
		"arguments": arguments,
		"context":   "mcp_tool_call",
		"mode":      "strict",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[Lakera] Failed to marshal request: %v", err)
		return c.handleError(err)
	}

	// Create the request
	url := fmt.Sprintf("%s/v1/guard/check", c.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		log.Printf("[Lakera] Failed to create request: %v", err)
		return c.handleError(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("User-Agent", "MCP-Policy-Proxy/1.0")

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[Lakera] Request failed: %v", err)
		return c.handleError(err)
	}
	defer resp.Body.Close()

	// FIX #5: Limit response body to 1MB to prevent memory exhaustion
	limitedReader := io.LimitReader(resp.Body, 1*1024*1024)
	respBody, err := io.ReadAll(limitedReader)
	if err != nil {
		// Check if it was truncated due to size limit
		if err == io.ErrUnexpectedEOF {
			log.Printf("[Lakera] Response body exceeded 1MB limit - treating as error")
			return c.handleError(fmt.Errorf("response body exceeded 1MB limit"))
		}
		log.Printf("[Lakera] Failed to read response body: %v", err)
		return c.handleError(err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		// SECURITY FIX: Don't log full response body - may contain sensitive info
		// Only log truncated info for debugging
		truncatedBody := string(respBody)
		if len(truncatedBody) > 200 {
			truncatedBody = truncatedBody[:200] + "..."
		}
		log.Printf("[Lakera] API returned status %d: %s", resp.StatusCode, truncatedBody)
		return c.handleError(fmt.Errorf("Lakera API returned error status %d", resp.StatusCode))
	}

	// Parse the response
	var lakeraResp LakeraResponse
	if err := json.Unmarshal(respBody, &lakeraResp); err != nil {
		log.Printf("[Lakera] Failed to parse response: %v", err)
		return c.handleError(err)
	}

	// Determine if the request should be allowed
	allowed := lakeraResp.Score < c.config.Threshold

	if !allowed {
		reason := "Security threshold exceeded"
		if len(lakeraResp.Reasons) > 0 {
			reason = lakeraResp.Reasons[0]
		}
		log.Printf("[Lakera] Blocked tool '%s' - score: %d, threshold: %d, reason: %s",
			toolName, lakeraResp.Score, c.config.Threshold, reason)
	} else {
		log.Printf("[Lakera] Allowed tool '%s' - score: %d, threshold: %d",
			toolName, lakeraResp.Score, c.config.Threshold)
	}

	return allowed, lakeraResp.Score, reasonFromResponse(lakeraResp), nil
}

// handleError handles errors from Lakera API - returns error for proxy to apply fail mode
// SECURITY FIX: Sanitized error messages - never expose internal details
func (c *LakeraClient) handleError(err error) (bool, int, string, error) {
	// Log full error details server-side only
	log.Printf("[Lakera] Error - returning to proxy for fail mode handling: %v", err)
	// SECURITY FIX: Return generic error message to client - don't leak internal details
	// The error is still returned for logging by the proxy
	return false, 0, "Lakera security service unavailable", err
}

// reasonFromResponse extracts a human-readable reason from the response
func reasonFromResponse(resp LakeraResponse) string {
	if len(resp.Reasons) > 0 {
		return resp.Reasons[0]
	}
	return fmt.Sprintf("risk_score=%d", resp.Score)
}

// HealthCheck checks if the Lakera API is reachable
func (c *LakeraClient) HealthCheck(ctx context.Context) error {
	if c.config.APIKey == "" {
		return fmt.Errorf("Lakera API key not configured")
	}

	url := fmt.Sprintf("%s/v1/health", c.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// GetConfig returns the current Lakera configuration
func (c *LakeraClient) GetConfig() *LakeraConfig {
	return c.config
}
