package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractTextMessages_FiltersCorrectly(t *testing.T) {
	payload := WhatsAppWebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []WhatsAppEntry{{
			Changes: []WhatsAppChange{{
				Value: WhatsAppValue{
					Messages: []WhatsAppIncomingMessage{
						{From: "123", ID: "m1", Type: "text", Text: WhatsAppTextBody{Body: "hola"}},
						{From: "123", ID: "m2", Type: "image"},
						{From: "456", ID: "m3", Type: "text", Text: WhatsAppTextBody{Body: "chau"}},
						{From: "789", ID: "m4", Type: "text", Text: WhatsAppTextBody{Body: ""}},
					},
				},
			}},
		}},
	}

	msgs := ExtractTextMessages(payload)

	require.Len(t, msgs, 2)
	assert.Equal(t, "hola", msgs[0].Text.Body)
	assert.Equal(t, "chau", msgs[1].Text.Body)
}

func TestExtractTextMessages_EmptyPayload(t *testing.T) {
	msgs := ExtractTextMessages(WhatsAppWebhookPayload{})

	assert.Empty(t, msgs)
}

func TestExtractTextMessages_StatusOnly(t *testing.T) {
	payload := WhatsAppWebhookPayload{
		Entry: []WhatsAppEntry{{
			Changes: []WhatsAppChange{{
				Value: WhatsAppValue{
					Statuses: []WhatsAppStatus{{ID: "1", Status: "delivered"}},
				},
			}},
		}},
	}

	msgs := ExtractTextMessages(payload)

	assert.Empty(t, msgs)
}

func TestWhatsAppVerifyRequest_Validate(t *testing.T) {
	valid := WhatsAppVerifyRequest{Mode: "subscribe", Token: "tok", Challenge: "ch"}
	assert.NoError(t, valid.Validate())

	missing := WhatsAppVerifyRequest{Mode: "subscribe", Token: "tok"}
	assert.Error(t, missing.Validate())

	badMode := WhatsAppVerifyRequest{Mode: "unsubscribe", Token: "tok", Challenge: "ch"}
	assert.Error(t, badMode.Validate())
}
