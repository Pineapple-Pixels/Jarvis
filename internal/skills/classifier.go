package skills

import "strings"

var keywordToTags = map[string][]string{
	"gast":      {"finance", "sheets"},
	"pagu":      {"finance", "sheets"},
	"pago":      {"finance", "sheets"},
	"lucas":     {"finance", "sheets"},
	"luquitas":  {"finance", "sheets"},
	"plata":     {"finance", "sheets"},
	"dolar":     {"finance", "sheets"},
	"usd":       {"finance", "sheets"},
	"presupuest":{"finance", "sheets"},

	"spotify":   {"spotify", "music", "player"},
	"musica":    {"spotify", "music", "player"},
	"cancion":   {"spotify", "music", "player"},
	"escuchando":{"spotify", "music", "player"},
	"play":      {"spotify", "music", "player"},
	"pause":     {"spotify", "music", "player"},

	"calendar":  {"calendar", "google"},
	"evento":    {"calendar", "google"},
	"agenda":    {"calendar", "google"},
	"agendame":  {"calendar", "google"},
	"reunion":   {"calendar", "google"},
	"hoy tengo": {"calendar", "google"},

	"github":    {"github", "code", "repos"},
	"repo":      {"github", "code", "repos"},
	"issue":     {"github", "code", "repos"},
	"pull request":{"github", "code", "repos"},
	"pr ":       {"github", "code", "repos"},

	"jira":      {"jira", "tasks", "project"},
	"ticket":    {"jira", "tasks", "project"},
	"sprint":    {"jira", "tasks", "project"},

	"clickup":   {"clickup", "tasks", "project"},

	"todoist":   {"todoist", "tasks", "todo"},
	"tarea":     {"todoist", "tasks", "todo"},
	"pendiente": {"todoist", "tasks", "todo"},

	"gmail":     {"gmail", "email", "google"},
	"mail":      {"gmail", "email", "google"},
	"correo":    {"gmail", "email", "google"},
	"inbox":     {"gmail", "email", "google"},

	"notion":    {"notion", "notes"},
	"obsidian":  {"obsidian", "notes"},
	"vault":     {"obsidian", "notes"},

	"nota":      {"memory", "notes"},
	"guarda":    {"memory", "notes"},
	"recordame": {"memory", "notes"},
	"recorda":   {"memory", "notes"},
	"busca":     {"memory", "notes"},

	"link":      {"links", "bookmarks"},
	"url":       {"links", "bookmarks"},
	"guardame este": {"links", "bookmarks"},

	"habito":    {"habits", "tracking"},
	"ejercicio": {"habits", "tracking"},
	"medite":    {"habits", "tracking"},
	"streak":    {"habits", "tracking"},

	"uade":      {"university", "uade"},
	"facultad":  {"university", "uade"},
	"parcial":   {"university", "uade"},
	"entrega":   {"university", "uade"},
	"materia":   {"university", "uade"},

	"proyecto":  {"projects", "notes"},
	"mythological": {"projects", "notes"},
	"vaultbreakers": {"projects", "notes"},
	"mvp":       {"projects", "notes"},
}

// ClassifyMessage analyzes a user message and returns matching skill tags.
// Returns empty slice if no keywords match (caller should load all skills as fallback).
func ClassifyMessage(msg string) []string {
	lower := strings.ToLower(msg)

	tagSet := make(map[string]struct{})
	for keyword, tags := range keywordToTags {
		if strings.Contains(lower, keyword) {
			for _, t := range tags {
				tagSet[t] = struct{}{}
			}
		}
	}

	if len(tagSet) == 0 {
		return nil
	}

	result := make([]string, 0, len(tagSet))
	for t := range tagSet {
		result = append(result, t)
	}

	return result
}
