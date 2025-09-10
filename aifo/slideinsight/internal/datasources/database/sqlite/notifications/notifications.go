// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
)

// Adapter implements the NotificationsRepository interface using SQLite
type Adapter struct {
	db *sql.DB
}

// NewAdapter creates a new notifications adapter
func NewAdapter(db *sql.DB) *Adapter {
	return &Adapter{db: db}
}

// CreateNotification creates a new notification
func (a *Adapter) CreateNotification(ctx context.Context, notification ports.NewNotification) (*ports.Notification, error) {
	query := `
		INSERT INTO notifications (user_id, title, body, link, type, priority, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, user_id, title, body, link, type, priority, is_read, is_dismissed, expires_at, created_at, read_at, dismissed_at
	`

	var result ports.Notification
	err := a.db.QueryRowContext(ctx, query,
		notification.UserID,
		notification.Title,
		notification.Body,
		notification.Link,
		notification.Type,
		notification.Priority,
		notification.ExpiresAt,
	).Scan(
		&result.ID,
		&result.UserID,
		&result.Title,
		&result.Body,
		&result.Link,
		&result.Type,
		&result.Priority,
		&result.IsRead,
		&result.IsDismissed,
		&result.ExpiresAt,
		&result.CreatedAt,
		&result.ReadAt,
		&result.DismissedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return &result, nil
}

// GetNotificationByID retrieves a notification by ID
func (a *Adapter) GetNotificationByID(ctx context.Context, id int) (*ports.Notification, error) {
	query := `
		SELECT id, user_id, title, body, link, type, priority, is_read, is_dismissed, expires_at, created_at, read_at, dismissed_at
		FROM notifications
		WHERE id = ?
	`

	var notification ports.Notification
	err := a.db.QueryRowContext(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Title,
		&notification.Body,
		&notification.Link,
		&notification.Type,
		&notification.Priority,
		&notification.IsRead,
		&notification.IsDismissed,
		&notification.ExpiresAt,
		&notification.CreatedAt,
		&notification.ReadAt,
		&notification.DismissedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return &notification, nil
}

// GetNotifications retrieves notifications with filtering
func (a *Adapter) GetNotifications(ctx context.Context, filter ports.NotificationFilter) ([]ports.Notification, error) {
	var conditions []string
	var args []interface{}

	baseQuery := `
		SELECT id, user_id, title, body, link, type, priority, is_read, is_dismissed, expires_at, created_at, read_at, dismissed_at
		FROM notifications
	`

	if filter.UserID != nil {
		conditions = append(conditions, "user_id = ?")
		args = append(args, *filter.UserID)
	}

	if filter.Type != nil {
		conditions = append(conditions, "type = ?")
		args = append(args, *filter.Type)
	}

	if filter.Priority != nil {
		conditions = append(conditions, "priority = ?")
		args = append(args, *filter.Priority)
	}

	if filter.IsRead != nil {
		conditions = append(conditions, "is_read = ?")
		args = append(args, *filter.IsRead)
	}

	if filter.IsDismissed != nil {
		conditions = append(conditions, "is_dismissed = ?")
		args = append(args, *filter.IsDismissed)
	}

	// Add condition to exclude expired notifications
	conditions = append(conditions, "(expires_at IS NULL OR expires_at > datetime('now'))")

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []ports.Notification
	for rows.Next() {
		var notification ports.Notification
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Body,
			&notification.Link,
			&notification.Type,
			&notification.Priority,
			&notification.IsRead,
			&notification.IsDismissed,
			&notification.ExpiresAt,
			&notification.CreatedAt,
			&notification.ReadAt,
			&notification.DismissedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (a *Adapter) MarkAsRead(ctx context.Context, notificationID int, userID int) error {
	query := `
		UPDATE notifications 
		SET is_read = TRUE, read_at = CURRENT_TIMESTAMP 
		WHERE id = ? AND user_id = ? AND is_read = FALSE
	`

	result, err := a.db.ExecContext(ctx, query, notificationID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found or already read")
	}

	return nil
}

// MarkAsDismissed marks a notification as dismissed
func (a *Adapter) MarkAsDismissed(ctx context.Context, notificationID int, userID int) error {
	query := `
		UPDATE notifications 
		SET is_dismissed = TRUE, dismissed_at = CURRENT_TIMESTAMP 
		WHERE id = ? AND user_id = ? AND is_dismissed = FALSE
	`

	result, err := a.db.ExecContext(ctx, query, notificationID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as dismissed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found or already dismissed")
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (a *Adapter) MarkAllAsRead(ctx context.Context, userID int) error {
	query := `
		UPDATE notifications 
		SET is_read = TRUE, read_at = CURRENT_TIMESTAMP 
		WHERE user_id = ? AND is_read = FALSE
	`

	_, err := a.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// MarkAllAsDismissed marks all notifications as dismissed for a user
func (a *Adapter) MarkAllAsDismissed(ctx context.Context, userID int) error {
	query := `
		UPDATE notifications 
		SET is_dismissed = TRUE, dismissed_at = CURRENT_TIMESTAMP 
		WHERE user_id = ? AND is_dismissed = FALSE
	`

	_, err := a.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as dismissed: %w", err)
	}

	return nil
}

// DeleteNotification deletes a notification
func (a *Adapter) DeleteNotification(ctx context.Context, notificationID int, userID int) error {
	query := `DELETE FROM notifications WHERE id = ? AND user_id = ?`

	result, err := a.db.ExecContext(ctx, query, notificationID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// GetNotificationStats gets notification statistics for a user
func (a *Adapter) GetNotificationStats(ctx context.Context, userID int) (*ports.NotificationStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_count,
			COALESCE(SUM(CASE WHEN is_read = FALSE THEN 1 ELSE 0 END), 0) as unread_count,
			COALESCE(SUM(CASE WHEN is_dismissed = FALSE THEN 1 ELSE 0 END), 0) as undismissed_count
		FROM notifications 
		WHERE user_id = ? AND (expires_at IS NULL OR expires_at > datetime('now'))
	`

	var stats ports.NotificationStats
	err := a.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.TotalCount,
		&stats.UnreadCount,
		&stats.UndismissedCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification stats: %w", err)
	}

	return &stats, nil
}

// CleanupExpiredNotifications removes expired notifications
func (a *Adapter) CleanupExpiredNotifications(ctx context.Context) error {
	query := `DELETE FROM notifications WHERE expires_at IS NOT NULL AND expires_at <= datetime('now')`

	_, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired notifications: %w", err)
	}

	return nil
}

// GetNotificationPreferences gets notification preferences for a user
func (a *Adapter) GetNotificationPreferences(ctx context.Context, userID int) (*ports.NotificationPreferences, error) {
	query := `
		SELECT id, user_id, email_enabled, push_enabled, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = ?
	`

	var prefs ports.NotificationPreferences
	err := a.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.ID,
		&prefs.UserID,
		&prefs.EmailEnabled,
		&prefs.PushEnabled,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create default preferences if they don't exist
			return a.createDefaultNotificationPreferences(ctx, userID)
		}
		return nil, fmt.Errorf("failed to get notification preferences: %w", err)
	}

	return &prefs, nil
}

// createDefaultNotificationPreferences creates default notification preferences for a user
func (a *Adapter) createDefaultNotificationPreferences(ctx context.Context, userID int) (*ports.NotificationPreferences, error) {
	query := `
		INSERT INTO notification_preferences (user_id, email_enabled, push_enabled)
		VALUES (?, TRUE, TRUE)
		RETURNING id, user_id, email_enabled, push_enabled, created_at, updated_at
	`

	var prefs ports.NotificationPreferences
	err := a.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.ID,
		&prefs.UserID,
		&prefs.EmailEnabled,
		&prefs.PushEnabled,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create default notification preferences: %w", err)
	}

	return &prefs, nil
}

// UpdateNotificationPreferences updates notification preferences for a user
func (a *Adapter) UpdateNotificationPreferences(ctx context.Context, userID int, preferences ports.NotificationPreferences) error {
	query := `
		UPDATE notification_preferences 
		SET email_enabled = ?, push_enabled = ?, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
	`

	result, err := a.db.ExecContext(ctx, query, preferences.EmailEnabled, preferences.PushEnabled, userID)
	if err != nil {
		return fmt.Errorf("failed to update notification preferences: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Create preferences if they don't exist
		_, err := a.createDefaultNotificationPreferences(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to create notification preferences: %w", err)
		}
	}

	return nil
}

// GetAllNotifications retrieves all notifications (admin only)
func (a *Adapter) GetAllNotifications(ctx context.Context, filter ports.NotificationFilter) ([]ports.Notification, error) {
	var conditions []string
	var args []interface{}

	baseQuery := `
		SELECT n.id, n.user_id, n.title, n.body, n.link, n.type, n.priority, n.is_read, n.is_dismissed, n.expires_at, n.created_at, n.read_at, n.dismissed_at,
		       u.email, u.first_name, u.last_name
		FROM notifications n
		LEFT JOIN users u ON n.user_id = u.id
	`

	if filter.Type != nil {
		conditions = append(conditions, "n.type = ?")
		args = append(args, *filter.Type)
	}

	if filter.Priority != nil {
		conditions = append(conditions, "n.priority = ?")
		args = append(args, *filter.Priority)
	}

	if filter.IsRead != nil {
		conditions = append(conditions, "n.is_read = ?")
		args = append(args, *filter.IsRead)
	}

	if filter.IsDismissed != nil {
		conditions = append(conditions, "n.is_dismissed = ?")
		args = append(args, *filter.IsDismissed)
	}

	// Add condition to exclude expired notifications
	conditions = append(conditions, "(n.expires_at IS NULL OR n.expires_at > datetime('now'))")

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY n.created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query all notifications: %w", err)
	}
	defer rows.Close()

	var notifications []ports.Notification
	for rows.Next() {
		var notification ports.Notification
		var email, firstName, lastName sql.NullString

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Body,
			&notification.Link,
			&notification.Type,
			&notification.Priority,
			&notification.IsRead,
			&notification.IsDismissed,
			&notification.ExpiresAt,
			&notification.CreatedAt,
			&notification.ReadAt,
			&notification.DismissedAt,
			&email,
			&firstName,
			&lastName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		// Add user information to notification (we'll store this as a new field)
		// For now, we'll just include the notification as-is
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// GetAllNotificationsWithUserInfo retrieves all notifications with user information (admin only)
func (a *Adapter) GetAllNotificationsWithUserInfo(ctx context.Context, filter ports.NotificationFilter) ([]ports.AdminNotification, error) {
	var conditions []string
	var args []interface{}

	baseQuery := `
		SELECT n.id, n.user_id, n.title, n.body, n.link, n.type, n.priority, n.is_read, n.is_dismissed, 
		       n.expires_at, n.created_at, n.read_at, n.dismissed_at,
		       u.id as user_id, u.short_uid, u.email, u.first_name, u.last_name
		FROM notifications n
		LEFT JOIN users u ON n.user_id = u.id
	`

	if filter.Type != nil {
		conditions = append(conditions, "n.type = ?")
		args = append(args, *filter.Type)
	}

	if filter.Priority != nil {
		conditions = append(conditions, "n.priority = ?")
		args = append(args, *filter.Priority)
	}

	if filter.IsRead != nil {
		conditions = append(conditions, "n.is_read = ?")
		args = append(args, *filter.IsRead)
	}

	if filter.IsDismissed != nil {
		conditions = append(conditions, "n.is_dismissed = ?")
		args = append(args, *filter.IsDismissed)
	}

	// Add condition to exclude expired notifications
	conditions = append(conditions, "(n.expires_at IS NULL OR n.expires_at > datetime('now'))")

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY n.created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query admin notifications: %w", err)
	}
	defer rows.Close()

	var notifications []ports.AdminNotification
	for rows.Next() {
		var notification ports.AdminNotification
		var user ports.AdminNotificationUser
		var userID, shortUID sql.NullString
		var email, firstName, lastName sql.NullString

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Body,
			&notification.Link,
			&notification.Type,
			&notification.Priority,
			&notification.IsRead,
			&notification.IsDismissed,
			&notification.ExpiresAt,
			&notification.CreatedAt,
			&notification.ReadAt,
			&notification.DismissedAt,
			&userID,
			&shortUID,
			&email,
			&firstName,
			&lastName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin notification: %w", err)
		}

		// Add user information if available
		if userID.Valid && shortUID.Valid {
			user.ID = notification.UserID
			user.UserUID = shortUID.String
			if email.Valid {
				user.Email = &email.String
			}
			if firstName.Valid {
				user.FirstName = &firstName.String
			}
			if lastName.Valid {
				user.LastName = &lastName.String
			}
			notification.User = &user
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// AdminDeleteNotification deletes any notification (admin only, no user restriction)
func (a *Adapter) AdminDeleteNotification(ctx context.Context, notificationID int) error {
	query := `DELETE FROM notifications WHERE id = ?`

	result, err := a.db.ExecContext(ctx, query, notificationID)
	if err != nil {
		return fmt.Errorf("failed to delete notification (admin): %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// AdminUpdateNotification updates any notification (admin only, no user restriction)
func (a *Adapter) AdminUpdateNotification(ctx context.Context, notificationID int, update ports.UpdateNotification) (*ports.Notification, error) {
	query := `
		UPDATE notifications 
		SET title = ?, body = ?, link = ?, type = ?, priority = ?, expires_at = ?
		WHERE id = ?
		RETURNING id, user_id, title, body, link, type, priority, is_read, is_dismissed, expires_at, created_at, read_at, dismissed_at
	`

	var result ports.Notification
	err := a.db.QueryRowContext(ctx, query,
		update.Title,
		update.Body,
		update.Link,
		update.Type,
		update.Priority,
		update.ExpiresAt,
		notificationID,
	).Scan(
		&result.ID,
		&result.UserID,
		&result.Title,
		&result.Body,
		&result.Link,
		&result.Type,
		&result.Priority,
		&result.IsRead,
		&result.IsDismissed,
		&result.ExpiresAt,
		&result.CreatedAt,
		&result.ReadAt,
		&result.DismissedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, fmt.Errorf("failed to update notification (admin): %w", err)
	}

	return &result, nil
}
