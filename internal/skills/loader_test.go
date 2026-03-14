package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	skillWithAllFields = `---
name: test-skill
description: A test skill
enabled: true
tags: [test, demo]
depends_on: [claude]
---
# Test Skill

This is the body.`

	skillDisabled = `---
name: disabled-skill
enabled: false
---
# Disabled

Should not appear in LoadEnabled.`

	skillNoFrontmatter = `# Plain Skill

No frontmatter here.`

	skillNameOnly = `---
name: minimal
---
# Minimal

Just a name.`
)

type loaderTestEnv struct {
	dir string
}

func setupLoaderTest(t *testing.T, files map[string]string) loaderTestEnv {
	t.Helper()
	dir := t.TempDir()

	for name, content := range files {
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
		require.NoError(t, err)
	}

	return loaderTestEnv{dir: dir}
}

func TestLoader_LoadAll_ParsesFrontmatter(t *testing.T) {
	env := setupLoaderTest(t, map[string]string{
		"test.md": skillWithAllFields,
	})
	loader := NewLoader(env.dir)

	skills, err := loader.LoadAll()

	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "test-skill", skills[0].Name)
	assert.Equal(t, "A test skill", skills[0].Description)
	assert.True(t, skills[0].IsEnabled())
	assert.Equal(t, []string{"test", "demo"}, skills[0].Tags)
	assert.Equal(t, []string{"claude"}, skills[0].DependsOn)
	assert.Contains(t, skills[0].Content, "# Test Skill")
	assert.Contains(t, skills[0].Content, "This is the body.")
}

func TestLoader_LoadAll_NoFrontmatter_UsesFilename(t *testing.T) {
	env := setupLoaderTest(t, map[string]string{
		"plain.md": skillNoFrontmatter,
	})
	loader := NewLoader(env.dir)

	skills, err := loader.LoadAll()

	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "plain", skills[0].Name)
	assert.Contains(t, skills[0].Content, "# Plain Skill")
}

func TestLoader_LoadAll_IgnoresNonMarkdown(t *testing.T) {
	env := setupLoaderTest(t, map[string]string{
		"skill.md":  skillNameOnly,
		"notes.txt": "not a skill",
		"data.json": `{"not": "a skill"}`,
	})
	loader := NewLoader(env.dir)

	skills, err := loader.LoadAll()

	require.NoError(t, err)
	assert.Len(t, skills, 1)
}

func TestLoader_LoadAll_InvalidDir_ReturnsError(t *testing.T) {
	loader := NewLoader("/nonexistent/path")

	_, err := loader.LoadAll()

	assert.Error(t, err)
}

func TestLoader_LoadEnabled_ExcludesDisabled(t *testing.T) {
	env := setupLoaderTest(t, map[string]string{
		"active.md":   skillWithAllFields,
		"inactive.md": skillDisabled,
	})
	loader := NewLoader(env.dir)

	skills, err := loader.LoadEnabled()

	require.NoError(t, err)
	assert.Len(t, skills, 1)
	assert.Equal(t, "test-skill", skills[0].Name)
}

func TestLoader_LoadEnabled_NilEnabledDefaultsToTrue(t *testing.T) {
	env := setupLoaderTest(t, map[string]string{
		"minimal.md": skillNameOnly,
	})
	loader := NewLoader(env.dir)

	skills, err := loader.LoadEnabled()

	require.NoError(t, err)
	assert.Len(t, skills, 1)
}

func TestFilterByTags_MatchesSingleTag(t *testing.T) {
	skills := []Skill{
		{Name: "a", Tags: []string{"finance", "sheets"}},
		{Name: "b", Tags: []string{"memory"}},
	}

	result := FilterByTags(skills, "finance")

	assert.Len(t, result, 1)
	assert.Equal(t, "a", result[0].Name)
}

func TestFilterByTags_NoTags_ReturnsAll(t *testing.T) {
	skills := []Skill{
		{Name: "a", Tags: []string{"finance"}},
		{Name: "b", Tags: []string{"memory"}},
	}

	result := FilterByTags(skills)

	assert.Len(t, result, 2)
}

func TestFilterByTags_NoMatch_ReturnsEmpty(t *testing.T) {
	skills := []Skill{
		{Name: "a", Tags: []string{"finance"}},
	}

	result := FilterByTags(skills, "nonexistent")

	assert.Empty(t, result)
}

func TestFormatForPrompt_JoinsContent(t *testing.T) {
	skills := []Skill{
		{Content: "Skill A content"},
		{Content: "Skill B content"},
	}

	result := FormatForPrompt(skills)

	assert.Contains(t, result, "Skill A content")
	assert.Contains(t, result, "Skill B content")
}

func TestParse_FullFrontmatter(t *testing.T) {
	skill, err := parse(skillWithAllFields)

	require.NoError(t, err)
	assert.Equal(t, "test-skill", skill.Name)
	assert.Equal(t, "A test skill", skill.Description)
	assert.Contains(t, skill.Content, "# Test Skill")
}

func TestParse_NoFrontmatter(t *testing.T) {
	skill, err := parse(skillNoFrontmatter)

	require.NoError(t, err)
	assert.Empty(t, skill.Name)
	assert.Contains(t, skill.Content, "# Plain Skill")
}

func TestParse_OnlyOpeningDelimiter(t *testing.T) {
	raw := "---\nname: broken"

	skill, err := parse(raw)

	require.NoError(t, err)
	assert.Empty(t, skill.Name)
	assert.Contains(t, skill.Content, "---")
}
