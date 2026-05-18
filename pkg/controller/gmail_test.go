package controller

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"jarvis/pkg/domain"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockGmailClient implements gmailClient for unit tests.
type mockGmailClient struct {
	mock.Mock
}

func (m *mockGmailClient) ListUnread(maxResults int) ([]domain.GmailEmail, error) {
	args := m.Called(maxResults)
	emails, _ := args.Get(0).([]domain.GmailEmail)
	return emails, args.Error(1)
}

func (m *mockGmailClient) GetMessage(messageID string) (domain.GmailEmail, error) {
	args := m.Called(messageID)
	return args.Get(0).(domain.GmailEmail), args.Error(1)
}

var _ gmailClient = (*mockGmailClient)(nil)

func newGmailControllerWithMock(client gmailClient) *GmailController {
	return &GmailController{client: client}
}

// --- ListUnread ---

func TestGmailController_ListUnread_HappyPath(t *testing.T) {
	now := time.Now().Format(time.RFC1123Z)
	emails := []domain.GmailEmail{
		{ID: "msg-1", From: "alice@example.com", Subject: "Hello", Snippet: "Hi there", Date: now},
		{ID: "msg-2", From: "bob@example.com", Subject: "World", Snippet: "Bye", Date: now},
	}
	client := new(mockGmailClient)
	client.On("ListUnread", defaultMaxResults).Return(emails, nil)
	ctrl := newGmailControllerWithMock(client)
	req := test.NewMockRequest()

	resp := ctrl.ListUnread(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	client.AssertExpectations(t)
}

func TestGmailController_ListUnread_EmptyResult(t *testing.T) {
	client := new(mockGmailClient)
	client.On("ListUnread", defaultMaxResults).Return([]domain.GmailEmail{}, nil)
	ctrl := newGmailControllerWithMock(client)
	req := test.NewMockRequest()

	resp := ctrl.ListUnread(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	client.AssertExpectations(t)
}

func TestGmailController_ListUnread_ServiceError(t *testing.T) {
	client := new(mockGmailClient)
	client.On("ListUnread", defaultMaxResults).Return([]domain.GmailEmail(nil), errors.New("gmail: list messages: API error"))
	ctrl := newGmailControllerWithMock(client)
	req := test.NewMockRequest()

	resp := ctrl.ListUnread(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
	client.AssertExpectations(t)
}

func TestGmailController_ListUnread_CustomMaxResults(t *testing.T) {
	client := new(mockGmailClient)
	client.On("ListUnread", 5).Return([]domain.GmailEmail{}, nil)
	ctrl := newGmailControllerWithMock(client)
	req := test.NewMockRequest().WithQuery("max_results", "5")

	resp := ctrl.ListUnread(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	client.AssertExpectations(t)
}

func TestGmailController_ListUnread_InvalidMaxResults_UsesDefault(t *testing.T) {
	client := new(mockGmailClient)
	client.On("ListUnread", defaultMaxResults).Return([]domain.GmailEmail{}, nil)
	ctrl := newGmailControllerWithMock(client)
	req := test.NewMockRequest().WithQuery("max_results", "not-a-number")

	resp := ctrl.ListUnread(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	client.AssertExpectations(t)
}

// --- GetMessage ---

func TestGmailController_GetMessage_MissingID(t *testing.T) {
	ctrl := NewGmailController(nil)
	req := test.NewMockRequest()

	resp := ctrl.GetMessage(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestGmailController_GetMessage_EmptyID(t *testing.T) {
	ctrl := NewGmailController(nil)
	req := test.NewMockRequest().WithParam("id", "")

	resp := ctrl.GetMessage(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestGmailController_GetMessage_HappyPath(t *testing.T) {
	email := domain.GmailEmail{
		ID:      "msg-abc",
		From:    "sender@example.com",
		Subject: "Test subject",
		Snippet: "Short preview",
		Date:    "Mon, 01 Jan 2026 10:00:00 +0000",
	}
	client := new(mockGmailClient)
	client.On("GetMessage", "msg-abc").Return(email, nil)
	ctrl := newGmailControllerWithMock(client)
	req := test.NewMockRequest().WithParam("id", "msg-abc")

	resp := ctrl.GetMessage(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	client.AssertExpectations(t)
}

func TestGmailController_GetMessage_ServiceError(t *testing.T) {
	client := new(mockGmailClient)
	client.On("GetMessage", "msg-missing").Return(domain.GmailEmail{}, errors.New("gmail: get message: not found"))
	ctrl := newGmailControllerWithMock(client)
	req := test.NewMockRequest().WithParam("id", "msg-missing")

	resp := ctrl.GetMessage(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
	client.AssertExpectations(t)
}
