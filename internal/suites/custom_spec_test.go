package suites

import (
	"fmt"
	"strings"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Custom Spec Type for Testing
// =============================================================================

type currencySpec struct{}

func (cs *currencySpec) SpecID() string { return "currency" }
func (cs *currencySpec) Type() string   { return "string" }
func (cs *currencySpec) Format() string { return "currency" }

func (cs *currencySpec) Encode(val any) (string, error) {
	switch v := val.(type) {
	case int:
		dollars := v / 100
		cents := v % 100
		return fmt.Sprintf("$%d.%02d", dollars, cents), nil
	default:
		return "", fmt.Errorf("currency encode: expected int, got %T", val)
	}
}

func (cs *currencySpec) Decode(val any) (any, error) {
	s, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("currency decode: expected string, got %T", val)
	}
	s = strings.TrimPrefix(s, "$")
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("currency decode: invalid format %q", s)
	}
	var dollars, cents int
	fmt.Sscanf(parts[0], "%d", &dollars)
	fmt.Sscanf(parts[1], "%d", &cents)
	return dollars*100 + cents, nil
}

// =============================================================================
// 1. Register and Use Custom Spec on Request
// =============================================================================

func TestCustomSpec_RequestDecode(t *testing.T) {
	type schema struct {
		Request struct {
			Query struct {
				Amount int `json:"amount" spec:"currency" validate:"required"`
			}
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			assert.Equal(t, 1999, s.Request.Query.Amount, "Expected Amount=1999 (cents)")
			return c.SendString(200, "ok")
		},
	}

	m := gofi.NewServeMux()
	m.RegisterSpec(&currencySpec{})

	rec, err := m.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/products",
		Query: map[string]string{
			"amount": "$19.99",
		},
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)
}

// =============================================================================
// 2. Custom Spec on Response Headers
// =============================================================================

func TestCustomSpec_ResponseEncode(t *testing.T) {
	type schema struct {
		Ok struct {
			Header struct {
				Price int `json:"X-Price" spec:"currency"`
			}
			Body string
		}
	}

	handler := gofi.RouteOptions{
		Schema: &schema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[schema](c)
			assert.Nil(t, err)
			s.Ok.Header.Price = 4250
			s.Ok.Body = "priced"
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	m.RegisterSpec(&currencySpec{})

	rec, err := m.Inject(gofi.InjectOptions{
		Method:  "GET",
		Path:    "/price",
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, "$42.50", rec.Header().Get("X-Price"))
}

// =============================================================================
// 3. DefineCustomSpec helper
// =============================================================================

func TestCustomSpec_DefineCustomSpec(t *testing.T) {
	spec := gofi.DefineCustomSpec(gofi.SpecDefinition{
		SpecID: "uppercase",
		Type:   "string",
		Format: "uppercase",
		Encode: func(val any) (string, error) {
			s, ok := val.(string)
			if !ok {
				return "", fmt.Errorf("expected string")
			}
			return strings.ToUpper(s), nil
		},
		Decode: func(val any) (any, error) {
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("expected string")
			}
			return strings.ToUpper(s), nil
		},
	})

	assert.Equal(t, "uppercase", spec.SpecID())
	assert.Equal(t, "string", spec.Type())
	assert.Equal(t, "uppercase", spec.Format())

	encoded, err := spec.Encode("hello")
	assert.Nil(t, err)
	assert.Equal(t, "HELLO", encoded)

	decoded, err := spec.Decode("hello")
	assert.Nil(t, err)
	assert.Equal(t, "HELLO", decoded)
}
