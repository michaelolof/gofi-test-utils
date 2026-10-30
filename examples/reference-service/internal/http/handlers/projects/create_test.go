package projects_test

import (
	"strings"
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"example.com/projectsvc/internal/http/schemas"
	"example.com/projectsvc/internal/testutil/testharness"
)

func TestCreateProjectHandler_Returns201WithProject(t *testing.T) {
	h := testharness.New(t)

	resp, err := h.Test(gofi.TestOptions{
		Method:  "POST",
		Path:    "/api/v1/projects",
		Headers: testharness.AuthHeaders("token-alice"),
		Body:    strings.NewReader(`{"name":"Payments"}`),
	})
	require.NoError(t, err)
	require.Equal(t, 201, resp.Code(), resp.BodyString())

	var out schemas.SuccessBody[schemas.Project]
	require.NoError(t, resp.BindBody(&out))
	assert.Equal(t, "success", out.Status)
	assert.Equal(t, "Payments", out.Data.Name)
	assert.NotEmpty(t, out.Data.ID)
}

func TestCreateProjectHandler_InvalidName_Returns400(t *testing.T) {
	h := testharness.New(t)

	resp, err := h.Test(gofi.TestOptions{
		Method:  "POST",
		Path:    "/api/v1/projects",
		Headers: testharness.AuthHeaders("token-alice"),
		Body:    strings.NewReader(`{"name":"P"}`),
	})
	require.NoError(t, err)
	require.Equal(t, 400, resp.Code(), resp.BodyString())

	var out schemas.ErrBody
	require.NoError(t, resp.BindBody(&out))
	assert.Equal(t, int64(1793299514301), out.ErrorID)
	assert.Contains(t, out.Message, "name")
}

func TestCreateProjectHandler_NoToken_Returns401(t *testing.T) {
	h := testharness.New(t)

	resp, err := h.Test(gofi.TestOptions{
		Method: "POST",
		Path:   "/api/v1/projects",
		Body:   strings.NewReader(`{"name":"Payments"}`),
	})
	require.NoError(t, err)
	assert.Equal(t, 401, resp.Code(), resp.BodyString())
}
