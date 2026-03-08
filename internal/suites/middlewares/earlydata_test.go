package middleware

import (
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestEarlyDataMiddleware(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.EarlyData())

	app.Get("/safe", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "safe")
		},
	})

	app.Post("/unsafe", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "unsafe")
		},
	})

	// Test case 1: Safe request without Early-Data header
	resp1 := app.Test("GET", "/safe")
	if resp1.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp1.StatusCode)
	}

	// Test case 2: Unsafe request without Early-Data header
	resp2, _ := app.Inject(gofi.InjectOptions{Method: "POST", Path: "/unsafe", Handler: &gofi.RouteOptions{Handler: func(c gofi.Context) error { return c.SendString(200, "unsafe") }}})
	if resp2.StatusCode != 200 {
		t.Errorf("Expected 200 for normal POST, got %d", resp2.StatusCode)
	}

	// Test case 3: Safe request WITH Early-Data header (Should be allowed by default)
	resp3, _ := app.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/safe",
		Headers: map[string]string{
			"Early-Data": "1",
		},
		Handler: &gofi.RouteOptions{Handler: func(c gofi.Context) error { return c.SendString(200, "safe") }},
	})
	if resp3.StatusCode != 200 {
		t.Errorf("Expected 200 for early data GET, got %d", resp3.StatusCode)
	}

	// Test case 4: Unsafe request WITH Early-Data header (Should be rejected)
	resp4, _ := app.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/unsafe",
		Headers: map[string]string{
			"Early-Data": "1",
		},
		Handler: &gofi.RouteOptions{Handler: func(c gofi.Context) error { return c.SendString(200, "unsafe") }},
	})
	if resp4.StatusCode != 425 {
		t.Errorf("Expected 425 Too Early for early data POST, got %d", resp4.StatusCode)
	}
}
