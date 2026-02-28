package codegen

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
// 1. Valid registered custom spec usage
// =============================================================================

func TestCodegen_CustomSpec_ValidUsage(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &CustomSpecSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[CustomSpecSchema](c)
			assert.Nil(t, err)

			// Verify decoded request
			assert.Equal(t, 1999, s.Request.Query.Amount, "Expected Amount=1999 (cents)")

			// Map response
			s.Ok.Header.Price = 4250
			s.Ok.Body = "priced"
			return c.Send(200, s.Ok)
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
	// Verify encoded response header
	assert.Equal(t, "$42.50", rec.Header().Get("X-Price"))
}

// =============================================================================
// 2. Fallback when Custom Spec is missing (Unregistered)
// =============================================================================

func TestCodegen_CustomSpec_MissingFallback(t *testing.T) {
	handler := gofi.RouteOptions{
		Schema: &CustomSpecSchema{},
		Handler: func(c gofi.Context) error {
			s, err := gofi.ValidateAndBind[CustomSpecSchema](c)
			assert.Nil(t, err)

			// Since spec is missing, fallback uses strconv.ParseInt.
			// The query provides "1999" directly (because "$19.99" isn't a valid integer).
			assert.Equal(t, 1999, s.Request.Query.Amount)

			// Map response
			s.Ok.Header.Price = 4250
			s.Ok.Body = "priced"
			return c.Send(200, s.Ok)
		},
	}

	m := gofi.NewServeMux()
	// Deliberately NOT registering currencySpec

	rec, err := m.Inject(gofi.InjectOptions{
		Method: "GET",
		Path:   "/products",
		Query: map[string]string{
			"amount": "1999", // fallback parses this as int cleanly
		},
		Handler: &handler,
	})
	assert.Nil(t, err)
	assert.Equal(t, 200, rec.Code)

	// Verify encoded response header - uses fmt.Sprint fallback since spec is missing
	assert.Equal(t, "4250", rec.Header().Get("X-Price"))
}
