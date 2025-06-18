// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Default cost for bcrypt - can be adjusted based on security vs performance needs
const DefaultBcryptCost = 12

var (
	ErrEmptyPassword     = errors.New("password cannot be empty")
	ErrPasswordMismatch  = errors.New("password does not match stored hash")
	ErrInvalidHashFormat = errors.New("invalid password hash format")
)

// HashPassword creates a bcrypt hash of the provided password with the default cost
func HashPassword(password string) (string, error) {
	return HashPasswordWithCost(password, DefaultBcryptCost)
}

// HashPasswordWithCost creates a bcrypt hash of the provided password with a specified cost
func HashPasswordWithCost(password string, cost int) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	// Generate bcrypt hash from the password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword compares a plaintext password against a stored hash
func VerifyPassword(hashedPassword, plainPassword string) error {
	if plainPassword == "" {
		return ErrEmptyPassword
	}

	if hashedPassword == "" {
		return ErrInvalidHashFormat
	}

	// Compare the provided password with the stored hash
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordMismatch
		}
		return fmt.Errorf("error verifying password: %w", err)
	}

	return nil
}

// IsHashedPassword checks if a string appears to be a bcrypt hash
func IsHashedPassword(password string) bool {
	// bcrypt hashes start with $2a$, $2b$, or $2y$
	return len(password) >= 60 && (password[:4] == "$2a$" ||
		password[:4] == "$2b$" || password[:4] == "$2y$")
}
