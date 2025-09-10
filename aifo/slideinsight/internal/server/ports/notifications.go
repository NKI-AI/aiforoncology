// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package ports

import (
	"context"
	"time"
)

// Notification represents a notification in the system
type Notification struct {
	ID          int        `json:"id" db:"id"`
	UserID      int        `json:"user_id" db:"user_id"`
	Title       string     `json:"title" db:"title"`
	Body        string     `json:"body" db:"body"`
	Link        *string    `json:"link" db:"link"`
	Type        string     `json:"type" db:"type"`         // info, success, warning, error
	Priority    string     `json:"priority" db:"priority"` // low, normal, high, urgent
	IsRead      bool       `json:"is_read" db:"is_read"`
	IsDismissed bool       `json:"is_dismissed" db:"is_dismissed"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ReadAt      *time.Time `json:"read_at" db:"read_at"`
	DismissedAt *time.Time `json:"dismissed_at" db:"dismissed_at"`
}

// NewNotification represents the data needed to create a new notification
type NewNotification struct {
	UserID    int        `json:"user_id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Link      *string    `json:"link"`
	Type      string     `json:"type"`     // info, success, warning, error
	Priority  string     `json:"priority"` // low, normal, high, urgent
	ExpiresAt *time.Time `json:"expires_at"`
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	EmailEnabled bool      `json:"email_enabled" db:"email_enabled"`
	PushEnabled  bool      `json:"push_enabled" db:"push_enabled"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// NotificationFilter represents filters for querying notifications
type NotificationFilter struct {
	UserID      *int    `json:"user_id"`
	Type        *string `json:"type"`
	Priority    *string `json:"priority"`
	IsRead      *bool   `json:"is_read"`
	IsDismissed *bool   `json:"is_dismissed"`
	Limit       int     `json:"limit"`
	Offset      int     `json:"offset"`
}

// NotificationStats represents notification statistics for a user
type NotificationStats struct {
	TotalCount       int `json:"total_count"`
	UnreadCount      int `json:"unread_count"`
	UndismissedCount int `json:"undismissed_count"`
}

// AdminNotification represents a notification with user information for admin views
type AdminNotification struct {
	ID          int                    `json:"id" db:"id"`
	UserID      int                    `json:"user_id" db:"user_id"`
	Title       string                 `json:"title" db:"title"`
	Body        string                 `json:"body" db:"body"`
	Link        *string                `json:"link" db:"link"`
	Type        string                 `json:"type" db:"type"`
	Priority    string                 `json:"priority" db:"priority"`
	IsRead      bool                   `json:"is_read" db:"is_read"`
	IsDismissed bool                   `json:"is_dismissed" db:"is_dismissed"`
	ExpiresAt   *time.Time             `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	ReadAt      *time.Time             `json:"read_at" db:"read_at"`
	DismissedAt *time.Time             `json:"dismissed_at" db:"dismissed_at"`
	User        *AdminNotificationUser `json:"user,omitempty"`
}

// AdminNotificationUser represents user information in admin notification views
type AdminNotificationUser struct {
	ID        int     `json:"id" db:"user_id"`
	UserUID   string  `json:"user_uid" db:"user_uid"`
	Email     *string `json:"email" db:"email"`
	FirstName *string `json:"first_name" db:"first_name"`
	LastName  *string `json:"last_name" db:"last_name"`
}

// UpdateNotification represents notification updates for admin operations
type UpdateNotification struct {
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Link      *string    `json:"link"`
	Type      string     `json:"type"`
	Priority  string     `json:"priority"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// NotificationsRepository defines the interface for notification database operations
type NotificationsRepository interface {
	// Create a new notification
	CreateNotification(ctx context.Context, notification NewNotification) (*Notification, error)

	// Get notification by ID
	GetNotificationByID(ctx context.Context, id int) (*Notification, error)

	// Get notifications for a user with filtering
	GetNotifications(ctx context.Context, filter NotificationFilter) ([]Notification, error)

	// Get all notifications (admin only)
	GetAllNotifications(ctx context.Context, filter NotificationFilter) ([]Notification, error)

	// Get all notifications with user information (admin only)
	GetAllNotificationsWithUserInfo(ctx context.Context, filter NotificationFilter) ([]AdminNotification, error)

	// Mark notification as read
	MarkAsRead(ctx context.Context, notificationID int, userID int) error

	// Mark notification as dismissed
	MarkAsDismissed(ctx context.Context, notificationID int, userID int) error

	// Mark all notifications as read for a user
	MarkAllAsRead(ctx context.Context, userID int) error

	// Mark all notifications as dismissed for a user
	MarkAllAsDismissed(ctx context.Context, userID int) error

	// Delete notification
	DeleteNotification(ctx context.Context, notificationID int, userID int) error

	// Admin operations (no user restrictions)
	AdminDeleteNotification(ctx context.Context, notificationID int) error
	AdminUpdateNotification(ctx context.Context, notificationID int, update UpdateNotification) (*Notification, error)

	// Get notification stats for a user
	GetNotificationStats(ctx context.Context, userID int) (*NotificationStats, error)

	// Clean up expired notifications
	CleanupExpiredNotifications(ctx context.Context) error

	// Notification preferences
	GetNotificationPreferences(ctx context.Context, userID int) (*NotificationPreferences, error)
	UpdateNotificationPreferences(ctx context.Context, userID int, preferences NotificationPreferences) error
}
