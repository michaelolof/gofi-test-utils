package validations

import (
	"bytes"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

func TestValidation_Comparison(t *testing.T) {
	type schema struct {
		Request struct {
			Body struct {
				EqValue  int    `json:"eq_value" validate:"eq=10"`
				NeValue  int    `json:"ne_value" validate:"ne=10"`
				GtValue  int    `json:"gt_value" validate:"gt=10"`
				GteValue int    `json:"gte_value" validate:"gte=10"`
				LtValue  int    `json:"lt_value" validate:"lt=10"`
				LteValue int    `json:"lte_value" validate:"lte=10"`
				LenValue string `json:"len_value" validate:"len=5"`
				OneOf    string `json:"one_of" validate:"oneof=a b c"`
				Required string `json:"required" validate:"required"`
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
				"eq_value": 10,
				"ne_value": 11,
				"gt_value": 11,
				"gte_value": 10,
				"lt_value": 9,
				"lte_value": 10,
				"len_value": "abcde",
				"one_of": "a",
				"required": "present"
			}`,
			expectCode: 200,
		},
		{
			name:       "invalid eq",
			body:       `{"eq_value": 11, "ne_value": 11, "gt_value": 11, "gte_value": 10, "lt_value": 9, "lte_value": 10, "len_value": "abcde", "one_of": "a", "required": "present"}`,
			expectCode: 500,
		},
		{
			name:       "invalid ne",
			body:       `{"eq_value": 10, "ne_value": 10, "gt_value": 11, "gte_value": 10, "lt_value": 9, "lte_value": 10, "len_value": "abcde", "one_of": "a", "required": "present"}`,
			expectCode: 500,
		},
		{
			name:       "invalid gt",
			body:       `{"eq_value": 10, "ne_value": 11, "gt_value": 10, "gte_value": 10, "lt_value": 9, "lte_value": 10, "len_value": "abcde", "one_of": "a", "required": "present"}`,
			expectCode: 500,
		},
		{
			name:       "invalid gte",
			body:       `{"eq_value": 10, "ne_value": 11, "gt_value": 11, "gte_value": 9, "lt_value": 9, "lte_value": 10, "len_value": "abcde", "one_of": "a", "required": "present"}`,
			expectCode: 500,
		},
		{
			name:       "invalid lt",
			body:       `{"eq_value": 10, "ne_value": 11, "gt_value": 11, "gte_value": 10, "lt_value": 10, "lte_value": 10, "len_value": "abcde", "one_of": "a", "required": "present"}`,
			expectCode: 500,
		},
		{
			name:       "invalid lte",
			body:       `{"eq_value": 10, "ne_value": 11, "gt_value": 11, "gte_value": 10, "lt_value": 9, "lte_value": 11, "len_value": "abcde", "one_of": "a", "required": "present"}`,
			expectCode: 500,
		},
		{
			name:       "invalid len",
			body:       `{"eq_value": 10, "ne_value": 11, "gt_value": 11, "gte_value": 10, "lt_value": 9, "lte_value": 10, "len_value": "abcd", "one_of": "a", "required": "present"}`,
			expectCode: 500,
		},
		{
			name:       "invalid oneof",
			body:       `{"eq_value": 10, "ne_value": 11, "gt_value": 11, "gte_value": 10, "lt_value": 9, "lte_value": 10, "len_value": "abcde", "one_of": "d", "required": "present"}`,
			expectCode: 500,
		},
		{
			name:       "missing required",
			body:       `{"eq_value": 10, "ne_value": 11, "gt_value": 11, "gte_value": 10, "lt_value": 9, "lte_value": 10, "len_value": "abcde", "one_of": "a"}`,
			expectCode: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := gofi.DefineHandler(gofi.RouteOptions{
				Schema: &schema{},
				Handler: func(c gofi.Context) error {
					s, err := gofi.ValidateAndBind[schema](c)
					if err != nil {
						return err
					}

					ignore(s)
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

func ignore(v any) {

}
