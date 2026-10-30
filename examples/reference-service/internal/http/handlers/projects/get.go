package projects

import (
	"errors"

	"github.com/michaelolof/gofi"

	"example.com/projectsvc/internal/apperr"
	"example.com/projectsvc/internal/domain"
	"example.com/projectsvc/internal/http/deps"
	"example.com/projectsvc/internal/http/schemas"
)

// GetProjectHandler returns one of the caller's projects by ID.
var GetProjectHandler = gofi.DefineHandler(gofi.RouteOptions{
	Info: gofi.Info{
		Tags:        []string{"projects"},
		Description: "Returns one of the caller's projects. A project owned by someone else answers 404.",
	},

	Schema: &getProjectSchema{},

	Handler: func(c gofi.Context) error {
		s, err := gofi.ValidateAndBind[getProjectSchema](c)
		if err != nil {
			return apperr.Wrap(1793299514311, apperr.InvalidInput, err)
		}

		caller, _ := domain.PrincipalFrom(c.Context())
		project, err := deps.MustGet(c).Projects().Get(c.Context(), caller.UserID, s.Request.Path.ProjectID)
		if errors.Is(err, domain.ErrNotFound) {
			return apperr.Wrap(1793299514312, apperr.NotFound, err).WithMsg("project not found")
		}
		if err != nil {
			return apperr.Wrap(1793299514313, apperr.Internal, err)
		}

		s.Ok.Body.Data = toProject(project)
		return c.Send(200, s.Ok)
	},
})

type getProjectSchema struct {
	Request struct {
		Header schemas.AuthHeader
		Path   struct {
			ProjectID string `json:"project_id" validate:"required" example:"prj_01J9Z3"`
		}
	}

	Ok  schemas.SuccessSchema[schemas.Project]
	Err schemas.ErrSchema
}
