package domain

import "errors"

// Sentinel errors for the memory/storage layer.
var (
	ErrStoreOpen    = errors.New("failed to open store")
	ErrStoreMigrate = errors.New("failed to run migrations")
	ErrStoreSave    = errors.New("failed to save memory")
	ErrStoreSearch  = errors.New("failed to search memories")
	ErrStoreFTS     = errors.New("failed to perform full-text search")
	ErrStoreHybrid  = errors.New("hybrid search failed")
	ErrStoreDelete  = errors.New("failed to delete memory")
	ErrStoreClose   = errors.New("failed to close store")
	ErrStorePing    = errors.New("failed to ping database")
)

// Sentinel errors for the conversation/context engine.
var (
	ErrConversationLoad    = errors.New("failed to load conversation")
	ErrConversationSave    = errors.New("failed to save conversation")
	ErrConversationClear   = errors.New("failed to clear conversation")
	ErrConversationReplace = errors.New("failed to replace conversation")
	ErrCompactFailed       = errors.New("conversation compact failed")
	ErrCompactChunk        = errors.New("failed to compact chunk")
	ErrCompactMerge        = errors.New("failed to merge summaries")
)

// Sentinel errors for the finance module.
var (
	ErrFinanceParseExpense = errors.New("failed to parse expense")
	ErrFinanceWriteSheets  = errors.New("failed to write to sheets")
)

// Sentinel errors for the Claude API client.
var (
	ErrClaudeMarshal   = errors.New("failed to marshal request")
	ErrClaudeRequest   = errors.New("failed to create request")
	ErrClaudeSend      = errors.New("failed to send request")
	ErrClaudeRead      = errors.New("failed to read response")
	ErrClaudeUnmarshal = errors.New("failed to unmarshal response")
	ErrClaudeAPI       = errors.New("claude api error")
	ErrClaudeEmpty     = errors.New("claude returned empty response")
	ErrClaudeJSON      = errors.New("failed to unmarshal json response")
)

// Sentinel errors for embeddings.
var (
	ErrEmbedGenerate = errors.New("failed to generate embedding")
	ErrEmbedParse    = errors.New("failed to parse embedding vector")
)

// Sentinel errors for the skills loader.
var (
	ErrSkillsReadDir     = errors.New("failed to read skills directory")
	ErrSkillsFrontmatter = errors.New("failed to parse skill frontmatter")
)

// Sentinel errors for external integrations.
var (
	ErrSheetsCreate = errors.New("failed to create sheets service")
	ErrSheetsAppend = errors.New("failed to append row")
	ErrSheetsRead   = errors.New("failed to read range")

	ErrNotionRequest = errors.New("notion api request failed")
	ErrNotionParse   = errors.New("failed to parse notion response")

	ErrObsidianRead  = errors.New("failed to read note")
	ErrObsidianWrite = errors.New("failed to write note")

	ErrCalendarCreate = errors.New("failed to create calendar service")
	ErrCalendarList   = errors.New("failed to list events")
	ErrCalendarInsert = errors.New("failed to create event")

	ErrWhatsAppSend    = errors.New("failed to send whatsapp message")
	ErrWhatsAppAPI     = errors.New("whatsapp api error")
	ErrWhatsAppWebhook = errors.New("whatsapp webhook error")
	ErrWhatsAppVerify  = errors.New("whatsapp verification failed")

	ErrGitHubRequest = errors.New("github api request failed")
	ErrGitHubParse   = errors.New("failed to parse github response")

	ErrJiraRequest = errors.New("jira api request failed")
	ErrJiraParse   = errors.New("failed to parse jira response")

	ErrSpotifyRequest = errors.New("spotify api request failed")
	ErrSpotifyParse   = errors.New("failed to parse spotify response")

	ErrTodoistRequest = errors.New("todoist api request failed")
	ErrTodoistParse   = errors.New("failed to parse todoist response")

	ErrGmailCreate  = errors.New("failed to create gmail service")
	ErrGmailList    = errors.New("failed to list emails")
	ErrGmailGet     = errors.New("failed to get email")

	ErrClickUpRequest = errors.New("clickup api request failed")
	ErrClickUpParse   = errors.New("failed to parse clickup response")

	ErrFigmaRequest = errors.New("figma api request failed")
	ErrFigmaParse   = errors.New("failed to parse figma response")
)

// Sentinel errors for cron.
var (
	ErrCronNoClaude = errors.New("job has no claude client and no RunFn")
)

// Sentinel errors for habits.
var (
	ErrHabitLog   = errors.New("failed to log habit")
	ErrHabitQuery = errors.New("failed to query habits")
)

// Sentinel errors for links.
var (
	ErrLinkSave   = errors.New("failed to save link")
	ErrLinkSearch = errors.New("failed to search links")
)

// Sentinel errors for finance summary.
var (
	ErrFinanceSummary = errors.New("failed to generate finance summary")
)

// Sentinel errors for project status.
var (
	ErrProjectStatus = errors.New("failed to get project status")
)

// Sentinel errors for validation.
var (
	ErrValidation = errors.New("validation error")
)

// Sentinel errors for the migrator.
var (
	ErrMigrateTable  = errors.New("failed to create migrations table")
	ErrMigrateRead   = errors.New("failed to read migration file")
	ErrMigrateApply  = errors.New("failed to apply migration")
	ErrMigrateRecord = errors.New("failed to record migration")
)

// Wrap re-exports errors.Is and errors.As for convenience,
// so consumers only need to import this package.
var (
	Is  = errors.Is
	As  = errors.As
	New = errors.New
)

// Wrap wraps a sentinel error with additional context.
func Wrap(sentinel error, detail string) error {
	return &wrappedError{sentinel: sentinel, detail: detail}
}

// Wrapf wraps a sentinel error with a formatted detail and an optional cause.
func Wrapf(sentinel error, cause error) error {
	return &wrappedError{sentinel: sentinel, detail: cause.Error(), cause: cause}
}

type wrappedError struct {
	sentinel error
	detail   string
	cause    error
}

func (e *wrappedError) Error() string {
	if e.detail != "" {
		return e.sentinel.Error() + ": " + e.detail
	}
	return e.sentinel.Error()
}

func (e *wrappedError) Is(target error) bool {
	return target == e.sentinel
}

func (e *wrappedError) Unwrap() error {
	if e.cause != nil {
		return e.cause
	}
	return e.sentinel
}
