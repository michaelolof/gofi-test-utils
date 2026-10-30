package health

import (
	"github.com/michaelolof/gofi"

	"example.com/projectsvc/internal/http/schemas"
)

// GetHealthHandler reports that the service is up.
var GetHealthHandler = gofi.DefineHandler(gofi.RouteOptions{
	Info: gofi.Info{
		Tags:        []string{"health"},
		Description: "Reports that the service is up. Needs no credentials.",
		Security:    gofi.NoSecurity,
	},

	Schema: &getHealthSchema{},

	Handler: func(c gofi.Context) error {
		s, err := gofi.ValidateAndBind[getHealthSchema](c)
		if err != nil {
			return err
		}

		s.Ok.Body.Data = schemas.Health{Status: "ok"}
		return c.Send(200, s.Ok)
	},
})

type getHealthSchema struct {
	Ok  schemas.SuccessSchema[schemas.Health]
	Err schemas.ErrSchema
}
