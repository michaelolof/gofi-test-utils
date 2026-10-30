package schemas

import "time"

// Project is the response body for a project.
type Project struct {
	ID        string    `json:"id" validate:"required" example:"prj_01J9Z3"`
	Name      string    `json:"name" validate:"required" example:"Payments"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
}
