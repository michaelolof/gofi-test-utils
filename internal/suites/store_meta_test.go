package suites

import (
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 1. Global Store
// =============================================================================

func TestStore_GlobalStore_SetGet(t *testing.T) {
	r := gofi.NewRouter()
	r.GlobalStore().Set("dbHost", "localhost:5432")
	r.GlobalStore().Set("appName", "test-app")

	r.Get("/check", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			host, ok := c.GlobalStore().Get("dbHost")
			assert.True(t, ok)
			assert.Equal(t, "localhost:5432", host)

			name, ok := c.GlobalStore().Get("appName")
			assert.True(t, ok)
			assert.Equal(t, "test-app", name)

			assert.True(t, c.GlobalStore().Has("dbHost"))
			assert.False(t, c.GlobalStore().Has("nonexistent"))
			return c.SendString(200, "ok")
		},
	})

	w := mustTest(t, r, "GET", "/check")
	assert.Equal(t, 200, w.StatusCode)
}

func TestStore_GlobalStore_TryGet(t *testing.T) {
	r := gofi.NewRouter()
	r.GlobalStore().Set("key", "value")

	r.Get("/tryget", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			val := c.GlobalStore().TryGet("key")
			assert.Equal(t, "value", val)
			return c.SendString(200, "ok")
		},
	})

	mustTest(t, r, "GET", "/tryget")
}

func TestStore_GlobalStore_TryGet_Panics(t *testing.T) {
	r := gofi.NewRouter()

	r.Get("/tryget-panic", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			assert.Panics(t, func() {
				c.GlobalStore().TryGet("nonexistent")
			})
			return c.SendString(200, "ok")
		},
	})

	mustTest(t, r, "GET", "/tryget-panic")
}

// =============================================================================
// 2. DataStore (per-request)
// =============================================================================

func TestStore_DataStore_PerRequest(t *testing.T) {
	r := gofi.NewRouter()

	r.Use(func(c gofi.Context) error {
		c.DataStore().Set("requestID", "req-1234")
		return c.Next()
	})

	r.Get("/data", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			id, ok := c.DataStore().Get("requestID")
			assert.True(t, ok)
			assert.Equal(t, "req-1234", id)
			return c.SendString(200, "ok")
		},
	})

	w := mustTest(t, r, "GET", "/data")
	assert.Equal(t, 200, w.StatusCode)
}

// =============================================================================
// 3. Route Meta
// =============================================================================

func TestMeta_RouteMeta(t *testing.T) {
	type metaInfo struct {
		RequiresAuth bool
		RateLimit    int
	}

	r := gofi.NewRouter()
	r.Get("/public", gofi.RouteOptions{
		Meta: metaInfo{RequiresAuth: false, RateLimit: 100},
		Handler: func(c gofi.Context) error {
			meta, ok := c.Meta().This()
			assert.True(t, ok, "Expected meta to be available")

			info, ok := meta.(metaInfo)
			assert.True(t, ok, "Expected meta to be metaInfo type")
			assert.False(t, info.RequiresAuth)
			assert.Equal(t, 100, info.RateLimit)
			return c.SendString(200, "ok")
		},
	})

	mustTest(t, r, "GET", "/public")
}

func TestMeta_RouterMeta_All(t *testing.T) {
	r := gofi.NewRouter()
	r.Get("/a", gofi.RouteOptions{
		Meta:    "meta-a",
		Handler: func(c gofi.Context) error { return c.SendString(200, "a") },
	})
	r.Post("/b", gofi.RouteOptions{
		Meta:    "meta-b",
		Handler: func(c gofi.Context) error { return c.SendString(200, "b") },
	})

	all := r.Meta().All()
	assert.GreaterOrEqual(t, len(all), 2, "Expected at least 2 meta entries")
}

func TestMeta_RouterMeta_Route(t *testing.T) {
	r := gofi.NewRouter()
	r.Get("/users", gofi.RouteOptions{
		Meta:    "users-meta",
		Handler: func(c gofi.Context) error { return c.SendString(200, "users") },
	})

	val, ok := r.Meta().Route("/users", "GET")
	assert.True(t, ok, "Expected route meta to be found")
	assert.Equal(t, "users-meta", val)

	_, ok = r.Meta().Route("/nonexistent", "GET")
	assert.False(t, ok, "Expected route meta NOT to be found for /nonexistent")
}

func TestMeta_RouterMeta_Filter(t *testing.T) {
	r := gofi.NewRouter()
	r.Get("/api/users", gofi.RouteOptions{
		Meta:    "api-users",
		Handler: func(c gofi.Context) error { return c.SendString(200, "users") },
	})
	r.Get("/api/posts", gofi.RouteOptions{
		Meta:    "api-posts",
		Handler: func(c gofi.Context) error { return c.SendString(200, "posts") },
	})
	r.Get("/health", gofi.RouteOptions{
		Meta:    "health",
		Handler: func(c gofi.Context) error { return c.SendString(200, "healthy") },
	})

	filtered := r.Meta().Filter(func(path, method string) bool {
		return len(path) >= 4 && path[:4] == "/api"
	})
	assert.GreaterOrEqual(t, len(filtered), 2, "Expected at least 2 filtered entries")
}
