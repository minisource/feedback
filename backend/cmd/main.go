package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minisource/feedback/config"
	_ "github.com/minisource/feedback/docs" // Swagger docs
	"github.com/minisource/feedback/internal/database"
	"github.com/minisource/feedback/internal/router"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/go-sdk/auth"
)

// @title Feedback Service API
// @version 1.0
// @description Feedback management service for Minisource
// @host localhost:5012
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logging.NewLogger(&logging.LoggerConfig{
		FilePath: cfg.Logging.FilePath,
		Encoding: cfg.Logging.Encoding,
		Level:    cfg.Logging.Level,
		Logger:   cfg.Logging.Logger,
	})

	logger.Info(logging.General, logging.Startup, "Starting feedback service...", nil)

	// Initialize MongoDB
	db, err := database.NewMongoDB(cfg.MongoDB)
	if err != nil {
		logger.Fatal(logging.General, logging.Startup, "Failed to connect to MongoDB", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
	}
	defer func() {
		if err := db.Close(context.Background()); err != nil {
			logger.Error(logging.General, logging.Startup, "Failed to close MongoDB", map[logging.ExtraKey]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Create indexes
	if err := db.CreateIndexes(context.Background()); err != nil {
		logger.Error(logging.General, logging.Startup, "Failed to create indexes", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info(logging.General, logging.Startup, "MongoDB connected successfully", nil)

	// Initialize auth client
	authClient := auth.NewClient(auth.ClientConfig{
		BaseURL:      cfg.Auth.BaseURL,
		ClientID:     cfg.Auth.ClientID,
		ClientSecret: cfg.Auth.ClientSecret,
		Timeout:      time.Duration(cfg.Auth.Timeout) * time.Second,
		AutoRefresh:  true,
		Logger:       logger,
	})

	logger.Info(logging.General, logging.Startup, "Auth client initialized", nil)

	// Setup router
	r := router.NewRouter(db, cfg, logger, authClient)
	app := r.App()

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		logger.Info(logging.General, logging.Startup, fmt.Sprintf("Server starting on %s", addr), nil)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(logging.General, logging.Startup, "Shutting down server...", nil)

	// Graceful shutdown with timeout
	shutdownTimeout := 30 * time.Second
	if cfg.Server.IdleTimeout > 0 {
		shutdownTimeout = time.Duration(cfg.Server.IdleTimeout) * time.Second
	}
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error(logging.General, logging.Startup, "Server forced to shutdown", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info(logging.General, logging.Startup, "Server exited", nil)
}
