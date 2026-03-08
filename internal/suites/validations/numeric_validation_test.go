package validations

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

func TestValidation_NumericAndGeo(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Numeric   string `json:"numeric" validate:"numeric"`
				Number    string `json:"number" validate:"number"`
				Hex       string `json:"hex" validate:"hexadecimal"`
				Latitude  string `json:"latitude" validate:"latitude"`
				Longitude string `json:"longitude" validate:"longitude"`
			}
		}
	}

	m := gofi.NewRouter()

	tests := []struct {
		name       string
		body       string
		expectCode int
	}{
		{
			name: "valid all",
			body: `{
				"numeric": "12345",
				"number": "1235",
				"hex": "123abc",
				"latitude": "45.0",
				"longitude": "90.0"
			}`,
			expectCode: 200,
		},
		{
			name:       "invalid numeric",
			body:       `{"numeric": "123a", "number": "123", "hex": "abc", "latitude": "0", "longitude": "0"}`,
			expectCode: 500,
		},
		{
			name:       "invalid number",
			body:       `{"numeric": "123", "number": "123.a", "hex": "abc", "latitude": "0", "longitude": "0"}`,
			expectCode: 500,
		},
		{
			name:       "invalid hexadecimal",
			body:       `{"numeric": "123", "number": "123", "hex": "ghi", "latitude": "0", "longitude": "0"}`,
			expectCode: 500,
		},
		{
			name:       "invalid latitude",
			body:       `{"numeric": "123", "number": "123", "hex": "abc", "latitude": "91", "longitude": "0"}`,
			expectCode: 500,
		},
		{
			name:       "invalid longitude",
			body:       `{"numeric": "123", "number": "123", "hex": "abc", "latitude": "0", "longitude": "181"}`,
			expectCode: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := gofi.DefineHandler(gofi.RouteOptions{
				Schema: &schema{},
				Handler: func(c gofi.Context) error {
					_, err := gofi.ValidateAndBind[schema](c)
					if err != nil {
						fmt.Println(err)
						return err
					}
					return c.SendString(200, "ok")
				},
			})

			rec, err := m.Inject(gofi.InjectOptions{
				Method:  "POST",
				Path:    "/test",
				Body:    bytes.NewBufferString(tt.body),
				Handler: &handler,
			})

			assert.Nil(t, err)
			assert.Equal(t, tt.expectCode, rec.StatusCode, tt.name)
		})
	}
}
