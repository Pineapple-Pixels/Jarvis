package domain

import (
	"strings"
	"testing"

	stderrors "errors"

	"github.com/stretchr/testify/assert"
)

const (
	errPathRequired       = "validation error: path is required"
	errPathRelative       = "validation error: path must be relative"
	errPathTraversal      = "validation error: path must not contain '..'"
	errPathMaxLen         = "validation error: path exceeds maximum length"
	errContentRequired    = "validation error: content is required"
	errContentMaxLen      = "validation error: content exceeds maximum length"
	errSummaryRequired    = "validation error: summary is required"
	errStartRequired      = "validation error: start is required"
	errEndRequired        = "validation error: end is required"
	errStartFormat        = "validation error: invalid start time format, use RFC3339"
	errEndFormat          = "validation error: invalid end time format, use RFC3339"
	errEndAfterStart      = "validation error: end must be after start"
	errURLRequired        = "validation error: url is required"
	errURLMaxLen          = "validation error: url exceeds maximum length"
	errURLFormat          = "validation error: url must be a valid http or https URL"
	errTodoistContent     = "validation error: content is required"
	errTodoistDateFormat  = "validation error: due_date must be in YYYY-MM-DD format"
	errMessageRequired    = "validation error: message is required"
	errMessageMaxLen      = "validation error: message exceeds maximum length"
	errSenderMaxLen       = "validation error: sender exceeds maximum length"
	errSessionIDMaxLen    = "validation error: session_id exceeds maximum length"
	errNoteContentMax     = "validation error: content exceeds maximum length"
	errTooManyTags        = "validation error: too many tags"
	errInvalidTag         = "validation error: invalid tag"
)

func TestObsidianNoteRequest_Validate_Success(t *testing.T) {
	r := ObsidianNoteRequest{Path: "notes/test.md", Content: "hello"}

	assert.NoError(t, r.Validate())
}

func TestObsidianNoteRequest_Validate_EmptyPath(t *testing.T) {
	r := ObsidianNoteRequest{Path: "", Content: "hello"}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errPathRequired, err.Error())
}

func TestObsidianNoteRequest_Validate_EmptyContent(t *testing.T) {
	r := ObsidianNoteRequest{Path: "test.md", Content: ""}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errContentRequired, err.Error())
}

func TestObsidianNoteRequest_Validate_PathTraversal(t *testing.T) {
	r := ObsidianNoteRequest{Path: "../../../etc/passwd", Content: "x"}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errPathTraversal, err.Error())
}

func TestObsidianNoteRequest_Validate_AbsolutePath(t *testing.T) {
	r := ObsidianNoteRequest{Path: "/etc/passwd", Content: "x"}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errPathRelative, err.Error())
}

func TestObsidianNoteRequest_Validate_PathTooLong(t *testing.T) {
	r := ObsidianNoteRequest{Path: strings.Repeat("a", 1025), Content: "x"}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errPathMaxLen, err.Error())
}

func TestObsidianNoteRequest_Validate_ContentTooLong(t *testing.T) {
	r := ObsidianNoteRequest{Path: "test.md", Content: strings.Repeat("a", 100_001)}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errContentMaxLen, err.Error())
}

func TestValidatePath_Success(t *testing.T) {
	assert.NoError(t, ValidatePath("notes/daily.md"))
}

func TestValidatePath_Empty(t *testing.T) {
	err := ValidatePath("")

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errPathRequired, err.Error())
}

func TestValidatePath_Traversal(t *testing.T) {
	err := ValidatePath("../secret.md")

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errPathTraversal, err.Error())
}

func TestValidatePath_Absolute(t *testing.T) {
	err := ValidatePath("/etc/passwd")

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errPathRelative, err.Error())
}

func TestCalendarEventRequest_Validate_Success(t *testing.T) {
	r := CalendarEventRequest{Summary: "Meeting", Start: "2026-03-11T10:00:00Z", End: "2026-03-11T11:00:00Z"}

	start, end, err := r.Validate()

	assert.NoError(t, err)
	assert.Equal(t, 10, start.Hour())
	assert.Equal(t, 11, end.Hour())
}

func TestCalendarEventRequest_Validate_EmptySummary(t *testing.T) {
	r := CalendarEventRequest{Summary: "", Start: "2026-03-11T10:00:00Z", End: "2026-03-11T11:00:00Z"}

	_, _, err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errSummaryRequired, err.Error())
}

func TestCalendarEventRequest_Validate_EmptyStart(t *testing.T) {
	r := CalendarEventRequest{Summary: "x", Start: "", End: "2026-03-11T11:00:00Z"}

	_, _, err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errStartRequired, err.Error())
}

func TestCalendarEventRequest_Validate_EmptyEnd(t *testing.T) {
	r := CalendarEventRequest{Summary: "x", Start: "2026-03-11T10:00:00Z", End: ""}

	_, _, err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errEndRequired, err.Error())
}

func TestCalendarEventRequest_Validate_InvalidStartFormat(t *testing.T) {
	r := CalendarEventRequest{Summary: "x", Start: "not-a-date", End: "2026-03-11T11:00:00Z"}

	_, _, err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errStartFormat, err.Error())
}

func TestCalendarEventRequest_Validate_InvalidEndFormat(t *testing.T) {
	r := CalendarEventRequest{Summary: "x", Start: "2026-03-11T10:00:00Z", End: "bad"}

	_, _, err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errEndFormat, err.Error())
}

func TestCalendarEventRequest_Validate_EndBeforeStart(t *testing.T) {
	r := CalendarEventRequest{Summary: "x", Start: "2026-03-11T12:00:00Z", End: "2026-03-11T10:00:00Z"}

	_, _, err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errEndAfterStart, err.Error())
}

func TestCalendarEventRequest_Validate_EndEqualsStart(t *testing.T) {
	r := CalendarEventRequest{Summary: "x", Start: "2026-03-11T10:00:00Z", End: "2026-03-11T10:00:00Z"}

	_, _, err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errEndAfterStart, err.Error())
}

func TestLinkSaveRequest_Validate_Success(t *testing.T) {
	r := LinkSaveRequest{URL: "https://example.com/page"}

	assert.NoError(t, r.Validate())
}

func TestLinkSaveRequest_Validate_Empty(t *testing.T) {
	r := LinkSaveRequest{URL: ""}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errURLRequired, err.Error())
}

func TestLinkSaveRequest_Validate_InvalidScheme(t *testing.T) {
	r := LinkSaveRequest{URL: "ftp://files.example.com"}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errURLFormat, err.Error())
}

func TestLinkSaveRequest_Validate_NoHost(t *testing.T) {
	r := LinkSaveRequest{URL: "https://"}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errURLFormat, err.Error())
}

func TestLinkSaveRequest_Validate_NotAURL(t *testing.T) {
	r := LinkSaveRequest{URL: "just some text"}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errURLFormat, err.Error())
}

func TestLinkSaveRequest_Validate_TooLong(t *testing.T) {
	r := LinkSaveRequest{URL: "https://example.com/" + strings.Repeat("a", 2030)}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errURLMaxLen, err.Error())
}

func TestTodoistCreateTaskRequest_Validate_Success(t *testing.T) {
	r := TodoistCreateTaskRequest{Content: "Buy milk"}

	assert.NoError(t, r.Validate())
}

func TestTodoistCreateTaskRequest_Validate_WithValidDate(t *testing.T) {
	date := "2026-03-15"
	r := TodoistCreateTaskRequest{Content: "Buy milk", DueDate: &date}

	assert.NoError(t, r.Validate())
}

func TestTodoistCreateTaskRequest_Validate_EmptyContent(t *testing.T) {
	r := TodoistCreateTaskRequest{Content: ""}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errTodoistContent, err.Error())
}

func TestTodoistCreateTaskRequest_Validate_InvalidDateFormat(t *testing.T) {
	date := "15/03/2026"
	r := TodoistCreateTaskRequest{Content: "test", DueDate: &date}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errTodoistDateFormat, err.Error())
}

func TestTodoistCreateTaskRequest_Validate_NilDate(t *testing.T) {
	r := TodoistCreateTaskRequest{Content: "test", DueDate: nil}

	assert.NoError(t, r.Validate())
}

func TestChatRequest_Validate_Success(t *testing.T) {
	r := ChatRequest{Message: "hola", Sender: "Sebas"}

	assert.NoError(t, r.Validate())
}

func TestChatRequest_Validate_EmptyMessage(t *testing.T) {
	r := ChatRequest{Message: "", Sender: "Sebas"}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errMessageRequired, err.Error())
}

func TestChatRequest_Validate_MessageTooLong(t *testing.T) {
	r := ChatRequest{Message: strings.Repeat("a", 10_001), Sender: "Sebas"}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errMessageMaxLen, err.Error())
}

func TestChatRequest_Validate_SenderTooLong(t *testing.T) {
	r := ChatRequest{Message: "hola", Sender: strings.Repeat("x", 201)}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errSenderMaxLen, err.Error())
}

func TestChatRequest_Validate_SessionIDTooLong(t *testing.T) {
	r := ChatRequest{Message: "hola", Sender: "Sebas", SessionID: strings.Repeat("x", 257)}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errSessionIDMaxLen, err.Error())
}

func TestNoteRequest_Validate_Success(t *testing.T) {
	r := NoteRequest{Content: "some note", Tags: []string{"tag1"}}

	assert.NoError(t, r.Validate())
}

func TestNoteRequest_Validate_EmptyContent(t *testing.T) {
	r := NoteRequest{Content: ""}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errContentRequired, err.Error())
}

func TestNoteRequest_Validate_ContentTooLong(t *testing.T) {
	r := NoteRequest{Content: strings.Repeat("a", 50_001)}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errNoteContentMax, err.Error())
}

func TestNoteRequest_Validate_TooManyTags(t *testing.T) {
	tags := make([]string, 21)
	for i := range tags {
		tags[i] = "tag"
	}
	r := NoteRequest{Content: "note", Tags: tags}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errTooManyTags, err.Error())
}

func TestNoteRequest_Validate_EmptyTag(t *testing.T) {
	r := NoteRequest{Content: "note", Tags: []string{"valid", ""}}

	err := r.Validate()

	assert.True(t, stderrors.Is(err, ErrValidation))
	assert.Equal(t, errInvalidTag, err.Error())
}

func TestNoteRequest_Validate_NoTags(t *testing.T) {
	r := NoteRequest{Content: "note"}

	assert.NoError(t, r.Validate())
}
