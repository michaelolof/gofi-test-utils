package suites

import (
	"io"
	"net/http"
	"net/http/httptest"
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
			r := gofi.NewServeMux()
			tc.register(r, "/test", gofi.RouteOptions{
				Handler: func(c gofi.Context) error {
					return c.SendString(200, tc.method+"-ok")
				},
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/test", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code, "Expected 200 for %s", tc.method)
		})
	}
}

func TestRouting_MethodFunc(t *testing.T) {
	r := gofi.NewServeMux()
	r.Method("GET", "/method-test", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "method-ok")
		},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/method-test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "method-ok", w.Body.String())
}

// =============================================================================
// 2. Static Routes
// =============================================================================

func TestRouting_StaticRoute(t *testing.T) {
	r := gofi.NewServeMux()
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
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tc.path, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
			assert.Equal(t, tc.body, w.Body.String())
		})
	}
}

// =============================================================================
// 3. Parameterized Routes
// =============================================================================

func TestRouting_SingleParam(t *testing.T) {
	r := gofi.NewServeMux()
	r.Get("/user/{name}", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			name := c.Request().PathValue("name")
			return c.SendString(200, "user:"+name)
		},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/user/alice", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, "user:alice", w.Body.String())
}

func TestRouting_MultipleParams(t *testing.T) {
	r := gofi.NewServeMux()
	r.Get("/users/{userId}/posts/{postId}", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			uid := c.Request().PathValue("userId")
			pid := c.Request().PathValue("postId")
			return c.SendString(200, uid+":"+pid)
		},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users/42/posts/99", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, "42:99", w.Body.String())
}

func TestRouting_WildcardRoute(t *testing.T) {
	r := gofi.NewServeMux()
	r.Get("/files/{path...}", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			p := c.Request().PathValue("path")
			return c.SendString(200, "file:"+p)
		},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/files/images/logo.png", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, "file:images/logo.png", w.Body.String())
}

// =============================================================================
// 4. 404 Handling
// =============================================================================

func TestRouting_404(t *testing.T) {
	r := gofi.NewServeMux()
	r.Get("/exists", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "ok") },
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/does-not-exist", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
}

// =============================================================================
// 5. Route Nesting & Groups
// =============================================================================

func TestRouting_RouteNesting(t *testing.T) {
	r := gofi.NewServeMux()

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
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tc.path, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
			assert.Equal(t, tc.body, w.Body.String())
		})
	}
}

func TestRouting_DeepNesting(t *testing.T) {
	r := gofi.NewServeMux()
	r.Get("/v1/api/deep/nested/resource/action", gofi.RouteOptions{
		Handler: func(c gofi.Context) error { return c.SendString(200, "deep-ok") },
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/api/deep/nested/resource/action", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, "deep-ok", w.Body.String())
}

// =============================================================================
// 6. Handle / HandleFunc
// =============================================================================

func TestRouting_HandleFunc(t *testing.T) {
	r := gofi.NewServeMux()
	r.HandleFunc("GET /plain", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		io.WriteString(w, "plain-handler")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/plain", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, "plain-handler", w.Body.String())
}

// =============================================================================
// 7. Multiple Routes Same Path Different Methods
// =============================================================================

func TestRouting_SamePathDifferentMethods(t *testing.T) {
	r := gofi.NewServeMux()
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
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/resource", nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tc.code, w.Code)
			assert.Equal(t, tc.body, w.Body.String())
		})
	}
}
