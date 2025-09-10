// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const (
	// Password requirements
	MinPasswordLength = 8
	MaxPasswordLength = 128

	// Token lengths
	ResetTokenLength        = 32 // 32 bytes = 64 hex chars
	VerificationTokenLength = 32

	// Security constants
	MaxPasswordHistoryMonths = 6
	MaxPasswordHistoryCount  = 10
)

// PasswordRequirements defines password security requirements
type PasswordRequirements struct {
	MinLength        int
	MaxLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireDigits    bool
	RequireSpecial   bool
	ForbidCommon     bool
}

// DefaultPasswordRequirements returns default password security requirements
// TODO: Database configurable
func DefaultPasswordRequirements() PasswordRequirements {
	return PasswordRequirements{
		MinLength:        MinPasswordLength,
		MaxLength:        MaxPasswordLength,
		RequireUppercase: false,
		RequireLowercase: false,
		RequireDigits:    false,
		RequireSpecial:   false,
		ForbidCommon:     false,
	}
}

// PasswordValidationError represents a password validation error
type PasswordValidationError struct {
	Field   string
	Message string
}

func (e PasswordValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidatePassword validates a password against security requirements
func ValidatePassword(password string, requirements PasswordRequirements) []PasswordValidationError {
	var errors []PasswordValidationError

	// Check length
	if len(password) < requirements.MinLength {
		errors = append(errors, PasswordValidationError{
			Field:   "password",
			Message: fmt.Sprintf("password must be at least %d characters long", requirements.MinLength),
		})
	}

	if len(password) > requirements.MaxLength {
		errors = append(errors, PasswordValidationError{
			Field:   "password",
			Message: fmt.Sprintf("password must be no more than %d characters long", requirements.MaxLength),
		})
	}

	// Check character requirements
	if requirements.RequireUppercase && !hasUppercase(password) {
		errors = append(errors, PasswordValidationError{
			Field:   "password",
			Message: "password must contain at least one uppercase letter",
		})
	}

	if requirements.RequireLowercase && !hasLowercase(password) {
		errors = append(errors, PasswordValidationError{
			Field:   "password",
			Message: "password must contain at least one lowercase letter",
		})
	}

	if requirements.RequireDigits && !hasDigit(password) {
		errors = append(errors, PasswordValidationError{
			Field:   "password",
			Message: "password must contain at least one digit",
		})
	}

	if requirements.RequireSpecial && !hasSpecialChar(password) {
		errors = append(errors, PasswordValidationError{
			Field:   "password",
			Message: "password must contain at least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)",
		})
	}

	// Check against common passwords
	if requirements.ForbidCommon && isCommonPassword(password) {
		errors = append(errors, PasswordValidationError{
			Field:   "password",
			Message: "password is too common, please choose a more secure password",
		})
	}

	// Additional security checks
	if hasRepeatingChars(password, 3) {
		errors = append(errors, PasswordValidationError{
			Field:   "password",
			Message: "password must not contain more than 2 consecutive identical characters",
		})
	}

	if hasSequentialChars(password) {
		errors = append(errors, PasswordValidationError{
			Field:   "password",
			Message: "password must not contain sequential characters (e.g., 123, abc)",
		})
	}

	return errors
}

// CheckPasswordHistory verifies that a new password hasn't been used recently
func CheckPasswordHistory(newPasswordHash string, passwordHistory []string) bool {
	for _, oldHash := range passwordHistory {
		if subtle.ConstantTimeCompare([]byte(newPasswordHash), []byte(oldHash)) == 1 {
			return false // Password found in history
		}
	}
	return true // Password not in history
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GeneratePasswordResetToken generates a secure password reset token
func GeneratePasswordResetToken() (string, error) {
	return GenerateSecureToken(ResetTokenLength)
}

// GenerateEmailVerificationToken generates a secure email verification token
func GenerateEmailVerificationToken() (string, error) {
	return GenerateSecureToken(VerificationTokenLength)
}

// Helper functions for password validation

func hasUppercase(s string) bool {
	for _, c := range s {
		if unicode.IsUpper(c) {
			return true
		}
	}
	return false
}

func hasLowercase(s string) bool {
	for _, c := range s {
		if unicode.IsLower(c) {
			return true
		}
	}
	return false
}

func hasDigit(s string) bool {
	for _, c := range s {
		if unicode.IsDigit(c) {
			return true
		}
	}
	return false
}

func hasSpecialChar(s string) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, c := range s {
		if strings.ContainsRune(specialChars, c) {
			return true
		}
	}
	return false
}

func hasRepeatingChars(s string, maxRepeats int) bool {
	if len(s) < maxRepeats {
		return false
	}

	count := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			count++
			if count >= maxRepeats {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}

func hasSequentialChars(s string) bool {
	if len(s) < 3 {
		return false
	}

	// Check for sequential numbers (123, 789, etc.)
	numPattern := regexp.MustCompile(`(?i)(012|123|234|345|456|567|678|789|890|987|876|765|654|543|432|321|210)`)
	if numPattern.MatchString(s) {
		return true
	}

	// Check for sequential letters (abc, xyz, etc.)
	lowerS := strings.ToLower(s)
	for i := 0; i < len(lowerS)-2; i++ {
		c1, c2, c3 := lowerS[i], lowerS[i+1], lowerS[i+2]
		if c1 >= 'a' && c1 <= 'z' && c2 >= 'a' && c2 <= 'z' && c3 >= 'a' && c3 <= 'z' {
			if (c2 == c1+1 && c3 == c2+1) || (c2 == c1-1 && c3 == c2-1) {
				return true
			}
		}
	}

	return false
}

func isCommonPassword(password string) bool {
	// Convert to lowercase for checking
	lower := strings.ToLower(password)

	// List of common passwords (this should be expanded in production)
	commonPasswords := []string{
		"password", "123456", "password123", "admin", "qwerty", "letmein",
		"welcome", "monkey", "1234567890", "abc123", "111111", "123123",
		"password1", "12345678", "sunshine", "iloveyou", "princess", "admin123",
		"welcome123", "login", "passw0rd", "master", "hello", "guest", "root",
		"test", "user", "demo", "default", "changeme", "secret", "pass",
	}

	for _, common := range commonPasswords {
		if lower == common {
			return true
		}
	}

	// Check for simple patterns - same character repeated 8 or more times
	// Since Go doesn't support backreferences, we'll check this in hasRepeatingChars instead
	if hasRepeatingChars(password, 8) {
		return true
	}

	if regexp.MustCompile(`^(password|admin|user|test|demo)\d*$`).MatchString(lower) { // Common words with numbers
		return true
	}

	return false
}

// CalculatePasswordStrength calculates password strength score (0-100)
func CalculatePasswordStrength(password string) int {
	score := 0

	// Length scoring
	if len(password) >= 8 {
		score += 20
	}
	if len(password) >= 12 {
		score += 10
	}
	if len(password) >= 16 {
		score += 10
	}

	// Character diversity
	if hasLowercase(password) {
		score += 10
	}
	if hasUppercase(password) {
		score += 10
	}
	if hasDigit(password) {
		score += 10
	}
	if hasSpecialChar(password) {
		score += 15
	}

	// Penalty for common patterns
	if isCommonPassword(password) {
		score -= 30
	}
	if hasRepeatingChars(password, 3) {
		score -= 10
	}
	if hasSequentialChars(password) {
		score -= 15
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}
