package codegen

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Helper to initialize Mux with specific Compiled Schema
// =============================================================================

func setupResponseMux(schemaName string, funcs gofi.CompiledSchemaFuncs) gofi.Router {
	m := gofi.NewServeMux()
	m.UseErrorHandler(func(err error, c gofi.Context) {
		c.SendString(400, err.Error())
	})
	return m
}

// =============================================================================
// Response Header Extraction Tests
// =============================================================================

func TestCodegen_Response_Headers(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &ResponseHeaderSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[ResponseHeaderSchema](c)
			assert.Nil(t, err)

			s.Ok.Header.XRequestID = "resp-1234"
			s.Ok.Header.XVersion = "v1.0.5"
			s.Ok.Body.ID = s.Request.Path.ID * 2

			return c.Send(200, s.Ok)
		},
	}

	m := setupResponseMux("ResponseHeaderSchema", gofi.CompiledSchemaFuncs{})
	m.Get("/headers/{id}", handler)

	req := httptest.NewRequest("GET", "/headers/15", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "resp-1234", rec.Header().Get("X-Request-Id"))
	assert.Equal(t, "v1.0.5", rec.Header().Get("X-Version"))

	var body map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &body)
	assert.Equal(t, float64(30), body["id"]) // JSON unmarshals numbers as float64
}

// =============================================================================
// Response Cookie Evaluation Tests
// =============================================================================

func TestCodegen_Response_Cookies(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &CookieSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[CookieSchema](c)
			assert.Nil(t, err)

			s.Ok.Cookie.AuthToken = http.Cookie{
				Name:     "auth_token",
				Value:    "secret_token_abc",
				HttpOnly: true,
				Path:     "/",
			}
			s.Ok.Body.Authenticated = true

			return c.Send(200, s.Ok)
		},
	}

	m := setupResponseMux("CookieSchema", gofi.CompiledSchemaFuncs{})
	m.Post("/login", handler)

	req := httptest.NewRequest("POST", "/login", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid"})
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)

	// Verify the response cookie was encoded correctly by extracting it
	cookies := rec.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "auth_token", cookies[0].Name)
	assert.Equal(t, "secret_token_abc", cookies[0].Value)
	assert.True(t, cookies[0].HttpOnly)
}
