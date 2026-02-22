package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/aphilas/pgmcp/pkg/jsonrpc"
)

func main() {
	server := jsonrpc.NewStdioServer(os.Stdin, os.Stdout, os.Stderr)
	log.Printf("starting stdio jsonrpc server\n")

	server.RegisterMethod("ping", func(params json.RawMessage) (json.RawMessage, *jsonrpc.Error) {
		return jsonrpc.EmptyResult(), nil
	})

	server.Serve()
}
