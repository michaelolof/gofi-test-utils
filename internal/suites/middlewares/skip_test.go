package middleware

import (
	"strings"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestSkipConfig(t *testing.T) {
	app := gofi.NewRouter()

	wrappedMiddleware := func(c gofi.Context) error {
		c.Writer().Header().Set("X-Skipped", "false")
		return c.Next()
	}

	app.Use(middleware.Skip(middleware.SkipConfig{
		SkipFilter: func(c gofi.Context) bool {
			return strings.HasPrefix(c.Path(), "/skip")
		},
		Handler: wrappedMiddleware,
	}))

	app.Get("/skip/me", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "skipped route")
		},
	})

	app.Get("/apply/me", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "applied route")
		},
	})

	// Test case 1: Should Skip
	resp1 := mustTest(t, app, "GET", "/skip/me")
	if resp1.HeaderMap.Get("X-Skipped") == "false" {
		t.Errorf("Expected middleware to be skipped (no X-Skipped header)")
	}

	// Test case 2: Should Not Skip
	resp2 := mustTest(t, app, "GET", "/apply/me")
	if resp2.HeaderMap.Get("X-Skipped") != "false" {
		t.Errorf("Expected middleware to execute (X-Skipped header should be false)")
	}
}
