package usecase

import (
	"time"

	"jarvis/pkg/domain"
)

// CalendarProvider is the interface the usecase layer requires from a calendar integration.
// Only the methods actually called by usecase code are listed.
type CalendarProvider interface {
	GetTodayEvents() ([]domain.CalendarEvent, error)
	CreateEvent(summary string, start, end time.Time) (string, error)
}

// GmailProvider is the interface the usecase layer requires from an email integration.
type GmailProvider interface {
	ListUnread(maxResults int) ([]domain.GmailEmail, error)
}

// TodoistProvider is the interface the usecase layer requires from a task integration.
type TodoistProvider interface {
	GetTasks() ([]domain.TodoistTask, error)
	CreateTask(content string, dueDate *string) (domain.TodoistTask, error)
}

// GitHubProvider is the interface the usecase layer requires from a GitHub integration.
type GitHubProvider interface {
	ListIssues(owner, repo string) ([]domain.GitHubIssue, error)
	CreateIssue(owner, repo, title, body string) (domain.GitHubIssue, error)
}

// JiraProvider is the interface the usecase layer requires from a Jira integration.
type JiraProvider interface {
	GetMyIssues() ([]domain.JiraIssue, error)
}

// SpotifyProvider is the interface the usecase layer requires from a Spotify integration.
type SpotifyProvider interface {
	GetCurrentlyPlaying() (*domain.SpotifyTrack, error)
	Next() error
}
