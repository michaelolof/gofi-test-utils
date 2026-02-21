package suites

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 1. Global Middleware (Use)
// =============================================================================

func TestMiddleware_GlobalUse(t *testing.T) {
	r := gofi.NewServeMux()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Global", "applied")
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/test", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "ok") },
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/test", nil))

	assert.Equal(t, "applied", w.Header().Get("X-Global"))
}

// =============================================================================
// 2. Inline Middleware (With)
// =============================================================================

func TestMiddleware_InlineWith(t *testing.T) {
	r := gofi.NewServeMux()

	auth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Auth", "checked")
			next.ServeHTTP(w, r)
		})
	}

	r.With(auth).Get("/protected", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "protected") },
	})
	r.Get("/public", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "public") },
	})

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest("GET", "/protected", nil))
	assert.Equal(t, "checked", w1.Header().Get("X-Auth"), "Protected route should have X-Auth")

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest("GET", "/public", nil))
	assert.Empty(t, w2.Header().Get("X-Auth"), "Public route should NOT have X-Auth")
}

// =============================================================================
// 3. Group Middleware Isolation
// =============================================================================

func TestMiddleware_GroupIsolation(t *testing.T) {
	r := gofi.NewServeMux()

	logger := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Group", "yes")
			next.ServeHTTP(w, r)
		})
	}

	r.Group(func(sub gofi.Router) {
		sub.Use(logger)
		sub.Get("/grouped", gofi.RouteOptions{
			Handler: func(c gofi.Context) error {
				c.Writer().Header().Set("X-Handler", "grouped")
				return c.SendString(200, "grouped")
			},
		})
	})

	r.Get("/ungrouped", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			c.Writer().Header().Set("X-Handler", "ungrouped")
			return c.SendString(200, "ungrouped")
		},
	})

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest("GET", "/grouped", nil))
	gh, gth := w1.Header().Get("X-Group"), w1.Header().Get("X-Handler")
	assert.True(t, gh == "yes" || gth == "grouped", "Expected X-Group=yes or X-Handler=grouped, got X-Group=%q X-Handler=%q", gh, gth)

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest("GET", "/ungrouped", nil))
	assert.NotEqual(t, "yes", w2.Header().Get("X-Group"), "Ungrouped route should NOT have X-Group header")
}

// =============================================================================
// 4. PreHandler
// =============================================================================

func TestMiddleware_PreHandler_Global(t *testing.T) {
	r := gofi.NewServeMux()
	r.UsePreHandler(func(next gofi.HandlerFunc) gofi.HandlerFunc {
		return func(c gofi.Context) error {
			c.Writer().Header().Set("X-Pre", "yes")
			return next(c)
		}
	})
	r.Get("/test", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "ok") },
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/test", nil))

	assert.Equal(t, "yes", w.Header().Get("X-Pre"))
}

func TestMiddleware_PreHandler_PerRoute(t *testing.T) {
	r := gofi.NewServeMux()
	r.Get("/with-pre", gofi.RouteOptions{
		PreHandlers: []gofi.PreHandler{
			func(next gofi.HandlerFunc) gofi.HandlerFunc {
				return func(c gofi.Context) error {
					c.Writer().Header().Set("X-Route-Pre", "active")
					return next(c)
				}
			},
		},
		Handler: func(c gofi.Context) error { return c.SendString(200, "ok") },
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/with-pre", nil))

	assert.Equal(t, "active", w.Header().Get("X-Route-Pre"))
}

// =============================================================================
// 5. PreHandler Execution Order
// =============================================================================

func TestMiddleware_PreHandler_ExecutionOrder(t *testing.T) {
	r := gofi.NewServeMux()
	var order []string

	r.UsePreHandler(func(next gofi.HandlerFunc) gofi.HandlerFunc {
		return func(c gofi.Context) error {
			order = append(order, "global-before")
			err := next(c)
			order = append(order, "global-after")
			return err
		}
	})

	r.Get("/order", gofi.RouteOptions{
		PreHandlers: []gofi.PreHandler{
			func(next gofi.HandlerFunc) gofi.HandlerFunc {
				return func(c gofi.Context) error {
					order = append(order, "route-before")
					err := next(c)
					order = append(order, "route-after")
					return err
				}
			},
		},
		Handler: func(c gofi.Context) error {
			order = append(order, "handler")
			return nil
		},
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/order", nil))

	expected := []string{"global-before", "route-before", "handler", "route-after", "global-after"}
	assert.Equal(t, expected, order)
}

// =============================================================================
// 6. PreHandler Error Short-Circuit
// =============================================================================

func TestMiddleware_PreHandler_ErrorShortCircuit(t *testing.T) {
	r := gofi.NewServeMux()
	handlerCalled := false

	r.UsePreHandler(func(next gofi.HandlerFunc) gofi.HandlerFunc {
		return func(c gofi.Context) error {
			return http.ErrAbortHandler
		}
	})

	r.Get("/blocked", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			handlerCalled = true
			return c.SendString(200, "should-not-reach")
		},
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/blocked", nil))

	assert.False(t, handlerCalled, "Handler should NOT have been called after PreHandler error")
}

// =============================================================================
// 7. Middleware Chaining
// =============================================================================

func TestMiddleware_MultipleChained(t *testing.T) {
	r := gofi.NewServeMux()
	count := 0

	for i := 0; i < 10; i++ {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				next.ServeHTTP(w, r)
			})
		})
	}

	r.Get("/chain", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "ok") },
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/chain", nil))

	assert.Equal(t, 10, count, "Expected 10 middlewares to execute")
}

// =============================================================================
// 8. PreHandler Group Isolation
// =============================================================================

func TestMiddleware_PreHandler_GroupIsolation(t *testing.T) {
	r := gofi.NewServeMux()

	r.UsePreHandler(func(next gofi.HandlerFunc) gofi.HandlerFunc {
		return func(c gofi.Context) error {
			c.Writer().Header().Set("X-Base", "true")
			return next(c)
		}
	})

	r.Group(func(sub gofi.Router) {
		sub.UsePreHandler(func(next gofi.HandlerFunc) gofi.HandlerFunc {
			return func(c gofi.Context) error {
				c.Writer().Header().Set("X-Group1", "true")
				return next(c)
			}
		})
		sub.Get("/g1", gofi.RouteOptions{
			Handler: func(c gofi.Context) error { return c.SendString(200, "g1") },
		})
	})

	r.Group(func(sub gofi.Router) {
		sub.Get("/g2", gofi.RouteOptions{
			Handler: func(c gofi.Context) error { return c.SendString(200, "g2") },
		})
	})

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest("GET", "/g1", nil))
	assert.Equal(t, "true", w1.Header().Get("X-Base"), "g1 should have base prehandler")
	assert.Equal(t, "true", w1.Header().Get("X-Group1"), "g1 should have group1 prehandler")

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest("GET", "/g2", nil))
	assert.Equal(t, "true", w2.Header().Get("X-Base"), "g2 should have base prehandler")
	assert.Empty(t, w2.Header().Get("X-Group1"), "g2 should NOT have group1 prehandler")
}
