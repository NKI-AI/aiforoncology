// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package errors

import (
	"errors"
	"fmt"
)

// User-specific error types
var (
	// ErrUserNotFound is returned when a user doesn't exist
	ErrUserNotFound = fmt.Errorf("%w: user not found", ErrNotFound)

	// ErrUserNotFoundByUID is returned when a user with specific UID doesn't exist
	ErrUserNotFoundByUID = fmt.Errorf("%w: user not found by UID", ErrNotFound)

	// ErrUserNotFoundByEmail is returned when a user with specific email doesn't exist
	ErrUserNotFoundByEmail = fmt.Errorf("%w: user not found by email", ErrNotFound)

	// ErrEmailEmpty is returned when email is empty
	ErrEmailEmpty = fmt.Errorf("%w: email cannot be empty", ErrInvalidInput)

	// ErrPasswordEmpty is returned when password is empty
	ErrPasswordEmpty = fmt.Errorf("%w: password cannot be empty", ErrInvalidInput)

	// ErrUserUIDEmpty is returned when user UID is empty
	ErrUserUIDEmpty = fmt.Errorf("%w: user UID cannot be empty", ErrInvalidInput)

	// ErrTenantRequired is returned when tenant is required but not provided
	ErrTenantRequired = fmt.Errorf("%w: tenant is required", ErrInvalidInput)

	// ErrUserAlreadyExists is returned when trying to create a user that already exists
	ErrUserAlreadyExists = fmt.Errorf("%w: user already exists", ErrAlreadyExists)

	// ErrUserHasDependencies is returned when trying to delete a user that has dependencies
	ErrUserHasDependencies = fmt.Errorf("%w: user has dependencies", ErrAlreadyExists)
)

// User error helper functions that return formatted errors with context

// NewUserNotFoundByUIDError creates a user not found error with UID context
func NewUserNotFoundByUIDError(userUID string) error {
	return WithDetails(ErrUserNotFoundByUID, "UID '%s'", userUID)
}

// NewUserNotFoundByEmailError creates a user not found error with email context
func NewUserNotFoundByEmailError(email string) error {
	return WithDetails(ErrUserNotFoundByEmail, "email '%s'", email)
}

// NewUserAlreadyExistsError creates a user already exists error with email context
func NewUserAlreadyExistsError(email string) error {
	return WithDetails(ErrUserAlreadyExists, "email '%s'", email)
}

// NewUserHasDependenciesError creates a user has dependencies error with context
func NewUserHasDependenciesError(userUID, dependencies string) error {
	return WithDetails(ErrUserHasDependencies, "UID '%s' - %s", userUID, dependencies)
}

// User error type checks
func IsUserNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrUserNotFoundByUID) ||
		errors.Is(err, ErrUserNotFoundByEmail)
}

func IsUserAlreadyExists(err error) bool {
	return errors.Is(err, ErrUserAlreadyExists)
}

func IsUserHasDependencies(err error) bool {
	return errors.Is(err, ErrUserHasDependencies)
}

func IsEmailEmpty(err error) bool {
	return errors.Is(err, ErrEmailEmpty)
}

func IsPasswordEmpty(err error) bool {
	return errors.Is(err, ErrPasswordEmpty)
}
