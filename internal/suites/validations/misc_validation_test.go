package validations

import (
	"bytes"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

func TestValidation_Misc(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				Boolean           string `json:"boolean" validate:"boolean"`
				JSON              string `json:"json" validate:"json"`
				Datetime          string `json:"datetime" validate:"datetime=2006-01-02"`
				Timezone          string `json:"timezone" validate:"timezone"`
				CreditCard        string `json:"credit_card" validate:"credit_card"`
				EIN               string `json:"ein" validate:"ein"`
				LuhnChecksum      string `json:"luhn_checksum" validate:"luhn_checksum"`
				Cron              string `json:"cron" validate:"cron"`
				JWT               string `json:"jwt" validate:"jwt"`
				HTML              string `json:"html" validate:"html"`
				HTMLEncoded       string `json:"html_encoded" validate:"html_encoded"`
				SpiceDBID         string `json:"spicedb_id" validate:"spicedb_id"`
				SpiceDBPermission string `json:"spicedb_permission" validate:"spicedb_permission"`
				SpiceDBType       string `json:"spicedb_type" validate:"spicedb_type"`
				BtcAddr           string `json:"btc_addr" validate:"btc_addr"`
				BtcAddrBech32     string `json:"btc_addr_bech32" validate:"btc_addr_bech32"`
				EthAddr           string `json:"eth_addr" validate:"eth_addr"`
				MongoDB           string `json:"mongodb" validate:"mongodb"`
				Default           string `json:"default" validate:"isdefault"`
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
				"boolean": "true",
				"json": "{\"k\":\"v\"}",
				"datetime": "2023-10-27",
				"timezone": "UTC",
				"credit_card": "4111111111111111",
				"ein": "12-3456789",
				"luhn_checksum": "79927398713",
				"cron": "* * * * *",
				"jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.t-99p_tX9p_tX9p_tX9p_tX9p_tX9p_tX9p_tX9p_tX",
				"html": "<div></div>",
				"html_encoded": "&lt;div&gt;&lt;/div&gt;",
				"spicedb_id": "user123",
				"spicedb_permission": "read",
				"spicedb_type": "user",
				"btc_addr": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
				"btc_addr_bech32": "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh",
				"eth_addr": "0x323b5d4ba3602ed4b0791a61d4074223d6b086ed",
				"mongodb": "507f1f77bcf86cd799439011",
				"default": ""
			}`,
			expectCode: 200,
		},
		{
			name:       "invalid boolean",
			body:       `{"boolean": "invalid"}`,
			expectCode: 500,
		},
		{
			name:       "invalid datetime",
			body:       `{"datetime": "2023-10-27 10:00:00"}`,
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
			assert.Equal(t, tt.expectCode, rec.StatusCode, tt.name)
		})
	}
}
