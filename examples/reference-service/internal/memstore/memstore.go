// Package memstore is an in-memory implementation of the domain stores, used
// by the demo binary and the test harness.
package memstore

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"example.com/projectsvc/internal/domain"
)

type Projects struct {
	mu   sync.Mutex
	byID map[string]domain.Project
}

func NewProjects() *Projects { return &Projects{byID: map[string]domain.Project{}} }

func (s *Projects) Create(_ context.Context, p domain.Project) (domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p.ID = fmt.Sprintf("prj_%d", len(s.byID)+1)
	p.CreatedAt = time.Now().UTC()
	s.byID[p.ID] = p
	return p, nil
}

func (s *Projects) Get(_ context.Context, ownerID, id string) (domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.byID[id]
	if !ok || p.OwnerID != ownerID {
		return domain.Project{}, domain.ErrNotFound
	}
	return p, nil
}

type APIKeys struct {
	mu sync.Mutex
	n  int
}

func (s *APIKeys) Create(_ context.Context, projectID string) (domain.APIKey, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.n++
	return domain.APIKey{ID: fmt.Sprintf("key_%d", s.n), ProjectID: projectID, Prefix: "sk_test"}, fmt.Sprintf("sk_test_secret_%d", s.n), nil
}

// Tokens authenticates a fixed set of bearer tokens.
type Tokens map[string]string

func (t Tokens) Authenticate(_ context.Context, token string) (domain.Principal, error) {
	userID, ok := t[token]
	if !ok {
		return domain.Principal{}, errors.New("unknown token")
	}
	return domain.Principal{UserID: userID}, nil
}
