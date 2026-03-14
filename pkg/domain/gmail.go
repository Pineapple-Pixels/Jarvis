package domain

// Gmail query parameter constants.
const (
	QueryParamMaxResults = "max_results"
)

// GmailEmail represents an email message.
type GmailEmail struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Snippet string `json:"snippet"`
	Date    string `json:"date"`
}

// GmailListResponse is the response for listing emails.
type GmailListResponse struct {
	Success bool         `json:"success"`
	Emails  []GmailEmail `json:"emails,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// GmailMessageResponse is the response for a single email.
type GmailMessageResponse struct {
	Success bool        `json:"success"`
	Email   *GmailEmail `json:"email,omitempty"`
	Error   string      `json:"error,omitempty"`
}
