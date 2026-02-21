package suites

import (
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "injected", rec.Body.String())
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/users/{userId}",
		Paths:   map[string]string{"userId": "abc-123"},
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, "user:abc-123", rec.Body.String())
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/search",
		Query:   map[string]string{"search": "golang"},
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, "search:golang", rec.Body.String())
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/auth",
		Headers: map[string]string{"X-Token": "secret-123"},
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, "token:secret-123", rec.Body.String())
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/profile",
		Cookies: []http.Cookie{
			{Name: "session", Value: "sess-xyz"},
		},
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, "session:sess-xyz", rec.Body.String())
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/users",
		Body:    strings.NewReader(`{"name":"Bob"}`),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, "name:Bob", rec.Body.String())
}

// =============================================================================
// 7. Inject with PreHandlers
// =============================================================================

func TestInject_WithPreHandlers(t *testing.T) {
	m := gofi.NewServeMux()
	m.UsePreHandler(func(next gofi.HandlerFunc) gofi.HandlerFunc {
		return func(c gofi.Context) error {
			c.Writer().Header().Set("X-Injected-Pre", "yes")
			return next(c)
		}
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
	assert.Equal(t, "yes", rec.Header().Get("X-Injected-Pre"))
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

	m := gofi.NewServeMux()
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
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "all-ok", rec.Body.String())
}
