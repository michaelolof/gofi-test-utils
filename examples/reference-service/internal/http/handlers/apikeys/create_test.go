package apikeys_test

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

func TestCreateAPIKeyHandler_Returns201WithSecret(t *testing.T) {
	h := testharness.New(t)
	project, err := h.Deps.ProjectStore.Create(context.Background(), domain.Project{Name: "Payments", OwnerID: "alice"})
	require.NoError(t, err)

	resp, err := h.Test(gofi.TestOptions{
		Method:  "POST",
		Path:    "/api/v1/projects/:project_id/api-keys",
		Paths:   map[string]string{"project_id": project.ID},
		Headers: testharness.AuthHeaders("token-alice"),
	})
	require.NoError(t, err)
	require.Equal(t, 201, resp.Code(), resp.BodyString())

	var out schemas.SuccessBody[schemas.CreatedAPIKey]
	require.NoError(t, resp.BindBody(&out))
	assert.NotEmpty(t, out.Data.Secret)
}
