// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package domain

// Notification represents a notification for API responses
type Notification struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Body        string  `json:"body"`
	Link        *string `json:"link,omitempty"`
	Type        string  `json:"type"`     // info, success, warning, error
	Priority    string  `json:"priority"` // low, normal, high, urgent
	IsRead      bool    `json:"isRead"`
	IsDismissed bool    `json:"isDismissed"`
	ExpiresAt   *string `json:"expiresAt,omitempty"`
	CreatedAt   string  `json:"createdAt"`
	ReadAt      *string `json:"readAt,omitempty"`
	DismissedAt *string `json:"dismissedAt,omitempty"`
}

// NewNotificationRequest represents the request payload to create a new notification
type NewNotificationRequest struct {
	UserUID   *string `json:"userUid,omitempty"` // Optional: if not provided, send to current user
	Email     *string `json:"email,omitempty"`   // Alternative to UserUID
	Title     string  `json:"title" validate:"required"`
	Body      string  `json:"body" validate:"required"`
	Link      *string `json:"link,omitempty"`
	Type      string  `json:"type,omitempty"`      // info, success, warning, error (default: info)
	Priority  string  `json:"priority,omitempty"`  // low, normal, high, urgent (default: normal)
	ExpiresAt *string `json:"expiresAt,omitempty"` // ISO 8601 timestamp
}

// UpdateNotificationRequest represents the request payload to update a notification (admin only)
type UpdateNotificationRequest struct {
	Title     string  `json:"title" validate:"required"`
	Body      string  `json:"body" validate:"required"`
	Link      *string `json:"link,omitempty"`
	Type      string  `json:"type,omitempty"`      // info, success, warning, error
	Priority  string  `json:"priority,omitempty"`  // low, normal, high, urgent
	ExpiresAt *string `json:"expiresAt,omitempty"` // ISO 8601 timestamp
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	EmailEnabled bool   `json:"emailEnabled"`
	PushEnabled  bool   `json:"pushEnabled"`
	UpdatedAt    string `json:"updatedAt"`
}

// NotificationStats represents notification statistics for a user
type NotificationStats struct {
	TotalCount       int `json:"totalCount"`
	UnreadCount      int `json:"unreadCount"`
	UndismissedCount int `json:"undismissedCount"`
}

// NotificationsResponse represents a response containing a list of notifications with pagination
type NotificationsResponse struct {
	Notifications []Notification     `json:"notifications"`
	Pagination    PaginationInfo     `json:"pagination"`
	Stats         *NotificationStats `json:"stats,omitempty"`
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type string      `json:"type"` // "notification", "stats", "error"
	Data interface{} `json:"data"`
}

// NotificationWebSocketData represents notification data sent over WebSocket
type NotificationWebSocketData struct {
	Notification *Notification      `json:"notification,omitempty"`
	Stats        *NotificationStats `json:"stats,omitempty"`
}

// NotificationFilter represents filters for querying notifications
type NotificationFilter struct {
	Type        *string `json:"type,omitempty"`
	Priority    *string `json:"priority,omitempty"`
	IsRead      *bool   `json:"isRead,omitempty"`
	IsDismissed *bool   `json:"isDismissed,omitempty"`
}

// AdminNotification represents a notification with user information for admin views
type AdminNotification struct {
	ID          int                    `json:"id"`
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	Link        *string                `json:"link,omitempty"`
	Type        string                 `json:"type"`     // info, success, warning, error
	Priority    string                 `json:"priority"` // low, normal, high, urgent
	IsRead      bool                   `json:"isRead"`
	IsDismissed bool                   `json:"isDismissed"`
	ExpiresAt   *string                `json:"expiresAt,omitempty"`
	CreatedAt   string                 `json:"createdAt"`
	ReadAt      *string                `json:"readAt,omitempty"`
	DismissedAt *string                `json:"dismissedAt,omitempty"`
	User        *AdminNotificationUser `json:"user,omitempty"`
}

// AdminNotificationUser represents user information in admin notification views
type AdminNotificationUser struct {
	ID        int     `json:"id"`
	UserUID   string  `json:"userUid"`
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
	FullName  string  `json:"fullName"` // Computed field
}

// AdminNotificationsResponse represents a response containing admin notifications with user info
type AdminNotificationsResponse struct {
	Notifications []AdminNotification `json:"notifications"`
	Pagination    PaginationInfo      `json:"pagination"`
	Stats         *NotificationStats  `json:"stats,omitempty"`
}
