package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClickUpCreateTaskRequest_Validate_Valid(t *testing.T) {
	r := ClickUpCreateTaskRequest{ListID: "list1", Name: "task"}

	assert.NoError(t, r.Validate())
}

func TestClickUpCreateTaskRequest_Validate_MissingListID(t *testing.T) {
	r := ClickUpCreateTaskRequest{Name: "task"}

	err := r.Validate()

	assert.True(t, errors.Is(err, ErrValidation))
}

func TestClickUpCreateTaskRequest_Validate_MissingName(t *testing.T) {
	r := ClickUpCreateTaskRequest{ListID: "list1"}

	err := r.Validate()

	assert.True(t, errors.Is(err, ErrValidation))
}

func TestClickUpCreateTaskRequest_Validate_NameTooLong(t *testing.T) {
	r := ClickUpCreateTaskRequest{ListID: "list1", Name: strings.Repeat("a", 501)}

	err := r.Validate()

	assert.True(t, errors.Is(err, ErrValidation))
}

func TestClickUpCreateTaskRequest_Validate_DescTooLong(t *testing.T) {
	r := ClickUpCreateTaskRequest{ListID: "list1", Name: "task", Description: strings.Repeat("a", 10001)}

	err := r.Validate()

	assert.True(t, errors.Is(err, ErrValidation))
}

func TestClickUpUpdateStatusRequest_Validate_Valid(t *testing.T) {
	r := ClickUpUpdateStatusRequest{Status: "done"}

	assert.NoError(t, r.Validate())
}

func TestClickUpUpdateStatusRequest_Validate_Empty(t *testing.T) {
	r := ClickUpUpdateStatusRequest{}

	err := r.Validate()

	assert.True(t, errors.Is(err, ErrValidation))
}
