package middleware

import (
	"bytes"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func TestCSRF_SignedTokenFlow(t *testing.T) {
	app := gofi.NewRouter()

	cfg := middleware.CSRFConfig{}
	rv := reflect.ValueOf(&cfg).Elem()

	signTokensField := rv.FieldByName("SignTokens")
	signingKeyField := rv.FieldByName("SigningKey")
	if !signTokensField.IsValid() || !signingKeyField.IsValid() {
		t.Skip("middleware.CSRFConfig on this gofi version does not expose signed-token fields")
	}

	if signTokensField.CanSet() {
		signTokensField.SetBool(true)
	}
	if signingKeyField.CanSet() {
		signingKeyField.SetBytes([]byte("unit-test-signing-key"))
	}

	app.Use(middleware.CSRF(cfg))

	handler := gofi.DefineHandler(gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(http.StatusOK, "ok")
		},
	})

	resp1, err := app.Inject(gofi.InjectOptions{
		Method:  http.MethodGet,
		Path:    "/csrf",
		Handler: &handler,
	})
	if err != nil {
		t.Fatalf("inject get failed: %v", err)
	}

	cookies := resp1.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected csrf cookie to be set")
	}

	csrfCookie := cookies[0]
	if !strings.Contains(csrfCookie.Value, ".") {
		t.Fatalf("expected signed csrf token, got %q", csrfCookie.Value)
	}

	resp2, err := app.Inject(gofi.InjectOptions{
		Method: http.MethodPost,
		Path:   "/csrf",
		Headers: map[string]string{
			"X-CSRF-Token": csrfCookie.Value,
		},
		Cookies: []http.Cookie{*csrfCookie},
		Body:    bytes.NewBufferString("{}"),
		Handler: &handler,
	})
	if err != nil {
		t.Fatalf("inject post failed: %v", err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp2.StatusCode)
	}

	tampered := *csrfCookie
	tampered.Value += "x"
	resp3, err := app.Inject(gofi.InjectOptions{
		Method: http.MethodPost,
		Path:   "/csrf",
		Headers: map[string]string{
			"X-CSRF-Token": tampered.Value,
		},
		Cookies: []http.Cookie{tampered},
		Body:    bytes.NewBufferString("{}"),
		Handler: &handler,
	})
	if err != nil {
		t.Fatalf("inject tampered post failed: %v", err)
	}
	if resp3.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for tampered token, got %d", resp3.StatusCode)
	}

}
