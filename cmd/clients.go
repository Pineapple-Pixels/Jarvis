package main

import (
	"log"

	"jarvis/clients"
	"jarvis/config"
	"jarvis/pkg/domain"
)

type Clients struct {
	AI       domain.AIProvider
	AILight  domain.AIProvider
	Sheets   *clients.SheetsClient
	WhatsApp *clients.WhatsAppClient
	Calendar *clients.CalendarClient
	Notion   *clients.NotionClient
	Obsidian *clients.ObsidianVault
	GitHub   *clients.GitHubClient
	Jira     *clients.JiraClient
	Spotify  *clients.SpotifyClient
	Todoist  *clients.TodoistClient
	Gmail    *clients.GmailClient
	ClickUp  *clients.ClickUpClient
	Figma       *clients.FigmaClient
	Telegram    *clients.TelegramClient
	Transcriber domain.Transcriber
}

func NewClients(cfg config.Config) Clients {
	ai := newAIProviderWithFailover(cfg)
	return Clients{
		AI:       ai,
		AILight:  newAILightProvider(cfg, ai),
		Sheets:   newSheetsClient(cfg),
		WhatsApp: newWhatsAppClient(cfg),
		Calendar: newCalendarClient(cfg),
		Notion:   newNotionClient(cfg),
		Obsidian: newObsidianVault(cfg),
		GitHub:   newGitHubClient(cfg),
		Jira:     newJiraClient(cfg),
		Spotify:  newSpotifyClient(cfg),
		Todoist:  newTodoistClient(cfg),
		Gmail:    newGmailClient(cfg),
		ClickUp:  newClickUpClient(cfg),
		Figma:       newFigmaClient(cfg),
		Telegram:    newTelegramClient(cfg),
		Transcriber: newTranscriber(cfg),
	}
}

func newAIProviderWithFailover(cfg config.Config) domain.AIProvider {
	primary := newAIProvider(cfg)
	fallback := newAIFallback(cfg)
	return clients.NewFailoverProvider(primary, fallback)
}

func newAIProvider(cfg config.Config) domain.AIProvider {
	switch cfg.AI.Provider {
	case "openai":
		log.Printf("AI provider: OpenAI (model: %s)", cfg.AI.OpenAIModel)
		return clients.NewOpenAIClient(cfg.AI.OpenAIAPIKey, cfg.AI.OpenAIModel)
	default:
		log.Printf("AI provider: Claude (model: %s)", cfg.AI.ClaudeModel)
		return clients.NewClaudeClient(cfg.AI.ClaudeAPIKey, cfg.AI.ClaudeModel)
	}
}

func newAIFallback(cfg config.Config) domain.AIProvider {
	switch cfg.AI.Provider {
	case "openai":
		if cfg.AI.ClaudeAPIKey == "" {
			return nil
		}
		log.Printf("AI fallback: Claude (model: %s)", cfg.AI.ClaudeModel)
		return clients.NewClaudeClient(cfg.AI.ClaudeAPIKey, cfg.AI.ClaudeModel)
	default:
		if cfg.AI.OpenAIAPIKey == "" {
			return nil
		}
		log.Printf("AI fallback: OpenAI (model: %s)", cfg.AI.OpenAIModel)
		return clients.NewOpenAIClient(cfg.AI.OpenAIAPIKey, cfg.AI.OpenAIModel)
	}
}

func newAILightProvider(cfg config.Config, primary domain.AIProvider) domain.AIProvider {
	switch cfg.AI.Provider {
	case "openai":
		if cfg.AI.OpenAIModelLight == "" || cfg.AI.OpenAIModelLight == cfg.AI.OpenAIModel {
			return primary
		}
		log.Printf("AI light provider: OpenAI (model: %s)", cfg.AI.OpenAIModelLight)
		return clients.NewOpenAIClient(cfg.AI.OpenAIAPIKey, cfg.AI.OpenAIModelLight)
	default:
		if cfg.AI.ClaudeModelLight == "" || cfg.AI.ClaudeModelLight == cfg.AI.ClaudeModel {
			return primary
		}
		log.Printf("AI light provider: Claude (model: %s)", cfg.AI.ClaudeModelLight)
		return clients.NewClaudeClient(cfg.AI.ClaudeAPIKey, cfg.AI.ClaudeModelLight)
	}
}

func newSheetsClient(cfg config.Config) *clients.SheetsClient {
	if cfg.Google.SheetsID == "" || cfg.Google.SheetsCredFile == "" {
		return nil
	}
	client, err := clients.NewSheetsClient(cfg.Google.SheetsCredFile, cfg.Google.SheetsID)
	if err != nil {
		log.Printf("WARNING: sheets client not available: %v", err)
		return nil
	}
	return client
}

func newWhatsAppClient(cfg config.Config) *clients.WhatsAppClient {
	if cfg.WhatsApp.PhoneID == "" || cfg.WhatsApp.Token == "" {
		return nil
	}
	log.Println("WhatsApp client configured")
	return clients.NewWhatsAppClient(cfg.WhatsApp.PhoneID, cfg.WhatsApp.Token)
}

func newCalendarClient(cfg config.Config) *clients.CalendarClient {
	if cfg.Google.CalendarID == "" || cfg.Google.SheetsCredFile == "" {
		return nil
	}
	client, err := clients.NewCalendarClient(cfg.Google.SheetsCredFile, cfg.Google.CalendarID)
	if err != nil {
		log.Printf("WARNING: calendar client not available: %v", err)
		return nil
	}
	return client
}

func newNotionClient(cfg config.Config) *clients.NotionClient {
	if cfg.Integrations.NotionAPIKey == "" {
		return nil
	}
	log.Println("Notion client configured")
	return clients.NewNotionClient(cfg.Integrations.NotionAPIKey)
}

func newObsidianVault(cfg config.Config) *clients.ObsidianVault {
	if cfg.Integrations.ObsidianVaultPath == "" {
		return nil
	}
	log.Printf("Obsidian vault: %s", cfg.Integrations.ObsidianVaultPath)
	return clients.NewObsidianVault(cfg.Integrations.ObsidianVaultPath)
}

func newGitHubClient(cfg config.Config) *clients.GitHubClient {
	if cfg.Integrations.GitHubToken == "" {
		return nil
	}
	log.Println("GitHub client configured")
	return clients.NewGitHubClient(cfg.Integrations.GitHubToken)
}

func newJiraClient(cfg config.Config) *clients.JiraClient {
	if cfg.Integrations.JiraBaseURL == "" || cfg.Integrations.JiraEmail == "" || cfg.Integrations.JiraAPIToken == "" {
		return nil
	}
	log.Println("Jira client configured")
	return clients.NewJiraClient(cfg.Integrations.JiraBaseURL, cfg.Integrations.JiraEmail, cfg.Integrations.JiraAPIToken)
}

func newSpotifyClient(cfg config.Config) *clients.SpotifyClient {
	if cfg.Integrations.SpotifyAccessToken == "" {
		return nil
	}
	log.Println("Spotify client configured")
	return clients.NewSpotifyClient(cfg.Integrations.SpotifyAccessToken)
}

func newTodoistClient(cfg config.Config) *clients.TodoistClient {
	if cfg.Integrations.TodoistAPIToken == "" {
		return nil
	}
	log.Println("Todoist client configured")
	return clients.NewTodoistClient(cfg.Integrations.TodoistAPIToken)
}

func newGmailClient(cfg config.Config) *clients.GmailClient {
	if cfg.Integrations.GmailUserEmail == "" || cfg.Google.SheetsCredFile == "" {
		return nil
	}
	client, err := clients.NewGmailClient(cfg.Google.SheetsCredFile, cfg.Integrations.GmailUserEmail)
	if err != nil {
		log.Printf("WARNING: gmail client not available: %v", err)
		return nil
	}
	log.Println("Gmail client configured")
	return client
}

func newClickUpClient(cfg config.Config) *clients.ClickUpClient {
	if cfg.Integrations.ClickUpAPIToken == "" {
		return nil
	}
	log.Println("ClickUp client configured")
	return clients.NewClickUpClient(cfg.Integrations.ClickUpAPIToken, cfg.Integrations.ClickUpTeamID)
}

func newTranscriber(cfg config.Config) domain.Transcriber {
	if cfg.AI.OpenAIAPIKey == "" {
		return nil
	}
	log.Println("Transcriber configured (Whisper)")
	return clients.NewOpenAIClient(cfg.AI.OpenAIAPIKey, "")
}

func newTelegramClient(cfg config.Config) *clients.TelegramClient {
	if cfg.Telegram.BotToken == "" {
		return nil
	}
	log.Println("Telegram client configured")
	return clients.NewTelegramClient(cfg.Telegram.BotToken)
}

func newFigmaClient(cfg config.Config) *clients.FigmaClient {
	if cfg.Integrations.FigmaAccessToken == "" {
		return nil
	}
	log.Println("Figma client configured")
	return clients.NewFigmaClient(cfg.Integrations.FigmaAccessToken)
}
