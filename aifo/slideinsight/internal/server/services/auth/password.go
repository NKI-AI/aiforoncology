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
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/auth"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"go.uber.org/zap"
)

func (s *authService) ChangePassword(ctx context.Context, email string, currentPassword, newPassword string) error {
	// Verify current password
	err := s.userService.VerifyPassword(ctx, email, currentPassword)
	if err != nil {
		utils.WarnWithKeyValue(ctx, "Password change attempt with invalid current password", "email", email)
		return errors.ErrInvalidCredentials
	}

	// Get user for password history check
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Validate new password strength
	requirements := auth.DefaultPasswordRequirements()
	if validationErrors := auth.ValidatePassword(newPassword, requirements); len(validationErrors) > 0 {
		return fmt.Errorf("%w: %s", errors.ErrPasswordValidation, validationErrors[0].Message)
	}

	// Hash new password
	newPasswordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Check against password history
	passwordHistory, err := s.db.GetPasswordHistory(ctx, user.ID, auth.MaxPasswordHistoryMonths)
	if err != nil {
		return fmt.Errorf("failed to get password history: %w", err)
	}

	var historyHashes []string
	for _, hist := range passwordHistory {
		historyHashes = append(historyHashes, hist.PasswordHash)
	}

	if !auth.CheckPasswordHistory(newPasswordHash, historyHashes) {
		return fmt.Errorf("%w: password has been used recently, please choose a different password", errors.ErrPasswordValidation)
	}

	// Add old password to history before changing
	if err := s.db.AddPasswordToHistory(ctx, user.ID, user.Password); err != nil {
		utils.LogError(ctx, "Failed to add password to history", zap.Error(err), zap.String("email", email))
		// Continue anyway - this is not critical
	}

	// Update password
	if err := s.userService.UpdatePassword(ctx, email, newPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Clear the password reset requirement flag
	mustReset := false
	updates := ports.UserUpdates{
		MustResetPassword: &mustReset,
	}
	if err := s.db.UpdateUserByUID(ctx, user.ShortUID, updates); err != nil {
		utils.LogError(ctx, "Failed to clear MustResetPassword flag", zap.Error(err), zap.String("email", email))
		// Continue anyway - password was already updated
	}

	// Clean up old password history entries
	if err := s.db.CleanupOldPasswordHistory(ctx, user.ID, auth.MaxPasswordHistoryCount); err != nil {
		utils.LogError(ctx, "Failed to cleanup old password history", zap.Error(err), zap.String("email", email))
		// Continue anyway - this is not critical
	}

	utils.InfoWithKeyValue(ctx, "Password changed successfully", "email", email)
	return nil
}

func (s *authService) ForcedChangePassword(ctx context.Context, email string, currentPassword, newPassword string) error {
	// Verify current password
	err := s.userService.VerifyPasswordByEmail(ctx, email, currentPassword)
	if err != nil {
		utils.WarnWithKeyValue(ctx, "Password change attempt with invalid current password", "email", email)
		return errors.ErrInvalidCredentials
	}

	// Get user for password history check
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Validate new password strength
	requirements := auth.DefaultPasswordRequirements()
	if validationErrors := auth.ValidatePassword(newPassword, requirements); len(validationErrors) > 0 {
		return fmt.Errorf("%w: %s", errors.ErrPasswordValidation, validationErrors[0].Message)
	}

	// Hash new password
	newPasswordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Check against password history
	passwordHistory, err := s.db.GetPasswordHistory(ctx, user.ID, auth.MaxPasswordHistoryMonths)
	if err != nil {
		return fmt.Errorf("failed to get password history: %w", err)
	}

	var historyHashes []string
	for _, hist := range passwordHistory {
		historyHashes = append(historyHashes, hist.PasswordHash)
	}

	if !auth.CheckPasswordHistory(newPasswordHash, historyHashes) {
		return fmt.Errorf("%w: password has been used recently, please choose a different password", errors.ErrPasswordValidation)
	}

	// Add old password to history before changing
	if err := s.db.AddPasswordToHistory(ctx, user.ID, user.Password); err != nil {
		utils.LogError(ctx, "Failed to add password to history", zap.Error(err), zap.String("email", email))
	}

	// Update password
	if err := s.userService.UpdatePassword(ctx, user.Email, newPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Clear the password reset requirement flag
	mustReset := false
	updates := ports.UserUpdates{
		MustResetPassword: &mustReset,
	}
	if err := s.db.UpdateUserByUID(ctx, user.ShortUID, updates); err != nil {
		utils.LogError(ctx, "Failed to clear MustResetPassword flag", zap.Error(err), zap.String("email", email))
		// Continue anyway - password was already updated
	}

	// Clean up old password history
	if err := s.db.CleanupOldPasswordHistory(ctx, user.ID, auth.MaxPasswordHistoryCount); err != nil {
		utils.LogError(ctx, "Failed to cleanup old password history", zap.Error(err), zap.String("email", email))
	}

	utils.InfoWithKeyValue(ctx, "Password changed successfully", "email", email)
	return nil
}

func (s *authService) ResetPassword(ctx context.Context, email string) error {
	// Get user by email
	user, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists - always return success for security
		utils.InfoWithKeyValue(ctx, "Password reset requested for non-existent email", "email", email)
		// Send dummy email to prevent user enumeration - use a default tenant ID or skip email
		// We can't determine tenant for non-existent email, so we'll skip sending dummy email
		// This is actually more secure as it doesn't reveal email existence patterns
		return nil
	}

	// Generate secure reset token
	token, err := auth.GeneratePasswordResetToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Token expires in 1 hour
	expiresAt := time.Now().Add(1 * time.Hour)

	// Store token in database
	if err := s.db.CreatePasswordResetToken(ctx, user.ID, token, expiresAt); err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	// Send password reset email
	ctxWithTenant := context.WithValue(ctx, "tenantId", user.TenantID)
	if err := s.asyncEmailService.SendPasswordResetEmailAsync(ctxWithTenant, user.Email, token); err != nil {
		utils.LogError(ctx, "Failed to send password reset email", zap.Error(err), zap.String("email", user.Email))
		// Continue anyway - token was created successfully
	}

	utils.InfoWithKeyValue(ctx, "Password reset token generated", "email", email)
	return nil
}

func (s *authService) ResetPasswordConfirm(ctx context.Context, token, newPassword string) error {
	// Get and validate token
	var user ports.User
	_, user, err := s.db.GetPasswordResetToken(ctx, token)
	if err != nil {
		return errors.ErrInvalidResetToken
	}

	// Validate new password strength
	requirements := auth.DefaultPasswordRequirements()
	if validationErrors := auth.ValidatePassword(newPassword, requirements); len(validationErrors) > 0 {
		return fmt.Errorf("%w: %s", errors.ErrPasswordValidation, validationErrors[0].Message)
	}

	// Hash new password
	newPasswordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Check against password history
	passwordHistory, err := s.db.GetPasswordHistory(ctx, user.ID, auth.MaxPasswordHistoryMonths)
	if err != nil {
		return fmt.Errorf("failed to get password history: %w", err)
	}

	var historyHashes []string
	for _, hist := range passwordHistory {
		historyHashes = append(historyHashes, hist.PasswordHash)
	}

	// Also check current password
	historyHashes = append(historyHashes, user.Password)

	if !auth.CheckPasswordHistory(newPasswordHash, historyHashes) {
		return fmt.Errorf("%w: password has been used recently, please choose a different password", errors.ErrPasswordValidation)
	}

	// Add old password to history
	if err := s.db.AddPasswordToHistory(ctx, user.ID, user.Password); err != nil {
		utils.LogError(ctx, "Failed to add password to history", zap.Error(err), zap.String("email", user.Email))
	}

	// Update password
	if err := s.userService.UpdatePassword(ctx, user.Email, newPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Clear the password reset requirement flag
	mustReset := false
	updates := ports.UserUpdates{
		MustResetPassword: &mustReset,
	}
	if err := s.db.UpdateUserByUID(ctx, user.ShortUID, updates); err != nil {
		utils.LogError(ctx, "Failed to clear MustResetPassword flag", zap.Error(err), zap.String("email", user.Email))
		// Continue anyway - password was already updated
	}

	// Mark token as used
	if err := s.db.MarkPasswordResetTokenAsUsed(ctx, token); err != nil {
		utils.LogError(ctx, "Failed to mark reset token as used", zap.Error(err), zap.String("token", token))
	}

	// Clean up old password history
	if err := s.db.CleanupOldPasswordHistory(ctx, user.ID, auth.MaxPasswordHistoryCount); err != nil {
		utils.LogError(ctx, "Failed to cleanup old password history", zap.Error(err), zap.String("email", user.Email))
	}

	utils.InfoWithKeyValue(ctx, "Password reset completed", "email", user.Email)
	return nil
}
