// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package users

import (
	"context"
	"database/sql"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
)

// AuthService handles authentication-related user operations
type AuthService struct {
	db *sql.DB
}

// NewAuthService creates a new auth service instance
func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

// DeactivateUser marks a user as inactive by setting is_active to false
func (s *AuthService) DeactivateUser(ctx context.Context, email string) error {
	result, err := s.db.Exec("UPDATE users SET is_active = FALSE, deactivated_at = CURRENT_TIMESTAMP WHERE email = ?", email)
	if err != nil {
		return errors.NewDatabaseUpdateError("user deactivation", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.NewUserNotFoundByEmailError(email)
	}

	return nil
}

// ActivateUser marks a user as active by setting is_active to true and clearing deactivation timestamp
func (s *AuthService) ActivateUser(ctx context.Context, email string) error {
	result, err := s.db.Exec("UPDATE users SET is_active = TRUE, deactivated_at = NULL, deactivated_by = NULL WHERE email = ?", email)
	if err != nil {
		return errors.NewDatabaseUpdateError("user activation", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.NewUserNotFoundByEmailError(email)
	}

	return nil
}

// Password History Methods

// AddPasswordToHistory adds a password hash to the user's password history
func (s *AuthService) AddPasswordToHistory(ctx context.Context, userID int, passwordHash string) error {
	// Calculate expiry date (6 months from now)
	expiresAt := time.Now().AddDate(0, 6, 0)

	_, err := s.db.Exec("INSERT INTO password_history (user_id, password_hash, expires_at) VALUES (?, ?, ?)",
		userID, passwordHash, expiresAt.Format(time.RFC3339))
	if err != nil {
		return errors.NewDatabaseInsertError("password history", err)
	}
	return nil
}

// GetPasswordHistory retrieves password history for a user within 6 months (using expires_at field)
func (s *AuthService) GetPasswordHistory(ctx context.Context, userID int, months int) ([]ports.PasswordHistory, error) {
	// Use the expires_at field to get only non-expired passwords
	rows, err := s.db.Query(`
		SELECT id, user_id, password_hash, created_at 
		FROM password_history 
		WHERE user_id = ? AND expires_at > datetime('now')
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("password history", err)
	}
	defer rows.Close()

	var history []ports.PasswordHistory
	for rows.Next() {
		var h ports.PasswordHistory
		if err := rows.Scan(&h.ID, &h.UserID, &h.PasswordHash, &h.CreatedAt); err != nil {
			return nil, errors.NewDatabaseScanError("password history", err)
		}
		history = append(history, h)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("password history", err)
	}

	return history, nil
}

// CleanupOldPasswordHistory removes expired password history entries
func (s *AuthService) CleanupOldPasswordHistory(ctx context.Context, userID int, keepCount int) error {
	// Clean up expired entries first
	_, err := s.db.Exec(`
		DELETE FROM password_history 
		WHERE user_id = ? AND expires_at < datetime('now')
	`, userID)
	if err != nil {
		return errors.NewDatabaseDeleteError("expired password history", err)
	}

	// Then clean up excess entries if needed (keep only the most recent ones)
	_, err = s.db.Exec(`
		DELETE FROM password_history 
		WHERE user_id = ? AND id NOT IN (
			SELECT id FROM password_history 
			WHERE user_id = ? 
			ORDER BY created_at DESC 
			LIMIT ?
		)
	`, userID, userID, keepCount)
	if err != nil {
		return errors.NewDatabaseDeleteError("old password history", err)
	}
	return nil
}

// Authentication Attempt Methods

// RecordAuthAttempt records an authentication attempt
func (s *AuthService) RecordAuthAttempt(ctx context.Context, ipAddress, email string, success bool, failReason string) error {
	_, err := s.db.Exec("INSERT INTO auth_attempts (ip_address, email, success, fail_reason) VALUES (?, ?, ?, ?)", ipAddress, email, success, failReason)
	if err != nil {
		return errors.NewDatabaseInsertError("auth attempt", err)
	}
	return nil
}

// GetRecentAuthAttempts retrieves recent authentication attempts for an IP address
func (s *AuthService) GetRecentAuthAttempts(ctx context.Context, ipAddress string, since time.Time) ([]ports.AuthAttempt, error) {
	query := `
		SELECT id, ip_address, email, success, fail_reason, attempted_at 
		FROM auth_attempts 
		WHERE ip_address = ? AND attempted_at > ?
		ORDER BY attempted_at DESC`

	rows, err := s.db.Query(query, ipAddress, since)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("auth attempts", err)
	}
	defer rows.Close()

	var attempts []ports.AuthAttempt
	for rows.Next() {
		var attempt ports.AuthAttempt
		if err := rows.Scan(&attempt.ID, &attempt.IPAddress, &attempt.Email, &attempt.Success, &attempt.FailReason, &attempt.AttemptedAt); err != nil {
			return nil, errors.NewDatabaseScanError("auth attempt", err)
		}
		attempts = append(attempts, attempt)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("auth attempts", err)
	}

	return attempts, nil
}

// GetRecentAuthAttemptsForUser retrieves recent authentication attempts for an email
func (s *AuthService) GetRecentAuthAttemptsForUser(ctx context.Context, email string, since time.Time) ([]ports.AuthAttempt, error) {
	query := `
		SELECT id, ip_address, email, success, fail_reason, attempted_at 
		FROM auth_attempts 
		WHERE email = ? AND attempted_at > ?
		ORDER BY attempted_at DESC`

	rows, err := s.db.Query(query, email, since)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("auth attempts for user", err)
	}
	defer rows.Close()

	var attempts []ports.AuthAttempt
	for rows.Next() {
		var attempt ports.AuthAttempt
		if err := rows.Scan(&attempt.ID, &attempt.IPAddress, &attempt.Email, &attempt.Success, &attempt.FailReason, &attempt.AttemptedAt); err != nil {
			return nil, errors.NewDatabaseScanError("auth attempt", err)
		}
		attempts = append(attempts, attempt)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("auth attempts", err)
	}

	return attempts, nil
}

// CleanupOldAuthAttempts removes old authentication attempts
func (s *AuthService) CleanupOldAuthAttempts(ctx context.Context, olderThan time.Time) error {
	_, err := s.db.Exec("DELETE FROM auth_attempts WHERE attempted_at < ?", olderThan)
	if err != nil {
		return errors.NewDatabaseDeleteError("old auth attempts", err)
	}
	return nil
}
