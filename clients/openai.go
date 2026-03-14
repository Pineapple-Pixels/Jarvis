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
	openaiDefaultBaseURL   = "https://api.openai.com/v1/chat/completions"
	openaiDefaultMaxTokens = 2048
	openaiDefaultTimeout   = 30 * time.Second
)

// Compile-time check: *OpenAIClient implements domain.AIProvider.
var _ domain.AIProvider = (*OpenAIClient)(nil)

type OpenAIClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

type openaiChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiChatRequest struct {
	Model     string              `json:"model"`
	Messages  []openaiChatMessage `json:"messages"`
	MaxTokens int                 `json:"max_tokens,omitempty"`
}

type openaiChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		apiKey:     apiKey,
		model:      model,
		baseURL:    openaiDefaultBaseURL,
		httpClient: &http.Client{Timeout: openaiDefaultTimeout},
	}
}

func NewOpenAIClientWithBaseURL(apiKey, model, baseURL string) *OpenAIClient {
	c := NewOpenAIClient(apiKey, model)
	c.baseURL = baseURL
	return c
}

func (c *OpenAIClient) Complete(system, userMessage string, opts ...domain.CompletionOption) (string, error) {
	return c.CompleteMessages(system, []domain.Message{
		{Role: domain.RoleUser, Content: userMessage},
	}, opts...)
}

func (c *OpenAIClient) CompleteMessages(system string, messages []domain.Message, opts ...domain.CompletionOption) (string, error) {
	cfg := domain.ApplyOptions(openaiDefaultMaxTokens, opts...)

	var apiMsgs []openaiChatMessage
	if system != "" {
		apiMsgs = append(apiMsgs, openaiChatMessage{Role: "system", Content: system})
	}
	for _, m := range messages {
		apiMsgs = append(apiMsgs, openaiChatMessage{Role: m.Role, Content: m.Content})
	}

	reqBody := openaiChatRequest{
		Model:     c.model,
		Messages:  apiMsgs,
		MaxTokens: cfg.MaxTokens,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", domain.Wrapf(domain.ErrClaudeMarshal, err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return "", domain.Wrapf(domain.ErrClaudeRequest, err)
	}

	req.Header.Set(headerContentType, contentTypeJSON)
	req.Header.Set(headerAuthorization, "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", domain.Wrapf(domain.ErrClaudeSend, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", domain.Wrapf(domain.ErrClaudeRead, err)
	}

	var result openaiChatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", domain.Wrapf(domain.ErrClaudeUnmarshal, err)
	}

	if result.Error != nil {
		return "", domain.Wrap(domain.ErrClaudeAPI, result.Error.Type+": "+result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", domain.ErrClaudeEmpty
	}

	text := result.Choices[0].Message.Content
	if result.Usage.PromptTokens > 0 || result.Usage.CompletionTokens > 0 {
		log.Printf("openai: model=%s in=%d out=%d", c.model, result.Usage.PromptTokens, result.Usage.CompletionTokens)
	}

	return text, nil
}

func (c *OpenAIClient) CompleteJSON(system, userMessage string, target any, opts ...domain.CompletionOption) error {
	text, err := c.Complete(system, userMessage, opts...)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(text), target); err != nil {
		return domain.Wrapf(domain.ErrClaudeJSON, err)
	}

	return nil
}
