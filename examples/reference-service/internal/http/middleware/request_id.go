// Package middleware holds the service's Gofi middleware and its router
// error handler.
package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/michaelolof/gofi"
)

type requestIDKey struct{}

// RequestID reads X-Request-ID or generates one, stores it for later
// middleware and handlers, and echoes it on the response.
func RequestID(c gofi.Context) error {
	rid := c.HeaderVal("x-request-id")
	if rid == "" || len(rid) > 128 {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		rid = hex.EncodeToString(b)
	}
	gofi.StoreSet(c.DataStore(), requestIDKey{}, rid)
	c.Writer().Header().Set("X-Request-ID", rid)
	return c.Next()
}

// GetRequestID returns the ID set by RequestID, or "" if it did not run.
func GetRequestID(c gofi.Context) string {
	v, _ := gofi.StoreGet(c.DataStore(), requestIDKey{})
	rid, _ := v.(string)
	return rid
}
