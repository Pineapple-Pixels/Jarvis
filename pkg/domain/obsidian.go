package domain

import (
	"path/filepath"
	"strings"
	"time"
)

const (
	QueryParamPath = "path"
	QueryParamDir  = "dir"

	obsidianMaxContentLen = 100_000
	obsidianMaxPathLen    = 1024
)

type ObsidianNoteRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (r ObsidianNoteRequest) Validate() error {
	if r.Path == "" {
		return Wrap(ErrValidation, "path is required")
	}
	if len(r.Path) > obsidianMaxPathLen {
		return Wrap(ErrValidation, "path exceeds maximum length")
	}
	if filepath.IsAbs(r.Path) || strings.HasPrefix(r.Path, "/") || strings.HasPrefix(r.Path, "\\") {
		return Wrap(ErrValidation, "path must be relative")
	}
	if strings.Contains(r.Path, "..") {
		return Wrap(ErrValidation, "path must not contain '..'")
	}
	if r.Content == "" {
		return Wrap(ErrValidation, "content is required")
	}
	if len(r.Content) > obsidianMaxContentLen {
		return Wrap(ErrValidation, "content exceeds maximum length")
	}
	return nil
}

// ValidatePath validates a path query parameter for read operations.
func ValidatePath(path string) error {
	if path == "" {
		return Wrap(ErrValidation, "path is required")
	}
	if filepath.IsAbs(path) || strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		return Wrap(ErrValidation, "path must be relative")
	}
	if strings.Contains(path, "..") {
		return Wrap(ErrValidation, "path must not contain '..'")
	}
	return nil
}

type ObsidianNoteResponse struct {
	Success bool   `json:"success"`
	Path    string `json:"path,omitempty"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ObsidianListResponse struct {
	Success bool            `json:"success"`
	Notes   []ObsidianNote  `json:"notes"`
	Error   string          `json:"error,omitempty"`
}

type ObsidianNote struct {
	Path       string    `json:"path"`
	Title      string    `json:"title"`
	ModifiedAt time.Time `json:"modified_at"`
}
