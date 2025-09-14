package mcp

import (
	"encoding/json"
	"fmt"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// JSONRPCNotification represents a JSON-RPC 2.0 notification
type JSONRPCNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Standard JSON-RPC error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// MCP-specific error codes
const (
	ResourceNotFound = -32002
	ToolNotFound     = -32001
)

// NewJSONRPCResponse creates a successful JSON-RPC response
func NewJSONRPCResponse(id interface{}, result interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
}

// NewJSONRPCError creates an error JSON-RPC response
func NewJSONRPCError(id interface{}, code int, message string, data interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	}
}

// NewJSONRPCNotification creates a JSON-RPC notification
func NewJSONRPCNotification(method string, params interface{}) *JSONRPCNotification {
	return &JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

// ParseJSONRPCRequest parses a JSON-RPC request from bytes
func ParseJSONRPCRequest(data []byte) (*JSONRPCRequest, error) {
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse JSON-RPC request: %w", err)
	}

	// Validate required fields
	if req.JSONRPC != "2.0" {
		return nil, fmt.Errorf("invalid JSON-RPC version: %s", req.JSONRPC)
	}

	if req.Method == "" {
		return nil, fmt.Errorf("missing method field")
	}

	if req.ID == nil {
		return nil, fmt.Errorf("missing id field")
	}

	return &req, nil
}

// Marshal converts a JSON-RPC response to bytes
func (r *JSONRPCResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Marshal converts a JSON-RPC notification to bytes
func (n *JSONRPCNotification) Marshal() ([]byte, error) {
	return json.Marshal(n)
}

// IsNotification checks if a request is a notification (no ID)
func (r *JSONRPCRequest) IsNotification() bool {
	return r.ID == nil
}

// GetStringParam extracts a string parameter from the request
func (r *JSONRPCRequest) GetStringParam(name string) (string, error) {
	params, ok := r.Params.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("params is not an object")
	}

	value, exists := params[name]
	if !exists {
		return "", fmt.Errorf("parameter '%s' not found", name)
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("parameter '%s' is not a string", name)
	}

	return str, nil
}

// GetIntParam extracts an integer parameter from the request
func (r *JSONRPCRequest) GetIntParam(name string) (int, error) {
	params, ok := r.Params.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("params is not an object")
	}

	value, exists := params[name]
	if !exists {
		return 0, fmt.Errorf("parameter '%s' not found", name)
	}

	// Handle both int and float64 (from JSON)
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("parameter '%s' is not a number", name)
	}
}

// GetOptionalStringParam extracts an optional string parameter
func (r *JSONRPCRequest) GetOptionalStringParam(name string) (string, bool) {
	params, ok := r.Params.(map[string]interface{})
	if !ok {
		return "", false
	}

	value, exists := params[name]
	if !exists {
		return "", false
	}

	str, ok := value.(string)
	if !ok {
		return "", false
	}

	return str, true
}

// GetParams returns the raw params as a map
func (r *JSONRPCRequest) GetParams() (map[string]interface{}, error) {
	params, ok := r.Params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("params is not an object")
	}
	return params, nil
}