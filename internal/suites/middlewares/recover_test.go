package middleware

import (
	"encoding/json"
	"errors"
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
	resp := mustTest(t, app, "GET", "/panic")

	if resp.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", resp.StatusCode)
	}

	var body struct {
		Status     string `json:"status"`
		StatusCode int    `json:"statusCode"`
		Message    string `json:"message"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("Expected JSON error response, got error: %v body: %s", err, string(resp.Body))
	}
	if body.Status != "error" || body.StatusCode != 500 || body.Message != "panic recovered: something went wrong" {
		t.Fatalf("Unexpected error response: %+v", body)
	}

	if !strings.Contains(out.String(), "panic recovered: something went wrong") {
		t.Errorf("Expected panic log, got: %s", out.String())
	}

	if !strings.Contains(out.String(), "goroutine") {
		t.Errorf("Expected stack trace in output")
	}

	// Test safe route to ensure server still works
	resp = mustTest(t, app, "GET", "/safe")
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

	resp := mustTest(t, app, "GET", "/panic")

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

func TestRecoverForwardsPanicToRouterErrorHandler(t *testing.T) {
	app := gofi.NewRouter()

	var out stringWriter
	logger := log.New(&out, "", 0)

	app.Use(middleware.Recover(middleware.RecoverConfig{
		EnableStackTrace: false,
		Output:           logger,
	}))

	var capturedErr error
	app.UseErrorHandler(func(err error, c gofi.Context) {
		capturedErr = err
		_ = c.SendString(555, "custom:"+err.Error())
	})

	app.Get("/panic", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			panic("boom")
		},
	})

	resp := mustTest(t, app, "GET", "/panic")

	if resp.StatusCode != 555 {
		t.Fatalf("Expected custom status code 555, got %d", resp.StatusCode)
	}
	if string(resp.Body) != "custom:panic recovered: boom" {
		t.Fatalf("Expected custom body, got %s", string(resp.Body))
	}
	if capturedErr == nil {
		t.Fatal("Expected router error handler to receive recovered panic error")
	}
	if capturedErr.Error() != "panic recovered: boom" {
		t.Fatalf("Unexpected captured error: %v", capturedErr)
	}
}

func TestRecoverCanHandlePanicLocally(t *testing.T) {
	app := gofi.NewRouter()

	var out stringWriter
	logger := log.New(&out, "", 0)

	globalHandlerCalled := false
	app.UseErrorHandler(func(err error, c gofi.Context) {
		globalHandlerCalled = true
		_ = c.SendString(500, err.Error())
	})

	app.Use(middleware.Recover(middleware.RecoverConfig{
		EnableStackTrace: false,
		Output:           logger,
		ErrorHandler: func(c gofi.Context, r any) error {
			if err := c.SendString(204, ""); err != nil {
				return err
			}
			return nil
		},
	}))

	app.Get("/panic", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			panic(errors.New("handled locally"))
		},
	})

	resp := mustTest(t, app, "GET", "/panic")

	if resp.StatusCode != 204 {
		t.Fatalf("Expected local handler status code 204, got %d", resp.StatusCode)
	}
	if globalHandlerCalled {
		t.Fatal("Expected router error handler not to be called when recover handles panic locally")
	}
	if !strings.Contains(out.String(), "panic recovered: handled locally") {
		t.Fatalf("Expected panic log, got: %s", out.String())
	}
}
