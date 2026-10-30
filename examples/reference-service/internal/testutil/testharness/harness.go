// Package testharness builds the real server with in-memory dependencies so
// handler tests run through the same routes, middleware and error handler
// as production.
package testharness

import (
	"io"
	"log/slog"
	"testing"

	"github.com/michaelolof/gofi"

	"example.com/projectsvc/internal/http/deps"
	"example.com/projectsvc/internal/http/server"
	"example.com/projectsvc/internal/memstore"
)

// Harness is one isolated copy of the service.
type Harness struct {
	Deps   deps.AppDeps
	router gofi.Router
}

// New returns a harness with empty stores and two known tokens.
func New(t *testing.T) *Harness {
	t.Helper()
	d := deps.AppDeps{
		ProjectStore: memstore.NewProjects(),
		APIKeyStore:  &memstore.APIKeys{},
		Authn:        memstore.Tokens{"token-alice": "alice", "token-bob": "bob"},
	}
	srv, err := server.New(server.Args{
		CORSOrigins: "http://localhost:3000",
		Deps:        d,
		Log:         slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatalf("testharness: %v", err)
	}
	return &Harness{Deps: d, router: srv.Router()}
}

// Test sends a request through the full router.
func (h *Harness) Test(opts gofi.TestOptions) (*gofi.InjectResponse, error) {
	return h.router.Test(opts)
}

// AuthHeaders returns request headers for the given token.
func AuthHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}
}
