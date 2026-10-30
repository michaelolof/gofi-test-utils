package apikeys

import (
	"errors"

	"github.com/michaelolof/gofi"

	"example.com/projectsvc/internal/apperr"
	"example.com/projectsvc/internal/domain"
	"example.com/projectsvc/internal/http/deps"
	"example.com/projectsvc/internal/http/middleware"
	"example.com/projectsvc/internal/http/schemas"
)

// CreateAPIKeyHandler issues a new API key for one of the caller's projects.
var CreateAPIKeyHandler = gofi.DefineHandler(gofi.RouteOptions{
	Info: gofi.Info{
		Tags:        []string{"api-keys"},
		Description: "Issues a new API key for one of the caller's projects. The secret is returned only in this response.",
	},

	Schema: &createAPIKeySchema{},

	Handler: func(c gofi.Context) error {
		s, err := gofi.ValidateAndBind[createAPIKeySchema](c)
		if err != nil {
			return apperr.Wrap(1793299514401, apperr.InvalidInput, err)
		}

		d := deps.MustGet(c)
		caller, _ := middleware.GetPrincipal(c)
		if _, err := d.Projects().Get(c.Context(), caller.UserID, s.Request.Path.ProjectID); errors.Is(err, domain.ErrNotFound) {
			return apperr.Wrap(1793299514402, apperr.NotFound, err).WithMsg("project not found")
		} else if err != nil {
			return apperr.Wrap(1793299514403, apperr.Internal, err)
		}

		key, secret, err := d.APIKeys().Create(c.Context(), s.Request.Path.ProjectID)
		if err != nil {
			return apperr.Wrap(1793299514404, apperr.Internal, err)
		}

		s.Created.Body.Data = schemas.CreatedAPIKey{ID: key.ID, Prefix: key.Prefix, Secret: secret}
		return c.Send(201, s.Created)
	},
})

type createAPIKeySchema struct {
	Request struct {
		Header schemas.AuthHeader
		Path   struct {
			ProjectID string `json:"project_id" validate:"required"`
		}
	}

	Created schemas.SuccessSchema[schemas.CreatedAPIKey]
	Err     schemas.ErrSchema
}
