package controller

import (
	"net/http"
	"testing"

	"asistente/pkg/usecase"
	"asistente/test"

	"github.com/stretchr/testify/assert"
)

const (
	validExpenseBody = `{"message":"gaste 5000 en el super","sender":"Sebas"}`
	noMessageBody    = `{"message":"","sender":"Sebas"}`
	noSenderBody     = `{"message":"gaste 5000 en el super"}`
	invalidJSONBody  = `{invalid`

	mockExpenseJSON = `{"amount":5000,"category":"Supermercado","description":"super","paid_by":"","date":"2026-03-10"}`
)

func TestFinanceController_PostExpense_InvalidJSON(t *testing.T) {
	ctrl := NewFinanceController(nil)
	req := test.NewMockRequest().WithBody(invalidJSONBody)

	resp := ctrl.PostExpense(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFinanceController_PostExpense_EmptyMessage(t *testing.T) {
	ctrl := NewFinanceController(nil)
	req := test.NewMockRequest().WithBody(noMessageBody)

	resp := ctrl.PostExpense(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestFinanceController_PostExpense_DefaultSender(t *testing.T) {
	srv, ai := test.NewMockClaudeServer(test.ClaudeResponse{Text: mockExpenseJSON})
	defer srv.Close()
	uc := usecase.NewFinanceUseCase(ai, nil)
	ctrl := NewFinanceController(uc)
	req := test.NewMockRequest().WithBody(noSenderBody)

	resp := ctrl.PostExpense(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
}
