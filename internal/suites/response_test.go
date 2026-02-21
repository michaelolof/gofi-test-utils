package suites

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 1. SendString & SendBytes
// =============================================================================

func TestResponse_SendString(t *testing.T) {
	handler := gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(200, "hello-world")
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/hello",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "hello-world", rec.Body.String())
}

func TestResponse_SendBytes(t *testing.T) {
	handler := gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendBytes(200, []byte("raw-bytes"))
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/bytes",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "raw-bytes", rec.Body.String())
}

// =============================================================================
// 2. JSON Response via Send
// =============================================================================

func TestResponse_Send_StringBody(t *testing.T) {
	type respSchema struct {
		Ok struct {
			Body string
		}
	}

	handler := gofi.RouteOptions{
		Schema: &respSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[respSchema](c)
			assert.Nil(t, err)
			s.Ok.Body = "plain-response"
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/raw",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
}

// =============================================================================
// 3. Response Headers
// =============================================================================

func TestResponse_Headers(t *testing.T) {
	type respSchema struct {
		Ok struct {
			Header struct {
				RequestID string `json:"X-Request-Id" validate:"required"`
			}
			Body string
		}
	}

	handler := gofi.RouteOptions{
		Schema: &respSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[respSchema](c)
			assert.Nil(t, err)
			s.Ok.Header.RequestID = "resp-123"
			s.Ok.Body = "with-headers"
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, "resp-123", rec.Header().Get("X-Request-Id"))
}

// =============================================================================
// 4. Response Cookies
// =============================================================================

func TestResponse_Cookies(t *testing.T) {
	type respSchema struct {
		Ok struct {
			Cookie struct {
				SessionID http.Cookie `json:"session_id"`
			}
			Body string
		}
	}

	handler := gofi.RouteOptions{
		Schema: &respSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[respSchema](c)
			assert.Nil(t, err)
			s.Ok.Cookie.SessionID = http.Cookie{
				Name:     "session_id",
				Value:    "xyz-789",
				Path:     "/",
				Secure:   true,
				HttpOnly: true,
			}
			s.Ok.Body = "with-cookie"
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/login",
		Body:    strings.NewReader("{}"),
		Handler: &handler,
	})
	assert.Nil(t, err)

	cookies := rec.Result().Cookies()
	assert.NotEmpty(t, cookies, "Expected cookies to be set")
	c := cookies[0]
	assert.Equal(t, "session_id", c.Name)
	assert.Equal(t, "xyz-789", c.Value)
	assert.True(t, c.Secure)
	assert.True(t, c.HttpOnly)
}

// =============================================================================
// 5. Multiple Status Codes
// =============================================================================

func TestResponse_Created(t *testing.T) {
	type schema struct {
		Created struct {
			Body struct {
				ID int `json:"id"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			s.Created.Body.ID = 99
			return c.Send(201, s.Created)
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/items",
		Body:    strings.NewReader("{}"),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 201, rec.Code)

	var result struct {
		ID int `json:"id"`
	}
	json.Unmarshal(rec.Body.Bytes(), &result)
	assert.Equal(t, 99, result.ID)
}

// =============================================================================
// 8. time.Time Header Response
// =============================================================================

func TestResponse_Header_TimeTime(t *testing.T) {
	type respSchema struct {
		Ok struct {
			Header struct {
				ExpiresAt time.Time `json:"X-Expires-At"`
			}
			Body string
		}
	}

	fixedTime := time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC)

	handler := gofi.RouteOptions{
		Schema: &respSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[respSchema](c)
			assert.Nil(t, err)
			s.Ok.Header.ExpiresAt = fixedTime
			s.Ok.Body = "with-time"
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.Nil(t, err)

	headerVal := rec.Header().Get("X-Expires-At")
	assert.NotEmpty(t, headerVal, "Expected X-Expires-At header to be set")

	// Parse the time back from the header — gofi uses RFC3339Nano by default
	parsed, parseErr := time.Parse(time.RFC3339Nano, headerVal)
	assert.Nil(t, parseErr, "Expected header to be valid RFC3339Nano time")
	assert.True(t, fixedTime.Equal(parsed), "Expected parsed time to equal original: got %v vs %v", parsed, fixedTime)
}

// =============================================================================
// 9. *time.Time Pointer Header Response
// =============================================================================

func TestResponse_Header_TimeTimePointer(t *testing.T) {
	type respSchema struct {
		Ok struct {
			Header struct {
				ModifiedAt *time.Time `json:"X-Modified-At"`
			}
			Body string
		}
	}

	fixedTime := time.Date(2026, 6, 15, 8, 30, 0, 0, time.UTC)

	handler := gofi.RouteOptions{
		Schema: &respSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[respSchema](c)
			assert.Nil(t, err)
			s.Ok.Header.ModifiedAt = &fixedTime
			s.Ok.Body = "with-time-ptr"
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.Nil(t, err)

	headerVal := rec.Header().Get("X-Modified-At")
	assert.NotEmpty(t, headerVal, "Expected X-Modified-At header to be set")

	parsed, parseErr := time.Parse(time.RFC3339Nano, headerVal)
	assert.Nil(t, parseErr, "Expected header to be valid RFC3339Nano time")
	assert.True(t, fixedTime.Equal(parsed), "Expected parsed time to equal original: got %v vs %v", parsed, fixedTime)
}

// =============================================================================
// 10. *http.Cookie Pointer Response Cookie
// =============================================================================

func TestResponse_Cookie_Pointer(t *testing.T) {
	type respSchema struct {
		Ok struct {
			Cookie struct {
				Session *http.Cookie `json:"session"`
			}
			Body string
		}
	}

	handler := gofi.RouteOptions{
		Schema: &respSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[respSchema](c)
			assert.Nil(t, err)
			s.Ok.Cookie.Session = &http.Cookie{
				Name:     "session",
				Value:    "ptr-cookie-val",
				Path:     "/",
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}
			s.Ok.Body = "cookie-ptr"
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "POST",
		Path:    "/auth",
		Body:    strings.NewReader("{}"),
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)

	cookies := rec.Result().Cookies()
	assert.NotEmpty(t, cookies, "Expected response cookies to be set")
	ck := cookies[0]
	assert.Equal(t, "session", ck.Name)
	assert.Equal(t, "ptr-cookie-val", ck.Value)
	assert.Equal(t, "/", ck.Path)
	assert.True(t, ck.Secure)
	assert.True(t, ck.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, ck.SameSite)
}

// =============================================================================
// 11. Mixed Pointer / Value Body Struct
// =============================================================================

func TestResponse_Send_MixedPointerValueBody(t *testing.T) {
	type respSchema struct {
		Ok struct {
			Body struct {
				ID      int      `json:"id"`
				Name    *string  `json:"name"`
				Email   string   `json:"email"`
				Score   *float64 `json:"score"`
				IsAdmin *bool    `json:"is_admin"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &respSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[respSchema](c)
			assert.Nil(t, err)

			name := "Bob"
			score := 95.5
			isAdmin := true

			s.Ok.Body.ID = 1
			s.Ok.Body.Name = &name
			s.Ok.Body.Email = "bob@example.com"
			s.Ok.Body.Score = &score
			s.Ok.Body.IsAdmin = &isAdmin
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/profile",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)

	var result struct {
		ID      int      `json:"id"`
		Name    *string  `json:"name"`
		Email   string   `json:"email"`
		Score   *float64 `json:"score"`
		IsAdmin *bool    `json:"is_admin"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	assert.Nil(t, err)
	assert.Equal(t, 1, result.ID)
	assert.NotNil(t, result.Name)
	assert.Equal(t, "Bob", *result.Name)
	assert.Equal(t, "bob@example.com", result.Email)
	assert.NotNil(t, result.Score)
	assert.InDelta(t, 95.5, *result.Score, 0.01)
	assert.NotNil(t, result.IsAdmin)
	assert.True(t, *result.IsAdmin)
}

// =============================================================================
// 13. Response Empty Values — Required Fields Should Fail
// =============================================================================

func TestResponse_EmptyValues_RequiredHeader_ZeroValue(t *testing.T) {
	type schema struct {
		Ok struct {
			Header struct {
				RequestID string `json:"X-Request-Id" validate:"required"`
			}
			Body string
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			// Don't set RequestID — leave it as zero value ""
			s.Ok.Body = "test"
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	rec, _ := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	// Error handler should produce non-200 response because required header is empty
	assert.NotEqual(t, 200, rec.Code, "Expected error for empty required response header")
}

func TestResponse_EmptyValues_RequiredBody_EmptyStruct(t *testing.T) {
	type schema struct {
		Ok struct {
			Body struct {
				Name string `json:"name" validate:"required"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			// Don't set Name — leave as zero value ""
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	rec, _ := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.NotEqual(t, 200, rec.Code, "Expected error for empty required body field")
}

func TestResponse_EmptyValues_RequiredCookie_EmptyValue(t *testing.T) {
	type schema struct {
		Ok struct {
			Cookie struct {
				Token string `json:"token" validate:"required"`
			}
			Body string
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			// Don't set Token — leave as zero value ""
			s.Ok.Body = "test"
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	rec, _ := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.NotEqual(t, 200, rec.Code, "Expected error for empty required response cookie")
}

// =============================================================================
// 14. Response Empty Values — Non-Required Fields Should Pass
// =============================================================================

func TestResponse_EmptyValues_OptionalHeader_ZeroValue(t *testing.T) {
	type schema struct {
		Ok struct {
			Header struct {
				Tag   string `json:"X-Tag"`
				Count int    `json:"X-Count"`
			}
			Body string
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			// Leave all headers as zero values
			s.Ok.Body = "ok"
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
	// Zero-value headers should still be set (as empty string / "0")
	assert.Equal(t, "", rec.Header().Get("X-Tag"))
	assert.Equal(t, "0", rec.Header().Get("X-Count"))
}

func TestResponse_EmptyValues_OptionalHeader_WithDefault(t *testing.T) {
	type schema struct {
		Ok struct {
			Header struct {
				Version string `json:"X-Version" default:"1.0"`
			}
			Body string
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			// Don't set Version — should use default
			s.Ok.Body = "ok"
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "1.0", rec.Header().Get("X-Version"))
}

func TestResponse_EmptyValues_NilPointerBodyFields(t *testing.T) {
	type schema struct {
		Ok struct {
			Body struct {
				Name    *string  `json:"name"`
				Count   *int     `json:"count"`
				Balance *float64 `json:"balance"`
				Active  *bool    `json:"active"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			// Leave all pointers nil
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/test",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)

	var result map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	assert.Nil(t, err)
	// Nil pointers should serialize as null
	assert.Nil(t, result["name"])
	assert.Nil(t, result["count"])
	assert.Nil(t, result["balance"])
	assert.Nil(t, result["active"])
}
