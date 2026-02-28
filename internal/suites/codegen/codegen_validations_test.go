package codegen

import (
	"net/http/httptest"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Helper to initialize Mux with specific Compiled Schema
// =============================================================================

func setupValidationMux() gofi.Router {
	m := gofi.NewServeMux()
	m.UseErrorHandler(func(err error, c gofi.Context) {
		c.SendString(400, err.Error())
	})
	return m
}

// =============================================================================
// Inline Rules Evaluation Tests
// =============================================================================

func TestCodegen_Validations_Valid(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &InlineValidationsSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[InlineValidationsSchema](c)
			assert.Nil(t, err)

			assert.Equal(t, "DEV123", s.Request.Query.Code)
			assert.Equal(t, 25, s.Request.Query.Age)

			s.Ok.Body = "valid"
			return c.SendString(200, s.Ok.Body)
		},
	}

	m := setupValidationMux()
	m.Get("/verify", handler)

	// Length 6 (between 4 and 8), Age 25 (between 18 and 65)
	req := httptest.NewRequest("GET", "/verify?code=DEV123&age=25", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
}

func TestCodegen_Validations_StringLength(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &InlineValidationsSchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[InlineValidationsSchema](c)
			return err
		},
	}

	m := setupValidationMux()
	m.Get("/verify", handler)

	// Code length 2 (less than 4), Age 25 (valid)
	req := httptest.NewRequest("GET", "/verify?code=AB&age=25", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Code)
	assert.Contains(t, rec.Body.String(), "[query] code: length must be at least 4")
}

func TestCodegen_Validations_StringMaxLength(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &InlineValidationsSchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[InlineValidationsSchema](c)
			return err
		},
	}

	m := setupValidationMux()
	m.Get("/verify", handler)

	// Code length 10 (greater than 8), Age 25 (valid)
	req := httptest.NewRequest("GET", "/verify?code=TOOLONGCODE&age=25", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Code)
	assert.Contains(t, rec.Body.String(), "[query] code: length must be at most 8")
}

func TestCodegen_Validations_NumberBounds(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &InlineValidationsSchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[InlineValidationsSchema](c)
			return err
		},
	}

	m := setupValidationMux()
	m.Get("/verify", handler)

	// Code length 6 (valid), Age 15 (less than 18)
	req := httptest.NewRequest("GET", "/verify?code=DEV123&age=15", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Code)
	assert.Contains(t, rec.Body.String(), "[query] age: value must be at least 18")
}

func TestCodegen_Validations_NumberMaxBounds(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &InlineValidationsSchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[InlineValidationsSchema](c)
			return err
		},
	}

	m := setupValidationMux()
	m.Get("/verify", handler)

	// Code length 6 (valid), Age 75 (greater than 65)
	req := httptest.NewRequest("GET", "/verify?code=DEV123&age=75", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Code)
	assert.Contains(t, rec.Body.String(), "[query] age: value must be at most 65")
}
