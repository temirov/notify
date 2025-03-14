package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/temirov/notify/pkg/config"
	"github.com/temirov/notify/pkg/db"
	"github.com/temirov/notify/pkg/logging"
	"github.com/temirov/notify/pkg/service"
	"github.com/temirov/notify/pkg/transport"
)

func main() {
	// 1. Load config
	cfg := config.LoadConfig()

	// 2. Logger
	logger := logging.NewLogger(cfg.LogLevel)
	logger.Info("Starting Notification Service...")

	// 3. DB
	gormDB, err := db.InitDB(cfg.DatabasePath, logger)
	if err != nil {
		logger.Error("Failed to init DB", "error", err)
		os.Exit(1)
	}

	// 4. Start worker
	retryCtx, retryCancel := context.WithCancel(context.Background())
	go service.StartRetryWorker(retryCtx, gormDB, logger, cfg.RetryIntervalSec, cfg.MaxRetries)

	// 5. Create router
	router := transport.CreateRouter(gormDB, logger, cfg.AuthToken)

	// 6. Server
	server := &http.Server{
		Addr:    cfg.ServerAddress(),
		Handler: router,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("HTTP server listening", "address", cfg.ServerAddress())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			stopChan <- os.Interrupt
		}
	}()

	<-stopChan
	logger.Info("Shutting down notification service...")

	// 7. Stop worker
	retryCancel()

	// 8. Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Graceful shutdown failed", "error", err)
	} else {
		logger.Info("Server shut down gracefully.")
	}
}
