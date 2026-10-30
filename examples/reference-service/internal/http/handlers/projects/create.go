package projects

import (
	"github.com/michaelolof/gofi"

	"example.com/projectsvc/internal/apperr"
	"example.com/projectsvc/internal/domain"
	"example.com/projectsvc/internal/http/deps"
	"example.com/projectsvc/internal/http/middleware"
	"example.com/projectsvc/internal/http/schemas"
)

// CreateProjectHandler creates a project owned by the caller.
var CreateProjectHandler = gofi.DefineHandler(gofi.RouteOptions{
	Info: gofi.Info{
		Tags:        []string{"projects"},
		Description: "Creates a project owned by the authenticated caller.",
	},

	Schema: &createProjectSchema{},

	Handler: func(c gofi.Context) error {
		s, err := gofi.ValidateAndBind[createProjectSchema](c)
		if err != nil {
			return apperr.Wrap(1793299514301, apperr.InvalidInput, err)
		}

		caller, _ := middleware.GetPrincipal(c)
		project, err := deps.MustGet(c).Projects().Create(c.Context(), domain.Project{
			Name:    s.Request.Body.Name,
			OwnerID: caller.UserID,
		})
		if err != nil {
			return apperr.Wrap(1793299514302, apperr.Internal, err)
		}

		s.Created.Body.Data = toProject(project)
		return c.Send(201, s.Created)
	},
})

type createProjectSchema struct {
	Request struct {
		Header schemas.AuthHeader
		Body   struct {
			Name string `json:"name" validate:"required,min=2,max=80" example:"Payments"`
		}
	}

	Created schemas.SuccessSchema[schemas.Project]
	Err     schemas.ErrSchema
}
