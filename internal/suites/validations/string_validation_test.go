package validations

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

func TestValidation_StringsExhaustive(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Alpha           string `json:"alpha" validate:"alpha"`
				Alphanum        string `json:"alphanum" validate:"alphanum"`
				AlphaUnicode    string `json:"alphaunicode" validate:"alphaunicode"`
				AlphaUnicodeNum string `json:"alphaunicodenum" validate:"alphaunicodenum"`
				Ascii           string `json:"ascii" validate:"ascii"`
				PrintAscii      string `json:"printascii" validate:"printascii"`
				Multibyte       string `json:"multibyte" validate:"multibyte"`
				Email           string `json:"email" validate:"email"`
				E164            string `json:"e164" validate:"e164"`
				Isbn10          string `json:"isbn10" validate:"isbn10"`
				Isbn13          string `json:"isbn13" validate:"isbn13"`
				Issn            string `json:"issn" validate:"issn"`
				UUID            string `json:"uuid" validate:"uuid"`
				ULID            string `json:"ulid" validate:"ulid"`
				SSN             string `json:"ssn" validate:"ssn"`
				BIC             string `json:"bic" validate:"bic"`
				Semver          string `json:"semver" validate:"semver"`
				CVE             string `json:"cve" validate:"cve"`
			}
		}
	}

	m := gofi.NewServeMux()

	tests := []struct {
		name       string
		body       string
		expectCode int
	}{
		{
			name: "valid all",
			body: `{
				"alpha": "abc",
				"alphanum": "abc123",
				"alphaunicode": "abc世界",
				"alphaunicodenum": "abc123世界",
				"ascii": "abc!@#",
				"printascii": "abc !@#",
				"multibyte": "世界",
				"email": "test@example.com",
				"e164": "+1234567890",
				"isbn10": "0306406152",
				"isbn13": "9780306406157",
				"issn": "2049-3630",
				"uuid": "550e8400-e29b-41d4-a716-446655440000",
				"ulid": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
				"ssn": "666-55-4321",
				"bic": "ABCDEFF1XXX",
				"semver": "1.2.3",
				"cve": "CVE-2021-12345"
			}`,
			expectCode: 200,
		},
		{
			name:       "invalid email",
			body:       `{"email": "not-an-email"}`,
			expectCode: 500,
		},
		{
			name:       "invalid uuid",
			body:       `{"uuid": "not-a-uuid"}`,
			expectCode: 500,
		},
		{
			name:       "invalid semver",
			body:       `{"semver": "1.2"}`,
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
			assert.Equal(t, tt.expectCode, rec.Code, tt.name)
		})
	}
}
