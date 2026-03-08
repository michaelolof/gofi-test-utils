package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/michaelolof/gofi"
)

const PORT string = "6803"

// MultipartSchema defines a complex multipart request.
type MultipartSchema struct {
	Request struct {
		Header struct {
			ContentType string `json:"Content-Type" default:"multipart/form-data"`
		}
		Body struct {
			// Single file upload
			ProfileImage *multipart.FileHeader `json:"profile_image" validate:"required"`
			// Multiple file upload
			Documents []*multipart.FileHeader `json:"documents" validate:"required,min=1"`
			// Standard form fields
			DisplayName string `json:"display_name" validate:"required,min=3"`
			// Array of strings in multipart (sent as multiple fields with same name)
			Tags []string `json:"tags" validate:"required,min=1"`
		}
	}

	Ok struct {
		Body struct {
			Message       string   `json:"message"`
			ImageName     string   `json:"image_name"`
			DocumentCount int      `json:"document_count"`
			DisplayName   string   `json:"display_name"`
			Tags          []string `json:"tags"`
		}
	}
}

func main() {
	r := gofi.NewRouter()

	// Define the multipart handler
	multipartHandler := gofi.DefineHandler(gofi.RouteOptions{
		Schema: &MultipartSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[MultipartSchema](c)
			if err != nil {
				return err
			}

			// Process the profile image
			img, err := s.Request.Body.ProfileImage.Open()
			if err != nil {
				return err
			}
			defer img.Close()

			fmt.Printf("Uploaded Image: %s (%d bytes)\n", s.Request.Body.ProfileImage.Filename, s.Request.Body.ProfileImage.Size)
			fmt.Printf("Uploaded Documents: %d files\n", len(s.Request.Body.Documents))
			fmt.Printf("User: %s, Tags: %v\n", s.Request.Body.DisplayName, s.Request.Body.Tags)

			// Prepare response
			s.Ok.Body.Message = "Upload successful"
			s.Ok.Body.ImageName = s.Request.Body.ProfileImage.Filename
			s.Ok.Body.DocumentCount = len(s.Request.Body.Documents)
			s.Ok.Body.DisplayName = s.Request.Body.DisplayName
			s.Ok.Body.Tags = s.Request.Body.Tags

			return c.Send(http.StatusOK, s.Ok)
		},
	})

	r.Post("/upload", multipartHandler)

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
		testMultipartUpload()
	}()

	log.Printf("Server listening on :%s\n", PORT)
	if err := r.Listen(fmt.Sprintf(":%s", PORT)); err != nil {
		log.Fatal(err)
	}
}

func testMultipartUpload() {
	fmt.Println("\n--- Testing Multipart Upload ---")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 1. Add standard fields
	_ = writer.WriteField("display_name", "Gofi User")
	_ = writer.WriteField("tags", "go")
	_ = writer.WriteField("tags", "api")
	_ = writer.WriteField("tags", "fast")

	// 2. Add profile image
	part, _ := writer.CreateFormFile("profile_image", "avatar.png")
	_, _ = part.Write([]byte("fake-image-binary-data"))

	// 3. Add multiple documents
	for i := 1; i <= 3; i++ {
		docPart, _ := writer.CreateFormFile("documents", fmt.Sprintf("doc%d.txt", i))
		_, _ = docPart.Write([]byte(fmt.Sprintf("content of document %d", i)))
	}

	_ = writer.Close()

	req, _ := http.NewRequest("POST", fmt.Sprintf("http://localhost:%s/upload", PORT), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		fmt.Printf("Upload request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Upload failed with status %d: %s\n", resp.StatusCode, string(respBody))
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", string(respBody))
}
