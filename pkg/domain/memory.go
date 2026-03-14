package domain

import "time"

const (
	DefaultSearchLimit = 5

	DefaultVecWeight = 0.6
	DefaultFTSWeight = 0.4

	DecayLambda = 0.05

	SearchModeFTS    = "fts"
	SearchModeVector = "vector"
	SearchModeHybrid = "hybrid"

	QueryParamQ     = "q"
	QueryParamLimit = "limit"
	QueryParamMode  = "mode"
	PathParamID     = "id"
)

type Memory struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	Score     float64   `json:"score,omitempty"`
}

type ConversationMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	maxNoteContentLen = 50_000
	maxTagLen         = 100
	maxTagCount       = 20
)

type NoteRequest struct {
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

func (r NoteRequest) Validate() error {
	if r.Content == "" {
		return Wrap(ErrValidation, "content is required")
	}
	if len(r.Content) > maxNoteContentLen {
		return Wrap(ErrValidation, "content exceeds maximum length")
	}
	if len(r.Tags) > maxTagCount {
		return Wrap(ErrValidation, "too many tags")
	}
	for _, tag := range r.Tags {
		if tag == "" || len(tag) > maxTagLen {
			return Wrap(ErrValidation, "invalid tag")
		}
	}
	return nil
}

type NoteResponse struct {
	Success bool   `json:"success"`
	ID      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type SearchResponse struct {
	Results []Memory `json:"results"`
}
