// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package errors

import (
	"fmt"
)

// Auth-specific error types
var (
	// ErrInvalidCredentials is returned for incorrect username/password
	ErrInvalidCredentials = fmt.Errorf("%w: invalid username or password", ErrUnauthorized)

	// ErrTokenExpired is returned when JWT token has expired
	ErrTokenExpired = fmt.Errorf("%w: token has expired", ErrUnauthorized)

	// ErrInvalidToken is returned when JWT token is invalid
	ErrInvalidToken = fmt.Errorf("%w: invalid token", ErrUnauthorized)

	// ErrMissingToken is returned when no JWT token is provided
	ErrMissingToken = fmt.Errorf("%w: missing authentication token", ErrUnauthorized)

	// ErrInsufficientPermissions is returned when user doesn't have required permissions
	ErrInsufficientPermissions = fmt.Errorf("%w: insufficient permissions", ErrUnauthorized)
)
