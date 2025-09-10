// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package sqlite

import (
	"context"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// User-related methods that delegate to the users adapter

// Basic CRUD operations

// CreateUser adds a new user to the database
func (db *DB) CreateUser(ctx context.Context, newUser ports.NewUser) error {
	return db.users.CreateUser(ctx, newUser)
}

// GetUserByUID retrieves a specific user by its UID
func (db *DB) GetUserByUID(ctx context.Context, userUID string) (ports.User, error) {
	return db.users.GetUserByUID(ctx, userUID)
}

// GetUserByEmail retrieves a specific user by email address
func (db *DB) GetUserByEmail(ctx context.Context, email string) (ports.User, error) {
	return db.users.GetUserByEmail(ctx, email)
}

// GetUserByInternalID retrieves a specific user by its internal database ID
func (db *DB) GetUserByInternalID(ctx context.Context, userID int) (ports.User, error) {
	return db.users.GetUserByInternalID(ctx, userID)
}

// UpdateUserPassword updates the password for a user with the specified email
func (db *DB) UpdateUserPassword(ctx context.Context, email string, hashedPassword string) error {
	return db.users.UpdateUserPassword(ctx, email, hashedPassword)
}

// UpdateUser updates user information (excluding password) for a user with the specified email
func (db *DB) UpdateUser(ctx context.Context, email string, updates ports.UserUpdates) error {
	return db.users.UpdateUser(ctx, email, updates)
}

// UpdateUserByUID updates user information (excluding password) for a user with the specified UID
func (db *DB) UpdateUserByUID(ctx context.Context, userUID string, updates ports.UserUpdates) error {
	return db.users.UpdateUserByUID(ctx, userUID, updates)
}

// DeleteUser removes a user from the database after checking for dependencies
func (db *DB) DeleteUser(ctx context.Context, userUID string) error {
	return db.users.DeleteUser(ctx, userUID)
}

// Search and listing operations

// LoadAllUsers retrieves users from the database with optional search/filter and pagination support
func (db *DB) LoadAllUsers(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.User, error) {
	return db.users.LoadAllUsers(ctx, search, limit, offset)
}

// GetUserCount returns the total count of users matching optional search criteria
func (db *DB) GetUserCount(ctx context.Context, search utils.SearchParams) (int, error) {
	return db.users.GetUserCount(ctx, search)
}

// Authentication operations

// DeactivateUser marks a user as inactive
func (db *DB) DeactivateUser(ctx context.Context, email string) error {
	return db.users.DeactivateUser(ctx, email)
}

// ActivateUser marks a user as active
func (db *DB) ActivateUser(ctx context.Context, email string) error {
	return db.users.ActivateUser(ctx, email)
}

// Password history operations

// AddPasswordToHistory adds a password hash to the user's password history
func (db *DB) AddPasswordToHistory(ctx context.Context, userID int, passwordHash string) error {
	return db.users.AddPasswordToHistory(ctx, userID, passwordHash)
}

// GetPasswordHistory retrieves password history for a user
func (db *DB) GetPasswordHistory(ctx context.Context, userID int, months int) ([]ports.PasswordHistory, error) {
	return db.users.GetPasswordHistory(ctx, userID, months)
}

// CleanupOldPasswordHistory removes expired password history entries
func (db *DB) CleanupOldPasswordHistory(ctx context.Context, userID int, keepCount int) error {
	return db.users.CleanupOldPasswordHistory(ctx, userID, keepCount)
}

// Authentication attempt operations

// RecordAuthAttempt records an authentication attempt
func (db *DB) RecordAuthAttempt(ctx context.Context, ipAddress, email string, success bool, failReason string) error {
	return db.users.RecordAuthAttempt(ctx, ipAddress, email, success, failReason)
}

// GetRecentAuthAttempts retrieves recent authentication attempts for an IP address
func (db *DB) GetRecentAuthAttempts(ctx context.Context, ipAddress string, since time.Time) ([]ports.AuthAttempt, error) {
	return db.users.GetRecentAuthAttempts(ctx, ipAddress, since)
}

// GetRecentAuthAttemptsForUser retrieves recent authentication attempts for an email
func (db *DB) GetRecentAuthAttemptsForUser(ctx context.Context, email string, since time.Time) ([]ports.AuthAttempt, error) {
	return db.users.GetRecentAuthAttemptsForUser(ctx, email, since)
}

// CleanupOldAuthAttempts removes old authentication attempts
func (db *DB) CleanupOldAuthAttempts(ctx context.Context, olderThan time.Time) error {
	return db.users.CleanupOldAuthAttempts(ctx, olderThan)
}

// Password reset token operations

// CreatePasswordResetToken creates a new password reset token
func (db *DB) CreatePasswordResetToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	return db.users.CreatePasswordResetToken(ctx, userID, token, expiresAt)
}

// GetPasswordResetToken retrieves a password reset token and associated user
func (db *DB) GetPasswordResetToken(ctx context.Context, token string) (ports.PasswordResetToken, ports.User, error) {
	return db.users.GetPasswordResetToken(ctx, token)
}

// MarkPasswordResetTokenAsUsed marks a password reset token as used
func (db *DB) MarkPasswordResetTokenAsUsed(ctx context.Context, token string) error {
	return db.users.MarkPasswordResetTokenAsUsed(ctx, token)
}

// CleanupExpiredPasswordResetTokens removes expired password reset tokens
func (db *DB) CleanupExpiredPasswordResetTokens(ctx context.Context) error {
	return db.users.CleanupExpiredPasswordResetTokens(ctx)
}

// Email verification token operations

// CreateEmailVerificationToken creates a new email verification token
func (db *DB) CreateEmailVerificationToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	return db.users.CreateEmailVerificationToken(ctx, userID, token, expiresAt)
}

// GetEmailVerificationToken retrieves an email verification token and associated user
func (db *DB) GetEmailVerificationToken(ctx context.Context, token string) (ports.EmailVerificationToken, ports.User, error) {
	return db.users.GetEmailVerificationToken(ctx, token)
}

// MarkEmailVerificationTokenAsUsed marks an email verification token as used
func (db *DB) MarkEmailVerificationTokenAsUsed(ctx context.Context, token string) error {
	return db.users.MarkEmailVerificationTokenAsUsed(ctx, token)
}

// CleanupExpiredEmailVerificationTokens removes expired email verification tokens
func (db *DB) CleanupExpiredEmailVerificationTokens(ctx context.Context) error {
	return db.users.CleanupExpiredEmailVerificationTokens(ctx)
}
