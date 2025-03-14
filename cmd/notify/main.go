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
	// 1. Load config from env variables
	cfg := config.LoadConfig()

	// 2. Initialize logger with LOG_LEVEL
	logger := logging.NewLogger(cfg.LogLevel)
	logger.Info("Starting Notification Service...")

	// 3. Initialize DB (SQLite)
	gormDB, err := db.InitDB(cfg.DatabasePath, logger)
	if err != nil {
		logger.Error("Failed to initialize DB", "error", err)
		os.Exit(1)
	}

	// 4. Start the background retry worker
	retryCtx, retryCancel := context.WithCancel(context.Background())
	go service.StartRetryWorker(retryCtx, gormDB, logger, cfg.RetryIntervalSec, cfg.MaxRetries)

	// 5. Create the router/handlers
	router := transport.CreateRouter(gormDB, logger, cfg.AuthToken, cfg)

	// 6. Start HTTP server
	server := &http.Server{
		Addr:    cfg.ServerAddress(),
		Handler: router,
	}

	// Listen for OS signals for graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("HTTP server listening", "address", cfg.ServerAddress())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			stopChan <- os.Interrupt // force shutdown
		}
	}()

	// Wait until we get an interrupt
	<-stopChan
	logger.Info("Shutting down notification service...")

	// Stop background worker
	retryCancel()

	// Shutdown HTTP server gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Graceful shutdown failed", "error", err)
	} else {
		logger.Info("Server shut down gracefully.")
	}
}
