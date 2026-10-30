// Package deps gives handlers and middleware access to the service's
// dependencies. Deps is stored in the router's GlobalStore once, at setup.
package deps

import (
	"errors"

	"github.com/michaelolof/gofi"

	"example.com/projectsvc/internal/domain"
)

type depsKey struct{}

// Deps is everything the HTTP layer needs from the rest of the service.
type Deps interface {
	Projects() domain.ProjectStore
	APIKeys() domain.APIKeyStore
	Authenticator() domain.Authenticator
}

// Register stores d in the router's GlobalStore. Call it during setup,
// before registering routes.
func Register(r gofi.Router, d Deps) {
	gofi.StoreSet(r.GlobalStore(), depsKey{}, d)
}

// Get returns the registered Deps, or an error if Register was never called.
func Get(c gofi.Context) (Deps, error) {
	v, ok := gofi.StoreGet(c.GlobalStore(), depsKey{})
	if !ok {
		return nil, errors.New("deps: not registered; call deps.Register(r, d) during setup")
	}
	d, ok := v.(Deps)
	if !ok {
		return nil, errors.New("deps: stored value does not implement Deps")
	}
	return d, nil
}

// MustGet returns the registered Deps and panics if Register was never
// called. A missing registration is a wiring bug, not a request error.
func MustGet(c gofi.Context) Deps {
	d, err := Get(c)
	if err != nil {
		panic(err)
	}
	return d
}

// AppDeps is the Deps implementation built in main and in the test harness.
type AppDeps struct {
	ProjectStore domain.ProjectStore
	APIKeyStore  domain.APIKeyStore
	Authn        domain.Authenticator
}

func (d AppDeps) Projects() domain.ProjectStore       { return d.ProjectStore }
func (d AppDeps) APIKeys() domain.APIKeyStore         { return d.APIKeyStore }
func (d AppDeps) Authenticator() domain.Authenticator { return d.Authn }
