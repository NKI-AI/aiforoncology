// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.

// SlideScope API
//
//	@title			SlideScope API
//	@version		1.0
//	@description	A digital pathology slide management and viewing API
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	API Support
//	@contact.url	http://www.example.com/support
//	@contact.email	support@example.com
//
//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT
//
//	@host		localhost:8080
//	@BasePath	/api/v1
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"aifo.dev/aifo/slidescope/internal/config"
	"aifo.dev/aifo/slidescope/internal/datasources"
	"aifo.dev/aifo/slidescope/internal/datasources/database"
	"aifo.dev/aifo/slidescope/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse command-line flags
	configFile := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, warnings, err := config.LoadFromFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Merge with environment variables
	cfg = config.MergeWithEnv(cfg)

	// Log configuration warnings
	for _, warning := range warnings {
		slog.Warn("Configuration warning", "field", warning.Field, "message", warning.Message)
	}

	// Initialize logger based on configuration
	initLogger(cfg.Logging.Level)

	// Connect to database
	db, err := database.NewDatabase(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.CloseConnections()

	// Create and start server
	app := server.NewServer(ctx, cfg, &datasources.DataSources{DB: db})
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	slog.Info("Starting server", "address", serverAddr)
	log.Fatal(app.Listen(serverAddr))
}

// initLogger configures the global logger based on the specified level
func initLogger(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
