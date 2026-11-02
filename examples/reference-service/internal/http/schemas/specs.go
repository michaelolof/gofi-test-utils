package schemas

import (
	"errors"
	"strings"

	"github.com/michaelolof/gofi"
)

// ProjectID identifies a project. A malformed ID never reaches a handler.
type ProjectID string

// ProjectIDSpec reads and writes a ProjectID. server.New registers it before
// the routes that use it.
var ProjectIDSpec = gofi.DefineCustomSpec(gofi.SpecDefinition{
	SpecID: "project-id",
	Type:   "string",
	Encode: func(val any) (string, error) {
		id, ok := val.(ProjectID)
		if !ok {
			return "", errors.New("project-id spec expects a ProjectID")
		}
		return string(id), nil
	},
	Decode: func(val any) (any, error) {
		s, ok := val.(string)
		if !ok || !strings.HasPrefix(s, "prj_") {
			return nil, errors.New("project ID must start with prj_")
		}
		return ProjectID(s), nil
	},
})
