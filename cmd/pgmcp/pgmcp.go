package main

import (
	"log"
	"os"

	"github.com/aphilas/pgmcp/mcp"
	"github.com/aphilas/pgmcp/pkg/jsonrpc"
)

func main() {
	transport := jsonrpc.NewStdioServer(os.Stdin, os.Stdout, os.Stderr)
	log.Printf("starting stdio jsonrpc server\n")

	server, err := mcp.NewServer(transport)
	if err != nil {
		log.Fatalf("creating server: %v\n", err)
	}

	log.Printf("starting server with protocol version %s\n", server.ProtocolVersion)
	server.Transport.Serve()
}
