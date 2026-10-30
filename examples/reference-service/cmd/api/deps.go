package main

import (
	"example.com/projectsvc/internal/http/deps"
	"example.com/projectsvc/internal/memstore"
)

func newDeps() deps.Deps {
	return deps.AppDeps{
		ProjectStore: memstore.NewProjects(),
		APIKeyStore:  &memstore.APIKeys{},
		Authn:        memstore.Tokens{"dev-token": "user_dev"},
	}
}
