package controller

import (
	"net/http"
	"testing"

	"asistente/test"

	"github.com/stretchr/testify/assert"
)

const (
	validNoteWriteBody    = `{"path":"notes/test.md","content":"hello world"}`
	pathTraversalBody     = `{"path":"../../../etc/passwd","content":"x"}`
	absolutePathBody      = `{"path":"/etc/passwd","content":"x"}`
	emptyPathWriteBody    = `{"path":"","content":"hello"}`
	emptyContentWriteBody = `{"path":"test.md","content":""}`
	invalidObsidianJSON   = `{nope`
)

func TestObsidianController_WriteNote_InvalidJSON(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest().WithBody(invalidObsidianJSON)

	resp := ctrl.WriteNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestObsidianController_WriteNote_PathTraversal(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest().WithBody(pathTraversalBody)

	resp := ctrl.WriteNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: path must not contain '..'", errorFromBody(t, resp.Body))
}

func TestObsidianController_WriteNote_AbsolutePath(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest().WithBody(absolutePathBody)

	resp := ctrl.WriteNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: path must be relative", errorFromBody(t, resp.Body))
}

func TestObsidianController_WriteNote_EmptyPath(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest().WithBody(emptyPathWriteBody)

	resp := ctrl.WriteNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: path is required", errorFromBody(t, resp.Body))
}

func TestObsidianController_WriteNote_EmptyContent(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest().WithBody(emptyContentWriteBody)

	resp := ctrl.WriteNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: content is required", errorFromBody(t, resp.Body))
}

func TestObsidianController_ReadNote_MissingPath(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest()

	resp := ctrl.ReadNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestObsidianController_ReadNote_PathTraversal(t *testing.T) {
	ctrl := NewObsidianController(nil)
	req := test.NewMockRequest().WithQuery("path", "../secret.md")

	resp := ctrl.ReadNote(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: path must not contain '..'", errorFromBody(t, resp.Body))
}
