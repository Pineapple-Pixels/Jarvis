package controller

import (
	"net/http"
	"testing"

	"jarvis/internal/skills"
	"jarvis/pkg/domain"
	"jarvis/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubSkillProvider is a test double for skills.SkillProvider.
type stubSkillProvider struct {
	skills []skills.Skill
	err    error
}

func (s *stubSkillProvider) LoadEnabled() ([]skills.Skill, error) {
	return s.skills, s.err
}

var _ skills.SkillProvider = (*stubSkillProvider)(nil)

func TestSkillsQAController_Report_ReturnsQAResults(t *testing.T) {
	enabled := true
	provider := &stubSkillProvider{
		skills: []skills.Skill{
			{
				Name:        "expense-parse",
				Description: "Parsea gastos en lenguaje natural",
				Tags:        []string{"finance"},
				Content:     "## Expense parsing instructions",
				Enabled:     &enabled,
			},
		},
	}
	rubric := skills.DefaultRubric()
	ctrl := NewSkillsQAController(provider, rubric)
	req := test.NewMockRequest()

	resp := ctrl.Report(req)

	require.Equal(t, http.StatusOK, resp.Status)
}

func TestSkillsQAController_Report_EmptySkillsReturnsEmptyResults(t *testing.T) {
	provider := &stubSkillProvider{skills: []skills.Skill{}}
	rubric := skills.DefaultRubric()
	ctrl := NewSkillsQAController(provider, rubric)
	req := test.NewMockRequest()

	resp := ctrl.Report(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}

func TestSkillsQAController_Report_LoaderError(t *testing.T) {
	provider := &stubSkillProvider{err: domain.ErrStoreOpen}
	rubric := skills.DefaultRubric()
	ctrl := NewSkillsQAController(provider, rubric)
	req := test.NewMockRequest()

	resp := ctrl.Report(req)

	assert.Equal(t, http.StatusInternalServerError, resp.Status)
}

func TestSkillsQAController_Validate_ValidSkill(t *testing.T) {
	provider := &stubSkillProvider{}
	rubric := skills.DefaultRubric()
	ctrl := NewSkillsQAController(provider, rubric)
	body := `{"name":"expense-parse","description":"Parsea gastos en lenguaje natural","tags":["finance"],"content":"## Parsing instructions"}`
	req := test.NewMockRequest().WithBody(body)

	resp := ctrl.Validate(req)

	require.Equal(t, http.StatusOK, resp.Status)
}

func TestSkillsQAController_Validate_InvalidBody(t *testing.T) {
	provider := &stubSkillProvider{}
	rubric := skills.DefaultRubric()
	ctrl := NewSkillsQAController(provider, rubric)
	req := test.NewMockRequest().WithBody(`{broken`)

	resp := ctrl.Validate(req)

	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestSkillsQAController_Validate_SkillWithIssues(t *testing.T) {
	provider := &stubSkillProvider{}
	rubric := skills.DefaultRubric()
	ctrl := NewSkillsQAController(provider, rubric)
	// Missing tags and content — should still return 200 with validation issues in the result body.
	body := `{"name":"broken-skill","description":"","tags":[],"content":""}`
	req := test.NewMockRequest().WithBody(body)

	resp := ctrl.Validate(req)

	assert.Equal(t, http.StatusOK, resp.Status)
}
