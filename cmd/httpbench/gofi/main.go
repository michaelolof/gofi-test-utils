package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/michaelolof/gofi-test-utils/internal/utils"

	"github.com/michaelolof/gofi"
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
		FormData utils.SmallPayloadValidate
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
		Multipart utils.SmallPayloadValidate
	}
	Ok struct {
		Body string
	}
	Err struct {
		Body string
	}
}

func main() {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	data := make([]utils.SmallPayload, 100)
	for i := 0; i < 100; i++ {
		data[i] = utils.SmallPayload{ID: i, Name: fmt.Sprintf("Item %d", i)}
	}

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
		Schema: &SmallPayloadSchema{},
		Handler: func(c gofi.Context) error {
			return c.Send(200, data)
		},
	})
	largeData := utils.GenerateLargeData()
	largeDataValidate := utils.GenerateLargeDataValidate()

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
		Schema: &LargePayloadSchema{},
		Handler: func(c gofi.Context) error {
			return c.Send(200, largeData)
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
		Schema: &SmallValidateSchema{},
		Handler: func(c gofi.Context) error {
			return c.Send(200, utils.SmallPayloadValidate{ID: 1, Name: "test"})
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
		Schema: &LargeValidateSchema{},
		Handler: func(c gofi.Context) error {
			return c.Send(200, largeDataValidate)
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

	log.Printf("Gofi listening on :%s\n", port)
	log.Fatal(mux.Listen(":" + port))
}
