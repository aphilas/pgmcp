// Package jsonrpc is a subset of JSON-RPC for working with MCP.
package jsonrpc

import (
	"encoding/json"
	"log"
)

const Version = "2.0"

// Request represents a JSON-RPC request object.
// A Request is a Notification if the ID is omitted.
type Request struct {
	JSONRPC string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
	ID      *json.RawMessage `json:"id,omitempty"`
}

type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC response object.
// One of Result or Error must be provided.
type Response struct {
	JSONRPC string           `json:"jsonrpc"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *Error           `json:"error,omitempty"`
	ID      *json.RawMessage `json:"id"`
}

// Error represents the JSON-RPC error object.
type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Implement Error() string to satisfy error interface.
func (err Error) Error() string {
	return err.Message
}

// Pre-defined Error Codes
const (
	CodeParseError     = -32700 // Invalid JSON received by the server.
	CodeInvalidRequest = -32600 // The JSON sent is not a valid Request object.
	CodeMethodNotFound = -32601 // The method does not exist / is not available.
	CodeInvalidParams  = -32602 // Invalid method parameter(s).
	CodeInternalError  = -32603 // Internal JSON-RPC error.
)

// ServerError codes (-32000 to -32099) are reserved for implementation-defined
// server-errors.
const (
	CodeServerErrorMin = -32099
	CodeServerErrorMax = -32000
)

// NewRequest is a helper to create a new Request.
func NewRequest(method string, params any, id any) (*Request, error) {
	var p json.RawMessage
	if params != nil {
		var err error
		p, err = json.Marshal(params)
		if err != nil {
			return nil, err
		}
	}

	var i *json.RawMessage
	if id != nil {
		idBytes, err := json.Marshal(id)
		if err != nil {
			return nil, err
		}
		rawID := json.RawMessage(idBytes)
		i = &rawID
	}

	return &Request{
		JSONRPC: Version,
		Method:  method,
		Params:  p,
		ID:      i,
	}, nil
}

// NewResponse is a helper to create a success Response.
func NewResponse(id *json.RawMessage, result any) *Response {
	resBytes, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}

	return &Response{
		JSONRPC: Version,
		Result:  resBytes,
		ID:      id,
	}
}

// NewErrorResponse is a helper to create an error Response.
func NewErrorResponse(id *json.RawMessage, error *Error) *Response {
	return &Response{
		JSONRPC: Version,
		Error:   error,
		ID:      id,
	}
}

func JSONRawMessage(v any) json.RawMessage {
	buf, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Marshaling json: %v", err)
	}

	return buf
}
