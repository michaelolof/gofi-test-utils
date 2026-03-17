package middleware

import (
	"errors"
	"testing"
	"time"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestTimeoutConfig(t *testing.T) {
	app := gofi.NewRouter()

	var capturedErr error
	app.UseErrorHandler(func(err error, c gofi.Context) {
		capturedErr = err
		_ = c.SendString(408, "Request Timeout")
	})

	app.Use(middleware.Timeout(middleware.TimeoutConfig{
		Timeout: 50 * time.Millisecond,
	}))

	app.Get("/fast", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "fast")
		},
	})

	app.Get("/slow", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			time.Sleep(100 * time.Millisecond)
			return c.SendString(200, "slow")
		},
	})

	// Test case 1: Fast handler completes
	resp1 := mustTest(t, app, "GET", "/fast")
	if resp1.StatusCode != 200 {
		t.Errorf("Expected 200 for fast route, got %d", resp1.StatusCode)
	}

	// Test case 2: Slow handler gets timed out
	resp2, _ := app.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/slow",
		Handler: &gofi.RouteOptions{
			Handler: func(c gofi.Context) error {
				time.Sleep(100 * time.Millisecond)
				return c.SendString(200, "slow")
			},
		},
	})

	if resp2.StatusCode != 408 {
		t.Errorf("Expected 408 Request Timeout for slow route, got %d", resp2.StatusCode)
	}

	if string(resp2.Body) != "Request Timeout" {
		t.Errorf("Expected 'Request Timeout' body, got '%s'", string(resp2.Body))
	}

	if capturedErr == nil {
		t.Fatal("Expected router error handler to receive timeout error")
	}
	if capturedErr.Error() != "request timeout" {
		t.Fatalf("Expected 'request timeout' error, got %v", capturedErr)
	}
}

func TestTimeoutCanHandleLocally(t *testing.T) {
	app := gofi.NewRouter()

	globalHandlerCalled := false
	app.UseErrorHandler(func(err error, c gofi.Context) {
		globalHandlerCalled = true
		_ = c.SendString(500, err.Error())
	})

	app.Use(middleware.Timeout(middleware.TimeoutConfig{
		Timeout: 20 * time.Millisecond,
		ErrorHandler: func(c gofi.Context) error {
			return c.SendString(418, "timeout handled locally")
		},
	}))

	app.Get("/slow", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			time.Sleep(60 * time.Millisecond)
			return errors.New("should not complete")
		},
	})

	resp := mustTest(t, app, "GET", "/slow")

	if resp.StatusCode != 418 {
		t.Fatalf("Expected local timeout handler status 418, got %d", resp.StatusCode)
	}
	if string(resp.Body) != "timeout handled locally" {
		t.Fatalf("Expected local timeout handler body, got %q", string(resp.Body))
	}
	if globalHandlerCalled {
		t.Fatal("Expected router error handler not to be called when timeout middleware handles locally")
	}
}
