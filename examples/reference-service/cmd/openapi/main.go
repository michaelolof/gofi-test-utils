// Command openapi writes the service's OpenAPI document to stdout without
// starting a server or touching any external system.
package main

import (
	"encoding/json"
	"log/slog"
	"os"

	"github.com/michaelolof/gofi"

	"example.com/projectsvc/internal/http/deps"
	"example.com/projectsvc/internal/http/server"
)

func main() {
	srv, err := server.New(server.Args{CORSOrigins: "*", Deps: deps.AppDeps{}, Log: slog.Default()})
	if err != nil {
		slog.Error("server setup failed", "error", err)
		os.Exit(1)
	}
	spec := gofi.OpenAPISpec(srv.Router(), server.DocsOptions())

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(spec); err != nil {
		slog.Error("encode failed", "error", err)
		os.Exit(1)
	}
}
