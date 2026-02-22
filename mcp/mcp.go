package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/aphilas/pgmcp/pkg/jsonrpc"
	"github.com/aphilas/pgmcp/pkg/types"
)

const ProtocolVersion = "2025-11-25"

type Server struct {
	ProtocolVersion string
	ServerInfo      Implementation
	Transport       Transport

	Tools map[string]Tooler
}

type Transport interface {
	RegisterMethod(name string, method jsonrpc.Method)
	Serve()
}

func NewServer(transport Transport) (*Server, error) {
	calculatorTool, err := NewCalculator()
	if err != nil {
		return nil, fmt.Errorf("creating calculator tool: %w", err)
	}

	s := Server{
		ServerInfo: Implementation{
			Name:    "pgmcp",
			Version: "0.0.1",
		},
		ProtocolVersion: ProtocolVersion,
		Tools: map[string]Tooler{
			"calculator": calculatorTool,
		},
		Transport: transport,
	}

	methods := map[string]jsonrpc.Method{
		"initialize":                s.Initialize,
		"notifications/initialized": s.NotificationsInitialized,
		"tools/list":                s.ListTools,
		"tools/call":                s.CallTool,
	}

	for name, method := range methods {
		s.Transport.RegisterMethod(name, method)
	}

	return &s, nil
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

// InitializeParams contains parameters for an initialize request.
// Omitted: capabilities.
type InitializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	ClientInfo      Implementation `json:"clientInfo"`
}

// InitializeResult is the server's response to an initialize request
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      Implementation     `json:"serverInfo"`
	Instructions    *string            `json:"instructions,omitempty"`
}

type UnsupportedProtocolErrorData struct {
	Supported []string `json:"supported"`
	Requested string   `json:"requested"`
}

func (s Server) Initialize(p json.RawMessage) (json.RawMessage, *jsonrpc.Error) {
	var params InitializeParams
	err := json.Unmarshal(p, &params)
	if err != nil {
		return nil, &jsonrpc.Error{
			Message: "Invalid params",
			Code:    jsonrpc.CodeInvalidParams,
		}
	}

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

	return types.NewRawJSON(InitializeResult{
		ProtocolVersion: s.ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{},
		},
		ServerInfo: s.ServerInfo,
	}), nil
}

// NotificationsInitialized is called when the client sends the
// "notifications/initialized" notification.
func (s Server) NotificationsInitialized(p json.RawMessage) (json.RawMessage, *jsonrpc.Error) {
	// TODO: Handle initialized notification
	return nil, nil
}
