package benchmarks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// =============================================================================
// Chi Helpers
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

func noopMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// =============================================================================
// Chi + Schema helpers
// =============================================================================

var chiValidator = validator.New()

type chiSingleParam struct {
	Name string `validate:"required"`
}

type chiTwoParam struct {
	UserID string `validate:"required"`
	PostID string `validate:"required"`
}

type chiFiveParam struct {
	A string `validate:"required"`
	B string `validate:"required"`
	C string `validate:"required"`
	D string `validate:"required"`
	E string `validate:"required"`
}

type chiJSONBind struct {
	ID   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

// =============================================================================
// Global vars for loaded Chi routers
// =============================================================================

var (
	staticChi http.Handler
	githubChi http.Handler
	gplusChi  http.Handler
	parseChi  http.Handler
)

func init() {
	calcMem("Chi", func() { staticChi = loadChi(staticRoutes) })
}

func init() {
	calcMem("Chi", func() { githubChi = loadChi(githubAPI) })
}

func init() {
	calcMem("Chi", func() { gplusChi = loadChi(gplusAPI) })
}

func init() {
	calcMem("Chi", func() { parseChi = loadChi(parseAPI) })
}

// =============================================================================
// 1. MICRO BENCHMARKS — Chi
// =============================================================================

func BenchmarkChi_Static(b *testing.B) {
	router := loadChiSingle("GET", "/")
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, router, req)
}

func BenchmarkChi_Param(b *testing.B) {
	router := loadChiSingle("GET", "/user/:name")
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, req)
}

func BenchmarkChi_Param5(b *testing.B) {
	router := loadChiSingle("GET", "/:a/:b/:c/:d/:e")
	req, _ := http.NewRequest("GET", "/test/test/test/test/test", nil)
	benchRequest(b, router, req)
}

func BenchmarkChi_Param20(b *testing.B) {
	router := loadChiSingle("GET", "/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/:k/:l/:m/:n/:o/:p/:q/:r/:s/:t")
	req, _ := http.NewRequest("GET", "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t", nil)
	benchRequest(b, router, req)
}

func BenchmarkChi_ParamWrite(b *testing.B) {
	mux := chi.NewRouter()
	mux.Get("/user/{name}", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, chi.URLParam(r, "name"))
	})
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, mux, req)
}

func BenchmarkChi_MultiParam(b *testing.B) {
	router := loadChiSingle("GET", "/users/:userID/posts/:postID")
	req, _ := http.NewRequest("GET", "/users/123/posts/456", nil)
	benchRequest(b, router, req)
}

func BenchmarkChi_Wildcard(b *testing.B) {
	mux := chi.NewRouter()
	mux.Get("/files/*", mockHandler)
	req, _ := http.NewRequest("GET", "/files/images/logo.png", nil)
	benchRequest(b, mux, req)
}

func BenchmarkChi_Deep(b *testing.B) {
	router := loadChiSingle("GET", "/v1/api/deep/nested/resource/action")
	req, _ := http.NewRequest("GET", "/v1/api/deep/nested/resource/action", nil)
	benchRequest(b, router, req)
}

func BenchmarkChi_404(b *testing.B) {
	router := loadChiSingle("GET", "/")
	req, _ := http.NewRequest("GET", "/not-found", nil)
	benchRequest(b, router, req)
}

// =============================================================================
// Chi + Schema Micro Benchmarks
// =============================================================================

func BenchmarkChiS_Static(b *testing.B) {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, r, req)
}

func BenchmarkChiS_Param(b *testing.B) {
	r := chi.NewRouter()
	r.Get("/user/{name}", func(w http.ResponseWriter, r *http.Request) {
		p := chiSingleParam{Name: chi.URLParam(r, "name")}
		chiValidator.Struct(&p) //nolint:errcheck
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, r, req)
}

func BenchmarkChiS_Param5(b *testing.B) {
	r := chi.NewRouter()
	r.Get("/{a}/{b}/{c}/{d}/{e}", func(w http.ResponseWriter, r *http.Request) {
		p := chiFiveParam{
			A: chi.URLParam(r, "a"), B: chi.URLParam(r, "b"),
			C: chi.URLParam(r, "c"), D: chi.URLParam(r, "d"),
			E: chi.URLParam(r, "e"),
		}
		chiValidator.Struct(&p) //nolint:errcheck
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	req, _ := http.NewRequest("GET", "/test/test/test/test/test", nil)
	benchRequest(b, r, req)
}

func BenchmarkChiS_ParamWrite(b *testing.B) {
	r := chi.NewRouter()
	r.Get("/user/{name}", func(w http.ResponseWriter, r *http.Request) {
		p := chiSingleParam{Name: chi.URLParam(r, "name")}
		chiValidator.Struct(&p) //nolint:errcheck
		io.WriteString(w, p.Name)
	})
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, r, req)
}

func BenchmarkChiS_MultiParam(b *testing.B) {
	r := chi.NewRouter()
	r.Get("/users/{userID}/posts/{postID}", func(w http.ResponseWriter, r *http.Request) {
		p := chiTwoParam{
			UserID: chi.URLParam(r, "userID"),
			PostID: chi.URLParam(r, "postID"),
		}
		chiValidator.Struct(&p) //nolint:errcheck
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	req, _ := http.NewRequest("GET", "/users/123/posts/456", nil)
	benchRequest(b, r, req)
}

// =============================================================================
// 2. MIDDLEWARE SCALABILITY — Chi
// =============================================================================

func BenchmarkChi_Middleware5(b *testing.B) {
	r := chi.NewRouter()
	for i := 0; i < 5; i++ {
		r.Use(noopMiddleware)
	}
	r.Get("/", mockHandler)
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
// 3. DATA HANDLING — Chi
// =============================================================================

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

func BenchmarkChiS_BindJSON_Small(b *testing.B) {
	b.StopTimer()
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var p chiJSONBind
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			w.WriteHeader(400)
			return
		}
		if err := chiValidator.Struct(&p); err != nil {
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
// 4. CONCURRENCY — Chi
// =============================================================================

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
// 5. ROUTE GROUPS — Chi
// =============================================================================

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
// 6-9. API Benchmarks — Chi
// =============================================================================

func BenchmarkChi_StaticAll(b *testing.B) { benchRoutes(b, staticChi, staticRoutes) }

// GitHub API
func BenchmarkChi_GithubStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/user/repos", nil)
	benchRequest(b, githubChi, req)
}
func BenchmarkChi_GithubParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/repos/julienschmidt/httprouter/stargazers", nil)
	benchRequest(b, githubChi, req)
}
func BenchmarkChi_GithubAll(b *testing.B) { benchRoutes(b, githubChi, githubAPI) }

// Google+ API
func BenchmarkChi_GPlusStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people", nil)
	benchRequest(b, gplusChi, req)
}
func BenchmarkChi_GPlusParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327", nil)
	benchRequest(b, gplusChi, req)
}
func BenchmarkChi_GPlus2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327/activities/123456789", nil)
	benchRequest(b, gplusChi, req)
}
func BenchmarkChi_GPlusAll(b *testing.B) { benchRoutes(b, gplusChi, gplusAPI) }

// Parse.com API
func BenchmarkChi_ParseStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequest(b, parseChi, req)
}
func BenchmarkChi_ParseParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequest(b, parseChi, req)
}
func BenchmarkChi_Parse2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequest(b, parseChi, req)
}
func BenchmarkChi_ParseAll(b *testing.B) { benchRoutes(b, parseChi, parseAPI) }
