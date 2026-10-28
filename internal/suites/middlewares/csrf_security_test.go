package middleware

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/middleware"
)

func newCSRFApp(t *testing.T) (gofi.Router, *gofi.RouteOptions) {
	t.Helper()
	app := gofi.NewRouter()
	app.UseErrorHandler(func(err error, c gofi.Context) {
		_ = c.SendString(http.StatusForbidden, err.Error())
	})
	app.Use(mustCSRF(t, middleware.CSRFConfig{}))
	handler := gofi.DefineHandler(gofi.RouteOptions{
		Handler: func(c gofi.Context) error {
			return c.SendString(http.StatusOK, "ok")
		},
	})
	return app, &handler
}

func issueCSRF(t *testing.T, app gofi.Router, handler *gofi.RouteOptions, session string) *http.Cookie {
	t.Helper()
	opts := gofi.InjectOptions{Method: http.MethodGet, Path: "/csrf", Handler: handler}
	if session != "" {
		opts.Headers = map[string]string{"X-Session": session}
	}
	resp, err := app.Inject(opts)
	if err != nil {
		t.Fatalf("inject get failed: %v", err)
	}
	return csrfCookieFrom(t, resp)
}

func csrfPost(t *testing.T, app gofi.Router, handler *gofi.RouteOptions, cookie http.Cookie, session string, extra map[string]string) *gofi.InjectResponse {
	t.Helper()
	headers := map[string]string{"X-CSRF-Token": cookie.Value}
	if session != "" {
		headers["X-Session"] = session
	}
	for k, v := range extra {
		headers[k] = v
	}
	resp, err := app.Inject(gofi.InjectOptions{
		Method:  http.MethodPost,
		Path:    "/csrf",
		Headers: headers,
		Cookies: []http.Cookie{cookie},
		Body:    bytes.NewBufferString("{}"),
		Handler: handler,
	})
	if err != nil {
		t.Fatalf("inject post failed: %v", err)
	}
	return resp
}

func TestCSRF_SignedTokenFlow(t *testing.T) {
	app, handler := newCSRFApp(t)
	cookie := issueCSRF(t, app, handler, "alice")

	if parts := strings.Split(cookie.Value, "."); len(parts) != 4 || parts[0] != "v1" {
		t.Fatalf("expected signed v1 token, got %q", cookie.Value)
	}
	if resp := csrfPost(t, app, handler, *cookie, "alice", nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	tampered := *cookie
	tampered.Value += "x"
	resp := csrfPost(t, app, handler, tampered, "alice", nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for tampered token, got %d", resp.StatusCode)
	}
	if string(resp.Body) != "invalid CSRF token" {
		t.Fatalf("expected invalid csrf token body, got %q", string(resp.Body))
	}
}

func TestCSRF_TokenIsBoundToSession(t *testing.T) {
	app, handler := newCSRFApp(t)

	// A token issued to one session must not validate for another.
	cookie := issueCSRF(t, app, handler, "alice")
	if resp := csrfPost(t, app, handler, *cookie, "mallory", nil); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for a token from another session, got %d", resp.StatusCode)
	}

	// A token issued before login must not survive it (session fixation).
	anon := issueCSRF(t, app, handler, "")
	if resp := csrfPost(t, app, handler, *anon, "alice", nil); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for a pre-login token after login, got %d", resp.StatusCode)
	}
}

func TestCSRF_ForgedUnsignedTokenRejected(t *testing.T) {
	app, handler := newCSRFApp(t)

	// Plain double-submit accepted any matching cookie and header pair.
	forged := http.Cookie{Name: "csrf_", Value: "attacker-chosen"}
	if resp := csrfPost(t, app, handler, forged, "alice", nil); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for an unsigned token, got %d", resp.StatusCode)
	}
}

func TestCSRF_CrossOriginRejected(t *testing.T) {
	app, handler := newCSRFApp(t)
	cookie := issueCSRF(t, app, handler, "alice")

	resp := csrfPost(t, app, handler, *cookie, "alice", map[string]string{"Origin": "https://evil.example"})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for a cross-origin request with a valid token, got %d", resp.StatusCode)
	}
}
