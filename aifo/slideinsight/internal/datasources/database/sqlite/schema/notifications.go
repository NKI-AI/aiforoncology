// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package schema

import "database/sql"

// NotificationsSchema handles notification-related schema creation
type NotificationsSchema struct{}

// CreateTables creates notification-related tables
func (s *NotificationsSchema) CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
		-- Notifications table
		CREATE TABLE IF NOT EXISTS notifications (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id         INTEGER NOT NULL,
			title           TEXT NOT NULL,
			body            TEXT NOT NULL,
			link            TEXT,
			type            TEXT NOT NULL DEFAULT 'info', -- info, success, warning, error
			priority        TEXT NOT NULL DEFAULT 'normal', -- low, normal, high, urgent
			is_read         BOOLEAN DEFAULT FALSE,
			is_dismissed    BOOLEAN DEFAULT FALSE,
			expires_at      TIMESTAMP NULL, -- NULL means no expiration
			created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			read_at         TIMESTAMP NULL,
			dismissed_at    TIMESTAMP NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);

		-- Notification preferences table (for future use)
		CREATE TABLE IF NOT EXISTS notification_preferences (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id         INTEGER NOT NULL UNIQUE,
			email_enabled   BOOLEAN DEFAULT TRUE,
			push_enabled    BOOLEAN DEFAULT TRUE,
			created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
	`)
	return err
}

// CreateTriggers creates notification-related triggers
func (s *NotificationsSchema) CreateTriggers(db *sql.DB) error {
	_, err := db.Exec(`
		-- Trigger for notification_preferences table
		CREATE TRIGGER IF NOT EXISTS notification_preferences_updated_at 
		AFTER UPDATE ON notification_preferences 
		FOR EACH ROW 
		BEGIN
			UPDATE notification_preferences SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

		-- Trigger to automatically clean up expired notifications
		CREATE TRIGGER IF NOT EXISTS notifications_cleanup 
		AFTER INSERT ON notifications 
		FOR EACH ROW 
		BEGIN
			DELETE FROM notifications 
			WHERE expires_at IS NOT NULL 
			AND expires_at < datetime('now');
		END;
	`)
	return err
}

// CreateIndexes creates notification-related indexes
func (s *NotificationsSchema) CreateIndexes(db *sql.DB) error {
	_, err := db.Exec(`
		-- Indexes for notifications
		CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
		CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
		CREATE INDEX IF NOT EXISTS idx_notifications_is_dismissed ON notifications(is_dismissed);
		CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
		CREATE INDEX IF NOT EXISTS idx_notifications_expires_at ON notifications(expires_at);
		CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
		CREATE INDEX IF NOT EXISTS idx_notifications_priority ON notifications(priority);
		
		-- Composite indexes for common queries
		CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id, is_read, is_dismissed);
		CREATE INDEX IF NOT EXISTS idx_notifications_user_created ON notifications(user_id, created_at);
	`)
	return err
}
