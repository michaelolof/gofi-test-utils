package projects

import (
	"example.com/projectsvc/internal/domain"
	"example.com/projectsvc/internal/http/schemas"
)

func toProject(p domain.Project) schemas.Project {
	return schemas.Project{ID: p.ID, Name: p.Name, CreatedAt: p.CreatedAt}
}
