package main

import (
	"bufio"
	"fmt"
	"time"

	"github.com/michaelolof/gofi"
)

func main() {
	r := gofi.NewRouter()

	type streamSchema struct {
		Ok struct {
			Header struct {
				ContentType string `json:"content-type" default:"text/event-stream"`
			}
			Body string `validate:"required" pattern:"event-stream"`
		}
	}

	var streamHandler = gofi.DefineHandler(gofi.RouteOptions{

		Schema: &streamSchema{},

		Handler: func(c gofi.Context) error {
			var s streamSchema
			return c.SendStream(200, s, func(w *bufio.Writer) error {
				for i := 0; i < 10; i++ {
					// Write a chunk of data
					if _, err := fmt.Fprintf(w, "data: Chunk %d\n\n", i+1); err != nil {
						return err
					}

					// Flush the writer to send the data immediately
					if err := w.Flush(); err != nil {
						return err
					}

					// Simulate a delay between chunks
					time.Sleep(1 * time.Second)
				}
				return nil
			})
		},
	})

	r.Get("/stream", streamHandler)

	// Start the server
	if err := gofi.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
