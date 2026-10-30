package routes

import (
	"log/slog"
	"time"

	"github.com/michaelolof/gofi"

	"example.com/projectsvc/internal/http/deps"
	"example.com/projectsvc/internal/http/handlers/projects"
)

// RegisterRoutes registers every route. In a small service the middleware
// lives in this file and is mounted in one Use call.
func RegisterRoutes(r gofi.Router, d deps.Deps) {
	r.Use(
		requestLogger,
		requireJSON,
	)

	r.Post("/projects", projects.CreateProjectHandler)
	r.Get("/projects/:project_id", projects.GetProjectHandler)
}

func requestLogger(c gofi.Context) error {
	start := time.Now()
	err := c.Next()
	slog.Info("request", "method", c.Method(), "path", c.Path(), "failed", err != nil, "duration", time.Since(start))
	return err
}

func requireJSON(c gofi.Context) error {
	if c.Method() == "POST" && c.HeaderVal("content-type") != "application/json" {
		return gofi.NewHTTPError(415, "content-type must be application/json")
	}
	return c.Next()
}
