package jsonrpc

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
)

// StdioServer implements a JSON-RPC server that communicates over standard
// input and output.
type StdioServer struct {
	// in is the input stream for reading JSON-RPC requests.
	in io.Reader
	// out is the output stream for writing JSON-RPC responses.
	out io.Writer
	// err is the error stream for logging errors.
	err io.Writer

	methods map[string]Method
}

// NewStdioServer creates a new StdioServer with the given input, output, and
// error streams.
func NewStdioServer(in io.Reader, out io.Writer, err io.Writer) *StdioServer {
	return &StdioServer{
		in:      in,
		out:     out,
		err:     err,
		methods: make(map[string]Method),
	}
}

// RehgisterMethod registers a JSON-RPC method with the server.
func (s *StdioServer) RegisterMethod(name string, method Method) {
	s.methods[name] = method
}

// Handle processes a JSON-RPC request and returns a response.
func (s *StdioServer) Handle(req *Request) *Response {
	method, ok := s.methods[req.Method]
	if !ok {
		return NewErrorResponse(req.ID,
			&Error{
				Code:    CodeMethodNotFound,
				Message: "Method not found",
			},
		)
	}

	result, err := method(req.Params)
	if err != nil {
		return NewErrorResponse(req.ID, err)
	}

	return NewResponse(req.ID, result)
}

// Serve starts the server and listens for incoming JSON-RPC requests on the
// input stream. It processes each request and writes the corresponding response
// to the output stream.
func (s *StdioServer) Serve() {
	logger := log.New(s.err, "jsonrpc: ", log.LstdFlags)

	writeResponse := func(resp *Response) {
		respBytes, err := json.Marshal(resp)
		if err != nil {
			logger.Printf("Failed to marshal response: %v", err)
			return
		}

		s.out.Write(respBytes)
		s.out.Write([]byte("\n"))
	}

	scanner := bufio.NewScanner(s.in)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			logger.Printf("Failed to parse request: %v", err)
			writeResponse(NewErrorResponse(nil, &Error{
				Code:    CodeParseError,
				Message: "Invalid request: invalid JSON",
			}))
			continue
		}

		// JSONRPC MUST be version 2.0
		if req.JSONRPC != Version {
			logger.Printf("Invalid JSON-RPC version: %s", req.JSONRPC)
			writeResponse(NewErrorResponse(req.ID, &Error{
				Code:    CodeInvalidRequest,
				Message: "Invalid request: unsupported JSON-RPC version",
			}))
			continue
		}

		// TODO: ID MUST be a string or integer

		resp := s.Handle(&req)
		respBytes, err := json.Marshal(resp)
		if err != nil {
			logger.Printf("Failed to marshal response: %v", err)
			continue
		}

		// Write response to the output stream
		s.out.Write(respBytes)
		s.out.Write([]byte("\n"))
	}
}
