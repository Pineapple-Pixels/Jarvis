package clients

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ObsidianVault struct {
	basePath string
}

type ObsidianNoteRef struct {
	Path       string    `json:"path"`
	Title      string    `json:"title"`
	ModifiedAt time.Time `json:"modified_at"`
}

func NewObsidianVault(basePath string) *ObsidianVault {
	return &ObsidianVault{basePath: basePath}
}

func (v *ObsidianVault) ReadNote(path string) (string, error) {
	content, err := os.ReadFile(v.fullPath(path))
	if err != nil {
		return "", fmt.Errorf("obsidian: read note: %w", err)
	}
	return string(content), nil
}

func (v *ObsidianVault) WriteNote(path, content string) error {
	full := v.fullPath(path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return fmt.Errorf("obsidian: create dir: %w", err)
	}
	return os.WriteFile(full, []byte(content), 0o644)
}

func (v *ObsidianVault) ListNotes(dir string) ([]ObsidianNoteRef, error) {
	searchDir := v.basePath
	if dir != "" {
		searchDir = v.fullPath(dir)
	}

	var notes []ObsidianNoteRef
	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		rel, _ := filepath.Rel(v.basePath, path)
		notes = append(notes, ObsidianNoteRef{
			Path:       rel,
			Title:      strings.TrimSuffix(info.Name(), ".md"),
			ModifiedAt: info.ModTime(),
		})
		return nil
	})

	return notes, err
}

func (v *ObsidianVault) SearchNotes(query string) ([]ObsidianNoteRef, error) {
	queryLower := strings.ToLower(query)
	var results []ObsidianNoteRef

	err := filepath.Walk(v.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		if strings.Contains(strings.ToLower(string(content)), queryLower) {
			rel, _ := filepath.Rel(v.basePath, path)
			results = append(results, ObsidianNoteRef{
				Path:       rel,
				Title:      strings.TrimSuffix(info.Name(), ".md"),
				ModifiedAt: info.ModTime(),
			})
		}
		return nil
	})

	return results, err
}

func (v *ObsidianVault) fullPath(path string) string {
	return filepath.Join(v.basePath, path)
}
