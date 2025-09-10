// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package errors

import (
	"errors"
	"fmt"
)

// Auth-specific error types
var (
	// ErrInvalidCredentials is returned for incorrect email/password
	ErrInvalidCredentials = fmt.Errorf("%w: invalid email or password", ErrUnauthorized)

	// ErrPasswordResetRequired is returned when user must reset their password before logging in
	ErrPasswordResetRequired = errors.New("password reset required")

	// ErrAccountInactive is returned when user account is not active
	ErrAccountInactive = fmt.Errorf("%w: account is not active", ErrUnauthorized)

	// ErrEmailExists is returned when trying to register with an existing email
	ErrEmailExists = fmt.Errorf("%w: email already registered", ErrAlreadyExists)

	// ErrInvalidResetToken is returned when password reset token is invalid or expired
	ErrInvalidResetToken = fmt.Errorf("%w: invalid or expired reset token", ErrUnauthorized)

	// ErrInvalidVerificationToken is returned when email verification token is invalid or expired
	ErrInvalidVerificationToken = fmt.Errorf("%w: invalid or expired verification token", ErrUnauthorized)

	// ErrPasswordValidation is returned when password doesn't meet requirements
	ErrPasswordValidation = fmt.Errorf("%w: password validation failed", ErrInvalidInput)

	// ErrTokenExpired is returned when JWT token has expired
	ErrTokenExpired = fmt.Errorf("%w: token has expired", ErrUnauthorized)

	// ErrInvalidToken is returned when JWT token is invalid
	ErrInvalidToken = fmt.Errorf("%w: invalid token", ErrUnauthorized)

	// ErrMissingToken is returned when no JWT token is provided
	ErrMissingToken = fmt.Errorf("%w: missing authentication token", ErrUnauthorized)

	// ErrInsufficientPermissions is returned when user doesn't have required permissions
	ErrInsufficientPermissions = fmt.Errorf("%w: insufficient permissions", ErrUnauthorized)
)

// IsPasswordResetRequired checks if the error is a password reset required error
func IsPasswordResetRequired(err error) bool {
	return errors.Is(err, ErrPasswordResetRequired)
}

// IsAccountInactive checks if the error is an account inactive error
func IsAccountInactive(err error) bool {
	return errors.Is(err, ErrAccountInactive)
}

// IsInvalidResetToken checks if the error is an invalid reset token error
func IsInvalidResetToken(err error) bool {
	return errors.Is(err, ErrInvalidResetToken)
}

// IsInvalidVerificationToken checks if the error is an invalid verification token error
func IsInvalidVerificationToken(err error) bool {
	return errors.Is(err, ErrInvalidVerificationToken)
}

// IsPasswordValidation checks if the error is a password validation error
func IsPasswordValidation(err error) bool {
	return errors.Is(err, ErrPasswordValidation)
}
