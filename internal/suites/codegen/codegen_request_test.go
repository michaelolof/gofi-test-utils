package codegen

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Helper to initialize Mux with specific Compiled Schema
// =============================================================================

func setupRequestMux(schemaName string, funcs gofi.CompiledSchemaFuncs) gofi.Router {
	m := gofi.NewServeMux()
	m.UseErrorHandler(func(err error, c gofi.Context) {
		c.SendString(400, err.Error())
	})
	// NOTE: Because `RegisterCompiledSchema` is called in `init()`, the schemas
	// are already registered globally. But we ensure the fast path is used here.
	return m
}

// =============================================================================
// Header / Query / Path Extraction Tests
// =============================================================================

func TestCodegen_Request_HeaderQueryPath_Valid(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &HeaderQueryPathSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[HeaderQueryPathSchema](c)
			assert.Nil(t, err)

			// Validate Headers
			assert.Equal(t, "Bearer token123", s.Request.Header.Authorization)
			assert.Equal(t, "req-555", s.Request.Header.XRequestID)

			// Validate Queries
			assert.Equal(t, 2, s.Request.Query.Page)
			assert.Equal(t, "desc", s.Request.Query.Sort)
			assert.True(t, s.Request.Query.Active)

			// Validate Paths
			assert.Equal(t, 101, s.Request.Path.ID)
			assert.Equal(t, "books", s.Request.Path.Category)
			assert.Equal(t, 4.5, s.Request.Path.Rating)

			return c.SendString(200, "ok")
		},
	}

	m := setupRequestMux("HeaderQueryPathSchema", gofi.CompiledSchemaFuncs{})
	m.Get("/items/{id}/category/{category}/rating/{rating}", handler)

	req := httptest.NewRequest("GET", "/items/101/category/books/rating/4.5?page=2&sort=desc&active=true", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("X-Request-Id", "req-555")

	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
}

func TestCodegen_Request_HeaderQueryPath_Defaults(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &HeaderQueryPathSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[HeaderQueryPathSchema](c)
			assert.Nil(t, err)

			// Default values should be applied
			assert.Equal(t, "auto-generated", s.Request.Header.XRequestID)
			assert.Equal(t, 1, s.Request.Query.Page)
			assert.Equal(t, "", s.Request.Query.Sort)
			assert.False(t, s.Request.Query.Active)
			assert.Equal(t, 0.0, s.Request.Path.Rating) // zero value when omitted

			return c.SendString(200, "ok")
		},
	}

	m := setupRequestMux("HeaderQueryPathSchema", gofi.CompiledSchemaFuncs{})
	m.Get("/items/{id}/category/{category}", handler)

	req := httptest.NewRequest("GET", "/items/101/category/books", nil)
	req.Header.Set("Authorization", "Bearer token123")

	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
}

func TestCodegen_Request_HeaderQueryPath_MissingRequired(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &HeaderQueryPathSchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[HeaderQueryPathSchema](c)
			// Should return joined errors
			assert.NotNil(t, err)

			msg := err.Error()
			assert.Contains(t, msg, "[header] Authorization: field is required")
			assert.Contains(t, msg, "[path] id: field is required")
			assert.Contains(t, msg, "[path] category: field is required")

			return err // Framework error handler maps this to 400
		},
	}

	m := setupRequestMux("HeaderQueryPathSchema", gofi.CompiledSchemaFuncs{})
	m.Get("/bare", handler)

	req := httptest.NewRequest("GET", "/bare", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Code)
}

func TestCodegen_Request_HeaderQueryPath_TypeConversionErrors(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &HeaderQueryPathSchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[HeaderQueryPathSchema](c)
			assert.NotNil(t, err)

			msg := err.Error()
			assert.Contains(t, msg, "[query] page: invalid integer")
			assert.Contains(t, msg, "[path] id: invalid integer")
			assert.Contains(t, msg, "[path] rating: invalid float")

			return err
		},
	}

	m := setupRequestMux("HeaderQueryPathSchema", gofi.CompiledSchemaFuncs{})
	m.Get("/items/{id}/category/{category}/rating/{rating}", handler)

	req := httptest.NewRequest("GET", "/items/not-an-int/category/books/rating/not-a-float?page=not-an-int", nil)
	req.Header.Set("Authorization", "Bearer token123")

	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Code)
}

// =============================================================================
// Cookie Extraction Tests
// =============================================================================

func TestCodegen_Request_Cookie_Valid(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &CookieSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[CookieSchema](c)
			assert.Nil(t, err)

			// session_id is string
			assert.Equal(t, "sess_12345", s.Request.Cookie.SessionID)

			// tracking is http.Cookie
			assert.Equal(t, "tracking", s.Request.Cookie.Tracking.Name)
			assert.Equal(t, "user_abc", s.Request.Cookie.Tracking.Value)

			return c.SendString(200, "ok")
		},
	}

	m := setupRequestMux("CookieSchema", gofi.CompiledSchemaFuncs{})
	m.Post("/login", handler)

	req := httptest.NewRequest("POST", "/login", strings.NewReader(`{}`))
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "sess_12345"})
	req.AddCookie(&http.Cookie{Name: "tracking", Value: "user_abc"})

	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
}

func TestCodegen_Request_Cookie_MissingRequired(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &CookieSchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[CookieSchema](c)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "[cookie] session_id: field is required")
			return err
		},
	}

	m := setupRequestMux("CookieSchema", gofi.CompiledSchemaFuncs{})
	m.Post("/login", handler)

	req := httptest.NewRequest("POST", "/login", strings.NewReader(`{}`))
	// missing session_id cookie

	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Code)
}
