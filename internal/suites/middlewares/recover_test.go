package middleware

import (
	"log"
	"strings"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

type stringWriter struct {
	strings.Builder
}

func (w *stringWriter) Write(p []byte) (n int, err error) {
	return w.Builder.Write(p)
}

func TestRecover(t *testing.T) {
	app := gofi.NewRouter()

	var out stringWriter
	logger := log.New(&out, "", 0)

	app.Use(middleware.Recover(middleware.RecoverConfig{
		EnableStackTrace: true,
		Output:           logger,
	}))

	panicRoute := gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			panic("something went wrong")
		},
	}
	app.Get("/panic", panicRoute)

	safeRoute := gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "safe")
		},
	}
	// Test normal route
	app.Get("/safe", safeRoute)

	// Test panic route (using Test to go through the whole stack since Panic Recovery needs to wrap Next())
	resp := app.Test("GET", "/panic")

	if resp.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", resp.StatusCode)
	}

	if !strings.Contains(out.String(), "panic recovered: something went wrong") {
		t.Errorf("Expected panic log, got: %s", out.String())
	}

	if !strings.Contains(out.String(), "goroutine") {
		t.Errorf("Expected stack trace in output")
	}

	// Test safe route to ensure server still works
	resp = app.Test("GET", "/safe")
	if resp.StatusCode != 200 || string(resp.Body) != "safe" {
		t.Errorf("Server broke after panic. Expected 200 'safe', got %d %s", resp.StatusCode, string(resp.Body))
	}
}

func TestRecoverNoStackTrace(t *testing.T) {
	app := gofi.NewRouter()

	var out stringWriter
	logger := log.New(&out, "", 0)

	app.Use(middleware.Recover(middleware.RecoverConfig{
		EnableStackTrace: false,
		Output:           logger,
	}))

	panicRoute := gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			panic("test string")
		},
	}
	app.Get("/panic", panicRoute)

	resp := app.Test("GET", "/panic")

	if resp.StatusCode != 500 {
		t.Errorf("Expected 500, got %d", resp.StatusCode)
	}

	output := out.String()
	if !strings.Contains(output, "panic recovered: test string") {
		t.Errorf("Expected log containing 'panic recovered: test string', got %s", output)
	}

	if strings.Contains(output, "goroutine") {
		t.Errorf("Did not expect stack trace, got: %s", output)
	}
}
