// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	// Test successful hashing
	password := "securePassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed with valid password: %v", err)
	}
	if !IsHashedPassword(hash) {
		t.Errorf("Output of HashPassword doesn't appear to be a valid bcrypt hash: %s", hash)
	}

	// Test empty password
	_, err = HashPassword("")
	if err != ErrEmptyPassword {
		t.Errorf("HashPassword with empty password should return ErrEmptyPassword, got: %v", err)
	}
}

func TestHashPasswordWithCost(t *testing.T) {
	// Using a faster maximum cost for testing (actual MaxCost is 31 and is too slow)
	testMaxCost := 14 // Much faster than bcrypt.MaxCost (31)

	// Test successful hashing with various costs
	testCases := []struct {
		name     string
		password string
		cost     int
		wantErr  bool
	}{
		{"valid password with min cost", "password123", bcrypt.MinCost, false},
		{"valid password with default cost", "password123", DefaultBcryptCost, false},
		{"valid password with higher cost", "password123", testMaxCost, false},
		{"empty password", "", DefaultBcryptCost, true},
		{"cost too high", "password123", bcrypt.MaxCost + 1, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := HashPasswordWithCost(tc.password, tc.cost)

			// Check error scenarios
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			// Check success scenarios
			if err != nil {
				t.Fatalf("HashPasswordWithCost failed: %v", err)
			}

			if !IsHashedPassword(hash) {
				t.Errorf("Output doesn't appear to be a valid bcrypt hash: %s", hash)
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	// Create a hash for testing
	password := "correctPassword123"
	wrongPassword := "wrongPassword456"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to create test hash: %v", err)
	}

	// Test successful verification
	err = VerifyPassword(hash, password)
	if err != nil {
		t.Errorf("VerifyPassword failed with correct password: %v", err)
	}

	// Test password mismatch
	err = VerifyPassword(hash, wrongPassword)
	if err != ErrPasswordMismatch {
		t.Errorf("VerifyPassword with wrong password should return ErrPasswordMismatch, got: %v", err)
	}

	// Test empty plain password
	err = VerifyPassword(hash, "")
	if err != ErrEmptyPassword {
		t.Errorf("VerifyPassword with empty password should return ErrEmptyPassword, got: %v", err)
	}

	// Test empty hashed password
	err = VerifyPassword("", password)
	if err != ErrInvalidHashFormat {
		t.Errorf("VerifyPassword with empty hash should return ErrInvalidHashFormat, got: %v", err)
	}

	// Test invalid hash format
	err = VerifyPassword("not-a-valid-hash", password)
	if err == nil || err == ErrPasswordMismatch {
		t.Errorf("VerifyPassword with invalid hash format should return an error different from ErrPasswordMismatch, got: %v", err)
	}
}

func TestIsHashedPassword(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid $2a$ hash", "$2a$10$abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklm", true},
		{"valid $2b$ hash", "$2b$10$abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklm", true},
		{"valid $2y$ hash", "$2y$10$abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklm", true},
		{"wrong prefix", "$3a$10$abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklm", false},
		{"too short", "$2a$10$short", false},
		{"empty string", "", false},
		{"plain text", "plainTextPassword", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsHashedPassword(tc.input)
			if result != tc.expected {
				t.Errorf("IsHashedPassword(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}

	// Test with actual generated hash
	password := "testPassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to create test hash: %v", err)
	}
	if !IsHashedPassword(hash) {
		t.Errorf("Generated hash is not recognized as valid: %s", hash)
	}
}

func TestRealBcryptCompatibility(t *testing.T) {
	// This test verifies our wrapper functions are compatible with the underlying bcrypt library
	password := "realPasswordTest"

	// Create hash with our function - use a low cost for faster tests
	hash, err := HashPasswordWithCost(password, bcrypt.MinCost)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Verify directly with bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		t.Errorf("Our hash is not compatible with bcrypt verification: %v", err)
	}

	// Create hash directly with bcrypt - use a low cost for faster tests
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword failed: %v", err)
	}

	// Verify with our function
	err = VerifyPassword(string(bcryptHash), password)
	if err != nil {
		t.Errorf("bcrypt hash is not compatible with our verification: %v", err)
	}
}
