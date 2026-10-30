package middleware

import (
	"strings"

	"github.com/michaelolof/gofi"

	"example.com/projectsvc/internal/apperr"
	"example.com/projectsvc/internal/domain"
)

// Authenticate verifies the bearer token and attaches the caller to the
// request context, where handlers and domain code read it with
// domain.PrincipalFrom.
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
		if err := gofi.SetRequestContext(c, domain.WithPrincipal(c.Context(), p)); err != nil {
			return apperr.Wrap(1793299514203, apperr.Internal, err)
		}
		return c.Next()
	}
}
