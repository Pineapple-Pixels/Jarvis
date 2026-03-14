package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type WhatsAppClient struct {
	phoneNumberID string
	accessToken   string
	httpClient    *http.Client
}

func NewWhatsAppClient(phoneNumberID, accessToken string) *WhatsAppClient {
	return &WhatsAppClient{
		phoneNumberID: phoneNumberID,
		accessToken:   accessToken,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *WhatsAppClient) SendTextMessage(to, text string) error {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", c.phoneNumberID)

	body := map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text":              map[string]string{"body": text},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("whatsapp: marshal body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("whatsapp: create request: %w", err)
	}

	req.Header.Set(headerContentType, contentTypeJSON)
	req.Header.Set(headerAuthorization, "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("whatsapp: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("whatsapp: api error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
