package main

import (
	"jarvis/config"
	"jarvis/internal/hooks"
	"jarvis/pkg/domain"
	"jarvis/pkg/service"
	"jarvis/pkg/usecase"
)

func NewScheduler(cl Clients, cfg config.Config, memorySvc service.MemoryService, hooksRegistry *hooks.Registry) *usecase.Scheduler {
	return usecase.NewScheduler([]domain.Job{
		usecase.NewDailyBriefingJob(cl.AI, cfg.WhatsApp.To, cl.WhatsApp, cl.Calendar, cl.Gmail, memorySvc),
		usecase.NewWeeklyFinanceJob(cl.AI, cfg.WhatsApp.To, cl.WhatsApp),
		usecase.NewBudgetAlertJob(cl.AI, cfg.WhatsApp.To, cl.WhatsApp),
		usecase.NewDailyJournalJob(cl.AI, cfg.WhatsApp.To, cl.WhatsApp),
		usecase.NewSessionPruningJob(memorySvc),
	}, hooksRegistry)
}
