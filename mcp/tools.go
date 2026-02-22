package mcp

import (
	"encoding/json"
	"log"

	"github.com/aphilas/pgmcp/pkg/jsonrpc"
	"github.com/google/jsonschema-go/jsonschema"
)

type Tooler interface {
	Definition() Tool
	Execute(params json.RawMessage) (*CallToolResult, *jsonrpc.Error)
}

// Tool defines a tool the client can call. Omitted: icons, annotations,
// execution, _meta.
type Tool struct {
	Name        string  `json:"name"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`

	// InputSchema and OutputSchema are JSON Schema objects defining the
	// expected input parameters and output of the tool, respectively. Type is
	// always "object".
	InputSchema  *jsonschema.Schema `json:"inputSchema"`
	OutputSchema *jsonschema.Schema `json:"outputSchema,omitempty"`
}

// ListToolsParams contains parameters for a tools/list request. We do NOT
// support pagination. Omitted: cursor.
type ListToolsParams struct{}

// ListToolsResult is the server's response to a tools/list request.
// Omitted: nextCursor.
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

// CallToolParams contains parameters for a tools/call request.
// Omitted: task.
type CallToolParams struct {
	// The name of the tool.
	Name string `json:"name"`

	// Arguments is a JSON object containing the arguments to pass to the tool.
	// Type: map[string]any.
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// TextContent represents text content returned by a tool call.
type TextContent struct {
	Type string `json:"type"` // always "text"
	Text string `json:"text"`
}

// CallToolResult is the server's response to a tools/call request.
type CallToolResult struct {
	// Content represents the unstructured result of the tool call.
	Content []TextContent `json:"content"`
	// StructuredContent is a JSON object containing the structured content
	// returned by the tool. Type: map[string]any.
	StructuredContent json.RawMessage `json:"structuredContent,omitempty"`
	// IsError indicates whether the tool call resulted in an error. Errors that
	// originate from the tool should be reported inside the result object.
	// Other exceptional conditions should MCP protocol-level errors.
	IsError *bool `json:"isError,omitempty"`
}

func NewTextResult(text string) *CallToolResult {
	return &CallToolResult{
		Content: []TextContent{
			TextContent{
				Type: "text",
				Text: text,
			},
		},
	}
}

func NewErrorTextResult(text string) *CallToolResult {
	isError := true
	return &CallToolResult{
		Content: []TextContent{
			{
				Type: "text",
				Text: text,
			},
		},
		IsError: &isError,
	}
}

// ListTools is called when the client sends the "tools/list" request. It
// returns a list of tools the server supports.
func (s Server) ListTools(p json.RawMessage) (json.RawMessage, *jsonrpc.Error) {
	tools := ListToolsResult{
		Tools: make([]Tool, 0, len(s.Tools)),
	}
	for _, tool := range s.Tools {
		tools.Tools = append(tools.Tools, tool.Definition())
	}

	resultBytes, err := json.Marshal(tools)
	if err != nil {
		return nil, &jsonrpc.Error{
			Code:    jsonrpc.CodeInternalError,
			Message: "Failed to marshal tools list",
		}
	}

	return resultBytes, nil
}

// CallTool is called when the client sends the "tools/call" request. It
// executes the specified tool with the provided arguments and returns the
// result.
func (s Server) CallTool(p json.RawMessage) (json.RawMessage, *jsonrpc.Error) {
	var params CallToolParams
	err := json.Unmarshal(p, &params)
	if err != nil {
		log.Printf("Failed to unmarshal CallTool params: %v", err)
		return nil, &jsonrpc.Error{
			Code:    jsonrpc.CodeInvalidParams,
			Message: "Invalid params",
		}
	}

	tool, ok := s.Tools[params.Name]
	if !ok {
		return nil, &jsonrpc.Error{
			Code:    jsonrpc.CodeInvalidParams,
			Message: "Tool not found",
		}
	}

	result, toolErr := tool.Execute(params.Arguments)
	if toolErr != nil {
		return nil, toolErr
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, &jsonrpc.Error{
			Code:    jsonrpc.CodeInternalError,
			Message: "Failed to marshal tool result",
		}
	}

	return resultBytes, nil
}
