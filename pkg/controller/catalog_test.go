package controller

import (
	"net/http"
	"testing"
	"time"

	"jarvis/pkg/domain"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
)

func TestCatalogController_List_ReturnsEntries(t *testing.T) {
	now := time.Now()
	catalog := new(test.MockCatalogService)
	catalog.On("GetAll").Return([]domain.CatalogEntry{
		{Name: "expense_parse", Type: domain.CatalogTypeTool, UsageCount: 10, SuccessCount: 8, ErrorCount: 2, LastUsed: &now},
		{Name: "note_search", Type: domain.CatalogTypeSkill, UsageCount: 5, SuccessCount: 5, ErrorCount: 0},
	}, nil)
	ctrl := NewCatalogController(catalog)
	req := test.NewMockRequest()

	resp := ctrl.List(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	catalog.AssertExpectations(t)
}

func TestCatalogController_List_EmptyResult(t *testing.T) {
	catalog := new(test.MockCatalogService)
	catalog.On("GetAll").Return([]domain.CatalogEntry{}, nil)
	ctrl := NewCatalogController(catalog)
	req := test.NewMockRequest()

	resp := ctrl.List(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	catalog.AssertExpectations(t)
}

func TestCatalogController_List_ServiceError(t *testing.T) {
	catalog := new(test.MockCatalogService)
	catalog.On("GetAll").Return([]domain.CatalogEntry(nil), domain.ErrStoreOpen)
	ctrl := NewCatalogController(catalog)
	req := test.NewMockRequest()

	resp := ctrl.List(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
	catalog.AssertExpectations(t)
}

func TestCatalogController_Get_HappyPath(t *testing.T) {
	now := time.Now()
	entry := &domain.CatalogEntry{
		Name:         "expense_parse",
		Type:         domain.CatalogTypeTool,
		UsageCount:   20,
		SuccessCount: 18,
		ErrorCount:   2,
		LastUsed:     &now,
		Tags:         []string{"finance"},
	}
	catalog := new(test.MockCatalogService)
	catalog.On("GetByName", "expense_parse", "tool").Return(entry, nil)
	ctrl := NewCatalogController(catalog)
	req := test.NewMockRequest().WithParam("name", "expense_parse")

	resp := ctrl.Get(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	catalog.AssertExpectations(t)
}

func TestCatalogController_Get_DefaultsToToolType(t *testing.T) {
	catalog := new(test.MockCatalogService)
	catalog.On("GetByName", "some_tool", "tool").Return((*domain.CatalogEntry)(nil), nil)
	ctrl := NewCatalogController(catalog)
	req := test.NewMockRequest().WithParam("name", "some_tool")

	resp := ctrl.Get(req)

	assert.Equal(t, http.StatusNotFound, resp.Status)
	catalog.AssertExpectations(t)
}

func TestCatalogController_Get_CustomType(t *testing.T) {
	catalog := new(test.MockCatalogService)
	catalog.On("GetByName", "assistant", "agent").Return((*domain.CatalogEntry)(nil), nil)
	ctrl := NewCatalogController(catalog)
	req := test.NewMockRequest().WithParam("name", "assistant").WithQuery("type", "agent")

	resp := ctrl.Get(req)

	assert.Equal(t, http.StatusNotFound, resp.Status)
	catalog.AssertExpectations(t)
}

func TestCatalogController_Get_NotFound(t *testing.T) {
	catalog := new(test.MockCatalogService)
	catalog.On("GetByName", "unknown", "tool").Return((*domain.CatalogEntry)(nil), nil)
	ctrl := NewCatalogController(catalog)
	req := test.NewMockRequest().WithParam("name", "unknown")

	resp := ctrl.Get(req)

	assert.Equal(t, http.StatusNotFound, resp.Status)
	catalog.AssertExpectations(t)
}

func TestCatalogController_Get_ServiceError(t *testing.T) {
	catalog := new(test.MockCatalogService)
	catalog.On("GetByName", "broken", "tool").Return((*domain.CatalogEntry)(nil), domain.ErrStoreOpen)
	ctrl := NewCatalogController(catalog)
	req := test.NewMockRequest().WithParam("name", "broken")

	resp := ctrl.Get(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
	catalog.AssertExpectations(t)
}
