package domain

// ClickUpTask represents a ClickUp task.
type ClickUpTask struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	Assignees []string `json:"assignees"`
	URL       string   `json:"url"`
	DueDate   string   `json:"due_date,omitempty"`
}

// ClickUpCreateTaskRequest is the payload for creating a ClickUp task.
type ClickUpCreateTaskRequest struct {
	ListID      string `json:"list_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Validate checks that required fields are present and within limits.
func (r ClickUpCreateTaskRequest) Validate() error {
	if r.ListID == "" {
		return Wrap(ErrValidation, "list_id is required")
	}
	if r.Name == "" {
		return Wrap(ErrValidation, "name is required")
	}
	if len(r.Name) > 500 {
		return Wrap(ErrValidation, "name exceeds 500 characters")
	}
	if len(r.Description) > 10000 {
		return Wrap(ErrValidation, "description exceeds 10000 characters")
	}
	return nil
}

// ClickUpUpdateStatusRequest is the payload for updating a task's status.
type ClickUpUpdateStatusRequest struct {
	Status string `json:"status"`
}

// Validate checks that the status field is present.
func (r ClickUpUpdateStatusRequest) Validate() error {
	if r.Status == "" {
		return Wrap(ErrValidation, "status is required")
	}
	return nil
}

// ClickUpTaskListResponse is the response for listing tasks.
type ClickUpTaskListResponse struct {
	Success bool          `json:"success"`
	Tasks   []ClickUpTask `json:"tasks,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// ClickUpTaskResponse is the response for a single task operation.
type ClickUpTaskResponse struct {
	Success bool         `json:"success"`
	Task    *ClickUpTask `json:"task,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// ClickUpActionResponse is the response for task status updates.
type ClickUpActionResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
