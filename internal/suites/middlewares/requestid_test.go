package middleware

import (
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestRequestID(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.RequestID())

	app.Get("/id-test", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			// Extract ID from headers to prove it was set
			id := c.Writer().Header().Get("X-Request-Id")
			return c.SendString(200, id)
		},
	})

	// Test case 1: No ID provided in request
	resp := mustTest(t, app, "GET", "/id-test")
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	headerID := resp.HeaderMap.Get("X-Request-Id")
	if headerID == "" {
		t.Errorf("Expected X-Request-Id in response headers, got empty")
	}

	bodyID := string(resp.Body)
	if bodyID != headerID {
		t.Errorf("Expected body to match header. Body: %s, Header: %s", bodyID, headerID)
	}

	// Test case 2: ID provided in request (Should be reused)
	injectReq, _ := app.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/id-test",
		Headers: map[string]string{
			"X-Request-Id": "my-custom-id-123",
		},
		Handler: &gofi.RouteOptions{
			Handler: func(c gofi.Context) error {
				id := c.Writer().Header().Get("X-Request-Id")
				return c.SendString(200, id)
			},
		},
	})

	if injectReq.HeaderMap.Get("X-Request-Id") != "my-custom-id-123" {
		t.Errorf("Expected provided ID to be echoed, got %s", injectReq.HeaderMap.Get("X-Request-Id"))
	}
}

func TestRequestIDCustomConfig(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.RequestID(middleware.RequestIDConfig{
		Header: "X-Trace-Id",
		Generator: func() string {
			return "static-trace"
		},
	}))

	app.Get("/trace", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "hello")
		},
	})

	resp := mustTest(t, app, "GET", "/trace")

	if resp.HeaderMap.Get("X-Trace-Id") != "static-trace" {
		t.Errorf("Expected custom header to be set with custom generator value. Got: %s", resp.HeaderMap.Get("X-Trace-Id"))
	}
}
