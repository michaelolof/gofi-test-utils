package projects_test

import (
	"context"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"example.com/projectsvc/internal/domain"
	"example.com/projectsvc/internal/http/schemas"
	"example.com/projectsvc/internal/testutil/testharness"
)

func TestGetProjectHandler_ReturnsCallersProject(t *testing.T) {
	h := testharness.New(t)
	project, err := h.Deps.ProjectStore.Create(context.Background(), domain.Project{Name: "Payments", OwnerID: "alice"})
	require.NoError(t, err)

	resp, err := h.Test(gofi.TestOptions{
		Method:  "GET",
		Path:    "/api/v1/projects/:project_id",
		Paths:   map[string]string{"project_id": project.ID},
		Headers: testharness.AuthHeaders("token-alice"),
	})
	require.NoError(t, err)
	require.Equal(t, 200, resp.Code(), resp.BodyString())

	var out schemas.SuccessBody[schemas.Project]
	require.NoError(t, resp.BindBody(&out))
	assert.Equal(t, project.ID, out.Data.ID)
}

func TestGetProjectHandler_OtherOwner_Returns404(t *testing.T) {
	h := testharness.New(t)
	project, err := h.Deps.ProjectStore.Create(context.Background(), domain.Project{Name: "Payments", OwnerID: "alice"})
	require.NoError(t, err)

	resp, err := h.Test(gofi.TestOptions{
		Method:  "GET",
		Path:    "/api/v1/projects/:project_id",
		Paths:   map[string]string{"project_id": project.ID},
		Headers: testharness.AuthHeaders("token-bob"),
	})
	require.NoError(t, err)
	require.Equal(t, 404, resp.Code(), resp.BodyString())

	var out schemas.ErrBody
	require.NoError(t, resp.BindBody(&out))
	assert.Equal(t, "project not found", out.Message)
	assert.Equal(t, int64(1793299514312), out.ErrorID)
}
