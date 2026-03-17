package middleware

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestStaticMiddleware(t *testing.T) {
	// Create a temporary test directory with some static files
	dir := t.TempDir()

	// Create index.html
	indexContent := "<h1>Hello Static</h1>"
	err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(indexContent), 0644)
	if err != nil {
		t.Fatal("Failed to create index.html in TempDir:", err)
	}

	// Create app.css
	cssContent := "body { color: red; }"
	err = os.WriteFile(filepath.Join(dir, "app.css"), []byte(cssContent), 0644)
	if err != nil {
		t.Fatal("Failed to create app.css in TempDir:", err)
	}

	app := gofi.NewRouter()

	// Use Static middleware mounted on /assets prefix
	app.Use(middleware.Static(middleware.StaticConfig{
		Root:   dir,
		Prefix: "/assets",
	}))

	// A fallback route to test that 404s on static files continue down the chain
	app.Get("/assets/fallback", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "fallback payload")
		},
	})

	// Register the test files exactly to trigger the middleware
	app.Get("/assets/app.css", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(404, "") },
	})
	app.Get("/assets/", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(404, "") },
	})

	// Test case 1: Retrieve valid static file (app.css)
	resp1 := mustTest(t, app, "GET", "/assets/app.css")
	if resp1.StatusCode != 200 {
		t.Errorf("Expected 200 for app.css, got %d", resp1.StatusCode)
	}

	if !strings.Contains(string(resp1.Body), "color: red") {
		t.Errorf("Expected css content, got %s", string(resp1.Body))
	}

	// Test case 2: Retrieve index file implicitly
	resp2 := mustTest(t, app, "GET", "/assets/")
	if resp2.StatusCode != 200 {
		t.Errorf("Expected 200 for index.html implicitly, got %d", resp2.StatusCode)
	}

	if !strings.Contains(string(resp2.Body), "Hello Static") {
		t.Errorf("Expected html content from index, got %s", string(resp2.Body))
	}

	// Test case 3: Try to access nonexistent file, should hit fallback handler
	resp3 := mustTest(t, app, "GET", "/assets/fallback")
	if resp3.StatusCode != 200 {
		t.Errorf("Expected 200 from fallback route since file wasn't found, got %d", resp3.StatusCode)
	}

	if string(resp3.Body) != "fallback payload" {
		t.Errorf("Expected fallback payload, got %s", string(resp3.Body))
	}

	// Test case 4: Non-GET methods ignored
	resp4, err := app.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/assets/app.css",
		Handler: &gofi.RouteOptions{Handler: func(c gofi.Context) error { return c.SendString(404, "") }},
	})
	if err != nil {
		t.Errorf("Inject error: %v", err)
	} else if resp4 != nil && resp4.StatusCode != 404 {
		t.Errorf("Expected 404 for POST request out of static scope, got %d", resp4.StatusCode)
	}
}
