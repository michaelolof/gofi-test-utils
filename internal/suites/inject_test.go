package suites

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 1. Basic Inject
// =============================================================================

func TestInject_BasicGET(t *testing.T) {
	handler := gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "injected")
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
	assert.Equal(t, "injected", string(rec.Body))
}

// =============================================================================
// 2. Inject with Path Params
// =============================================================================

func TestInject_PathParams(t *testing.T) {
	type schema struct {
		Request struct {
			Path struct {
				UserID string `json:"userId" validate:"required"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			return c.SendString(200, "user:"+s.Request.Path.UserID)
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/users/{userId}",
		Paths:   map[string]string{"userId": "abc-123"},
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, "user:abc-123", string(rec.Body))
}

// =============================================================================
// 3. Inject with Query Params
// =============================================================================

func TestInject_QueryParams(t *testing.T) {
	type schema struct {
		Request struct {
			Query struct {
				Search string `json:"search" validate:"required"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			return c.SendString(200, "search:"+s.Request.Query.Search)
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/search",
		Query:   map[string]string{"search": "golang"},
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, "search:golang", string(rec.Body))
}

// =============================================================================
// 4. Inject with Headers
// =============================================================================

func TestInject_Headers(t *testing.T) {
	type schema struct {
		Request struct {
			Header struct {
				Token string `json:"X-Token" validate:"required"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			return c.SendString(200, "token:"+s.Request.Header.Token)
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/auth",
		Headers: map[string]string{"X-Token": "secret-123"},
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, "token:secret-123", string(rec.Body))
}

// =============================================================================
// 5. Inject with Cookies
// =============================================================================

func TestInject_Cookies(t *testing.T) {
	type schema struct {
		Request struct {
			Cookie struct {
				Session string `json:"session" validate:"required"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			return c.SendString(200, "session:"+s.Request.Cookie.Session)
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/profile",
		Cookies: []http.Cookie{
			{Name: "session", Value: "sess-xyz"},
		},
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, "session:sess-xyz", string(rec.Body))
}

// =============================================================================
// 6. Inject with Body
// =============================================================================

func TestInject_Body(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Name string `json:"name" validate:"required"`
			} `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			return c.SendString(200, "name:"+s.Request.Body.Name)
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`{"name":"Bob"}`),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, "name:Bob", string(rec.Body))
}

// =============================================================================
// 7. Inject with Middleware
// =============================================================================

func TestInject_WithMiddleware(t *testing.T) {
	m := gofi.NewRouter()
	m.Use(func(c gofi.Context) error {
		c.Writer().Header().Set("X-Injected-Pre", "yes")
		return c.Next()
	})

	handler := gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "ok")
		},
	}

	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, "yes", rec.HeaderMap.Get("X-Injected-Pre"))
}

// =============================================================================
// 8. Inject with Combined All Parts
// =============================================================================

func TestInject_AllParts(t *testing.T) {
	type schema struct {
		Request struct {
			Path struct {
				ID string `json:"id" validate:"required"`
			}
			Query struct {
				Verbose bool `json:"verbose"`
			}
			Header struct {
				Auth string `json:"Authorization" validate:"required"`
			}
			Cookie struct {
				Sess string `json:"sess" validate:"required"`
			}
			Body struct {
				Data string `json:"data" validate:"required"`
			} `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			assert.Equal(t, "42", s.Request.Path.ID)
			assert.True(t, s.Request.Query.Verbose)
			assert.Equal(t, "Bearer tok", s.Request.Header.Auth)
			assert.Equal(t, "s123", s.Request.Cookie.Sess)
			assert.Equal(t, "payload", s.Request.Body.Data)
			return c.SendString(200, "all-ok")
		},
	}

	m := gofi.NewRouter()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/items/{id}",
		Paths:   map[string]string{"id": "42"},
		Query:   map[string]string{"verbose": "true"},
		Headers: map[string]string{"Authorization": "Bearer tok"},
		Cookies: []http.Cookie{{Name: "sess", Value: "s123"}},
		Body:    strings.NewReader(`{"data":"payload"}`),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
	assert.Equal(t, "all-ok", string(rec.Body))
}

// =============================================================================
// 9. Inject panic uses router error handler
// =============================================================================

func TestInject_PanicUsesErrorHandler(t *testing.T) {
	m := gofi.NewRouter()

	var capturedErr error
	m.UseErrorHandler(func(err error, c gofi.Context) {
		capturedErr = err
		_ = c.SendString(555, "inject:"+err.Error())
	})

	handler := gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			panic("inject boom")
		},
	}

	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/panic",
		Handler: &handler,
	})
	assert.NotNil(t, err)
	assert.Equal(t, 555, rec.StatusCode)
	assert.Equal(t, "inject:panic recovered in Inject: inject boom", string(rec.Body))
	assert.NotNil(t, capturedErr)
	assert.Equal(t, "panic recovered in Inject: inject boom", capturedErr.Error())

	var httpErr *gofi.HTTPError
	assert.ErrorAs(t, err, &httpErr)
	if assert.NotNil(t, httpErr) {
		assert.Equal(t, 500, httpErr.Code)
	}
}

func TestInject_PanicDefaultErrorHandler(t *testing.T) {
	m := gofi.NewRouter()

	handler := gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			panic("default inject panic")
		},
	}

	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/panic-default",
		Handler: &handler,
	})
	assert.NotNil(t, err)
	assert.Equal(t, 500, rec.StatusCode)

	var resp struct {
		Status     string `json:"status"`
		StatusCode int    `json:"statusCode"`
		Message    string `json:"message"`
	}
	decodeErr := json.Unmarshal(rec.Body, &resp)
	assert.Nil(t, decodeErr)
	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, "panic recovered in Inject: default inject panic", resp.Message)
}
