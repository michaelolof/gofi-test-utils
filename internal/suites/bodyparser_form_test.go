package suites

import (
	"bytes"
	"net/url"
	"strings"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Form Body Binding
// =============================================================================

func TestSchema_FormBody_Struct(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"application/x-www-form-urlencoded"`
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/users",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:    strings.NewReader("name=Alice&age=30"),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
}

func TestSchema_FormBody_Array(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"application/x-www-form-urlencoded"`
			}
			Body struct {
				Tags []string `json:"tags" validate:"required"`
			} `validate:"required"`
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.Equal(t, []string{"id1", "id2"}, s.Request.Body.Tags)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/tags",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:    strings.NewReader("tags=id1&tags=id2"),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
}

func TestSchema_FormBody_ValidationErrors(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"application/x-www-form-urlencoded"`
			}
			Body struct {
				Name string `json:"name" validate:"required"`
				Age  int    `json:"age" validate:"required,min=18"`
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

	m := gofi.NewServeMux()

	// Case 1: Missing required field
	m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/users",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:    strings.NewReader("age=20"),
		Handler: &handler,
	})
	assert.NotNil(t, gotErr)

	// Case 2: Validation min failure
	m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/users",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:    strings.NewReader("name=Alice&age=15"),
		Handler: &handler,
	})
	assert.NotNil(t, gotErr)
}

func TestSchema_FormBody_PointerFields(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"application/x-www-form-urlencoded"`
			}
			Body struct {
				Name     *string `json:"name"`
				Age      *int    `json:"age"`
				IsActive *bool   `json:"is_active"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.NotNil(t, s.Request.Body.Name)
			assert.Equal(t, "Alice", *s.Request.Body.Name)
			assert.NotNil(t, s.Request.Body.Age)
			assert.Equal(t, 30, *s.Request.Body.Age)
			assert.NotNil(t, s.Request.Body.IsActive)
			assert.True(t, *s.Request.Body.IsActive)
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/users",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:    strings.NewReader("name=Alice&age=30&is_active=true"),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
}

func TestSchema_FormBody_ArrayPointer(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"application/x-www-form-urlencoded"`
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/tags",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:    strings.NewReader("tags=id1&tags=id2"),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
}

func TestSchema_FormBody_StructArray(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"application/x-www-form-urlencoded"`
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

	m := gofi.NewServeMux()
	// Using dot notation for nested fields in form
	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/users",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:    strings.NewReader("users.0.name=Alice&users.0.age=30&users.1.name=Bob&users.1.age=25"),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
}

func TestSchema_FormBody_StructArrayPointer(t *testing.T) {
	type User struct {
		Name string `json:"name"`
	}
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"application/x-www-form-urlencoded"`
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

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/users",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:    strings.NewReader("users.0.name=Alice"),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
}

func TestBodyParser_FormBody_FormValueEncode(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"application/x-www-form-urlencoded"`
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

	formData := url.Values{}
	formData.Set("name", "Alice")
	formData.Set("age", "30")

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/users",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:    bytes.NewBufferString(formData.Encode()),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
}

func TestBodyParser_FormBody_SendResponseOkNoHeader(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"application/x-www-form-urlencoded"`
			}
			Body struct {
				Name string `json:"name" validate:"required"`
			} `validate:"required"`
		}

		Ok struct {
			Body struct {
				Message string `json:"message" validate:"required"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.Equal(t, "Alice", s.Request.Body.Name)

			s.Ok.Body.Message = "ok"
			return c.Send(200, s.Ok)
		},
	}

	formData := url.Values{}
	formData.Set("name", "Alice")
	formData.Set("age", "30")

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/users",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:    bytes.NewBufferString(formData.Encode()),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestBodyParser_FormBody_SendResponseOkJSONHeader(t *testing.T) {
	type bodySchema struct {
		Request struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"application/x-www-form-urlencoded"`
			}
			Body struct {
				Name string `json:"name" validate:"required"`
			} `validate:"required"`
		}

		Ok struct {
			Header struct {
				ContentType string `json:"Content-Type" default:"application/json"`
			}
			Body struct {
				Message string `json:"message" validate:"required"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &bodySchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[bodySchema](c)
			assert.Nil(t, err)
			assert.Equal(t, "Alice", s.Request.Body.Name)

			s.Ok.Body.Message = "ok"
			return c.Send(200, s.Ok)
		},
	}

	formData := url.Values{}
	formData.Set("name", "Alice")
	formData.Set("age", "30")

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method: "POST",
		Path:   "/users",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:    bytes.NewBufferString(formData.Encode()),
		Handler: &handler,
	})

	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}
