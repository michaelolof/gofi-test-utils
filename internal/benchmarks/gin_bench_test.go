package benchmarks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// =============================================================================
// Gin Helpers
// =============================================================================

func ginHandler(c *gin.Context) {
	c.String(200, "OK")
}

// colonToGin converts ":param" to ":param" (Gin uses colon syntax natively).
func loadGin(routes []route) *gin.Engine {
	r := gin.New()
	for _, rt := range routes {
		r.Handle(rt.method, rt.path, ginHandler)
	}
	return r
}

func loadGinSingle(method, path string) *gin.Engine {
	r := gin.New()
	r.Handle(method, path, ginHandler)
	return r
}

// =============================================================================
// Global vars for loaded Gin routers
// =============================================================================

var (
	staticGin http.Handler
	githubGin http.Handler
	gplusGin  http.Handler
	parseGin  http.Handler
)

func init() {
	calcMem("Gin", func() { staticGin = loadGin(staticRoutes) })
}

func init() {
	calcMem("Gin", func() { githubGin = loadGin(githubAPI) })
}

func init() {
	calcMem("Gin", func() { gplusGin = loadGin(gplusAPI) })
}

func init() {
	calcMem("Gin", func() { parseGin = loadGin(parseAPI) })
}

// =============================================================================
// Gin + Schema
// =============================================================================

type ginSingleParam struct {
	Name string `uri:"name" binding:"required"`
}

type ginTwoParam struct {
	UserID string `uri:"userID" binding:"required"`
	PostID string `uri:"postID" binding:"required"`
}

type ginFiveParam struct {
	A string `uri:"a" binding:"required"`
	B string `uri:"b" binding:"required"`
	C string `uri:"c" binding:"required"`
	D string `uri:"d" binding:"required"`
	E string `uri:"e" binding:"required"`
}

type ginJSONBind struct {
	ID   int    `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// =============================================================================
// 1. MICRO BENCHMARKS — Gin
// =============================================================================

func BenchmarkGin_Static(b *testing.B) {
	r := loadGinSingle("GET", "/")
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, r, req)
}

func BenchmarkGinS_Static(b *testing.B) {
	r := loadGinSingle("GET", "/")
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, r, req)
}

func BenchmarkGin_Param(b *testing.B) {
	r := loadGinSingle("GET", "/user/:name")
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, r, req)
}

func BenchmarkGinS_Param(b *testing.B) {
	r := gin.New()
	r.GET("/user/:name", func(c *gin.Context) {
		var p ginSingleParam
		if err := c.ShouldBindUri(&p); err != nil {
			c.String(400, "Bad")
			return
		}
		c.String(200, "OK")
	})
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, r, req)
}

func BenchmarkGin_Param5(b *testing.B) {
	r := loadGinSingle("GET", "/:a/:b/:c/:d/:e")
	req, _ := http.NewRequest("GET", "/test/test/test/test/test", nil)
	benchRequest(b, r, req)
}

func BenchmarkGinS_Param5(b *testing.B) {
	r := gin.New()
	r.GET("/:a/:b/:c/:d/:e", func(c *gin.Context) {
		var p ginFiveParam
		if err := c.ShouldBindUri(&p); err != nil {
			c.String(400, "Bad")
			return
		}
		c.String(200, "OK")
	})
	req, _ := http.NewRequest("GET", "/test/test/test/test/test", nil)
	benchRequest(b, r, req)
}

func BenchmarkGin_Param20(b *testing.B) {
	r := loadGinSingle("GET", "/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/:k/:l/:m/:n/:o/:p/:q/:r/:s/:t")
	req, _ := http.NewRequest("GET", "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t", nil)
	benchRequest(b, r, req)
}

func BenchmarkGin_ParamWrite(b *testing.B) {
	r := gin.New()
	r.GET("/user/:name", func(c *gin.Context) {
		io.WriteString(c.Writer, c.Param("name"))
	})
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, r, req)
}

func BenchmarkGinS_ParamWrite(b *testing.B) {
	r := gin.New()
	r.GET("/user/:name", func(c *gin.Context) {
		var p ginSingleParam
		if err := c.ShouldBindUri(&p); err != nil {
			c.String(400, "Bad")
			return
		}
		io.WriteString(c.Writer, p.Name)
	})
	req, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, r, req)
}

func BenchmarkGin_MultiParam(b *testing.B) {
	r := loadGinSingle("GET", "/users/:userID/posts/:postID")
	req, _ := http.NewRequest("GET", "/users/123/posts/456", nil)
	benchRequest(b, r, req)
}

func BenchmarkGinS_MultiParam(b *testing.B) {
	r := gin.New()
	r.GET("/users/:userID/posts/:postID", func(c *gin.Context) {
		var p ginTwoParam
		if err := c.ShouldBindUri(&p); err != nil {
			c.String(400, "Bad")
			return
		}
		c.String(200, "OK")
	})
	req, _ := http.NewRequest("GET", "/users/123/posts/456", nil)
	benchRequest(b, r, req)
}

func BenchmarkGin_Wildcard(b *testing.B) {
	r := gin.New()
	r.GET("/files/*path", ginHandler)
	req, _ := http.NewRequest("GET", "/files/images/logo.png", nil)
	benchRequest(b, r, req)
}

func BenchmarkGin_Deep(b *testing.B) {
	r := loadGinSingle("GET", "/v1/api/deep/nested/resource/action")
	req, _ := http.NewRequest("GET", "/v1/api/deep/nested/resource/action", nil)
	benchRequest(b, r, req)
}

func BenchmarkGin_404(b *testing.B) {
	r := loadGinSingle("GET", "/")
	req, _ := http.NewRequest("GET", "/not-found", nil)
	benchRequest(b, r, req)
}

// =============================================================================
// 2. MIDDLEWARE — Gin
// =============================================================================

func ginNoopMiddleware(c *gin.Context) {
	c.Next()
}

func BenchmarkGin_Middleware5(b *testing.B) {
	r := gin.New()
	for i := 0; i < 5; i++ {
		r.Use(ginNoopMiddleware)
	}
	r.GET("/", ginHandler)
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, r, req)
}

func BenchmarkGin_Middleware10(b *testing.B) {
	r := gin.New()
	for i := 0; i < 10; i++ {
		r.Use(ginNoopMiddleware)
	}
	r.GET("/", ginHandler)
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, r, req)
}

func BenchmarkGin_Middleware20(b *testing.B) {
	r := gin.New()
	for i := 0; i < 20; i++ {
		r.Use(ginNoopMiddleware)
	}
	r.GET("/", ginHandler)
	req, _ := http.NewRequest("GET", "/", nil)
	benchRequest(b, r, req)
}

// =============================================================================
// 3. DATA HANDLING — Gin
// =============================================================================

func BenchmarkGin_BindJSON_Small(b *testing.B) {
	b.StopTimer()
	r := gin.New()
	r.POST("/", func(c *gin.Context) {
		var p SmallPayload
		if err := json.NewDecoder(c.Request.Body).Decode(&p); err != nil {
			c.String(400, "Bad")
			return
		}
		c.String(200, "OK")
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

func BenchmarkGinS_BindJSON_Small(b *testing.B) {
	b.StopTimer()
	r := gin.New()
	r.POST("/", func(c *gin.Context) {
		var p ginJSONBind
		if err := c.ShouldBindJSON(&p); err != nil {
			c.String(400, "Bad")
			return
		}
		c.String(200, "OK")
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

func BenchmarkGin_JSONResponse_Large(b *testing.B) {
	b.StopTimer()
	data := make([]SmallPayload, 100)
	for i := 0; i < 100; i++ {
		data[i] = SmallPayload{ID: i, Name: fmt.Sprintf("Item %d", i)}
	}
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, data)
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
// 4. CONCURRENCY — Gin
// =============================================================================

func BenchmarkGin_Parallel(b *testing.B) {
	r := gin.New()
	r.GET("/", ginHandler)
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
// 5. ROUTE GROUPS — Gin
// =============================================================================

func BenchmarkGin_RouteGroup(b *testing.B) {
	r := gin.New()
	r.Use(ginNoopMiddleware)
	api := r.Group("/api")
	api.Use(ginNoopMiddleware)
	v1 := api.Group("/v1")
	v1.Use(ginNoopMiddleware)
	v1.GET("/users", ginHandler)
	v1.GET("/users/:id", ginHandler)
	v1.POST("/users", ginHandler)

	req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
	benchRequest(b, r, req)
}

// =============================================================================
// 6-9. API Benchmarks — Gin
// =============================================================================

func BenchmarkGin_StaticAll(b *testing.B) { benchRoutes(b, staticGin, staticRoutes) }
func BenchmarkGin_GithubAll(b *testing.B) { benchRoutes(b, githubGin, githubAPI) }
func BenchmarkGin_GPlusAll(b *testing.B)  { benchRoutes(b, gplusGin, gplusAPI) }
func BenchmarkGin_ParseAll(b *testing.B)  { benchRoutes(b, parseGin, parseAPI) }

func BenchmarkGin_GithubStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/user/repos", nil)
	benchRequest(b, githubGin, req)
}
func BenchmarkGin_GithubParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/repos/julienschmidt/httprouter/stargazers", nil)
	benchRequest(b, githubGin, req)
}
func BenchmarkGin_GPlusStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people", nil)
	benchRequest(b, gplusGin, req)
}
func BenchmarkGin_GPlusParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327", nil)
	benchRequest(b, gplusGin, req)
}
func BenchmarkGin_GPlus2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/people/118051310819094153327/activities/123456789", nil)
	benchRequest(b, gplusGin, req)
}
func BenchmarkGin_ParseStatic(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/users", nil)
	benchRequest(b, parseGin, req)
}
func BenchmarkGin_ParseParam(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go", nil)
	benchRequest(b, parseGin, req)
}
func BenchmarkGin_Parse2Params(b *testing.B) {
	req, _ := http.NewRequest("GET", "/1/classes/go/123456789", nil)
	benchRequest(b, parseGin, req)
}
