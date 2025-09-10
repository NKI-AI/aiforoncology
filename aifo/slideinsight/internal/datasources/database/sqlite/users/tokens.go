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

// TokenService handles user token operations (password reset, email verification)
type TokenService struct {
	db *sql.DB
}

// NewTokenService creates a new token service instance
func NewTokenService(db *sql.DB) *TokenService {
	return &TokenService{db: db}
}

// Password Reset Token Methods

// CreatePasswordResetToken creates a new password reset token
func (s *TokenService) CreatePasswordResetToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	_, err := s.db.Exec("INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES (?, ?, ?)", userID, token, expiresAt)
	if err != nil {
		return errors.NewDatabaseInsertError("password reset token", err)
	}
	return nil
}

// GetPasswordResetToken retrieves a password reset token and associated user
func (s *TokenService) GetPasswordResetToken(ctx context.Context, token string) (ports.PasswordResetToken, ports.User, error) {
	var resetToken ports.PasswordResetToken
	var user ports.User

	query := `
		SELECT 
			prt.id, prt.user_id, prt.token, prt.expires_at, prt.used, prt.used_at, prt.created_at,
			u.id, u.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, u.short_uid, u.email, u.first_name, u.last_name, u.password, u.must_reset_password, u.is_active, u.email_verified, u.password_changed_at, u.created_at, u.updated_at
		FROM password_reset_tokens prt
		JOIN users u ON prt.user_id = u.id
		LEFT JOIN tenants t ON u.tenant_id = t.id
		WHERE prt.token = ? AND prt.used = FALSE AND prt.expires_at > datetime('now')`

	err := s.db.QueryRow(query, token).Scan(
		&resetToken.ID, &resetToken.UserID, &resetToken.Token, &resetToken.ExpiresAt, &resetToken.Used, &resetToken.UsedAt, &resetToken.CreatedAt,
		&user.ID, &user.TenantID, &user.TenantUID, &user.ShortUID, &user.Email, &user.FirstName, &user.LastName, &user.Password, &user.MustResetPassword, &user.IsActive, &user.EmailVerified, &user.PasswordChangedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.PasswordResetToken{}, ports.User{}, errors.NewPasswordResetTokenNotFoundError()
		}
		return ports.PasswordResetToken{}, ports.User{}, errors.NewDatabaseQueryError("password reset token", err)
	}

	return resetToken, user, nil
}

// MarkPasswordResetTokenAsUsed marks a password reset token as used
func (s *TokenService) MarkPasswordResetTokenAsUsed(ctx context.Context, token string) error {
	result, err := s.db.Exec("UPDATE password_reset_tokens SET used = TRUE, used_at = CURRENT_TIMESTAMP WHERE token = ?", token)
	if err != nil {
		return errors.NewDatabaseUpdateError("password reset token", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.ErrTokenNotFound
	}

	return nil
}

// CleanupExpiredPasswordResetTokens removes expired password reset tokens
func (s *TokenService) CleanupExpiredPasswordResetTokens(ctx context.Context) error {
	_, err := s.db.Exec("DELETE FROM password_reset_tokens WHERE expires_at < datetime('now') OR used = TRUE")
	if err != nil {
		return errors.NewDatabaseDeleteError("expired password reset tokens", err)
	}
	return nil
}

// Email Verification Token Methods

// CreateEmailVerificationToken creates a new email verification token
func (s *TokenService) CreateEmailVerificationToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	_, err := s.db.Exec("INSERT INTO email_verification_tokens (user_id, token, expires_at) VALUES (?, ?, ?)", userID, token, expiresAt)
	if err != nil {
		return errors.NewDatabaseInsertError("email verification token", err)
	}
	return nil
}

// GetEmailVerificationToken retrieves an email verification token and associated user
func (s *TokenService) GetEmailVerificationToken(ctx context.Context, token string) (ports.EmailVerificationToken, ports.User, error) {
	var verifyToken ports.EmailVerificationToken
	var user ports.User

	query := `
		SELECT 
			evt.id, evt.user_id, evt.token, evt.expires_at, evt.used, evt.used_at, evt.created_at,
			u.id, u.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, u.short_uid, u.email, u.first_name, u.last_name, u.password, u.must_reset_password, u.is_active, u.email_verified, u.password_changed_at, u.created_at, u.updated_at
		FROM email_verification_tokens evt
		JOIN users u ON evt.user_id = u.id
		LEFT JOIN tenants t ON u.tenant_id = t.id
		WHERE evt.token = ? AND evt.used = FALSE AND evt.expires_at > datetime('now')`

	err := s.db.QueryRow(query, token).Scan(
		&verifyToken.ID, &verifyToken.UserID, &verifyToken.Token, &verifyToken.ExpiresAt, &verifyToken.Used, &verifyToken.UsedAt, &verifyToken.CreatedAt,
		&user.ID, &user.TenantID, &user.TenantUID, &user.ShortUID, &user.Email, &user.FirstName, &user.LastName, &user.Password, &user.MustResetPassword, &user.IsActive, &user.EmailVerified, &user.PasswordChangedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.EmailVerificationToken{}, ports.User{}, errors.NewEmailVerificationTokenNotFoundError()
		}
		return ports.EmailVerificationToken{}, ports.User{}, errors.NewDatabaseQueryError("email verification token", err)
	}

	return verifyToken, user, nil
}

// MarkEmailVerificationTokenAsUsed marks an email verification token as used
func (s *TokenService) MarkEmailVerificationTokenAsUsed(ctx context.Context, token string) error {
	result, err := s.db.Exec("UPDATE email_verification_tokens SET used = TRUE, used_at = CURRENT_TIMESTAMP WHERE token = ?", token)
	if err != nil {
		return errors.NewDatabaseUpdateError("email verification token", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.ErrTokenNotFound
	}

	return nil
}

// CleanupExpiredEmailVerificationTokens removes expired email verification tokens
func (s *TokenService) CleanupExpiredEmailVerificationTokens(ctx context.Context) error {
	_, err := s.db.Exec("DELETE FROM email_verification_tokens WHERE expires_at < datetime('now') OR used = TRUE")
	if err != nil {
		return errors.NewDatabaseDeleteError("expired email verification tokens", err)
	}
	return nil
}
