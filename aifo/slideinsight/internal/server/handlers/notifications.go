// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/notifications"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// failedAttempts tracks failed WebSocket connection attempts to reduce log spam
var (
	failedAttempts = make(map[string]int)
	failedMutex    sync.RWMutex
)

// shouldLogFailure determines if we should log this failure based on attempt count
func shouldLogFailure(userUID string) bool {
	failedMutex.Lock()
	defer failedMutex.Unlock()

	failedAttempts[userUID]++
	count := failedAttempts[userUID]

	// Log the first 3 attempts, then every 10th attempt, then every 100th
	if count <= 3 {
		return true
	} else if count <= 30 && count%10 == 0 {
		return true
	} else if count%100 == 0 {
		return true
	}
	return false
}

// clearFailedAttempts clears the failed attempt count for a userUID (on successful connection)
func clearFailedAttempts(userUID string) {
	failedMutex.Lock()
	defer failedMutex.Unlock()
	delete(failedAttempts, userUID)
}

// CreateNotification creates a new notification
// @Summary Create a new notification
// @Description Create a new notification for a user
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notification body domain.NewNotificationRequest true "Notification data"
// @Success 201 {object} domain.Notification "Created notification"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/notifications [post]
func CreateNotification(service services.NotificationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Info("CreateNotification handler called")

		var req domain.NewNotificationRequest
		if err := c.BodyParser(&req); err != nil {
			log.Error("Failed to parse request body", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid request body")
		}

		log.Info(fmt.Sprintf("Parsed notification request: title=%s, userUid=%v, email=%v", req.Title, req.UserUID, req.Email))

		notification, err := service.CreateNotification(c.UserContext(), req)
		if err != nil {
			log.Error("Failed to create notification in service", "error", err)
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		log.Info(fmt.Sprintf("Notification created successfully: id=%d, title=%s", notification.ID, notification.Title))

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status": "success",
			"data":   notification,
		})
	}
}

// GetNotifications retrieves notifications for the current user
// @Summary Get notifications
// @Description Get notifications for the current user with pagination and filtering
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of notifications to return" default(20)
// @Param offset query int false "Number of notifications to skip" default(0)
// @Param type query string false "Filter by notification type"
// @Param priority query string false "Filter by notification priority"
// @Param isRead query bool false "Filter by read status"
// @Param isDismissed query bool false "Filter by dismissed status"
// @Success 200 {object} domain.NotificationsResponse "Notifications with pagination and stats"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/notifications [get]
func GetNotifications(service services.NotificationService, userService services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get current user
		p := middleware.FromContext(c.UserContext())
		if p == nil {
			return middleware.SendError(c, fiber.StatusUnauthorized, "Authentication required")
		}

		user, err := userService.GetInternalUserByEmail(c.UserContext(), p.Email)
		if err != nil {
			return middleware.SendError(c, fiber.StatusUnauthorized, "User not found")
		}

		// Parse query parameters
		limit := c.QueryInt("limit", 20)
		offset := c.QueryInt("offset", 0)

		filter := domain.NotificationFilter{}
		if typeParam := c.Query("type"); typeParam != "" {
			filter.Type = &typeParam
		}
		if priorityParam := c.Query("priority"); priorityParam != "" {
			filter.Priority = &priorityParam
		}
		if isReadParam := c.Query("isRead"); isReadParam != "" {
			if isRead, err := strconv.ParseBool(isReadParam); err == nil {
				filter.IsRead = &isRead
			}
		}
		if isDismissedParam := c.Query("isDismissed"); isDismissedParam != "" {
			if isDismissed, err := strconv.ParseBool(isDismissedParam); err == nil {
				filter.IsDismissed = &isDismissed
			}
		}

		notifications, stats, err := service.GetNotifications(c.UserContext(), user.ID, filter, limit, offset)
		if err != nil {
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		// Calculate pagination info
		totalPages := 0
		if stats != nil && limit > 0 {
			totalPages = (stats.TotalCount + limit - 1) / limit
		}

		pagination := domain.PaginationInfo{
			Page:       (offset / limit) + 1,
			Limit:      limit,
			Total:      0,
			TotalPages: totalPages,
			HasNext:    false,
			HasPrev:    offset > 0,
		}

		if stats != nil {
			pagination.Total = stats.TotalCount
			pagination.HasNext = offset+limit < stats.TotalCount
		}

		response := domain.NotificationsResponse{
			Notifications: notifications,
			Pagination:    pagination,
			Stats:         stats,
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   response,
		})
	}
}

// MarkNotificationAsRead marks a notification as read
// @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/notifications/{id}/read [put]
func MarkNotificationAsRead(service services.NotificationService, userService services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get current user
		p := middleware.FromContext(c.UserContext())
		if p == nil {
			return middleware.SendError(c, fiber.StatusUnauthorized, "Authentication required")
		}

		user, err := userService.GetInternalUserByEmail(c.UserContext(), p.Email)
		if err != nil {
			return middleware.SendError(c, fiber.StatusUnauthorized, "User not found")
		}

		// Parse notification ID
		notificationID, err := c.ParamsInt("id")
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid notification ID")
		}

		err = service.MarkAsRead(c.UserContext(), notificationID, user.ID)
		if err != nil {
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Notification marked as read",
		})
	}
}

// MarkNotificationAsDismissed marks a notification as dismissed
// @Summary Mark notification as dismissed
// @Description Mark a specific notification as dismissed
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/notifications/{id}/dismiss [put]
func MarkNotificationAsDismissed(service services.NotificationService, userService services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get current user
		p := middleware.FromContext(c.UserContext())
		if p == nil {
			return middleware.SendError(c, fiber.StatusUnauthorized, "Authentication required")
		}

		user, err := userService.GetInternalUserByEmail(c.UserContext(), p.Email)
		if err != nil {
			return middleware.SendError(c, fiber.StatusUnauthorized, "User not found")
		}

		// Parse notification ID
		notificationID, err := c.ParamsInt("id")
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid notification ID")
		}

		err = service.MarkAsDismissed(c.UserContext(), notificationID, user.ID)
		if err != nil {
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Notification dismissed",
		})
	}
}

// MarkAllNotificationsAsRead marks all notifications as read
// @Summary Mark all notifications as read
// @Description Mark all notifications as read for the current user
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/notifications/read-all [put]
func MarkAllNotificationsAsRead(service services.NotificationService, userService services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get current user
		p := middleware.FromContext(c.UserContext())
		if p == nil {
			return middleware.SendError(c, fiber.StatusUnauthorized, "Authentication required")
		}

		user, err := userService.GetInternalUserByEmail(c.UserContext(), p.Email)
		if err != nil {
			return middleware.SendError(c, fiber.StatusUnauthorized, "User not found")
		}

		err = service.MarkAllAsRead(c.UserContext(), user.ID)
		if err != nil {
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "All notifications marked as read",
		})
	}
}

// DeleteNotification deletes a notification
// @Summary Delete notification
// @Description Delete a specific notification
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/notifications/{id} [delete]
func DeleteNotification(service services.NotificationService, userService services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get current user
		p := middleware.FromContext(c.UserContext())
		if p == nil {
			return middleware.SendError(c, fiber.StatusUnauthorized, "Authentication required")
		}

		user, err := userService.GetInternalUserByEmail(c.UserContext(), p.Email)
		if err != nil {
			return middleware.SendError(c, fiber.StatusUnauthorized, "User not found")
		}

		// Parse notification ID
		notificationID, err := c.ParamsInt("id")
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid notification ID")
		}

		err = service.DeleteNotification(c.UserContext(), notificationID, user.ID)
		if err != nil {
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Notification deleted",
		})
	}
}

// GetNotificationStats gets notification statistics
// @Summary Get notification statistics
// @Description Get notification statistics for the current user
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.NotificationStats "Notification statistics"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/notifications/stats [get]
func GetNotificationStats(service services.NotificationService, userService services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get current user
		p := middleware.FromContext(c.UserContext())
		if p == nil {
			return middleware.SendError(c, fiber.StatusUnauthorized, "Authentication required")
		}

		user, err := userService.GetInternalUserByEmail(c.UserContext(), p.Email)
		if err != nil {
			return middleware.SendError(c, fiber.StatusUnauthorized, "User not found")
		}

		stats, err := service.GetNotificationStats(c.UserContext(), user.ID)
		if err != nil {
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   stats,
		})
	}
}

// NotificationWebSocket handles WebSocket connections for real-time notifications
// @Summary WebSocket for notifications
// @Description Establish WebSocket connection for real-time notifications
// @Tags notifications
// @Security BearerAuth
// @Router /api/v1/notifications/ws [get]
func NotificationWebSocket(hub *notifications.Hub, userService services.UserService) fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("WebSocket handler panic", "error", r)
			}
		}()

		// Helper function to send error and handle connection cleanup
		sendErrorAndClose := func(errorType, message, code string) {
			errorResponse := map[string]string{
				"type":    "error",
				"message": message,
				"code":    code,
			}
			// Send error response
			if writeErr := conn.WriteJSON(errorResponse); writeErr != nil {
				log.Debug("Failed to send error response to WebSocket", "error", writeErr)
			}
			// Let the connection close naturally rather than forcing it
			time.Sleep(100 * time.Millisecond) // Give time for message to be sent
		}

		// Get user UID from query parameter
		userUid := conn.Query("userUid")
		if userUid == "" {
			log.Debug("WebSocket connection rejected: no user UID provided")
			sendErrorAndClose("auth_error", "Authentication required - userUid parameter missing", "USER_UID_REQUIRED")
			return
		}

		// Look up the user by UID to get the internal user ID
		ctx := context.Background()
		user, err := userService.GetUserByUID(ctx, userUid)
		if err != nil {
			// Only log if this failure should be logged (to reduce spam from repeated failures)
			if shouldLogFailure(userUid) {
				failedMutex.RLock()
				attemptCount := failedAttempts[userUid]
				failedMutex.RUnlock()

				// Use Debug level to reduce log noise
				log.Debug("WebSocket connection failed: user lookup failed",
					"userUid", userUid,
					"remote_addr", conn.RemoteAddr().String(),
					"reason", "user_not_found",
					"attempt_count", attemptCount,
					"error", err.Error())
			}

			sendErrorAndClose("auth_error", "Authentication failed - user not found or session expired", "USER_NOT_FOUND")
			return
		}

		userID := user.ID

		// Clear failed attempts on successful connection
		clearFailedAttempts(userUid)

		log.Info("WebSocket connection established", "userUid", userUid, "userID", userID, "remote_addr", conn.RemoteAddr().String())

		// Register the connection
		hub.RegisterConnection(userID, conn)
		defer hub.UnregisterConnection(userID, conn)

		// Send initial connection confirmation and stats
		confirmMsg := map[string]interface{}{
			"type":    "connection",
			"status":  "authenticated",
			"userUid": userUid,
			"userID":  userID,
		}
		if err := conn.WriteJSON(confirmMsg); err != nil {
			log.Error("Failed to send connection confirmation", "error", err)
		}

		// Send initial stats
		stats, err := hub.GetNotificationStats(ctx, userID)
		if err != nil {
			log.Warn("Failed to get notification stats for WebSocket", "userID", userID, "error", err)
		} else {
			message := domain.WebSocketMessage{
				Type: "stats",
				Data: domain.NotificationWebSocketData{
					Stats: stats,
				},
			}
			if err := conn.WriteJSON(message); err != nil {
				log.Error("Failed to send initial stats", "error", err)
			}
		}

		// Handle incoming messages (ping/pong, etc.)
		for {
			var msg map[string]interface{}
			if err := conn.ReadJSON(&msg); err != nil {
				log.Debug("WebSocket connection closed", "userID", userID, "userUid", userUid, "error", err)
				break
			}

			// Handle ping messages
			if msgType, ok := msg["type"].(string); ok && msgType == "ping" {
				response := map[string]string{"type": "pong"}
				if err := conn.WriteJSON(response); err != nil {
					log.Error("Failed to send pong", "userID", userID, "error", err)
					break
				}
			}
		}
	})
}

// CreateTestNotification creates a test notification for the current user
// @Summary Create a test notification
// @Description Create a test notification for the current user (for testing purposes)
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 201 {object} domain.Notification "Created test notification"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/notifications/test [post]
func CreateTestNotification(service services.NotificationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create a test notification for the current user
		req := domain.NewNotificationRequest{
			Title:    "Test Notification",
			Body:     "This is a test notification created at " + time.Now().Format("15:04:05"),
			Type:     "info",
			Priority: "normal",
			Link:     nil, // No link for test notification
		}

		notification, err := service.CreateNotification(c.UserContext(), req)
		if err != nil {
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status": "success",
			"data":   notification,
		})
	}
}

// GetAllNotifications retrieves all notifications for admin purposes
// @Summary Get all notifications (admin only)
// @Description Get all notifications in the system for admin monitoring
// @Tags notifications,admin
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of notifications to return" default(50)
// @Param offset query int false "Number of notifications to skip" default(0)
// @Param type query string false "Filter by notification type"
// @Param priority query string false "Filter by notification priority"
// @Param isRead query bool false "Filter by read status"
// @Param isDismissed query bool false "Filter by dismissed status"
// @Success 200 {object} domain.AdminNotificationsResponse "All notifications with pagination and user info"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 403 {object} domain.ErrorResponse "Forbidden - Admin access required"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/admin/notifications [get]
func GetAllNotifications(service services.NotificationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Info("GetAllNotifications handler called (admin)")

		// Parse query parameters
		limit := c.QueryInt("limit", 50)
		offset := c.QueryInt("offset", 0)

		filter := domain.NotificationFilter{}
		if typeParam := c.Query("type"); typeParam != "" {
			filter.Type = &typeParam
		}
		if priorityParam := c.Query("priority"); priorityParam != "" {
			filter.Priority = &priorityParam
		}
		if isReadParam := c.Query("isRead"); isReadParam != "" {
			if isRead, err := strconv.ParseBool(isReadParam); err == nil {
				filter.IsRead = &isRead
			}
		}
		if isDismissedParam := c.Query("isDismissed"); isDismissedParam != "" {
			if isDismissed, err := strconv.ParseBool(isDismissedParam); err == nil {
				filter.IsDismissed = &isDismissed
			}
		}

		log.Info(fmt.Sprintf("Getting all notifications with filter: limit=%d, offset=%d, filter=%+v", limit, offset, filter))

		// Get all notifications with user information
		notifications, totalCount, err := service.GetAllNotificationsWithUserInfo(c.UserContext(), filter, limit, offset)
		if err != nil {
			log.Error("Failed to get all notifications with user info", "error", err)
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		// Calculate pagination info
		totalPages := 0
		if totalCount > 0 && limit > 0 {
			totalPages = (totalCount + limit - 1) / limit
		}

		pagination := domain.PaginationInfo{
			Page:       (offset / limit) + 1,
			Limit:      limit,
			Total:      totalCount,
			TotalPages: totalPages,
			HasNext:    offset+limit < totalCount,
			HasPrev:    offset > 0,
		}

		response := domain.AdminNotificationsResponse{
			Notifications: notifications,
			Pagination:    pagination,
			Stats:         nil, // No stats for admin view
		}

		log.Info(fmt.Sprintf("Retrieved all notifications successfully: count=%d, total=%d", len(notifications), totalCount))

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   response,
		})
	}
}

// AdminDeleteNotification deletes any notification (admin only)
// @Summary Delete notification (admin only)
// @Description Delete any notification in the system (admin only)
// @Tags notifications,admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 403 {object} domain.ErrorResponse "Forbidden - Admin access required"
// @Failure 404 {object} domain.ErrorResponse "Not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/admin/notifications/{id} [delete]
func AdminDeleteNotification(service services.NotificationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Info("AdminDeleteNotification handler called")

		// Parse notification ID
		notificationID, err := c.ParamsInt("id")
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid notification ID")
		}

		err = service.AdminDeleteNotification(c.UserContext(), notificationID)
		if err != nil {
			log.Error("Failed to delete notification (admin)", "error", err, "id", notificationID)
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		log.Info("Notification deleted successfully (admin)", "id", notificationID)

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Notification deleted",
		})
	}
}

// AdminUpdateNotification updates any notification (admin only)
// @Summary Update notification (admin only)
// @Description Update any notification in the system (admin only)
// @Tags notifications,admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Param notification body domain.UpdateNotificationRequest true "Updated notification data"
// @Success 200 {object} domain.Notification "Updated notification"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 403 {object} domain.ErrorResponse "Forbidden - Admin access required"
// @Failure 404 {object} domain.ErrorResponse "Not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/admin/notifications/{id} [put]
func AdminUpdateNotification(service services.NotificationService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Info("AdminUpdateNotification handler called")

		// Parse notification ID
		notificationID, err := c.ParamsInt("id")
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid notification ID")
		}

		var req domain.UpdateNotificationRequest
		if err := c.BodyParser(&req); err != nil {
			log.Error("Failed to parse request body", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid request body")
		}

		log.Info("Parsed notification update request", "id", notificationID, "title", req.Title)

		notification, err := service.AdminUpdateNotification(c.UserContext(), notificationID, req)
		if err != nil {
			log.Error("Failed to update notification (admin)", "error", err, "id", notificationID)
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		log.Info("Notification updated successfully (admin)", "id", notificationID, "title", notification.Title)

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   notification,
		})
	}
}
