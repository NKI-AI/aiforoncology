// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"aifo.dev/aifo/slideinsight/internal/queue"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"github.com/gofiber/fiber/v2"
)

// GetQueueStats returns current queue statistics
// @Summary Get queue statistics
// @Description Retrieve current statistics about the task queue including total tasks, completed tasks, failed tasks, and per-type metrics
// @Tags system
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Queue statistics including running status, workers, rate limits, and task counts"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/system/queue [get]
func GetQueueStats(queueManager *queue.QueueManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if queueManager == nil {
			return middleware.SendError(c, fiber.StatusServiceUnavailable, "queue manager not available")
		}

		stats := queueManager.Stats()
		return c.JSON(fiber.Map{
			"status": "success",
			"data":   stats,
		})
	}
}
