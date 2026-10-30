// Package routes maps URL patterns to handlers and route-level middleware.
package routes

import (
	"github.com/michaelolof/gofi"

	"example.com/projectsvc/internal/http/deps"
	"example.com/projectsvc/internal/http/handlers/apikeys"
	"example.com/projectsvc/internal/http/handlers/health"
	"example.com/projectsvc/internal/http/handlers/projects"
	"example.com/projectsvc/internal/http/middleware"
)

// RegisterAPIRoutes registers the public health route and the product API
// under /api/v1, where every route requires a bearer token.
func RegisterAPIRoutes(r gofi.Router, d deps.Deps) {
	r.Get("/health", health.GetHealthHandler)

	r.Route("/api/v1", func(r gofi.Router) {
		r.Use(middleware.Authenticate(d.Authenticator()))

		r.Post("/projects", projects.CreateProjectHandler)
		r.Get("/projects/:project_id", projects.GetProjectHandler)

		r.Post("/projects/:project_id/api-keys", apikeys.CreateAPIKeyHandler)
	})
}
