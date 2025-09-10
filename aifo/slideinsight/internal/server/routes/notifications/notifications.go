// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package notifications

import (
	"aifo.dev/aifo/slideinsight/internal/config"
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/notifications"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"github.com/gofiber/fiber/v2"
)

// SetupNotificationRoutes sets up notification-related routes
func SetupNotificationRoutes(api fiber.Router, service services.NotificationService, userService services.UserService, config config.Config) {
	// Create notification hub for WebSocket management
	hub := notifications.NewHub(service)

	// Create notifications group
	notifications := api.Group("/notifications")

	// Apply authentication middleware to all notification routes
	notifications.Use(middleware.Protected(config.Auth))

	// Notification CRUD routes
	notifications.Post("/", handlers.CreateNotification(service))
	notifications.Post("/test", handlers.CreateTestNotification(service))
	notifications.Get("/", handlers.GetNotifications(service, userService))
	notifications.Get("/stats", handlers.GetNotificationStats(service, userService))
	notifications.Put("/read-all", handlers.MarkAllNotificationsAsRead(service, userService))
	notifications.Put("/:id/read", handlers.MarkNotificationAsRead(service, userService))
	notifications.Put("/:id/dismiss", handlers.MarkNotificationAsDismissed(service, userService))
	notifications.Delete("/:id", handlers.DeleteNotification(service, userService))

	// WebSocket route for real-time notifications - configured separately to avoid middleware conflicts
	// Create a separate sub-router for WebSocket to avoid logging middleware noise
	wsRouter := api.Group("/notifications")

	// Add the WebSocket route without authentication middleware
	// Authentication is handled within the WebSocket handler itself
	wsRouter.Get("/ws", handlers.NotificationWebSocket(hub, userService))

	// Admin routes
	adminNotifications := api.Group("/admin/notifications")
	adminNotifications.Use(middleware.Protected(config.Auth))
	// TODO: Add admin role checking middleware here when available
	adminNotifications.Get("/", handlers.GetAllNotifications(service))
	adminNotifications.Delete("/:id", handlers.AdminDeleteNotification(service))
	adminNotifications.Put("/:id", handlers.AdminUpdateNotification(service))
}
