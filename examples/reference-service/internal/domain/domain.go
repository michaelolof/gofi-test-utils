// Package domain holds the service's business types and the interfaces the
// HTTP layer depends on. It does not import Gofi.
package domain

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned by stores when a record does not exist.
var ErrNotFound = errors.New("not found")

type Project struct {
	ID        string
	Name      string
	OwnerID   string
	CreatedAt time.Time
}

type ProjectStore interface {
	Create(ctx context.Context, p Project) (Project, error)
	Get(ctx context.Context, ownerID, id string) (Project, error)
}

type APIKey struct {
	ID        string
	ProjectID string
	Prefix    string
}

type APIKeyStore interface {
	Create(ctx context.Context, projectID string) (APIKey, string, error)
}

// Principal is the authenticated caller.
type Principal struct {
	UserID string
}

type principalKey struct{}

// WithPrincipal returns ctx carrying the authenticated caller.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

// PrincipalFrom returns the caller stored by WithPrincipal.
func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(Principal)
	return p, ok
}

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (Principal, error)
}
