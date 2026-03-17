package middleware

import (
	"net/http"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestEarlyDataMiddleware(t *testing.T) {
	app := gofi.NewRouter()

	var capturedErr error
	app.UseErrorHandler(func(err error, c gofi.Context) {
		capturedErr = err
		_ = c.SendString(http.StatusTooEarly, "Too Early")
	})

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
	resp1 := mustTest(t, app, "GET", "/safe")
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
	if string(resp4.Body) != "Too Early" {
		t.Errorf("Expected 'Too Early' body, got %q", string(resp4.Body))
	}
	if capturedErr == nil {
		t.Fatal("Expected router error handler to receive early data rejection")
	}
	if capturedErr.Error() != "too early" {
		t.Fatalf("Expected 'too early' error, got %v", capturedErr)
	}
}

func TestEarlyDataCanHandleLocally(t *testing.T) {
	app := gofi.NewRouter()

	globalHandlerCalled := false
	app.UseErrorHandler(func(err error, c gofi.Context) {
		globalHandlerCalled = true
		_ = c.SendString(500, err.Error())
	})

	app.Use(middleware.EarlyData(middleware.EarlyDataConfig{
		ErrorHandler: func(c gofi.Context) error {
			return c.SendString(460, "handled locally")
		},
	}))

	resp, _ := app.Inject(gofi.InjectOptions{
		Method: http.MethodPost,
		Path:   "/unsafe",
		Headers: map[string]string{
			"Early-Data": "1",
		},
		Handler: &gofi.RouteOptions{Handler: func(c gofi.Context) error { return c.SendString(200, "unsafe") }},
	})

	if resp.StatusCode != 460 {
		t.Fatalf("Expected local early data handler status 460, got %d", resp.StatusCode)
	}
	if string(resp.Body) != "handled locally" {
		t.Fatalf("Expected local early data handler body, got %q", string(resp.Body))
	}
	if globalHandlerCalled {
		t.Fatal("Expected router error handler not to be called when EarlyData handles locally")
	}
}
