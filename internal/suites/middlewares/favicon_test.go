package middleware

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
	"github.com/valyala/fasthttp"
)

func TestFaviconDefault(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.Favicon())

	app.Get("/favicon.ico", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(404, "") },
	})

	resp := mustTest(t, app, "GET", "/favicon.ico")

	if resp.StatusCode != fasthttp.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	if resp.HeaderMap.Get(fasthttp.HeaderContentType) != "image/x-icon" {
		t.Errorf("Expected image/x-icon content type, got '%s'", resp.HeaderMap.Get(fasthttp.HeaderContentType))
	}

	if resp.HeaderMap.Get(fasthttp.HeaderCacheControl) == "" {
		t.Errorf("Expected non-empty Cache-Control header")
	}

	// Small built-in placeholder is > 0 length
	if len(resp.Body) == 0 {
		t.Errorf("Expected non-empty body payload")
	}

	// Ensure other routes are unaffected
	respOther := mustTest(t, app, "GET", "/")
	if respOther.StatusCode != fasthttp.StatusNotFound {
		t.Errorf("Expected 404 for unhandled route, got %d", respOther.StatusCode)
	}
}

func TestFaviconCustom(t *testing.T) {
	dir := t.TempDir()
	iconPath := filepath.Join(dir, "icon.ico")
	customData := []byte("custom_icon_data")

	err := os.WriteFile(iconPath, customData, 0644)
	if err != nil {
		t.Fatal(err)
	}

	app := gofi.NewRouter()

	app.Use(middleware.Favicon(middleware.FaviconConfig{
		File:         iconPath,
		URL:          "/my-favicon.ico",
		CacheControl: "no-cache",
	}))

	app.Get("/my-favicon.ico", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(404, "") },
	})

	resp := mustTest(t, app, "GET", "/my-favicon.ico")

	if resp.StatusCode != fasthttp.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	if string(resp.Body) != "custom_icon_data" {
		t.Errorf("Expected body 'custom_icon_data', got '%s'", string(resp.Body))
	}

	if resp.HeaderMap.Get(fasthttp.HeaderCacheControl) != "no-cache" {
		t.Errorf("Expected Cache-Control 'no-cache', got '%s'", resp.HeaderMap.Get("Cache-Control"))
	}
}
