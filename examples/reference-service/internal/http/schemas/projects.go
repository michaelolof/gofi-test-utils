package schemas

import "time"

// Project is the response body for a project.
type Project struct {
	ID        ProjectID `json:"id" validate:"required" spec:"project-id" example:"prj_01J9Z3"`
	Name      string    `json:"name" validate:"required" example:"Payments"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
}
