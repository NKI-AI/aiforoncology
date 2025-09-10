// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package domain

import "time"

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	RefreshExpiresIn int    `json:"refresh_expires_in,omitempty"`
}

// ChangePasswordRequest represents the request to change a user's password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

// ForcedChangePasswordRequest represents the request to change a password without authentication
// This is used when a user must reset their password before getting a JWT token
type ForcedChangePasswordRequest struct {
	Email           string `json:"email" validate:"required,email"`
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

// ResetPasswordRequest represents the request to initiate password reset
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordConfirmRequest represents the request to confirm password reset
type ResetPasswordConfirmRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}

// RegisterUserRequest represents the request to register a new user
type RegisterUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Password  string `json:"password" validate:"required,min=8"`
}

// ResendVerificationRequest represents the request to resend email verification
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// SwitchUserRequest represents the request to switch to another user (admin only)
type SwitchUserRequest struct {
	TargetUserUID string `json:"targetUserUid" validate:"required"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	Token     string    `json:"token"`
	UserID    int       `json:"userId"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expiresAt"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"createdAt"`
}

// EmailVerificationToken represents an email verification token for new registrations
type EmailVerificationToken struct {
	Token     string    `json:"token"`
	UserID    int       `json:"userId"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expiresAt"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"createdAt"`
}

// PasswordHistory represents a historical password for security tracking
type PasswordHistory struct {
	ID           int       `json:"id"`
	UserID       int       `json:"userId"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	CreatedAt    time.Time `json:"createdAt"`
}

// AuthAttempt represents a login attempt for rate limiting
type AuthAttempt struct {
	ID          int       `json:"id"`
	IP          string    `json:"ip"`
	Success     bool      `json:"success"`
	FailReason  string    `json:"failReason,omitempty"`
	AttemptedAt time.Time `json:"attemptedAt"`
}
