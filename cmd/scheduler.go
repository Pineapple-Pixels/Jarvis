package main

import (
	"asistente/config"
	"asistente/internal/hooks"
	"asistente/pkg/domain"
	"asistente/pkg/usecase"
)

func NewScheduler(cl Clients, cfg config.Config, hooksRegistry *hooks.Registry) *usecase.Scheduler {
	return usecase.NewScheduler([]domain.Job{
		usecase.NewDailyBriefingJob(cl.AI, cfg.WhatsAppTo, cl.WhatsApp, cl.Calendar),
		usecase.NewWeeklyFinanceJob(cl.AI, cfg.WhatsAppTo, cl.WhatsApp),
		usecase.NewBudgetAlertJob(cl.AI, cfg.WhatsAppTo, cl.WhatsApp),
		usecase.NewDailyJournalJob(cl.AI, cfg.WhatsAppTo, cl.WhatsApp),
	}, hooksRegistry)
}
