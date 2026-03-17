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
