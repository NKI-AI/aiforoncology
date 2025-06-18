// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Note: A compatibility test between old and new JWT config is available
// but skipped to avoid circular imports. Run it manually if needed.

func TestJWTEnvVars(t *testing.T) {
	// Save original env vars to restore after test
	originalSecret := os.Getenv("SLIDESCOPE_JWT_SECRET")
	originalAlgorithm := os.Getenv("SLIDESCOPE_JWT_ALGORITHM")
	originalExpiration := os.Getenv("SLIDESCOPE_JWT_EXPIRATION_MINUTES")
	originalRefresh := os.Getenv("SLIDESCOPE_JWT_REFRESH_EXPIRATION_MINUTES")
	defer func() {
		os.Setenv("SLIDESCOPE_JWT_SECRET", originalSecret)
		os.Setenv("SLIDESCOPE_JWT_ALGORITHM", originalAlgorithm)
		os.Setenv("SLIDESCOPE_JWT_EXPIRATION_MINUTES", originalExpiration)
		os.Setenv("SLIDESCOPE_JWT_REFRESH_EXPIRATION_MINUTES", originalRefresh)
	}()

	// Set test environment variables
	os.Setenv("SLIDESCOPE_JWT_SECRET", "test-secret-value")
	os.Setenv("SLIDESCOPE_JWT_ALGORITHM", "RS256")
	os.Setenv("SLIDESCOPE_JWT_EXPIRATION_MINUTES", "120")
	os.Setenv("SLIDESCOPE_JWT_REFRESH_EXPIRATION_MINUTES", "120")

	// Create base config with expected values
	config := DefaultConfig()
	expectedSecret := config.Auth.JWTSecret
	expectedAlgorithm := config.Auth.JWTAlgorithm
	expectedExpiration := config.Auth.JWTExpirationMinutes
	expectedRefresh := config.Auth.JWTRefreshExpirationMinutes

	// Apply environment variables
	config = MergeWithEnv(config)

	// Check that the environment variables were NOT applied (for security)
	assert.Equal(t, expectedSecret, config.Auth.JWTSecret, "JWT secret should NOT be overrideable by env var")
	assert.Equal(t, expectedAlgorithm, config.Auth.JWTAlgorithm, "JWT algorithm should NOT be overrideable by env var")
	assert.Equal(t, expectedExpiration, config.Auth.JWTExpirationMinutes, "JWT expiration should NOT be overrideable by env var")
	assert.Equal(t, expectedRefresh, config.Auth.JWTRefreshExpirationMinutes, "Refresh expiration should NOT be overrideable by env var")
}

func TestCookieEnvVars(t *testing.T) {
	// Save original env vars to restore after test
	originalName := os.Getenv("COOKIE_NAME")
	originalPath := os.Getenv("COOKIE_PATH")
	originalHTTPOnly := os.Getenv("COOKIE_HTTPONLY")
	originalSecure := os.Getenv("COOKIE_SECURE")
	originalSameSite := os.Getenv("COOKIE_SAMESITE")
	defer func() {
		os.Setenv("COOKIE_NAME", originalName)
		os.Setenv("COOKIE_PATH", originalPath)
		os.Setenv("COOKIE_HTTPONLY", originalHTTPOnly)
		os.Setenv("COOKIE_SECURE", originalSecure)
		os.Setenv("COOKIE_SAMESITE", originalSameSite)
	}()

	// Set test environment variables
	os.Setenv("COOKIE_NAME", "test_cookie")
	os.Setenv("COOKIE_PATH", "/api")
	os.Setenv("COOKIE_HTTPONLY", "false")
	os.Setenv("COOKIE_SECURE", "true")
	os.Setenv("COOKIE_SAMESITE", "strict")

	// Create base config
	config := DefaultConfig()

	// Apply environment variables
	config = MergeWithEnv(config)

	// Check that the environment variables were applied
	assert.Equal(t, "test_cookie", config.Auth.Cookie.Name, "Cookie name should be loaded from env var")
	assert.Equal(t, "/api", config.Auth.Cookie.Path, "Cookie path should be loaded from env var")
	assert.False(t, config.Auth.Cookie.HTTPOnly, "Cookie HTTPOnly should be loaded from env var")
	assert.True(t, config.Auth.Cookie.Secure, "Cookie Secure should be loaded from env var")
	assert.Equal(t, "strict", config.Auth.Cookie.SameSite, "Cookie SameSite should be loaded from env var")
}

func TestInvalidJWTExpirationEnvVar(t *testing.T) {
	// This test is no longer relevant since JWT settings are not configurable via environment variables
	// for security reasons. JWT configuration must be done via config file only.
	t.Skip("JWT settings are no longer configurable via environment variables for security reasons")
}

func TestGetEnvHelperFunctions(t *testing.T) {
	// Test getEnvOrDefault
	assert.Equal(t, "default", GetEnvOrDefault("NONEXISTENT_ENV_VAR", "default"),
		"Default value should be returned when env var is not set")

	// Test getEnvBool
	assert.True(t, GetEnvBool("NONEXISTENT_ENV_VAR", true),
		"Default boolean value should be returned when env var is not set")

	// Test various boolean string representations
	testCases := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"yes", true},
		{"YES", true},
		{"1", true},
		{"t", true},
		{"y", true},
		{"false", false},
		{"no", false},
		{"0", false},
		{"random", false},
	}

	for _, tc := range testCases {
		os.Setenv("TEST_BOOL_VAR", tc.value)
		assert.Equal(t, tc.expected, GetEnvBool("TEST_BOOL_VAR", !tc.expected),
			"Boolean env var %s should be parsed correctly", tc.value)
	}
	os.Unsetenv("TEST_BOOL_VAR")

	// Test getEnvDuration
	assert.Equal(t, 30*time.Minute, GetEnvDuration("NONEXISTENT_ENV_VAR", 30),
		"Default duration should be returned when env var is not set")

	os.Setenv("TEST_DURATION_VAR", "45")
	assert.Equal(t, 45*time.Minute, GetEnvDuration("TEST_DURATION_VAR", 30),
		"Duration should be parsed from env var when set")

	os.Setenv("TEST_DURATION_VAR", "not-a-number")
	assert.Equal(t, 30*time.Minute, GetEnvDuration("TEST_DURATION_VAR", 30),
		"Default duration should be used when env var is invalid")
	os.Unsetenv("TEST_DURATION_VAR")
}

func TestGetJWTSecretBytes(t *testing.T) {
	config := DefaultConfig()
	config.Auth.JWTSecret = "test-secret"

	secretBytes := config.Auth.GetJWTSecretBytes()
	assert.Equal(t, []byte("test-secret"), secretBytes, "GetJWTSecretBytes should return the secret as bytes")
}
