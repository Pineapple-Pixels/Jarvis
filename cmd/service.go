package main

import (
	"log"

	"jarvis/clients"
	"jarvis/config"
	"jarvis/pkg/domain"
	"jarvis/pkg/service"
)

func NewMemoryService(cfg config.Config) service.MemoryService {
	if cfg.Database.PostgresDSN == "" {
		log.Fatal("POSTGRES_DSN is required")
	}
	store, err := service.NewPGMemoryService(cfg.Database.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to init postgres store: %v", err)
	}
	log.Println("Storage backend: PostgreSQL")
	return store
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

// NewCatalogService derives a CatalogService from the memory service.
// The type assertion is intentionally isolated here so that NewControllers
// never needs to inspect the concrete type behind the MemoryService interface.
func NewCatalogService(memorySvc service.MemoryService) service.CatalogService {
	if pgMem, ok := memorySvc.(*service.PGMemoryService); ok {
		return service.NewCatalogServiceFromDB(pgMem.DB())
	}
	return service.NullCatalogService{}
}
