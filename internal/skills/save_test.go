package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_Save(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader(dir)

	enabled := true
	skill := Skill{
		Name:        "test skill",
		Description: "a test",
		Tags:        []string{"test", "demo"},
		Content:     "You are a test assistant.",
		Enabled:     &enabled,
	}

	err := loader.Save(skill)

	require.NoError(t, err)

	path := filepath.Join(dir, "test-skill.md")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "name: test skill")
	assert.Contains(t, content, "enabled: true")
	assert.Contains(t, content, "You are a test assistant.")
	assert.Contains(t, content, "- test")
	assert.Contains(t, content, "- demo")
}

func TestLoader_Save_EmptyName(t *testing.T) {
	loader := NewLoader(t.TempDir())

	err := loader.Save(Skill{Content: "test"})

	assert.Error(t, err)
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"café latte", "caf-latte"},
		{"test_123", "test_123"},
		{"A B C", "a-b-c"},
		{"already-slugified", "already-slugified"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, slugify(tt.input))
		})
	}
}

func TestCachedLoader_Save_Invalidates(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader(dir)
	cached := NewCachedLoader(loader)

	enabled := true
	err := cached.Save(Skill{
		Name:    "cached-test",
		Content: "test content",
		Enabled: &enabled,
	})

	require.NoError(t, err)

	skills, err := cached.LoadEnabled()
	require.NoError(t, err)
	assert.Len(t, skills, 1)
	assert.Equal(t, "cached-test", skills[0].Name)
}
