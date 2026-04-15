package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

// JSONRPCError represents a JSON-RPC 2.0 error object
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
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

// ToolCallParams represents the params for tools/call method
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolsListParams represents the params for tools/list method
type ToolsListParams struct {
	Limit  int    `json:"limit,omitempty"`
	Cursor string `json:"cursor,omitempty"`
}

// ParsedRequest holds extracted request data for processing
type ParsedRequest struct {
	Method    string
	ToolName  string
	Arguments json.RawMessage
	Params    any
	ID        any
	IsBatch   bool
	BatchReqs []ParsedRequest
}

// JSONRPCParseError codes
const (
	ParseErrorCode     = -32700
	InvalidRequestCode = -32600
	MethodNotFoundCode = -32601
	InvalidParamsCode  = -32602
	InternalErrorCode  = -32603
)

// ParseJSONRPC parses a JSON-RPC 2.0 message and returns a ParsedRequest or error
func ParseJSONRPC(data []byte) (*ParsedRequest, error) {
	// Try to parse as request first
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err == nil {
		parsed, err := validateAndExtractRequest(req)
		if err == nil {
			return parsed, nil
		}
	}

	// Try batch request
	if strings.HasPrefix(strings.TrimSpace(string(data)), "[") {
		var batchReqs []JSONRPCRequest
		if err := json.Unmarshal(data, &batchReqs); err == nil {
			if len(batchReqs) == 0 {
				return nil, newJSONRPCError(InvalidRequestCode, "Empty batch request")
			}
			return parseBatchRequest(batchReqs)
		}
	}

	// Try notification (no id)
	var notification JSONRPCRequest
	if err := json.Unmarshal(data, &notification); err == nil {
		if notification.JSONRPC == "2.0" && notification.ID == nil {
			parsed, err := validateAndExtractRequest(notification)
			if err == nil {
				parsed.ID = nil // explicit nil for notifications
				return parsed, nil
			}
		}
	}

	return nil, newJSONRPCError(ParseErrorCode, "Invalid JSON-RPC 2.0 message")
}

// parseBatchRequest handles batch requests
func parseBatchRequest(reqs []JSONRPCRequest) (*ParsedRequest, error) {
	batch := make([]ParsedRequest, 0, len(reqs))

	for i, req := range reqs {
		parsed, err := validateAndExtractRequest(req)
		if err != nil {
			// Per JSON-RPC 2.0 spec, continue processing other requests
			log.Printf("Warning: Skipping invalid batch request at index %d: %v", i, err)
			continue
		}
		batch = append(batch, *parsed)
	}

	if len(batch) == 0 {
		return nil, newJSONRPCError(InvalidRequestCode, "No valid requests in batch")
	}

	return &ParsedRequest{
		IsBatch:   true,
		BatchReqs: batch,
	}, nil
}

// validateAndExtractRequest validates and extracts data from a JSON-RPC request
func validateAndExtractRequest(req JSONRPCRequest) (*ParsedRequest, error) {
	// Validate jsonrpc version
	if req.JSONRPC != "2.0" {
		return nil, newJSONRPCError(InvalidRequestCode, "jsonrpc field must be '2.0'")
	}

	// Validate method presence
	if req.Method == "" {
		return nil, newJSONRPCError(InvalidRequestCode, "method field is required")
	}

	parsed := &ParsedRequest{
		Method: req.Method,
		ID:     req.ID,
	}

	// Extract tool-specific data based on method
	switch req.Method {
	case "tools/call":
		params, err := extractToolCallParams(req.Params)
		if err != nil {
			return nil, err
		}
		parsed.ToolName = params.Name
		parsed.Arguments = params.Arguments
		parsed.Params = params

	case "tools/list":
		if len(req.Params) > 0 {
			params, err := extractToolsListParams(req.Params)
			if err != nil {
				return nil, err
			}
			parsed.Params = params
		}

	case "resources/list", "resources/read", "prompts/list", "prompts/get":
		// Other supported methods - just pass through params
		if len(req.Params) > 0 {
			var params any
			if err := json.Unmarshal(req.Params, &params); err == nil {
				parsed.Params = params
			}
		}
	}

	return parsed, nil
}

// extractToolCallParams extracts tool call parameters
func extractToolCallParams(params json.RawMessage) (*ToolCallParams, error) {
	if len(params) == 0 {
		return nil, newJSONRPCError(InvalidParamsCode, "tools/call requires parameters")
	}

	var toolParams ToolCallParams
	if err := json.Unmarshal(params, &toolParams); err != nil {
		return nil, newJSONRPCError(InvalidParamsCode, "Invalid tool call parameters")
	}

	if toolParams.Name == "" {
		return nil, newJSONRPCError(InvalidParamsCode, "tool name is required")
	}

	return &toolParams, nil
}

// extractToolsListParams extracts tools list parameters
func extractToolsListParams(params json.RawMessage) (*ToolsListParams, error) {
	var listParams ToolsListParams
	if err := json.Unmarshal(params, &listParams); err != nil {
		// Default to empty params
		return &listParams, nil
	}
	return &listParams, nil
}

// newJSONRPCError creates a new JSONRPCError with given code and message
func newJSONRPCError(code int, message string) error {
	return errors.New(fmt.Sprintf("JSONRPCError %d: %s", code, message))
}

// CreateErrorResponse creates a JSON-RPC 2.0 error response
func CreateErrorResponse(id any, code int, message string) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
}

// CreateSuccessResponse creates a JSON-RPC 2.0 success response
func CreateSuccessResponse(id any, result any) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
}

// SerializeResponse serializes a JSON-RPC response to bytes
func SerializeResponse(resp JSONRPCResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// SerializeBatchResponse serializes a batch of responses
func SerializeBatchResponse(resps []JSONRPCResponse) ([]byte, error) {
	// Filter out notifications (those with null id)
	filtered := make([]JSONRPCResponse, 0)
	for _, r := range resps {
		if r.ID != nil {
			filtered = append(filtered, r)
		}
	}
	if len(filtered) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(filtered)
}

// GetToolInfo extracts tool name and arguments for semantic analysis
func GetToolInfo(parsed *ParsedRequest) (toolName string, args string, ok bool) {
	if parsed.ToolName == "" {
		return "", "", false
	}

	args = string(parsed.Arguments)
	return parsed.ToolName, args, true
}
