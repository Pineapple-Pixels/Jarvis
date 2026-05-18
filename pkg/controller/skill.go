package controller

import (
	"log"
	"net/http"
	"regexp"

	"jarvis/internal/skills"
	"jarvis/pkg/domain"
	"jarvis/web"
)

// skillNamePattern restricts skill names to safe identifiers only.
var skillNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// SkillController handles skill management endpoints.
type SkillController struct {
	writer skills.SkillWriter
}

// NewSkillController creates a new SkillController.
func NewSkillController(writer skills.SkillWriter) *SkillController {
	return &SkillController{writer: writer}
}

// ListSkills returns all enabled skills.
func (c *SkillController) ListSkills(req web.Request) web.Response {
	loaded, err := c.writer.LoadEnabled()
	if err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.SkillListResponse{Error: err.Error()})
	}

	infos := make([]domain.SkillInfo, len(loaded))
	for i, s := range loaded {
		infos[i] = domain.SkillInfo{
			Name:        s.Name,
			Description: s.Description,
			Tags:        s.Tags,
			Enabled:     s.IsEnabled(),
		}
	}

	return web.NewJSONResponse(http.StatusOK, domain.SkillListResponse{Success: true, Skills: infos})
}

// CreateSkill saves a new skill to disk.
// NOTE: This endpoint should be restricted to admin users in production.
func (c *SkillController) CreateSkill(req web.Request) web.Response {
	var payload domain.SkillCreateRequest
	if err := web.DecodeJSON(req.Body(), &payload); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.SkillResponse{Error: "invalid body"})
	}

	if err := payload.Validate(); err != nil {
		return web.NewJSONResponse(http.StatusBadRequest, domain.SkillResponse{Error: err.Error()})
	}

	if !skillNamePattern.MatchString(payload.Name) {
		return web.NewJSONResponse(http.StatusBadRequest, domain.SkillResponse{Error: "name must match ^[a-zA-Z0-9_-]+$"})
	}

	log.Printf("skill: creating skill %q via API", payload.Name)

	enabled := true
	skill := skills.Skill{
		Name:        payload.Name,
		Description: payload.Description,
		Tags:        payload.Tags,
		Content:     payload.Content,
		Enabled:     &enabled,
	}

	if err := c.writer.Save(skill); err != nil {
		return web.NewJSONResponse(http.StatusInternalServerError, domain.SkillResponse{Error: err.Error()})
	}

	return web.NewJSONResponse(http.StatusCreated, domain.SkillResponse{
		Success: true,
		Message: "Skill '" + payload.Name + "' created",
	})
}
