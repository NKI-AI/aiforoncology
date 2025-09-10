// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/auth"
	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"go.uber.org/zap"
)

func (s *authService) RegisterUser(ctx context.Context, req domain.RegisterUserRequest) error {
	utils.InfoWithKeyValue(ctx, "Starting user registration process",
		"email", req.Email)

	// Validate password strength
	requirements := auth.DefaultPasswordRequirements()
	if validationErrors := auth.ValidatePassword(req.Password, requirements); len(validationErrors) > 0 {
		utils.WarnWithKeyValue(ctx, "Password validation failed during registration",
			"email", req.Email,
			"error", validationErrors[0].Message)
		return fmt.Errorf("%w: %s", errors.ErrPasswordValidation, validationErrors[0].Message)
	}

	// Check if user already exists by email
	if _, err := s.db.GetUserByEmail(ctx, req.Email); err == nil {
		utils.WarnWithKeyValue(ctx, "Registration failed - email already registered",
			"email", req.Email)
		return errors.ErrEmailExists
	}

	// Always use domain-based tenant lookup
	utils.InfoWithKeyValue(ctx, "Using domain-based tenant lookup for registration",
		"email", req.Email)

	// Extract domain from email and look up tenant
	emailDomain, err := extractDomainFromEmail(req.Email)
	if err != nil {
		utils.ErrorWithKeyValue(ctx, "Failed to extract domain from email",
			"email", req.Email,
			"error", err)
		return fmt.Errorf("invalid email format: %w", err)
	}

	utils.InfoWithKeyValue(ctx, "Extracted domain from email, looking up tenant",
		"email", req.Email,
		"domain", emailDomain)

	// Look up tenant by domain
	tenant, err := s.db.GetTenantByDomain(ctx, emailDomain)
	if err != nil {
		utils.ErrorWithKeyValue(ctx, "Failed to find tenant for domain",
			"email", req.Email,
			"domain", emailDomain,
			"error", err)
		return fmt.Errorf("no verified tenant found for domain '%s'. Please contact your administrator to register your organization's domain", emailDomain)
	}

	utils.InfoWithKeyValue(ctx, "Successfully found tenant using domain-based lookup",
		"email", req.Email,
		"domain", emailDomain,
		"tenant_uid", tenant.TenantUID,
		"tenant_name", tenant.Name)

	// Create user (inactive by default)
	user := domain.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  req.Password,
		TenantUID: tenant.TenantUID,
	}

	utils.InfoWithKeyValue(ctx, "Creating user in database",
		"email", req.Email,
		"tenant_uid", tenant.TenantUID)

	createdUser, err := s.userService.CreateUser(ctx, user)
	if err != nil {
		utils.ErrorWithKeyValue(ctx, "Failed to create user in database",
			"email", req.Email,
			"tenant_uid", tenant.TenantUID,
			"error", err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	utils.InfoWithKeyValue(ctx, "User created successfully in database",
		"email", createdUser.Email,
		"tenant_uid", createdUser.TenantUID)

	// Generate email verification token
	token, err := auth.GenerateEmailVerificationToken()
	if err != nil {
		utils.ErrorWithKeyValue(ctx, "Failed to generate email verification token",
			"email", req.Email,
			"error", err)
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	// Token expires in 24 hours
	expiresAt := time.Now().Add(24 * time.Hour)

	// Get the created user's internal ID
	internalUser, err := s.db.GetUserByEmail(ctx, createdUser.Email)
	if err != nil {
		utils.ErrorWithKeyValue(ctx, "Failed to retrieve created user for verification token",
			"email", req.Email,
			"error", err)
		return fmt.Errorf("failed to get created user: %w", err)
	}

	// Store verification token
	if err := s.db.CreateEmailVerificationToken(ctx, internalUser.ID, token, expiresAt); err != nil {
		utils.ErrorWithKeyValue(ctx, "Failed to store email verification token",
			"email", req.Email,
			"user_id", internalUser.ID,
			"error", err)
		return fmt.Errorf("failed to create email verification token: %w", err)
	}

	// Send verification email
	ctxWithTenant := context.WithValue(ctx, "tenantId", tenant.ID)
	if err := s.asyncEmailService.SendEmailVerificationEmailAsync(ctxWithTenant, req.Email, token); err != nil {
		utils.LogError(ctx, "Failed to send email verification email", zap.Error(err), zap.String("email", req.Email))
		// Continue anyway - token was created successfully
	}

	utils.InfoWithKeyValue(ctx, "User registration completed successfully",
		"email", req.Email,
		"tenant_uid", tenant.TenantUID,
		"verification_token_expires_at", expiresAt.Format(time.RFC3339))
	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	// Get and validate token
	var user ports.User
	_, user, err := s.db.GetEmailVerificationToken(ctx, token)
	if err != nil {
		return errors.ErrInvalidVerificationToken
	}

	// Update user as verified and active
	emailVerified := true
	isActive := true
	updates := ports.UserUpdates{
		EmailVerified: &emailVerified,
		IsActive:      &isActive,
	}

	if err := s.db.UpdateUserByUID(ctx, user.ShortUID, updates); err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	// Mark token as used
	if err := s.db.MarkEmailVerificationTokenAsUsed(ctx, token); err != nil {
		utils.LogError(ctx, "Failed to mark verification token as used", zap.Error(err), zap.String("token", token))
	}

	// Send welcome email now that the user is verified and active
	ctxWithTenant := context.WithValue(ctx, "tenantId", user.TenantID)
	if err := s.asyncEmailService.SendWelcomeEmailAsync(ctxWithTenant, user.Email); err != nil {
		utils.LogError(ctx, "Failed to send welcome email", zap.Error(err), zap.String("email", user.Email))
		// Continue anyway - user was already activated
	}

	utils.InfoWithKeyValue(ctx, "Email verified and user activated", "email", user.Email)
	return nil
}

func (s *authService) ResendVerification(ctx context.Context, email string) error {
	// Get user by email
	user, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists - always return success for security
		utils.InfoWithKeyValue(ctx, "Verification resend requested for non-existent email", "email", email)
		return nil
	}

	// Check if user is already verified
	if user.EmailVerified {
		utils.InfoWithKeyValue(ctx, "Verification resend requested for already verified user", "email", email)
		return nil
	}

	// Generate new verification token
	token, err := auth.GenerateEmailVerificationToken()
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	// Token expires in 24 hours
	expiresAt := time.Now().Add(24 * time.Hour)

	// Store new verification token (this will create a new token, old ones will still exist but unused)
	if err := s.db.CreateEmailVerificationToken(ctx, user.ID, token, expiresAt); err != nil {
		return fmt.Errorf("failed to create email verification token: %w", err)
	}

	// Send verification email
	ctxWithTenant := context.WithValue(ctx, "tenantId", user.TenantID)
	if err := s.asyncEmailService.SendEmailVerificationEmailAsync(ctxWithTenant, user.Email, token); err != nil {
		utils.LogError(ctx, "Failed to send email verification email", zap.Error(err), zap.String("email", user.Email))
		// Continue anyway - token was created successfully
	}

	utils.InfoWithKeyValue(ctx, "Verification email resent", "email", email)
	return nil
}

// extractDomainFromEmail extracts the domain part from an email address
func extractDomainFromEmail(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return "", fmt.Errorf("email is empty")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid email format")
	}

	domain := strings.TrimSpace(parts[1])
	if domain == "" {
		return "", fmt.Errorf("domain part is empty")
	}

	return domain, nil
}
