package benchmarks

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi-test-utils/internal/utils"
	"github.com/valyala/fasthttp"
)

type requestLargeBindSchema struct {
	Request struct {
		Body []utils.LargePayload
	}
}

type requestLargeValidateSchema struct {
	Request struct {
		Body []utils.LargePayloadValidate `validate:"required,min=1,max=100"`
	}
}

type requestNestedLeaf struct {
	Value string `json:"value" validate:"required,min=2,max=30"`
}

type requestNestedGroup struct {
	Name  string              `json:"name" validate:"required"`
	Items []requestNestedLeaf `json:"items" validate:"required,min=1,max=10"`
}

type requestNestedSchema struct {
	Request struct {
		Body struct {
			Groups []requestNestedGroup `json:"groups" validate:"required,min=1,max=10"`
		} `validate:"required"`
	}
}

type requestPointerItem struct {
	Name *string `json:"name" validate:"required"`
}

type requestPointerSchema struct {
	Request struct {
		Body struct {
			Title *string              `json:"title" validate:"required"`
			Note  *string              `json:"note"`
			Items []requestPointerItem `json:"items" validate:"present,min=1,max=5"`
		} `validate:"required"`
	}
}

type requestSmallBindSchema struct {
	Request struct {
		Body utils.SmallPayload
	}
}

type requestSmallValidateSchema struct {
	Request struct {
		Body struct {
			ID   int    `json:"id" validate:"required"`
			Name string `json:"name" validate:"required"`
		} `validate:"required"`
	}
}

type requestRequiredBodySchema struct {
	Request struct {
		Body struct {
			Name string `json:"name" validate:"required"`
		} `validate:"required"`
	}
}

type requestCollectionMaxSchema struct {
	Request struct {
		Body struct {
			Items []requestNestedLeaf `json:"items" validate:"max=1"`
		} `validate:"required"`
	}
}

type requestTimeFailureSchema struct {
	Request struct {
		Body struct {
			When time.Time `json:"when" validate:"required"`
		} `validate:"required"`
	}
}

type requestCustomFailureSchema struct {
	Request struct {
		Body struct {
			Value string `json:"value" spec:"ws6bad" validate:"required"`
		} `validate:"required"`
	}
}

type requestMultiSourceFailureSchema struct {
	Request struct {
		Header struct {
			Attempts int `json:"X-Attempts" validate:"required,gte=1"`
		}
		Query struct {
			Limit int `json:"limit" validate:"required,gte=1"`
		}
		Path struct {
			ID int `json:"id" validate:"required,gte=1"`
		}
		Cookie struct {
			Session int `json:"session" validate:"required,gte=1"`
		}
	}
}

// requestBadStringSpec deliberately returns a value that cannot be assigned to
// its string destination. It keeps the custom-spec failure path reproducible
// without depending on application code.
type requestBadStringSpec struct{}

func (*requestBadStringSpec) SpecID() string { return "ws6bad" }
func (*requestBadStringSpec) Type() string   { return "string" }
func (*requestBadStringSpec) Format() string { return "string" }
func (*requestBadStringSpec) Decode(any) (any, error) {
	return struct{}{}, nil
}
func (*requestBadStringSpec) Encode(value any) (string, error) {
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("expected string, got %T", value)
	}
	return text, nil
}

var requestBenchmarkTimestamp = time.Date(2026, time.October, 3, 12, 0, 0, 123456789, time.UTC)

func requestLargeBindData() []utils.LargePayload {
	data := utils.GenerateLargeData()
	for i := range data {
		data[i].CreatedAt = requestBenchmarkTimestamp
		for event := range data[i].Events {
			data[i].Events[event].Timestamp = requestBenchmarkTimestamp
		}
	}
	return data
}

func requestLargeValidateData() []utils.LargePayloadValidate {
	data := utils.GenerateLargeDataValidate()
	for i := range data {
		data[i].ID = fmt.Sprintf("00000000-0000-4000-8000-%012x", i+1)
		data[i].CreatedAt = requestBenchmarkTimestamp
		for event := range data[i].Events {
			data[i].Events[event].Timestamp = requestBenchmarkTimestamp
		}
	}
	return data
}

func requestBenchmarkRouter(schema any, operation string, verifyBoundResult bool) gofi.Router {
	router := gofi.NewRouter()
	var handler gofi.HandlerFunc

	sendResult := func(c gofi.Context, err error) error {
		if err != nil {
			return c.SendString(400, "Bad")
		}
		return c.SendString(200, "OK")
	}

	switch operation {
	case "bind":
		if verifyBoundResult {
			handler = func(c gofi.Context) error {
				bound, err := gofi.ValidateAndBind[requestLargeBindSchema](c)
				if err == nil && (bound == nil || len(bound.Request.Body) != 50 || bound.Request.Body[0].ID != "uuid-0") {
					return c.SendString(500, "Invalid bound result")
				}
				return sendResult(c, err)
			}
		} else {
			handler = func(c gofi.Context) error {
				_, err := gofi.ValidateAndBind[requestLargeBindSchema](c)
				return sendResult(c, err)
			}
		}
	case "validate":
		handler = func(c gofi.Context) error {
			return sendResult(c, gofi.Validate(c))
		}
	case "bind-validate":
		if verifyBoundResult {
			handler = func(c gofi.Context) error {
				bound, err := gofi.ValidateAndBind[requestLargeValidateSchema](c)
				if err == nil && (bound == nil || len(bound.Request.Body) != 50 || bound.Request.Body[0].UserID != 1) {
					return c.SendString(500, "Invalid bound result")
				}
				return sendResult(c, err)
			}
		} else {
			handler = func(c gofi.Context) error {
				_, err := gofi.ValidateAndBind[requestLargeValidateSchema](c)
				return sendResult(c, err)
			}
		}
	case "small-bind-validate":
		handler = func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[requestSmallValidateSchema](c)
			return sendResult(c, err)
		}
	case "nested":
		if verifyBoundResult {
			handler = func(c gofi.Context) error {
				bound, err := gofi.ValidateAndBind[requestNestedSchema](c)
				if err == nil && (bound == nil || len(bound.Request.Body.Groups) != 2 || len(bound.Request.Body.Groups[0].Items) != 2) {
					return c.SendString(500, "Invalid bound result")
				}
				return sendResult(c, err)
			}
		} else {
			handler = func(c gofi.Context) error {
				_, err := gofi.ValidateAndBind[requestNestedSchema](c)
				return sendResult(c, err)
			}
		}
	case "pointers":
		if verifyBoundResult {
			handler = func(c gofi.Context) error {
				bound, err := gofi.ValidateAndBind[requestPointerSchema](c)
				if err == nil && (bound == nil || bound.Request.Body.Title == nil || *bound.Request.Body.Title != "benchmark" || bound.Request.Body.Note != nil || len(bound.Request.Body.Items) != 2) {
					return c.SendString(500, "Invalid bound result")
				}
				return sendResult(c, err)
			}
		} else {
			handler = func(c gofi.Context) error {
				_, err := gofi.ValidateAndBind[requestPointerSchema](c)
				return sendResult(c, err)
			}
		}
	default:
		panic("unknown request benchmark operation: " + operation)
	}

	router.Post("/", gofi.RouteOptions{
		Schema:  schema,
		Handler: handler,
	})
	return router
}

func benchmarkGofiJSONRequest(b *testing.B, router gofi.Router, payload []byte, expectedStatus int) {
	b.Helper()

	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.SetContentType("application/json")
	ctx.Request.SetBody(payload)
	handler := router.Handler()

	// A non-timed preflight verifies that the benchmark is measuring the
	// intended success or failure path before any result is recorded.
	handler(&ctx)
	if status := ctx.Response.StatusCode(); status != expectedStatus {
		b.Fatalf("preflight returned status %d, want %d; body=%q", status, expectedStatus, ctx.Response.Body())
	}
	if body := string(ctx.Response.Body()); body != map[bool]string{true: "OK", false: "Bad"}[expectedStatus == 200] {
		b.Fatalf("preflight returned body %q for status %d", body, expectedStatus)
	}

	b.ReportAllocs()
	b.SetBytes(int64(len(payload)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Response.Reset()
		handler(&ctx)
	}
	b.StopTimer()
	b.ReportMetric(float64(len(payload)), "payload-B")

	if status := ctx.Response.StatusCode(); status != expectedStatus {
		b.Fatalf("last request returned status %d, want %d", status, expectedStatus)
	}
}

func preflightGofiJSONRequest(b *testing.B, router gofi.Router, payload []byte, expectedStatus int) {
	b.Helper()

	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.SetContentType("application/json")
	ctx.Request.SetBody(payload)
	router.Handler()(&ctx)
	if status := ctx.Response.StatusCode(); status != expectedStatus {
		b.Fatalf("bound-result preflight returned status %d, want %d; body=%q", status, expectedStatus, ctx.Response.Body())
	}
}

type requestFailureCase struct {
	name    string
	router  gofi.Router
	method  string
	uri     string
	payload []byte
	prepare func(*fasthttp.RequestCtx)
}

func requestFailureRouter[T any](schema *T, path string, register func(gofi.Router)) gofi.Router {
	router := gofi.NewRouter()
	if register != nil {
		register(router)
	}
	router.Post(path, gofi.RouteOptions{
		Schema: schema,
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[T](c)
			if err != nil {
				return c.SendString(400, "Bad")
			}
			return c.SendString(200, "OK")
		},
	})
	return router
}

func requestFailureCases() []requestFailureCase {
	largeValidateData := requestLargeValidateData()
	earlyFailureData := requestLargeValidateData()
	earlyFailureData[0].ID = "invalid-uuid"
	lateFailureData := requestLargeValidateData()
	lastPayload := len(lateFailureData) - 1
	lastEvent := len(lateFailureData[lastPayload].Events) - 1
	lateFailureData[lastPayload].Events[lastEvent].Metrics.CPUUsage = 101

	mustJSON := func(value any) []byte {
		payload, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		return payload
	}

	return []requestFailureCase{
		{
			name:    "MalformedJSON",
			router:  requestFailureRouter(&requestLargeValidateSchema{}, "/", nil),
			method:  "POST",
			uri:     "/",
			payload: []byte(`{"id":`),
		},
		{
			name:    "MissingRequiredRootBody",
			router:  requestFailureRouter(&requestRequiredBodySchema{}, "/", nil),
			method:  "POST",
			uri:     "/",
			payload: nil,
		},
		{
			name:    "FirstFieldFailure",
			router:  requestFailureRouter(&requestLargeValidateSchema{}, "/", nil),
			method:  "POST",
			uri:     "/",
			payload: mustJSON(earlyFailureData),
		},
		{
			name:    "DeeplyNestedFailure",
			router:  requestFailureRouter(&requestLargeValidateSchema{}, "/", nil),
			method:  "POST",
			uri:     "/",
			payload: mustJSON(lateFailureData),
		},
		{
			name:    "CollectionMaximumFailure",
			router:  requestFailureRouter(&requestCollectionMaxSchema{}, "/", nil),
			method:  "POST",
			uri:     "/",
			payload: []byte(`{"items":[{"value":"one"},{"value":"two"}]}`),
		},
		{
			name:   "MultipleHeaderQueryPathCookieErrors",
			router: requestFailureRouter(&requestMultiSourceFailureSchema{}, "/:id", nil),
			method: "POST",
			uri:    "/bad?limit=bad",
			prepare: func(ctx *fasthttp.RequestCtx) {
				ctx.Request.Header.Set("X-Attempts", "bad")
				ctx.Request.Header.SetCookie("session", "bad")
			},
		},
		{
			name: "CustomSpecDecodeFailure",
			router: requestFailureRouter(&requestCustomFailureSchema{}, "/", func(router gofi.Router) {
				router.RegisterSpec(&requestBadStringSpec{})
			}),
			method:  "POST",
			uri:     "/",
			payload: []byte(`{"value":"anything"}`),
		},
		{
			name:    "InvalidTime",
			router:  requestFailureRouter(&requestTimeFailureSchema{}, "/", nil),
			method:  "POST",
			uri:     "/",
			payload: []byte(`{"when":"not-a-time"}`),
		},
		{
			name:    "ValidControl",
			router:  requestFailureRouter(&requestLargeValidateSchema{}, "/", nil),
			method:  "POST",
			uri:     "/",
			payload: mustJSON(largeValidateData),
		},
	}
}

func runFailureCase(tc requestFailureCase) (int, string) {
	var ctx fasthttp.RequestCtx
	ctx.Request.Header.SetMethod(tc.method)
	ctx.Request.SetRequestURI(tc.uri)
	if len(tc.payload) > 0 {
		ctx.Request.Header.SetContentType("application/json")
		ctx.Request.SetBody(tc.payload)
	}
	if tc.prepare != nil {
		tc.prepare(&ctx)
	}
	tc.router.Handler()(&ctx)
	return ctx.Response.StatusCode(), string(ctx.Response.Body())
}

func TestGofiRequestFailureMatrix(t *testing.T) {
	for _, tc := range requestFailureCases() {
		t.Run(tc.name, func(t *testing.T) {
			status, body := runFailureCase(tc)
			expectedStatus := 400
			expectedBody := "Bad"
			if tc.name == "ValidControl" {
				expectedStatus = 200
				expectedBody = "OK"
			}
			if status != expectedStatus || body != expectedBody {
				t.Fatalf("request returned status %d body %q, want %d %q", status, body, expectedStatus, expectedBody)
			}
		})
	}
}

func BenchmarkGofiS_RequestJSONFailures(b *testing.B) {
	for _, tc := range requestFailureCases() {
		if tc.name == "ValidControl" {
			continue
		}
		b.Run(tc.name, func(b *testing.B) {
			status, body := runFailureCase(tc)
			if status != 400 || body != "Bad" {
				b.Fatalf("preflight returned status %d body %q", status, body)
			}

			b.ReportAllocs()
			b.SetBytes(int64(len(tc.payload)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				status, _ = runFailureCase(tc)
			}
			b.StopTimer()
			if status != 400 {
				b.Fatalf("last request returned status %d, want 400", status)
			}
		})
	}
}

func BenchmarkGofiS_RequestJSON(b *testing.B) {
	largeBindPayload, err := json.Marshal(requestLargeBindData())
	if err != nil {
		b.Fatal(err)
	}
	largeValidateData := requestLargeValidateData()
	largeValidatePayload, err := json.Marshal(largeValidateData)
	if err != nil {
		b.Fatal(err)
	}

	earlyFailureData := requestLargeValidateData()
	earlyFailureData[0].ID = "invalid-uuid"
	earlyFailurePayload, err := json.Marshal(earlyFailureData)
	if err != nil {
		b.Fatal(err)
	}

	lateFailureData := requestLargeValidateData()
	lastPayload := len(lateFailureData) - 1
	lastEvent := len(lateFailureData[lastPayload].Events) - 1
	lateFailureData[lastPayload].Events[lastEvent].Metrics.CPUUsage = 101
	lateFailurePayload, err := json.Marshal(lateFailureData)
	if err != nil {
		b.Fatal(err)
	}

	nestedPayload := []byte(`{"groups":[{"name":"alpha","items":[{"value":"one"},{"value":"two"}]},{"name":"beta","items":[{"value":"three"}]}]}`)
	pointerPayload := []byte(`{"title":"benchmark","items":[{"name":"one"},{"name":"two"}]}`)

	cases := []struct {
		name           string
		schema         any
		operation      string
		payload        []byte
		expectedStatus int
	}{
		{name: "LargeBindOnly", schema: &requestLargeBindSchema{}, operation: "bind", payload: largeBindPayload, expectedStatus: 200},
		{name: "LargeValidateOnly", schema: &requestLargeValidateSchema{}, operation: "validate", payload: largeValidatePayload, expectedStatus: 200},
		{name: "LargeBindAndValidate", schema: &requestLargeValidateSchema{}, operation: "bind-validate", payload: largeValidatePayload, expectedStatus: 200},
		{name: "NestedArraysAndObjects", schema: &requestNestedSchema{}, operation: "nested", payload: nestedPayload, expectedStatus: 200},
		{name: "PointerAndOptionalFields", schema: &requestPointerSchema{}, operation: "pointers", payload: pointerPayload, expectedStatus: 200},
		{name: "EarlyValidationFailure", schema: &requestLargeValidateSchema{}, operation: "bind-validate", payload: earlyFailurePayload, expectedStatus: 400},
		{name: "LateDeepValidationFailure", schema: &requestLargeValidateSchema{}, operation: "bind-validate", payload: lateFailurePayload, expectedStatus: 400},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			if len(tc.payload) == 0 {
				b.Fatal(fmt.Errorf("empty payload for %s", tc.name))
			}
			if tc.expectedStatus == 200 && tc.operation != "validate" {
				verifyRouter := requestBenchmarkRouter(tc.schema, tc.operation, true)
				preflightGofiJSONRequest(b, verifyRouter, tc.payload, tc.expectedStatus)
			}
			benchmarkGofiJSONRequest(b, requestBenchmarkRouter(tc.schema, tc.operation, false), tc.payload, tc.expectedStatus)
		})
	}
}

type requestParallelWorkload struct {
	router  gofi.Router
	payload []byte
}

func benchmarkGofiJSONParallel(b *testing.B, workloads ...requestParallelWorkload) {
	b.Helper()
	if len(workloads) == 0 {
		b.Fatal("at least one workload is required")
	}

	for _, workload := range workloads {
		var ctx fasthttp.RequestCtx
		ctx.Request.Header.SetMethod("POST")
		ctx.Request.SetRequestURI("/")
		ctx.Request.Header.SetContentType("application/json")
		ctx.Request.SetBody(workload.payload)
		workload.router.Handler()(&ctx)
		if status := ctx.Response.StatusCode(); status != 200 || string(ctx.Response.Body()) != "OK" {
			b.Fatalf("preflight returned status %d body %q", status, ctx.Response.Body())
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		contexts := make([]fasthttp.RequestCtx, len(workloads))
		for index, workload := range workloads {
			contexts[index].Request.Header.SetMethod("POST")
			contexts[index].Request.SetRequestURI("/")
			contexts[index].Request.Header.SetContentType("application/json")
			contexts[index].Request.SetBody(workload.payload)
		}

		iteration := 0
		for pb.Next() {
			index := iteration % len(workloads)
			ctx := &contexts[index]
			ctx.Response.Reset()
			workloads[index].router.Handler()(ctx)
			iteration++
		}
	})
}

// Run with -cpu=1,8 to exercise low and high in-process concurrency. The HTTP
// WS-6 runner complements this with socket-level latency and runtime metrics.
func BenchmarkGofiS_RequestJSONConcurrent(b *testing.B) {
	largeBindPayload, err := json.Marshal(requestLargeBindData())
	if err != nil {
		b.Fatal(err)
	}
	largeValidatePayload, err := json.Marshal(requestLargeValidateData())
	if err != nil {
		b.Fatal(err)
	}
	smallPayload := []byte(`{"id":1,"name":"test"}`)

	small := requestParallelWorkload{
		router:  requestFailureRouter(&requestSmallBindSchema{}, "/", nil),
		payload: smallPayload,
	}
	smallBindValidate := requestParallelWorkload{
		router:  requestBenchmarkRouter(&requestSmallValidateSchema{}, "small-bind-validate", false),
		payload: smallPayload,
	}
	largeBind := requestParallelWorkload{
		router:  requestBenchmarkRouter(&requestLargeBindSchema{}, "bind", false),
		payload: largeBindPayload,
	}
	largeValidate := requestParallelWorkload{
		router:  requestBenchmarkRouter(&requestLargeValidateSchema{}, "validate", false),
		payload: largeValidatePayload,
	}
	largeBindValidate := requestParallelWorkload{
		router:  requestBenchmarkRouter(&requestLargeValidateSchema{}, "bind-validate", false),
		payload: largeValidatePayload,
	}

	cases := []struct {
		name      string
		workloads []requestParallelWorkload
	}{
		{name: "SmallBind", workloads: []requestParallelWorkload{small}},
		{name: "SmallBindAndValidate", workloads: []requestParallelWorkload{smallBindValidate}},
		{name: "LargeBindOnly", workloads: []requestParallelWorkload{largeBind}},
		{name: "LargeValidateOnly", workloads: []requestParallelWorkload{largeValidate}},
		{name: "LargeBindAndValidate", workloads: []requestParallelWorkload{largeBindValidate}},
		{name: "MixedSmallAndLargeBind", workloads: []requestParallelWorkload{small, largeBind}},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			benchmarkGofiJSONParallel(b, tc.workloads...)
		})
	}
}
