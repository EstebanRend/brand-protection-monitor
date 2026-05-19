package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"brand-protection-monitor/backend/internal/certificates"
	"brand-protection-monitor/backend/internal/config"
	"brand-protection-monitor/backend/internal/db"
	"brand-protection-monitor/backend/internal/httpapi"
	"brand-protection-monitor/backend/internal/keywords"
	"brand-protection-monitor/backend/internal/monitor"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(cfg.DatabaseURL, cfg.MigrationsDir); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	keywordRepo := keywords.NewRepository(database)
	certificateRepo := certificates.NewRepository(database)
	stateRepo := monitor.NewStateRepository(database)
	ctClient := monitor.NewCTClient(cfg.CTLogBaseURL)

	monitorService := monitor.NewService(ctClient, keywordRepo, certificateRepo, stateRepo, cfg.BatchSize, cfg.CTLogBaseURL)
	apiServer := httpapi.NewServer(keywordRepo, certificateRepo, stateRepo, monitorService)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go monitorService.Start(ctx, cfg.MonitorInterval)

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: apiServer.Router(),
	}

	go func() {
		log.Printf("backend running on http://localhost:%s", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()
	_ = server.Shutdown(context.Background())
}
