package middleware

import (
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestHelmet_Default(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.Helmet())

	app.Get("/", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "hello")
		},
	})

	resp := app.Test("GET", "/")

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	headers := map[string]string{
		"X-XSS-Protection":                  "0",
		"X-Content-Type-Options":            "nosniff",
		"X-Frame-Options":                   "SAMEORIGIN",
		"Referrer-Policy":                   "no-referrer",
		"Cross-Origin-Embedder-Policy":      "require-corp",
		"Cross-Origin-Opener-Policy":        "same-origin",
		"Cross-Origin-Resource-Policy":      "same-origin",
		"Origin-Agent-Cluster":              "?1",
		"X-DNS-Prefetch-Control":            "off",
		"X-Download-Options":                "noopen",
		"X-Permitted-Cross-Domain-Policies": "none",
	}

	for k, expected := range headers {
		actual := resp.HeaderMap.Get(k)
		if actual != expected {
			t.Errorf("Expected %s for header %s, got: %s", expected, k, actual)
		}
	}

	if resp.HeaderMap.Get("Strict-Transport-Security") != "" {
		t.Errorf("Did not expect HSTS header to be set by default")
	}
}

func TestHelmet_Custom(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.Helmet(middleware.HelmetConfig{
		XSSProtection:         "1; mode=block",
		XFrameOptions:         "DENY",
		ContentSecurityPolicy: "default-src 'self'",
		HSTSMaxAge:            31536000,
		HSTSPreload:           true,
	}))

	app.Get("/", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "hello")
		},
	})

	resp := app.Test("GET", "/")

	if resp.HeaderMap.Get("X-XSS-Protection") != "1; mode=block" {
		t.Errorf("Expected custom XSS header")
	}

	if resp.HeaderMap.Get("X-Frame-Options") != "DENY" {
		t.Errorf("Expected custom Frame options")
	}

	if resp.HeaderMap.Get("Content-Security-Policy") != "default-src 'self'" {
		t.Errorf("Expected custom CSP header")
	}

	expectedHSTS := "max-age=31536000; includeSubDomains; preload"
	if resp.HeaderMap.Get("Strict-Transport-Security") != expectedHSTS {
		t.Errorf("Expected HSTS '%s', got: %s", expectedHSTS, resp.HeaderMap.Get("Strict-Transport-Security"))
	}
}
