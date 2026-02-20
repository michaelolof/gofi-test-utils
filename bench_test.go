package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/michaelolof/gofi"
)

// =============================================================================
// Helper Types & Functions
// =============================================================================

type route struct {
	method string
	path   string
}

// mockResponseWriter is a minimal, zero-allocation ResponseWriter for hot-path benchmarks.
type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) { return http.Header{} }
func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}
func (m *mockResponseWriter) WriteHeader(int) {}

func mockHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func gofiHandler(c gofi.Context) error {
	return c.SendString(200, "OK")
}

// benchRequest measures routing performance for a single request.
func benchRequest(b *testing.B, router http.Handler, r *http.Request) {
	w := new(mockResponseWriter)
	u := r.URL
	rq := u.RawQuery
	r.RequestURI = u.RequestURI()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		u.RawQuery = rq
		router.ServeHTTP(w, r)
	}
}

// benchRoutes measures routing performance across all routes in a set.
func benchRoutes(b *testing.B, router http.Handler, routes []route) {
	w := new(mockResponseWriter)
	reqs := make([]*http.Request, len(routes))
	for i, r := range routes {
		req, _ := http.NewRequest(r.method, r.path, nil)
		req.RequestURI = r.path
		reqs[i] = req
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, req := range reqs {
			router.ServeHTTP(w, req)
		}
	}
}

// calcMem measures the heap memory required to load a routing structure.
func calcMem(name string, load func()) {
	m := new(runtime.MemStats)

	runtime.GC()
	runtime.ReadMemStats(m)
	before := m.HeapAlloc

	load()

	runtime.GC()
	runtime.ReadMemStats(m)
	after := m.HeapAlloc

	fmt.Fprintf(os.Stdout, "%s: %d Bytes\n", name, after-before)
}

// =============================================================================
// Schema Types for GofiS Benchmarks
// =============================================================================

// Schema for routes with a single path parameter
type singleParamSchema struct {
	Request struct {
		Path struct {
			Name string `json:"name" validate:"required"`
		}
	}
}

// Schema for routes with two path parameters
type twoParamSchema struct {
	Request struct {
		Path struct {
			UserID string `json:"userID" validate:"required"`
			PostID string `json:"postID" validate:"required"`
		}
	}
}

// Schema for routes with 5 path parameters
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

// Schema for JSON body binding
type jsonBindSchema struct {
	Request struct {
		Body struct {
			ID   int    `json:"id" validate:"required"`
			Name string `json:"name" validate:"required"`
		} `validate:"required"`
	}
}

// Schema for JSON response
type jsonResponseSchema struct {
	Ok struct {
		Body []SmallPayload
	}
}

// =============================================================================
// Loading Helpers — Gofi (no schema)
// =============================================================================

func loadGofi(routes []route) http.Handler {
	mux := gofi.NewServeMux()
	for _, r := range routes {
		path := colonToBrace(r.path)
		switch r.method {
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
			panic("Unknown HTTP method: " + r.method)
		}
	}
	return mux
}

func loadGofiSingle(method, path string) http.Handler {
	mux := gofi.NewServeMux()
	path = colonToBrace(path)
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

// Generic schema for routes — binds + validates, then returns OK.
func gofiSchemaHandler[T any](c gofi.Context) error {
	_, err := gofi.ValidateAndBind[T](c)
	if err != nil {
		return err
	}
	return c.SendString(200, "OK")
}

// Schema for a generic single-param route (uses "name" param)
func loadGofiSSingle(method, path string, schema any, handler func(gofi.Context) error) http.Handler {
	mux := gofi.NewServeMux()
	path = colonToBrace(path)
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

// loadGofiS loads all routes with a generic empty schema (schema compilation overhead).
func loadGofiS(routes []route) http.Handler {
	type emptySchema struct {
		Request struct{}
	}
	mux := gofi.NewServeMux()
	for _, r := range routes {
		path := colonToBrace(r.path)
		opts := gofi.RouteOptions{
			Schema:  &emptySchema{},
			Handler: gofiHandler,
		}
		switch r.method {
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
			panic("Unknown HTTP method: " + r.method)
		}
	}
	return mux
}

// =============================================================================
// Loading Helpers — Chi
// =============================================================================

func loadChi(routes []route) http.Handler {
	mux := chi.NewRouter()
	for _, r := range routes {
		path := colonToBrace(r.path)
		switch r.method {
		case "GET":
			mux.Get(path, mockHandler)
		case "POST":
			mux.Post(path, mockHandler)
		case "PUT":
			mux.Put(path, mockHandler)
		case "PATCH":
			mux.Patch(path, mockHandler)
		case "DELETE":
			mux.Delete(path, mockHandler)
		default:
			panic("Unknown HTTP method: " + r.method)
		}
	}
	return mux
}

func loadChiSingle(method, path string) http.Handler {
	mux := chi.NewRouter()
	path = colonToBrace(path)
	switch method {
	case "GET":
		mux.Get(path, mockHandler)
	case "POST":
		mux.Post(path, mockHandler)
	case "PUT":
		mux.Put(path, mockHandler)
	case "PATCH":
		mux.Patch(path, mockHandler)
	case "DELETE":
		mux.Delete(path, mockHandler)
	default:
		panic("Unknown HTTP method: " + method)
	}
	return mux
}

// colonToBrace converts ":param" style routes to "{param}" style.
func colonToBrace(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if len(p) > 0 && p[0] == ':' {
			parts[i] = "{" + p[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

// =============================================================================
// 1. MICRO BENCHMARKS — Basic Routing
// =============================================================================

// --- Static Route: GET / ---

func BenchmarkGofi_Static(b *testing.B) {
	router := loadGofiSingle("GET", "/")
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, router, req)
}

func BenchmarkGofiS_Static(b *testing.B) {
	type schema struct{ Request struct{} }
	router := loadGofiSSingle("GET", "/", &schema{}, gofiSchemaHandler[schema])
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, router, req)
}

func BenchmarkChi_Static(b *testing.B) {
	router := loadChiSingle("GET", "/")
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, router, req)
}

// --- Single Param: GET /user/:name ---

func BenchmarkGofi_Param(b *testing.B) {
	router := loadGofiSingle("GET", "/user/:name")
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, req)
}

func BenchmarkGofiS_Param(b *testing.B) {
	router := loadGofiSSingle("GET", "/user/:name", &singleParamSchema{}, gofiSchemaHandler[singleParamSchema])
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, req)
}

func BenchmarkChi_Param(b *testing.B) {
	router := loadChiSingle("GET", "/user/:name")
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, req)
}

// --- 5 Params: GET /:a/:b/:c/:d/:e ---

func BenchmarkGofi_Param5(b *testing.B) {
	router := loadGofiSingle("GET", "/:a/:b/:c/:d/:e")
	req, _ := http.NewRequest("GET", "/test/test/test/test/test", nil)
	benchRequest(b, router, req)
}

func BenchmarkGofiS_Param5(b *testing.B) {
	router := loadGofiSSingle("GET", "/:a/:b/:c/:d/:e", &fiveParamSchema{}, gofiSchemaHandler[fiveParamSchema])
	req, _ := http.NewRequest("GET", "/test/test/test/test/test", nil)
	benchRequest(b, router, req)
}

func BenchmarkChi_Param5(b *testing.B) {
	router := loadChiSingle("GET", "/:a/:b/:c/:d/:e")
	req, _ := http.NewRequest("GET", "/test/test/test/test/test", nil)
	benchRequest(b, router, req)
}

// --- 20 Params: GET /:a/:b/.../:t ---

func BenchmarkGofi_Param20(b *testing.B) {
	router := loadGofiSingle("GET", "/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/:k/:l/:m/:n/:o/:p/:q/:r/:s/:t")
	req, _ := http.NewRequest("GET", "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t", nil)
	benchRequest(b, router, req)
}

func BenchmarkChi_Param20(b *testing.B) {
	router := loadChiSingle("GET", "/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/:k/:l/:m/:n/:o/:p/:q/:r/:s/:t")
	req, _ := http.NewRequest("GET", "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t", nil)
	benchRequest(b, router, req)
}

// --- Param Write: GET /user/:name (writes param to response) ---

func BenchmarkGofi_ParamWrite(b *testing.B) {
	mux := gofi.NewServeMux()
	mux.Get("/user/{name}", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			io.WriteString(c.Writer(), c.Request().PathValue("name"))
			return nil
		},
	})
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, mux, req)
}

func BenchmarkGofiS_ParamWrite(b *testing.B) {
	mux := gofi.NewServeMux()
	mux.Get("/user/{name}", gofi.RouteOptions{
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
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, mux, req)
}

func BenchmarkChi_ParamWrite(b *testing.B) {
	mux := chi.NewRouter()
	mux.Get("/user/{name}", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, chi.URLParam(r, "name"))
	})
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, mux, req)
}

// --- Multi Param: GET /users/:userID/posts/:postID ---

func BenchmarkGofi_MultiParam(b *testing.B) {
	router := loadGofiSingle("GET", "/users/:userID/posts/:postID")
	req, _ := http.NewRequest("GET", "/users/123/posts/456", nil)
	benchRequest(b, router, req)
}

func BenchmarkGofiS_MultiParam(b *testing.B) {
	router := loadGofiSSingle("GET", "/users/:userID/posts/:postID", &twoParamSchema{}, gofiSchemaHandler[twoParamSchema])
	req, _ := http.NewRequest("GET", "/users/123/posts/456", nil)
	benchRequest(b, router, req)
}

func BenchmarkChi_MultiParam(b *testing.B) {
	router := loadChiSingle("GET", "/users/:userID/posts/:postID")
	req, _ := http.NewRequest("GET", "/users/123/posts/456", nil)
	benchRequest(b, router, req)
}

// --- Wildcard: GET /files/* ---

func BenchmarkGofi_Wildcard(b *testing.B) {
	mux := gofi.NewServeMux()
	mux.Get("/files/{path...}", gofi.RouteOptions{Handler: gofiHandler})
	req, _ := http.NewRequest("GET", "/files/images/logo.png", nil)
	benchRequest(b, mux, req)
}

func BenchmarkChi_Wildcard(b *testing.B) {
	mux := chi.NewRouter()
	mux.Get("/files/*", mockHandler)
	req, _ := http.NewRequest("GET", "/files/images/logo.png", nil)
	benchRequest(b, mux, req)
}

// --- Deep Nesting: GET /v1/api/deep/nested/resource/action ---

func BenchmarkGofi_Deep(b *testing.B) {
	router := loadGofiSingle("GET", "/v1/api/deep/nested/resource/action")
	req, _ := http.NewRequest("GET", "/v1/api/deep/nested/resource/action", nil)
	benchRequest(b, router, req)
}

func BenchmarkChi_Deep(b *testing.B) {
	router := loadChiSingle("GET", "/v1/api/deep/nested/resource/action")
	req, _ := http.NewRequest("GET", "/v1/api/deep/nested/resource/action", nil)
	benchRequest(b, router, req)
}

// --- 404 Handling ---

func BenchmarkGofi_404(b *testing.B) {
	router := loadGofiSingle("GET", "/")
	req, _ := http.NewRequest("GET", "/not-found", nil)
	benchRequest(b, router, req)
}

func BenchmarkChi_404(b *testing.B) {
	router := loadChiSingle("GET", "/")
	req, _ := http.NewRequest("GET", "/not-found", nil)
	benchRequest(b, router, req)
}

// =============================================================================
// 2. MIDDLEWARE SCALABILITY
// =============================================================================

func noopMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// --- 5 Middlewares ---

func BenchmarkGofi_Middleware5(b *testing.B) {
	r := gofi.NewServeMux()
	for i := 0; i < 5; i++ {
		r.Use(noopMiddleware)
	}
	r.Get("/", gofi.RouteOptions{Handler: gofiHandler})
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, r, req)
}

func BenchmarkChi_Middleware5(b *testing.B) {
	r := chi.NewRouter()
	for i := 0; i < 5; i++ {
		r.Use(noopMiddleware)
	}
	r.Get("/", mockHandler)
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, r, req)
}

// --- 10 Middlewares ---

func BenchmarkGofi_Middleware10(b *testing.B) {
	r := gofi.NewServeMux()
	for i := 0; i < 10; i++ {
		r.Use(noopMiddleware)
	}
	r.Get("/", gofi.RouteOptions{Handler: gofiHandler})
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, r, req)
}

func BenchmarkChi_Middleware10(b *testing.B) {
	r := chi.NewRouter()
	for i := 0; i < 10; i++ {
		r.Use(noopMiddleware)
	}
	r.Get("/", mockHandler)
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, r, req)
}

// --- 20 Middlewares ---

func BenchmarkGofi_Middleware20(b *testing.B) {
	r := gofi.NewServeMux()
	for i := 0; i < 20; i++ {
		r.Use(noopMiddleware)
	}
	r.Get("/", gofi.RouteOptions{Handler: gofiHandler})
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, r, req)
}

func BenchmarkChi_Middleware20(b *testing.B) {
	r := chi.NewRouter()
	for i := 0; i < 20; i++ {
		r.Use(noopMiddleware)
	}
	r.Get("/", mockHandler)
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, r, req)
}

// =============================================================================
// 3. DATA HANDLING & I/O
// =============================================================================

type SmallPayload struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// --- JSON Binding (Small Payload) ---

func BenchmarkGofi_BindJSON_Small(b *testing.B) {
	b.StopTimer()
	r := gofi.NewServeMux()
	r.Post("/", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			var p SmallPayload
			if err := json.NewDecoder(c.Request().Body).Decode(&p); err != nil {
				return err
			}
			return c.SendString(200, "OK")
		},
	})
	payload := `{"id": 1, "name": "test"}`

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkGofiS_BindJSON_Small(b *testing.B) {
	b.StopTimer()
	r := gofi.NewServeMux()
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
	payload := `{"id": 1, "name": "test"}`

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkChi_BindJSON_Small(b *testing.B) {
	b.StopTimer()
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var p SmallPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	payload := `{"id": 1, "name": "test"}`

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// --- JSON Response (Large Payload — 100 items) ---

func BenchmarkGofi_JSONResponse_Large(b *testing.B) {
	b.StopTimer()
	data := make([]SmallPayload, 100)
	for i := 0; i < 100; i++ {
		data[i] = SmallPayload{ID: i, Name: fmt.Sprintf("Item %d", i)}
	}

	r := gofi.NewServeMux()
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
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkChi_JSONResponse_Large(b *testing.B) {
	b.StopTimer()
	data := make([]SmallPayload, 100)
	for i := 0; i < 100; i++ {
		data[i] = SmallPayload{ID: i, Name: fmt.Sprintf("Item %d", i)}
	}

	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(data)
	})

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// =============================================================================
// 4. CONCURRENCY
// =============================================================================

func BenchmarkGofi_Parallel(b *testing.B) {
	r := gofi.NewServeMux()
	r.Get("/", gofi.RouteOptions{Handler: gofiHandler})
	req := httptest.NewRequest("GET", "/", nil)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		w := httptest.NewRecorder()
		for pb.Next() {
			r.ServeHTTP(w, req)
		}
	})
}

func BenchmarkChi_Parallel(b *testing.B) {
	r := chi.NewRouter()
	r.Get("/", mockHandler)
	req := httptest.NewRequest("GET", "/", nil)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		w := httptest.NewRecorder()
		for pb.Next() {
			r.ServeHTTP(w, req)
		}
	})
}

// =============================================================================
// 5. ROUTE GROUPS
// =============================================================================

func BenchmarkGofi_RouteGroup(b *testing.B) {
	r := gofi.NewServeMux()
	r.Use(noopMiddleware)
	r.Route("/api", func(api gofi.Router) {
		api.Use(noopMiddleware)
		api.Route("/v1", func(v1 gofi.Router) {
			v1.Use(noopMiddleware)
			v1.Get("/users", gofi.RouteOptions{Handler: gofiHandler})
			v1.Get("/users/{id}", gofi.RouteOptions{Handler: gofiHandler})
			v1.Post("/users", gofi.RouteOptions{Handler: gofiHandler})
		})
	})

	req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
	benchRequest(b, r, req)
}

func BenchmarkChi_RouteGroup(b *testing.B) {
	r := chi.NewRouter()
	r.Use(noopMiddleware)
	r.Route("/api", func(api chi.Router) {
		api.Use(noopMiddleware)
		api.Route("/v1", func(v1 chi.Router) {
			v1.Use(noopMiddleware)
			v1.Get("/users", mockHandler)
			v1.Get("/users/{id}", mockHandler)
			v1.Post("/users", mockHandler)
		})
	})

	req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
	benchRequest(b, r, req)
}

// =============================================================================
// 6. STATIC ROUTES (157 routes — Go docs directory inspired)
// =============================================================================

var staticRoutes = []route{
	{"GET", "/"},
	{"GET", "/cmd.html"},
	{"GET", "/code.html"},
	{"GET", "/contrib.html"},
	{"GET", "/contribute.html"},
	{"GET", "/debugging_with_gdb.html"},
	{"GET", "/docs.html"},
	{"GET", "/effective_go.html"},
	{"GET", "/files.log"},
	{"GET", "/gccgo_contribute.html"},
	{"GET", "/gccgo_install.html"},
	{"GET", "/go-logo-black.png"},
	{"GET", "/go-logo-blue.png"},
	{"GET", "/go-logo-white.png"},
	{"GET", "/go1.1.html"},
	{"GET", "/go1.2.html"},
	{"GET", "/go1.html"},
	{"GET", "/go1compat.html"},
	{"GET", "/go_faq.html"},
	{"GET", "/go_mem.html"},
	{"GET", "/go_spec.html"},
	{"GET", "/help.html"},
	{"GET", "/ie.css"},
	{"GET", "/install-source.html"},
	{"GET", "/install.html"},
	{"GET", "/logo-153x55.png"},
	{"GET", "/Makefile"},
	{"GET", "/root.html"},
	{"GET", "/share.png"},
	{"GET", "/sieve.gif"},
	{"GET", "/tos.html"},
	{"GET", "/articles/"},
	{"GET", "/articles/go_command.html"},
	{"GET", "/articles/index.html"},
	{"GET", "/articles/wiki/"},
	{"GET", "/articles/wiki/edit.html"},
	{"GET", "/articles/wiki/final-noclosure.go"},
	{"GET", "/articles/wiki/final-noerror.go"},
	{"GET", "/articles/wiki/final-parsetemplate.go"},
	{"GET", "/articles/wiki/final-template.go"},
	{"GET", "/articles/wiki/final.go"},
	{"GET", "/articles/wiki/get.go"},
	{"GET", "/articles/wiki/http-sample.go"},
	{"GET", "/articles/wiki/index.html"},
	{"GET", "/articles/wiki/Makefile"},
	{"GET", "/articles/wiki/notemplate.go"},
	{"GET", "/articles/wiki/part1-noerror.go"},
	{"GET", "/articles/wiki/part1.go"},
	{"GET", "/articles/wiki/part2.go"},
	{"GET", "/articles/wiki/part3-errorhandling.go"},
	{"GET", "/articles/wiki/part3.go"},
	{"GET", "/articles/wiki/test.bash"},
	{"GET", "/articles/wiki/test_edit.good"},
	{"GET", "/articles/wiki/test_Test.txt.good"},
	{"GET", "/articles/wiki/test_view.good"},
	{"GET", "/articles/wiki/view.html"},
	{"GET", "/codewalk/"},
	{"GET", "/codewalk/codewalk.css"},
	{"GET", "/codewalk/codewalk.js"},
	{"GET", "/codewalk/codewalk.xml"},
	{"GET", "/codewalk/functions.xml"},
	{"GET", "/codewalk/markov.go"},
	{"GET", "/codewalk/markov.xml"},
	{"GET", "/codewalk/pig.go"},
	{"GET", "/codewalk/popout.png"},
	{"GET", "/codewalk/run"},
	{"GET", "/codewalk/sharemem.xml"},
	{"GET", "/codewalk/urlpoll.go"},
	{"GET", "/devel/"},
	{"GET", "/devel/release.html"},
	{"GET", "/devel/weekly.html"},
	{"GET", "/gopher/"},
	{"GET", "/gopher/appenginegopher.jpg"},
	{"GET", "/gopher/appenginegophercolor.jpg"},
	{"GET", "/gopher/appenginelogo.gif"},
	{"GET", "/gopher/bumper.png"},
	{"GET", "/gopher/bumper192x108.png"},
	{"GET", "/gopher/bumper320x180.png"},
	{"GET", "/gopher/bumper480x270.png"},
	{"GET", "/gopher/bumper640x360.png"},
	{"GET", "/gopher/doc.png"},
	{"GET", "/gopher/frontpage.png"},
	{"GET", "/gopher/gopherbw.png"},
	{"GET", "/gopher/gophercolor.png"},
	{"GET", "/gopher/gophercolor16x16.png"},
	{"GET", "/gopher/help.png"},
	{"GET", "/gopher/pkg.png"},
	{"GET", "/gopher/project.png"},
	{"GET", "/gopher/ref.png"},
	{"GET", "/gopher/run.png"},
	{"GET", "/gopher/talks.png"},
	{"GET", "/gopher/pencil/"},
	{"GET", "/gopher/pencil/gopherhat.jpg"},
	{"GET", "/gopher/pencil/gopherhelmet.jpg"},
	{"GET", "/gopher/pencil/gophermega.jpg"},
	{"GET", "/gopher/pencil/gopherrunning.jpg"},
	{"GET", "/gopher/pencil/gopherswim.jpg"},
	{"GET", "/gopher/pencil/gopherswrench.jpg"},
	{"GET", "/play/"},
	{"GET", "/play/fib.go"},
	{"GET", "/play/hello.go"},
	{"GET", "/play/life.go"},
	{"GET", "/play/peano.go"},
	{"GET", "/play/pi.go"},
	{"GET", "/play/sieve.go"},
	{"GET", "/play/solitaire.go"},
	{"GET", "/play/tree.go"},
	{"GET", "/progs/"},
	{"GET", "/progs/cgo1.go"},
	{"GET", "/progs/cgo2.go"},
	{"GET", "/progs/cgo3.go"},
	{"GET", "/progs/cgo4.go"},
	{"GET", "/progs/defer.go"},
	{"GET", "/progs/defer.out"},
	{"GET", "/progs/defer2.go"},
	{"GET", "/progs/defer2.out"},
	{"GET", "/progs/eff_bytesize.go"},
	{"GET", "/progs/eff_bytesize.out"},
	{"GET", "/progs/eff_qr.go"},
	{"GET", "/progs/eff_sequence.go"},
	{"GET", "/progs/eff_sequence.out"},
	{"GET", "/progs/eff_unused1.go"},
	{"GET", "/progs/eff_unused2.go"},
	{"GET", "/progs/error.go"},
	{"GET", "/progs/error2.go"},
	{"GET", "/progs/error3.go"},
	{"GET", "/progs/error4.go"},
	{"GET", "/progs/go1.go"},
	{"GET", "/progs/gobs1.go"},
	{"GET", "/progs/gobs2.go"},
	{"GET", "/progs/image_draw.go"},
	{"GET", "/progs/image_package1.go"},
	{"GET", "/progs/image_package1.out"},
	{"GET", "/progs/image_package2.go"},
	{"GET", "/progs/image_package2.out"},
	{"GET", "/progs/image_package3.go"},
	{"GET", "/progs/image_package3.out"},
	{"GET", "/progs/image_package4.go"},
	{"GET", "/progs/image_package4.out"},
	{"GET", "/progs/image_package5.go"},
	{"GET", "/progs/image_package5.out"},
	{"GET", "/progs/image_package6.go"},
	{"GET", "/progs/image_package6.out"},
	{"GET", "/progs/interface.go"},
	{"GET", "/progs/interface2.go"},
	{"GET", "/progs/interface2.out"},
	{"GET", "/progs/json1.go"},
	{"GET", "/progs/json2.go"},
	{"GET", "/progs/json2.out"},
	{"GET", "/progs/json3.go"},
	{"GET", "/progs/json4.go"},
	{"GET", "/progs/json5.go"},
	{"GET", "/progs/run"},
	{"GET", "/progs/slices.go"},
	{"GET", "/progs/timeout1.go"},
	{"GET", "/progs/timeout2.go"},
	{"GET", "/progs/update.bash"},
}

var (
	staticGofi  http.Handler
	staticGofiS http.Handler
	staticChi   http.Handler
	staticEcho  http.Handler
	staticEchoS http.Handler
)

func init() {
	fmt.Println("#Static Routes:", len(staticRoutes))
	calcMem("Gofi", func() { staticGofi = loadGofi(staticRoutes) })
	calcMem("GofiS", func() { staticGofiS = loadGofiS(staticRoutes) })
	calcMem("Chi", func() { staticChi = loadChi(staticRoutes) })
	calcMem("Echo", func() { staticEcho = loadEcho(staticRoutes) })
	calcMem("EchoS", func() { staticEchoS = loadEchoSchema(staticRoutes) })
	fmt.Println()
}

func BenchmarkGofi_StaticAll(b *testing.B) {
	benchRoutes(b, staticGofi, staticRoutes)
}

func BenchmarkGofiS_StaticAll(b *testing.B) {
	benchRoutes(b, staticGofiS, staticRoutes)
}

func BenchmarkChi_StaticAll(b *testing.B) {
	benchRoutes(b, staticChi, staticRoutes)
}

// =============================================================================
// 7. REAL-WORLD API: GitHub API v3 (~160 routes)
// =============================================================================

var githubAPI = []route{
	// OAuth Authorizations
	{"GET", "/authorizations"},
	{"GET", "/authorizations/:id"},
	{"POST", "/authorizations"},
	{"DELETE", "/authorizations/:id"},
	{"GET", "/applications/:client_id/tokens/:access_token"},
	{"DELETE", "/applications/:client_id/tokens"},
	{"DELETE", "/applications/:client_id/tokens/:access_token"},

	// Activity
	{"GET", "/events"},
	{"GET", "/repos/:owner/:repo/events"},
	{"GET", "/networks/:owner/:repo/events"},
	{"GET", "/orgs/:org/events"},
	{"GET", "/users/:user/received_events"},
	{"GET", "/users/:user/received_events/public"},
	{"GET", "/users/:user/events"},
	{"GET", "/users/:user/events/public"},
	{"GET", "/users/:user/events/orgs/:org"},
	{"GET", "/feeds"},
	{"GET", "/notifications"},
	{"GET", "/repos/:owner/:repo/notifications"},
	{"PUT", "/notifications"},
	{"PUT", "/repos/:owner/:repo/notifications"},
	{"GET", "/notifications/threads/:id"},
	{"GET", "/notifications/threads/:id/subscription"},
	{"PUT", "/notifications/threads/:id/subscription"},
	{"DELETE", "/notifications/threads/:id/subscription"},
	{"GET", "/repos/:owner/:repo/stargazers"},
	{"GET", "/users/:user/starred"},
	{"GET", "/user/starred"},
	{"GET", "/user/starred/:owner/:repo"},
	{"PUT", "/user/starred/:owner/:repo"},
	{"DELETE", "/user/starred/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/subscribers"},
	{"GET", "/users/:user/subscriptions"},
	{"GET", "/user/subscriptions"},
	{"GET", "/repos/:owner/:repo/subscription"},
	{"PUT", "/repos/:owner/:repo/subscription"},
	{"DELETE", "/repos/:owner/:repo/subscription"},
	{"GET", "/user/subscriptions/:owner/:repo"},
	{"PUT", "/user/subscriptions/:owner/:repo"},
	{"DELETE", "/user/subscriptions/:owner/:repo"},

	// Gists
	{"GET", "/users/:user/gists"},
	{"GET", "/gists"},
	{"GET", "/gists/:id"},
	{"POST", "/gists"},
	{"PUT", "/gists/:id/star"},
	{"DELETE", "/gists/:id/star"},
	{"GET", "/gists/:id/star"},
	{"POST", "/gists/:id/forks"},
	{"DELETE", "/gists/:id"},

	// Git Data
	{"GET", "/repos/:owner/:repo/git/blobs/:sha"},
	{"POST", "/repos/:owner/:repo/git/blobs"},
	{"GET", "/repos/:owner/:repo/git/commits/:sha"},
	{"POST", "/repos/:owner/:repo/git/commits"},
	{"GET", "/repos/:owner/:repo/git/refs"},
	{"POST", "/repos/:owner/:repo/git/refs"},
	{"GET", "/repos/:owner/:repo/git/tags/:sha"},
	{"POST", "/repos/:owner/:repo/git/tags"},
	{"GET", "/repos/:owner/:repo/git/trees/:sha"},
	{"POST", "/repos/:owner/:repo/git/trees"},

	// Issues
	{"GET", "/issues"},
	{"GET", "/user/issues"},
	{"GET", "/orgs/:org/issues"},
	{"GET", "/repos/:owner/:repo/issues"},
	{"GET", "/repos/:owner/:repo/issues/:number"},
	{"POST", "/repos/:owner/:repo/issues"},
	{"GET", "/repos/:owner/:repo/assignees"},
	{"GET", "/repos/:owner/:repo/assignees/:assignee"},
	{"GET", "/repos/:owner/:repo/issues/:number/comments"},
	{"POST", "/repos/:owner/:repo/issues/:number/comments"},
	{"GET", "/repos/:owner/:repo/issues/:number/events"},
	{"GET", "/repos/:owner/:repo/labels"},
	{"GET", "/repos/:owner/:repo/labels/:name"},
	{"POST", "/repos/:owner/:repo/labels"},
	{"DELETE", "/repos/:owner/:repo/labels/:name"},
	{"GET", "/repos/:owner/:repo/issues/:number/labels"},
	{"POST", "/repos/:owner/:repo/issues/:number/labels"},
	{"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name"},
	{"PUT", "/repos/:owner/:repo/issues/:number/labels"},
	{"DELETE", "/repos/:owner/:repo/issues/:number/labels"},
	{"GET", "/repos/:owner/:repo/milestones/:number/labels"},
	{"GET", "/repos/:owner/:repo/milestones"},
	{"GET", "/repos/:owner/:repo/milestones/:number"},
	{"POST", "/repos/:owner/:repo/milestones"},
	{"DELETE", "/repos/:owner/:repo/milestones/:number"},

	// Miscellaneous
	{"GET", "/emojis"},
	{"GET", "/gitignore/templates"},
	{"GET", "/gitignore/templates/:name"},
	{"POST", "/markdown"},
	{"POST", "/markdown/raw"},
	{"GET", "/meta"},
	{"GET", "/rate_limit"},

	// Organizations
	{"GET", "/users/:user/orgs"},
	{"GET", "/user/orgs"},
	{"GET", "/orgs/:org"},
	{"GET", "/orgs/:org/members"},
	{"GET", "/orgs/:org/members/:user"},
	{"DELETE", "/orgs/:org/members/:user"},
	{"GET", "/orgs/:org/public_members"},
	{"GET", "/orgs/:org/public_members/:user"},
	{"PUT", "/orgs/:org/public_members/:user"},
	{"DELETE", "/orgs/:org/public_members/:user"},
	{"GET", "/orgs/:org/teams"},
	{"GET", "/teams/:id"},
	{"POST", "/orgs/:org/teams"},
	{"DELETE", "/teams/:id"},
	{"GET", "/teams/:id/members"},
	{"GET", "/teams/:id/members/:user"},
	{"PUT", "/teams/:id/members/:user"},
	{"DELETE", "/teams/:id/members/:user"},
	{"GET", "/teams/:id/repos"},
	{"GET", "/teams/:id/repos/:owner/:repo"},
	{"PUT", "/teams/:id/repos/:owner/:repo"},
	{"DELETE", "/teams/:id/repos/:owner/:repo"},
	{"GET", "/user/teams"},

	// Pull Requests
	{"GET", "/repos/:owner/:repo/pulls"},
	{"GET", "/repos/:owner/:repo/pulls/:number"},
	{"POST", "/repos/:owner/:repo/pulls"},
	{"GET", "/repos/:owner/:repo/pulls/:number/commits"},
	{"GET", "/repos/:owner/:repo/pulls/:number/files"},
	{"GET", "/repos/:owner/:repo/pulls/:number/merge"},
	{"PUT", "/repos/:owner/:repo/pulls/:number/merge"},
	{"GET", "/repos/:owner/:repo/pulls/:number/comments"},
	{"PUT", "/repos/:owner/:repo/pulls/:number/comments"},

	// Repositories
	{"GET", "/user/repos"},
	{"GET", "/users/:user/repos"},
	{"GET", "/orgs/:org/repos"},
	{"GET", "/repositories"},
	{"POST", "/user/repos"},
	{"POST", "/orgs/:org/repos"},
	{"GET", "/repos/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/contributors"},
	{"GET", "/repos/:owner/:repo/languages"},
	{"GET", "/repos/:owner/:repo/teams"},
	{"GET", "/repos/:owner/:repo/tags"},
	{"GET", "/repos/:owner/:repo/branches"},
	{"GET", "/repos/:owner/:repo/branches/:branch"},
	{"DELETE", "/repos/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/collaborators"},
	{"GET", "/repos/:owner/:repo/collaborators/:user"},
	{"PUT", "/repos/:owner/:repo/collaborators/:user"},
	{"DELETE", "/repos/:owner/:repo/collaborators/:user"},
	{"GET", "/repos/:owner/:repo/comments"},
	{"GET", "/repos/:owner/:repo/commits/:sha/comments"},
	{"POST", "/repos/:owner/:repo/commits/:sha/comments"},
	{"GET", "/repos/:owner/:repo/comments/:id"},
	{"DELETE", "/repos/:owner/:repo/comments/:id"},
	{"GET", "/repos/:owner/:repo/commits"},
	{"GET", "/repos/:owner/:repo/commits/:sha"},
	{"GET", "/repos/:owner/:repo/readme"},
	{"GET", "/repos/:owner/:repo/keys"},
	{"GET", "/repos/:owner/:repo/keys/:id"},
	{"POST", "/repos/:owner/:repo/keys"},
	{"DELETE", "/repos/:owner/:repo/keys/:id"},
	{"GET", "/repos/:owner/:repo/downloads"},
	{"GET", "/repos/:owner/:repo/downloads/:id"},
	{"DELETE", "/repos/:owner/:repo/downloads/:id"},
	{"GET", "/repos/:owner/:repo/forks"},
	{"POST", "/repos/:owner/:repo/forks"},
	{"GET", "/repos/:owner/:repo/hooks"},
	{"GET", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/hooks"},
	{"POST", "/repos/:owner/:repo/hooks/:id/tests"},
	{"DELETE", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/merges"},
	{"GET", "/repos/:owner/:repo/releases"},
	{"GET", "/repos/:owner/:repo/releases/:id"},
	{"POST", "/repos/:owner/:repo/releases"},
	{"DELETE", "/repos/:owner/:repo/releases/:id"},
	{"GET", "/repos/:owner/:repo/releases/:id/assets"},
	{"GET", "/repos/:owner/:repo/stats/contributors"},
	{"GET", "/repos/:owner/:repo/stats/commit_activity"},
	{"GET", "/repos/:owner/:repo/stats/code_frequency"},
	{"GET", "/repos/:owner/:repo/stats/participation"},
	{"GET", "/repos/:owner/:repo/stats/punch_card"},
	{"GET", "/repos/:owner/:repo/statuses/:ref"},
	{"POST", "/repos/:owner/:repo/statuses/:ref"},

	// Search
	{"GET", "/search/repositories"},
	{"GET", "/search/code"},
	{"GET", "/search/issues"},
	{"GET", "/search/users"},
	{"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword"},
	{"GET", "/legacy/repos/search/:keyword"},
	{"GET", "/legacy/user/search/:keyword"},
	{"GET", "/legacy/user/email/:email"},

	// Users
	{"GET", "/users/:user"},
	{"GET", "/user"},
	{"GET", "/users"},
	{"GET", "/user/emails"},
	{"POST", "/user/emails"},
	{"DELETE", "/user/emails"},
	{"GET", "/users/:user/followers"},
	{"GET", "/user/followers"},
	{"GET", "/users/:user/following"},
	{"GET", "/user/following"},
	{"GET", "/user/following/:user"},
	{"GET", "/users/:user/following/:target_user"},
	{"PUT", "/user/following/:user"},
	{"DELETE", "/user/following/:user"},
	{"GET", "/users/:user/keys"},
	{"GET", "/user/keys"},
	{"GET", "/user/keys/:id"},
	{"POST", "/user/keys"},
	{"DELETE", "/user/keys/:id"},
}

var (
	githubGofi  http.Handler
	githubGofiS http.Handler
	githubChi   http.Handler
	githubEcho  http.Handler
	githubEchoS http.Handler
)

func init() {
	fmt.Println("#GithubAPI Routes:", len(githubAPI))
	calcMem("Gofi", func() { githubGofi = loadGofi(githubAPI) })
	calcMem("GofiS", func() { githubGofiS = loadGofiS(githubAPI) })
	calcMem("Chi", func() { githubChi = loadChi(githubAPI) })
	calcMem("Echo", func() { githubEcho = loadEcho(githubAPI) })
	calcMem("EchoS", func() { githubEchoS = loadEchoSchema(githubAPI) })
	fmt.Println()
}

// Static
func BenchmarkGofi_GithubStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/user/repos", nil)
	benchRequest(b, githubGofi, req)
}

func BenchmarkGofiS_GithubStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/user/repos", nil)
	benchRequest(b, githubGofiS, req)
}

func BenchmarkChi_GithubStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/user/repos", nil)
	benchRequest(b, githubChi, req)
}

// Param
func BenchmarkGofi_GithubParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/repos/julienschmidt/httprouter/stargazers", nil)
	benchRequest(b, githubGofi, req)
}

func BenchmarkGofiS_GithubParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/repos/julienschmidt/httprouter/stargazers", nil)
	benchRequest(b, githubGofiS, req)
}

func BenchmarkChi_GithubParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/repos/julienschmidt/httprouter/stargazers", nil)
	benchRequest(b, githubChi, req)
}

// All Routes
func BenchmarkGofi_GithubAll(b *testing.B) {
	benchRoutes(b, githubGofi, githubAPI)
}

func BenchmarkGofiS_GithubAll(b *testing.B) {
	benchRoutes(b, githubGofiS, githubAPI)
}

func BenchmarkChi_GithubAll(b *testing.B) {
	benchRoutes(b, githubChi, githubAPI)
}

// =============================================================================
// 8. REAL-WORLD API: Google+ API (13 routes)
// =============================================================================

var gplusAPI = []route{
	// People
	{"GET", "/people/:userId"},
	{"GET", "/people"},
	{"GET", "/activities/:activityId/people/:collection"},
	{"GET", "/people/:userId/people/:collection"},
	{"GET", "/people/:userId/openIdConnect"},

	// Activities
	{"GET", "/people/:userId/activities/:collection"},
	{"GET", "/activities/:activityId"},
	{"GET", "/activities"},

	// Comments
	{"GET", "/activities/:activityId/comments"},
	{"GET", "/comments/:commentId"},

	// Moments
	{"POST", "/people/:userId/moments/:collection"},
	{"GET", "/people/:userId/moments/:collection"},
	{"DELETE", "/moments/:id"},
}

var (
	gplusGofi  http.Handler
	gplusGofiS http.Handler
	gplusChi   http.Handler
	gplusEcho  http.Handler
	gplusEchoS http.Handler
)

func init() {
	fmt.Println("#GPlusAPI Routes:", len(gplusAPI))
	calcMem("Gofi", func() { gplusGofi = loadGofi(gplusAPI) })
	calcMem("GofiS", func() { gplusGofiS = loadGofiS(gplusAPI) })
	calcMem("Chi", func() { gplusChi = loadChi(gplusAPI) })
	calcMem("Echo", func() { gplusEcho = loadEcho(gplusAPI) })
	calcMem("EchoS", func() { gplusEchoS = loadEchoSchema(gplusAPI) })
	fmt.Println()
}

// Static
func BenchmarkGofi_GPlusStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people", nil)
	benchRequest(b, gplusGofi, req)
}

func BenchmarkGofiS_GPlusStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people", nil)
	benchRequest(b, gplusGofiS, req)
}

func BenchmarkChi_GPlusStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people", nil)
	benchRequest(b, gplusChi, req)
}

// One Param
func BenchmarkGofi_GPlusParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327", nil)
	benchRequest(b, gplusGofi, req)
}

func BenchmarkGofiS_GPlusParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327", nil)
	benchRequest(b, gplusGofiS, req)
}

func BenchmarkChi_GPlusParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327", nil)
	benchRequest(b, gplusChi, req)
}

// Two Params
func BenchmarkGofi_GPlus2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327/activities/123456789", nil)
	benchRequest(b, gplusGofi, req)
}

func BenchmarkGofiS_GPlus2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327/activities/123456789", nil)
	benchRequest(b, gplusGofiS, req)
}

func BenchmarkChi_GPlus2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327/activities/123456789", nil)
	benchRequest(b, gplusChi, req)
}

// All Routes
func BenchmarkGofi_GPlusAll(b *testing.B) {
	benchRoutes(b, gplusGofi, gplusAPI)
}

func BenchmarkGofiS_GPlusAll(b *testing.B) {
	benchRoutes(b, gplusGofiS, gplusAPI)
}

func BenchmarkChi_GPlusAll(b *testing.B) {
	benchRoutes(b, gplusChi, gplusAPI)
}

// =============================================================================
// 9. REAL-WORLD API: Parse.com API (26 routes)
// =============================================================================

var parseAPI = []route{
	// Objects
	{"POST", "/1/classes/:className"},
	{"GET", "/1/classes/:className/:objectId"},
	{"PUT", "/1/classes/:className/:objectId"},
	{"GET", "/1/classes/:className"},
	{"DELETE", "/1/classes/:className/:objectId"},

	// Users
	{"POST", "/1/users"},
	{"GET", "/1/login"},
	{"GET", "/1/users/:objectId"},
	{"PUT", "/1/users/:objectId"},
	{"GET", "/1/users"},
	{"DELETE", "/1/users/:objectId"},
	{"POST", "/1/requestPasswordReset"},

	// Roles
	{"POST", "/1/roles"},
	{"GET", "/1/roles/:objectId"},
	{"PUT", "/1/roles/:objectId"},
	{"GET", "/1/roles"},
	{"DELETE", "/1/roles/:objectId"},

	// Files
	{"POST", "/1/files/:fileName"},

	// Analytics
	{"POST", "/1/events/:eventName"},

	// Push Notifications
	{"POST", "/1/push"},

	// Installations
	{"POST", "/1/installations"},
	{"GET", "/1/installations/:objectId"},
	{"PUT", "/1/installations/:objectId"},
	{"GET", "/1/installations"},
	{"DELETE", "/1/installations/:objectId"},

	// Cloud Functions
	{"POST", "/1/functions"},
}

var (
	parseGofi  http.Handler
	parseGofiS http.Handler
	parseChi   http.Handler
	parseEcho  http.Handler
	parseEchoS http.Handler
)

func init() {
	fmt.Println("#ParseAPI Routes:", len(parseAPI))
	calcMem("Gofi", func() { parseGofi = loadGofi(parseAPI) })
	calcMem("GofiS", func() { parseGofiS = loadGofiS(parseAPI) })
	calcMem("Chi", func() { parseChi = loadChi(parseAPI) })
	calcMem("Echo", func() { parseEcho = loadEcho(parseAPI) })
	calcMem("EchoS", func() { parseEchoS = loadEchoSchema(parseAPI) })
	fmt.Println()
}

// Static
func BenchmarkGofi_ParseStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequest(b, parseGofi, req)
}

func BenchmarkGofiS_ParseStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequest(b, parseGofiS, req)
}

func BenchmarkChi_ParseStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequest(b, parseChi, req)
}

// One Param
func BenchmarkGofi_ParseParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequest(b, parseGofi, req)
}

func BenchmarkGofiS_ParseParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequest(b, parseGofiS, req)
}

func BenchmarkChi_ParseParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequest(b, parseChi, req)
}

// Two Params
func BenchmarkGofi_Parse2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequest(b, parseGofi, req)
}

func BenchmarkGofiS_Parse2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequest(b, parseGofiS, req)
}

func BenchmarkChi_Parse2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequest(b, parseChi, req)
}

// All Routes
func BenchmarkGofi_ParseAll(b *testing.B) {
	benchRoutes(b, parseGofi, parseAPI)
}

func BenchmarkGofiS_ParseAll(b *testing.B) {
	benchRoutes(b, parseGofiS, parseAPI)
}

func BenchmarkChi_ParseAll(b *testing.B) {
	benchRoutes(b, parseChi, parseAPI)
}
