// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package errors

import (
	"fmt"
)

// Database operation error types
var (
	// Generic database operation errors
	ErrDatabaseInsert      = fmt.Errorf("%w: failed to insert record", ErrInternal)
	ErrDatabaseUpdate      = fmt.Errorf("%w: failed to update record", ErrInternal)
	ErrDatabaseDelete      = fmt.Errorf("%w: failed to delete record", ErrInternal)
	ErrDatabaseQuery       = fmt.Errorf("%w: failed to query database", ErrInternal)
	ErrDatabaseScan        = fmt.Errorf("%w: failed to scan database row", ErrInternal)
	ErrDatabaseConnect     = fmt.Errorf("%w: failed to connect to database", ErrInternal)
	ErrDatabaseIterateRows = fmt.Errorf("%w: error iterating over rows", ErrInternal)
	ErrDatabaseCheckRows   = fmt.Errorf("%w: error checking rows affected", ErrInternal)

	// Token operation errors
	ErrTokenNotFound          = fmt.Errorf("%w: token not found", ErrNotFound)
	ErrPasswordResetToken     = fmt.Errorf("%w: password reset token not found or expired", ErrNotFound)
	ErrEmailVerificationToken = fmt.Errorf("%w: email verification token not found or expired", ErrNotFound)

	// Validation errors for common fields
	ErrNoFieldsToUpdate = fmt.Errorf("%w: no fields to update", ErrInvalidInput)
	ErrNoRowsAffected   = fmt.Errorf("%w: no rows affected", ErrNotFound)
)

// Database error helper functions that return formatted errors with context

// NewDatabaseInsertError creates a database insert error with entity context
func NewDatabaseInsertError(entity string, err error) error {
	return WithDetails(ErrDatabaseInsert, "%s: %v", entity, err)
}

// NewDatabaseUpdateError creates a database update error with entity context
func NewDatabaseUpdateError(entity string, err error) error {
	return WithDetails(ErrDatabaseUpdate, "%s: %v", entity, err)
}

// NewDatabaseDeleteError creates a database delete error with entity context
func NewDatabaseDeleteError(entity string, err error) error {
	return WithDetails(ErrDatabaseDelete, "%s: %v", entity, err)
}

// NewDatabaseQueryError creates a database query error with entity context
func NewDatabaseQueryError(entity string, err error) error {
	return WithDetails(ErrDatabaseQuery, "%s: %v", entity, err)
}

// NewDatabaseScanError creates a database scan error with entity context
func NewDatabaseScanError(entity string, err error) error {
	return WithDetails(ErrDatabaseScan, "%s: %v", entity, err)
}

// NewDatabaseIterateRowsError creates a database row iteration error with entity context
func NewDatabaseIterateRowsError(entity string, err error) error {
	return WithDetails(ErrDatabaseIterateRows, "%s: %v", entity, err)
}

// NewDatabaseCheckRowsError creates a database check rows affected error
func NewDatabaseCheckRowsError(err error) error {
	return WithDetails(ErrDatabaseCheckRows, "%v", err)
}

// NewPasswordHashError creates a password hashing error
func NewPasswordHashError(err error) error {
	return WithDetails(ErrInternal, "failed to hash password: %v", err)
}

// NewUIDGenerationError creates a UID generation error
func NewUIDGenerationError(err error) error {
	return WithDetails(ErrInternal, "failed to generate UID: %v", err)
}

// NewGenericOperationError creates a generic operation error with context
func NewGenericOperationError(operation string, err error) error {
	return WithDetails(ErrInternal, "failed to %s: %v", operation, err)
}

// Token-specific error functions

// NewPasswordResetTokenNotFoundError creates a password reset token not found error
func NewPasswordResetTokenNotFoundError() error {
	return ErrPasswordResetToken
}

// NewEmailVerificationTokenNotFoundError creates an email verification token not found error
func NewEmailVerificationTokenNotFoundError() error {
	return ErrEmailVerificationToken
}

// Validation helper functions

// NewNoFieldsToUpdateError creates a no fields to update error
func NewNoFieldsToUpdateError() error {
	return ErrNoFieldsToUpdate
}

// NewNoRowsAffectedError creates a no rows affected error with entity context
func NewNoRowsAffectedError(entity, identifier string) error {
	return WithDetails(ErrNoRowsAffected, "%s with identifier '%s'", entity, identifier)
}
