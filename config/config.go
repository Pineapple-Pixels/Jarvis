package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// ServerConfig holds HTTP server and security settings.
type ServerConfig struct {
	Port          string
	WebhookSecret string
	AllowInsecure bool
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	PostgresDSN string
}

// AIConfig holds AI provider configuration.
type AIConfig struct {
	Provider        string
	ClaudeAPIKey    string
	ClaudeModel     string
	ClaudeModelLight string
	OpenAIAPIKey    string
	OpenAIModel     string
	OpenAIModelLight string
	CompactThreshold int
	MaxHistoryMsgs  int
}

// WhatsAppConfig holds WhatsApp Business Cloud API settings.
type WhatsAppConfig struct {
	PhoneID     string
	Token       string
	To          string
	VerifyToken string
	AppSecret   string
}

// TelegramConfig holds Telegram bot settings.
type TelegramConfig struct {
	BotToken    string
	SecretToken string
	BotUsername string
}

// GoogleConfig holds Google service credentials and IDs.
type GoogleConfig struct {
	SheetsID       string
	SheetsCredFile string
	SheetsSheetName string
	CalendarID     string
}

// IntegrationsConfig holds third-party integration tokens.
type IntegrationsConfig struct {
	NotionAPIKey      string
	NotionPageID      string
	ObsidianVaultPath string
	GitHubToken       string
	JiraBaseURL       string
	JiraEmail         string
	JiraAPIToken      string
	SpotifyAccessToken string
	TodoistAPIToken   string
	GmailUserEmail    string
	ClickUpAPIToken   string
	ClickUpTeamID     string
	FigmaAccessToken  string
}

// RuntimeConfig holds runtime behavior settings.
type RuntimeConfig struct {
	AgentsDir       string
	RulesDir        string
	ProfilesDir     string
	DefaultProfile  string
	HooksConfigFile string
	SkillsDir       string
	DryRunTools     string
}

// Config is the root configuration loaded at startup.
type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	AI           AIConfig
	WhatsApp     WhatsAppConfig
	Telegram     TelegramConfig
	Google       GoogleConfig
	Integrations IntegrationsConfig
	Runtime      RuntimeConfig
}

// Validate checks that required fields are present and consistent.
// It returns an error describing the first (or all) violations found.
func (c Config) Validate() error {
	var errs []error

	port, err := strconv.Atoi(c.Server.Port)
	if err != nil || port <= 0 {
		errs = append(errs, fmt.Errorf("SERVER_PORT must be a positive integer, got %q", c.Server.Port))
	}

	if c.Database.PostgresDSN == "" {
		errs = append(errs, errors.New("POSTGRES_DSN is required"))
	}

	if c.AI.ClaudeAPIKey == "" && c.AI.OpenAIAPIKey == "" {
		errs = append(errs, errors.New("at least one AI API key is required (CLAUDE_API_KEY or OPENAI_API_KEY)"))
	}

	return errors.Join(errs...)
}

func Load() Config {
	return Config{
		Server: ServerConfig{
			Port:          envOr("PORT", "8080"),
			WebhookSecret: os.Getenv("WEBHOOK_SECRET"),
			AllowInsecure: os.Getenv("ALLOW_INSECURE") == "true",
		},
		Database: DatabaseConfig{
			PostgresDSN: os.Getenv("POSTGRES_DSN"),
		},
		AI: AIConfig{
			Provider:         envOr("AI_PROVIDER", "claude"),
			ClaudeAPIKey:     os.Getenv("CLAUDE_API_KEY"),
			ClaudeModel:      envOr("CLAUDE_MODEL", "claude-sonnet-4-6"),
			ClaudeModelLight: envOr("CLAUDE_MODEL_LIGHT", ""),
			OpenAIAPIKey:     os.Getenv("OPENAI_API_KEY"),
			OpenAIModel:      envOr("OPENAI_MODEL", "gpt-4o"),
			OpenAIModelLight: envOr("OPENAI_MODEL_LIGHT", ""),
			CompactThreshold: envOrInt("COMPACT_THRESHOLD", 20),
			MaxHistoryMsgs:   envOrInt("MAX_HISTORY_MESSAGES", 30),
		},
		WhatsApp: WhatsAppConfig{
			PhoneID:     os.Getenv("WHATSAPP_PHONE_NUMBER_ID"),
			Token:       os.Getenv("WHATSAPP_ACCESS_TOKEN"),
			To:          os.Getenv("WHATSAPP_TO_NUMBER"),
			VerifyToken: os.Getenv("WHATSAPP_VERIFY_TOKEN"),
			AppSecret:   os.Getenv("WHATSAPP_APP_SECRET"),
		},
		Telegram: TelegramConfig{
			BotToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
			SecretToken: os.Getenv("TELEGRAM_SECRET_TOKEN"),
			BotUsername: os.Getenv("TELEGRAM_BOT_USERNAME"),
		},
		Google: GoogleConfig{
			SheetsID:        os.Getenv("GOOGLE_SHEETS_ID"),
			SheetsCredFile:  envOr("GOOGLE_CREDENTIALS_FILE", "credentials.json"),
			SheetsSheetName: envOr("GOOGLE_SHEETS_NAME", "Gastos"),
			CalendarID:      os.Getenv("GOOGLE_CALENDAR_ID"),
		},
		Integrations: IntegrationsConfig{
			NotionAPIKey:       os.Getenv("NOTION_API_KEY"),
			NotionPageID:       os.Getenv("NOTION_DEFAULT_PAGE_ID"),
			ObsidianVaultPath:  os.Getenv("OBSIDIAN_VAULT_PATH"),
			GitHubToken:        os.Getenv("GITHUB_TOKEN"),
			JiraBaseURL:        os.Getenv("JIRA_BASE_URL"),
			JiraEmail:          os.Getenv("JIRA_EMAIL"),
			JiraAPIToken:       os.Getenv("JIRA_API_TOKEN"),
			SpotifyAccessToken: os.Getenv("SPOTIFY_ACCESS_TOKEN"),
			TodoistAPIToken:    os.Getenv("TODOIST_API_TOKEN"),
			GmailUserEmail:     os.Getenv("GMAIL_USER_EMAIL"),
			ClickUpAPIToken:    os.Getenv("CLICKUP_API_TOKEN"),
			ClickUpTeamID:      os.Getenv("CLICKUP_TEAM_ID"),
			FigmaAccessToken:   os.Getenv("FIGMA_ACCESS_TOKEN"),
		},
		Runtime: RuntimeConfig{
			AgentsDir:       envOr("AGENTS_DIR", "agents"),
			RulesDir:        envOr("RULES_DIR", "rules"),
			ProfilesDir:     envOr("PROFILES_DIR", "config/profiles"),
			DefaultProfile:  envOr("DEFAULT_PROFILE", "full"),
			HooksConfigFile: envOr("HOOKS_CONFIG_FILE", "config/hooks.yaml"),
			SkillsDir:       envOr("SKILLS_DIR", "skills"),
			DryRunTools:     os.Getenv("DRY_RUN_TOOLS"),
		},
	}
}

func envOrInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
