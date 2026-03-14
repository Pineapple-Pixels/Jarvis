package controller

import (
	"net/http"
	"testing"

	"asistente/test"

	"github.com/stretchr/testify/assert"
)

const (
	validNotionBody   = `{"title":"My Page","content":"Hello"}`
	emptyTitleBody    = `{"title":"","content":"Hello"}`
	invalidNotionJSON = `{bad`
)

func TestNotionController_CreatePage_InvalidJSON(t *testing.T) {
	ctrl := NewNotionController(nil, "parent-id")
	req := test.NewMockRequest().WithBody(invalidNotionJSON)

	resp := ctrl.CreatePage(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestNotionController_CreatePage_EmptyTitle(t *testing.T) {
	ctrl := NewNotionController(nil, "parent-id")
	req := test.NewMockRequest().WithBody(emptyTitleBody)

	resp := ctrl.CreatePage(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "title is required", errorFromBody(t, resp.Body))
}

func TestNotionController_GetPage_MissingID(t *testing.T) {
	ctrl := NewNotionController(nil, "parent-id")
	req := test.NewMockRequest()

	resp := ctrl.GetPage(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "id is required", errorFromBody(t, resp.Body))
}
