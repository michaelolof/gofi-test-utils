package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// =============================================================================
// Echo Helpers
// =============================================================================

func echoHandler(c echo.Context) error {
	return c.String(200, "OK")
}

// colonToColon keeps :param as-is for Echo (native syntax).
func loadEcho(routes []route) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	for _, r := range routes {
		switch r.method {
		case "GET":
			e.GET(r.path, echoHandler)
		case "POST":
			e.POST(r.path, echoHandler)
		case "PUT":
			e.PUT(r.path, echoHandler)
		case "PATCH":
			e.PATCH(r.path, echoHandler)
		case "DELETE":
			e.DELETE(r.path, echoHandler)
		}
	}
	return e
}

func loadEchoSingle(method, path string) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	switch method {
	case "GET":
		e.GET(path, echoHandler)
	case "POST":
		e.POST(path, echoHandler)
	case "PUT":
		e.PUT(path, echoHandler)
	case "PATCH":
		e.PATCH(path, echoHandler)
	case "DELETE":
		e.DELETE(path, echoHandler)
	}
	return e
}

// =============================================================================
// Echo + Schema: uses echo.Bind + go-playground/validator
// =============================================================================

var echoValidator = validator.New()

type echoSchemaValidator struct {
	v *validator.Validate
}

func (ev *echoSchemaValidator) Validate(i interface{}) error {
	return ev.v.Struct(i)
}

func loadEchoSchemaSingle(method, path string, handler echo.HandlerFunc) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Validator = &echoSchemaValidator{v: echoValidator}
	switch method {
	case "GET":
		e.GET(path, handler)
	case "POST":
		e.POST(path, handler)
	case "PUT":
		e.PUT(path, handler)
	case "PATCH":
		e.PATCH(path, handler)
	case "DELETE":
		e.DELETE(path, handler)
	}
	return e
}

// Echo schema structs for path param binding
type echoSingleParam struct {
	Name string `param:"name" validate:"required"`
}

type echoTwoParam struct {
	UserID string `param:"userID" validate:"required"`
	PostID string `param:"postID" validate:"required"`
}

type echoFiveParam struct {
	A string `param:"a" validate:"required"`
	B string `param:"b" validate:"required"`
	C string `param:"c" validate:"required"`
	D string `param:"d" validate:"required"`
	E string `param:"e" validate:"required"`
}

type echoJSONBind struct {
	ID   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

// =============================================================================
// 1. MICRO BENCHMARKS — Echo
// =============================================================================

func BenchmarkEcho_Static(b *testing.B) {
	e := loadEchoSingle("GET", "/")
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, e, req)
}

func BenchmarkEchoS_Static(b *testing.B) {
	e := loadEchoSchemaSingle("GET", "/", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, e, req)
}

func BenchmarkEcho_Param(b *testing.B) {
	e := loadEchoSingle("GET", "/user/:name")
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, e, req)
}

func BenchmarkEchoS_Param(b *testing.B) {
	e := loadEchoSchemaSingle("GET", "/user/:name", func(c echo.Context) error {
		var p echoSingleParam
		if err := c.Bind(&p); err != nil {
			return err
		}
		if err := c.Validate(&p); err != nil {
			return err
		}
		return c.String(200, "OK")
	})
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, e, req)
}

func BenchmarkEcho_Param5(b *testing.B) {
	e := loadEchoSingle("GET", "/:a/:b/:c/:d/:e")
	req, _ := http.NewRequest("GET", "/test/test/test/test/test", nil)
	benchRequest(b, e, req)
}

func BenchmarkEchoS_Param5(b *testing.B) {
	e := loadEchoSchemaSingle("GET", "/:a/:b/:c/:d/:e", func(c echo.Context) error {
		var p echoFiveParam
		if err := c.Bind(&p); err != nil {
			return err
		}
		if err := c.Validate(&p); err != nil {
			return err
		}
		return c.String(200, "OK")
	})
	req, _ := http.NewRequest("GET", "/test/test/test/test/test", nil)
	benchRequest(b, e, req)
}

func BenchmarkEcho_Param20(b *testing.B) {
	e := loadEchoSingle("GET", "/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/:k/:l/:m/:n/:o/:p/:q/:r/:s/:t")
	req, _ := http.NewRequest("GET", "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t", nil)
	benchRequest(b, e, req)
}

func BenchmarkEcho_ParamWrite(b *testing.B) {
	e := echo.New()
	e.HideBanner = true
	e.GET("/user/:name", func(c echo.Context) error {
		io.WriteString(c.Response(), c.Param("name"))
		return nil
	})
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, e, req)
}

func BenchmarkEchoS_ParamWrite(b *testing.B) {
	e := echo.New()
	e.HideBanner = true
	e.Validator = &echoSchemaValidator{v: echoValidator}
	e.GET("/user/:name", func(c echo.Context) error {
		var p echoSingleParam
		if err := c.Bind(&p); err != nil {
			return err
		}
		if err := c.Validate(&p); err != nil {
			return err
		}
		io.WriteString(c.Response(), p.Name)
		return nil
	})
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, e, req)
}

func BenchmarkEcho_MultiParam(b *testing.B) {
	e := loadEchoSingle("GET", "/users/:userID/posts/:postID")
	req, _ := http.NewRequest("GET", "/users/123/posts/456", nil)
	benchRequest(b, e, req)
}

func BenchmarkEchoS_MultiParam(b *testing.B) {
	e := loadEchoSchemaSingle("GET", "/users/:userID/posts/:postID", func(c echo.Context) error {
		var p echoTwoParam
		if err := c.Bind(&p); err != nil {
			return err
		}
		if err := c.Validate(&p); err != nil {
			return err
		}
		return c.String(200, "OK")
	})
	req, _ := http.NewRequest("GET", "/users/123/posts/456", nil)
	benchRequest(b, e, req)
}

func BenchmarkEcho_Wildcard(b *testing.B) {
	e := echo.New()
	e.HideBanner = true
	e.GET("/files/*", echoHandler)
	req, _ := http.NewRequest("GET", "/files/images/logo.png", nil)
	benchRequest(b, e, req)
}

func BenchmarkEcho_Deep(b *testing.B) {
	e := loadEchoSingle("GET", "/v1/api/deep/nested/resource/action")
	req, _ := http.NewRequest("GET", "/v1/api/deep/nested/resource/action", nil)
	benchRequest(b, e, req)
}

func BenchmarkEcho_404(b *testing.B) {
	e := loadEchoSingle("GET", "/")
	req, _ := http.NewRequest("GET", "/not-found", nil)
	benchRequest(b, e, req)
}

// =============================================================================
// 2. MIDDLEWARE — Echo
// =============================================================================

func echoNoopMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}

func BenchmarkEcho_Middleware5(b *testing.B) {
	e := echo.New()
	e.HideBanner = true
	for i := 0; i < 5; i++ {
		e.Use(echoNoopMiddleware)
	}
	e.GET("/", echoHandler)
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, e, req)
}

func BenchmarkEcho_Middleware10(b *testing.B) {
	e := echo.New()
	e.HideBanner = true
	for i := 0; i < 10; i++ {
		e.Use(echoNoopMiddleware)
	}
	e.GET("/", echoHandler)
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, e, req)
}

func BenchmarkEcho_Middleware20(b *testing.B) {
	e := echo.New()
	e.HideBanner = true
	for i := 0; i < 20; i++ {
		e.Use(echoNoopMiddleware)
	}
	e.GET("/", echoHandler)
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, e, req)
}

// =============================================================================
// 3. DATA HANDLING — Echo
// =============================================================================

func BenchmarkEcho_BindJSON_Small(b *testing.B) {
	b.StopTimer()
	e := echo.New()
	e.HideBanner = true
	e.POST("/", func(c echo.Context) error {
		var p SmallPayload
		if err := json.NewDecoder(c.Request().Body).Decode(&p); err != nil {
			return c.String(400, "Bad")
		}
		return c.String(200, "OK")
	})
	payload := `{"id": 1, "name": "test"}`
	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)
	}
}

func BenchmarkEchoS_BindJSON_Small(b *testing.B) {
	b.StopTimer()
	e := echo.New()
	e.HideBanner = true
	e.Validator = &echoSchemaValidator{v: echoValidator}
	e.POST("/", func(c echo.Context) error {
		var p echoJSONBind
		if err := c.Bind(&p); err != nil {
			return c.String(400, "Bad")
		}
		if err := c.Validate(&p); err != nil {
			return c.String(400, "Invalid")
		}
		return c.String(200, "OK")
	})
	payload := `{"id": 1, "name": "test"}`
	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)
	}
}

func BenchmarkEcho_JSONResponse_Large(b *testing.B) {
	b.StopTimer()
	data := make([]SmallPayload, 100)
	for i := 0; i < 100; i++ {
		data[i] = SmallPayload{ID: i, Name: fmt.Sprintf("Item %d", i)}
	}
	e := echo.New()
	e.HideBanner = true
	e.GET("/", func(c echo.Context) error {
		return c.JSON(200, data)
	})
	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)
	}
}

// =============================================================================
// 4. CONCURRENCY — Echo
// =============================================================================

func BenchmarkEcho_Parallel(b *testing.B) {
	e := echo.New()
	e.HideBanner = true
	e.GET("/", echoHandler)
	req := httptest.NewRequest("GET", "/", nil)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		w := httptest.NewRecorder()
		for pb.Next() {
			e.ServeHTTP(w, req)
		}
	})
}

// =============================================================================
// 5. ROUTE GROUPS — Echo
// =============================================================================

func BenchmarkEcho_RouteGroup(b *testing.B) {
	e := echo.New()
	e.HideBanner = true
	e.Use(echoNoopMiddleware)
	api := e.Group("/api")
	api.Use(echoNoopMiddleware)
	v1 := api.Group("/v1")
	v1.Use(echoNoopMiddleware)
	v1.GET("/users", echoHandler)
	v1.GET("/users/:id", echoHandler)
	v1.POST("/users", echoHandler)

	req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
	benchRequest(b, e, req)
}

// =============================================================================
// Echo + Schema — API route loading
// =============================================================================

func loadEchoSchema(routes []route) http.Handler {
	e := echo.New()
	e.HideBanner = true
	e.Validator = &echoSchemaValidator{v: echoValidator}
	for _, r := range routes {
		switch r.method {
		case "GET":
			e.GET(r.path, echoHandler)
		case "POST":
			e.POST(r.path, echoHandler)
		case "PUT":
			e.PUT(r.path, echoHandler)
		case "PATCH":
			e.PATCH(r.path, echoHandler)
		case "DELETE":
			e.DELETE(r.path, echoHandler)
		}
	}
	return e
}

// Variables staticEcho, staticEchoS, githubEcho, etc. declared in bench_test.go

// --- Echo API benchmarks ---
func BenchmarkEcho_StaticAll(b *testing.B) { benchRoutes(b, staticEcho, staticRoutes) }
func BenchmarkEcho_GithubAll(b *testing.B) { benchRoutes(b, githubEcho, githubAPI) }
func BenchmarkEcho_GPlusAll(b *testing.B)  { benchRoutes(b, gplusEcho, gplusAPI) }
func BenchmarkEcho_ParseAll(b *testing.B)  { benchRoutes(b, parseEcho, parseAPI) }

func BenchmarkEcho_GithubStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/user/repos", nil)
	benchRequest(b, githubEcho, req)
}
func BenchmarkEcho_GithubParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/repos/julienschmidt/httprouter/stargazers", nil)
	benchRequest(b, githubEcho, req)
}
func BenchmarkEcho_GPlusStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people", nil)
	benchRequest(b, gplusEcho, req)
}
func BenchmarkEcho_GPlusParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327", nil)
	benchRequest(b, gplusEcho, req)
}
func BenchmarkEcho_GPlus2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327/activities/123456789", nil)
	benchRequest(b, gplusEcho, req)
}
func BenchmarkEcho_ParseStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequest(b, parseEcho, req)
}
func BenchmarkEcho_ParseParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequest(b, parseEcho, req)
}
func BenchmarkEcho_Parse2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequest(b, parseEcho, req)
}

// --- EchoS API benchmarks ---

func BenchmarkEchoS_StaticAll(b *testing.B) { benchRoutes(b, staticEchoS, staticRoutes) }
func BenchmarkEchoS_GithubAll(b *testing.B) { benchRoutes(b, githubEchoS, githubAPI) }
func BenchmarkEchoS_GPlusAll(b *testing.B)  { benchRoutes(b, gplusEchoS, gplusAPI) }
func BenchmarkEchoS_ParseAll(b *testing.B)  { benchRoutes(b, parseEchoS, parseAPI) }

func BenchmarkEchoS_GithubStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/user/repos", nil)
	benchRequest(b, githubEchoS, req)
}
func BenchmarkEchoS_GithubParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/repos/julienschmidt/httprouter/stargazers", nil)
	benchRequest(b, githubEchoS, req)
}
func BenchmarkEchoS_GPlusStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people", nil)
	benchRequest(b, gplusEchoS, req)
}
func BenchmarkEchoS_GPlusParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327", nil)
	benchRequest(b, gplusEchoS, req)
}
func BenchmarkEchoS_GPlus2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327/activities/123456789", nil)
	benchRequest(b, gplusEchoS, req)
}
func BenchmarkEchoS_ParseStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequest(b, parseEchoS, req)
}
func BenchmarkEchoS_ParseParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequest(b, parseEchoS, req)
}
func BenchmarkEchoS_Parse2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequest(b, parseEchoS, req)
}
