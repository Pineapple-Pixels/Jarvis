package domain

import "net/url"

const (
	LinkTag = "link"

	linkMaxURLLen = 2048
)

type LinkSaveRequest struct {
	URL   string   `json:"url"`
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
}

func (r LinkSaveRequest) Validate() error {
	if r.URL == "" {
		return Wrap(ErrValidation, "url is required")
	}
	if len(r.URL) > linkMaxURLLen {
		return Wrap(ErrValidation, "url exceeds maximum length")
	}
	u, err := url.Parse(r.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Wrap(ErrValidation, "url must be a valid http or https URL")
	}
	return nil
}

type LinkResponse struct {
	Success bool   `json:"success"`
	ID      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type LinkListResponse struct {
	Success bool     `json:"success"`
	Results []Memory `json:"results"`
	Error   string   `json:"error,omitempty"`
}
