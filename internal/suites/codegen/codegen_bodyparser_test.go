package codegen

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Helper to initialize Mux with specific Compiled Schema
// =============================================================================

func setupBodyParserMux(schemaName string, funcs gofi.CompiledSchemaFuncs) gofi.Router {
	m := gofi.NewServeMux()
	m.UseErrorHandler(func(err error, c gofi.Context) {
		c.SendString(400, err.Error())
	})
	return m
}

// =============================================================================
// JSON Body Extraction and Validation Tests
// =============================================================================

func TestCodegen_BodyParser_JSON_Valid(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &JSONBodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[JSONBodySchema](c)
			assert.Nil(t, err)

			assert.Equal(t, "John Doe", s.Request.Body.Name)
			assert.Equal(t, "john@example.com", s.Request.Body.Email)
			assert.Equal(t, 25, s.Request.Body.Age)

			s.Ok.Body.ID = 100
			s.Ok.Body.Name = s.Request.Body.Name
			s.Ok.Body.Email = s.Request.Body.Email

			return c.Send(200, s.Ok)
		},
	}

	m := setupBodyParserMux("JSONBodySchema", gofi.CompiledSchemaFuncs{})
	m.Post("/users", handler)

	body := `{"name": "John Doe", "email": "john@example.com", "age": 25}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	rec := httptest.NewRecorder()

	m.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)

	var respBody map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &respBody)
	assert.Equal(t, float64(100), respBody["id"])
	assert.Equal(t, "John Doe", respBody["name"])
	assert.Equal(t, "john@example.com", respBody["email"])
}

func TestCodegen_BodyParser_JSON_MissingRequiredBody(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &JSONBodySchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[JSONBodySchema](c)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "[body]: request body is required")
			return err
		},
	}

	m := setupBodyParserMux("JSONBodySchema", gofi.CompiledSchemaFuncs{})
	m.Post("/users", handler)

	// Empty request body
	req := httptest.NewRequest("POST", "/users", strings.NewReader(""))
	rec := httptest.NewRecorder()

	m.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Code)
}

func TestCodegen_BodyParser_JSON_MalformedJSON(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &JSONBodySchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[JSONBodySchema](c)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "[body]: failed to parse JSON")
			return err
		},
	}

	m := setupBodyParserMux("JSONBodySchema", gofi.CompiledSchemaFuncs{})
	m.Post("/users", handler)

	// Malformed JSON (trailing comma)
	body := `{"name": "John", "email": "john@ex.com",}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	rec := httptest.NewRecorder()

	m.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Code)
}

func TestCodegen_BodyParser_JSON_FieldValidations(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &JSONBodySchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[JSONBodySchema](c)
			assert.NotNil(t, err)

			msg := err.Error()
			// From generated `JSONBodySchema` validators we expect these:
			assert.Contains(t, msg, "[body] name: field is required")
			assert.Contains(t, msg, "[body] email: field is required")
			assert.Contains(t, msg, "[body] age: value must be at least 1")

			return err
		},
	}

	m := setupBodyParserMux("JSONBodySchema", gofi.CompiledSchemaFuncs{})
	m.Post("/users", handler)

	// Missing Name and Email, Age below 1
	body := `{"age": 0}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	rec := httptest.NewRecorder()

	m.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Code)
}

func TestCodegen_BodyParser_JSON_FieldMaxValidation(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &JSONBodySchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[JSONBodySchema](c)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "[body] age: value must be at most 150")
			return err
		},
	}

	m := setupBodyParserMux("JSONBodySchema", gofi.CompiledSchemaFuncs{})
	m.Post("/users", handler)

	// Age exceeds 150
	body := `{"name": "John", "email": "j@ex.com", "age": 200}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	rec := httptest.NewRecorder()

	m.ServeHTTP(rec, req)

	assert.Equal(t, 400, rec.Code)
}
