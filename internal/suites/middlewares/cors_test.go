package middleware

import (
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestCORS_Basic(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.CORS())

	app.Get("/cors", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "hello cors")
		},
	})

	// Test case 1: No Origin header
	resp := mustTest(t, app, "GET", "/cors")
	if resp.HeaderMap.Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("Expected strictly no CORS headers when Origin is omitted")
	}

	// Test case 2: Basic Origin request
	injectResp, _ := app.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/cors",
		Headers: map[string]string{
			"Origin": "https://example.com",
		},
		Handler: &gofi.RouteOptions{
			Handler: func(c gofi.Context) error {
				return c.SendString(200, "hello cors")
			},
		},
	})

	if injectResp.HeaderMap.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected * for origin, got: %s", injectResp.HeaderMap.Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_PreflightAndConfig(t *testing.T) {
	app := gofi.NewRouter()

	app.Use(middleware.CORS(middleware.CORSConfig{
		AllowOrigins:     "https://foo.com, https://bar.com",
		AllowMethods:     "GET, POST",
		AllowCredentials: true,
		AllowHeaders:     "Accept, Content-Type, X-My-Header",
		ExposeHeaders:    "X-My-Custom",
	}))

	// Test case 1: Disallowed Origin should not get headers
	req1, _ := app.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/",
		Headers: map[string]string{"Origin": "https://baz.com"},
		Handler: &gofi.RouteOptions{Handler: func(c gofi.Context) error { return nil }},
	})

	if req1.HeaderMap.Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("Did not expect Origin header for disallowed origin")
	}

	// Test case 2: Preflight OPTIONS request
	req2, _ := app.Inject(gofi.InjectOptions{
		Method:  "OPTIONS",
		Path:    "/",
		Headers: map[string]string{"Origin": "https://foo.com"},
		Handler: &gofi.RouteOptions{Handler: func(c gofi.Context) error { return nil }},
	})

	if req2.StatusCode != 204 {
		t.Errorf("Expected 204 No Content for preflight, got %d", req2.StatusCode)
	}

	if req2.HeaderMap.Get("Access-Control-Allow-Origin") != "https://foo.com" {
		t.Errorf("Expected origin to be matched and set, got: %s", req2.HeaderMap.Get("Access-Control-Allow-Origin"))
	}

	if req2.HeaderMap.Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("Expected credentials to be true")
	}

	if req2.HeaderMap.Get("Access-Control-Allow-Methods") != "GET, POST" {
		t.Errorf("Expected specific methods")
	}

	if req2.HeaderMap.Get("Access-Control-Allow-Headers") != "Accept, Content-Type, X-My-Header" {
		t.Errorf("Expected specific allow headers")
	}

	if req2.HeaderMap.Get("Access-Control-Expose-Headers") != "X-My-Custom" {
		t.Errorf("Expected specific expose headers")
	}
}
