package clients

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type CalendarClient struct {
	service    *calendar.Service
	calendarID string
}

type CalendarEvent struct {
	ID       string    `json:"id"`
	Summary  string    `json:"summary"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Location string    `json:"location,omitempty"`
}

func NewCalendarClient(credentialsFile, calendarID string) (*CalendarClient, error) {
	ctx := context.Background()
	srv, err := calendar.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("calendar: create service: %w", err)
	}

	return &CalendarClient{
		service:    srv,
		calendarID: calendarID,
	}, nil
}

func (c *CalendarClient) GetTodayEvents() ([]CalendarEvent, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	return c.ListEvents(startOfDay, endOfDay)
}

func (c *CalendarClient) ListEvents(timeMin, timeMax time.Time) ([]CalendarEvent, error) {
	events, err := c.service.Events.List(c.calendarID).
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, fmt.Errorf("calendar: list events: %w", err)
	}

	result := make([]CalendarEvent, 0, len(events.Items))
	for _, item := range events.Items {
		e := CalendarEvent{
			ID:       item.Id,
			Summary:  item.Summary,
			Location: item.Location,
		}

		if item.Start.DateTime != "" {
			e.Start, _ = time.Parse(time.RFC3339, item.Start.DateTime)
		}
		if item.End.DateTime != "" {
			e.End, _ = time.Parse(time.RFC3339, item.End.DateTime)
		}

		result = append(result, e)
	}

	return result, nil
}

func (c *CalendarClient) CreateEvent(summary string, start, end time.Time) (string, error) {
	event := &calendar.Event{
		Summary: summary,
		Start: &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: end.Format(time.RFC3339),
		},
	}

	created, err := c.service.Events.Insert(c.calendarID, event).Do()
	if err != nil {
		return "", fmt.Errorf("calendar: create event: %w", err)
	}

	return created.Id, nil
}
