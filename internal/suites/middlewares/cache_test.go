package middleware

import (
	"strconv"
	"testing"
	"time"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestCacheMiddleware(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.Cache(middleware.CacheConfig{
		Expiration:  100 * time.Millisecond,
		CacheHeader: true,
	}))

	counter := 0

	app.Get("/cache", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			counter++
			return c.SendString(200, "Count: "+strconv.Itoa(counter))
		},
	})

	// Request 1: cache miss, runs handler
	resp1 := mustTest(t, app, "GET", "/cache")
	if string(resp1.Body) != "Count: 1" {
		t.Errorf("Expected Count: 1, got %s", string(resp1.Body))
	}

	// Request 2: cache hit, skips handler
	resp2 := mustTest(t, app, "GET", "/cache")
	if string(resp2.Body) != "Count: 1" {
		t.Errorf("Expected Count: 1 (cached), got %s", string(resp2.Body))
	}

	if resp2.HeaderMap.Get("Cache-Control") == "" {
		t.Errorf("Expected Cache-Control header to be set")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Request 3: cache miss, runs handler again
	resp3 := mustTest(t, app, "GET", "/cache")
	if string(resp3.Body) != "Count: 2" {
		t.Errorf("Expected Count: 2 (expired), got %s", string(resp3.Body))
	}
}

func TestCacheNotSupportedMethod(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.Cache(middleware.CacheConfig{
		Expiration: 1 * time.Minute,
	}))

	counter := 0

	app.Post("/nocache", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			counter++
			return c.SendString(200, "Count: "+strconv.Itoa(counter))
		},
	})

	resp1, _ := app.Inject(gofi.InjectOptions{Method: "POST", Path: "/nocache", Handler: &gofi.RouteOptions{Handler: func(c gofi.Context) error { counter++; return c.SendString(200, "Count: "+strconv.Itoa(counter)) }}})
	resp2, _ := app.Inject(gofi.InjectOptions{Method: "POST", Path: "/nocache", Handler: &gofi.RouteOptions{Handler: func(c gofi.Context) error { counter++; return c.SendString(200, "Count: "+strconv.Itoa(counter)) }}})

	if string(resp1.Body) == string(resp2.Body) {
		t.Errorf("Expected POST not to be cached. Body1: %s, Body2: %s", string(resp1.Body), string(resp2.Body))
	}
}
