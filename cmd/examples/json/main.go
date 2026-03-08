package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

const PORT string = "6804"

// UserProfile represents a nested structure in the JSON schema.
type UserProfile struct {
	Bio      string `json:"bio" validate:"max=200"`
	Website  string `json:"website" validate:"omitempty,url"`
	Location string `json:"location"`
}

// JSONSchema defines a comprehensive JSON request and response.
type JSONSchema struct {
	Request struct {
		Body struct {
			// Basic types
			Username string  `json:"username" validate:"required,min=3,max=32"`
			Email    string  `json:"email" validate:"required,email"`
			Age      int     `json:"age" validate:"required,min=18,max=120"`
			Score    float64 `json:"score" validate:"min=0,max=100"`
			Active   bool    `json:"active"`

			// Time
			JoinedAt time.Time `json:"joined_at" validate:"required"`

			// Slices and Maps
			Tags     []string          `json:"tags" validate:"min=1,max=5"`
			Metadata map[string]string `json:"metadata"`

			// Nested Struct
			Profile UserProfile `json:"profile"`

			// Pointer and Optional fields
			Nickname *string `json:"nickname" validate:"omitempty,min=2"`
			Referrer string  `json:"referrer" default:"none"`
		}
	}

	Ok struct {
		Body struct {
			Message   string    `json:"message"`
			Timestamp time.Time `json:"timestamp"`
			UserID    string    `json:"user_id"`
			Received  struct {
				Username string    `json:"username"`
				JoinedAt time.Time `json:"joined_at"`
				Tags     []string  `json:"tags"`
			} `json:"received"`
		}
	}
}

func main() {
	r := gofi.NewRouter()

	// Apply robust Core Middlewares
	r.Use(middleware.Recover())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Compress())
	r.Use(middleware.ResponseTime())

	r.RegisterBodyParser(&gofi.JSONBodyParser{MaintainOrder: true})

	// Define the JSON handler
	jsonHandler := gofi.DefineHandler(gofi.RouteOptions{
		Schema: &JSONSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[JSONSchema](c)
			if err != nil {
				return err
			}

			fmt.Printf("Received Request for User: %s\n", s.Request.Body.Username)
			fmt.Printf("Email: %s, Age: %d, Active: %v\n", s.Request.Body.Email, s.Request.Body.Age, s.Request.Body.Active)
			fmt.Printf("Joined At: %s\n", s.Request.Body.JoinedAt)
			fmt.Printf("Tags: %v\n", s.Request.Body.Tags)
			fmt.Printf("Profile Bio: %s\n", s.Request.Body.Profile.Bio)

			if s.Request.Body.Nickname != nil {
				fmt.Printf("Nickname: %s\n", *s.Request.Body.Nickname)
			}
			fmt.Printf("Referrer: %s\n", s.Request.Body.Referrer)

			// Prepare response
			s.Ok.Body.Message = "User data processed successfully"
			s.Ok.Body.Timestamp = time.Now()
			s.Ok.Body.UserID = "user_12345"
			s.Ok.Body.Received.Username = s.Request.Body.Username
			s.Ok.Body.Received.JoinedAt = s.Request.Body.JoinedAt
			s.Ok.Body.Received.Tags = s.Request.Body.Tags

			return c.Send(http.StatusOK, s.Ok)
		},
	})

	r.Post("/users", jsonHandler)

	// Serve API Documentation
	gofi.ServeDocs(r, gofi.DocsOptions{
		Info: gofi.DocsInfoOptions{
			Title:       "Gofi JSON Example API",
			Description: "An extensive example of JSON request and response handling in Gofi.",
			Version:     "1.0.0",
		},
		Views: []gofi.DocsView{
			{RoutePrefix: "/docs", Template: gofi.StopLight()},
		},
	})

	// Run an automated test in a separate goroutine
	go func() {
		time.Sleep(2 * time.Second)
		testJSONService()
	}()

	log.Printf("JSON Example Server listening on :%s\n", PORT)
	log.Printf("Documentation available at http://localhost:%s/docs\n", PORT)

	if err := r.Listen(fmt.Sprintf(":%s", PORT)); err != nil {
		log.Fatal(err)
	}
}

func testJSONService() {
	fmt.Println("\n--- Testing JSON Service ---")

	nickname := "GoGopher"
	payload := map[string]interface{}{
		"username":  "john_doe",
		"email":     "john@example.com",
		"age":       25,
		"score":     95.5,
		"active":    true,
		"joined_at": time.Now().Format(time.RFC3339),
		"tags":      []string{"golang", "api", "gofi"},
		"metadata": map[string]string{
			"source": "example_client",
		},
		"profile": map[string]string{
			"bio":      "A passionate developer.",
			"website":  "https://example.com",
			"location": "San Francisco",
		},
		"nickname": &nickname,
	}

	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", fmt.Sprintf("http://localhost:%s/users", PORT), bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		fmt.Printf("JSON request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", string(respBody))

	if resp.StatusCode == http.StatusOK {
		fmt.Println("--- JSON Service Test Passed ---")
	} else {
		fmt.Printf("--- JSON Service Test Failed with status %d ---\n", resp.StatusCode)
	}
}
