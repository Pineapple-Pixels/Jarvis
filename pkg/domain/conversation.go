package domain

const (
	MaxHistoryMessages   = 50
	CompactThreshold     = 40
	ChunkSize            = 10
	MinMessagesToCompact = 5

	CompactSummaryPrefix = "Resumen de conversacion anterior: "

	DefaultSystemPrompt = "Sos un asistente personal inteligente. Hablás en español rioplatense.\n"
	SkillsSectionHeader = "Tus capacidades:\n\n"
)

const (
	maxMessageLen   = 10_000
	maxSenderLen    = 200
	maxSessionIDLen = 256
)

type ChatRequest struct {
	Message   string `json:"message"`
	Sender    string `json:"sender"`
	SessionID string `json:"session_id"`
}

func (r ChatRequest) Validate() error {
	if r.Message == "" {
		return Wrap(ErrValidation, "message is required")
	}
	if len(r.Message) > maxMessageLen {
		return Wrap(ErrValidation, "message exceeds maximum length")
	}
	if len(r.Sender) > maxSenderLen {
		return Wrap(ErrValidation, "sender exceeds maximum length")
	}
	if len(r.SessionID) > maxSessionIDLen {
		return Wrap(ErrValidation, "session_id exceeds maximum length")
	}
	return nil
}

type ChatResponse struct {
	Success  bool   `json:"success"`
	Response string `json:"response,omitempty"`
	Error    string `json:"error,omitempty"`
}
