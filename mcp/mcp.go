package mcp

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aphilas/pgmcp/pkg/jsonrpc"
)

const ProtocolVersion = "2025-11-25"

type Params = any
type Result = any
type Method = func(json.RawMessage) (json.RawMessage, error)

type Server struct {
	ProtocolVersion string
	ServerInfo      Implementation
}

func NewServer() *Server {
	s := Server{
		ServerInfo: Implementation{
			Name:    "pgmcp",
			Version: "0.0.1",
		},
		ProtocolVersion: ProtocolVersion,
	}

	return &s
}

func (s Server) Handle(r jsonrpc.Request) *jsonrpc.Response {
	switch r.Method {
	case "initialize":
		var p InitializeRequestParams

		_, jerr := s.initialize(p)
		if jerr != nil {
			return jsonrpc.NewErrorResponse(
				r.ID,
				jerr,
			)
		}

		return jsonrpc.NewResponse(r.ID, make(map[string]any))
	case "tools/list":
		return jsonrpc.NewResponse(r.ID, map[string]any{
			"tools": []string{},
		})
	default:
		return jsonrpc.NewErrorResponse(
			r.ID,
			&jsonrpc.Error{
				Code:    jsonrpc.CodeMethodNotFound,
				Message: fmt.Sprintf("Method %q not found", r.Method),
			},
		)
	}
}

// Implementation describes the MCP implementation. Omitted: icons.
type Implementation struct {
	Name        string  `json:"name"`
	Title       *string `json:"title,omitempty"`
	Version     string  `json:"version"`
	Description *string `json:"description,omitempty"`
	WebsiteURL  *string `json:"websiteUrl,omitempty"`
}

// ServerCapabilities defines capabilities a server may support. Omitted:
// experimental, logging, completions, prompts, resources, tasks.
type ServerCapabilities struct {
	Tools *ToolsCapability `json:"tools,omitempty"`
}

// ToolsCapability indicates if the server offers tools to call.
type ToolsCapability struct {
	ListChanged *bool `json:"listChanged,omitempty"`
}

// InitializeRequestParams contains parameters for an initialize request.
// Omitted: capabilities.
type InitializeRequestParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	ClientInfo      Implementation `json:"clientInfo"`
}

// InitializeRequest is sent from the client to the server when it first
// connects
type InitializeRequest struct {
	jsonrpc.Request
	Method string                  `json:"method"` // "initialize"
	Params InitializeRequestParams `json:"params"`
}

// InitializeResult is the server's response to an initialize request
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      Implementation     `json:"serverInfo"`
	Instructions    *string            `json:"instructions,omitempty"`
}

// InitializedNotification is sent from the client to the server after
// initialization has finished
type InitializedNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"` // "notifications/initialized"
}

type UnsupportedProtocolErrorData struct {
	Supported []string `json:"supported"`
	Requested string   `json:"requested"`
}

func (s Server) initialize(params InitializeRequestParams) (*InitializeResult, *jsonrpc.Error) {
	log.Printf("Received initialization request from %v", params.ClientInfo.Name)

	if params.ProtocolVersion != s.ProtocolVersion {
		return nil, &jsonrpc.Error{
			Code:    jsonrpc.CodeInvalidParams,
			Message: "Unsupported protocol version",
			Data: jsonrpc.JSONRawMessage(UnsupportedProtocolErrorData{
				Supported: []string{s.ProtocolVersion},
				Requested: params.ProtocolVersion,
			}),
		}
	}

	return &InitializeResult{
		ProtocolVersion: s.ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{},
		},
		ServerInfo: s.ServerInfo,
	}, nil
}
