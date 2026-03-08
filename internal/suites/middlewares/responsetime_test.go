package middleware

import (
	"strings"
	"testing"
	"time"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestResponseTimeConfig(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.ResponseTime(middleware.ResponseTimeConfig{
		Header: "X-My-Timer",
		Suffix: " ms",
	}))

	app.Get("/time", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			// Sleep briefly to accumulate some time
			time.Sleep(10 * time.Millisecond)
			return c.SendString(200, "hello")
		},
	})

	resp := app.Test("GET", "/time")

	header := resp.HeaderMap.Get("X-My-Timer")
	if header == "" {
		t.Fatalf("Expected X-My-Timer header to be present")
	}

	if !strings.HasSuffix(header, " ms") {
		t.Errorf("Expected header to end with ' ms', got '%s'", header)
	}
}
