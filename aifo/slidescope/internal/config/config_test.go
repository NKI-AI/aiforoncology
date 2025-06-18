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

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test a few key defaults
	assert.Equal(t, "3000", cfg.Server.Port)
	assert.Equal(t, "your-secret-key-here", cfg.Auth.JWTSecret)
	assert.Equal(t, true, cfg.Server.CORS.Enabled)
	assert.Equal(t, []string{"*"}, cfg.Server.CORS.AllowOrigins)
}

func TestLoadFromFile(t *testing.T) {
	// Save original environment variables
	origEnv := os.Getenv("PORT")
	defer func() {
		// Restore original environment
		os.Setenv("PORT", origEnv)
	}()

	// Set environment variable for testing
	os.Setenv("PORT", "4000")

	// Load configuration
	cfg := DefaultConfig()
	cfg = MergeWithEnv(cfg)

	// Test that environment variable overrides default
	assert.Equal(t, "4000", cfg.Server.Port)
}

func TestMergeWithEnv(t *testing.T) {
	// Reset environment variables for testing
	envVars := map[string]string{
		"PORT":         "5000",
		"DATABASE_URL": "sqlite://test.db",
	}

	// Save originals and set new values
	originals := make(map[string]string)
	for key, value := range envVars {
		originals[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	// Ensure we restore environment after test
	defer func() {
		for key, value := range originals {
			os.Setenv(key, value)
		}
	}()

	// Create default config and merge with environment
	cfg := DefaultConfig()
	expectedJWTSecret := cfg.Auth.JWTSecret // JWT should not change
	cfg = MergeWithEnv(cfg)

	// Test environment variables were applied for non-JWT settings
	assert.Equal(t, "5000", cfg.Server.Port)
	assert.Equal(t, "sqlite://test.db", cfg.Database.URL)

	// Test that JWT settings are NOT overridden by environment variables
	assert.Equal(t, expectedJWTSecret, cfg.Auth.JWTSecret, "JWT secret should not be overrideable by env vars")
}

func TestLoadFromFileWithConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test config to the file
	testConfig := []byte(`
server:
  port: "5000"
database:
  url: "sqlite://test.db"
`)
	if _, err := tmpFile.Write(testConfig); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Load config from file
	cfg, warnings, err := LoadFromFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Empty(t, warnings)
	assert.Equal(t, "5000", cfg.Server.Port)
	assert.Equal(t, "sqlite://test.db", cfg.Database.URL)
}

func TestValidateConfig(t *testing.T) {
	// Create a config with potential validation issues
	cfg := Config{
		Server: ServerConfig{
			Environment: "production",
		},
		Database: DatabaseConfig{
			URL: "sqlite://localhost:27017/",
		},
		Auth: AuthConfig{
			JWTSecret: "your-secret-key-here", // This should trigger a warning in production
		},
		Storage: StorageConfig{
			SlidesDir: "", // This should trigger a warning
			MasksDir:  "", // This should trigger a warning
		},
	}

	warnings := validateConfig(cfg)

	// We should have 3 warnings
	assert.Len(t, warnings, 3)

	// Check specific warnings
	found := make(map[string]bool)
	for _, w := range warnings {
		found[w.Field] = true
	}

	assert.True(t, found["Auth.JWTSecret"])
	assert.True(t, found["Storage.SlidesDir"])
	assert.True(t, found["Storage.MasksDir"])
}

func TestAuthConfigString(t *testing.T) {
	// Test with default secret
	auth := AuthConfig{
		JWTSecret:                   "your-secret-key-here",
		JWTAlgorithm:                "HS256",
		JWTExpirationMinutes:        5 * time.Minute,
		JWTRefreshExpirationMinutes: 30 * time.Minute,
		Cookie: CookieConfig{
			Name:     "_auth",
			Path:     "/",
			HTTPOnly: true,
			Secure:   false,
			SameSite: "lax",
		},
	}

	str := auth.String()
	assert.Contains(t, str, "[DEFAULT]", "Should redact default secret")
	assert.Contains(t, str, "HS256", "Should include algorithm")
	assert.Contains(t, str, "5m0s", "Should include expiration duration")
	assert.Contains(t, str, "30m0s", "Should include refresh duration")
	assert.Contains(t, str, "_auth", "Should include cookie name")

	// Test with empty secret
	auth.JWTSecret = ""
	str = auth.String()
	assert.Contains(t, str, "[EMPTY]", "Should indicate empty secret")

	// Test with custom secret
	auth.JWTSecret = "custom-secret"
	str = auth.String()
	assert.Contains(t, str, "[REDACTED]", "Should redact custom secret")
	assert.NotContains(t, str, "custom-secret", "Should not expose actual secret")
}
