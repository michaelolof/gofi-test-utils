package benchmarks

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// =============================================================================
// Chi + Schema: manual struct binding + go-playground/validator
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
// Micro Benchmarks — Chi + Schema
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
// Data Handling — Chi + Schema
// =============================================================================

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
