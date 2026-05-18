package controller

import (
	"net/http"
	"time"

	"jarvis/pkg/domain"
	"jarvis/web"
)

type calendarClient interface {
	GetTodayEvents() ([]domain.CalendarEvent, error)
	CreateEvent(summary string, start, end time.Time) (string, error)
}

type CalendarController struct {
	client calendarClient
}

func NewCalendarController(client calendarClient) *CalendarController {
	return &CalendarController{client: client}
}

func (c *CalendarController) GetTodayEvents(req web.Request) web.Response {
	events, err := c.client.GetTodayEvents()
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.CalendarListResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusOK, domain.CalendarListResponse{
		Success: true, Events: events,
	})
}

func (c *CalendarController) CreateEvent(req web.Request) web.Response {
	var payload domain.CalendarEventRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.CalendarEventResponse{Error: "invalid body"})
	}

	start, end, err := payload.Validate()
	if err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.CalendarEventResponse{Error: err.Error()})
	}

	id, err := c.client.CreateEvent(payload.Summary, start, end)
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.CalendarEventResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusCreated, domain.CalendarEventResponse{Success: true, ID: id})
}

