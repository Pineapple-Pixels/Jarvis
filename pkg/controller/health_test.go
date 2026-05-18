package controller

import (
	"net/http"
	"testing"

	"jarvis/pkg/usecase"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthController_DetailedHealth_ReturnsHealthyReport(t *testing.T) {
	catalog := new(test.MockCatalogService)
	catalog.On("GetAll").Return(nil, nil)
	checker := usecase.NewHealthChecker([]usecase.IntegrationCheck{}, catalog)
	checker.Start(0)
	ctrl := NewHealthController(checker)
	req := test.NewMockRequest()

	resp := ctrl.DetailedHealth(req)

	require.Equal(t, http.StatusOK, resp.Status)
}

func TestHealthController_DetailedHealth_ReportsIntegrationStatus(t *testing.T) {
	okCheck := usecase.IntegrationCheck{
		Name:  "postgres",
		Check: func() error { return nil },
	}
	catalog := new(test.MockCatalogService)
	catalog.On("GetAll").Return(nil, nil)
	checker := usecase.NewHealthChecker([]usecase.IntegrationCheck{okCheck}, catalog)
	checker.Start(0)
	ctrl := NewHealthController(checker)
	req := test.NewMockRequest()

	resp := ctrl.DetailedHealth(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestHealthController_DetailedHealth_NoChecksReturnsUnknownOrHealthy(t *testing.T) {
	catalog := new(test.MockCatalogService)
	catalog.On("GetAll").Return(nil, nil)
	checker := usecase.NewHealthChecker([]usecase.IntegrationCheck{}, catalog)
	ctrl := NewHealthController(checker)
	req := test.NewMockRequest()

	resp := ctrl.DetailedHealth(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}
