package middleware

import (
	"testing"

	"github.com/michaelolof/gofi"
)

func mustTest(t *testing.T, router gofi.Router, method, path string) *gofi.InjectResponse {
	t.Helper()

	resp, err := router.Test(gofi.TestOptions{Method: method, Path: path})
	if err != nil {
		t.Fatalf("router.Test(%s, %s) failed: %v", method, path, err)
	}

	return resp
}

// must unwraps a middleware constructor that takes a validated config.
func must(mw gofi.MiddlewareFunc, err error) gofi.MiddlewareFunc {
	if err != nil {
		panic(err)
	}
	return mw
}
