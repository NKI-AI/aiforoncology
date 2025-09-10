// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package sqlite

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
)

// CreateNotification proxies the notifications adapter call
func (db *DB) CreateNotification(ctx context.Context, notification ports.NewNotification) (*ports.Notification, error) {
	return db.notifications.CreateNotification(ctx, notification)
}

// GetNotificationByID proxies the notifications adapter call
func (db *DB) GetNotificationByID(ctx context.Context, id int) (*ports.Notification, error) {
	return db.notifications.GetNotificationByID(ctx, id)
}

// GetNotifications proxies the notifications adapter call
func (db *DB) GetNotifications(ctx context.Context, filter ports.NotificationFilter) ([]ports.Notification, error) {
	return db.notifications.GetNotifications(ctx, filter)
}

// GetAllNotifications proxies the notifications adapter call
func (db *DB) GetAllNotifications(ctx context.Context, filter ports.NotificationFilter) ([]ports.Notification, error) {
	return db.notifications.GetAllNotifications(ctx, filter)
}

// MarkAsRead proxies the notifications adapter call
func (db *DB) MarkAsRead(ctx context.Context, notificationID int, userID int) error {
	return db.notifications.MarkAsRead(ctx, notificationID, userID)
}

// MarkAsDismissed proxies the notifications adapter call
func (db *DB) MarkAsDismissed(ctx context.Context, notificationID int, userID int) error {
	return db.notifications.MarkAsDismissed(ctx, notificationID, userID)
}

// MarkAllAsRead proxies the notifications adapter call
func (db *DB) MarkAllAsRead(ctx context.Context, userID int) error {
	return db.notifications.MarkAllAsRead(ctx, userID)
}

// MarkAllAsDismissed proxies the notifications adapter call
func (db *DB) MarkAllAsDismissed(ctx context.Context, userID int) error {
	return db.notifications.MarkAllAsDismissed(ctx, userID)
}

// DeleteNotification proxies the notifications adapter call
func (db *DB) DeleteNotification(ctx context.Context, notificationID int, userID int) error {
	return db.notifications.DeleteNotification(ctx, notificationID, userID)
}

// GetNotificationStats proxies the notifications adapter call
func (db *DB) GetNotificationStats(ctx context.Context, userID int) (*ports.NotificationStats, error) {
	return db.notifications.GetNotificationStats(ctx, userID)
}

// CleanupExpiredNotifications proxies the notifications adapter call
func (db *DB) CleanupExpiredNotifications(ctx context.Context) error {
	return db.notifications.CleanupExpiredNotifications(ctx)
}

// GetNotificationPreferences proxies the notifications adapter call
func (db *DB) GetNotificationPreferences(ctx context.Context, userID int) (*ports.NotificationPreferences, error) {
	return db.notifications.GetNotificationPreferences(ctx, userID)
}

// UpdateNotificationPreferences proxies the notifications adapter call
func (db *DB) UpdateNotificationPreferences(ctx context.Context, userID int, preferences ports.NotificationPreferences) error {
	return db.notifications.UpdateNotificationPreferences(ctx, userID, preferences)
}

// GetAllNotificationsWithUserInfo proxies the notifications adapter call
func (db *DB) GetAllNotificationsWithUserInfo(ctx context.Context, filter ports.NotificationFilter) ([]ports.AdminNotification, error) {
	return db.notifications.GetAllNotificationsWithUserInfo(ctx, filter)
}

// AdminDeleteNotification proxies the notifications adapter call
func (db *DB) AdminDeleteNotification(ctx context.Context, notificationID int) error {
	return db.notifications.AdminDeleteNotification(ctx, notificationID)
}

// AdminUpdateNotification proxies the notifications adapter call
func (db *DB) AdminUpdateNotification(ctx context.Context, notificationID int, update ports.UpdateNotification) (*ports.Notification, error) {
	return db.notifications.AdminUpdateNotification(ctx, notificationID, update)
}
