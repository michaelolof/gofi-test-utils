package middleware

import (
	"log"
	"strings"
	"testing"
	"time"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestLogger(t *testing.T) {
	app := gofi.NewRouter()

	var out stringWriter
	logger := log.New(&out, "", 0)

	app.Use(middleware.Logger(middleware.LoggerConfig{
		Output: logger,
	}))

	app.Get("/log-test", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			time.Sleep(10 * time.Millisecond) // Simulate work
			return c.SendString(200, "hello logger")
		},
	})

	resp := mustTest(t, app, "GET", "/log-test")

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	output := out.String()

	if !strings.Contains(output, "200") {
		t.Errorf("Expected 200 in output, got: %s", output)
	}

	if !strings.Contains(output, "GET") {
		t.Errorf("Expected GET in output, got: %s", output)
	}

	if !strings.Contains(output, "/log-test") {
		t.Errorf("Expected /log-test in output, got: %s", output)
	}

	// The logged IP is the socket peer; forwarding headers need a TrustPolicy.
	if !strings.Contains(output, "0.0.0.0") {
		t.Errorf("Expected the peer IP in output, got: %s", output)
	}
}
