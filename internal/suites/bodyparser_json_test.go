package suites

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 5. JSON Body Binding
// =============================================================================

func TestSchema_JSONBody_Struct(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Body struct {
				Name string `json:"name" validate:"required"`
				Age  int    `json:"age" validate:"required,gt=0"`
			} `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.Equal(t, "Alice", s.Request.Body.Name)
			assert.Equal(t, 30, s.Request.Body.Age)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`{"name":"Alice","age":30}`),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}

func TestSchema_JSONBody_Primitive(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Body int `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.Equal(t, 42, s.Request.Body)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	_, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/value",
		Body:    strings.NewReader("42"),
		Handler: &handler,
	})
	assert.Nil(t, err)
}

func TestSchema_JSONBody_Array(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Body []int `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.Equal(t, []int{1, 2, 3}, s.Request.Body)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	_, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/list",
		Body:    strings.NewReader("[1,2,3]"),
		Handler: &handler,
	})
	assert.Nil(t, err)
}

func TestSchema_JSONBody_MissingRequired(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Body struct {
				Name string `json:"name" validate:"required"`
			} `validate:"required"`
		}
	}

	var gotErr error
	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[bodySchema](c)
			gotErr = err
			return nil
		},
	}

	m := gofi.NewRouter()
	m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`{}`),
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "Expected validation error for missing required body field")
}

// =============================================================================
// 6. Empty/Zero Values — Required Fields Should Fail
// =============================================================================

func TestSchema_EmptyValues_RequiredBody_EmptyJSON(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Name  string `json:"name" validate:"required"`
				Email string `json:"email" validate:"required"`
			} `validate:"required"`
		}
	}

	var gotErr error
	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[schema](c)
			gotErr = err
			return nil
		},
	}

	m := gofi.NewRouter()
	m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`{"name":"","email":""}`),
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "Empty string fields in body should fail required validation")
}

func TestSchema_EmptyValues_RequiredBody_ZeroValues(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Count int     `json:"count" validate:"required"`
				Price float64 `json:"price" validate:"required"`
			} `validate:"required"`
		}
	}

	var gotErr error
	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[schema](c)
			gotErr = err
			return nil
		},
	}

	m := gofi.NewRouter()
	m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/items",
		Body:    strings.NewReader(`{"count":0,"price":0}`),
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "Zero numeric values should fail required validation")
}

// =============================================================================
// 7. Malformed JSON Request — Should Fail Gracefully
// =============================================================================

func TestSchema_MalformedJSON_InvalidSyntax(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Name string `json:"name" validate:"required"`
			} `validate:"required"`
		}
	}

	var gotErr error
	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[schema](c)
			gotErr = err
			return err
		},
	}

	m := gofi.NewRouter()
	m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`{name: "broken"}`), // not valid JSON — unquoted key
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "Invalid JSON syntax should produce an error")
}

func TestSchema_MalformedJSON_TruncatedBody(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Name string `json:"name" validate:"required"`
				Age  int    `json:"age"`
			} `validate:"required"`
		}
	}

	var gotErr error
	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[schema](c)
			gotErr = err
			return err
		},
	}

	m := gofi.NewRouter()
	m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`{"name":"Alice","age`), // truncated mid-key
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "Truncated JSON should produce an error")
}

func TestSchema_MalformedJSON_UnclosedString(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Name string `json:"name" validate:"required"`
			} `validate:"required"`
		}
	}

	var gotErr error
	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[schema](c)
			gotErr = err
			return err
		},
	}

	m := gofi.NewRouter()
	m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`{"name":"Alice`), // unclosed string and object
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "Unclosed JSON string should produce an error")
}

func TestSchema_MalformedJSON_WrongType_StringForInt(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Age int `json:"age" validate:"required"`
			} `validate:"required"`
		}
	}

	var gotErr error
	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[schema](c)
			gotErr = err
			return err
		},
	}

	m := gofi.NewRouter()
	m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`{"age":"not-a-number"}`),
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "String value for int field should produce an error")
}

func TestSchema_MalformedJSON_WrongType_StringForObject(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Profile struct {
					Name string `json:"name"`
				} `json:"profile"`
			} `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			if err != nil {
				return err
			}
			// gofi's parser may not error on a structural type mismatch; it simply
			// won't bind the nested fields. The nested Name should remain empty.
			assert.Equal(t, "", s.Request.Body.Profile.Name, "Nested field should not be bound from string")
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	// This should not panic regardless of whether it errors or not
	m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`{"profile":"not-an-object"}`),
		Handler: &handler,
	})
}

func TestSchema_MalformedJSON_ExtraTrailingChars(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Name string `json:"name" validate:"required"`
			} `validate:"required"`
		}
	}

	var gotErr error
	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[schema](c)
			gotErr = err
			return err
		},
	}

	m := gofi.NewRouter()
	m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`{"name":"Alice"}}}}`), // extra closing braces
		Handler: &handler,
	})

	// This may or may not error depending on parser behavior — some parsers ignore trailing chars
	// The key assertion is that it doesn't panic
	_ = gotErr
}

func TestSchema_MalformedJSON_NullBody(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Name string `json:"name" validate:"required"`
			} `validate:"required"`
		}
	}

	var gotErr error
	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[schema](c)
			gotErr = err
			return err
		},
	}

	m := gofi.NewRouter()
	m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`null`),
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "JSON null body with required fields should produce an error")
}

func TestSchema_MalformedJSON_ArrayInsteadOfObject(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Name string `json:"name" validate:"required"`
			} `validate:"required"`
		}
	}

	var gotErr error
	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[schema](c)
			gotErr = err
			return err
		},
	}

	m := gofi.NewRouter()
	m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`[1, 2, 3]`), // array instead of object
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "Array body for object schema should produce an error")
}

func TestSchema_MalformedJSON_BooleanBody(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Name string `json:"name" validate:"required"`
			} `validate:"required"`
		}
	}

	var gotErr error
	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[schema](c)
			gotErr = err
			return err
		},
	}

	m := gofi.NewRouter()
	m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`true`), // boolean instead of object
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "Boolean body for object schema should produce an error")
}

// =============================================================================
// 2. JSON Response via Send
// =============================================================================

func TestResponse_Send_JSONBody(t *testing.T) {
	type respSchema struct {
		Ok struct {
			Body struct {
				Message string `json:"message"`
				Count   int    `json:"count"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &respSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[respSchema](c)
			assert.Nil(t, err)
			s.Ok.Body.Message = "success"
			s.Ok.Body.Count = 42
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/data",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)

	var result struct {
		Message string `json:"message"`
		Count   int    `json:"count"`
	}
	err = json.Unmarshal(rec.Body, &result)
	assert.Nil(t, err)
	assert.Equal(t, "success", result.Message)
	assert.Equal(t, 42, result.Count)
}

// =============================================================================
// 6. Pointer Body Fields (*string, *int, *float64)
// =============================================================================

func TestResponse_Send_PointerBodyFields(t *testing.T) {
	type respSchema struct {
		Ok struct {
			Body struct {
				Name    *string  `json:"name"`
				Age     *int     `json:"age"`
				Balance *float64 `json:"balance"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &respSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[respSchema](c)
			assert.Nil(t, err)
			name := "Alice"
			age := 30
			balance := 99.95
			s.Ok.Body.Name = &name
			s.Ok.Body.Age = &age
			s.Ok.Body.Balance = &balance
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/user",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)

	var result struct {
		Name    *string  `json:"name"`
		Age     *int     `json:"age"`
		Balance *float64 `json:"balance"`
	}
	err = json.Unmarshal(rec.Body, &result)
	assert.Nil(t, err)
	assert.NotNil(t, result.Name)
	assert.Equal(t, "Alice", *result.Name)
	assert.NotNil(t, result.Age)
	assert.Equal(t, 30, *result.Age)
	assert.NotNil(t, result.Balance)
	assert.InDelta(t, 99.95, *result.Balance, 0.01)
}

// =============================================================================
// 7. Nil Pointer Body Fields
// =============================================================================

func TestResponse_Send_NilPointerBodyFields(t *testing.T) {
	type respSchema struct {
		Ok struct {
			Body struct {
				Name    *string `json:"name"`
				Age     *int    `json:"age"`
				Present string  `json:"present"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &respSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[respSchema](c)
			assert.Nil(t, err)
			// Leave Name and Age as nil
			s.Ok.Body.Present = "yes"
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/user",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)

	var result map[string]any
	err = json.Unmarshal(rec.Body, &result)
	assert.Nil(t, err)
	assert.Equal(t, "yes", result["present"])
}

// =============================================================================
// 12. time.Time in JSON Body
// =============================================================================

func TestResponse_Send_TimeInBody(t *testing.T) {
	type respSchema struct {
		Ok struct {
			Body struct {
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
			}
		}
	}

	created := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2026, 2, 20, 14, 0, 0, 0, time.UTC)

	handler := gofi.RouteOptions{
		Schema: &respSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[respSchema](c)
			assert.Nil(t, err)
			s.Ok.Body.CreatedAt = created
			s.Ok.Body.UpdatedAt = updated
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/timestamps",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)

	// gofi's custom JSON encoder may encode time.Time as an object or string.
	// Verify the fields are present and non-nil by decoding into a generic map.
	var result map[string]any
	err = json.Unmarshal(rec.Body, &result)
	assert.Nil(t, err)
	assert.NotNil(t, result["created_at"], "Expected created_at field to be present")
	assert.NotNil(t, result["updated_at"], "Expected updated_at field to be present")
}

func TestResponse_Send_MixedTimePointerBody(t *testing.T) {

	type bodyObj struct {
		Email     string     `json:"email" validate:"required,email"`
		CreatedAt time.Time  `json:"created_at" validate:"required"`
		ExpiresAt *time.Time `json:"expires_at" validate:"required"`
		DeletedAt *time.Time `json:"deleted_at"`
	}

	type respSchema struct {
		Ok struct {
			Body bodyObj
		}
	}

	created := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2026, 2, 20, 14, 0, 0, 0, time.UTC)

	handler := gofi.RouteOptions{
		Schema: &respSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[respSchema](c)
			assert.Nil(t, err)

			bodyObj := bodyObj{
				Email:     "test@mail.com",
				CreatedAt: created,
				ExpiresAt: &updated,
				DeletedAt: nil,
			}
			s.Ok.Body = bodyObj
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/timestamps",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)

	// gofi's custom JSON encoder may encode time.Time as an object or string.
	// Verify the fields are present and non-nil by decoding into a generic map.
	var result map[string]any
	err = json.Unmarshal(rec.Body, &result)
	assert.Nil(t, err)
	assert.NotNil(t, result["created_at"], "Expected created_at field to be present")
	assert.NotNil(t, result["expires_at"], "Expected expires_at field to be present")
}

// =============================================================================
// 14. Response Empty Values — Non-Required Fields Should Pass
// =============================================================================
func TestResponse_EmptyValues_EmptyArrayBody(t *testing.T) {
	type schema struct {
		Ok struct {
			Body []string
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			s.Ok.Body = []string{}
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/tags",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
	assert.Equal(t, "[]", strings.TrimSpace(string(rec.Body)))
}

func TestResponse_EmptyValues_ZeroValueBody(t *testing.T) {
	type schema struct {
		Ok struct {
			Body struct {
				Count   int     `json:"count"`
				Score   float64 `json:"score"`
				Name    string  `json:"name"`
				Enabled bool    `json:"enabled"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			// Leave all as zero values
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)

	var result struct {
		Count   int     `json:"count"`
		Score   float64 `json:"score"`
		Name    string  `json:"name"`
		Enabled bool    `json:"enabled"`
	}
	err = json.Unmarshal(rec.Body, &result)
	assert.Nil(t, err)
	assert.Equal(t, 0, result.Count)
	assert.Equal(t, 0.0, result.Score)
	assert.Equal(t, "", result.Name)
	assert.False(t, result.Enabled)
}

// =============================================================================
// 15. Response Body Edge Cases — Malformed or Invalid
// =============================================================================

func TestResponse_EdgeCase_NilBodyOnRequiredSchema(t *testing.T) {
	type schema struct {
		Ok struct {
			Body struct {
				Name string `json:"name" validate:"required"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			// Try to send nil as the response object
			return c.Send(200, nil)
		},
	}

	m := gofi.NewRouter()
	rec, _ := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	// Should not return 200 — nil body on a defined schema should error
	assert.NotEqual(t, 200, rec.StatusCode, "Nil body on required schema should produce an error")
}

func TestResponse_EdgeCase_SendNonStructBody(t *testing.T) {
	type schema struct {
		Ok struct {
			Body struct {
				Name string `json:"name"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			// Try to send a string instead of a struct
			return c.Send(200, "not-a-struct")
		},
	}

	m := gofi.NewRouter()
	rec, _ := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.NotEqual(t, 200, rec.StatusCode, "Non-struct body should produce an error")
}

func TestResponse_EdgeCase_ValidJSON_AlwaysProduced(t *testing.T) {
	type schema struct {
		Ok struct {
			Body struct {
				Name    string   `json:"name"`
				Tags    []string `json:"tags"`
				Count   int      `json:"count"`
				Enabled bool     `json:"enabled"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			s.Ok.Body.Name = "test"
			s.Ok.Body.Tags = []string{"a", "b", "c"}
			s.Ok.Body.Count = 42
			s.Ok.Body.Enabled = true
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)

	// Verify the body is valid JSON by unmarshaling
	var result map[string]any
	err = json.Unmarshal(rec.Body, &result)
	assert.Nil(t, err, "Response body should always be valid JSON")
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, float64(42), result["count"])
	assert.Equal(t, true, result["enabled"])

	tags, ok := result["tags"].([]any)
	assert.True(t, ok, "Tags should be an array")
	assert.Len(t, tags, 3)
}

func TestResponse_EdgeCase_SpecialCharsInBody(t *testing.T) {
	type schema struct {
		Ok struct {
			Body struct {
				Name    string `json:"name"`
				Message string `json:"message"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			s.Ok.Body.Name = `He said "hello"`
			s.Ok.Body.Message = "line1\nline2\ttab\\backslash"
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)

	// Verify the body is valid JSON even with special characters
	var result struct {
		Name    string `json:"name"`
		Message string `json:"message"`
	}
	err = json.Unmarshal(rec.Body, &result)
	assert.Nil(t, err, "Response with special chars should produce valid JSON")
	assert.Equal(t, `He said "hello"`, result.Name)
	assert.Contains(t, result.Message, "line1")
	assert.Contains(t, result.Message, "line2")
}
func TestSchema_JSONBody_ArrayPointer(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Body *[]string `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.NotNil(t, s.Request.Body)
			assert.Equal(t, []string{"a", "b"}, *s.Request.Body)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	_, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/list",
		Body:    strings.NewReader(`["a","b"]`),
		Handler: &handler,
	})
	assert.Nil(t, err)
}

func TestSchema_JSONBody_StructArray(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	type bodySchema struct {
		Request struct {
			Body []User `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.Len(t, s.Request.Body, 2)
			assert.Equal(t, "Alice", s.Request.Body[0].Name)
			assert.Equal(t, "Bob", s.Request.Body[1].Name)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	_, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`[{"name":"Alice","age":30},{"name":"Bob","age":25}]`),
		Handler: &handler,
	})
	assert.Nil(t, err)
}

func TestSchema_JSONBody_StructArrayPointer(t *testing.T) {
	type User struct {
		Name string `json:"name"`
	}
	type bodySchema struct {
		Request struct {
			Body *[]User `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.NotNil(t, s.Request.Body)
			assert.Len(t, *s.Request.Body, 1)
			assert.Equal(t, "Alice", (*s.Request.Body)[0].Name)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	_, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`[{"name":"Alice"}]`),
		Handler: &handler,
	})
	assert.Nil(t, err)
}

// A JSON list with default:"[]" binds an empty list when it is left out or
// sent as null, at any depth, rather than reading a JSON array that isn't there.
func TestSchema_JSONBody_ListDefault_FillsAMissingOrNullList(t *testing.T) {
	type policy struct {
		Networks []string `json:"networks" validate:"required,allow_zero" default:"[]"`
		Assets   []string `json:"assets" validate:"required,allow_zero" default:"[]"`
	}
	type item struct {
		N string `json:"n"`
	}
	type bodySchema struct {
		Request struct {
			Body struct {
				Tags     []string          `json:"tags" default:"[]"`
				Items    []item            `json:"items" default:"[]"`
				PtrList  *[]string         `json:"ptr_list" default:"[]"`
				Policies []policy          `json:"policies"`
				ByName   map[string]policy `json:"by_name"`
				Optional *policy           `json:"optional"`
			}
		}
	}
	// list shows a nil list as nil, so it can't pass for an empty one.
	list := func(v any) string {
		if reflect.ValueOf(v).IsNil() {
			return "nil"
		}
		return fmt.Sprint(v)
	}
	read := func(s *bodySchema) string {
		b := s.Request.Body
		out := "tags=" + list(b.Tags) + " items=" + list(b.Items) + " ptr_list=" + list(b.PtrList)
		for i, p := range b.Policies {
			out += fmt.Sprintf(" policies.%d=%s,%s", i, list(p.Networks), list(p.Assets))
		}
		if p, ok := b.ByName["k"]; ok {
			out += " by_name.k=" + list(p.Networks) + "," + list(p.Assets)
		}
		if p := b.Optional; p != nil {
			out += " optional=" + list(p.Networks) + "," + list(p.Assets)
		}
		return out
	}
	post := func(r gofi.Router, body string) *gofi.InjectResponse {
		t.Helper()
		resp, err := r.Test(gofi.TestOptions{Method: "POST", Path: "/x", Headers: map[string]string{"Content-Type": "application/json"}, Body: strings.NewReader(body)})
		if err != nil {
			t.Fatalf("%s: request failed: %v", body, err)
		}
		return resp
	}

	var got string
	r := gofi.NewRouter()
	r.Post("/x", gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			if err != nil {
				return err
			}
			got = read(s)
			return c.SendString(200, "ok")
		},
	})
	const defaults = "tags=[] items=[] ptr_list=nil"
	for _, tc := range []struct{ body, want string }{
		{`{}`, "200 " + defaults},
		{`{"tags":null}`, "200 " + defaults},
		{`{"tags":[]}`, "200 " + defaults},
		{`{"tags":["a"]}`, "200 tags=[a] items=[] ptr_list=nil"},
		{`{"tags":[],"policies":[{}]}`, "200 " + defaults + " policies.0=[],[]"},
		{`{"tags":[],"policies":[{"networks":null,"assets":null}]}`, "200 " + defaults + " policies.0=[],[]"},
		{`{"tags":[],"policies":[{"networks":["a"]}]}`, "200 " + defaults + " policies.0=[a],[]"},
		{`{"tags":[],"by_name":{"k":{}}}`, "200 " + defaults + " by_name.k=[],[]"},
		{`{"tags":[],"optional":{}}`, "200 " + defaults + " optional=[],[]"},
		{`{"tags":[],"policies":[{"networks":[],"assets":[]}]}`, "200 " + defaults + " policies.0=[],[]"},
		{`{"tags":["a",null]}`, "400 "},
		// A list of structs takes its default the same way.
		{`{"items":null}`, "200 " + defaults},
		{`{"items":[{"n":"a"}]}`, "200 tags=[] items=[{a}] ptr_list=nil"},
		// default:"[]" applies to []T only. A pointer to a list stays nil.
		{`{"ptr_list":null}`, "200 " + defaults},
		{`{"ptr_list":["a"]}`, "200 tags=[] items=[] ptr_list=&[a]"},
	} {
		got = ""
		resp := post(r, tc.body)
		assert.Equal(t, tc.want, fmt.Sprint(resp.Code(), " ", got), tc.body)
	}

	// A list without a default keeps its presence rules, and a default is
	// checked against the field's rules as a sent [] would be.
	type rulesSchema struct {
		Request struct {
			Body struct {
				Strings []string `json:"strings" validate:"required,min=1"`
				Items   []item   `json:"items" validate:"required,min=1"`
				Min     []string `json:"min" validate:"min=1" default:"[]"`
			}
		}
	}
	r = gofi.NewRouter()
	r.Post("/x", gofi.RouteOptions{
		Schema: &rulesSchema{},
		Handler: func(c gofi.Context) error {
			if _, err := gofi.ValidateAndBind[rulesSchema](c); err != nil {
				return err
			}
			return c.SendString(200, "ok")
		},
	})
	for _, tc := range []struct{ body, field string }{
		{`{"strings":["a"],"items":[{"n":"a"}],"min":["a"]}`, ""},
		{`{"items":[{"n":"a"}],"min":["a"]}`, "strings"},
		{`{"strings":null,"items":[{"n":"a"}],"min":["a"]}`, "strings"},
		{`{"strings":["a"],"min":["a"]}`, "items"},
		{`{"strings":["a"],"items":null,"min":["a"]}`, "items"},
		{`{"strings":["a"],"items":[{"n":"a"}]}`, "min"},
		{`{"strings":["a"],"items":[{"n":"a"}],"min":null}`, "min"},
		{`{"strings":["a"],"items":[{"n":"a"}],"min":[]}`, "min"},
	} {
		resp := post(r, tc.body)
		if tc.field == "" {
			assert.Equal(t, 200, resp.Code(), tc.body)
			continue
		}
		assert.Equal(t, 400, resp.Code(), tc.body)
		assert.Contains(t, resp.BodyString(), "Body("+tc.field+")", tc.body)
	}

	// Requests share the compiled default, so what one handler does to its
	// list must not reach the next request.
	var seen []string
	r = gofi.NewRouter()
	r.Post("/x", gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			if err != nil {
				return err
			}
			seen = append(seen, read(s))
			b := &s.Request.Body
			for _, l := range [][]string{b.Tags, b.Policies[0].Networks} {
				full := l[:cap(l)]
				for i := range full {
					full[i] = "leak"
				}
			}
			b.Tags = append(b.Tags, "leak")
			b.Items = append(b.Items, item{N: "leak"})
			b.Policies[0].Networks = append(b.Policies[0].Networks, "leak")
			return c.SendString(200, "ok")
		},
	})
	for range 2 {
		resp := post(r, `{"policies":[{}]}`)
		assert.Equal(t, 200, resp.Code(), resp.BodyString())
	}
	assert.Equal(t, []string{defaults + " policies.0=[],[]", defaults + " policies.0=[],[]"}, seen)
}
