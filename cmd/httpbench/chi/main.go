package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/michaelolof/gofi-test-utils/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func main() {
	port := "8081"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	data := make([]utils.SmallPayload, 100)
	for i := 0; i < 100; i++ {
		data[i] = utils.SmallPayload{ID: i, Name: fmt.Sprintf("Item %d", i)}
	}
	largeData := utils.GenerateLargeData()
	largeDataValidate := utils.GenerateLargeDataValidate()

	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Get("/user/{name}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(chi.URLParam(r, "name")))
	})

	r.Get("/users/{userID}/posts/{postID}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(chi.URLParam(r, "userID") + ":" + chi.URLParam(r, "postID")))
	})

	r.Post("/json", func(w http.ResponseWriter, r *http.Request) {
		var p utils.SmallPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Bad", 400)
			return
		}
		w.Write([]byte("OK"))
	})

	r.Get("/json-response", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	r.Post("/json-large", func(w http.ResponseWriter, r *http.Request) {
		var p []utils.LargePayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Bad", 400)
			return
		}
		w.Write([]byte("OK"))
	})

	r.Get("/json-response-large", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(largeData)
	})

	r.Post("/json-validate-small", func(w http.ResponseWriter, r *http.Request) {
		var p utils.SmallPayloadValidate
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Bad", 400)
			return
		}
		if err := validate.Struct(p); err != nil {
			http.Error(w, "Bad", 400)
			return
		}
		w.Write([]byte("OK"))
	})

	r.Get("/json-response-validate-small", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(utils.SmallPayloadValidate{ID: 1, Name: "test"})
	})

	r.Post("/json-validate-large", func(w http.ResponseWriter, r *http.Request) {
		var p []utils.LargePayloadValidate
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Bad", 400)
			return
		}
		for _, item := range p {
			if err := validate.Struct(item); err != nil {
				http.Error(w, "Bad", 400)
				return
			}
		}
		w.Write([]byte("OK"))
	})

	r.Get("/json-response-validate-large", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(largeDataValidate)
	})

	r.Post("/multipart", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "Bad", 400)
			return
		}
		var p utils.SmallPayloadValidate
		fmt.Sscanf(r.FormValue("id"), "%d", &p.ID)
		p.Name = r.FormValue("name")
		if err := validate.Struct(p); err != nil {
			http.Error(w, "Bad", 400)
			return
		}
		w.Write([]byte("OK"))
	})

	r.Post("/formdata", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad", 400)
			return
		}
		var p utils.SmallPayloadValidate
		fmt.Sscanf(r.FormValue("id"), "%d", &p.ID)
		p.Name = r.FormValue("name")
		if err := validate.Struct(p); err != nil {
			http.Error(w, "Bad", 400)
			return
		}
		w.Write([]byte("OK"))
	})

	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
	r.With(mw, mw, mw, mw, mw).Get("/middlewares", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Get("/query", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		limit := r.URL.Query().Get("limit")
		w.Write([]byte(q + limit))
	})

	log.Printf("Chi listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
