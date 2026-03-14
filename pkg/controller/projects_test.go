package controller

import (
	"net/http"
	"testing"

	"asistente/test"

	"github.com/stretchr/testify/assert"
)

func TestProjectController_GetStatus_MissingName(t *testing.T) {
	ctrl := NewProjectController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetStatus(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "project name is required", errorFromBody(t, resp.Body))
}
