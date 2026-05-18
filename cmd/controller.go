package main

import (
	"net/http"
	"strings"
	"time"

	"jarvis/config"
	"jarvis/internal/agents"
	"jarvis/internal/hooks"
	"jarvis/internal/profiles"
	"jarvis/internal/rules"
	"jarvis/internal/skills"
	"jarvis/pkg/controller"
	"jarvis/pkg/domain"
	"jarvis/pkg/service"
	"jarvis/pkg/usecase"
	"jarvis/web"
)

type Controllers struct {
	Finance  *controller.FinanceController
	Memory   *controller.MemoryController
	Chat     *controller.ConversationController
	WhatsApp *controller.WhatsAppController
	Notion   *controller.NotionController
	Obsidian *controller.ObsidianController
	Calendar *controller.CalendarController
	GitHub   *controller.GitHubController
	Jira     *controller.JiraController
	Spotify  *controller.SpotifyController
	Todoist  *controller.TodoistController
	Gmail    *controller.GmailController
	ClickUp  *controller.ClickUpController
	Habit    *controller.HabitController
	Link     *controller.LinkController
	Project  *controller.ProjectController
	Figma    *controller.FigmaController
	Telegram *controller.TelegramController
	Skill    *controller.SkillController
	Trigger  *controller.TriggerController
	Usage    *controller.UsageController
	Catalog  *controller.CatalogController
	SkillsQA *controller.SkillsQAController
	Health   *controller.HealthController
	Pairing  web.Handler

	healthChecker *usecase.HealthChecker
}

// Close stops background goroutines owned by the Controllers layer.
func (c *Controllers) Close() {
	if c.healthChecker != nil {
		c.healthChecker.Stop()
	}
}

// buildAvailableIntegrations returns a single map of integration availability
// derived from nil-checks on the Clients struct. This map is the single source
// of truth consumed by ToolRegistry, DependencyChecker, and HealthChecker,
// eliminating the previous three separate nil-check blocks (Issue A3).
func buildAvailableIntegrations(cl Clients) map[string]bool {
	return map[string]bool{
		"calendar": cl.Calendar != nil,
		"sheets":   cl.Sheets != nil,
		"github":   cl.GitHub != nil,
		"jira":     cl.Jira != nil,
		"spotify":  cl.Spotify != nil,
		"todoist":  cl.Todoist != nil,
		"gmail":    cl.Gmail != nil,
		"notion":   cl.Notion != nil,
		"obsidian": cl.Obsidian != nil,
		"clickup":  cl.ClickUp != nil,
		"figma":    cl.Figma != nil,
		"whatsapp": cl.WhatsApp != nil,
		"telegram": cl.Telegram != nil,
	}
}

// buildHealthChecker constructs a HealthChecker from the available integrations
// map and starts the periodic check loop (Issue A1).
func buildHealthChecker(available map[string]bool, catalogSvc service.CatalogService) *usecase.HealthChecker {
	var checks []usecase.IntegrationCheck
	// Only integrations surfaced in health (excludes sheets — internal-only).
	healthIntegrations := []string{
		"calendar", "github", "jira", "spotify", "todoist",
		"gmail", "notion", "obsidian", "clickup", "figma",
		"whatsapp", "telegram",
	}
	for _, name := range healthIntegrations {
		if available[name] {
			name := name
			checks = append(checks, usecase.IntegrationCheck{
				Name:  name,
				Check: func() error { return nil },
			})
		}
	}
	checker := usecase.NewHealthChecker(checks, catalogSvc)
	checker.Start(5 * time.Minute)
	return checker
}

// buildMessageRouter assembles the ToolRegistry, AgentOrchestrator, and
// MessageRouter. Returns nil when neither WhatsApp nor Telegram is configured
// (Issue A1).
func buildMessageRouter(
	cfg config.Config,
	cl Clients,
	available map[string]bool,
	financeUC *usecase.FinanceUseCase,
	memorySvc service.MemoryService,
	embedder service.Embedder,
	chatUC *usecase.ConversationUseCase,
	catalogSvc service.CatalogService,
	skillsLoader skills.SkillProvider,
	hooksRegistry *hooks.Registry,
	usageTracker *usecase.UsageTracker,
) *usecase.MessageRouter {
	if cl.WhatsApp == nil && cl.Telegram == nil {
		return nil
	}

	// Reminder manager sends reminders via the first available channel.
	var reminderMgr *usecase.ReminderManager
	if cl.WhatsApp != nil && cfg.WhatsApp.To != "" {
		reminderMgr = usecase.NewReminderManager(func(text string) {
			_ = cl.WhatsApp.SendTextMessage(cfg.WhatsApp.To, text)
		})
	}

	var sw skills.SkillWriter
	if writer, ok := skillsLoader.(skills.SkillWriter); ok {
		sw = writer
	}

	toolReg := usecase.BuildToolRegistry(
		financeUC, memorySvc, embedder,
		cl.Calendar, cl.Gmail, cl.Todoist, cl.GitHub, cl.Jira,
		cl.Spotify, cl.Notion, cl.Obsidian, sw, reminderMgr,
	)
	toolReg.SetCatalog(catalogSvc)
	toolReg.SetHooks(hooksRegistry)
	if cfg.Runtime.DryRunTools != "" {
		toolReg.SetDryRunTools(strings.Split(cfg.Runtime.DryRunTools, ","))
	}

	var agent *usecase.AgentUseCase
	var orchestrator *usecase.AgentOrchestrator
	if tp, ok := cl.AI.(domain.ToolUseProvider); ok {
		agent = usecase.NewAgentUseCase(tp, toolReg)
		if loaded, err := skillsLoader.LoadEnabled(); err == nil {
			agent.SetSkills(loaded)
		}
		agentDefs := usecase.DefaultAgents()
		agentsLoader := agents.NewLoader(cfg.Runtime.AgentsDir)
		if loaded, err := agentsLoader.LoadAll(); err == nil && len(loaded) > 0 {
			agentDefs = loaded
		}
		orchestrator = usecase.NewAgentOrchestrator(tp, toolReg, agentDefs)
	}

	var transcriber domain.Transcriber
	if cl.Transcriber != nil {
		transcriber = cl.Transcriber
	}

	rulesLoader := rules.NewLoader(cfg.Runtime.RulesDir)
	router := usecase.NewMessageRouter(
		chatUC, cl.AI, agent, orchestrator, transcriber,
		skillsLoader, rulesLoader, hooksRegistry, usageTracker, cfg.WhatsApp.To,
	)

	profileLoader := profiles.NewLoader(cfg.Runtime.ProfilesDir)
	if profile, err := profileLoader.Load(cfg.Runtime.DefaultProfile); err == nil {
		router.SetProfile(profile)
	}

	// Re-use the availability map — no redundant nil-checks here.
	router.SetDependencyChecker(skills.NewDependencyChecker(available))

	return router
}

// NewControllers is a thin orchestrator: it wires use-cases, calls the two
// helper builders, and populates the Controllers struct (Issue A1).
// catalogSvc is injected rather than derived via type assertion (Issue C3).
func NewControllers(
	cl Clients,
	cfg config.Config,
	memorySvc service.MemoryService,
	financeSvc service.FinanceService,
	embedder service.Embedder,
	skillsLoader skills.SkillProvider,
	hooksRegistry *hooks.Registry,
	scheduler *usecase.Scheduler,
	catalogSvc service.CatalogService,
) Controllers {
	financeUC := usecase.NewFinanceUseCase(cl.AI, financeSvc)
	financeUC.SetMemoryService(memorySvc)

	chatUC := usecase.NewConversationUseCase(memorySvc, cl.AI, hooksRegistry, cfg.AI.MaxHistoryMsgs, cfg.AI.CompactThreshold)

	usageTracker := usecase.NewUsageTracker()

	// Single source of truth for integration availability (Issue A3).
	available := buildAvailableIntegrations(cl)

	c := Controllers{
		Finance: controller.NewFinanceController(financeUC),
		Memory:  controller.NewMemoryController(memorySvc, embedder),
		Chat:    controller.NewConversationController(chatUC, cl.AI, skillsLoader, hooksRegistry),
		Habit:   controller.NewHabitController(usecase.NewHabitUseCase(memorySvc)),
		Link:    controller.NewLinkController(usecase.NewLinkUseCase(memorySvc, embedder)),
		Project: controller.NewProjectController(usecase.NewProjectUseCase(memorySvc, embedder, cl.AI)),
		Trigger: controller.NewTriggerController(scheduler),
		Usage:   controller.NewUsageController(usageTracker),
		Catalog: controller.NewCatalogController(catalogSvc),
	}

	if cl.Notion != nil {
		c.Notion = controller.NewNotionController(cl.Notion, cfg.Integrations.NotionPageID)
	}
	if cl.Obsidian != nil {
		c.Obsidian = controller.NewObsidianController(cl.Obsidian)
	}
	if available["calendar"] {
		c.Calendar = controller.NewCalendarController(cl.Calendar)
	}
	if available["github"] {
		c.GitHub = controller.NewGitHubController(cl.GitHub)
	}
	if cl.Jira != nil {
		c.Jira = controller.NewJiraController(cl.Jira)
	}
	if cl.Spotify != nil {
		c.Spotify = controller.NewSpotifyController(cl.Spotify)
	}
	if cl.Todoist != nil {
		c.Todoist = controller.NewTodoistController(cl.Todoist)
	}
	if cl.Gmail != nil {
		c.Gmail = controller.NewGmailController(cl.Gmail)
	}
	if cl.ClickUp != nil {
		c.ClickUp = controller.NewClickUpController(cl.ClickUp)
	}
	if available["figma"] {
		c.Figma = controller.NewFigmaController(cl.Figma)
	}

	if sw, ok := skillsLoader.(skills.SkillWriter); ok {
		c.Skill = controller.NewSkillController(sw)
	}

	rubric, _ := skills.LoadRubric("skills/qa-rubric.yaml")
	c.SkillsQA = controller.NewSkillsQAController(skillsLoader, rubric)

	checker := buildHealthChecker(available, catalogSvc)
	c.healthChecker = checker
	c.Health = controller.NewHealthController(checker)

	router := buildMessageRouter(
		cfg, cl, available,
		financeUC, memorySvc, embedder,
		chatUC, catalogSvc, skillsLoader, hooksRegistry, usageTracker,
	)

	if router != nil {
		c.Pairing = func(req web.Request) web.Response {
			return web.NewJSONResponse(http.StatusOK, map[string]any{
				"success": true, "pairing_code": router.GetPairingCode(),
			})
		}
	}

	if available["whatsapp"] && cfg.WhatsApp.VerifyToken != "" && router != nil {
		c.WhatsApp = controller.NewWhatsAppController(router, cl.WhatsApp, cfg.WhatsApp.VerifyToken, cfg.WhatsApp.AppSecret)
	}

	if available["telegram"] && router != nil {
		c.Telegram = controller.NewTelegramController(router, cl.Telegram, cfg.Telegram.SecretToken, cfg.Telegram.BotUsername)
	}

	return c
}
