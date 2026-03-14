package clients

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"asistente/pkg/domain"
)

const (
	claudeDefaultBaseURL   = "https://api.anthropic.com/v1/messages"
	claudeDefaultMaxTokens = 2048
	claudeDefaultTimeout   = 30 * time.Second
	anthropicVersion       = "2023-06-01"
	claudeHeaderAPIKey     = "x-api-key"
	claudeHeaderVersion    = "anthropic-version"
)

// Compile-time check: *ClaudeClient implements domain.AIProvider.
var _ domain.AIProvider = (*ClaudeClient)(nil)

type ClaudeClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system,omitempty"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage domain.Usage `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewClaudeClient(apiKey, model string) *ClaudeClient {
	return &ClaudeClient{
		apiKey:     apiKey,
		model:      model,
		baseURL:    claudeDefaultBaseURL,
		httpClient: &http.Client{Timeout: claudeDefaultTimeout},
	}
}

func NewClaudeClientWithBaseURL(apiKey, model, baseURL string) *ClaudeClient {
	c := NewClaudeClient(apiKey, model)
	c.baseURL = baseURL
	return c
}

func (c *ClaudeClient) Complete(system, userMessage string, opts ...domain.CompletionOption) (string, error) {
	return c.CompleteMessages(system, []domain.Message{
		{Role: domain.RoleUser, Content: userMessage},
	}, opts...)
}

// CompleteWithUsage is like Complete but also returns token usage.
func (c *ClaudeClient) CompleteWithUsage(system, userMessage string, opts ...domain.CompletionOption) (string, domain.Usage, error) {
	return c.completeMessagesWithUsage(system, []domain.Message{
		{Role: domain.RoleUser, Content: userMessage},
	}, opts...)
}

func (c *ClaudeClient) CompleteMessages(system string, messages []domain.Message, opts ...domain.CompletionOption) (string, error) {
	text, usage, err := c.completeMessagesWithUsage(system, messages, opts...)
	if err == nil && (usage.InputTokens > 0 || usage.OutputTokens > 0) {
		log.Printf("claude: model=%s in=%d out=%d", c.model, usage.InputTokens, usage.OutputTokens)
	}
	return text, err
}

func (c *ClaudeClient) CompleteJSON(system, userMessage string, target any, opts ...domain.CompletionOption) error {
	text, err := c.Complete(system, userMessage, opts...)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(text), target); err != nil {
		return domain.Wrapf(domain.ErrClaudeJSON, err)
	}

	return nil
}

func (c *ClaudeClient) completeMessagesWithUsage(system string, messages []domain.Message, opts ...domain.CompletionOption) (string, domain.Usage, error) {
	cfg := domain.ApplyOptions(claudeDefaultMaxTokens, opts...)

	// Convert domain.Message to Claude's wire format.
	apiMsgs := make([]claudeMessage, len(messages))
	for i, m := range messages {
		apiMsgs[i] = claudeMessage{Role: m.Role, Content: m.Content}
	}

	reqBody := claudeRequest{
		Model:     c.model,
		MaxTokens: cfg.MaxTokens,
		System:    system,
		Messages:  apiMsgs,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", domain.Usage{}, domain.Wrapf(domain.ErrClaudeMarshal, err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return "", domain.Usage{}, domain.Wrapf(domain.ErrClaudeRequest, err)
	}

	req.Header.Set(headerContentType, contentTypeJSON)
	req.Header.Set(claudeHeaderAPIKey, c.apiKey)
	req.Header.Set(claudeHeaderVersion, anthropicVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", domain.Usage{}, domain.Wrapf(domain.ErrClaudeSend, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", domain.Usage{}, domain.Wrapf(domain.ErrClaudeRead, err)
	}

	var result claudeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", domain.Usage{}, domain.Wrapf(domain.ErrClaudeUnmarshal, err)
	}

	if result.Error != nil {
		return "", domain.Usage{}, domain.Wrap(domain.ErrClaudeAPI, result.Error.Type+": "+result.Error.Message)
	}

	if len(result.Content) == 0 {
		return "", domain.Usage{}, domain.ErrClaudeEmpty
	}

	return result.Content[0].Text, result.Usage, nil
}
