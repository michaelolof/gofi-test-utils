package validations

import (
	"bytes"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

func TestValidation_Network(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				IP           string `json:"ip" validate:"ip"`
				IPv4         string `json:"ipv4" validate:"ipv4"`
				IPv6         string `json:"ipv6" validate:"ipv6"`
				CIDR         string `json:"cidr" validate:"cidr"`
				CIDRv4       string `json:"cidrv4" validate:"cidrv4"`
				CIDRv6       string `json:"cidrv6" validate:"cidrv6"`
				MAC          string `json:"mac" validate:"mac"`
				FQDN         string `json:"fqdn" validate:"fqdn"`
				Hostname     string `json:"hostname" validate:"hostname"`
				HostnamePort string `json:"hostname_port" validate:"hostname_port"`
				URL          string `json:"url" validate:"url"`
				HttpURL      string `json:"http_url" validate:"http_url"`
				URI          string `json:"uri" validate:"uri"`
				URLEncoded   string `json:"url_encoded" validate:"url_encoded"`
				UrnRFC2141   string `json:"urn_rfc2141" validate:"urn_rfc2141"`
				DataURI      string `json:"datauri" validate:"datauri"`
				FileUrl      string `json:"file_url" validate:"fileUrl"`
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
				"ip": "1.1.1.1",
				"ipv4": "192.168.1.1",
				"ipv6": "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
				"cidr": "192.168.1.0/24",
				"cidrv4": "1.1.1.0/24",
				"cidrv6": "2001:db8::/32",
				"mac": "01:23:45:67:89:ab",
				"fqdn": "www.google.com",
				"hostname": "google.com",
				"hostname_port": "localhost:8080",
				"url": "https://google.com",
				"http_url": "http://google.com",
				"uri": "mailto:test@example.com",
				"url_encoded": "http%3A%2F%2Fgoogle.com%2F",
				"urn_rfc2141": "urn:ietf:rfc:2141",
				"datauri": "data:text/plain;base64,SGVsbG8sIFdvcmxkIQ==",
				"file_url": "file:///path/to/file"
			}`,
			expectCode: 200,
		},
		{
			name:       "invalid ip",
			body:       `{"ip": "invalid"}`,
			expectCode: 500,
		},
		{
			name:       "invalid url",
			body:       `{"url": "invalid"}`,
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
