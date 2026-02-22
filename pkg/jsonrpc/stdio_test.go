package jsonrpc

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func newTestServer(input string) (*StdioServer, *bytes.Buffer) {
	in := strings.NewReader(input)
	out := &bytes.Buffer{}
	s := NewStdioServer(in, out, io.Discard)
	return s, out
}

func TestHandleMethodFound(t *testing.T) {
	s, _ := newTestServer("")
	s.RegisterMethod("ping", func(params json.RawMessage) (json.RawMessage, *Error) {
		return EmptyResult(), nil
	})

	req := &Request{
		JSONRPC: Version,
		Method:  "ping",
		ID:      json.RawMessage(`1`),
	}
	resp := s.Handle(req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleMethodNotFound(t *testing.T) {
	s, _ := newTestServer("")

	req := &Request{
		JSONRPC: Version,
		Method:  "nonexistent",
		ID:      json.RawMessage(`1`),
	}
	resp := s.Handle(req)

	if resp.Error == nil {
		t.Fatal("expected error, got nil")
	}
	if resp.Error.Code != CodeMethodNotFound {
		t.Errorf("Error.Code = %d, want %d", resp.Error.Code, CodeMethodNotFound)
	}
}

func TestHandleMethodReturnsError(t *testing.T) {
	s, _ := newTestServer("")
	s.RegisterMethod("fail", func(params json.RawMessage) (json.RawMessage, *Error) {
		return nil, &Error{Code: CodeInternalError, Message: "boom"}
	})

	req := &Request{
		JSONRPC: Version,
		Method:  "fail",
		ID:      json.RawMessage(`1`),
	}
	resp := s.Handle(req)

	if resp.Error == nil {
		t.Fatal("expected error, got nil")
	}
	if resp.Error.Code != CodeInternalError {
		t.Errorf("Error.Code = %d, want %d", resp.Error.Code, CodeInternalError)
	}
	if resp.Error.Message != "boom" {
		t.Errorf("Error.Message = %q, want %q", resp.Error.Message, "boom")
	}
}

func TestHandlePassesParams(t *testing.T) {
	s, _ := newTestServer("")
	s.RegisterMethod("echo", func(params json.RawMessage) (json.RawMessage, *Error) {
		return params, nil
	})

	req := &Request{
		JSONRPC: Version,
		Method:  "echo",
		Params:  json.RawMessage(`{"msg":"hello"}`),
		ID:      json.RawMessage(`1`),
	}
	resp := s.Handle(req)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if string(resp.Result) != `{"msg":"hello"}` {
		t.Errorf("Result = %s, want %s", resp.Result, `{"msg":"hello"}`)
	}
}

func TestServeRoundTrip(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"ping","params":{}}` + "\n"
	s, out := newTestServer(input)
	s.RegisterMethod("ping", func(params json.RawMessage) (json.RawMessage, *Error) {
		return EmptyResult(), nil
	})

	s.Serve()

	var resp Response
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal response: %v", err)
	}
	if resp.JSONRPC != Version {
		t.Errorf("JSONRPC = %q, want %q", resp.JSONRPC, Version)
	}
	if string(resp.ID) != "1" {
		t.Errorf("ID = %s, want 1", resp.ID)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
}

func TestServeMultipleRequests(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"ping"}` + "\n"
	s, out := newTestServer(input)
	s.RegisterMethod("ping", func(params json.RawMessage) (json.RawMessage, *Error) {
		return EmptyResult(), nil
	})

	s.Serve()

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d responses, want 2", len(lines))
	}

	for i, line := range lines {
		var resp Response
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Fatalf("response %d: Unmarshal: %v", i, err)
		}
		if resp.Error != nil {
			t.Errorf("response %d: unexpected error: %v", i, resp.Error)
		}
	}
}

func TestServeInvalidJSON(t *testing.T) {
	input := "not json\n" + `{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n"
	errBuf := &bytes.Buffer{}
	in := strings.NewReader(input)
	out := &bytes.Buffer{}
	s := NewStdioServer(in, out, errBuf)
	s.RegisterMethod("ping", func(params json.RawMessage) (json.RawMessage, *Error) {
		return EmptyResult(), nil
	})

	s.Serve()

	// Invalid JSON returns an error response, then the valid request succeeds.
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d responses, want 2", len(lines))
	}

	// First response should be an InvalidRequest error.
	var errResp Response
	if err := json.Unmarshal([]byte(lines[0]), &errResp); err != nil {
		t.Fatalf("Unmarshal error response: %v", err)
	}
	if errResp.Error == nil {
		t.Fatal("expected error in first response")
	}
	if errResp.Error.Code != CodeInvalidRequest {
		t.Errorf("Error.Code = %d, want %d", errResp.Error.Code, CodeInvalidRequest)
	}

	// Second response should be a success.
	var okResp Response
	if err := json.Unmarshal([]byte(lines[1]), &okResp); err != nil {
		t.Fatalf("Unmarshal ok response: %v", err)
	}
	if okResp.Error != nil {
		t.Errorf("unexpected error in second response: %v", okResp.Error)
	}

	// Error should still be logged to stderr.
	if !strings.Contains(errBuf.String(), "Failed to parse request") {
		t.Errorf("stderr = %q, want parse error log", errBuf.String())
	}
}

func TestServeMethodNotFoundResponse(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"unknown"}` + "\n"
	s, out := newTestServer(input)

	s.Serve()

	var resp Response
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error response")
	}
	if resp.Error.Code != CodeMethodNotFound {
		t.Errorf("Error.Code = %d, want %d", resp.Error.Code, CodeMethodNotFound)
	}
}
