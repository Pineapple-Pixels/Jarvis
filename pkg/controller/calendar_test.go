package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"jarvis/pkg/domain"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockCalendarClient implements calendarClient for unit tests.
type mockCalendarClient struct {
	mock.Mock
}

func (m *mockCalendarClient) GetTodayEvents() ([]domain.CalendarEvent, error) {
	args := m.Called()
	events, _ := args.Get(0).([]domain.CalendarEvent)
	return events, args.Error(1)
}

func (m *mockCalendarClient) CreateEvent(summary string, start, end time.Time) (string, error) {
	args := m.Called(summary, start, end)
	return args.String(0), args.Error(1)
}

var _ calendarClient = (*mockCalendarClient)(nil)

func newCalendarControllerWithMock(client calendarClient) *CalendarController {
	return &CalendarController{client: client}
}

const (
	validEventBody     = `{"summary":"Meeting","start":"2026-03-11T10:00:00Z","end":"2026-03-11T11:00:00Z"}`
	noSummaryEventBody = `{"summary":"","start":"2026-03-11T10:00:00Z","end":"2026-03-11T11:00:00Z"}`
	badStartEventBody  = `{"summary":"x","start":"not-a-date","end":"2026-03-11T11:00:00Z"}`
	endBeforeStartBody = `{"summary":"x","start":"2026-03-11T12:00:00Z","end":"2026-03-11T10:00:00Z"}`
	invalidEventJSON   = `{broken`
)

// --- CreateEvent: validation errors ---

func TestCalendarController_CreateEvent_InvalidJSON(t *testing.T) {
	ctrl := NewCalendarController(nil)
	req := test.NewMockRequest().WithBody(invalidEventJSON)

	resp := ctrl.CreateEvent(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestCalendarController_CreateEvent_EmptySummary(t *testing.T) {
	ctrl := NewCalendarController(nil)
	req := test.NewMockRequest().WithBody(noSummaryEventBody)

	resp := ctrl.CreateEvent(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: summary is required", errorFromBody(t, resp.Body))
}

func TestCalendarController_CreateEvent_InvalidStartFormat(t *testing.T) {
	ctrl := NewCalendarController(nil)
	req := test.NewMockRequest().WithBody(badStartEventBody)

	resp := ctrl.CreateEvent(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: invalid start time format, use RFC3339", errorFromBody(t, resp.Body))
}

func TestCalendarController_CreateEvent_EndBeforeStart(t *testing.T) {
	ctrl := NewCalendarController(nil)
	req := test.NewMockRequest().WithBody(endBeforeStartBody)

	resp := ctrl.CreateEvent(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
	assert.Equal(t, "validation error: end must be after start", errorFromBody(t, resp.Body))
}

// --- GetTodayEvents ---

func TestCalendarController_GetTodayEvents_HappyPath(t *testing.T) {
	start := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 11, 11, 0, 0, 0, time.UTC)
	events := []domain.CalendarEvent{
		{ID: "evt-1", Summary: "Stand-up", Start: start, End: end, Location: "Room A"},
		{ID: "evt-2", Summary: "Lunch", Start: start.Add(2 * time.Hour), End: end.Add(2 * time.Hour)},
	}
	client := new(mockCalendarClient)
	client.On("GetTodayEvents").Return(events, nil)
	ctrl := newCalendarControllerWithMock(client)
	req := test.NewMockRequest()

	resp := ctrl.GetTodayEvents(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	client.AssertExpectations(t)
}

func TestCalendarController_GetTodayEvents_EmptyDay(t *testing.T) {
	client := new(mockCalendarClient)
	client.On("GetTodayEvents").Return([]domain.CalendarEvent{}, nil)
	ctrl := newCalendarControllerWithMock(client)
	req := test.NewMockRequest()

	resp := ctrl.GetTodayEvents(req)

	assert.Equal(t, http.StatusOK, resp.Status)
	client.AssertExpectations(t)
}

func TestCalendarController_GetTodayEvents_ServiceError(t *testing.T) {
	client := new(mockCalendarClient)
	client.On("GetTodayEvents").Return([]domain.CalendarEvent(nil), errors.New("calendar: list events: API error"))
	ctrl := newCalendarControllerWithMock(client)
	req := test.NewMockRequest()

	resp := ctrl.GetTodayEvents(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
	client.AssertExpectations(t)
}

// --- CreateEvent: happy path + service error ---

func TestCalendarController_CreateEvent_HappyPath(t *testing.T) {
	start := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 11, 11, 0, 0, 0, time.UTC)
	client := new(mockCalendarClient)
	client.On("CreateEvent", "Meeting", start, end).Return("evt-created-123", nil)
	ctrl := newCalendarControllerWithMock(client)
	req := test.NewMockRequest().WithBody(validEventBody)

	resp := ctrl.CreateEvent(req)

	assert.Equal(t, http.StatusCreated, resp.Status)
	client.AssertExpectations(t)
}

func TestCalendarController_CreateEvent_ServiceError(t *testing.T) {
	start := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 11, 11, 0, 0, 0, time.UTC)
	client := new(mockCalendarClient)
	client.On("CreateEvent", "Meeting", start, end).Return("", errors.New("calendar: create event: quota exceeded"))
	ctrl := newCalendarControllerWithMock(client)
	req := test.NewMockRequest().WithBody(validEventBody)

	resp := ctrl.CreateEvent(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
	client.AssertExpectations(t)
}

// --- helpers ---

func errorFromBody(t *testing.T, body []byte) string {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(body, &m))
	v, _ := m["error"].(string)
	return v
}
