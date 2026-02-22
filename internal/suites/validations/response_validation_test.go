package validations

import (
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

func TestValidation_Response(t *testing.T) {
	type schema struct {
		Ok struct {
			Body struct {
				Email string `json:"email" validate:"email"`
			}
		}
	}

	m := gofi.NewServeMux()

	tests := []struct {
		name       string
		email      string
		expectCode int
		expectMsg  string
	}{
		{
			name:       "valid email",
			email:      "test@example.com",
			expectCode: 200,
		},
		{
			name:       "invalid email",
			email:      "not-an-email",
			expectCode: 500, // Response validation errors usually result in 500
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := gofi.DefineHandler(gofi.RouteOptions{
				Schema: &schema{},
				Handler: func(c gofi.Context) error {
					s := &schema{}
					s.Ok.Body.Email = tt.email
					return c.Send(200, s.Ok)
				},
			})

			rec, err := m.Inject(gofi.InjectOptions{
				Method:  "GET",
				Path:    "/test",
				Handler: &handler,
			})

			assert.Nil(t, err)
			assert.Equal(t, tt.expectCode, rec.Code, tt.name)
		})
	}
}
