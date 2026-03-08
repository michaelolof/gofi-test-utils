package middleware

import (
	"strings"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestCompress_Gzip(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.Compress())

	app.Get("/", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			str := strings.Repeat("Hello World! ", 50)
			return c.SendString(200, str)
		},
	})

	// Test case 1: Request with gzip accepted
	resp, _ := app.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/",
		Headers: map[string]string{
			"Accept-Encoding": "gzip",
		},
		Handler: &gofi.RouteOptions{
			Handler: func(c gofi.Context) error {
				str := strings.Repeat("Hello World! ", 50)
				return c.SendString(200, str)
			},
		},
	})

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	if resp.HeaderMap.Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding: gzip, got %s", resp.HeaderMap.Get("Content-Encoding"))
	}

	uncompressedLen := len(strings.Repeat("Hello World! ", 50))
	if len(resp.Body) >= uncompressedLen {
		t.Errorf("Expected response body to be compressed (smaller than %d bytes), got %d bytes", uncompressedLen, len(resp.Body))
	}

	// Test case 2: Request without gzip (should remain uncompressed)
	resp2, _ := app.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/",
		Handler: &gofi.RouteOptions{
			Handler: func(c gofi.Context) error {
				str := strings.Repeat("Hello World! ", 50)
				return c.SendString(200, str)
			},
		},
	})

	if resp2.HeaderMap.Get("Content-Encoding") != "" {
		t.Errorf("Did not expect Content-Encoding without Accept-Encoding")
	}

	if len(resp2.Body) != uncompressedLen {
		t.Errorf("Expected uncompressed body length %d, got %d", uncompressedLen, len(resp2.Body))
	}
}

func TestCompress_Brotli(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.Compress(middleware.CompressConfig{
		Level: 1, // Fast compression
	}))

	resp, _ := app.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/",
		Headers: map[string]string{
			"Accept-Encoding": "br, gzip, deflate",
		},
		Handler: &gofi.RouteOptions{
			Handler: func(c gofi.Context) error {
				str := strings.Repeat("Hello brotli ", 50)
				return c.SendString(200, str)
			},
		},
	})

	if resp.HeaderMap.Get("Content-Encoding") != "br" {
		t.Errorf("Expected Content-Encoding: br (since br is generally preferred when both offered via fasthttp), got %s", resp.HeaderMap.Get("Content-Encoding"))
	}
}
