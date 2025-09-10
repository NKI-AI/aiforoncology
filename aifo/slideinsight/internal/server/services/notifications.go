// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2/log"
)

// NotificationService defines the interface for notification operations
type NotificationService interface {
	// Notification CRUD operations
	CreateNotification(ctx context.Context, req domain.NewNotificationRequest) (*domain.Notification, error)
	GetNotifications(ctx context.Context, userID int, filter domain.NotificationFilter, limit, offset int) ([]domain.Notification, *domain.NotificationStats, error)
	GetAllNotifications(ctx context.Context, filter domain.NotificationFilter, limit, offset int) ([]domain.Notification, int, error)
	GetAllNotificationsWithUserInfo(ctx context.Context, filter domain.NotificationFilter, limit, offset int) ([]domain.AdminNotification, int, error)
	MarkAsRead(ctx context.Context, notificationID int, userID int) error
	MarkAsDismissed(ctx context.Context, notificationID int, userID int) error
	MarkAllAsRead(ctx context.Context, userID int) error
	MarkAllAsDismissed(ctx context.Context, userID int) error
	DeleteNotification(ctx context.Context, notificationID int, userID int) error
	GetNotificationStats(ctx context.Context, userID int) (*domain.NotificationStats, error)

	// Admin operations (no user restrictions)
	AdminDeleteNotification(ctx context.Context, notificationID int) error
	AdminUpdateNotification(ctx context.Context, notificationID int, req domain.UpdateNotificationRequest) (*domain.Notification, error)

	// Notification preferences
	GetNotificationPreferences(ctx context.Context, userID int) (*domain.NotificationPreferences, error)
	UpdateNotificationPreferences(ctx context.Context, userID int, preferences domain.NotificationPreferences) error

	// WebSocket operations
	RegisterWebSocketConnection(userID int, conn *websocket.Conn)
	UnregisterWebSocketConnection(userID int, conn *websocket.Conn)
	BroadcastToUser(userID int, message domain.WebSocketMessage)
	BroadcastNotificationToUser(userID int, notification *domain.Notification)

	// Utility operations
	CleanupExpiredNotifications(ctx context.Context) error
	Close()
}

type notificationService struct {
	db          ports.Database
	userService UserService

	// WebSocket connection management
	connections map[int][]*websocket.Conn // userID -> list of connections
	mutex       sync.RWMutex
}

// NewNotificationService creates a new notification service
func NewNotificationService(db ports.Database, userService UserService) NotificationService {
	return &notificationService{
		db:          db,
		userService: userService,
		connections: make(map[int][]*websocket.Conn),
	}
}

// CreateNotification creates a new notification
func (s *notificationService) CreateNotification(ctx context.Context, req domain.NewNotificationRequest) (*domain.Notification, error) {
	log.Info("NotificationService.CreateNotification called", "title", req.Title, "userUid", req.UserUID, "email", req.Email)

	// Determine target user ID
	var targetUserID int
	var err error

	if req.UserUID != nil {
		log.Info("Looking up user by UID", "userUid", *req.UserUID)
		// Get user by UID
		user, err := s.userService.GetUserByUID(ctx, *req.UserUID)
		if err != nil {
			log.Error("Failed to get user by UID", "userUid", *req.UserUID, "error", err)
			return nil, errors.WithDetails(errors.ErrUserNotFound, "user not found: %v", err)
		}
		targetUserID = user.ID
		log.Info("Found user by UID", "userUid", *req.UserUID, "userId", targetUserID, "email", user.Email)
	} else if req.Email != nil {
		log.Info("Looking up user by email", "email", *req.Email)
		// Get user by email
		user, err := s.userService.GetInternalUserByEmail(ctx, *req.Email)
		if err != nil {
			log.Error("Failed to get user by email", "email", *req.Email, "error", err)
			return nil, errors.WithDetails(errors.ErrUserNotFound, "user not found: %v", err)
		}
		targetUserID = user.ID
		log.Info("Found user by email", "email", *req.Email, "userId", targetUserID)
	} else {
		log.Info("No specific user provided, getting current user from context")
		// Get current user from context
		authService := NewBaseService(s.db)
		authCtx, err := authService.GetAuthContext(ctx)
		if err != nil {
			log.Error("Failed to get auth context", "error", err)
			return nil, errors.WithDetails(errors.ErrInvalidInput, "no target user specified and no authenticated user: %v", err)
		}
		targetUserID = authCtx.CreatorID
		log.Info("Using current user from context", "userId", targetUserID)
	}

	// Set defaults
	notificationType := "info"
	if req.Type != "" {
		notificationType = req.Type
	}

	priority := "normal"
	if req.Priority != "" {
		priority = req.Priority
	}

	log.Info("Creating notification with params", "targetUserId", targetUserID, "type", notificationType, "priority", priority)

	// Parse expiration time if provided
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		parsedTime, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			log.Error("Failed to parse expiration time", "expiresAt", *req.ExpiresAt, "error", err)
			return nil, errors.WithDetails(errors.ErrInvalidInput, "invalid expiration time: %v", err)
		}
		expiresAt = &parsedTime
		log.Info("Parsed expiration time", "expiresAt", parsedTime)
	}

	// Create the notification
	newNotification := ports.NewNotification{
		UserID:    targetUserID,
		Title:     req.Title,
		Body:      req.Body,
		Link:      req.Link,
		Type:      notificationType,
		Priority:  priority,
		ExpiresAt: expiresAt,
	}

	log.Info("Calling database CreateNotification", "notification", newNotification)
	notification, err := s.db.CreateNotification(ctx, newNotification)
	if err != nil {
		log.Error("Database CreateNotification failed", "error", err)
		return nil, errors.WithDetails(errors.ErrDatabaseInsert, "failed to create notification: %v", err)
	}

	log.Info("Database notification created successfully", "id", notification.ID)

	// Convert to domain model
	domainNotification := s.convertNotificationToDomain(*notification)

	// Broadcast to user via WebSocket
	log.Info("Broadcasting notification to user", "userId", targetUserID, "notificationId", notification.ID)
	go s.BroadcastNotificationToUser(targetUserID, &domainNotification)

	// Also broadcast updated stats to ensure the counter updates immediately
	log.Info("Broadcasting updated stats to user", "userId", targetUserID)
	go s.broadcastStatsToUser(ctx, targetUserID)

	log.Info("CreateNotification completed successfully", "id", notification.ID, "title", notification.Title)
	return &domainNotification, nil
}

// GetNotifications retrieves notifications for a user
func (s *notificationService) GetNotifications(ctx context.Context, userID int, filter domain.NotificationFilter, limit, offset int) ([]domain.Notification, *domain.NotificationStats, error) {
	// Create ports filter
	portsFilter := ports.NotificationFilter{
		UserID:      &userID,
		Type:        filter.Type,
		Priority:    filter.Priority,
		IsRead:      filter.IsRead,
		IsDismissed: filter.IsDismissed,
		Limit:       limit,
		Offset:      offset,
	}

	// Get notifications
	notifications, err := s.db.GetNotifications(ctx, portsFilter)
	if err != nil {
		return nil, nil, errors.WithDetails(errors.ErrDatabaseQuery, "failed to get notifications: %v", err)
	}

	// Get stats
	stats, err := s.db.GetNotificationStats(ctx, userID)
	if err != nil {
		log.Warn("Failed to get notification stats", "error", err)
	}

	// Convert to domain models
	domainNotifications := make([]domain.Notification, len(notifications))
	for i, notification := range notifications {
		domainNotifications[i] = s.convertNotificationToDomain(notification)
	}

	var domainStats *domain.NotificationStats
	if stats != nil {
		domainStats = &domain.NotificationStats{
			TotalCount:       stats.TotalCount,
			UnreadCount:      stats.UnreadCount,
			UndismissedCount: stats.UndismissedCount,
		}
	}

	return domainNotifications, domainStats, nil
}

// GetAllNotifications retrieves notifications for all users
func (s *notificationService) GetAllNotifications(ctx context.Context, filter domain.NotificationFilter, limit, offset int) ([]domain.Notification, int, error) {
	// Create ports filter
	portsFilter := ports.NotificationFilter{
		Type:        filter.Type,
		Priority:    filter.Priority,
		IsRead:      filter.IsRead,
		IsDismissed: filter.IsDismissed,
		Limit:       limit,
		Offset:      offset,
	}

	// Get notifications
	notifications, err := s.db.GetAllNotifications(ctx, portsFilter)
	if err != nil {
		return nil, 0, errors.WithDetails(errors.ErrDatabaseQuery, "failed to get notifications: %v", err)
	}

	// Convert to domain models
	domainNotifications := make([]domain.Notification, len(notifications))
	for i, notification := range notifications {
		domainNotifications[i] = s.convertNotificationToDomain(notification)
	}

	// Get total count for pagination (without limit/offset)
	countFilter := ports.NotificationFilter{
		Type:        filter.Type,
		Priority:    filter.Priority,
		IsRead:      filter.IsRead,
		IsDismissed: filter.IsDismissed,
		// No limit/offset for count
	}
	allNotifications, err := s.db.GetAllNotifications(ctx, countFilter)
	if err != nil {
		log.Warn("Failed to get total notification count", "error", err)
		return domainNotifications, len(notifications), nil
	}

	return domainNotifications, len(allNotifications), nil
}

// GetAllNotificationsWithUserInfo retrieves notifications for all users with user information
func (s *notificationService) GetAllNotificationsWithUserInfo(ctx context.Context, filter domain.NotificationFilter, limit, offset int) ([]domain.AdminNotification, int, error) {
	// Create ports filter
	portsFilter := ports.NotificationFilter{
		Type:        filter.Type,
		Priority:    filter.Priority,
		IsRead:      filter.IsRead,
		IsDismissed: filter.IsDismissed,
		Limit:       limit,
		Offset:      offset,
	}

	// Get notifications with user info
	notifications, err := s.db.GetAllNotificationsWithUserInfo(ctx, portsFilter)
	if err != nil {
		return nil, 0, errors.WithDetails(errors.ErrDatabaseQuery, "failed to get admin notifications: %v", err)
	}

	// Convert to domain models
	domainNotifications := make([]domain.AdminNotification, len(notifications))
	for i, notification := range notifications {
		domainNotifications[i] = s.convertAdminNotificationToDomain(notification)
	}

	// Get total count for pagination (without limit/offset)
	countFilter := ports.NotificationFilter{
		Type:        filter.Type,
		Priority:    filter.Priority,
		IsRead:      filter.IsRead,
		IsDismissed: filter.IsDismissed,
		// No limit/offset for count
	}
	allNotifications, err := s.db.GetAllNotificationsWithUserInfo(ctx, countFilter)
	if err != nil {
		log.Warn("Failed to get total admin notification count", "error", err)
		return domainNotifications, len(notifications), nil
	}

	return domainNotifications, len(allNotifications), nil
}

// MarkAsRead marks a notification as read
func (s *notificationService) MarkAsRead(ctx context.Context, notificationID int, userID int) error {
	err := s.db.MarkAsRead(ctx, notificationID, userID)
	if err != nil {
		return errors.WithDetails(errors.ErrDatabaseUpdate, "failed to mark notification as read: %v", err)
	}

	// Broadcast updated stats to user
	go s.broadcastStatsToUser(ctx, userID)

	return nil
}

// MarkAsDismissed marks a notification as dismissed
func (s *notificationService) MarkAsDismissed(ctx context.Context, notificationID int, userID int) error {
	err := s.db.MarkAsDismissed(ctx, notificationID, userID)
	if err != nil {
		return errors.WithDetails(errors.ErrDatabaseUpdate, "failed to mark notification as dismissed: %v", err)
	}

	// Broadcast updated stats to user
	go s.broadcastStatsToUser(ctx, userID)

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (s *notificationService) MarkAllAsRead(ctx context.Context, userID int) error {
	err := s.db.MarkAllAsRead(ctx, userID)
	if err != nil {
		return errors.WithDetails(errors.ErrDatabaseUpdate, "failed to mark all notifications as read: %v", err)
	}

	// Broadcast updated stats to user
	go s.broadcastStatsToUser(ctx, userID)

	return nil
}

// MarkAllAsDismissed marks all notifications as dismissed for a user
func (s *notificationService) MarkAllAsDismissed(ctx context.Context, userID int) error {
	err := s.db.MarkAllAsDismissed(ctx, userID)
	if err != nil {
		return errors.WithDetails(errors.ErrDatabaseUpdate, "failed to mark all notifications as dismissed: %v", err)
	}

	// Broadcast updated stats to user
	go s.broadcastStatsToUser(ctx, userID)

	return nil
}

// DeleteNotification deletes a notification for a user
func (s *notificationService) DeleteNotification(ctx context.Context, notificationID int, userID int) error {
	err := s.db.DeleteNotification(ctx, notificationID, userID)
	if err != nil {
		return errors.WithDetails(errors.ErrDatabaseDelete, "failed to delete notification: %v", err)
	}

	// Broadcast updated stats to user
	go s.broadcastStatsToUser(ctx, userID)

	return nil
}

// GetNotificationStats gets notification statistics for a user
func (s *notificationService) GetNotificationStats(ctx context.Context, userID int) (*domain.NotificationStats, error) {
	stats, err := s.db.GetNotificationStats(ctx, userID)
	if err != nil {
		return nil, errors.WithDetails(errors.ErrDatabaseQuery, "failed to get notification stats: %v", err)
	}

	return &domain.NotificationStats{
		TotalCount:       stats.TotalCount,
		UnreadCount:      stats.UnreadCount,
		UndismissedCount: stats.UndismissedCount,
	}, nil
}

// GetNotificationPreferences gets notification preferences for a user
func (s *notificationService) GetNotificationPreferences(ctx context.Context, userID int) (*domain.NotificationPreferences, error) {
	prefs, err := s.db.GetNotificationPreferences(ctx, userID)
	if err != nil {
		return nil, errors.WithDetails(errors.ErrDatabaseQuery, "failed to get notification preferences: %v", err)
	}

	return &domain.NotificationPreferences{
		EmailEnabled: prefs.EmailEnabled,
		PushEnabled:  prefs.PushEnabled,
		UpdatedAt:    prefs.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateNotificationPreferences updates notification preferences for a user
func (s *notificationService) UpdateNotificationPreferences(ctx context.Context, userID int, preferences domain.NotificationPreferences) error {
	portsPrefs := ports.NotificationPreferences{
		EmailEnabled: preferences.EmailEnabled,
		PushEnabled:  preferences.PushEnabled,
	}

	err := s.db.UpdateNotificationPreferences(ctx, userID, portsPrefs)
	if err != nil {
		return errors.WithDetails(errors.ErrDatabaseUpdate, "failed to update notification preferences: %v", err)
	}

	return nil
}

// RegisterWebSocketConnection registers a new WebSocket connection for a user
func (s *notificationService) RegisterWebSocketConnection(userID int, conn *websocket.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.connections[userID] = append(s.connections[userID], conn)
	log.Info(fmt.Sprintf("WebSocket connection registered: userID=%d, total_connections=%d", userID, len(s.connections[userID])))
}

// UnregisterWebSocketConnection removes a WebSocket connection for a user
func (s *notificationService) UnregisterWebSocketConnection(userID int, conn *websocket.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	connections := s.connections[userID]
	for i, existingConn := range connections {
		if existingConn == conn {
			// Remove connection from slice
			s.connections[userID] = append(connections[:i], connections[i+1:]...)
			break
		}
	}

	// Clean up empty connection lists
	if len(s.connections[userID]) == 0 {
		delete(s.connections, userID)
	}

	log.Info(fmt.Sprintf("WebSocket connection unregistered: userID=%d, remaining_connections=%d", userID, len(s.connections[userID])))
}

// BroadcastToUser sends a message to all WebSocket connections for a user
func (s *notificationService) BroadcastToUser(userID int, message domain.WebSocketMessage) {
	s.mutex.RLock()
	connections := s.connections[userID]
	s.mutex.RUnlock()

	if len(connections) == 0 {
		return
	}

	// Create a copy of connections to avoid holding the lock during message sending
	connsCopy := make([]*websocket.Conn, len(connections))
	copy(connsCopy, connections)

	for _, conn := range connsCopy {
		if err := conn.WriteJSON(message); err != nil {
			log.Warn("Failed to send WebSocket message", "userID", userID, "error", err)
			// Remove failed connections
			s.UnregisterWebSocketConnection(userID, conn)
		}
	}
}

// BroadcastNotificationToUser sends a notification to a user via WebSocket
func (s *notificationService) BroadcastNotificationToUser(userID int, notification *domain.Notification) {
	message := domain.WebSocketMessage{
		Type: "notification",
		Data: domain.NotificationWebSocketData{
			Notification: notification,
		},
	}

	s.BroadcastToUser(userID, message)
}

// broadcastStatsToUser sends updated stats to a user via WebSocket
func (s *notificationService) broadcastStatsToUser(ctx context.Context, userID int) {
	stats, err := s.GetNotificationStats(ctx, userID)
	if err != nil {
		log.Warn("Failed to get stats for WebSocket broadcast", "userID", userID, "error", err)
		return
	}

	message := domain.WebSocketMessage{
		Type: "stats",
		Data: domain.NotificationWebSocketData{
			Stats: stats,
		},
	}

	s.BroadcastToUser(userID, message)
}

// CleanupExpiredNotifications removes expired notifications
func (s *notificationService) CleanupExpiredNotifications(ctx context.Context) error {
	return s.db.CleanupExpiredNotifications(ctx)
}

// Close cleans up the service
func (s *notificationService) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Close all WebSocket connections
	for userID, connections := range s.connections {
		for _, conn := range connections {
			conn.Close()
		}
		delete(s.connections, userID)
	}
}

// convertNotificationToDomain converts a ports notification to a domain notification
func (s *notificationService) convertNotificationToDomain(notification ports.Notification) domain.Notification {
	domainNotification := domain.Notification{
		ID:          notification.ID,
		Title:       notification.Title,
		Body:        notification.Body,
		Link:        notification.Link,
		Type:        notification.Type,
		Priority:    notification.Priority,
		IsRead:      notification.IsRead,
		IsDismissed: notification.IsDismissed,
		CreatedAt:   notification.CreatedAt.Format(time.RFC3339),
	}

	if notification.ExpiresAt != nil {
		expiresAt := notification.ExpiresAt.Format(time.RFC3339)
		domainNotification.ExpiresAt = &expiresAt
	}

	if notification.ReadAt != nil {
		readAt := notification.ReadAt.Format(time.RFC3339)
		domainNotification.ReadAt = &readAt
	}

	if notification.DismissedAt != nil {
		dismissedAt := notification.DismissedAt.Format(time.RFC3339)
		domainNotification.DismissedAt = &dismissedAt
	}

	return domainNotification
}

// convertAdminNotificationToDomain converts a ports admin notification to a domain admin notification
func (s *notificationService) convertAdminNotificationToDomain(notification ports.AdminNotification) domain.AdminNotification {
	domainNotification := domain.AdminNotification{
		ID:          notification.ID,
		Title:       notification.Title,
		Body:        notification.Body,
		Link:        notification.Link,
		Type:        notification.Type,
		Priority:    notification.Priority,
		IsRead:      notification.IsRead,
		IsDismissed: notification.IsDismissed,
		CreatedAt:   notification.CreatedAt.Format(time.RFC3339),
	}

	if notification.ExpiresAt != nil {
		expiresAt := notification.ExpiresAt.Format(time.RFC3339)
		domainNotification.ExpiresAt = &expiresAt
	}

	if notification.ReadAt != nil {
		readAt := notification.ReadAt.Format(time.RFC3339)
		domainNotification.ReadAt = &readAt
	}

	if notification.DismissedAt != nil {
		dismissedAt := notification.DismissedAt.Format(time.RFC3339)
		domainNotification.DismissedAt = &dismissedAt
	}

	// Convert user information
	if notification.User != nil {
		domainUser := &domain.AdminNotificationUser{
			ID:        notification.User.ID,
			UserUID:   notification.User.UserUID,
			Email:     notification.User.Email,
			FirstName: notification.User.FirstName,
			LastName:  notification.User.LastName,
		}

		// Build full name
		fullName := ""
		if notification.User.FirstName != nil && notification.User.LastName != nil {
			fullName = *notification.User.FirstName + " " + *notification.User.LastName
		} else if notification.User.FirstName != nil {
			fullName = *notification.User.FirstName
		} else if notification.User.LastName != nil {
			fullName = *notification.User.LastName
		} else if notification.User.Email != nil {
			fullName = *notification.User.Email // fallback to email
		} else {
			fullName = "Unknown User" // final fallback
		}
		domainUser.FullName = fullName

		domainNotification.User = domainUser
	}

	return domainNotification
}

// AdminDeleteNotification deletes any notification (admin only, no user restriction)
func (s *notificationService) AdminDeleteNotification(ctx context.Context, notificationID int) error {
	err := s.db.AdminDeleteNotification(ctx, notificationID)
	if err != nil {
		return errors.WithDetails(errors.ErrDatabaseDelete, "failed to delete notification (admin): %v", err)
	}

	return nil
}

// AdminUpdateNotification updates any notification (admin only, no user restriction)
func (s *notificationService) AdminUpdateNotification(ctx context.Context, notificationID int, req domain.UpdateNotificationRequest) (*domain.Notification, error) {
	// Set defaults
	notificationType := req.Type
	if notificationType == "" {
		notificationType = "info"
	}

	priority := req.Priority
	if priority == "" {
		priority = "normal"
	}

	// Parse expires_at
	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return nil, errors.WithDetails(errors.ErrInvalidInput, "invalid expiresAt format: %v", err)
		}
		expiresAt = &parsed
	}

	// Create update struct
	update := ports.UpdateNotification{
		Title:     req.Title,
		Body:      req.Body,
		Link:      req.Link,
		Type:      notificationType,
		Priority:  priority,
		ExpiresAt: expiresAt,
	}

	// Update in database
	notification, err := s.db.AdminUpdateNotification(ctx, notificationID, update)
	if err != nil {
		return nil, errors.WithDetails(errors.ErrDatabaseUpdate, "failed to update notification (admin): %v", err)
	}

	// Convert to domain model
	domainNotification := s.convertNotificationToDomain(*notification)

	return &domainNotification, nil
}
