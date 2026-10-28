package middleware

import (
	"net/http"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

var testCSRFKey = []byte("0123456789abcdef0123456789abcdef")

func testSessionID(c gofi.Context) string {
	return c.Request().Header.Get("X-Session")
}

func mustCSRF(t *testing.T, cfg middleware.CSRFConfig) gofi.MiddlewareFunc {
	t.Helper()
	if cfg.SigningKey == nil {
		cfg.SigningKey = testCSRFKey
	}
	if cfg.SessionID == nil {
		cfg.SessionID = testSessionID
	}
	mw, err := middleware.CSRF(cfg)
	if err != nil {
		t.Fatalf("CSRF config rejected: %v", err)
	}
	return mw
}

func csrfCookieFrom(t *testing.T, resp *gofi.InjectResponse) *http.Cookie {
	t.Helper()
	for _, c := range resp.Cookies() {
		if c.Name == "csrf_" && c.Value != "" {
			return c
		}
	}
	t.Fatal("Expected csrf_ cookie to be set")
	return nil
}

func TestCSRF_RequiresKeyAndSession(t *testing.T) {
	if _, err := middleware.CSRF(middleware.CSRFConfig{SessionID: testSessionID}); err == nil {
		t.Error("Expected an error without a signing key")
	}
	if _, err := middleware.CSRF(middleware.CSRFConfig{SigningKey: []byte("short"), SessionID: testSessionID}); err == nil {
		t.Error("Expected an error for a signing key under 32 bytes")
	}
	if _, err := middleware.CSRF(middleware.CSRFConfig{SigningKey: testCSRFKey}); err == nil {
		t.Error("Expected an error without a SessionID function")
	}
}

func TestCSRF_GenerationAndValidation(t *testing.T) {
	app := gofi.NewRouter()

	var capturedErr error
	app.UseErrorHandler(func(err error, c gofi.Context) {
		capturedErr = err
		_ = c.SendString(http.StatusForbidden, "Invalid CSRF Token")
	})

	app.Use(mustCSRF(t, middleware.CSRFConfig{}))

	app.Get("/safe", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "safe ok")
		},
	})

	unsafe := &gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "unsafe ok") },
	}

	// 1. Send GET request (Safe Method) to generate the token
	resp1 := mustTest(t, app, "GET", "/safe")
	if resp1.StatusCode != 200 {
		t.Errorf("Expected 200 GET, got %d", resp1.StatusCode)
	}
	csrfCookie := csrfCookieFrom(t, resp1)

	// 2. Send POST request WITHOUT token (Should fail 403)
	resp2, _ := app.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/unsafe",
		Cookies: []http.Cookie{*csrfCookie},
		Handler: unsafe,
	})
	if resp2.StatusCode != 403 {
		t.Errorf("Expected 403 Forbidden for missing CSRF header, got %d", resp2.StatusCode)
	}
	if string(resp2.Body) != "Invalid CSRF Token" {
		t.Errorf("Expected invalid token body, got %q", string(resp2.Body))
	}
	if capturedErr == nil || capturedErr.Error() != "invalid CSRF token" {
		t.Fatalf("Expected 'invalid CSRF token' error, got %v", capturedErr)
	}

	// 3. Send POST request WITH correct token (Should pass 200)
	resp3, _ := app.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/unsafe",
		Cookies: []http.Cookie{*csrfCookie},
		Headers: map[string]string{"X-CSRF-Token": csrfCookie.Value},
		Handler: unsafe,
	})
	if resp3.StatusCode != 200 {
		t.Errorf("Expected 200 OK for valid CSRF header, got %d. Body: %s", resp3.StatusCode, string(resp3.Body))
	}
}

func TestCSRF_CustomExtractor(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(mustCSRF(t, middleware.CSRFConfig{
		KeyLookup: "query:my_token",
		ErrorHandler: func(c gofi.Context) error {
			return c.SendString(401, "Custom Error") // Send 401 instead of 403
		},
	}))
	handler := &gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "ok") },
	}

	issued, _ := app.Inject(gofi.InjectOptions{Method: "GET", Path: "/", Handler: handler})
	csrfCookie := csrfCookieFrom(t, issued)

	// Send POST request failing validation to check custom error
	resp, _ := app.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/",
		Cookies: []http.Cookie{*csrfCookie},
		Query:   map[string]string{"my_token": "invalid-token"},
		Handler: handler,
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
		Cookies: []http.Cookie{*csrfCookie},
		Query:   map[string]string{"my_token": csrfCookie.Value},
		Handler: handler,
	})
	if resp2.StatusCode != 200 {
		t.Errorf("Expected 200 status code for valid query token, got %d", resp2.StatusCode)
	}
}
