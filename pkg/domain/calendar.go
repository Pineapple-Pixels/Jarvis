package domain

import "time"

type CalendarEventRequest struct {
	Summary string `json:"summary"`
	Start   string `json:"start"`
	End     string `json:"end"`
}

func (r CalendarEventRequest) Validate() (start, end time.Time, err error) {
	if r.Summary == "" {
		return time.Time{}, time.Time{}, Wrap(ErrValidation, "summary is required")
	}
	if r.Start == "" {
		return time.Time{}, time.Time{}, Wrap(ErrValidation, "start is required")
	}
	if r.End == "" {
		return time.Time{}, time.Time{}, Wrap(ErrValidation, "end is required")
	}

	start, err = time.Parse(CalendarTimeFormat, r.Start)
	if err != nil {
		return time.Time{}, time.Time{}, Wrap(ErrValidation, "invalid start time format, use RFC3339")
	}

	end, err = time.Parse(CalendarTimeFormat, r.End)
	if err != nil {
		return time.Time{}, time.Time{}, Wrap(ErrValidation, "invalid end time format, use RFC3339")
	}

	if !end.After(start) {
		return time.Time{}, time.Time{}, Wrap(ErrValidation, "end must be after start")
	}

	return start, end, nil
}

type CalendarEventResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Error   string `json:"error,omitempty"`
}

type CalendarListResponse struct {
	Success bool            `json:"success"`
	Events  []CalendarEvent `json:"events"`
	Error   string          `json:"error,omitempty"`
}

type CalendarEvent struct {
	ID       string    `json:"id"`
	Summary  string    `json:"summary"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Location string    `json:"location,omitempty"`
}

const CalendarTimeFormat = time.RFC3339
