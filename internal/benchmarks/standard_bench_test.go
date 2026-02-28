package benchmarks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/michaelolof/gofi"
)

var stdBenchValidator = validator.New()

// =============================================================================
// Standard Types & Payloads
// =============================================================================

type StdRequest struct {
	ID    int    `json:"id" validate:"required,gte=1"`
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Email string `json:"email" validate:"required,min=5,max=50"`
}

type StdResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		ProcessedID int    `json:"processed_id"`
		EchoName    string `json:"echo_name"`
	} `json:"data"`
}

var stdValidPayload = `{"id": 42, "name": "benchmark user", "email": "test@example.com"}`

// =============================================================================
// Gofi Schema
// =============================================================================

type GofiStandardSchema struct {
	Request struct {
		Body struct {
			ID    int    `json:"id" validate:"required,gte=1"`
			Name  string `json:"name" validate:"required,min=3,max=50"`
			Email string `json:"email" validate:"required,min=5,max=50"`
		}
	}
	Ok struct {
		Body StdResponse
	}
}

// =============================================================================
// Echo Schema Validator Setup
// =============================================================================

type stdEchoValidator struct {
	v *validator.Validate
}

func (ev *stdEchoValidator) Validate(i interface{}) error {
	return ev.v.Struct(i)
}

// =============================================================================
// Benchmarks
// =============================================================================

// 1. Gofi (Reflection-less Manual Binding)
func BenchmarkStd_Gofi(b *testing.B) {
	b.StopTimer()
	r := gofi.NewServeMux()
	r.Post("/", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			var req StdRequest
			if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
				return c.SendString(400, "Bad Request")
			}

			resp := StdResponse{
				Success: true,
				Message: "OK",
			}
			resp.Data.ProcessedID = req.ID
			resp.Data.EchoName = req.Name

			// Gofi's c.Send uses JSON by default for structs without schema
			return c.Send(200, resp)
		},
	})

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", strings.NewReader(stdValidPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// 2. Gofi + Schema (Codegen if generated)
func BenchmarkStd_GofiSchema(b *testing.B) {
	b.StopTimer()
	r := gofi.NewServeMux()
	r.Post("/", gofi.RouteOptions{
		Schema: &GofiStandardSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[GofiStandardSchema](c)
			if err != nil {
				return err // mapped to 400 internally
			}

			s.Ok.Body.Success = true
			s.Ok.Body.Message = "OK"
			s.Ok.Body.Data.ProcessedID = s.Request.Body.ID
			s.Ok.Body.Data.EchoName = s.Request.Body.Name

			return c.Send(200, s.Ok)
		},
	})

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", strings.NewReader(stdValidPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// 3. Chi (Manual JSON Unmarshal + Marshal)
func BenchmarkStd_Chi(b *testing.B) {
	b.StopTimer()
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var req StdRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(400)
			return
		}

		resp := StdResponse{
			Success: true,
			Message: "OK",
		}
		resp.Data.ProcessedID = req.ID
		resp.Data.EchoName = req.Name

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(resp)
	})

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", strings.NewReader(stdValidPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// 4. Chi + Schema (go-playground/validator)
func BenchmarkStd_ChiSchema(b *testing.B) {
	b.StopTimer()
	r := chi.NewRouter()
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var req StdRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(400)
			return
		}

		if err := stdBenchValidator.Struct(&req); err != nil {
			w.WriteHeader(400)
			return
		}

		resp := StdResponse{
			Success: true,
			Message: "OK",
		}
		resp.Data.ProcessedID = req.ID
		resp.Data.EchoName = req.Name

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(resp)
	})

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", strings.NewReader(stdValidPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// 5. Echo (Manual JSON Unmarshal + JSON response)
func BenchmarkStd_Echo(b *testing.B) {
	b.StopTimer()
	e := echo.New()
	e.HideBanner = true
	e.POST("/", func(c echo.Context) error {
		var req StdRequest
		if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
			return c.String(400, "Bad Request")
		}

		resp := StdResponse{
			Success: true,
			Message: "OK",
		}
		resp.Data.ProcessedID = req.ID
		resp.Data.EchoName = req.Name

		return c.JSON(200, resp)
	})

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", strings.NewReader(stdValidPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)
	}
}

// 6. Echo + Schema (Echo Bind + go-playground/validator)
func BenchmarkStd_EchoSchema(b *testing.B) {
	b.StopTimer()
	e := echo.New()
	e.HideBanner = true
	e.Validator = &stdEchoValidator{v: stdBenchValidator}
	e.POST("/", func(c echo.Context) error {
		var req StdRequest
		if err := c.Bind(&req); err != nil {
			return c.String(400, "Bad Request")
		}

		if err := c.Validate(&req); err != nil {
			return c.String(400, "Bad Request")
		}

		resp := StdResponse{
			Success: true,
			Message: "OK",
		}
		resp.Data.ProcessedID = req.ID
		resp.Data.EchoName = req.Name

		return c.JSON(200, resp)
	})

	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", strings.NewReader(stdValidPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)
	}
}
