package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"jarvis/boot"
	"jarvis/config"
	"jarvis/internal/hooks"
	"jarvis/internal/skills"
	"jarvis/pkg/service"
	"jarvis/pkg/usecase"
)

type App struct {
	memorySvc service.MemoryService
	scheduler *usecase.Scheduler
	ctrls     Controllers
	server    boot.Gin
}

func NewApp(cfg config.Config) *App {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	if cfg.Server.WebhookSecret == "" {
		if cfg.Server.AllowInsecure {
			log.Println("WARNING: WEBHOOK_SECRET is not set — running in insecure mode (ALLOW_INSECURE=true)")
		} else {
			log.Println("WARNING: WEBHOOK_SECRET is not set — all endpoints will return 401; set ALLOW_INSECURE=true to bypass in local dev")
		}
	}

	cl := NewClients(cfg)
	memorySvc := NewMemoryService(cfg)
	hooksRegistry := hooks.NewRegistry()
	if cfg.Runtime.HooksConfigFile != "" {
		if defs, err := hooks.LoadExternalConfig(cfg.Runtime.HooksConfigFile); err == nil && len(defs) > 0 {
			hooksRegistry.RegisterExternal(defs)
			log.Printf("loaded %d external hooks from %s", len(defs), cfg.Runtime.HooksConfigFile)
		}
	}

	scheduler := NewScheduler(cl, cfg, memorySvc, hooksRegistry)
	catalogSvc := NewCatalogService(memorySvc)

	ctrls := NewControllers(cl, cfg, memorySvc,
		NewFinanceService(cl.Sheets, cfg.Google.SheetsSheetName),
		NewEmbedder(cl.AILight),
		skills.NewCachedLoader(skills.NewLoader(cfg.Runtime.SkillsDir)),
		hooksRegistry,
		scheduler,
		catalogSvc,
	)

	scheduler.Start()

	return &App{
		memorySvc: memorySvc,
		scheduler: scheduler,
		ctrls:     ctrls,
		server:    boot.NewGin(middlewareMapper(cfg.Server.WebhookSecret, cfg.Server.AllowInsecure), setupRoutes(ctrls)),
	}
}

func (a *App) Run() {
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("received signal %s, shutting down...", sig)
		if err := a.server.Shutdown(); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	if err := a.server.Run(); err != nil {
		log.Printf("server stopped: %v", err)
	}
}

func (a *App) Close() {
	a.scheduler.Stop()
	a.ctrls.Close()
	a.memorySvc.Close()
}
