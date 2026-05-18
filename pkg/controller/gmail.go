package controller

import (
	"net/http"
	"strconv"

	"jarvis/pkg/domain"
	"jarvis/web"
)

const defaultMaxResults = 10

type gmailClient interface {
	ListUnread(maxResults int) ([]domain.GmailEmail, error)
	GetMessage(messageID string) (domain.GmailEmail, error)
}

// GmailController handles Gmail API endpoints.
type GmailController struct {
	client gmailClient
}

func NewGmailController(client gmailClient) *GmailController {
	return &GmailController{client: client}
}

// ListUnread returns unread emails.
func (c *GmailController) ListUnread(req web.Request) web.Response {
	maxResults := defaultMaxResults
	if v, ok := req.Query(domain.QueryParamMaxResults); ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxResults = n
		}
	}

	emails, err := c.client.ListUnread(maxResults)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.GmailListResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.GmailListResponse{
		Success: true, Emails: emails,
	})
}

// GetMessage returns a single email by ID.
func (c *GmailController) GetMessage(req web.Request) web.Response {
	messageID, ok := req.Param(domain.PathParamID)
	if !ok || messageID == "" {
		return web.NewJSONResponse(http.StatusBadRequest, domain.GmailMessageResponse{Error: "id is required"})
	}

	email, err := c.client.GetMessage(messageID)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.GmailMessageResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.GmailMessageResponse{
		Success: true,
		Email:   &email,
	})
}
