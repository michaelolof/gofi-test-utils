package suites

import (
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 1. HTTP Method Routing
// =============================================================================

func TestRouting_AllHTTPMethods(t *testing.T) {
	methods := []struct {
		method   string
		register func(r gofi.Router, path string, opts gofi.RouteOptions)
	}{
		{"GET", func(r gofi.Router, p string, o gofi.RouteOptions) { r.Get(p, o) }},
		{"POST", func(r gofi.Router, p string, o gofi.RouteOptions) { r.Post(p, o) }},
		{"PUT", func(r gofi.Router, p string, o gofi.RouteOptions) { r.Put(p, o) }},
		{"DELETE", func(r gofi.Router, p string, o gofi.RouteOptions) { r.Delete(p, o) }},
		{"PATCH", func(r gofi.Router, p string, o gofi.RouteOptions) { r.Patch(p, o) }},
		{"HEAD", func(r gofi.Router, p string, o gofi.RouteOptions) { r.Head(p, o) }},
		{"OPTIONS", func(r gofi.Router, p string, o gofi.RouteOptions) { r.Options(p, o) }},
	}

	for _, tc := range methods {
		t.Run(tc.method, func(t *testing.T) {
			r := gofi.NewRouter()
			tc.register(r, "/test", gofi.RouteOptions{
				Handler: func(c gofi.Context) error {
					return c.SendString(200, tc.method+"-ok")
				},
			})

			w := r.Test(tc.method, "/test")
			assert.Equal(t, 200, w.StatusCode, "Expected 200 for %s", tc.method)
		})
	}
}

func TestRouting_MethodFunc(t *testing.T) {
	r := gofi.NewRouter()
	r.Method("GET", "/method-test", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "method-ok")
		},
	})

	w := r.Test("GET", "/method-test")
	assert.Equal(t, 200, w.StatusCode)
	assert.Equal(t, "method-ok", string(w.Body))
}

// =============================================================================
// 2. Static Routes
// =============================================================================

func TestRouting_StaticRoute(t *testing.T) {
	r := gofi.NewRouter()
	r.Get("/", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "root") },
	})
	r.Get("/health", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "healthy") },
	})

	tests := []struct {
		path, body string
	}{
		{"/", "root"},
		{"/health", "healthy"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			w := r.Test("GET", tc.path)
			assert.Equal(t, 200, w.StatusCode)
			assert.Equal(t, tc.body, string(w.Body))
		})
	}
}

// =============================================================================
// 3. Parameterized Routes
// =============================================================================

func TestRouting_SingleParam(t *testing.T) {
	r := gofi.NewRouter()
	r.Get("/user/:name", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			name := c.Param("name")
			return c.SendString(200, "user:"+name)
		},
	})

	w := r.Test("GET", "/user/alice")
	assert.Equal(t, "user:alice", string(w.Body))
}

func TestRouting_MultipleParams(t *testing.T) {
	r := gofi.NewRouter()
	r.Get("/users/:userId/posts/:postId", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			uid := c.Param("userId")
			pid := c.Param("postId")
			return c.SendString(200, uid+":"+pid)
		},
	})

	w := r.Test("GET", "/users/42/posts/99")
	assert.Equal(t, "42:99", string(w.Body))
}

func TestRouting_WildcardRoute(t *testing.T) {
	r := gofi.NewRouter()
	r.Get("/files/*path", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			p := c.Param("path")
			return c.SendString(200, "file:"+p)
		},
	})

	w := r.Test("GET", "/files/images/logo.png")
	assert.Equal(t, "file:images/logo.png", string(w.Body))
}

// =============================================================================
// 4. 404 Handling
// =============================================================================

func TestRouting_404(t *testing.T) {
	r := gofi.NewRouter()
	r.Get("/exists", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "ok") },
	})

	w := r.Test("GET", "/does-not-exist")
	assert.Equal(t, 404, w.StatusCode)
}

// =============================================================================
// 5. Route Nesting & Groups
// =============================================================================

func TestRouting_RouteNesting(t *testing.T) {
	r := gofi.NewRouter()

	r.Route("/api", func(api gofi.Router) {
		api.Get("/users", gofi.RouteOptions{
			Handler: func(c gofi.Context) error { return c.SendString(200, "users-list") },
		})

		api.Route("/v1", func(v1 gofi.Router) {
			v1.Get("/posts", gofi.RouteOptions{
				Handler: func(c gofi.Context) error { return c.SendString(200, "v1-posts") },
			})
		})
	})

	tests := []struct {
		path, body string
	}{
		{"/api/users", "users-list"},
		{"/api/v1/posts", "v1-posts"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			w := r.Test("GET", tc.path)
			assert.Equal(t, 200, w.StatusCode)
			assert.Equal(t, tc.body, string(w.Body))
		})
	}
}

func TestRouting_DeepNesting(t *testing.T) {
	r := gofi.NewRouter()
	r.Get("/v1/api/deep/nested/resource/action", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "deep-ok") },
	})

	w := r.Test("GET", "/v1/api/deep/nested/resource/action")
	assert.Equal(t, "deep-ok", string(w.Body))
}

// =============================================================================
// 6. Multiple Routes Same Path Different Methods
// =============================================================================

func TestRouting_SamePathDifferentMethods(t *testing.T) {
	r := gofi.NewRouter()
	r.Get("/resource", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "get-resource") },
	})
	r.Post("/resource", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(201, "post-resource") },
	})
	r.Delete("/resource", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "delete-resource") },
	})

	tests := []struct {
		method string
		code   int
		body   string
	}{
		{"GET", 200, "get-resource"},
		{"POST", 201, "post-resource"},
		{"DELETE", 200, "delete-resource"},
	}

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			w := r.Test(tc.method, "/resource")
			assert.Equal(t, tc.code, w.StatusCode)
			assert.Equal(t, tc.body, string(w.Body))
		})
	}
}
