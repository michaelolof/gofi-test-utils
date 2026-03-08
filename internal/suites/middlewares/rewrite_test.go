package middleware

import (
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestRewriteConfig(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.Rewrite(middleware.RewriteConfig{
		Rules: map[string]string{
			"/old":      "/new",
			"/api/v1/*": "/api/v2$1",
		},
	}))

	// Register actual target routes to verify rewrite works properly
	app.Get("/new", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "Reached /new")
		},
	})

	// FastHTTP doesn't magically jump to the next route inside the router loop.
	// Since Gofi's tree router executes before middleware applies this URI change natively,
	// checking `c.Path()` directly inside exact match handlers verifies Rewrite

	app.Get("/api/v2/users", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "/api/v2/users")
		},
	})

	app.Get("/normal", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "/normal")
		},
	})

	// Test case 1: Exact
	resp1, _ := app.Inject(gofi.InjectOptions{Method: "GET", Path: "/old", Handler: &gofi.RouteOptions{Handler: func(c gofi.Context) error { return c.SendString(200, c.Path()) }}})
	if string(resp1.Body) != "/new" {
		t.Errorf("Expected rewritten path to be /new, got %s", string(resp1.Body))
	}

	// Test case 2: Wildcard
	resp2, _ := app.Inject(gofi.InjectOptions{Method: "GET", Path: "/api/v1/users", Handler: &gofi.RouteOptions{Handler: func(c gofi.Context) error { return c.SendString(200, c.Path()) }}})
	if string(resp2.Body) != "/api/v2/users" {
		t.Errorf("Expected rewritten path to be /api/v2/users, got %s", string(resp2.Body))
	}

	// Test case 3: Unchanged
	resp3, _ := app.Inject(gofi.InjectOptions{Method: "GET", Path: "/normal", Handler: &gofi.RouteOptions{Handler: func(c gofi.Context) error { return c.SendString(200, c.Path()) }}})
	if string(resp3.Body) != "/normal" {
		t.Errorf("Expected unchanged path /normal, got %s", string(resp3.Body))
	}
}
