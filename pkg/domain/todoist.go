package domain

import "time"

// TodoistTask represents a Todoist task.
type TodoistTask struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	DueDate   string `json:"due_date,omitempty"`
	Priority  int    `json:"priority"`
	Completed bool   `json:"completed"`
	URL       string `json:"url"`
}

// TodoistCreateTaskRequest is the payload for creating a Todoist task.
type TodoistCreateTaskRequest struct {
	Content string  `json:"content"`
	DueDate *string `json:"due_date,omitempty"`
}

func (r TodoistCreateTaskRequest) Validate() error {
	if r.Content == "" {
		return Wrap(ErrValidation, "content is required")
	}
	if r.DueDate != nil && *r.DueDate != "" {
		if _, err := time.Parse(DateFormatISO, *r.DueDate); err != nil {
			return Wrap(ErrValidation, "due_date must be in YYYY-MM-DD format")
		}
	}
	return nil
}

// TodoistTaskListResponse is the response for listing tasks.
type TodoistTaskListResponse struct {
	Success bool          `json:"success"`
	Tasks   []TodoistTask `json:"tasks,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// TodoistTaskResponse is the response for a single task operation.
type TodoistTaskResponse struct {
	Success bool         `json:"success"`
	Task    *TodoistTask `json:"task,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// TodoistActionResponse is the response for task actions (complete, etc).
type TodoistActionResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
