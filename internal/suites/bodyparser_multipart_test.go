package suites

import (
	"bytes"
	"mime/multipart"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Multipart Body Binding
// =============================================================================

func TestSchema_MultipartBody_Struct(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"multipart/form-data"`
			}
			Body struct {
				Name string `json:"name" validate:"required"`
				Age  int    `json:"age" validate:"required"`
			} `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.Equal(t, "Alice", s.Request.Body.Name)
			assert.Equal(t, 30, s.Request.Body.Age)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("name", "Alice")
	_ = writer.WriteField("age", "30")
	_ = writer.Close()

	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/users",
		Headers: map[string]string{
			"Content-Type": writer.FormDataContentType(),
		},
		Body:    body,
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}

func TestSchema_MultipartBody_Files(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"multipart/form-data"`
			}
			Body struct {
				ProfileImage *multipart.FileHeader   `json:"profile_image" validate:"required"`
				Documents    []*multipart.FileHeader `json:"documents"`
			} `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)

			assert.NotNil(t, s.Request.Body.ProfileImage)
			assert.Equal(t, "image.png", s.Request.Body.ProfileImage.Filename)

			assert.Len(t, s.Request.Body.Documents, 2)
			assert.Equal(t, "doc1.pdf", s.Request.Body.Documents[0].Filename)
			assert.Equal(t, "doc2.pdf", s.Request.Body.Documents[1].Filename)

			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Single file
	part, _ := writer.CreateFormFile("profile_image", "image.png")
	_, _ = part.Write([]byte("fake image data"))

	// Multiple files
	part1, _ := writer.CreateFormFile("documents", "doc1.pdf")
	_, _ = part1.Write([]byte("doc1 data"))
	part2, _ := writer.CreateFormFile("documents", "doc2.pdf")
	_, _ = part2.Write([]byte("doc2 data"))

	_ = writer.Close()

	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/upload",
		Headers: map[string]string{
			"Content-Type": writer.FormDataContentType(),
		},
		Body:    body,
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}

func TestSchema_MultipartBody_Mixed(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"multipart/form-data"`
			}
			Body struct {
				UserName string                `json:"username" validate:"required"`
				Avatar   *multipart.FileHeader `json:"avatar" validate:"required"`
			} `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.Equal(t, "bob", s.Request.Body.UserName)
			assert.NotNil(t, s.Request.Body.Avatar)
			assert.Equal(t, "avatar.jpg", s.Request.Body.Avatar.Filename)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("username", "bob")
	part, _ := writer.CreateFormFile("avatar", "avatar.jpg")
	_, _ = part.Write([]byte("fake avatar"))
	_ = writer.Close()

	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/profile",
		Headers: map[string]string{
			"Content-Type": writer.FormDataContentType(),
		},
		Body:    body,
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}

func TestSchema_MultipartBody_ValidationErrors(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"multipart/form-data"`
			}
			Body struct {
				File *multipart.FileHeader `json:"file" validate:"required"`
			} `validate:"required"`
		}
	}

	var gotErr error
	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			_, err := gofi.ValidateAndBind[bodySchema](c)
			gotErr = err
			return nil
		},
	}

	m := gofi.NewRouter()

	// Missing file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("other", "field")
	_ = writer.Close()

	m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/upload",
		Headers: map[string]string{
			"Content-Type": writer.FormDataContentType(),
		},
		Body:    body,
		Handler: &handler,
	})
	assert.NotNil(t, gotErr, "Expected error for missing required file")
}

func TestSchema_MultipartBody_ArrayPointer(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"multipart/form-data"`
			}
			Body struct {
				Tags *[]string `json:"tags" validate:"required"`
			} `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.NotNil(t, s.Request.Body.Tags)
			assert.Equal(t, []string{"id1", "id2"}, *s.Request.Body.Tags)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("tags", "id1")
	_ = writer.WriteField("tags", "id2")
	_ = writer.Close()

	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/tags",
		Headers: map[string]string{
			"Content-Type": writer.FormDataContentType(),
		},
		Body:    body,
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}

func TestSchema_MultipartBody_StructArray(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"multipart/form-data"`
			}
			Body struct {
				Users []User `json:"users" validate:"required"`
			} `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.Len(t, s.Request.Body.Users, 2)
			assert.Equal(t, "Alice", s.Request.Body.Users[0].Name)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("users.0.name", "Alice")
	_ = writer.WriteField("users.0.age", "30")
	_ = writer.WriteField("users.1.name", "Bob")
	_ = writer.WriteField("users.1.age", "25")
	_ = writer.Close()

	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/users",
		Headers: map[string]string{
			"Content-Type": writer.FormDataContentType(),
		},
		Body:    body,
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}

func TestSchema_MultipartBody_StructArrayPointer(t *testing.T) {
	type User struct {
		Name string `json:"name"`
	}
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"multipart/form-data"`
			}
			Body struct {
				Users *[]User `json:"users" validate:"required"`
			} `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.NotNil(t, s.Request.Body.Users)
			assert.Len(t, *s.Request.Body.Users, 1)
			assert.Equal(t, "Alice", (*s.Request.Body.Users)[0].Name)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewRouter()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("users.0.name", "Alice")
	_ = writer.Close()

	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/users",
		Headers: map[string]string{
			"Content-Type": writer.FormDataContentType(),
		},
		Body:    body,
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.StatusCode)
}
