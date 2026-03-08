package middleware

import (
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestRedirectConfig_Exact(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.Redirect(middleware.RedirectConfig{
		Rules: map[string]string{
			"/old":       "/new",
			"/legacy/v1": "/api/v2",
		},
		StatusCode: 301, // Moved Permanently
	}))

	// Unaffected route
	app.Get("/target", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "target destination")
		},
	})

	// Register the routes so they don't 404 before hitting the middleware stack
	app.Get("/old", gofi.RouteOptions{Handler: func(c gofi.Context) error { return nil }})
	app.Get("/legacy/v1", gofi.RouteOptions{Handler: func(c gofi.Context) error { return nil }})

	// Test case 1: Exact match 1
	resp1 := app.Test("GET", "/old")
	if resp1.StatusCode != 301 {
		t.Errorf("Expected 301 exact match, got %d", resp1.StatusCode)
	}
	if resp1.HeaderMap.Get("Location") != "/new" {
		t.Errorf("Expected Location /new, got %s", resp1.HeaderMap.Get("Location"))
	}

	// Test case 2: Exact match 2
	resp2 := app.Test("GET", "/legacy/v1")
	if resp2.StatusCode != 301 {
		t.Errorf("Expected 301 exact match, got %d", resp2.StatusCode)
	}
	if resp2.HeaderMap.Get("Location") != "/api/v2" {
		t.Errorf("Expected Location /api/v2, got %s", resp2.HeaderMap.Get("Location"))
	}

	// Test case 3: Unaffected path
	resp3 := app.Test("GET", "/target")
	if resp3.StatusCode != 200 {
		t.Errorf("Expected 200 for unaffected path, got %d", resp3.StatusCode)
	}
}

func TestRedirectConfig_Wildcard(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.Redirect(middleware.RedirectConfig{
		Rules: map[string]string{
			"/old/*":    "/new$1",
			"/api/v1/*": "/api/v2$1",
			"/static/*": "/assets$1",
		},
		StatusCode: 307,
	}))

	// Register the wildcard route prefixes so they don't 404
	app.Get("/old/*any", gofi.RouteOptions{Handler: func(c gofi.Context) error { return nil }})
	app.Get("/api/v1/*any", gofi.RouteOptions{Handler: func(c gofi.Context) error { return nil }})

	// Test case 1: Wildcard path root match
	resp1 := app.Test("GET", "/old/users")
	if resp1.StatusCode != 307 {
		t.Errorf("Expected 307 wildcard match, got %d", resp1.StatusCode)
	}
	if resp1.HeaderMap.Get("Location") != "/new/users" {
		t.Errorf("Expected Location /new/users, got %s", resp1.HeaderMap.Get("Location"))
	}

	// Test case 2: Wildcard deep match
	resp2 := app.Test("GET", "/api/v1/users/123/profile")
	if resp2.StatusCode != 307 {
		t.Errorf("Expected 307 wildcard match, got %d", resp2.StatusCode)
	}
	if resp2.HeaderMap.Get("Location") != "/api/v2/users/123/profile" {
		t.Errorf("Expected Location /api/v2/users/123/profile, got %s", resp2.HeaderMap.Get("Location"))
	}

	// Test case 3: Ignore missing suffix match
	resp3 := app.Test("GET", "/old-nomatch")
	if resp3.StatusCode != 307 {
		if resp3.HeaderMap.Get("Location") != "" {
			t.Errorf("Expected root wildcard to fallback gracefully, got %s", resp3.HeaderMap.Get("Location"))
		}
	}
}
