package mcp

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aphilas/pgmcp/pkg/eval"
	"github.com/aphilas/pgmcp/pkg/jsonrpc"
	"github.com/aphilas/pgmcp/pkg/types"
	"github.com/google/jsonschema-go/jsonschema"
)

type Calculator struct {
	Tool         Tool
	InputSchema  *jsonschema.Resolved
	OutputSchema *jsonschema.Resolved
}

type CalculatorParams struct {
	Expression string `json:"expression" jsonschema:"The arithmetic expression to evaluate. Supported operations: +, -, *, /, parentheses."`
}

func NewCalculator() (*Calculator, error) {
	inputSchema, err := jsonschema.For[CalculatorParams](nil)
	if err != nil {
		return nil, fmt.Errorf("creating input schema: %w", err)
	}

	inputSchemaResolved, err := inputSchema.Resolve(nil)
	if err != nil {
		return nil, fmt.Errorf("resolving input schema: %w", err)
	}

	return &Calculator{
		Tool: Tool{
			Name:        "calculator",
			Title:       types.Ptr("Calculator"),
			Description: types.Ptr("A simple calculator that can perform basic arithmetic operations."),
			InputSchema: inputSchema,
		},
		InputSchema: inputSchemaResolved,
	}, nil
}

func (c *Calculator) Definition() Tool {
	return c.Tool
}

func (c *Calculator) Execute(params json.RawMessage) (*CallToolResult, *jsonrpc.Error) {
	var object map[string]any
	err := json.Unmarshal(params, &object)
	if err != nil {
		return NewErrorTextResult(fmt.Sprintf("Error parsing parameters: %s", err.Error())), nil
	}

	// Validate does NOT take a struct: See:
	// https://github.com/google/jsonschema-go/issues/23
	err = c.InputSchema.Validate(object)
	if err != nil {
		return NewErrorTextResult(fmt.Sprintf("Invalid parameters: %s", err.Error())), nil
	}

	var p CalculatorParams
	err = json.Unmarshal(params, &p)
	if err != nil {
		return NewErrorTextResult(fmt.Sprintf("Error parsing parameters: %s", err.Error())), nil
	}

	res, err := eval.Eval(p.Expression)
	if err != nil {
		// TODO: Verify not internal error
		return NewErrorTextResult(fmt.Sprintf("Error evaluating expression: %s", err.Error())), nil
	}

	return NewTextResult(strconv.Itoa(*res)), nil
}
