package framework

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// HTTPClient is a simple HTTP client for testing MCP proxy
type HTTPClient struct {
	BaseURL    string
	Timeout    time.Duration
	AuthHeader string
}

// NewHTTPClient creates a new HTTP client for testing
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		Timeout: 30 * time.Second,
	}
}

// SetAuth sets the authorization header
func (c *HTTPClient) SetAuth(token string) {
	c.AuthHeader = token
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      any             `json:"id,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      any           `json:"id,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SendJSONRPC sends a JSON-RPC request and returns the response
func (c *HTTPClient) SendJSONRPC(ctx context.Context, method string, params any) (*JSONRPCResponse, error) {
	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		ID:      1,
	}

	if params != nil {
		paramsBytes, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		reqBody.Params = paramsBytes
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.AuthHeader != "" {
		req.Header.Set("Authorization", c.AuthHeader)
	}

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var jsonResp JSONRPCResponse
	if err := json.Unmarshal(respBytes, &jsonResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &jsonResp, nil
}

// SendRawJSON sends raw JSON and returns the response
func (c *HTTPClient) SendRawJSON(ctx context.Context, jsonBody string) (int, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader([]byte(jsonBody)))
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.AuthHeader != "" {
		req.Header.Set("Authorization", c.AuthHeader)
	}

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read response: %w", err)
	}

	return resp.StatusCode, string(body), nil
}

// GetHealth checks the health endpoint
func (c *HTTPClient) GetHealth(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/health", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetMetrics retrieves metrics from the proxy
func (c *HTTPClient) GetMetrics(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/metrics", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// WaitForService waits for a service to be available
func WaitForService(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(url + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("service at %s not available after %v", url, timeout)
}

// Retry executes a function with retry logic
func Retry(attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("after %d attempts: %w", attempts, err)
}

// ToolCallParams represents parameters for tools/call
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// NewToolCallParams creates a new tool call params
func NewToolCallParams(name string, args map[string]interface{}) ToolCallParams {
	return ToolCallParams{
		Name:      name,
		Arguments: args,
	}
}

// GetEnv returns environment variable or default
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
