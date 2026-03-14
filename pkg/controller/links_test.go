package controller

import (
	"net/http"
	"testing"

	"asistente/test"

	"github.com/stretchr/testify/assert"
)

const (
	validLinkBody   = `{"url":"https://example.com","title":"Example"}`
	emptyURLBody    = `{"url":"","title":"Example"}`
	invalidURLBody  = `{"url":"not a url","title":"Example"}`
	ftpURLBody      = `{"url":"ftp://files.example.com","title":"Files"}`
	invalidLinkJSON = `{bad`
)

func TestLinkController_PostLink_InvalidJSON(t *testing.T) {
	ctrl := NewLinkController(nil)
	req := test.NewMockRequest().WithBody(invalidLinkJSON)

	resp := ctrl.PostLink(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestLinkController_PostLink_EmptyURL(t *testing.T) {
	ctrl := NewLinkController(nil)
	req := test.NewMockRequest().WithBody(emptyURLBody)

	resp := ctrl.PostLink(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: url is required", errorFromBody(t, resp.Body))
}

func TestLinkController_PostLink_InvalidURL(t *testing.T) {
	ctrl := NewLinkController(nil)
	req := test.NewMockRequest().WithBody(invalidURLBody)

	resp := ctrl.PostLink(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: url must be a valid http or https URL", errorFromBody(t, resp.Body))
}

func TestLinkController_PostLink_FTPScheme(t *testing.T) {
	ctrl := NewLinkController(nil)
	req := test.NewMockRequest().WithBody(ftpURLBody)

	resp := ctrl.PostLink(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: url must be a valid http or https URL", errorFromBody(t, resp.Body))
}
