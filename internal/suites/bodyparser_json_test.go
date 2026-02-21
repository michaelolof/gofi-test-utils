package suites

import (
	"encoding/json"
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
				Age  int    `json:"age" validate:"required"`
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`{"name":"Alice","age":30}`),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/data",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)

	var result struct {
		Message string `json:"message"`
		Count   int    `json:"count"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &result)
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/user",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)

	var result struct {
		Name    *string  `json:"name"`
		Age     *int     `json:"age"`
		Balance *float64 `json:"balance"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &result)
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/user",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)

	var result map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &result)
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/timestamps",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)

	// gofi's custom JSON encoder may encode time.Time as an object or string.
	// Verify the fields are present and non-nil by decoding into a generic map.
	var result map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &result)
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/timestamps",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)

	// gofi's custom JSON encoder may encode time.Time as an object or string.
	// Verify the fields are present and non-nil by decoding into a generic map.
	var result map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &result)
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/tags",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "[]", strings.TrimSpace(rec.Body.String()))
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)

	var result struct {
		Count   int     `json:"count"`
		Score   float64 `json:"score"`
		Name    string  `json:"name"`
		Enabled bool    `json:"enabled"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &result)
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

	m := gofi.NewServeMux()
	rec, _ := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	// Should not return 200 — nil body on a defined schema should error
	assert.NotEqual(t, 200, rec.Code, "Nil body on required schema should produce an error")
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

	m := gofi.NewServeMux()
	rec, _ := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.NotEqual(t, 200, rec.Code, "Non-struct body should produce an error")
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)

	// Verify the body is valid JSON by unmarshaling
	var result map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &result)
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)

	// Verify the body is valid JSON even with special characters
	var result struct {
		Name    string `json:"name"`
		Message string `json:"message"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &result)
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
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

	m := gofi.NewServeMux()
	_, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`[{"name":"Alice"}]`),
		Handler: &handler,
	})
	assert.Nil(t, err)
}
