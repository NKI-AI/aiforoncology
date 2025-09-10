// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package ports

import (
	"context"
	"time"

	"aifo.dev/aifo/slideinsight/internal/utils"
)

// User represents a user in the database
type User struct {
	ID                int
	TenantID          int
	TenantUID         string
	ShortUID          string
	Email             string
	FirstName         string
	LastName          string
	Password          string
	MustResetPassword bool
	IsActive          bool
	EmailVerified     bool
	PasswordChangedAt *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewUser represents a new user to be created to the database
type NewUser struct {
	TenantID          int
	ShortUID          string
	Email             string
	FirstName         string
	LastName          string
	Password          string
	MustResetPassword bool
	IsActive          bool
	EmailVerified     bool
}

// UserUpdates represents fields that can be updated for an existing user
type UserUpdates struct {
	Email             *string
	FirstName         *string
	LastName          *string
	MustResetPassword *bool
	IsActive          *bool
	EmailVerified     *bool
}

// PasswordHistory represents a historical password entry
type PasswordHistory struct {
	ID           int
	UserID       int
	PasswordHash string
	CreatedAt    time.Time
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        int
	UserID    int
	Token     string
	ExpiresAt time.Time
	Used      bool
	UsedAt    *time.Time
	CreatedAt time.Time
}

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	ID        int
	UserID    int
	Token     string
	ExpiresAt time.Time
	Used      bool
	UsedAt    *time.Time
	CreatedAt time.Time
}

// AuthAttempt represents an authentication attempt
type AuthAttempt struct {
	ID          int
	IPAddress   string
	Email       string
	Success     bool
	FailReason  string
	AttemptedAt time.Time
}

// UsersRepository defines the interface for user-related database operations
type UsersRepository interface {
	// CreateUser adds a new user to the database.
	CreateUser(ctx context.Context, newUser NewUser) error

	// GetUserByEmail retrieves a specific user by email address.
	GetUserByEmail(ctx context.Context, email string) (User, error)

	// GetUserByUID retrieves a specific user by its UID.
	GetUserByUID(ctx context.Context, userUID string) (User, error)

	// GetUserByInternalID retrieves a specific user by its internal database ID.
	GetUserByInternalID(ctx context.Context, userID int) (User, error)

	// LoadAllUsers retrieves users from the database with optional search/filter and pagination support.
	LoadAllUsers(ctx context.Context, search utils.SearchParams, limit, offset int) ([]User, error)

	// UpdateUserPassword updates the password for a user with the specified email.
	UpdateUserPassword(ctx context.Context, email string, hashedPassword string) error

	// UpdateUser updates user information (excluding password) for a user with the specified email.
	UpdateUser(ctx context.Context, email string, updates UserUpdates) error

	// UpdateUserByUID updates user information (excluding password) for a user with the specified UID.
	UpdateUserByUID(ctx context.Context, userUID string, updates UserUpdates) error

	// DeleteUser removes a user from the database after checking for dependencies.
	DeleteUser(ctx context.Context, userUID string) error

	// GetUserCount returns the total count of users matching optional search criteria.
	GetUserCount(ctx context.Context, search utils.SearchParams) (int, error)

	// DeactivateUser marks a user as inactive by setting is_active to false.
	DeactivateUser(ctx context.Context, email string) error

	// ActivateUser marks a user as active by setting is_active to true.
	ActivateUser(ctx context.Context, email string) error

	// Password History Methods
	AddPasswordToHistory(ctx context.Context, userID int, passwordHash string) error
	GetPasswordHistory(ctx context.Context, userID int, months int) ([]PasswordHistory, error)
	CleanupOldPasswordHistory(ctx context.Context, userID int, keepCount int) error

	// Password Reset Token Methods
	CreatePasswordResetToken(ctx context.Context, userID int, token string, expiresAt time.Time) error
	GetPasswordResetToken(ctx context.Context, token string) (PasswordResetToken, User, error)
	MarkPasswordResetTokenAsUsed(ctx context.Context, token string) error
	CleanupExpiredPasswordResetTokens(ctx context.Context) error

	// Email Verification Token Methods
	CreateEmailVerificationToken(ctx context.Context, userID int, token string, expiresAt time.Time) error
	GetEmailVerificationToken(ctx context.Context, token string) (EmailVerificationToken, User, error)
	MarkEmailVerificationTokenAsUsed(ctx context.Context, token string) error
	CleanupExpiredEmailVerificationTokens(ctx context.Context) error

	// Authentication Attempt Methods
	RecordAuthAttempt(ctx context.Context, ipAddress, email string, success bool, failReason string) error
	GetRecentAuthAttempts(ctx context.Context, ipAddress string, since time.Time) ([]AuthAttempt, error)
	GetRecentAuthAttemptsForUser(ctx context.Context, email string, since time.Time) ([]AuthAttempt, error)
	CleanupOldAuthAttempts(ctx context.Context, olderThan time.Time) error
}
