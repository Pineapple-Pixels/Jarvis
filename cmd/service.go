package main

import (
	"log"
	"os"
	"path/filepath"

	"asistente/clients"
	"asistente/config"
	"asistente/pkg/domain"
	"asistente/pkg/service"
)

func NewMemoryService(cfg config.Config) service.MemoryService {
	switch cfg.StorageBackend {
	case "postgres":
		if cfg.PostgresDSN == "" {
			log.Fatal("POSTGRES_DSN is required when STORAGE_BACKEND=postgres")
		}
		store, err := service.NewPGMemoryService(cfg.PostgresDSN)
		if err != nil {
			log.Fatalf("failed to init postgres store: %v", err)
		}
		log.Println("Storage backend: PostgreSQL")
		return store
	default:
		if err := os.MkdirAll(filepath.Dir(cfg.SQLiteDBPath), 0o755); err != nil {
			log.Fatalf("failed to create data dir: %v", err)
		}
		store, err := service.NewSQLiteMemoryService(cfg.SQLiteDBPath)
		if err != nil {
			log.Fatalf("failed to init sqlite store: %v", err)
		}
		log.Println("Storage backend: SQLite")
		return store
	}
}

func NewFinanceService(sheetsClient *clients.SheetsClient, sheetName string) service.FinanceService {
	if sheetsClient == nil {
		return nil
	}
	return service.NewSheetsFinanceService(sheetsClient, sheetName)
}

func NewEmbedder(ai domain.AIProvider) service.Embedder {
	inner := service.NewAIEmbedder(ai)
	return service.NewCachedEmbedder(inner, 500)
}
