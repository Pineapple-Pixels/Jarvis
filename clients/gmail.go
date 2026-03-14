package clients

import (
	"context"
	"fmt"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

const (
	gmailLabelUnread = "UNREAD"
	gmailLabelInbox  = "INBOX"

	gmailHeaderFrom    = "From"
	gmailHeaderSubject = "Subject"
	gmailHeaderDate    = "Date"
)

// GmailEmail represents a Gmail email message.
type GmailEmail struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Snippet string `json:"snippet"`
	Date    string `json:"date"`
}

// GmailClient is the Gmail API client.
type GmailClient struct {
	service   *gmail.Service
	userEmail string
}

// NewGmailClient creates a new Gmail API client using service account credentials.
func NewGmailClient(credentialsFile, userEmail string) (*GmailClient, error) {
	ctx := context.Background()
	srv, err := gmail.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("gmail: create service: %w", err)
	}

	return &GmailClient{
		service:   srv,
		userEmail: userEmail,
	}, nil
}

// ListUnread returns unread emails up to maxResults.
func (c *GmailClient) ListUnread(maxResults int) ([]GmailEmail, error) {
	if maxResults <= 0 {
		maxResults = 10
	}

	msgs, err := c.service.Users.Messages.List(c.userEmail).
		LabelIds(gmailLabelUnread, gmailLabelInbox).
		MaxResults(int64(maxResults)).
		Do()
	if err != nil {
		return nil, fmt.Errorf("gmail: list messages: %w", err)
	}

	emails := make([]GmailEmail, 0, len(msgs.Messages))
	for _, m := range msgs.Messages {
		email, err := c.getMessage(m.Id)
		if err != nil {
			continue
		}
		emails = append(emails, email)
	}

	return emails, nil
}

// GetMessage returns a single email by ID.
func (c *GmailClient) GetMessage(messageID string) (GmailEmail, error) {
	return c.getMessage(messageID)
}

func (c *GmailClient) getMessage(messageID string) (GmailEmail, error) {
	msg, err := c.service.Users.Messages.Get(c.userEmail, messageID).
		Format("metadata").
		MetadataHeaders(gmailHeaderFrom, gmailHeaderSubject, gmailHeaderDate).
		Do()
	if err != nil {
		return GmailEmail{}, fmt.Errorf("gmail: get message: %w", err)
	}

	email := GmailEmail{
		ID:      msg.Id,
		Snippet: msg.Snippet,
	}

	for _, header := range msg.Payload.Headers {
		switch header.Name {
		case gmailHeaderFrom:
			email.From = header.Value
		case gmailHeaderSubject:
			email.Subject = header.Value
		case gmailHeaderDate:
			email.Date = header.Value
		}
	}

	return email, nil
}
