package controller

import (
	"net/http"
	"testing"

	"asistente/test"

	"github.com/stretchr/testify/assert"
)

const (
	validChatBody   = `{"message":"hola","sender":"Sebas","session_id":"s1"}`
	emptyChatBody   = `{"message":"","sender":"Sebas"}`
	invalidChatBody = `{nope`
	noSessionBody   = `{"message":"hola","sender":"Sebas"}`
)

func TestConversationController_PostChat_InvalidJSON(t *testing.T) {
	ctrl := NewConversationController(nil, nil, nil, nil)
	req := test.NewMockRequest().WithBody(invalidChatBody)

	resp := ctrl.PostChat(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestConversationController_PostChat_EmptyMessage(t *testing.T) {
	ctrl := NewConversationController(nil, nil, nil, nil)
	req := test.NewMockRequest().WithBody(emptyChatBody)

	resp := ctrl.PostChat(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}
