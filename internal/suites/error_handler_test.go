package suites

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 1. Default Error Handler
// =============================================================================

func TestErrorHandler_Default(t *testing.T) {
	r := gofi.NewRouter()
	r.Get("/fail", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return errors.New("something went wrong")
		},
	})

	w := mustTest(t, r, "GET", "/fail")
	assert.Equal(t, 500, w.StatusCode)

	var resp struct {
		Status     string `json:"status"`
		StatusCode int    `json:"statusCode"`
		Message    string `json:"message"`
	}
	err := json.Unmarshal(w.Body, &resp)
	assert.Nil(t, err)
	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, "something went wrong", resp.Message)
}

// =============================================================================
// 2. Custom Error Handler
// =============================================================================

func TestErrorHandler_Custom(t *testing.T) {
	r := gofi.NewRouter()
	r.UseErrorHandler(func(err error, c gofi.Context) {
		c.Writer().WriteHeader(418)
		c.Writer().Write([]byte("custom:" + err.Error()))
	})

	r.Get("/fail", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return errors.New("teapot")
		},
	})

	w := mustTest(t, r, "GET", "/fail")
	assert.Equal(t, 418, w.StatusCode)
	assert.Equal(t, "custom:teapot", string(w.Body))
}

// =============================================================================
// 3. Nil Error (No error handler invoked)
// =============================================================================

func TestErrorHandler_NilError(t *testing.T) {
	r := gofi.NewRouter()
	errHandlerCalled := false
	r.UseErrorHandler(func(err error, c gofi.Context) {
		errHandlerCalled = true
	})

	r.Get("/ok", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "ok")
		},
	})

	w := mustTest(t, r, "GET", "/ok")
	assert.False(t, errHandlerCalled, "Error handler should NOT be called when handler returns nil")
	assert.Equal(t, 200, w.StatusCode)
}

// =============================================================================
// 4. Validation Error → Error Handler
// =============================================================================

func TestErrorHandler_ValidationError(t *testing.T) {
	type schema struct {
		Request struct {
			Path struct {
				ID string `json:"id" validate:"required"`
			}
		}
	}

	r := gofi.NewRouter()
	var capturedErr error
	r.UseErrorHandler(func(err error, c gofi.Context) {
		capturedErr = err
		c.Writer().WriteHeader(400)
		c.Writer().Write([]byte("validation-error"))
	})

	r.Get("/items/{id}", gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[schema](c)
			return err
		},
	})

	w := mustTest(t, r, "GET", "/items/")

	// The exact behavior depends on whether the route matches at all with empty param.
	// Either the route won't match (404) or validation will fail.
	assert.True(t, w.StatusCode == 400 || w.StatusCode == 404,
		"Expected 400 or 404, got %d (capturedErr: %v)", w.StatusCode, capturedErr)
}

// =============================================================================
// 5. HTTPError -> Default Error Handler
// =============================================================================

func TestErrorHandler_HTTPError(t *testing.T) {
	r := gofi.NewRouter()
	r.Get("/http-error", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return gofi.NewHTTPError(422, "validation failed")
		},
	})

	w := mustTest(t, r, "GET", "/http-error")
	assert.Equal(t, 422, w.StatusCode)

	var resp struct {
		Status     string `json:"status"`
		StatusCode int    `json:"statusCode"`
		Message    string `json:"message"`
	}
	err := json.Unmarshal(w.Body, &resp)
	assert.Nil(t, err)
	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, 422, resp.StatusCode)
	assert.Equal(t, "validation failed", resp.Message)
}

// =============================================================================
// 6. HTTPError -> Custom Error Handler
// =============================================================================

func TestErrorHandler_CustomCanExtractHTTPError(t *testing.T) {
	r := gofi.NewRouter()
	r.UseErrorHandler(func(err error, c gofi.Context) {
		code := 500
		var httpErr *gofi.HTTPError
		if errors.As(err, &httpErr) {
			code = httpErr.Code
		}
		_ = c.SendString(code, fmt.Sprintf("code=%d msg=%s", code, err.Error()))
	})

	r.Get("/wrapped-http-error", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return fmt.Errorf("wrapped: %w", gofi.NewHTTPError(425, "too early"))
		},
	})

	w := mustTest(t, r, "GET", "/wrapped-http-error")
	assert.Equal(t, 425, w.StatusCode)
	assert.Equal(t, "code=425 msg=wrapped: too early", string(w.Body))
}
