package validations

import (
	"bytes"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

func TestValidation_Hashes(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				MD4       string `json:"md4" validate:"md4"`
				MD5       string `json:"md5" validate:"md5"`
				SHA256    string `json:"sha256" validate:"sha256"`
				SHA384    string `json:"sha384" validate:"sha384"`
				SHA512    string `json:"sha512" validate:"sha512"`
				RIPEMD128 string `json:"ripemd128" validate:"ripemd128"`
				RIPEMD160 string `json:"ripemd160" validate:"ripemd160"`
				Tiger128  string `json:"tiger128" validate:"tiger128"`
				Tiger160  string `json:"tiger160" validate:"tiger160"`
				Tiger192  string `json:"tiger192" validate:"tiger192"`
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
				"md4": "31d6cfe0d16ae931b73c59d7e0c089c0",
				"md5": "d41d8cd98f00b204e9800998ecf8427e",
				"sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				"sha384": "38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b",
				"sha512": "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
				"ripemd128": "996452d5b6b15822f778d06d48259cae",
				"ripemd160": "9c1185a5c5e9fc54612808977ee8f548b2258d31",
				"tiger128": "3293ac630c13f0245f92bbb1766e1616",
				"tiger160": "3293ac630c13f0245f92bbb1766e16167a4e5849",
				"tiger192": "3293ac630c13f0245f92bbb1766e16167a4e58492dde73f3"
			}`,
			expectCode: 200,
		},
		{
			name: "invalid md5",
			// Format the body hashes properly like the first one
			body: `{
				"md5": "invalid",
				"md4": "31d6cfe0d16ae931b73c59d7e0c089c0",
				"sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				"sha384": "38b060a751ac9xxxxxxxx6384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b",
				"sha512": "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
				"ripemd128": "996452xxxxxxd5b6b15822f778d06d48259cae",
				"ripemd160": "9c1185a5c5e9fc54612808977ee8f548b2258d31",
				"tiger128": "3293ac63xxxxxxx0c13f0245f92bbb1766e1616",
				"tiger160": "3293ac630c13f0245f92bbb1766e16167a4e5849",
				"tiger192": "3293ac630c13f0245f92bbb1766e16167a4e58492dde73f3"
			}`,
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
					return c.SendString(200, "okay")
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
