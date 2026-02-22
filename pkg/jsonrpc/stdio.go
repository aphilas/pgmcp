package jsonrpc

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
)

type StdioServer struct {
	// in is the input stream for reading JSON-RPC requests.
	in io.Reader
	// out is the output stream for writing JSON-RPC responses.
	out io.Writer
	// err is the error stream for logging errors.
	err io.Writer

	methods map[string]Method
}

func NewStdioServer(in io.Reader, out io.Writer, err io.Writer) *StdioServer {
	return &StdioServer{
		in:      in,
		out:     out,
		err:     err,
		methods: make(map[string]Method),
	}
}

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

func (s *StdioServer) Serve() {
	logger := log.New(s.err, "jsonrpc: ", log.LstdFlags)

	scanner := bufio.NewScanner(s.in)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			logger.Printf("Failed to parse request: %v", err)
			
			resp := NewErrorResponse(nil, &Error{
				Code:    CodeInvalidRequest,
				Message: "Invalid request: invalid JSON",
			})
			respBytes, marshalErr := json.Marshal(resp)
			if marshalErr != nil {
				logger.Printf("Failed to marshal error response: %v", marshalErr)
				continue
			}
			
			s.out.Write(respBytes)
			s.out.Write([]byte("\n"))
			continue
		}

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
