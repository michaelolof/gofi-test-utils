package benchmarks

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/michaelolof/gofi"
)

// =============================================================================
// Gofi Helpers
// =============================================================================

func gofiHandler(c gofi.Context) error {
	return c.SendString(200, "OK")
}

// gofiSchemaHandler is a generic schema handler that validates and binds.
func gofiSchemaHandler[T any](c gofi.Context) error {
	_, err := gofi.ValidateAndBind[T](c)
	if err != nil {
		return err
	}
	return c.SendString(200, "OK")
}

// gofiNoopMiddleware is a no-op middleware using Gofi's MiddlewareFunc signature.
var gofiNoopMiddleware gofi.MiddlewareFunc = func(c gofi.Context) error {
	return c.Next()
}

func benchDoGofiTest(router gofi.Router, method, path string) {
	_, _ = router.Test(gofi.TestOptions{Method: method, Path: path})
}

// benchGofiRequest measures routing performance for a single request (Gofi — fasthttp).
func benchGofiRequest(b *testing.B, router gofi.Router, method, path string) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchDoGofiTest(router, method, path)
	}
}

// benchGofiBenchTest measures pure routing performance without response collection overhead.
func benchGofiBenchTest(b *testing.B, router gofi.Router, method, path string) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchDoGofiTest(router, method, path)
	}
}

// benchGofiRoutes measures routing performance across all routes (Gofi — fasthttp).
func benchGofiRoutes(b *testing.B, router gofi.Router, routes []route) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, r := range routes {
			benchDoGofiTest(router, r.method, r.path)
		}
	}
}

// =============================================================================
// Loading Helpers — Gofi (no schema)
// =============================================================================

func loadGofi(routes []route) gofi.Router {
	mux := gofi.NewRouter()
	for _, r := range routes {
		switch r.method {
		case "GET":
			mux.Get(r.path, gofi.RouteOptions{Handler: gofiHandler})
		case "POST":
			mux.Post(r.path, gofi.RouteOptions{Handler: gofiHandler})
		case "PUT":
			mux.Put(r.path, gofi.RouteOptions{Handler: gofiHandler})
		case "PATCH":
			mux.Patch(r.path, gofi.RouteOptions{Handler: gofiHandler})
		case "DELETE":
			mux.Delete(r.path, gofi.RouteOptions{Handler: gofiHandler})
		default:
			panic("Unknown HTTP method: " + r.method)
		}
	}
	return mux
}

func loadGofiSingle(method, path string) gofi.Router {
	mux := gofi.NewRouter()
	switch method {
	case "GET":
		mux.Get(path, gofi.RouteOptions{Handler: gofiHandler})
	case "POST":
		mux.Post(path, gofi.RouteOptions{Handler: gofiHandler})
	case "PUT":
		mux.Put(path, gofi.RouteOptions{Handler: gofiHandler})
	case "PATCH":
		mux.Patch(path, gofi.RouteOptions{Handler: gofiHandler})
	case "DELETE":
		mux.Delete(path, gofi.RouteOptions{Handler: gofiHandler})
	default:
		panic("Unknown HTTP method: " + method)
	}
	return mux
}

// =============================================================================
// Loading Helpers — Gofi with Schema
// =============================================================================

func loadGofiSSingle(method, path string, schema any, handler func(gofi.Context) error) gofi.Router {
	mux := gofi.NewRouter()
	opts := gofi.RouteOptions{Schema: schema, Handler: handler}
	switch method {
	case "GET":
		mux.Get(path, opts)
	case "POST":
		mux.Post(path, opts)
	case "PUT":
		mux.Put(path, opts)
	case "PATCH":
		mux.Patch(path, opts)
	case "DELETE":
		mux.Delete(path, opts)
	default:
		panic("Unknown HTTP method: " + method)
	}
	return mux
}

func loadGofiS(routes []route) gofi.Router {
	type emptySchema struct {
		Request struct{}
	}
	mux := gofi.NewRouter()
	for _, r := range routes {
		opts := gofi.RouteOptions{
			Schema:  &emptySchema{},
			Handler: gofiHandler,
		}
		switch r.method {
		case "GET":
			mux.Get(r.path, opts)
		case "POST":
			mux.Post(r.path, opts)
		case "PUT":
			mux.Put(r.path, opts)
		case "PATCH":
			mux.Patch(r.path, opts)
		case "DELETE":
			mux.Delete(r.path, opts)
		default:
			panic("Unknown HTTP method: " + r.method)
		}
	}
	return mux
}

// =============================================================================
// Schema Types for GofiS Benchmarks
// =============================================================================

type singleParamSchema struct {
	Request struct {
		Path struct {
			Name string `json:"name" validate:"required"`
		}
	}
}

type twoParamSchema struct {
	Request struct {
		Path struct {
			UserID string `json:"userID" validate:"required"`
			PostID string `json:"postID" validate:"required"`
		}
	}
}

type fiveParamSchema struct {
	Request struct {
		Path struct {
			A string `json:"a" validate:"required"`
			B string `json:"b" validate:"required"`
			C string `json:"c" validate:"required"`
			D string `json:"d" validate:"required"`
			E string `json:"e" validate:"required"`
		}
	}
}

type jsonBindSchema struct {
	Request struct {
		Body struct {
			ID   int    `json:"id" validate:"required"`
			Name string `json:"name" validate:"required"`
		} `validate:"required"`
	}
}

// =============================================================================
// Global vars for loaded Gofi routers (used in init + API benchmarks)
// =============================================================================

var (
	staticGofi  gofi.Router
	staticGofiS gofi.Router
	githubGofi  gofi.Router
	githubGofiS gofi.Router
	gplusGofi   gofi.Router
	gplusGofiS  gofi.Router
	parseGofi   gofi.Router
	parseGofiS  gofi.Router
)

func init() {
	fmt.Println("#Static Routes:", len(staticRoutes))
	calcMem("Gofi", func() { staticGofi = loadGofi(staticRoutes) })
	calcMem("GofiS", func() { staticGofiS = loadGofiS(staticRoutes) })
}

func init() {
	fmt.Println("#GithubAPI Routes:", len(githubAPI))
	calcMem("Gofi", func() { githubGofi = loadGofi(githubAPI) })
	calcMem("GofiS", func() { githubGofiS = loadGofiS(githubAPI) })
}

func init() {
	fmt.Println("#GPlusAPI Routes:", len(gplusAPI))
	calcMem("Gofi", func() { gplusGofi = loadGofi(gplusAPI) })
	calcMem("GofiS", func() { gplusGofiS = loadGofiS(gplusAPI) })
}

func init() {
	fmt.Println("#ParseAPI Routes:", len(parseAPI))
	calcMem("Gofi", func() { parseGofi = loadGofi(parseAPI) })
	calcMem("GofiS", func() { parseGofiS = loadGofiS(parseAPI) })
}

// =============================================================================
// 1. MICRO BENCHMARKS — Basic Routing (Gofi)
// =============================================================================

// --- Static Route: GET / ---
func BenchmarkGofi_Static(b *testing.B) {
	router := loadGofiSingle("GET", "/")
	benchGofiRequest(b, router, "GET", "/")
}

func BenchmarkGofiS_Static(b *testing.B) {
	type schema struct{ Request struct{} }
	router := loadGofiSSingle("GET", "/", &schema{}, gofiSchemaHandler[schema])
	benchGofiRequest(b, router, "GET", "/")
}

// --- Single Param: GET /user/:name ---
func BenchmarkGofi_Param(b *testing.B) {
	router := loadGofiSingle("GET", "/user/:name")
	benchGofiRequest(b, router, "GET", "/user/gordon")
}

func BenchmarkGofiS_Param(b *testing.B) {
	router := loadGofiSSingle("GET", "/user/:name", &singleParamSchema{}, gofiSchemaHandler[singleParamSchema])
	benchGofiRequest(b, router, "GET", "/user/gordon")
}

// --- 5 Params: GET /:a/:b/:c/:d/:e ---
func BenchmarkGofi_Param5(b *testing.B) {
	router := loadGofiSingle("GET", "/:a/:b/:c/:d/:e")
	benchGofiRequest(b, router, "GET", "/test/test/test/test/test")
}

func BenchmarkGofiS_Param5(b *testing.B) {
	router := loadGofiSSingle("GET", "/:a/:b/:c/:d/:e", &fiveParamSchema{}, gofiSchemaHandler[fiveParamSchema])
	benchGofiRequest(b, router, "GET", "/test/test/test/test/test")
}

// --- 20 Params: GET /:a/:b/.../:t ---
func BenchmarkGofi_Param20(b *testing.B) {
	router := loadGofiSingle("GET", "/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/:k/:l/:m/:n/:o/:p/:q/:r/:s/:t")
	benchGofiRequest(b, router, "GET", "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t")
}

// --- Param Write: GET /user/:name (writes param to response) ---
func BenchmarkGofi_ParamWrite(b *testing.B) {
	mux := gofi.NewRouter()
	mux.Get("/user/:name", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			io.WriteString(c.Writer(), c.Param("name"))
			return nil
		},
	})
	benchGofiRequest(b, mux, "GET", "/user/gordon")
}

func BenchmarkGofiS_ParamWrite(b *testing.B) {
	mux := gofi.NewRouter()
	mux.Get("/user/:name", gofi.RouteOptions{
		Schema: &singleParamSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[singleParamSchema](c)
			if err != nil {
				return err
			}
			io.WriteString(c.Writer(), s.Request.Path.Name)
			return nil
		},
	})
	benchGofiRequest(b, mux, "GET", "/user/gordon")
}

// --- Multi Param: GET /users/:userID/posts/:postID ---
func BenchmarkGofi_MultiParam(b *testing.B) {
	router := loadGofiSingle("GET", "/users/:userID/posts/:postID")
	benchGofiRequest(b, router, "GET", "/users/123/posts/456")
}

func BenchmarkGofiS_MultiParam(b *testing.B) {
	router := loadGofiSSingle("GET", "/users/:userID/posts/:postID", &twoParamSchema{}, gofiSchemaHandler[twoParamSchema])
	benchGofiRequest(b, router, "GET", "/users/123/posts/456")
}

// --- Wildcard: GET /files/* ---
func BenchmarkGofi_Wildcard(b *testing.B) {
	mux := gofi.NewRouter()
	mux.Get("/files/*path", gofi.RouteOptions{Handler: gofiHandler})
	benchGofiRequest(b, mux, "GET", "/files/images/logo.png")
}

// --- Deep Nesting: GET /v1/api/deep/nested/resource/action ---
func BenchmarkGofi_Deep(b *testing.B) {
	router := loadGofiSingle("GET", "/v1/api/deep/nested/resource/action")
	benchGofiRequest(b, router, "GET", "/v1/api/deep/nested/resource/action")
}

// --- 404 Handling ---
func BenchmarkGofi_404(b *testing.B) {
	router := loadGofiSingle("GET", "/")
	benchGofiRequest(b, router, "GET", "/not-found")
}

// =============================================================================
// 2. MIDDLEWARE SCALABILITY (Gofi)
// =============================================================================

func BenchmarkGofi_Middleware5(b *testing.B) {
	r := gofi.NewRouter()
	for i := 0; i < 5; i++ {
		r.Use(gofiNoopMiddleware)
	}
	r.Get("/", gofi.RouteOptions{Handler: gofiHandler})
	benchGofiRequest(b, r, "GET", "/")
}

func BenchmarkGofi_Middleware10(b *testing.B) {
	r := gofi.NewRouter()
	for i := 0; i < 10; i++ {
		r.Use(gofiNoopMiddleware)
	}
	r.Get("/", gofi.RouteOptions{Handler: gofiHandler})
	benchGofiRequest(b, r, "GET", "/")
}

func BenchmarkGofi_Middleware20(b *testing.B) {
	r := gofi.NewRouter()
	for i := 0; i < 20; i++ {
		r.Use(gofiNoopMiddleware)
	}
	r.Get("/", gofi.RouteOptions{Handler: gofiHandler})
	benchGofiRequest(b, r, "GET", "/")
}

// =============================================================================
// 3. DATA HANDLING & I/O (Gofi)
// =============================================================================

func BenchmarkGofi_BindJSON_Small(b *testing.B) {
	b.StopTimer()
	r := gofi.NewRouter()
	r.Post("/", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			var p SmallPayload
			if err := json.NewDecoder(c.Request().Body).Decode(&p); err != nil {
				return err
			}
			return c.SendString(200, "OK")
		},
	})

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchDoGofiTest(r, "POST", "/")
	}
}

func BenchmarkGofiS_BindJSON_Small(b *testing.B) {
	b.StopTimer()
	r := gofi.NewRouter()
	r.Post("/", gofi.RouteOptions{
		Schema: &jsonBindSchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[jsonBindSchema](c)
			if err != nil {
				return err
			}
			return c.SendString(200, "OK")
		},
	})

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchDoGofiTest(r, "POST", "/")
	}
}

func BenchmarkGofi_JSONResponse_Large(b *testing.B) {
	b.StopTimer()
	data := make([]SmallPayload, 100)
	for i := 0; i < 100; i++ {
		data[i] = SmallPayload{ID: i, Name: fmt.Sprintf("Item %d", i)}
	}

	r := gofi.NewRouter()
	r.Get("/", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			c.Writer().Header().Set("Content-Type", "application/json")
			c.Writer().WriteHeader(200)
			return json.NewEncoder(c.Writer()).Encode(data)
		},
	})

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchDoGofiTest(r, "GET", "/")
	}
}

// =============================================================================
// 4. CONCURRENCY (Gofi)
// =============================================================================

func BenchmarkGofi_Parallel(b *testing.B) {
	r := gofi.NewRouter()
	r.Get("/", gofi.RouteOptions{Handler: gofiHandler})

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			benchDoGofiTest(r, "GET", "/")
		}
	})
}

// =============================================================================
// 5. ROUTE GROUPS (Gofi)
// =============================================================================

func BenchmarkGofi_RouteGroup(b *testing.B) {
	r := gofi.NewRouter()
	r.Use(gofiNoopMiddleware)
	r.Route("/api", func(api gofi.Router) {
		api.Use(gofiNoopMiddleware)
		api.Route("/v1", func(v1 gofi.Router) {
			v1.Use(gofiNoopMiddleware)
			v1.Get("/users", gofi.RouteOptions{Handler: gofiHandler})
			v1.Get("/users/:id", gofi.RouteOptions{Handler: gofiHandler})
			v1.Post("/users", gofi.RouteOptions{Handler: gofiHandler})
		})
	})

	benchGofiRequest(b, r, "GET", "/api/v1/users/123")
}

// =============================================================================
// 6-9. API Benchmarks (Gofi)
// =============================================================================

// Static API
func BenchmarkGofi_StaticAll(b *testing.B)  { benchGofiRoutes(b, staticGofi, staticRoutes) }
func BenchmarkGofiS_StaticAll(b *testing.B) { benchGofiRoutes(b, staticGofiS, staticRoutes) }

// GitHub API
func BenchmarkGofi_GithubStatic(b *testing.B) {
	benchGofiRequest(b, githubGofi, "GET", "/user/repos")
}
func BenchmarkGofiS_GithubStatic(b *testing.B) {
	benchGofiRequest(b, githubGofiS, "GET", "/user/repos")
}
func BenchmarkGofi_GithubParam(b *testing.B) {
	benchGofiRequest(b, githubGofi, "GET", "/repos/julienschmidt/httprouter/stargazers")
}
func BenchmarkGofiS_GithubParam(b *testing.B) {
	benchGofiRequest(b, githubGofiS, "GET", "/repos/julienschmidt/httprouter/stargazers")
}
func BenchmarkGofi_GithubAll(b *testing.B)  { benchGofiRoutes(b, githubGofi, githubAPI) }
func BenchmarkGofiS_GithubAll(b *testing.B) { benchGofiRoutes(b, githubGofiS, githubAPI) }

// Google+ API
func BenchmarkGofi_GPlusStatic(b *testing.B) {
	benchGofiRequest(b, gplusGofi, "GET", "/people")
}
func BenchmarkGofiS_GPlusStatic(b *testing.B) {
	benchGofiRequest(b, gplusGofiS, "GET", "/people")
}
func BenchmarkGofi_GPlusParam(b *testing.B) {
	benchGofiRequest(b, gplusGofi, "GET", "/people/118051310819094153327")
}
func BenchmarkGofiS_GPlusParam(b *testing.B) {
	benchGofiRequest(b, gplusGofiS, "GET", "/people/118051310819094153327")
}
func BenchmarkGofi_GPlus2Params(b *testing.B) {
	benchGofiRequest(b, gplusGofi, "GET", "/people/118051310819094153327/activities/123456789")
}
func BenchmarkGofiS_GPlus2Params(b *testing.B) {
	benchGofiRequest(b, gplusGofiS, "GET", "/people/118051310819094153327/activities/123456789")
}
func BenchmarkGofi_GPlusAll(b *testing.B)  { benchGofiRoutes(b, gplusGofi, gplusAPI) }
func BenchmarkGofiS_GPlusAll(b *testing.B) { benchGofiRoutes(b, gplusGofiS, gplusAPI) }

// Parse.com API
func BenchmarkGofi_ParseStatic(b *testing.B) {
	benchGofiRequest(b, parseGofi, "GET", "/1/users")
}
func BenchmarkGofiS_ParseStatic(b *testing.B) {
	benchGofiRequest(b, parseGofiS, "GET", "/1/users")
}
func BenchmarkGofi_ParseParam(b *testing.B) {
	benchGofiRequest(b, parseGofi, "GET", "/1/classes/go")
}
func BenchmarkGofiS_ParseParam(b *testing.B) {
	benchGofiRequest(b, parseGofiS, "GET", "/1/classes/go")
}
func BenchmarkGofi_Parse2Params(b *testing.B) {
	benchGofiRequest(b, parseGofi, "GET", "/1/classes/go/123456789")
}
func BenchmarkGofiS_Parse2Params(b *testing.B) {
	benchGofiRequest(b, parseGofiS, "GET", "/1/classes/go/123456789")
}
func BenchmarkGofi_ParseAll(b *testing.B)  { benchGofiRoutes(b, parseGofi, parseAPI) }
func BenchmarkGofiS_ParseAll(b *testing.B) { benchGofiRoutes(b, parseGofiS, parseAPI) }

// =============================================================================
// BenchTest variants — pure routing without response collection overhead
// =============================================================================

func BenchmarkGofiBT_Static(b *testing.B) {
	router := loadGofiSingle("GET", "/")
	benchGofiBenchTest(b, router, "GET", "/")
}

func BenchmarkGofiBT_Param(b *testing.B) {
	router := loadGofiSingle("GET", "/user/:name")
	benchGofiBenchTest(b, router, "GET", "/user/gordon")
}

func BenchmarkGofiBT_Param5(b *testing.B) {
	router := loadGofiSingle("GET", "/:a/:b/:c/:d/:e")
	benchGofiBenchTest(b, router, "GET", "/test/test/test/test/test")
}

func BenchmarkGofiBT_Param20(b *testing.B) {
	router := loadGofiSingle("GET", "/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/:k/:l/:m/:n/:o/:p/:q/:r/:s/:t")
	benchGofiBenchTest(b, router, "GET", "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t")
}

func BenchmarkGofiBT_Deep(b *testing.B) {
	router := loadGofiSingle("GET", "/v1/api/deep/nested/resource/action")
	benchGofiBenchTest(b, router, "GET", "/v1/api/deep/nested/resource/action")
}

func BenchmarkGofiBT_Parallel(b *testing.B) {
	r := gofi.NewRouter()
	r.Get("/", gofi.RouteOptions{Handler: gofiHandler})

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			benchDoGofiTest(r, "GET", "/")
		}
	})
}
