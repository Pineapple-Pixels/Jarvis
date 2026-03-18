package controller

import (
	"net/http"
	"testing"

	"asistente/test"

	"github.com/stretchr/testify/assert"
)

func TestFigmaController_GetFile_MissingFileKey(t *testing.T) {
	ctrl := NewFigmaController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetFile(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFigmaController_GetNodes_MissingIDs(t *testing.T) {
	ctrl := NewFigmaController(nil)
	req := test.NewMockRequest().WithParam("file_key", "abc123")

	resp := ctrl.GetNodes(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFigmaController_GetImages_MissingFileKey(t *testing.T) {
	ctrl := NewFigmaController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetImages(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFigmaController_GetImages_MissingIDs(t *testing.T) {
	ctrl := NewFigmaController(nil)
	req := test.NewMockRequest().WithParam("file_key", "abc123")

	resp := ctrl.GetImages(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFigmaController_GetComments_MissingFileKey(t *testing.T) {
	ctrl := NewFigmaController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetComments(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFigmaController_GetProjectFiles_MissingProjectID(t *testing.T) {
	ctrl := NewFigmaController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetProjectFiles(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFigmaController_GetComponents_MissingFileKey(t *testing.T) {
	ctrl := NewFigmaController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetComponents(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}
