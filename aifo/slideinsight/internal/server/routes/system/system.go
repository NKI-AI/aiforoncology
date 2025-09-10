// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package system

import (
	"time"

	"aifo.dev/aifo/slideinsight/internal/config"
	"aifo.dev/aifo/slideinsight/internal/queue"
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/monitor"
	"github.com/gofiber/fiber/v2"
)

// SetupSystemRoutes configures all system-related routes
func SetupSystemRoutes(apiRoutes fiber.Router, queueManager *queue.QueueManager, config config.Config) {
	// System routes group
	systemRoutes := apiRoutes.Group("/v1/system")

	// All system routes require authentication
	systemRoutes.Use(middleware.Protected(config.Auth))

	// Queue management routes
	systemRoutes.Get("/queue", handlers.GetQueueStats(queueManager))

	// System monitoring routes
	systemRoutes.Get("/monitor", monitor.New(monitor.Config{
		Title:   "SlideInsight",
		Refresh: 10 * time.Second,
	}))

	// WebSocket route for real-time system monitoring
	systemRoutes.Get("/monitor/ws", handlers.SystemMonitorWebSocket())
}
