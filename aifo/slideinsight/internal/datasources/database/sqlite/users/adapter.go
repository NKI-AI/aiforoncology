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

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Adapter provides a unified interface for all user operations
type Adapter struct {
	crud   *CrudService
	search *SearchService
	auth   *AuthService
	tokens *TokenService
}

// NewAdapter creates a new users adapter
func NewAdapter(db *sql.DB) *Adapter {
	return &Adapter{
		crud:   NewCrudService(db),
		search: NewSearchService(db),
		auth:   NewAuthService(db),
		tokens: NewTokenService(db),
	}
}

// Basic CRUD operations

// CreateUser adds a new user to the database
func (a *Adapter) CreateUser(ctx context.Context, newUser ports.NewUser) error {
	return a.crud.CreateUser(ctx, newUser)
}

// GetUserByUID retrieves a specific user by its UID
func (a *Adapter) GetUserByUID(ctx context.Context, userUID string) (ports.User, error) {
	return a.crud.GetUserByUID(ctx, userUID)
}

// GetUserByEmail retrieves a specific user by email address
func (a *Adapter) GetUserByEmail(ctx context.Context, email string) (ports.User, error) {
	return a.crud.GetUserByEmail(ctx, email)
}

// GetUserByInternalID retrieves a specific user by its internal database ID
func (a *Adapter) GetUserByInternalID(ctx context.Context, userID int) (ports.User, error) {
	return a.crud.GetUserByInternalID(ctx, userID)
}

// UpdateUser updates user information (excluding password) for a user with the specified email
func (a *Adapter) UpdateUser(ctx context.Context, email string, updates ports.UserUpdates) error {
	return a.crud.UpdateUser(ctx, email, updates)
}

// UpdateUserByUID updates user information (excluding password) for a user with the specified UID
func (a *Adapter) UpdateUserByUID(ctx context.Context, userUID string, updates ports.UserUpdates) error {
	return a.crud.UpdateUserByUID(ctx, userUID, updates)
}

// DeleteUser removes a user from the database after checking for dependencies
func (a *Adapter) DeleteUser(ctx context.Context, userUID string) error {
	return a.crud.DeleteUser(ctx, userUID)
}

// Search and listing operations

// LoadAllUsers retrieves users from the database with optional search/filter and pagination support
func (a *Adapter) LoadAllUsers(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.User, error) {
	return a.search.LoadAllUsers(ctx, search, limit, offset)
}

// GetUserCount returns the total count of users matching optional search criteria
func (a *Adapter) GetUserCount(ctx context.Context, search utils.SearchParams) (int, error) {
	return a.search.GetUserCount(ctx, search)
}

// Authentication operations

// DeactivateUser marks a user as inactive
func (a *Adapter) DeactivateUser(ctx context.Context, email string) error {
	return a.auth.DeactivateUser(ctx, email)
}

// ActivateUser marks a user as active
func (a *Adapter) ActivateUser(ctx context.Context, email string) error {
	return a.auth.ActivateUser(ctx, email)
}

// Password history operations

// AddPasswordToHistory adds a password hash to the user's password history
func (a *Adapter) AddPasswordToHistory(ctx context.Context, userID int, passwordHash string) error {
	return a.auth.AddPasswordToHistory(ctx, userID, passwordHash)
}

// GetPasswordHistory retrieves password history for a user
func (a *Adapter) GetPasswordHistory(ctx context.Context, userID int, months int) ([]ports.PasswordHistory, error) {
	return a.auth.GetPasswordHistory(ctx, userID, months)
}

// CleanupOldPasswordHistory removes expired password history entries
func (a *Adapter) CleanupOldPasswordHistory(ctx context.Context, userID int, keepCount int) error {
	return a.auth.CleanupOldPasswordHistory(ctx, userID, keepCount)
}

// Authentication attempt operations

// RecordAuthAttempt records an authentication attempt
func (a *Adapter) RecordAuthAttempt(ctx context.Context, ipAddress, email string, success bool, failReason string) error {
	return a.auth.RecordAuthAttempt(ctx, ipAddress, email, success, failReason)
}

// GetRecentAuthAttempts retrieves recent authentication attempts for an IP address
func (a *Adapter) GetRecentAuthAttempts(ctx context.Context, ipAddress string, since time.Time) ([]ports.AuthAttempt, error) {
	return a.auth.GetRecentAuthAttempts(ctx, ipAddress, since)
}

// GetRecentAuthAttemptsForUser retrieves recent authentication attempts for an email
func (a *Adapter) GetRecentAuthAttemptsForUser(ctx context.Context, email string, since time.Time) ([]ports.AuthAttempt, error) {
	return a.auth.GetRecentAuthAttemptsForUser(ctx, email, since)
}

// CleanupOldAuthAttempts removes old authentication attempts
func (a *Adapter) CleanupOldAuthAttempts(ctx context.Context, olderThan time.Time) error {
	return a.auth.CleanupOldAuthAttempts(ctx, olderThan)
}

// Password reset token operations

// CreatePasswordResetToken creates a new password reset token
func (a *Adapter) CreatePasswordResetToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	return a.tokens.CreatePasswordResetToken(ctx, userID, token, expiresAt)
}

// GetPasswordResetToken retrieves a password reset token and associated user
func (a *Adapter) GetPasswordResetToken(ctx context.Context, token string) (ports.PasswordResetToken, ports.User, error) {
	return a.tokens.GetPasswordResetToken(ctx, token)
}

// MarkPasswordResetTokenAsUsed marks a password reset token as used
func (a *Adapter) MarkPasswordResetTokenAsUsed(ctx context.Context, token string) error {
	return a.tokens.MarkPasswordResetTokenAsUsed(ctx, token)
}

// CleanupExpiredPasswordResetTokens removes expired password reset tokens
func (a *Adapter) CleanupExpiredPasswordResetTokens(ctx context.Context) error {
	return a.tokens.CleanupExpiredPasswordResetTokens(ctx)
}

// Email verification token operations

// CreateEmailVerificationToken creates a new email verification token
func (a *Adapter) CreateEmailVerificationToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	return a.tokens.CreateEmailVerificationToken(ctx, userID, token, expiresAt)
}

// GetEmailVerificationToken retrieves an email verification token and associated user
func (a *Adapter) GetEmailVerificationToken(ctx context.Context, token string) (ports.EmailVerificationToken, ports.User, error) {
	return a.tokens.GetEmailVerificationToken(ctx, token)
}

// MarkEmailVerificationTokenAsUsed marks an email verification token as used
func (a *Adapter) MarkEmailVerificationTokenAsUsed(ctx context.Context, token string) error {
	return a.tokens.MarkEmailVerificationTokenAsUsed(ctx, token)
}

// CleanupExpiredEmailVerificationTokens removes expired email verification tokens
func (a *Adapter) CleanupExpiredEmailVerificationTokens(ctx context.Context) error {
	return a.tokens.CleanupExpiredEmailVerificationTokens(ctx)
}

// UpdateUserPassword updates the password for a user with the specified email
func (a *Adapter) UpdateUserPassword(ctx context.Context, email string, hashedPassword string) error {
	return a.crud.UpdateUserPassword(ctx, email, hashedPassword)
}
