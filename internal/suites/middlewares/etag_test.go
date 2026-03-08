package middleware

import (
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestETagMiddleware(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.ETag())

	app.Get("/etag", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "hello etag")
		},
	})

	// Test case 1: Initial request (gets ETag)
	resp1 := app.Test("GET", "/etag")
	if resp1.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp1.StatusCode)
	}

	etag := resp1.HeaderMap.Get("ETag")
	if etag == "" {
		t.Fatalf("Expected ETag header, got empty")
	}

	// Test case 2: Request with correct If-None-Match
	resp2, _ := app.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/etag",
		Headers: map[string]string{
			"If-None-Match": etag,
		},
		Handler: &gofi.RouteOptions{
			Handler: func(c gofi.Context) error {
				return c.SendString(200, "hello etag")
			},
		},
	})

	if resp2.StatusCode != 304 {
		t.Errorf("Expected 304 Not Modified, got %d", resp2.StatusCode)
	}

	if len(resp2.Body) > 0 {
		t.Errorf("Expected empty body for 304, got %s", string(resp2.Body))
	}

	// Test case 3: Request with incorrect If-None-Match
	resp3, _ := app.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/etag",
		Headers: map[string]string{
			"If-None-Match": "\"different-etag\"",
		},
		Handler: &gofi.RouteOptions{
			Handler: func(c gofi.Context) error {
				return c.SendString(200, "hello etag")
			},
		},
	})

	if resp3.StatusCode != 200 {
		t.Errorf("Expected 200 OK for mismatched ETag, got %d", resp3.StatusCode)
	}
}

func TestETagWeak(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.ETag(middleware.ETagConfig{Weak: true}))

	app.Get("/weak", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "weak etag response")
		},
	})

	resp := app.Test("GET", "/weak")

	etag := resp.HeaderMap.Get("ETag")
	// Weak Etag should start with W/
	if len(etag) < 2 || etag[:2] != "W/" {
		t.Errorf("Expected weak ETag to start with W/, got %s", etag)
	}
}
