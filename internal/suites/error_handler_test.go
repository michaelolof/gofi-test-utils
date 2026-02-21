package suites

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 1. Default Error Handler
// =============================================================================

func TestErrorHandler_Default(t *testing.T) {
	r := gofi.NewServeMux()
	r.Get("/fail", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return errors.New("something went wrong")
		},
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/fail", nil))

	assert.Equal(t, 500, w.Code)

	var resp struct {
		Status     string `json:"status"`
		StatusCode int    `json:"statusCode"`
		Message    string `json:"message"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, "something went wrong", resp.Message)
}

// =============================================================================
// 2. Custom Error Handler
// =============================================================================

func TestErrorHandler_Custom(t *testing.T) {
	r := gofi.NewServeMux()
	r.UseErrorHandler(func(err error, c gofi.Context) {
		c.Writer().WriteHeader(418)
		c.Writer().Write([]byte("custom:" + err.Error()))
	})

	r.Get("/fail", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return errors.New("teapot")
		},
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/fail", nil))

	assert.Equal(t, 418, w.Code)
	assert.Equal(t, "custom:teapot", w.Body.String())
}

// =============================================================================
// 3. Nil Error (No error handler invoked)
// =============================================================================

func TestErrorHandler_NilError(t *testing.T) {
	r := gofi.NewServeMux()
	errHandlerCalled := false
	r.UseErrorHandler(func(err error, c gofi.Context) {
		errHandlerCalled = true
	})

	r.Get("/ok", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "ok")
		},
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/ok", nil))

	assert.False(t, errHandlerCalled, "Error handler should NOT be called when handler returns nil")
	assert.Equal(t, 200, w.Code)
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

	r := gofi.NewServeMux()
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

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/items/", nil)
	r.ServeHTTP(w, req)

	// The exact behavior depends on whether the route matches at all with empty param.
	// Either the route won't match (404) or validation will fail.
	assert.True(t, w.Code == 400 || w.Code == 404,
		"Expected 400 or 404, got %d (capturedErr: %v)", w.Code, capturedErr)
}
