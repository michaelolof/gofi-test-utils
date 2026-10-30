// Package server builds the Gofi router for the service and owns its
// lifecycle.
package server

import (
	"context"
	"log/slog"

	"github.com/michaelolof/gofi"
	gofimw "github.com/michaelolof/gofi/middleware"

	"example.com/projectsvc/internal/http/deps"
	"example.com/projectsvc/internal/http/middleware"
	"example.com/projectsvc/internal/http/routes"
)

type Args struct {
	Addr        string
	CORSOrigins string
	Deps        deps.Deps
	Log         *slog.Logger
}

type Server struct {
	router gofi.Router
	addr   string
}

// New builds the router: dependencies, global middleware, error handler,
// routes and docs, in that order.
func New(args Args) (*Server, error) {
	cors, err := gofimw.CORS(gofimw.CORSConfig{
		AllowOrigins: args.CORSOrigins,
		AllowHeaders: "authorization, content-type",
	})
	if err != nil {
		return nil, err
	}

	r := gofi.NewRouter()
	deps.Register(r, args.Deps)

	r.Use(gofimw.Recover())
	r.Use(middleware.RequestID)
	r.Use(cors)
	r.UseErrorHandler(middleware.ErrorHandler(args.Log))

	routes.RegisterAPIRoutes(r, args.Deps)

	if err := gofi.ServeDocs(r, DocsOptions()); err != nil {
		return nil, err
	}
	return &Server{router: r, addr: args.Addr}, nil
}

// DocsOptions describes the API for the docs UI and the spec export tool.
func DocsOptions() gofi.DocsOptions {
	return gofi.DocsOptions{
		Info: gofi.DocsInfoOptions{Title: "Projects API", Version: "1.0.0"},
		Views: []gofi.DocsView{
			{RoutePrefix: "/docs", Template: gofi.StopLight()},
		},
	}
}

// Start serves until Shutdown is called.
func (s *Server) Start() error { return s.router.Listen(s.addr) }

// Shutdown stops accepting connections and waits for in-flight requests
// until ctx is done.
func (s *Server) Shutdown(ctx context.Context) error { return s.router.ShutdownWithContext(ctx) }

// Router returns the router, for tests (Router().Test) and the spec export tool.
func (s *Server) Router() gofi.Router { return s.router }
