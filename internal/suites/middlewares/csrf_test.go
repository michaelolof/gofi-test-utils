package middleware

import (
	"errors"
	"net/http"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestCSRF_GenerationAndValidation(t *testing.T) {
	app := gofi.NewRouter()

	var capturedErr error
	app.UseErrorHandler(func(err error, c gofi.Context) {
		capturedErr = err
		_ = c.SendString(http.StatusForbidden, "Invalid CSRF Token")
	})

	app.Use(middleware.CSRF())

	app.Get("/safe", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "safe ok")
		},
	})

	app.Post("/unsafe", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "unsafe ok")
		},
	})

	// 1. Send GET request (Safe Method) to generate the token
	resp1 := mustTest(t, app, "GET", "/safe")
	if resp1.StatusCode != 200 {
		t.Errorf("Expected 200 GET, got %d", resp1.StatusCode)
	}

	// Extract the cookie manually
	cookies := resp1.Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "csrf_" {
			csrfCookie = c
			break
		}
	}

	if csrfCookie == nil || csrfCookie.Value == "" {
		t.Fatalf("Expected csrf_ cookie to be set")
	}

	token := csrfCookie.Value

	// 2. Send POST request WITHOUT token (Should fail 403)
	resp2, _ := app.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/unsafe",
		Cookies: []http.Cookie{*csrfCookie}, // Has cookie, but no header token
		Handler: &gofi.RouteOptions{
			Handler: func(c gofi.Context) error { return c.SendString(200, "unsafe ok") },
		},
	})

	if resp2.StatusCode != 403 {
		t.Errorf("Expected 403 Forbidden for missing CSRF header, got %d", resp2.StatusCode)
	}
	if string(resp2.Body) != "Invalid CSRF Token" {
		t.Errorf("Expected invalid token body, got %q", string(resp2.Body))
	}
	if capturedErr == nil {
		t.Fatal("Expected router error handler to receive CSRF validation error")
	}
	if !errors.Is(capturedErr, capturedErr) || capturedErr.Error() != "invalid CSRF token" {
		t.Fatalf("Expected 'invalid CSRF token' error, got %v", capturedErr)
	}

	// 3. Send POST request WITH correct token (Should pass 200)
	resp3, _ := app.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/unsafe",
		Cookies: []http.Cookie{*csrfCookie},
		Headers: map[string]string{
			"X-CSRF-Token": token,
		},
		Handler: &gofi.RouteOptions{
			Handler: func(c gofi.Context) error { return c.SendString(200, "unsafe ok") },
		},
	})

	if resp3.StatusCode != 200 {
		t.Errorf("Expected 200 OK for valid CSRF header, got %d. Body: %s", resp3.StatusCode, string(resp3.Body))
	}
}

func TestCSRF_CustomExtractor(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.CSRF(middleware.CSRFConfig{
		KeyLookup: "query:my_token",
		ErrorHandler: func(c gofi.Context) error {
			return c.SendString(401, "Custom Error") // Send 401 instead of 403
		},
	}))

	// Send POST request failing validation to check custom error
	resp, _ := app.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/",
		Cookies: []http.Cookie{{Name: "csrf_", Value: "my-valid-token"}},
		Query: map[string]string{
			"my_token": "invalid-token",
		},
		Handler: &gofi.RouteOptions{
			Handler: func(c gofi.Context) error { return c.SendString(200, "ok") },
		},
	})

	if resp.StatusCode != 401 {
		t.Errorf("Expected custom 401 status code, got %d", resp.StatusCode)
	}
	if string(resp.Body) != "Custom Error" {
		t.Errorf("Expected custom error string, got %s", string(resp.Body))
	}

	// Send POST request passing validation using Query lookup
	resp2, _ := app.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/",
		Cookies: []http.Cookie{{Name: "csrf_", Value: "my-valid-token"}},
		Query: map[string]string{
			"my_token": "my-valid-token",
		},
		Handler: &gofi.RouteOptions{
			Handler: func(c gofi.Context) error { return c.SendString(200, "ok") },
		},
	})

	if resp2.StatusCode != 200 {
		t.Errorf("Expected 200 status code for valid query token, got %d", resp2.StatusCode)
	}
}
