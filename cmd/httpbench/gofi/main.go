package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/debug"

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
		Body []utils.LargePayloadGofi
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
		Body []utils.LargePayloadGofi
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

type ws6RuntimeStats struct {
	HeapAlloc        uint64 `json:"heapAlloc"`
	HeapInuse        uint64 `json:"heapInuse"`
	HeapIdle         uint64 `json:"heapIdle"`
	HeapReleased     uint64 `json:"heapReleased"`
	TotalAlloc       uint64 `json:"totalAlloc"`
	NumGC            uint32 `json:"numGC"`
	PauseTotalNs     uint64 `json:"pauseTotalNs"`
	LastGCPauseNs    uint64 `json:"lastGCPauseNs"`
	Goroutines       int    `json:"goroutines"`
	ForcedCollection bool   `json:"forcedCollection"`
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

	// WS-6 uses this benchmark-only endpoint to observe the server process
	// after each load phase. It is intentionally not part of Gofi itself.
	mux.Get("/__ws6/runtime", gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			forced := c.Query("gc") == "1"
			if forced {
				runtime.GC()
				debug.FreeOSMemory()
			}

			var memory runtime.MemStats
			runtime.ReadMemStats(&memory)
			lastPause := uint64(0)
			if memory.NumGC > 0 {
				lastPause = memory.PauseNs[(memory.NumGC+255)%256]
			}
			payload, err := json.Marshal(ws6RuntimeStats{
				HeapAlloc:        memory.HeapAlloc,
				HeapInuse:        memory.HeapInuse,
				HeapIdle:         memory.HeapIdle,
				HeapReleased:     memory.HeapReleased,
				TotalAlloc:       memory.TotalAlloc,
				NumGC:            memory.NumGC,
				PauseTotalNs:     memory.PauseTotalNs,
				LastGCPauseNs:    lastPause,
				Goroutines:       runtime.NumGoroutine(),
				ForcedCollection: forced,
			})
			if err != nil {
				return c.SendString(500, "runtime stats unavailable")
			}
			c.Writer().Header().Set("Content-Type", "application/json")
			return c.SendBytes(200, payload)
		},
	})

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
		Schema: &SmallResponseSchema{},
		Handler: func(c gofi.Context) error {
			response := SmallResponseSchema{}
			response.Ok.Body = data
			return c.Send(200, response.Ok)
		},
	})
	largeData := utils.GenerateLargeData()
	largeDataValidate := utils.GenerateLargeDataGofi()

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
		Schema: &LargeResponseSchema{},
		Handler: func(c gofi.Context) error {
			response := LargeResponseSchema{}
			response.Ok.Body = largeData
			return c.Send(200, response.Ok)
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
		Schema: &SmallValidateResponseSchema{},
		Handler: func(c gofi.Context) error {
			response := SmallValidateResponseSchema{}
			response.Ok.Body = utils.SmallPayloadValidate{ID: 1, Name: "test"}
			return c.Send(200, response.Ok)
		},
	})

	mux.Post("/json-validate-large", gofi.RouteOptions{
		Schema: &LargeValidateSchema{},
		Handler: func(c gofi.Context) error {
			if err := gofi.Validate(c); err != nil {
				return c.SendString(400, "Bad")
			}
			return c.SendString(200, "OK")
		},
	})

	mux.Post("/json-bind-validate-large", gofi.RouteOptions{
		Schema: &LargeValidateSchema{},
		Handler: func(c gofi.Context) error {
			if _, err := gofi.ValidateAndBind[LargeValidateSchema](c); err != nil {
				return c.SendString(400, "Bad")
			}
			return c.SendString(200, "OK")
		},
	})

	mux.Get("/json-response-validate-large", gofi.RouteOptions{
		Schema: &LargeValidateResponseSchema{},
		Handler: func(c gofi.Context) error {
			response := LargeValidateResponseSchema{}
			response.Ok.Body = largeDataValidate
			return c.Send(200, response.Ok)
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
