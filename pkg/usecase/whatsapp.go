package usecase

import (
	"context"
	"log"
	"strings"

	"asistente/internal/hooks"
	"asistente/internal/skills"
	"asistente/pkg/domain"
	"asistente/pkg/service"
)

const whatsAppSessionPrefix = "whatsapp-"

type WhatsAppUseCase struct {
	conversation *ConversationUseCase
	finance      *FinanceUseCase
	memorySvc    service.MemoryService
	embedder     service.Embedder
	ai           domain.AIProvider
	wa           domain.WhatsAppSender
	skills       skills.SkillProvider
	hooks        *hooks.Registry
	allowedFrom  string
}

func NewWhatsAppUseCase(
	conversation *ConversationUseCase,
	finance *FinanceUseCase,
	memorySvc service.MemoryService,
	embedder service.Embedder,
	ai domain.AIProvider,
	wa domain.WhatsAppSender,
	skillsProvider skills.SkillProvider,
	hooksRegistry *hooks.Registry,
	allowedFrom string,
) *WhatsAppUseCase {
	return &WhatsAppUseCase{
		conversation: conversation,
		finance:      finance,
		memorySvc:    memorySvc,
		embedder:     embedder,
		ai:           ai,
		wa:           wa,
		skills:       skillsProvider,
		hooks:        hooksRegistry,
		allowedFrom:  allowedFrom,
	}
}

// ProcessMessage handles an incoming WhatsApp text message.
// It detects intent, delegates to the appropriate usecase, and sends a reply.
func (uc *WhatsAppUseCase) ProcessMessage(from, messageID, text string) {
	if uc.allowedFrom != "" && from != uc.allowedFrom {
		log.Printf("whatsapp: ignoring message from unauthorized number %s", from)
		return
	}

	_ = uc.markAsRead(messageID)

	intent := detectIntent(text)
	log.Printf("whatsapp: from=%s intent=%s msg=%q", from, intent, truncate(text, 80))

	response, err := uc.handleIntent(intent, from, text)
	if err != nil {
		log.Printf("whatsapp: error handling intent %s: %v", intent, err)
		response = "Perdón, hubo un error procesando tu mensaje. Intentá de nuevo."
	}

	if err := uc.wa.SendTextMessage(from, response); err != nil {
		log.Printf("whatsapp: failed to send reply to %s: %v", from, err)
	}

	uc.hooks.Emit(context.Background(), hooks.WhatsAppMessageProcessed, map[string]string{
		"from": from, "intent": intent, "message": text,
	})
}

type intentType = string

const (
	intentExpense  intentType = "expense"
	intentNote     intentType = "note"
	intentChat     intentType = "chat"
)

// detectIntent classifies the message using simple prefix/keyword matching.
// This avoids burning an AI call for every incoming message.
func detectIntent(text string) intentType {
	lower := strings.ToLower(strings.TrimSpace(text))

	expensePrefixes := []string{"gaste ", "gasté ", "gastamos ", "pague ", "pagué ", "pagamos ", "compre ", "compré ", "compramos "}
	for _, p := range expensePrefixes {
		if strings.HasPrefix(lower, p) {
			return intentExpense
		}
	}

	expenseKeywords := []string{" lucas", " luquitas", " pesos", " dolares", " dólares", " usd"}
	for _, k := range expenseKeywords {
		if strings.Contains(lower, k) {
			return intentExpense
		}
	}

	notePrefixes := []string{"nota ", "nota: ", "recordá ", "recorda ", "recordame ", "acordate "}
	for _, p := range notePrefixes {
		if strings.HasPrefix(lower, p) {
			return intentNote
		}
	}

	return intentChat
}

func (uc *WhatsAppUseCase) handleIntent(intent, from, text string) (string, error) {
	switch intent {
	case intentExpense:
		return uc.handleExpense(text)
	case intentNote:
		return uc.handleNote(text)
	default:
		return uc.handleChat(from, text)
	}
}

func (uc *WhatsAppUseCase) handleExpense(text string) (string, error) {
	if uc.finance == nil {
		return "El módulo de finanzas no está configurado.", nil
	}
	return uc.finance.ProcessExpense(text, "Sebas")
}

func (uc *WhatsAppUseCase) handleNote(text string) (string, error) {
	if uc.memorySvc == nil {
		return "El módulo de notas no está configurado.", nil
	}

	content := stripNotePrefix(text)

	var embedding []float64
	if uc.embedder != nil {
		emb, err := uc.embedder.Embed(content)
		if err != nil {
			log.Printf("whatsapp: embedding failed, saving without: %v", err)
		} else {
			embedding = emb
		}
	}

	_, err := uc.memorySvc.Save(content, []string{"whatsapp"}, embedding)
	if err != nil {
		return "", domain.Wrapf(domain.ErrStoreSave, err)
	}

	return "Anotado!", nil
}

func (uc *WhatsAppUseCase) handleChat(from, text string) (string, error) {
	sessionID := whatsAppSessionPrefix + from

	if err := uc.conversation.Ingest(sessionID, domain.RoleUser, text); err != nil {
		return "", err
	}

	messages, err := uc.conversation.Assemble(sessionID)
	if err != nil {
		return "", err
	}

	systemPrompt := uc.buildSystemPrompt(text)

	response, err := uc.ai.CompleteMessages(systemPrompt, messages)
	if err != nil {
		return "", err
	}

	_ = uc.conversation.Ingest(sessionID, domain.RoleAssistant, response)

	return response, nil
}

func (uc *WhatsAppUseCase) buildSystemPrompt(message string) string {
	var sb strings.Builder
	sb.WriteString(domain.DefaultSystemPrompt)
	sb.WriteString("El usuario te habla por WhatsApp. Sé conciso.\n\n")

	if uc.skills == nil {
		return sb.String()
	}

	loaded, err := uc.skills.LoadEnabled()
	if err != nil || len(loaded) == 0 {
		return sb.String()
	}

	tags := skills.ClassifyMessage(message)
	var relevant []skills.Skill
	if len(tags) == 0 {
		relevant = loaded
	} else {
		relevant = skills.FilterByTags(loaded, tags...)
	}

	sb.WriteString(domain.SkillsSectionHeader)
	sb.WriteString(skills.FormatForPrompt(relevant))

	return sb.String()
}

func (uc *WhatsAppUseCase) markAsRead(messageID string) error {
	type readMarker interface {
		MarkAsRead(messageID string) error
	}
	if rm, ok := uc.wa.(readMarker); ok {
		return rm.MarkAsRead(messageID)
	}
	return nil
}

func stripNotePrefix(text string) string {
	lower := strings.ToLower(text)
	prefixes := []string{"nota: ", "nota ", "recordá ", "recorda ", "recordame ", "acordate "}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return strings.TrimSpace(text[len(p):])
		}
	}
	return text
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
