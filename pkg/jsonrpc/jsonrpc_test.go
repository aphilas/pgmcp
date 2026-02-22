package jsonrpc

import (
	"encoding/json"
	"testing"
)

func TestNewRequest(t *testing.T) {
	t.Run("with int id", func(t *testing.T) {
		req, err := NewRequest("ping", nil, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.JSONRPC != Version {
			t.Errorf("JSONRPC = %q, want %q", req.JSONRPC, Version)
		}
		if req.Method != "ping" {
			t.Errorf("Method = %q, want %q", req.Method, "ping")
		}
		if string(req.ID) != "1" {
			t.Errorf("ID = %s, want 1", req.ID)
		}
	})

	t.Run("with string id", func(t *testing.T) {
		req, err := NewRequest("ping", nil, "abc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(req.ID) != `"abc"` {
			t.Errorf("ID = %s, want %q", req.ID, "abc")
		}
	})

	t.Run("with nil id", func(t *testing.T) {
		req, err := NewRequest("ping", nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.ID != nil {
			t.Errorf("ID = %s, want nil", req.ID)
		}
	})

	t.Run("with params", func(t *testing.T) {
		params := json.RawMessage(`{"key":"value"}`)
		req, err := NewRequest("echo", params, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(req.Params) != `{"key":"value"}` {
			t.Errorf("Params = %s, want %s", req.Params, `{"key":"value"}`)
		}
	})
}

func TestNewResponse(t *testing.T) {
	id := json.RawMessage(`1`)
	result := json.RawMessage(`{"ok":true}`)
	resp := NewResponse(id, result)

	if resp.JSONRPC != Version {
		t.Errorf("JSONRPC = %q, want %q", resp.JSONRPC, Version)
	}
	if string(resp.Result) != `{"ok":true}` {
		t.Errorf("Result = %s, want %s", resp.Result, `{"ok":true}`)
	}
	if resp.Error != nil {
		t.Errorf("Error = %v, want nil", resp.Error)
	}
	if string(resp.ID) != "1" {
		t.Errorf("ID = %s, want 1", resp.ID)
	}
}

func TestNewErrorResponse(t *testing.T) {
	id := json.RawMessage(`1`)
	rpcErr := &Error{Code: CodeMethodNotFound, Message: "Method not found"}
	resp := NewErrorResponse(id, rpcErr)

	if resp.JSONRPC != Version {
		t.Errorf("JSONRPC = %q, want %q", resp.JSONRPC, Version)
	}
	if resp.Result != nil {
		t.Errorf("Result = %v, want nil", resp.Result)
	}
	if resp.Error == nil {
		t.Fatal("Error is nil, want non-nil")
	}
	if resp.Error.Code != CodeMethodNotFound {
		t.Errorf("Error.Code = %d, want %d", resp.Error.Code, CodeMethodNotFound)
	}
}

func TestErrorInterface(t *testing.T) {
	err := Error{Code: CodeInternalError, Message: "something broke"}
	if err.Error() != "something broke" {
		t.Errorf("Error() = %q, want %q", err.Error(), "something broke")
	}
}

func TestEmptyResult(t *testing.T) {
	r := EmptyResult()
	if r == nil {
		t.Fatal("EmptyResult() returned nil")
	}
	if string(r) != "{}" {
		t.Errorf("EmptyResult() = %s, want {}", r)
	}
}

func TestJSONRawMessage(t *testing.T) {
	raw := JSONRawMessage(42)
	if string(raw) != "42" {
		t.Errorf("JSONRawMessage(42) = %s, want 42", raw)
	}

	raw = JSONRawMessage("hello")
	if string(raw) != `"hello"` {
		t.Errorf("JSONRawMessage(\"hello\") = %s, want %q", raw, "hello")
	}
}

func TestRequestJSON(t *testing.T) {
	req, _ := NewRequest("ping", json.RawMessage(`{"a":1}`), 1)
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Method != "ping" {
		t.Errorf("Method = %q, want %q", decoded.Method, "ping")
	}
	if string(decoded.ID) != "1" {
		t.Errorf("ID = %s, want 1", decoded.ID)
	}
}

func TestResponseJSON(t *testing.T) {
	resp := NewResponse(json.RawMessage(`"abc"`), json.RawMessage(`{"ok":true}`))
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if string(decoded.ID) != `"abc"` {
		t.Errorf("ID = %s, want %q", decoded.ID, "abc")
	}
	if decoded.Error != nil {
		t.Errorf("Error = %v, want nil", decoded.Error)
	}
}

func TestServerRegisterMethod(t *testing.T) {
	s := NewServer()
	called := false
	s.RegisterMethod("test", func(params json.RawMessage) (json.RawMessage, *Error) {
		called = true
		return EmptyResult(), nil
	})

	method, ok := s.methods["test"]
	if !ok {
		t.Fatal("method not registered")
	}
	method(nil)
	if !called {
		t.Error("method was not called")
	}
}
