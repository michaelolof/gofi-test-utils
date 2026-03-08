package suites

import (
	"net/http"
	"strings"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 1. Path Parameter Binding & Validation
// =============================================================================

func TestSchema_PathBinding(t *testing.T) {
	type pathSchema struct {
		Request struct {
			Path struct {
				ID       int     `json:"id" validate:"required"`
				Category string  `json:"category" validate:"required"`
				Rating   float64 `json:"rating"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &pathSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[pathSchema](c)
			assert.Nil(t, err)
			assert.Equal(t, 42, s.Request.Path.ID)
			assert.Equal(t, "books", s.Request.Path.Category)
			assert.Equal(t, 4.5, s.Request.Path.Rating)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/items/{category}/{id}",
		Paths: map[string]string{
			"id":       "42",
			"category": "books",
			"rating":   "4.5",
		},
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}

func TestSchema_PathParam_ValidationError(t *testing.T) {
	type pathSchema struct {
		Request struct {
			Path struct {
				ID string `json:"id" validate:"required"`
			}
		}
	}

	var gotErr error
	handler := gofi.RouteOptions{
		Schema: &pathSchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[pathSchema](c)
			gotErr = err
			return err
		},
	}

	m := gofi.NewRouter()
	m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/items/{id}",
		Paths:   map[string]string{},
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "Expected validation error for missing required path param")
}

// =============================================================================
// 2. Query Parameter Binding
// =============================================================================

func TestSchema_QueryBinding(t *testing.T) {
	type querySchema struct {
		Request struct {
			Query struct {
				Page   int    `json:"page" default:"1"`
				Sort   string `json:"sort"`
				Active bool   `json:"active"`
			}
		}
	}

	t.Run("AllProvided", func(t *testing.T) {
		handler := gofi.RouteOptions{
			Schema: &querySchema{},
			Handler: func(c gofi.Context) error {
				s, err := gofi.ValidateAndBind[querySchema](c)
				assert.Nil(t, err)
				assert.Equal(t, 3, s.Request.Query.Page)
				assert.Equal(t, "name", s.Request.Query.Sort)
				assert.True(t, s.Request.Query.Active)
				return c.SendString(200, "ok")
			},
		}

		m := gofi.NewRouter()
		_, err := m.Inject(gofi.InjectOptions{
			Method: "GET",
			Path:   "/search",
			Query: map[string]string{
				"page":   "3",
				"sort":   "name",
				"active": "true",
			},
			Handler: &handler,
		})
		assert.Nil(t, err)
	})

	t.Run("DefaultValues", func(t *testing.T) {
		handler := gofi.RouteOptions{
			Schema: &querySchema{},
			Handler: func(c gofi.Context) error {
				s, err := gofi.ValidateAndBind[querySchema](c)
				assert.Nil(t, err)
				assert.Equal(t, 1, s.Request.Query.Page)
				return c.SendString(200, "ok")
			},
		}

		m := gofi.NewRouter()
		_, err := m.Inject(gofi.InjectOptions{
			Method:  "GET",
			Path:    "/search",
			Handler: &handler,
		})
		assert.Nil(t, err)
	})
}

// =============================================================================
// 3. Header Binding
// =============================================================================

func TestSchema_HeaderBinding(t *testing.T) {
	type headerSchema struct {
		Request struct {
			Header struct {
				RequestID string `json:"X-Request-Id" validate:"required"`
				Attempts  int    `json:"X-Attempts" default:"1"`
				IsDebug   bool   `json:"X-Debug"`
			}
		}
	}

	t.Run("AllProvided", func(t *testing.T) {
		handler := gofi.RouteOptions{
			Schema: &headerSchema{},
			Handler: func(c gofi.Context) error {
				s, err := gofi.ValidateAndBind[headerSchema](c)
				assert.Nil(t, err)
				assert.Equal(t, "req-abc", s.Request.Header.RequestID)
				assert.Equal(t, 5, s.Request.Header.Attempts)
				assert.True(t, s.Request.Header.IsDebug)
				return c.SendString(200, "ok")
			},
		}

		m := gofi.NewRouter()
		_, err := m.Inject(gofi.InjectOptions{
			Method: "GET",
			Path:   "/test",
			Headers: map[string]string{
				"X-Request-Id": "req-abc",
				"X-Attempts":   "5",
				"X-Debug":      "true",
			},
			Handler: &handler,
		})
		assert.Nil(t, err)
	})

	t.Run("DefaultHeader", func(t *testing.T) {
		handler := gofi.RouteOptions{
			Schema: &headerSchema{},
			Handler: func(c gofi.Context) error {
				s, err := gofi.ValidateAndBind[headerSchema](c)
				assert.Nil(t, err)
				assert.Equal(t, 1, s.Request.Header.Attempts)
				return c.SendString(200, "ok")
			},
		}

		m := gofi.NewRouter()
		_, err := m.Inject(gofi.InjectOptions{
			Method: "GET",
			Path:   "/test",
			Headers: map[string]string{
				"X-Request-Id": "req-xyz",
			},
			Handler: &handler,
		})
		assert.Nil(t, err)
	})

	t.Run("MissingRequired", func(t *testing.T) {
		var gotErr error
		handler := gofi.RouteOptions{
			Schema: &headerSchema{},
			Handler: func(c gofi.Context) error {
				_, err := gofi.ValidateAndBind[headerSchema](c)
				gotErr = err
				return nil
			},
		}

		m := gofi.NewRouter()
		m.Inject(gofi.InjectOptions{
			Method:  "GET",
			Path:    "/test",
			Headers: map[string]string{},
			Handler: &handler,
		})

		assert.NotNil(t, gotErr, "Expected validation error for missing required header")
	})
}

// =============================================================================
// 4. Cookie Binding
// =============================================================================

func TestSchema_CookieBinding(t *testing.T) {
	type cookieSchema struct {
		Request struct {
			Cookie struct {
				SessionID string      `json:"session_id" validate:"required"`
				Tracking  http.Cookie `json:"tracking"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &cookieSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[cookieSchema](c)
			assert.Nil(t, err)
			assert.Equal(t, "abc-123", s.Request.Cookie.SessionID)
			assert.Equal(t, "tracking", s.Request.Cookie.Tracking.Name)
			assert.Equal(t, "on", s.Request.Cookie.Tracking.Value)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/test",
		Cookies: []http.Cookie{
			{Name: "session_id", Value: "abc-123"},
			{Name: "tracking", Value: "on"},
		},
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}

// =============================================================================
// 5. Validate Only (no binding)
// =============================================================================

func TestSchema_ValidateOnly(t *testing.T) {
	type schema struct {
		Request struct {
			Path struct {
				ID string `json:"id" validate:"required"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			err := gofi.Validate(c)
			assert.Nil(t, err)
			return c.SendString(200, "validated")
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/items/{id}",
		Paths:   map[string]string{"id": "abc"},
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}

// =============================================================================
// 7. Combined Request Parts
// =============================================================================

func TestSchema_CombinedBinding(t *testing.T) {
	type combined struct {
		Request struct {
			Path struct {
				ID int `json:"id" validate:"required"`
			}
			Query struct {
				Format string `json:"format" default:"json"`
			}
			Header struct {
				Auth string `json:"Authorization" validate:"required"`
			}
			Body struct {
				Title string `json:"title" validate:"required"`
			} `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &combined{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[combined](c)
			assert.Nil(t, err)
			assert.Equal(t, 7, s.Request.Path.ID)
			assert.Equal(t, "xml", s.Request.Query.Format)
			assert.Equal(t, "Bearer tok123", s.Request.Header.Auth)
			assert.Equal(t, "Hello", s.Request.Body.Title)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/items/{id}",
		Paths:   map[string]string{"id": "7"},
		Query:   map[string]string{"format": "xml"},
		Headers: map[string]string{"Authorization": "Bearer tok123"},
		Body:    strings.NewReader(`{"title":"Hello"}`),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}

// =============================================================================
// 8. Empty/Zero Values — Required Fields Should Fail
// =============================================================================

func TestSchema_EmptyValues_RequiredQueryString(t *testing.T) {
	type schema struct {
		Request struct {
			Query struct {
				Name string `json:"name" validate:"required"`
			}
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
		Method:  "GET",
		Path:    "/search",
		Query:   map[string]string{"name": ""},
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "Empty string should fail required validation")
}

func TestSchema_EmptyValues_RequiredQueryInt_Zero(t *testing.T) {
	type schema struct {
		Request struct {
			Query struct {
				Count int `json:"count" validate:"required"`
			}
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
		Method:  "GET",
		Path:    "/items",
		Query:   map[string]string{},
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "Missing required int query param should fail")
}

func TestSchema_EmptyValues_RequiredHeader_EmptyString(t *testing.T) {
	type schema struct {
		Request struct {
			Header struct {
				Token string `json:"X-Token" validate:"required"`
			}
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
		Method:  "GET",
		Path:    "/test",
		Headers: map[string]string{"X-Token": ""},
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "Empty string header should fail required validation")
}

func TestSchema_EmptyValues_RequiredCookie_Missing(t *testing.T) {
	type schema struct {
		Request struct {
			Cookie struct {
				Session string `json:"session" validate:"required"`
			}
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
		Method:  "GET",
		Path:    "/test",
		Cookies: []http.Cookie{},
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "Missing required cookie should fail validation")
}

func TestSchema_EmptyValues_RequiredBody_EmptyPayload(t *testing.T) {
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
			return nil
		},
	}

	m := gofi.NewRouter()
	m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(""),
		Handler: &handler,
	})

	assert.NotNil(t, gotErr, "Empty body should fail required validation")
}

// =============================================================================
// 9. Empty/Zero Values — Non-Required Fields Should Pass
// =============================================================================

func TestSchema_EmptyValues_OptionalQuery_ZeroValues(t *testing.T) {
	type schema struct {
		Request struct {
			Query struct {
				Page   int    `json:"page"`
				Sort   string `json:"sort"`
				Active bool   `json:"active"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			assert.Equal(t, 0, s.Request.Query.Page)
			assert.Equal(t, "", s.Request.Query.Sort)
			assert.False(t, s.Request.Query.Active)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/items",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}

func TestSchema_EmptyValues_OptionalHeader_Missing(t *testing.T) {
	type schema struct {
		Request struct {
			Header struct {
				Tag   string `json:"X-Tag"`
				Debug bool   `json:"X-Debug"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			assert.Equal(t, "", s.Request.Header.Tag)
			assert.False(t, s.Request.Header.Debug)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Headers: map[string]string{},
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}

func TestSchema_EmptyValues_OptionalBody_NilPayload(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Name string `json:"name"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			assert.Equal(t, "", s.Request.Body.Name)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}

// NOTE: Sending an empty JSON array `[]` as an optional body causes a panic
// in gofi's body parser (index out of range). This is a known gofi library bug.
func TestSchema_EmptyValues_OptionalBody_EmptyArray(t *testing.T) {
	type schema struct {
		Request struct {
			Body []string `validate:""`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			assert.Empty(t, s.Request.Body)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/tags",
		Body:    strings.NewReader(`[]`),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}
