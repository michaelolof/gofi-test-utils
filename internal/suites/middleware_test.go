package suites

import (
	"net/http"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 1. Global Middleware (Use)
// =============================================================================

func TestMiddleware_GlobalUse(t *testing.T) {
	r := gofi.NewRouter()
	r.Use(func(c gofi.Context) error {
		c.Writer().Header().Set("X-Global", "applied")
		return c.Next()
	})
	r.Get("/test", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "ok") },
	})

	w := r.Test("GET", "/test")
	assert.Equal(t, "applied", w.HeaderMap.Get("X-Global"))
}

// =============================================================================
// 2. Inline Middleware (With)
// =============================================================================

func TestMiddleware_InlineWith(t *testing.T) {
	r := gofi.NewRouter()

	auth := func(c gofi.Context) error {
		c.Writer().Header().Set("X-Auth", "checked")
		return c.Next()
	}

	r.With(auth).Get("/protected", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "protected") },
	})
	r.Get("/public", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "public") },
	})

	w1 := r.Test("GET", "/protected")
	assert.Equal(t, "checked", w1.HeaderMap.Get("X-Auth"), "Protected route should have X-Auth")

	w2 := r.Test("GET", "/public")
	assert.Empty(t, w2.HeaderMap.Get("X-Auth"), "Public route should NOT have X-Auth")
}

// =============================================================================
// 3. Group Middleware Isolation
// =============================================================================

func TestMiddleware_GroupIsolation(t *testing.T) {
	r := gofi.NewRouter()

	logger := func(c gofi.Context) error {
		c.Writer().Header().Set("X-Group", "yes")
		return c.Next()
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

	w1 := r.Test("GET", "/grouped")
	gh, gth := w1.HeaderMap.Get("X-Group"), w1.HeaderMap.Get("X-Handler")
	assert.True(t, gh == "yes" || gth == "grouped", "Expected X-Group=yes or X-Handler=grouped, got X-Group=%q X-Handler=%q", gh, gth)

	w2 := r.Test("GET", "/ungrouped")
	assert.NotEqual(t, "yes", w2.HeaderMap.Get("X-Group"), "Ungrouped route should NOT have X-Group header")
}

// =============================================================================
// 4. Global Middleware (formerly PreHandler)
// =============================================================================

func TestMiddleware_GlobalMiddleware(t *testing.T) {
	r := gofi.NewRouter()
	r.Use(func(c gofi.Context) error {
		c.Writer().Header().Set("X-Pre", "yes")
		return c.Next()
	})
	r.Get("/test", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "ok") },
	})

	w := r.Test("GET", "/test")
	assert.Equal(t, "yes", w.HeaderMap.Get("X-Pre"))
}

func TestMiddleware_InlineWith_PerRoute(t *testing.T) {
	r := gofi.NewRouter()

	mw := func(c gofi.Context) error {
		c.Writer().Header().Set("X-Route-Pre", "active")
		return c.Next()
	}

	r.With(mw).Get("/with-pre", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "ok") },
	})

	w := r.Test("GET", "/with-pre")
	assert.Equal(t, "active", w.HeaderMap.Get("X-Route-Pre"))
}

// =============================================================================
// 5. Middleware Execution Order
// =============================================================================

func TestMiddleware_ExecutionOrder(t *testing.T) {
	r := gofi.NewRouter()
	var order []string

	r.Use(func(c gofi.Context) error {
		order = append(order, "global-before")
		err := c.Next()
		order = append(order, "global-after")
		return err
	})

	routeMW := func(c gofi.Context) error {
		order = append(order, "route-before")
		err := c.Next()
		order = append(order, "route-after")
		return err
	}

	r.With(routeMW).Get("/order", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			order = append(order, "handler")
			return nil
		},
	})

	r.Test("GET", "/order")

	expected := []string{"global-before", "route-before", "handler", "route-after", "global-after"}
	assert.Equal(t, expected, order)
}

// =============================================================================
// 6. Middleware Error Short-Circuit
// =============================================================================

func TestMiddleware_ErrorShortCircuit(t *testing.T) {
	r := gofi.NewRouter()
	handlerCalled := false

	r.Use(func(c gofi.Context) error {
		return http.ErrAbortHandler
	})

	r.Get("/blocked", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			handlerCalled = true
			return c.SendString(200, "should-not-reach")
		},
	})

	r.Test("GET", "/blocked")
	assert.False(t, handlerCalled, "Handler should NOT have been called after middleware error")
}

// =============================================================================
// 7. Middleware Chaining
// =============================================================================

func TestMiddleware_MultipleChained(t *testing.T) {
	r := gofi.NewRouter()
	count := 0

	for i := 0; i < 10; i++ {
		r.Use(func(c gofi.Context) error {
			count++
			return c.Next()
		})
	}

	r.Get("/chain", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "ok") },
	})

	r.Test("GET", "/chain")
	assert.Equal(t, 10, count, "Expected 10 middlewares to execute")
}

// =============================================================================
// 8. Middleware Group Isolation
// =============================================================================

func TestMiddleware_GroupIsolation_UseOnly(t *testing.T) {
	r := gofi.NewRouter()

	r.Use(func(c gofi.Context) error {
		c.Writer().Header().Set("X-Base", "true")
		return c.Next()
	})

	r.Group(func(sub gofi.Router) {
		sub.Use(func(c gofi.Context) error {
			c.Writer().Header().Set("X-Group1", "true")
			return c.Next()
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

	w1 := r.Test("GET", "/g1")
	assert.Equal(t, "true", w1.HeaderMap.Get("X-Base"), "g1 should have base middleware")
	assert.Equal(t, "true", w1.HeaderMap.Get("X-Group1"), "g1 should have group1 middleware")

	w2 := r.Test("GET", "/g2")
	assert.Equal(t, "true", w2.HeaderMap.Get("X-Base"), "g2 should have base middleware")
	assert.Empty(t, w2.HeaderMap.Get("X-Group1"), "g2 should NOT have group1 middleware")
}
