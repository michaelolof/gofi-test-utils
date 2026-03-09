package httpbench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi-test-utils/internal/utils"
)

type SmallPayloadSchema struct {
	Request struct {
		Body utils.SmallPayload
	}
	Ok struct {
		Body string
	}
	Err struct {
		Body string
	}
}

type SmallValidateSchema struct {
	Request struct {
		Body utils.SmallPayloadValidate
	}
	Ok struct {
		Body string
	}
	Err struct {
		Body string
	}
}

type LargePayloadSchema struct {
	Request struct {
		Body []utils.LargePayload
	}
	Ok struct {
		Body string
	}
	Err struct {
		Body string
	}
}

type LargeValidateSchema struct {
	Request struct {
		Body []utils.LargePayloadValidate
	}
	Ok struct {
		Body string
	}
	Err struct {
		Body string
	}
}

type FormDataSchema struct {
	Request struct {
		Header struct {
			ContentType string `json:"content-type" validate:"required" default:"application/x-www-form-urlencoded"`
		}
		Body utils.SmallPayloadValidate
	}
	Ok struct {
		Body string
	}
	Err struct {
		Body string
	}
}

type MultipartSchema struct {
	Request struct {
		Header struct {
			ContentType string `json:"content-type" validate:"required" default:"multipart/form-data"`
		}
		Body utils.SmallPayloadValidate
	}
	Ok struct {
		Body string
	}
	Err struct {
		Body string
	}
}

type SmallResponseSchema struct {
	Ok struct {
		Body []utils.SmallPayload
	}
}

type LargeResponseSchema struct {
	Ok struct {
		Body []utils.LargePayload
	}
}

type SmallValidateResponseSchema struct {
	Ok struct {
		Body utils.SmallPayloadValidate
	}
}

type LargeValidateResponseSchema struct {
	Ok struct {
		Body []utils.LargePayloadValidate
	}
}

func setupBenchmarkRouter() gofi.Router {
	data := make([]utils.SmallPayload, 100)
	for i := 0; i < 100; i++ {
		data[i] = utils.SmallPayload{ID: i, Name: fmt.Sprintf("Item %d", i)}
	}

	largeData := utils.GenerateLargeData()
	largeDataValidate := utils.GenerateLargeDataValidate()

	mux := gofi.NewRouter()

	// Static route
	mux.Get("/", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "OK")
		},
	})

	// Single param
	mux.Get("/user/:name", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			io.WriteString(c.Writer(), c.Param("name"))
			return nil
		},
	})

	// Multi param
	mux.Get("/users/:userID/posts/:postID", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, c.Param("userID")+":"+c.Param("postID"))
		},
	})

	// Middleware Chain
	mw := func(c gofi.Context) error {
		return c.Next()
	}
	mux.With(mw, mw, mw, mw, mw).Get("/middlewares", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "OK")
		},
	})

	// Query Processing
	mux.Get("/query", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			q := c.Query("q")
			limit := c.Query("limit")
			return c.SendString(200, q+limit)
		},
	})

	// JSON bind
	mux.Post("/json", gofi.RouteOptions{
		Schema: &SmallPayloadSchema{},
		Handler: func(c gofi.Context) error {
			if _, err := gofi.ValidateAndBind[SmallPayloadSchema](c); err != nil {
				return c.SendString(400, "Bad")
			}
			return c.SendString(200, "OK")
		},
	})

	// JSON response
	mux.Get("/json-response", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			c.Writer().Header().Set("Content-Type", "application/json")
			b, _ := json.Marshal(data)
			return c.SendBytes(200, b)
		},
	})

	mux.Post("/json-large", gofi.RouteOptions{
		Schema: &LargePayloadSchema{},
		Handler: func(c gofi.Context) error {
			if _, err := gofi.ValidateAndBind[LargePayloadSchema](c); err != nil {
				return c.SendString(400, "Bad")
			}
			return c.SendString(200, "OK")
		},
	})

	mux.Get("/json-response-large", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			c.Writer().Header().Set("Content-Type", "application/json")
			b, _ := json.Marshal(largeData)
			return c.SendBytes(200, b)
		},
	})

	mux.Post("/json-validate-small", gofi.RouteOptions{
		Schema: &SmallValidateSchema{},
		Handler: func(c gofi.Context) error {
			if _, err := gofi.ValidateAndBind[SmallValidateSchema](c); err != nil {
				return c.SendString(400, "Bad")
			}
			return c.SendString(200, "OK")
		},
	})

	mux.Get("/json-response-validate-small", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			c.Writer().Header().Set("Content-Type", "application/json")
			b, _ := json.Marshal(utils.SmallPayloadValidate{ID: 1, Name: "test"})
			return c.SendBytes(200, b)
		},
	})

	mux.Post("/json-validate-large", gofi.RouteOptions{
		Schema: &LargeValidateSchema{},
		Handler: func(c gofi.Context) error {
			if _, err := gofi.ValidateAndBind[LargeValidateSchema](c); err != nil {
				return c.SendString(400, "Bad")
			}
			return c.SendString(200, "OK")
		},
	})

	mux.Get("/json-response-validate-large", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			c.Writer().Header().Set("Content-Type", "application/json")
			b, _ := json.Marshal(largeDataValidate)
			return c.SendBytes(200, b)
		},
	})

	mux.Post("/multipart", gofi.RouteOptions{
		Schema: &MultipartSchema{},
		Handler: func(c gofi.Context) error {
			if _, err := gofi.ValidateAndBind[MultipartSchema](c); err != nil {
				return c.SendString(400, "Bad")
			}
			return c.SendString(200, "OK")
		},
	})

	mux.Post("/formdata", gofi.RouteOptions{
		Schema: &FormDataSchema{},
		Handler: func(c gofi.Context) error {
			if _, err := gofi.ValidateAndBind[FormDataSchema](c); err != nil {
				return c.SendString(400, "Bad")
			}
			return c.SendString(200, "OK")
		},
	})

	return mux
}

func doRequest(mux gofi.Router, method, uri, ctype string, body []byte) *fasthttp.RequestCtx {
	ctx := &fasthttp.RequestCtx{}
	// Important to init for logging/internals avoiding panic sometimes
	ctx.Init(&fasthttp.Request{}, nil, nil)
	ctx.Request.Header.SetMethod(method)
	ctx.Request.SetRequestURI(uri)
	if ctype != "" {
		ctx.Request.Header.SetContentType(ctype)
	}
	if body != nil {
		ctx.Request.SetBody(body)
	}
	mux.Handler()(ctx)
	return ctx
}

func TestBenchmarkEndpoints_Routing(t *testing.T) {
	mux := setupBenchmarkRouter()

	t.Run("Static Route", func(t *testing.T) {
		ctx := doRequest(mux, "GET", "/", "", nil)
		assert.Equal(t, 200, ctx.Response.StatusCode())
		assert.Equal(t, "OK", string(ctx.Response.Body()))
	})

	t.Run("Single Param", func(t *testing.T) {
		ctx := doRequest(mux, "GET", "/user/gordon", "", nil)
		assert.Equal(t, 200, ctx.Response.StatusCode())
		assert.Equal(t, "gordon", string(ctx.Response.Body()))
	})

	t.Run("Multi Param", func(t *testing.T) {
		ctx := doRequest(mux, "GET", "/users/123/posts/456", "", nil)
		assert.Equal(t, 200, ctx.Response.StatusCode())
		assert.Equal(t, "123:456", string(ctx.Response.Body()))
	})

	t.Run("Middleware Chain", func(t *testing.T) {
		ctx := doRequest(mux, "GET", "/middlewares", "", nil)
		assert.Equal(t, 200, ctx.Response.StatusCode())
		assert.Equal(t, "OK", string(ctx.Response.Body()))
	})

	t.Run("Query Processing", func(t *testing.T) {
		ctx := doRequest(mux, "GET", "/query?q=searchterm&limit=10", "", nil)
		assert.Equal(t, 200, ctx.Response.StatusCode())
		assert.Equal(t, "searchterm10", string(ctx.Response.Body()))
	})
}

func TestBenchmarkEndpoints_JSON(t *testing.T) {
	mux := setupBenchmarkRouter()

	t.Run("JSON Bind Small (Valid)", func(t *testing.T) {
		payload := utils.SmallPayload{ID: 1, Name: "test"}
		b, _ := json.Marshal(payload)

		ctx := doRequest(mux, "POST", "/json", "application/json", b)
		assert.Equal(t, 200, ctx.Response.StatusCode())
		assert.Equal(t, "OK", string(ctx.Response.Body()))
	})

	t.Run("JSON Bind Small (Invalid)", func(t *testing.T) {
		ctx := doRequest(mux, "POST", "/json", "application/json", []byte(`{malformed`))
		assert.Equal(t, 400, ctx.Response.StatusCode())
		assert.Equal(t, "Bad", string(ctx.Response.Body()))
	})

	t.Run("JSON Bind Large (Valid)", func(t *testing.T) {
		largeData := utils.GenerateLargeData()
		b, _ := json.Marshal(largeData)

		ctx := doRequest(mux, "POST", "/json-large", "application/json", b)
		assert.Equal(t, 200, ctx.Response.StatusCode())
		assert.Equal(t, "OK", string(ctx.Response.Body()))
	})

	t.Run("JSON Response Small", func(t *testing.T) {
		ctx := doRequest(mux, "GET", "/json-response", "", nil)
		assert.Equal(t, 200, ctx.Response.StatusCode())

		var decoded []utils.SmallPayload
		err := json.Unmarshal(ctx.Response.Body(), &decoded)
		assert.NoError(t, err)
		assert.Len(t, decoded, 100)
		assert.Equal(t, 0, decoded[0].ID)
		assert.Equal(t, "Item 0", decoded[0].Name)
	})

	t.Run("JSON Response Large", func(t *testing.T) {
		ctx := doRequest(mux, "GET", "/json-response-large", "", nil)
		assert.Equal(t, 200, ctx.Response.StatusCode())

		var decoded []utils.LargePayload
		err := json.Unmarshal(ctx.Response.Body(), &decoded)
		assert.NoError(t, err)
		assert.Len(t, decoded, 50)
		assert.Equal(t, "uuid-0", decoded[0].ID)
	})
}

func TestBenchmarkEndpoints_JSONValidate(t *testing.T) {
	mux := setupBenchmarkRouter()

	t.Run("JSON Validate Small (Valid)", func(t *testing.T) {
		payload := utils.SmallPayloadValidate{ID: 1, Name: "test"}
		b, _ := json.Marshal(payload)

		ctx := doRequest(mux, "POST", "/json-validate-small", "application/json", b)
		assert.Equal(t, 200, ctx.Response.StatusCode())
		assert.Equal(t, "OK", string(ctx.Response.Body()))
	})

	t.Run("JSON Validate Small (Validation Failure)", func(t *testing.T) {
		payload := map[string]interface{}{"id": 1}
		b, _ := json.Marshal(payload)

		ctx := doRequest(mux, "POST", "/json-validate-small", "application/json", b)
		assert.Equal(t, 400, ctx.Response.StatusCode())
		assert.Equal(t, "Bad", string(ctx.Response.Body()))
	})

	t.Run("JSON Validate Large (Valid)", func(t *testing.T) {
		largeDataValidate := utils.GenerateLargeDataValidate()
		b, _ := json.Marshal(largeDataValidate)

		ctx := doRequest(mux, "POST", "/json-validate-large", "application/json", b)
		assert.Equal(t, 200, ctx.Response.StatusCode())
		assert.Equal(t, "OK", string(ctx.Response.Body()))
	})

	t.Run("JSON Validate Large (Validation Failure)", func(t *testing.T) {
		largeDataValidate := utils.GenerateLargeDataValidate()
		largeDataValidate[0].AccountType = "invalid"
		b, _ := json.Marshal(largeDataValidate)

		ctx := doRequest(mux, "POST", "/json-validate-large", "application/json", b)
		assert.Equal(t, 400, ctx.Response.StatusCode())
		assert.Equal(t, "Bad", string(ctx.Response.Body()))
	})

	t.Run("JSON Validate Response Small", func(t *testing.T) {
		ctx := doRequest(mux, "GET", "/json-response-validate-small", "", nil)
		assert.Equal(t, 200, ctx.Response.StatusCode())

		var decoded utils.SmallPayloadValidate
		err := json.Unmarshal(ctx.Response.Body(), &decoded)
		assert.NoError(t, err)
		assert.Equal(t, 1, decoded.ID)
		assert.Equal(t, "test", decoded.Name)
	})

	t.Run("JSON Validate Response Large", func(t *testing.T) {
		ctx := doRequest(mux, "GET", "/json-response-validate-large", "", nil)
		assert.Equal(t, 200, ctx.Response.StatusCode())

		var decoded []utils.LargePayloadValidate
		err := json.Unmarshal(ctx.Response.Body(), &decoded)
		assert.NoError(t, err)
		assert.Len(t, decoded, 50)
	})
}

func TestBenchmarkEndpoints_OtherForms(t *testing.T) {
	mux := setupBenchmarkRouter()

	t.Run("FormData Bind (Valid)", func(t *testing.T) {
		form := url.Values{}
		form.Add("id", "1")
		form.Add("name", "test")

		ctx := doRequest(mux, "POST", "/formdata", "application/x-www-form-urlencoded", []byte(form.Encode()))
		assert.Equal(t, 200, ctx.Response.StatusCode())
		assert.Equal(t, "OK", string(ctx.Response.Body()))
	})

	t.Run("FormData Bind (Validation Failure)", func(t *testing.T) {
		form := url.Values{}
		form.Add("id", "1")

		ctx := doRequest(mux, "POST", "/formdata", "application/x-www-form-urlencoded", []byte(form.Encode()))
		assert.Equal(t, 400, ctx.Response.StatusCode())
		assert.Equal(t, "Bad", string(ctx.Response.Body()))
	})

	t.Run("Multipart Bind (Valid)", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("id", "1")
		_ = writer.WriteField("name", "test")
		writer.Close()

		ctx := doRequest(mux, "POST", "/multipart", writer.FormDataContentType(), body.Bytes())
		assert.Equal(t, 200, ctx.Response.StatusCode())
		assert.Equal(t, "OK", string(ctx.Response.Body()))
	})
}
