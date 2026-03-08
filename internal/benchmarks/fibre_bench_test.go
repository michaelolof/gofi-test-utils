package benchmarks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// =============================================================================
// Fiber Helpers
// =============================================================================

func fiberHandler(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func loadFiber(routes []route) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	for _, rt := range routes {
		app.Add(rt.method, rt.path, fiberHandler)
	}
	return app
}

func loadFiberSingle(method, path string) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Add(method, path, fiberHandler)
	return app
}

// benchFiberRequest measures routing performance for a single request using Fiber's Test method.
func benchFiberRequest(b *testing.B, app *fiber.App, method, path string) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(method, path, nil)
		app.Test(req, -1) //nolint:errcheck
	}
}

// benchFiberRoutes measures routing performance across all routes.
func benchFiberRoutes(b *testing.B, app *fiber.App, routes []route) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, r := range routes {
			req, _ := http.NewRequest(r.method, r.path, nil)
			app.Test(req, -1) //nolint:errcheck
		}
	}
}

// =============================================================================
// Global vars for loaded Fiber apps
// =============================================================================

var (
	staticFiber *fiber.App
	githubFiber *fiber.App
	gplusFiber  *fiber.App
	parseFiber  *fiber.App
)

func init() {
	calcMem("Fiber", func() { staticFiber = loadFiber(staticRoutes) })
}

func init() {
	calcMem("Fiber", func() { githubFiber = loadFiber(githubAPI) })
}

func init() {
	calcMem("Fiber", func() { gplusFiber = loadFiber(gplusAPI) })
}

func init() {
	calcMem("Fiber", func() { parseFiber = loadFiber(parseAPI) })
}

// =============================================================================
// Fiber + Schema
// =============================================================================

var fiberValidator = validator.New()

type fiberSingleParam struct {
	Name string `params:"name" validate:"required"`
}

type fiberTwoParam struct {
	UserID string `params:"userID" validate:"required"`
	PostID string `params:"postID" validate:"required"`
}

type fiberFiveParam struct {
	A string `params:"a" validate:"required"`
	B string `params:"b" validate:"required"`
	C string `params:"c" validate:"required"`
	D string `params:"d" validate:"required"`
	E string `params:"e" validate:"required"`
}

type fiberJSONBind struct {
	ID   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

// =============================================================================
// 1. MICRO BENCHMARKS — Fiber
// =============================================================================

func BenchmarkFiber_Static(b *testing.B) {
	app := loadFiberSingle("GET", "/")
	benchFiberRequest(b, app, "GET", "/")
}

func BenchmarkFiberS_Static(b *testing.B) {
	app := loadFiberSingle("GET", "/")
	benchFiberRequest(b, app, "GET", "/")
}

func BenchmarkFiber_Param(b *testing.B) {
	app := loadFiberSingle("GET", "/user/:name")
	benchFiberRequest(b, app, "GET", "/user/gordon")
}

func BenchmarkFiberS_Param(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/user/:name", func(c *fiber.Ctx) error {
		var p fiberSingleParam
		if err := c.ParamsParser(&p); err != nil {
			return c.SendStatus(400)
		}
		if err := fiberValidator.Struct(&p); err != nil {
			return c.SendStatus(400)
		}
		return c.SendString("OK")
	})
	benchFiberRequest(b, app, "GET", "/user/gordon")
}

func BenchmarkFiber_Param5(b *testing.B) {
	app := loadFiberSingle("GET", "/:a/:b/:c/:d/:e")
	benchFiberRequest(b, app, "GET", "/test/test/test/test/test")
}

func BenchmarkFiberS_Param5(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/:a/:b/:c/:d/:e", func(c *fiber.Ctx) error {
		var p fiberFiveParam
		if err := c.ParamsParser(&p); err != nil {
			return c.SendStatus(400)
		}
		if err := fiberValidator.Struct(&p); err != nil {
			return c.SendStatus(400)
		}
		return c.SendString("OK")
	})
	benchFiberRequest(b, app, "GET", "/test/test/test/test/test")
}

func BenchmarkFiber_Param20(b *testing.B) {
	app := loadFiberSingle("GET", "/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/:k/:l/:m/:n/:o/:p/:q/:r/:s/:t")
	benchFiberRequest(b, app, "GET", "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t")
}

func BenchmarkFiber_ParamWrite(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/user/:name", func(c *fiber.Ctx) error {
		return c.SendString(c.Params("name"))
	})
	benchFiberRequest(b, app, "GET", "/user/gordon")
}

func BenchmarkFiberS_ParamWrite(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/user/:name", func(c *fiber.Ctx) error {
		var p fiberSingleParam
		if err := c.ParamsParser(&p); err != nil {
			return c.SendStatus(400)
		}
		if err := fiberValidator.Struct(&p); err != nil {
			return c.SendStatus(400)
		}
		return c.SendString(p.Name)
	})
	benchFiberRequest(b, app, "GET", "/user/gordon")
}

func BenchmarkFiber_MultiParam(b *testing.B) {
	app := loadFiberSingle("GET", "/users/:userID/posts/:postID")
	benchFiberRequest(b, app, "GET", "/users/123/posts/456")
}

func BenchmarkFiberS_MultiParam(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/users/:userID/posts/:postID", func(c *fiber.Ctx) error {
		var p fiberTwoParam
		if err := c.ParamsParser(&p); err != nil {
			return c.SendStatus(400)
		}
		if err := fiberValidator.Struct(&p); err != nil {
			return c.SendStatus(400)
		}
		return c.SendString("OK")
	})
	benchFiberRequest(b, app, "GET", "/users/123/posts/456")
}

func BenchmarkFiber_Wildcard(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/files/*", fiberHandler)
	benchFiberRequest(b, app, "GET", "/files/images/logo.png")
}

func BenchmarkFiber_Deep(b *testing.B) {
	app := loadFiberSingle("GET", "/v1/api/deep/nested/resource/action")
	benchFiberRequest(b, app, "GET", "/v1/api/deep/nested/resource/action")
}

func BenchmarkFiber_404(b *testing.B) {
	app := loadFiberSingle("GET", "/")
	benchFiberRequest(b, app, "GET", "/not-found")
}

// =============================================================================
// 2. MIDDLEWARE — Fiber
// =============================================================================

func fiberNoopMiddleware(c *fiber.Ctx) error {
	return c.Next()
}

func BenchmarkFiber_Middleware5(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	for i := 0; i < 5; i++ {
		app.Use(fiberNoopMiddleware)
	}
	app.Get("/", fiberHandler)
	benchFiberRequest(b, app, "GET", "/")
}

func BenchmarkFiber_Middleware10(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	for i := 0; i < 10; i++ {
		app.Use(fiberNoopMiddleware)
	}
	app.Get("/", fiberHandler)
	benchFiberRequest(b, app, "GET", "/")
}

func BenchmarkFiber_Middleware20(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	for i := 0; i < 20; i++ {
		app.Use(fiberNoopMiddleware)
	}
	app.Get("/", fiberHandler)
	benchFiberRequest(b, app, "GET", "/")
}

// =============================================================================
// 3. DATA HANDLING — Fiber
// =============================================================================

func BenchmarkFiber_BindJSON_Small(b *testing.B) {
	b.StopTimer()
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Post("/", func(c *fiber.Ctx) error {
		var p SmallPayload
		if err := json.Unmarshal(c.Body(), &p); err != nil {
			return c.SendStatus(400)
		}
		return c.SendString("OK")
	})
	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/", nil)
		req.Header.Set("Content-Type", "application/json")
		app.Test(req, -1) //nolint:errcheck
	}
}

func BenchmarkFiberS_BindJSON_Small(b *testing.B) {
	b.StopTimer()
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Post("/", func(c *fiber.Ctx) error {
		var p fiberJSONBind
		if err := c.BodyParser(&p); err != nil {
			return c.SendStatus(400)
		}
		if err := fiberValidator.Struct(&p); err != nil {
			return c.SendStatus(400)
		}
		return c.SendString("OK")
	})
	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/", nil)
		req.Header.Set("Content-Type", "application/json")
		app.Test(req, -1) //nolint:errcheck
	}
}

func BenchmarkFiber_JSONResponse_Large(b *testing.B) {
	b.StopTimer()
	data := make([]SmallPayload, 100)
	for i := 0; i < 100; i++ {
		data[i] = SmallPayload{ID: i, Name: fmt.Sprintf("Item %d", i)}
	}
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(data)
	})
	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/", nil)
		app.Test(req, -1) //nolint:errcheck
	}
}

// =============================================================================
// 4. CONCURRENCY — Fiber
// =============================================================================

func BenchmarkFiber_Parallel(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/", fiberHandler)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/", nil)
			app.Test(req, -1) //nolint:errcheck
		}
	})
}

// =============================================================================
// 5. ROUTE GROUPS — Fiber
// =============================================================================

func BenchmarkFiber_RouteGroup(b *testing.B) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(fiberNoopMiddleware)
	api := app.Group("/api")
	api.Use(fiberNoopMiddleware)
	v1 := api.Group("/v1")
	v1.Use(fiberNoopMiddleware)
	v1.Get("/users", fiberHandler)
	v1.Get("/users/:id", fiberHandler)
	v1.Post("/users", fiberHandler)

	benchFiberRequest(b, app, "GET", "/api/v1/users/123")
}

// =============================================================================
// 6-9. API Benchmarks — Fiber
// =============================================================================

func BenchmarkFiber_StaticAll(b *testing.B) { benchFiberRoutes(b, staticFiber, staticRoutes) }
func BenchmarkFiber_GithubAll(b *testing.B) { benchFiberRoutes(b, githubFiber, githubAPI) }
func BenchmarkFiber_GPlusAll(b *testing.B)  { benchFiberRoutes(b, gplusFiber, gplusAPI) }
func BenchmarkFiber_ParseAll(b *testing.B)  { benchFiberRoutes(b, parseFiber, parseAPI) }

func BenchmarkFiber_GithubStatic(b *testing.B) {
	benchFiberRequest(b, githubFiber, "GET", "/user/repos")
}
func BenchmarkFiber_GithubParam(b *testing.B) {
	benchFiberRequest(b, githubFiber, "GET", "/repos/julienschmidt/httprouter/stargazers")
}
func BenchmarkFiber_GPlusStatic(b *testing.B) {
	benchFiberRequest(b, gplusFiber, "GET", "/people")
}
func BenchmarkFiber_GPlusParam(b *testing.B) {
	benchFiberRequest(b, gplusFiber, "GET", "/people/118051310819094153327")
}
func BenchmarkFiber_GPlus2Params(b *testing.B) {
	benchFiberRequest(b, gplusFiber, "GET", "/people/118051310819094153327/activities/123456789")
}
func BenchmarkFiber_ParseStatic(b *testing.B) {
	benchFiberRequest(b, parseFiber, "GET", "/1/users")
}
func BenchmarkFiber_ParseParam(b *testing.B) {
	benchFiberRequest(b, parseFiber, "GET", "/1/classes/go")
}
func BenchmarkFiber_Parse2Params(b *testing.B) {
	benchFiberRequest(b, parseFiber, "GET", "/1/classes/go/123456789")
}
