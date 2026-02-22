package main

import (
	"log"
	"os"

	"github.com/aphilas/pgmcp/pkg/jsonrpc"
)

func main() {
	server := jsonrpc.NewStdioServer(os.Stdin, os.Stdout, os.Stderr)
	log.Printf("starting stdio jsonrpc server\n")

	server.RegisterMethod("ping", func(params map[string]any) (map[string]any, *jsonrpc.Error) {
		return jsonrpc.EmptyResult(), nil
	})

	server.Serve()
}
