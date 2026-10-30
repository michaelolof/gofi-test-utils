package middleware

import (
	"strings"

	"github.com/michaelolof/gofi"

	"example.com/projectsvc/internal/apperr"
	"example.com/projectsvc/internal/domain"
)

type principalKey struct{}

// Authenticate verifies the bearer token and stores the caller for handlers.
func Authenticate(authn domain.Authenticator) gofi.MiddlewareFunc {
	return func(c gofi.Context) error {
		token, ok := strings.CutPrefix(c.HeaderVal("authorization"), "Bearer ")
		if !ok || token == "" {
			return apperr.New(1793299514201, apperr.Unauthorized, "missing bearer token")
		}
		p, err := authn.Authenticate(c.Context(), token)
		if err != nil {
			return apperr.Wrap(1793299514202, apperr.Unauthorized, err).WithMsg("invalid bearer token")
		}
		gofi.StoreSet(c.DataStore(), principalKey{}, p)
		return c.Next()
	}
}

// GetPrincipal returns the caller stored by Authenticate.
func GetPrincipal(c gofi.Context) (domain.Principal, bool) {
	v, _ := gofi.StoreGet(c.DataStore(), principalKey{})
	p, ok := v.(domain.Principal)
	return p, ok
}
