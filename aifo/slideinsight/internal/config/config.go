// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	Storage  StorageConfig  `yaml:"storage"`
	Logging  LoggingConfig  `yaml:"logging"`
	Email    EmailConfig    `yaml:"email"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port            string     `yaml:"port"`
	Host            string     `yaml:"host"`
	BasePath        string     `yaml:"base_path"`
	Environment     string     `yaml:"environment"`
	CORS            CORSConfig `yaml:"cors"`
	StaticAssetsDir string     `yaml:"static_assets_dir"` // Directory containing static assets
}

// CORSConfig holds CORS-related configuration
type CORSConfig struct {
	Enabled          bool     `yaml:"enabled"`
	AllowOrigins     []string `yaml:"allow_origins"`
	AllowMethods     []string `yaml:"allow_methods"`
	AllowHeaders     []string `yaml:"allow_headers"`
	ExposeHeaders    []string `yaml:"expose_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	URL string `yaml:"url"`
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	JWTSecret                   string        `yaml:"jwt_secret"`
	JWTAlgorithm                string        `yaml:"jwt_algorithm"`
	JWTExpirationMinutes        time.Duration `yaml:"jwt_expiration_minutes"`
	JWTRefreshExpirationMinutes time.Duration `yaml:"jwt_refresh_expiration_minutes"`
	Cookie                      CookieConfig  `yaml:"cookie"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface to handle integer values
// for JWTExpirationMinutes
func (a *AuthConfig) UnmarshalYAML(value *yaml.Node) error {
	// Create an auxiliary struct with the same fields but using int for duration
	type Aux struct {
		JWTSecret                   string       `yaml:"jwt_secret"`
		JWTAlgorithm                string       `yaml:"jwt_algorithm"`
		JWTExpirationMinutes        int          `yaml:"jwt_expiration_minutes"`
		JWTRefreshExpirationMinutes int          `yaml:"jwt_refresh_expiration_minutes"`
		Cookie                      CookieConfig `yaml:"cookie"`
	}

	// Unmarshal into the auxiliary struct
	var aux Aux
	if err := value.Decode(&aux); err != nil {
		return err
	}

	// Convert integer minutes to time.Duration

	a.JWTSecret = aux.JWTSecret
	a.JWTAlgorithm = aux.JWTAlgorithm
	a.JWTExpirationMinutes = time.Duration(aux.JWTExpirationMinutes) * time.Minute
	a.JWTRefreshExpirationMinutes = time.Duration(aux.JWTRefreshExpirationMinutes) * time.Minute
	a.Cookie = aux.Cookie

	return nil
}

// GetJWTSecretBytes returns the JWT secret as a byte slice
func (a *AuthConfig) GetJWTSecretBytes() []byte {
	return []byte(a.JWTSecret)
}

// String provides a custom string representation for AuthConfig that handles
// sensitive data appropriately and avoids truncation issues during logging
func (a *AuthConfig) String() string {
	secretDisplay := "[REDACTED]"
	if a.JWTSecret == "" {
		secretDisplay = "[EMPTY]"
	} else if a.JWTSecret == "your-secret-key-here" {
		secretDisplay = "[DEFAULT]"
	}

	return fmt.Sprintf("AuthConfig{JWTSecret:%s, JWTAlgorithm:%s, JWTExpirationMinutes:%v, JWTRefreshExpirationMinutes:%v, Cookie:{Name:%s, Path:%s, HTTPOnly:%t, Secure:%t, SameSite:%s}}",
		secretDisplay,
		a.JWTAlgorithm,
		a.JWTExpirationMinutes,
		a.JWTRefreshExpirationMinutes,
		a.Cookie.Name,
		a.Cookie.Path,
		a.Cookie.HTTPOnly,
		a.Cookie.Secure,
		a.Cookie.SameSite,
	)
}

// CookieConfig holds cookie-related configuration
type CookieConfig struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	HTTPOnly bool   `yaml:"http_only"`
	Secure   bool   `yaml:"secure"`
	SameSite string `yaml:"same_site"`
}

// StorageConfig holds storage-related configuration
type StorageConfig struct {
	SlidesDir string `yaml:"slides_dir"`
	MasksDir  string `yaml:"masks_dir"`
	CacheSize int    `yaml:"cache_size"`
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level string `yaml:"level"`
}

// EmailConfig holds email-related configuration
type EmailConfig struct {
	Provider string    `yaml:"provider"` // "console" or "ses"
	SES      SESConfig `yaml:"ses"`
}

// SESConfig holds AWS SES-specific configuration
type SESConfig struct {
	Region      string `yaml:"region"`
	FromAddress string `yaml:"from_address"`
	FromName    string `yaml:"from_name"`
}

// ValidationWarning represents a configuration validation warning
type ValidationWarning struct {
	Field   string
	Message string
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Port:        "3000",
			Host:        "localhost",
			BasePath:    "",
			Environment: "development",
			CORS: CORSConfig{
				Enabled:          true,
				AllowOrigins:     []string{"*"},
				AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
				ExposeHeaders:    []string{},
				AllowCredentials: true,
				MaxAge:           86400,
			},
			StaticAssetsDir: "dist",
		},
		Database: DatabaseConfig{
			URL: "sqlite://:memory:",
		},
		Auth: AuthConfig{
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
		},
		Storage: StorageConfig{
			SlidesDir: "./data/slides",
			MasksDir:  "./data/raster",
			CacheSize: 100,
		},
		Logging: LoggingConfig{
			Level: "info",
		},
		Email: EmailConfig{
			Provider: "console",
			SES: SESConfig{
				Region:      "us-east-1",
				FromAddress: "noreply@example.com",
				FromName:    "SlideInsight",
			},
		},
	}
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(filePath string) (Config, []ValidationWarning, error) {
	config := DefaultConfig()

	// If file path is empty, return default config without error
	if filePath == "" {
		return config, validateConfig(config), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return config, nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	warnings := validateConfig(config)
	return config, warnings, nil
}

// validateConfig checks the configuration values and returns warnings
func validateConfig(config Config) []ValidationWarning {
	var warnings []ValidationWarning

	// Check JWT Secret
	if config.Auth.JWTSecret == "your-secret-key-here" &&
		config.Server.Environment == "production" {
		warnings = append(warnings, ValidationWarning{
			Field:   "Auth.JWTSecret",
			Message: "Using default JWT secret in production environment is insecure",
		})
	}

	// Check Database URL
	if config.Database.URL == "" {
		warnings = append(warnings, ValidationWarning{
			Field:   "Database.URL",
			Message: "Database URL is not set",
		})
	}

	// Check Storage directories
	if config.Storage.SlidesDir == "" {
		warnings = append(warnings, ValidationWarning{
			Field:   "Storage.SlidesDir",
			Message: "Slides directory is not set",
		})
	}

	if config.Storage.MasksDir == "" {
		warnings = append(warnings, ValidationWarning{
			Field:   "Storage.MasksDir",
			Message: "Masks directory is not set",
		})
	}

	// Check Email configuration
	if config.Email.Provider == "ses" {
		if config.Email.SES.FromAddress == "" || config.Email.SES.FromAddress == "noreply@example.com" {
			warnings = append(warnings, ValidationWarning{
				Field:   "Email.SES.FromAddress",
				Message: "SES from address should be configured for production use",
			})
		}
		if config.Email.SES.Region == "" {
			warnings = append(warnings, ValidationWarning{
				Field:   "Email.SES.Region",
				Message: "SES region is not set",
			})
		}
	}

	// Add more validation checks as needed

	return warnings
}

// GetEnvBool gets boolean environment variable or returns default if not set
func GetEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	lowerValue := strings.ToLower(value)
	return lowerValue == "true" || lowerValue == "yes" || lowerValue == "1" || lowerValue == "t" || lowerValue == "y"
}

// MergeWithEnv overrides config values with environment variables when available
func MergeWithEnv(config Config) Config {
	// Database URL from environment
	if envDBURL, exists := os.LookupEnv("DATABASE_URL"); exists {
		if envDBURL != "" {
			log.Info("DATABASE_URL environment variable overrides config",
				"env_value", envDBURL,
				"config_value", config.Database.URL)
			config.Database.URL = envDBURL
		} else {
			log.Warn("DATABASE_URL environment variable is empty, ignoring")
		}
	}

	// Server port from environment
	if envPort, exists := os.LookupEnv("PORT"); exists {
		if envPort != "" {
			log.Info("PORT environment variable overrides config",
				"env_value", envPort,
				"config_value", config.Server.Port)
			config.Server.Port = envPort
		} else {
			log.Warn("PORT environment variable is empty, ignoring")
		}
	}

	// JWT configuration is NOT configurable via environment variables for security reasons
	// All JWT settings must be configured via config file or defaults

	// Cookie settings from environment
	if envName, exists := os.LookupEnv("COOKIE_NAME"); exists {
		if envName != "" {
			log.Info("COOKIE_NAME environment variable overrides config",
				"env_value", envName,
				"config_value", config.Auth.Cookie.Name)
			config.Auth.Cookie.Name = envName
		} else {
			log.Warn("COOKIE_NAME environment variable is empty, ignoring")
		}
	}

	if envPath, exists := os.LookupEnv("COOKIE_PATH"); exists {
		if envPath != "" {
			log.Info("COOKIE_PATH environment variable overrides config",
				"env_value", envPath,
				"config_value", config.Auth.Cookie.Path)
			config.Auth.Cookie.Path = envPath
		} else {
			log.Warn("COOKIE_PATH environment variable is empty, ignoring")
		}
	}

	// Boolean cookie settings - always apply, but log when they override
	originalHTTPOnly := config.Auth.Cookie.HTTPOnly
	config.Auth.Cookie.HTTPOnly = GetEnvBool("COOKIE_HTTPONLY", config.Auth.Cookie.HTTPOnly)
	if httpOnlyEnv, exists := os.LookupEnv("COOKIE_HTTPONLY"); exists {
		log.Info("COOKIE_HTTPONLY environment variable processed",
			"env_value", httpOnlyEnv,
			"old_value", originalHTTPOnly,
			"new_value", config.Auth.Cookie.HTTPOnly)
	}

	originalSecure := config.Auth.Cookie.Secure
	config.Auth.Cookie.Secure = GetEnvBool("COOKIE_SECURE", config.Auth.Cookie.Secure)
	if secureEnv, exists := os.LookupEnv("COOKIE_SECURE"); exists {
		log.Info("COOKIE_SECURE environment variable processed",
			"env_value", secureEnv,
			"old_value", originalSecure,
			"new_value", config.Auth.Cookie.Secure)
	}

	if envSameSite, exists := os.LookupEnv("COOKIE_SAMESITE"); exists {
		if envSameSite != "" {
			log.Info("COOKIE_SAMESITE environment variable overrides config",
				"env_value", envSameSite,
				"config_value", config.Auth.Cookie.SameSite)
			config.Auth.Cookie.SameSite = envSameSite
		} else {
			log.Warn("COOKIE_SAMESITE environment variable is empty, ignoring")
		}
	}

	return config
}
