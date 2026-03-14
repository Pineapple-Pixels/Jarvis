package controller

import (
	"net/http"
	"testing"

	"asistente/test"

	"github.com/stretchr/testify/assert"
)

const (
	validTaskBody   = `{"content":"Buy milk"}`
	emptyTaskBody   = `{"content":""}`
	taskWithDate    = `{"content":"Buy milk","due_date":"2026-03-15"}`
	taskBadDate     = `{"content":"Buy milk","due_date":"15/03/2026"}`
	invalidTaskJSON = `{nah`
)

func TestTodoistController_CreateTask_InvalidJSON(t *testing.T) {
	ctrl := NewTodoistController(nil)
	req := test.NewMockRequest().WithBody(invalidTaskJSON)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestTodoistController_CreateTask_EmptyContent(t *testing.T) {
	ctrl := NewTodoistController(nil)
	req := test.NewMockRequest().WithBody(emptyTaskBody)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: content is required", errorFromBody(t, resp.Body))
}

func TestTodoistController_CreateTask_InvalidDateFormat(t *testing.T) {
	ctrl := NewTodoistController(nil)
	req := test.NewMockRequest().WithBody(taskBadDate)

	resp := ctrl.CreateTask(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: due_date must be in YYYY-MM-DD format", errorFromBody(t, resp.Body))
}
