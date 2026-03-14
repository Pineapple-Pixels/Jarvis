package domain

const (
	PathParamName = "name"
)

type ProjectStatusRequest struct {
	Name string `json:"name"`
}

type ProjectStatusResponse struct {
	Success   bool   `json:"success"`
	Name      string `json:"name,omitempty"`
	Summary   string `json:"summary,omitempty"`
	NoteCount int    `json:"note_count,omitempty"`
	Error     string `json:"error,omitempty"`
}
