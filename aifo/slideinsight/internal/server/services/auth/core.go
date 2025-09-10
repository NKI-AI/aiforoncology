// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package auth

import (
	"context"
	"time"

	"aifo.dev/aifo/slideinsight/internal/config"
	"aifo.dev/aifo/slideinsight/internal/queue"
	"aifo.dev/aifo/slideinsight/internal/server/auth"
	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/email"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// UserService interface for user operations - avoid circular dependency
type UserService interface {
	VerifyPassword(ctx context.Context, email, password string) error
	VerifyPasswordByEmail(ctx context.Context, email, password string) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	UpdatePassword(ctx context.Context, email, password string) error
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
}

// AuthService is an interface that defines the methods for the auth service.
// Interface is needed for mocking in tests.
type AuthService interface {
	Login(ctx context.Context, email, password string) (domain.TokenResponse, error)
	GenerateJWT(user domain.User) (string, time.Time, error)
	GenerateRefreshJWT(user domain.User) (string, time.Time, error)
	GenerateSwitchedUserJWT(targetUser domain.User, originalUserUID, originalEmail string) (string, time.Time, error)
	ValidateRefreshToken(ctx context.Context, token string) (domain.User, error)
	GetAuthConfig() config.AuthConfig

	// Password management methods
	ChangePassword(ctx context.Context, email string, currentPassword, newPassword string) error
	ForcedChangePassword(ctx context.Context, email string, currentPassword, newPassword string) error
	ResetPassword(ctx context.Context, email string) error
	ResetPasswordConfirm(ctx context.Context, token, newPassword string) error

	// Registration and verification methods
	RegisterUser(ctx context.Context, req domain.RegisterUserRequest) error
	VerifyEmail(ctx context.Context, token string) error
	ResendVerification(ctx context.Context, email string) error

	// Rate limiting
	CheckRateLimit(ctx context.Context, ipAddress, email string) (auth.RateLimitResult, error)
	RecordAuthAttempt(ctx context.Context, ipAddress, email string, success bool, failReason string) error
}

type authService struct {
	db                ports.Database
	userService       UserService
	jwtConfig         config.AuthConfig
	rateLimiter       *auth.RateLimiter
	asyncEmailService email.AsyncEmailService
}

func NewAuthService(db ports.Database, jwtConfig config.AuthConfig, emailService ports.EmailService, userService UserService, queueManager *queue.QueueManager) AuthService {
	// Create async email service for this auth service
	asyncEmailService := email.NewAsyncEmailService(emailService, queueManager)

	return &authService{
		db:                db,
		userService:       userService,
		jwtConfig:         jwtConfig,
		rateLimiter:       auth.NewRateLimiter(db),
		asyncEmailService: asyncEmailService,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (domain.TokenResponse, error) {
	// Verify the user's credentials
	utils.InfoWithKeyValue(ctx, "Verifying password", "email", email)
	err := s.userService.VerifyPasswordByEmail(ctx, email, password)
	if err != nil {
		return domain.TokenResponse{}, errors.ErrInvalidCredentials
	}

	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.TokenResponse{}, err
	}

	// Check if password reset is required
	if user.MustResetPassword {
		utils.InfoWithKeyValue(ctx, "Login attempted for user requiring password reset", "email", email)
		return domain.TokenResponse{}, errors.ErrPasswordResetRequired
	}

	// Check if user account is active
	if !user.IsActive {
		utils.InfoWithKeyValue(ctx, "Login attempted for inactive user", "email", email)
		return domain.TokenResponse{}, errors.ErrAccountInactive
	}

	// Generate access and refresh tokens
	accessToken, accessExpiry, err := s.GenerateJWT(user)
	if err != nil {
		utils.ErrorWithKeyValue(ctx, "Failed to generate access token", "error", err)
		return domain.TokenResponse{}, err
	}

	refreshToken, refreshExpiry, err := s.GenerateRefreshJWT(user)
	if err != nil {
		utils.ErrorWithKeyValue(ctx, "Failed to generate refresh token", "error", err)
		return domain.TokenResponse{}, err
	}

	// Calculate expiration in seconds
	expiresIn := int(time.Until(accessExpiry).Seconds())
	refreshExpiresIn := int(time.Until(refreshExpiry).Seconds())

	return domain.TokenResponse{
		AccessToken:      accessToken,
		TokenType:        "bearer",
		ExpiresIn:        expiresIn,
		RefreshToken:     refreshToken,
		RefreshExpiresIn: refreshExpiresIn,
	}, nil
}

// GetAuthConfig returns the authentication configuration
func (s *authService) GetAuthConfig() config.AuthConfig {
	return s.jwtConfig
}

func (s *authService) CheckRateLimit(ctx context.Context, ipAddress, email string) (auth.RateLimitResult, error) {
	return s.rateLimiter.CheckRateLimit(ctx, ipAddress, email)
}

func (s *authService) RecordAuthAttempt(ctx context.Context, ipAddress, email string, success bool, failReason string) error {
	return s.rateLimiter.RecordAttempt(ctx, ipAddress, email, success, failReason)
}
