// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package schema

import "database/sql"

// UsersSchema handles user-related schema creation
type UsersSchema struct{}

// CreateTables creates user-related tables
func (s *UsersSchema) CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
		-- Users table
		CREATE TABLE IF NOT EXISTS users (
			id                    INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id             INTEGER NOT NULL,
			short_uid             TEXT    NOT NULL UNIQUE,
			email                 TEXT    NOT NULL,
			first_name            TEXT,
			last_name             TEXT,
			password              TEXT    NOT NULL,
			must_reset_password   BOOLEAN DEFAULT FALSE,
			is_active             BOOLEAN DEFAULT TRUE,
			email_verified        BOOLEAN DEFAULT FALSE,
			password_changed_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deactivated_at        TIMESTAMP NULL,
			deactivated_by        INTEGER NULL,
			created_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (tenant_id, email),
			FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
		);

		-- Password history table for preventing password reuse (enhanced for 6-month tracking)
		CREATE TABLE IF NOT EXISTS password_history (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id      INTEGER NOT NULL,
			password_hash TEXT   NOT NULL,
			created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at   TIMESTAMP NOT NULL, -- When this password can be reused (6 months from creation)
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);

		-- Password reset tokens
		CREATE TABLE IF NOT EXISTS password_reset_tokens (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id    INTEGER NOT NULL,
			token      TEXT    NOT NULL UNIQUE,
			expires_at TIMESTAMP NOT NULL,
			used       BOOLEAN DEFAULT FALSE,
			used_at    TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);

		-- Email verification tokens for user registration
		CREATE TABLE IF NOT EXISTS email_verification_tokens (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id    INTEGER NOT NULL,
			token      TEXT    NOT NULL UNIQUE,
			expires_at TIMESTAMP NOT NULL,
			used       BOOLEAN DEFAULT FALSE,
			used_at    TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);

		-- Authentication attempts for rate limiting and security monitoring
		CREATE TABLE IF NOT EXISTS auth_attempts (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			ip_address   TEXT    NOT NULL,
			email        TEXT,
			success      BOOLEAN NOT NULL,
			fail_reason  TEXT,
			attempted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

// CreateTriggers creates user-related triggers
func (s *UsersSchema) CreateTriggers(db *sql.DB) error {
	_, err := db.Exec(`
		-- Trigger for users table
		CREATE TRIGGER IF NOT EXISTS users_updated_at 
		AFTER UPDATE ON users 
		FOR EACH ROW 
		BEGIN
			UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

		-- Trigger to update password_changed_at when password is updated
		CREATE TRIGGER IF NOT EXISTS users_password_changed_at 
		AFTER UPDATE OF password ON users 
		FOR EACH ROW 
		WHEN OLD.password != NEW.password
		BEGIN
			UPDATE users SET password_changed_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

		-- Trigger to automatically clean up expired password history entries
		CREATE TRIGGER IF NOT EXISTS password_history_cleanup 
		AFTER INSERT ON password_history 
		FOR EACH ROW 
		BEGIN
			DELETE FROM password_history 
			WHERE user_id = NEW.user_id 
			AND expires_at < datetime('now');
		END;
	`)
	return err
}

// CreateIndexes creates user-related indexes
func (s *UsersSchema) CreateIndexes(db *sql.DB) error {
	_, err := db.Exec(`
		-- Indexes for password reset tokens
		CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token ON password_reset_tokens(token);
		CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
		CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);

		-- Indexes for email verification tokens
		CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_token ON email_verification_tokens(token);
		CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);
		CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_expires_at ON email_verification_tokens(expires_at);

		-- Indexes for auth attempts (for rate limiting)
		CREATE INDEX IF NOT EXISTS idx_auth_attempts_ip_attempted_at ON auth_attempts(ip_address, attempted_at);
		CREATE INDEX IF NOT EXISTS idx_auth_attempts_email_attempted_at ON auth_attempts(email, attempted_at);

		-- Indexes for password history (enhanced for 6-month lookups)
		CREATE INDEX IF NOT EXISTS idx_password_history_user_id ON password_history(user_id);
		CREATE INDEX IF NOT EXISTS idx_password_history_created_at ON password_history(created_at);
		CREATE INDEX IF NOT EXISTS idx_password_history_expires_at ON password_history(expires_at);
		CREATE INDEX IF NOT EXISTS idx_password_history_user_expires ON password_history(user_id, expires_at);

		-- Additional indexes for users table
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`)
	return err
}
