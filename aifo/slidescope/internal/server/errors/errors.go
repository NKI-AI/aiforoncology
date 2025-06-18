// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package errors

import (
	"errors"
	"fmt"
)

// Common error types that can be used across services
var (
	// ErrNotFound is returned when a resource doesn't exist
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput is returned for general validation errors
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned for authentication failures
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInternal is returned for unexpected server errors
	ErrInternal = errors.New("internal server error")

	// ErrAlreadyExists is returned when trying to create a resource that already exists
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrOutOfBounds is returned when requesting data outside valid range
	ErrOutOfBounds = errors.New("out of bounds")
)

// Domain-specific error types
var (
	// ErrSlideNotFound is returned when a slide doesn't exist
	ErrSlideNotFound = fmt.Errorf("%w: slide not found", ErrNotFound)

	// ErrMaskNotFound is returned when a mask doesn't exist
	ErrMaskNotFound = fmt.Errorf("%w: mask not found", ErrNotFound)

	// ErrTileOutOfBounds is returned when requesting a tile outside slide/mask bounds
	ErrTileOutOfBounds = fmt.Errorf("%w: tile coordinates out of bounds", ErrOutOfBounds)

	// ErrInvalidFormat is returned when an unsupported format is requested
	ErrInvalidFormat = fmt.Errorf("%w: unsupported format", ErrInvalidInput)
)

// WithDetails wraps an error with additional context details
func WithDetails(err error, format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", err, fmt.Sprintf(format, args...))
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsInvalidInput checks if the error is a validation error
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorized checks if the error is an authorization error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsOutOfBounds checks if the error is an out of bounds error
func IsOutOfBounds(err error) bool {
	return errors.Is(err, ErrOutOfBounds)
}

// IsAlreadyExists checks if the error is an already exists error
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}
