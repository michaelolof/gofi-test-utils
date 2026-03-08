package middleware

import (
	"testing"
	"time"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestTimeoutConfig(t *testing.T) {
	app := gofi.NewRouter()

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
	resp1 := app.Test("GET", "/fast")
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
}
