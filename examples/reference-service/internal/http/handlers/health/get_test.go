package health_test

import (
	"testing"

	"github.com/michaelolof/gofi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"example.com/projectsvc/internal/http/schemas"
	"example.com/projectsvc/internal/testutil/testharness"
)

func TestGetHealthHandler_NeedsNoToken(t *testing.T) {
	h := testharness.New(t)

	resp, err := h.Test(gofi.TestOptions{Method: "GET", Path: "/health"})
	require.NoError(t, err)
	require.Equal(t, 200, resp.Code(), resp.BodyString())

	var out schemas.SuccessBody[schemas.Health]
	require.NoError(t, resp.BindBody(&out))
	assert.Equal(t, "ok", out.Data.Status)
}
