package middleware

import (
	"log"
	"strings"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestLogger_SanitizesControlCharacters(t *testing.T) {
	app := gofi.NewRouter()

	var out stringWriter
	logger := log.New(&out, "", 0)

	app.Use(middleware.Logger(middleware.LoggerConfig{Output: logger}))

	handler := gofi.DefineHandler(gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "ok")
		},
	})

	_, err := app.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/log-sec",
		Headers: map[string]string{
			"X-Forwarded-For": "10.0.0.1\nINJECT\r\tEND",
		},
		Handler: &handler,
	})
	if err != nil {
		t.Fatalf("inject failed: %v", err)
	}

	output := out.String()
	if strings.Contains(output, "\nINJECT") || strings.Contains(output, "\r") || strings.Contains(output, "\t") {
		t.Fatalf("expected log output to sanitize control chars, got %q", output)
	}

	if strings.Count(output, "\n") != 1 {
		t.Fatalf("expected exactly one log line, got %q", output)
	}
}
