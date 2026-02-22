package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/michaelolof/gofi"
)

const PORT string = "688001"

func main() {
	r := gofi.NewServeMux()

	type userFormSchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"application/x-www-form-urlencoded"`
			}
			Body struct {
				Username string `json:"username" validate:"required,min=3"`
				Email    string `json:"email" validate:"required,email"`
				Age      int    `json:"age" validate:"required,gt=10"`
			}
		}

		Ok struct {
			Body struct {
				Message string `json:"message"`
			}
		}
	}

	userFormHandler := gofi.DefineHandler(gofi.RouteOptions{
		Schema: &userFormSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[userFormSchema](c)
			if err != nil {
				return err
			}

			fmt.Printf("Received Form: Username=%s, Email=%s, Age=%d\n",
				s.Request.Body.Username, s.Request.Body.Email, s.Request.Body.Age)

			s.Ok.Body.Message = fmt.Sprintf("User %s created successfully", s.Request.Body.Username)
			return c.Send(http.StatusOK, s.Ok)
		},
	})

	r.Post("/user", userFormHandler)

	gofi.ServeDocs(r, gofi.DocsOptions{
		Info: gofi.DocsInfoOptions{
			Title: "API Docs",
		},
		Views: []gofi.DocsView{
			{RoutePrefix: "/docs", Template: gofi.StopLight()},
		},
	})

	// Run tests in a goroutine
	go func() {
		time.Sleep(2 * time.Second)
		testFormData()
	}()

	log.Printf("Server listening on :%s\n", PORT)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", PORT), r); err != nil {
		log.Fatal(err)
	}
}

func testFormData() {
	fmt.Println("\n--- Testing Form Data ---")
	formData := url.Values{}
	formData.Set("username", "johndoe")
	formData.Set("email", "john@example.com")
	formData.Set("age", "30")

	req, _ := http.NewRequest("POST", fmt.Sprintf("http://localhost:%s/user", PORT), strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		fmt.Printf("Form request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Form request failed: %s\n", string(body))
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", string(body))
}
