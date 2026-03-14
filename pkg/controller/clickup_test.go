package controller

import (
	"net/http"
	"strings"
	"testing"

	"asistente/test"

	"github.com/stretchr/testify/assert"
)

func TestClickUpController_CreateTask_InvalidJSON(t *testing.T) {
	ctrl := NewClickUpController(nil)
	req := test.NewMockRequest().WithBody(`{bad`)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestClickUpController_CreateTask_MissingListID(t *testing.T) {
	ctrl := NewClickUpController(nil)
	req := test.NewMockRequest().WithBody(`{"name":"Task 1"}`)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: list_id is required", errorFromBody(t, resp.Body))
}

func TestClickUpController_CreateTask_MissingName(t *testing.T) {
	ctrl := NewClickUpController(nil)
	req := test.NewMockRequest().WithBody(`{"list_id":"123"}`)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: name is required", errorFromBody(t, resp.Body))
}

func TestClickUpController_CreateTask_NameTooLong(t *testing.T) {
	ctrl := NewClickUpController(nil)
	long := strings.Repeat("a", 501)
	req := test.NewMockRequest().WithBody(`{"list_id":"123","name":"` + long + `"}`)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: name exceeds 500 characters", errorFromBody(t, resp.Body))
}

func TestClickUpController_GetTask_MissingID(t *testing.T) {
	ctrl := NewClickUpController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "id is required", errorFromBody(t, resp.Body))
}

func TestClickUpController_UpdateTaskStatus_MissingID(t *testing.T) {
	ctrl := NewClickUpController(nil)
	req := test.NewMockRequest()

	resp := ctrl.UpdateTaskStatus(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "id is required", errorFromBody(t, resp.Body))
}

func TestClickUpController_UpdateTaskStatus_InvalidJSON(t *testing.T) {
	ctrl := NewClickUpController(nil)
	req := test.NewMockRequest().WithParam("id", "abc123").WithBody(`{bad`)

	resp := ctrl.UpdateTaskStatus(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestClickUpController_UpdateTaskStatus_EmptyStatus(t *testing.T) {
	ctrl := NewClickUpController(nil)
	req := test.NewMockRequest().WithParam("id", "abc123").WithBody(`{"status":""}`)

	resp := ctrl.UpdateTaskStatus(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: status is required", errorFromBody(t, resp.Body))
}
