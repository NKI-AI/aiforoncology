// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"fmt"
	"time"

	"aifo.dev/aifo/slideinsight/internal/config"
	"aifo.dev/aifo/slideinsight/internal/server/auth"
	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	authService "aifo.dev/aifo/slideinsight/internal/server/services/auth"
	"aifo.dev/aifo/slideinsight/internal/server/validation"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// requestInfo holds common request information for logging
type requestInfo struct {
	ID       interface{}
	IP       string
	Endpoint string
}

// getRequestInfo extracts common request information
func getRequestInfo(c *fiber.Ctx, endpoint string) requestInfo {
	return requestInfo{
		ID:       c.Locals("requestid"),
		IP:       c.IP(),
		Endpoint: endpoint,
	}
}

// maskToken returns a masked version of the token for safe logging
func maskToken(token string) string {
	if len(token) < 10 {
		return "***"
	}
	return token[:5] + "..." + token[len(token)-5:]
}

// setAuthCookies sets both access and refresh cookies
func setAuthCookies(c *fiber.Ctx, authConfig config.AuthConfig, tokenResp *domain.TokenResponse) {
	sameSite := auth.SameSiteFromConfig(authConfig.Cookie.SameSite)

	// Set access token cookie
	c.Cookie(&fiber.Cookie{
		Name:     authConfig.Cookie.Name,
		Value:    tokenResp.AccessToken,
		Path:     authConfig.Cookie.Path,
		MaxAge:   tokenResp.ExpiresIn,
		HTTPOnly: authConfig.Cookie.HTTPOnly,
		Secure:   authConfig.Cookie.Secure,
		SameSite: sameSite,
	})

	// Set refresh token cookie
	c.Cookie(&fiber.Cookie{
		Name:     authConfig.Cookie.Name + "_refresh",
		Value:    tokenResp.RefreshToken,
		Path:     authConfig.Cookie.Path,
		MaxAge:   tokenResp.RefreshExpiresIn,
		HTTPOnly: authConfig.Cookie.HTTPOnly,
		Secure:   authConfig.Cookie.Secure,
		SameSite: sameSite,
	})
}

// clearAuthCookies clears both access and refresh cookies
func clearAuthCookies(c *fiber.Ctx, authConfig config.AuthConfig) {
	sameSite := auth.SameSiteFromConfig(authConfig.Cookie.SameSite)
	expiry := time.Now().Add(-time.Hour)

	// Clear access token cookie
	c.Cookie(&fiber.Cookie{
		Name:     authConfig.Cookie.Name,
		Value:    "",
		Path:     authConfig.Cookie.Path,
		Expires:  expiry,
		HTTPOnly: authConfig.Cookie.HTTPOnly,
		Secure:   authConfig.Cookie.Secure,
		SameSite: sameSite,
	})

	// Clear refresh token cookie
	c.Cookie(&fiber.Cookie{
		Name:     authConfig.Cookie.Name + "_refresh",
		Value:    "",
		Path:     authConfig.Cookie.Path,
		Expires:  expiry,
		HTTPOnly: authConfig.Cookie.HTTPOnly,
		Secure:   authConfig.Cookie.Secure,
		SameSite: sameSite,
	})
}

// Login handles user authentication and sets a JWT token in a cookie
// @Summary User login
// @Description Authenticate user credentials and return JWT tokens in cookies
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body LoginInput true "User credentials (email and password)"
// @Success 200 {object} domain.TokenResponse "Authentication successful"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized - invalid credentials"
// @Failure 429 {object} domain.ErrorResponse "Too many requests - rate limited"
// @Router /api/v1/auth/login [post]
func Login(authService authService.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.WithContext(c.UserContext()).Info("Processing login request")

		var input LoginInput
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("Login request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the login input struct using modern validation
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		log.WithContext(c.UserContext()).Info("Attempting authentication", "email", input.Email)
		clientIP := c.Locals("client_ip").(string)

		// Authenticate the user and get token response
		tokenResp, err := authService.Login(c.UserContext(), input.Email, input.Password)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Authentication failed",
				"error", err,
				"email", input.Email,
				"error_type", fmt.Sprintf("%T", err))

			// Handle specific error types using proper error checking
			switch {
			case errors.IsPasswordResetRequired(err):
				authService.RecordAuthAttempt(c.UserContext(), clientIP, input.Email, false, "password_reset_required")
				return middleware.HandleError(c, err)
			case errors.IsAccountInactive(err):
				authService.RecordAuthAttempt(c.UserContext(), clientIP, input.Email, false, "account_inactive")
				return middleware.HandleError(c, err)
			default:
				authService.RecordAuthAttempt(c.UserContext(), clientIP, input.Email, false, "invalid_credentials")
				return middleware.HandleError(c, err)
			}
		}

		authService.RecordAuthAttempt(c.UserContext(), clientIP, input.Email, true, "")

		// Set cookies
		authConfig := authService.GetAuthConfig()
		setAuthCookies(c, authConfig, &tokenResp)

		// Log success with expiry times
		accessExpiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		refreshExpiry := time.Now().Add(time.Duration(tokenResp.RefreshExpiresIn) * time.Second)

		log.WithContext(c.UserContext()).Info("Authentication successful",
			"email", input.Email,
			"access_token", maskToken(tokenResp.AccessToken),
			"refresh_token", maskToken(tokenResp.RefreshToken),
			"access_expires_at", accessExpiry.Format(time.RFC3339),
			"refresh_expires_at", refreshExpiry.Format(time.RFC3339))

		return c.JSON(tokenResp)
	}
}

// Logout clears the authentication cookie
// @Summary User logout
// @Description Clear authentication cookies and log out the user
// @Tags authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Logout successful"
// @Router /api/v1/auth/logout [post]
func Logout(authService authService.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract email from principal for logging (optional)
		var email string
		if principal := c.Locals("principal"); principal != nil {
			if p, ok := principal.(*middleware.Principal); ok && p != nil {
				email = p.Email
			}
		}

		log.WithContext(c.UserContext()).Info("Processing logout request", "email", email)

		// Clear cookies
		authConfig := authService.GetAuthConfig()
		clearAuthCookies(c, authConfig)

		log.WithContext(c.UserContext()).Info("Logout successful", "email", email)

		return c.JSON(fiber.Map{"message": "Successfully logged out"})
	}
}

// Refresh issues new access and refresh tokens using the refresh token cookie
// @Summary Refresh authentication tokens
// @Description Use refresh token to generate new access and refresh tokens
// @Tags authentication
// @Produce json
// @Success 200 {object} domain.TokenResponse "Token refresh successful"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized - invalid refresh token"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/auth/refresh [post]
func Refresh(authService authService.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.WithContext(c.UserContext()).Info("Processing token refresh request")

		authConfig := authService.GetAuthConfig()
		refreshToken := c.Cookies(authConfig.Cookie.Name + "_refresh")
		if refreshToken == "" {
			log.WithContext(c.UserContext()).Warn("Refresh attempt with missing token")
			return middleware.SendError(c, fiber.StatusUnauthorized, "missing refresh token")
		}

		log.WithContext(c.UserContext()).Info("Validating refresh token", "token", maskToken(refreshToken))

		// Validate refresh token
		user, err := authService.ValidateRefreshToken(c.UserContext(), refreshToken)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Refresh token validation failed",
				"error", err,
				"token", maskToken(refreshToken),
				"error_type", fmt.Sprintf("%T", err))
			return middleware.SendError(c, fiber.StatusUnauthorized, "invalid refresh token")
		}

		log.WithContext(c.UserContext()).Info("Refresh token valid, generating new tokens", "email", user.Email)

		// Generate new tokens
		accessToken, accessExp, err := authService.GenerateJWT(user)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to generate access token",
				"error", err,
				"email", user.Email,
				"error_type", fmt.Sprintf("%T", err))
			return middleware.SendError(c, fiber.StatusInternalServerError, "could not refresh token")
		}

		refreshTokenNew, refreshExp, err := authService.GenerateRefreshJWT(user)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to generate refresh token",
				"error", err,
				"email", user.Email,
				"error_type", fmt.Sprintf("%T", err))
			return middleware.SendError(c, fiber.StatusInternalServerError, "could not refresh token")
		}

		tokenResp := &domain.TokenResponse{
			AccessToken:      accessToken,
			TokenType:        "bearer",
			ExpiresIn:        int(time.Until(accessExp).Seconds()),
			RefreshToken:     refreshTokenNew,
			RefreshExpiresIn: int(time.Until(refreshExp).Seconds()),
		}

		// Set cookies
		setAuthCookies(c, authConfig, tokenResp)

		log.WithContext(c.UserContext()).Info("Token refresh successful",
			"email", user.Email,
			"access_token", maskToken(tokenResp.AccessToken),
			"refresh_token", maskToken(tokenResp.RefreshToken),
			"access_expires_at", accessExp.Format(time.RFC3339),
			"refresh_expires_at", refreshExp.Format(time.RFC3339))

		return c.JSON(tokenResp)
	}
}

// GetCurrentUser returns information about the currently authenticated user
// @Summary Get current user information
// @Description Retrieve information about the currently authenticated user
// @Tags authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Current user information"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Router /api/v1/auth/me [get]
func GetCurrentUser(authService authService.AuthService, db ports.Database) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// The middleware has validated the token and set the principal in locals
		principal := c.Locals("principal")
		if principal == nil {
			log.WithContext(c.UserContext()).Warn("Attempt to access user info without authentication")
			return middleware.SendError(c, fiber.StatusUnauthorized, "not authenticated")
		}

		// Extract email from the principal
		p, ok := principal.(*middleware.Principal)
		if !ok || p == nil {
			log.WithContext(c.UserContext()).Warn("Invalid principal object in request context")
			return middleware.SendError(c, fiber.StatusUnauthorized, "not authenticated")
		}

		// Initialize default response with basic user info from token
		response := fiber.Map{
			"email":   p.Email,
			"userUid": p.UserUID,
			"scopes":  []string{}, // Default empty scopes
		}

		// Add switched session information if applicable
		if p.OriginalUserUID != "" {
			response["isSwitched"] = true
			response["originalUserUid"] = p.OriginalUserUID
			response["originalEmail"] = p.OriginalEmail
		} else {
			response["isSwitched"] = false
		}

		// Try to get additional user info and roles, but don't fail if this doesn't work
		// This makes the endpoint more resilient to database issues
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.WithContext(c.UserContext()).Warn("Recovered from panic during role check", "panic", r, "email", p.Email)
				}
			}()

			// Get user from database to check superadmin status
			userService := services.NewUserService(db)
			user, err := userService.GetInternalUserByEmail(c.UserContext(), p.Email)
			if err != nil {
				log.WithContext(c.UserContext()).Warn("Failed to get user for role check", "error", err, "email", p.Email)
				return // Continue with basic info
			}

			// Check if user has superadmin role
			isSuper, err := db.UserHasGlobalRole(c.UserContext(), user.ID, "superadmin")
			if err != nil {
				log.WithContext(c.UserContext()).Warn("Failed to check superadmin role", "error", err, "email", p.Email)
				return // Continue with basic info
			}

			if isSuper {
				response["scopes"] = []string{"superadmin"}
			}
		}()

		log.WithContext(c.UserContext()).Info("Retrieved current user", "email", p.Email, "userUid", p.UserUID, "scopes", response["scopes"])

		return c.JSON(response)
	}
}

// ChangePassword handles user password changes
// @Summary Change user password
// @Description Change the current user's password
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.ChangePasswordRequest true "Change password request"
// @Success 200 {object} map[string]string "Password changed successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 422 {object} domain.ErrorResponse "Password validation failed"
// @Router /api/v1/auth/change-password [post]
func ChangePassword(authService authService.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get email from auth middleware principal
		principal := c.Locals("principal")
		if principal == nil {
			log.WithContext(c.UserContext()).Warn("Change password attempt without authentication")
			return middleware.SendError(c, fiber.StatusUnauthorized, "not authenticated")
		}

		p, ok := principal.(*middleware.Principal)
		if !ok || p == nil {
			log.WithContext(c.UserContext()).Warn("Invalid principal object in change password request")
			return middleware.SendError(c, fiber.StatusUnauthorized, "not authenticated")
		}

		email := p.Email

		var input domain.ChangePasswordRequest
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("Change password request parsing failed", "error", err, "email", p.Email)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the change password input struct using modern validation
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		log.WithContext(c.UserContext()).Info("Processing password change request", "email", p.Email)

		err := authService.ChangePassword(c.UserContext(), email, input.CurrentPassword, input.NewPassword)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Password change failed", "error", err, "email", p.Email)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Password change successful", "email", p.Email)
		return c.JSON(fiber.Map{"message": "Password changed successfully"})
	}
}

// ResetPassword initiates password reset process
// @Summary Request password reset
// @Description Send password reset email to user
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body domain.ResetPasswordRequest true "Reset password request"
// @Success 200 {object} map[string]string "Reset email sent"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Router /api/v1/auth/reset-password [post]
func ResetPassword(authService authService.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.WithContext(c.UserContext()).Info("Processing password reset request")

		var input domain.ResetPasswordRequest
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("Reset password request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the reset password input struct using modern validation
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		log.WithContext(c.UserContext()).Info("Initiating password reset", "email", input.Email)

		err := authService.ResetPassword(c.UserContext(), input.Email)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Password reset initiation failed", "error", err, "email", input.Email)
			// Don't reveal the error to prevent user enumeration
		}

		log.WithContext(c.UserContext()).Info("Password reset response sent", "email", input.Email)
		return c.JSON(fiber.Map{"message": "If an account with this email exists, you will receive password reset instructions."})
	}
}

// ResetPasswordConfirm completes password reset process
// @Summary Confirm password reset
// @Description Reset password using token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body domain.ResetPasswordConfirmRequest true "Reset password confirmation"
// @Success 200 {object} map[string]string "Password reset successful"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Invalid or expired token"
// @Failure 422 {object} domain.ErrorResponse "Password validation failed"
// @Router /api/v1/auth/reset-password/confirm [post]
func ResetPasswordConfirm(authService authService.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.WithContext(c.UserContext()).Info("Processing password reset confirmation")

		var input domain.ResetPasswordConfirmRequest
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("Reset password confirm request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the reset password confirm input struct using modern validation
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		log.WithContext(c.UserContext()).Info("Confirming password reset", "token", maskToken(input.Token))

		err := authService.ResetPasswordConfirm(c.UserContext(), input.Token, input.NewPassword)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Password reset confirmation failed", "error", err, "token", maskToken(input.Token))
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Password reset confirmation successful", "token", maskToken(input.Token))
		return c.JSON(fiber.Map{"message": "Password reset successfully"})
	}
}

// RegisterUser handles new user registration
// @Summary Register new user
// @Description Register a new user account
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body domain.RegisterUserRequest true "User registration request"
// @Success 201 {object} map[string]string "Registration successful"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 409 {object} domain.ErrorResponse "Email already exists"
// @Failure 422 {object} domain.ErrorResponse "Validation failed"
// @Router /api/v1/auth/register [post]
func RegisterUser(authService authService.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.WithContext(c.UserContext()).Info("Processing user registration request")

		var input domain.RegisterUserRequest
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("User registration request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the register user input struct using modern validation
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		log.WithContext(c.UserContext()).Info("Registering new user", "email", input.Email)

		err := authService.RegisterUser(c.UserContext(), input)
		if err != nil {
			log.WithContext(c.UserContext()).Error("User registration failed", "error", err, "email", input.Email)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("User registration successful", "email", input.Email)
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Registration successful. Please check your email for verification instructions.",
		})
	}
}

// VerifyEmail handles email verification
// @Summary Verify email address
// @Description Verify user email using verification token
// @Tags authentication
// @Accept json
// @Produce json
// @Param token query string true "Email verification token"
// @Success 200 {object} map[string]string "Email verified successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - missing token"
// @Failure 401 {object} domain.ErrorResponse "Invalid or expired token"
// @Router /api/v1/auth/verify-email [get]
func VerifyEmail(authService authService.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.WithContext(c.UserContext()).Info("Processing email verification request")

		token := c.Query("token")
		if token == "" {
			log.WithContext(c.UserContext()).Info("Email verification attempt with missing token")
			return middleware.SendError(c, fiber.StatusBadRequest, "verification token is required")
		}

		log.WithContext(c.UserContext()).Info("Verifying email", "token", maskToken(token))

		err := authService.VerifyEmail(c.UserContext(), token)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Email verification failed", "error", err, "token", maskToken(token))
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Email verification successful", "token", maskToken(token))
		return c.JSON(fiber.Map{"message": "Email verified successfully. Your account is now active."})
	}
}

// ResendVerification handles resending email verification
// @Summary Resend email verification
// @Description Send a new email verification token to user
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body domain.ResendVerificationRequest true "Resend verification request"
// @Success 200 {object} map[string]string "Verification email sent"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Router /api/v1/auth/resend-verification [post]
func ResendVerification(authService authService.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.WithContext(c.UserContext()).Info("Processing resend verification request")

		var input domain.ResendVerificationRequest
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("Resend verification request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the resend verification input struct using modern validation
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		log.WithContext(c.UserContext()).Info("Initiating verification resend", "email", input.Email)

		err := authService.ResendVerification(c.UserContext(), input.Email)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Verification resend failed", "error", err, "email", input.Email)
			// Don't reveal the error to prevent user enumeration
		}

		log.WithContext(c.UserContext()).Info("Verification resend response sent", "email", input.Email)
		return c.JSON(fiber.Map{"message": "If an unverified account with this email exists, you will receive a new verification email."})
	}
}

// ForcedChangePassword handles password changes for users who must reset their password before authentication
// @Summary Forced password change without authentication
// @Description Change password for users who must reset their password (no JWT required)
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body domain.ForcedChangePasswordRequest true "Forced change password request"
// @Success 200 {object} map[string]string "Password changed successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Invalid credentials"
// @Failure 422 {object} domain.ErrorResponse "Password validation failed"
// @Router /api/v1/auth/forced-change-password [post]
func ForcedChangePassword(authService authService.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.WithContext(c.UserContext()).Info("Processing forced password change request")

		var input domain.ForcedChangePasswordRequest
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("Forced change password request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the forced change password input struct using modern validation
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		log.WithContext(c.UserContext()).Info("Processing forced password change request", "email", input.Email)

		err := authService.ForcedChangePassword(c.UserContext(), input.Email, input.CurrentPassword, input.NewPassword)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Forced password change failed", "error", err, "email", input.Email)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Forced password change successful", "email", input.Email)
		return c.JSON(fiber.Map{"message": "Password changed successfully"})
	}
}

// SwitchUser allows admin users to switch to another user for impersonation
// @Summary Switch to another user
// @Description Switch to view the platform as another user (admin only)
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.SwitchUserRequest true "Switch user request"
// @Success 200 {object} domain.TokenResponse "New token for switched user"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 403 {object} domain.ErrorResponse "Forbidden - platform.admin required"
// @Failure 404 {object} domain.ErrorResponse "Target user not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/auth/switch-user [post]
func SwitchUser(authService authService.AuthService, db ports.Database) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the principal from the JWT middleware
		principal := c.Locals("principal")
		if principal == nil {
			log.WithContext(c.UserContext()).Warn("Switch user attempted without authentication")
			return middleware.SendError(c, fiber.StatusUnauthorized, "authentication required")
		}

		p, ok := principal.(*middleware.Principal)
		if !ok || p == nil {
			log.WithContext(c.UserContext()).Warn("Invalid principal object in switch user request")
			return middleware.SendError(c, fiber.StatusUnauthorized, "invalid authentication")
		}

		// Check if user has platform.admin permission
		userID, err := db.GetUserIDByUID(c.UserContext(), p.UserUID)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to get user ID for permission check",
				"error", err,
				"userUID", p.UserUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "permission check failed")
		}

		hasPermission, err := db.UserHasRolePermission(c.UserContext(), userID, "platform.admin")
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to check platform.admin permission",
				"error", err,
				"userID", userID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "permission check failed")
		}

		if !hasPermission {
			log.WithContext(c.UserContext()).Warn("Switch user attempted without platform.admin permission",
				"userUID", p.UserUID)
			return middleware.SendError(c, fiber.StatusForbidden, "platform.admin permission required")
		}

		// Parse the request
		var input domain.SwitchUserRequest
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("Switch user request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		log.WithContext(c.UserContext()).Info("Processing switch user request",
			"adminEmail", p.Email,
			"targetUserUID", input.TargetUserUID)

		// Get the target user
		targetUser, err := db.GetUserByUID(c.UserContext(), input.TargetUserUID)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to get target user for switch",
				"error", err,
				"targetUserUID", input.TargetUserUID)
			return middleware.SendError(c, fiber.StatusNotFound, "target user not found")
		}

		// Check if target user is active
		if !targetUser.IsActive {
			log.WithContext(c.UserContext()).Warn("Switch user attempted to inactive user",
				"targetUserUID", input.TargetUserUID)
			return middleware.SendError(c, fiber.StatusBadRequest, "cannot switch to inactive user")
		}

		// Convert ports.User to domain.User
		targetDomainUser := domain.User{
			ID:                targetUser.ID,
			TenantID:          targetUser.TenantID,
			TenantUID:         targetUser.TenantUID,
			ShortUID:          targetUser.ShortUID,
			Email:             targetUser.Email,
			FirstName:         targetUser.FirstName,
			LastName:          targetUser.LastName,
			Password:          targetUser.Password,
			MustResetPassword: targetUser.MustResetPassword,
			IsActive:          targetUser.IsActive,
			EmailVerified:     targetUser.EmailVerified,
			CreatedAt:         targetUser.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         targetUser.UpdatedAt.Format(time.RFC3339),
		}

		// Generate new JWT with the target user info but include original admin info
		accessToken, accessExpiry, err := authService.GenerateSwitchedUserJWT(targetDomainUser, p.UserUID, p.Email)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to generate switched user token",
				"error", err,
				"targetUserUID", input.TargetUserUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "could not generate switch token")
		}

		// Generate refresh token for the switched user
		refreshToken, refreshExpiry, err := authService.GenerateRefreshJWT(targetDomainUser)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to generate refresh token for switched user",
				"error", err,
				"targetUserUID", input.TargetUserUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "could not generate refresh token")
		}

		tokenResp := &domain.TokenResponse{
			AccessToken:      accessToken,
			TokenType:        "bearer",
			ExpiresIn:        int(time.Until(accessExpiry).Seconds()),
			RefreshToken:     refreshToken,
			RefreshExpiresIn: int(time.Until(refreshExpiry).Seconds()),
		}

		// Set cookies
		authConfig := authService.GetAuthConfig()
		setAuthCookies(c, authConfig, tokenResp)

		log.WithContext(c.UserContext()).Info("Switch user successful",
			"adminEmail", p.Email,
			"targetUserUID", input.TargetUserUID,
			"targetUserEmail", targetUser.Email)

		return c.JSON(tokenResp)
	}
}

// SwitchBack allows switched users to return to their original identity
// @Summary Switch back to original user
// @Description Return to original admin identity after user switching
// @Tags authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.TokenResponse "New token for original user"
// @Failure 400 {object} domain.ErrorResponse "Bad request - not in switched mode"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/auth/switch-back [post]
func SwitchBack(authService authService.AuthService, db ports.Database) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the principal from the JWT middleware
		principal := c.Locals("principal")
		if principal == nil {
			log.WithContext(c.UserContext()).Warn("Switch back attempted without authentication")
			return middleware.SendError(c, fiber.StatusUnauthorized, "authentication required")
		}

		p, ok := principal.(*middleware.Principal)
		if !ok || p == nil {
			log.WithContext(c.UserContext()).Warn("Invalid principal object in switch back request")
			return middleware.SendError(c, fiber.StatusUnauthorized, "invalid authentication")
		}

		// Check if this is a switched user session
		if p.OriginalUserUID == "" {
			log.WithContext(c.UserContext()).Warn("Switch back attempted but not in switched mode",
				"userUID", p.UserUID)
			return middleware.SendError(c, fiber.StatusBadRequest, "not in switched user mode")
		}

		log.WithContext(c.UserContext()).Info("Processing switch back request",
			"currentUserUID", p.UserUID,
			"originalUserUID", p.OriginalUserUID)

		// Get the original admin user
		originalUser, err := db.GetUserByUID(c.UserContext(), p.OriginalUserUID)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to get original user for switch back",
				"error", err,
				"originalUserUID", p.OriginalUserUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "failed to switch back")
		}

		// Convert ports.User to domain.User
		originalDomainUser := domain.User{
			ID:                originalUser.ID,
			TenantID:          originalUser.TenantID,
			TenantUID:         originalUser.TenantUID,
			ShortUID:          originalUser.ShortUID,
			Email:             originalUser.Email,
			FirstName:         originalUser.FirstName,
			LastName:          originalUser.LastName,
			Password:          originalUser.Password,
			MustResetPassword: originalUser.MustResetPassword,
			IsActive:          originalUser.IsActive,
			EmailVerified:     originalUser.EmailVerified,
			CreatedAt:         originalUser.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         originalUser.UpdatedAt.Format(time.RFC3339),
		}

		// Generate new JWT with the original user info (no switching info)
		accessToken, accessExpiry, err := authService.GenerateJWT(originalDomainUser)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to generate original user token",
				"error", err,
				"originalUserUID", p.OriginalUserUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "could not generate token")
		}

		// Generate refresh token for the original user
		refreshToken, refreshExpiry, err := authService.GenerateRefreshJWT(originalDomainUser)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to generate refresh token for original user",
				"error", err,
				"originalUserUID", p.OriginalUserUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "could not generate refresh token")
		}

		tokenResp := &domain.TokenResponse{
			AccessToken:      accessToken,
			TokenType:        "bearer",
			ExpiresIn:        int(time.Until(accessExpiry).Seconds()),
			RefreshToken:     refreshToken,
			RefreshExpiresIn: int(time.Until(refreshExpiry).Seconds()),
		}

		// Set cookies
		authConfig := authService.GetAuthConfig()
		setAuthCookies(c, authConfig, tokenResp)

		log.WithContext(c.UserContext()).Info("Switch back successful",
			"originalUserEmail", originalUser.Email,
			"previousUserUID", p.UserUID)

		return c.JSON(tokenResp)
	}
}
