// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package handlers

import (
	"fmt"
	"log/slog"
	"time"

	"aifo.dev/aifo/slidescope/internal/server/domain"
	"aifo.dev/aifo/slidescope/internal/server/middleware"
	"aifo.dev/aifo/slidescope/internal/server/services"
	"github.com/gofiber/fiber/v2"
)

// LoginInput represents the login request payload
type LoginInput struct {
	Username string `json:"username" example:"testuser" validate:"required"`
	Password string `json:"password" example:"password123" validate:"required"`
}

// maskToken returns a masked version of the token for safe logging
func maskToken(token string) string {
	if len(token) < 10 {
		return "***"
	}
	return token[:5] + "..." + token[len(token)-5:]
}

// Login handles user authentication and sets a JWT token in a cookie matching Python implementation
// @Summary User login
// @Description Authenticate user credentials and return JWT tokens in cookies
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body LoginInput true "User credentials"
// @Success 200 {object} domain.TokenResponse "Authentication successful"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized - invalid credentials"
// @Router /api/v1/auth/login [post]
func Login(authService services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := fmt.Sprintf("%s-%d", c.IP(), time.Now().UnixNano())
		logger := slog.With("request_id", requestID, "ip", c.IP(), "endpoint", "login")

		logger.Info("Processing login request")

		var input LoginInput
		if err := c.BodyParser(&input); err != nil {
			logger.Warn("Login request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		if input.Username == "" {
			logger.Warn("Login attempt with empty username")
			return middleware.SendError(c, fiber.StatusBadRequest, "username is required")
		}

		if input.Password == "" {
			logger.Warn("Login attempt with empty password", "username", input.Username)
			return middleware.SendError(c, fiber.StatusBadRequest, "password is required")
		}

		logger.Info("Attempting authentication", "username", input.Username)

		// Authenticate the user and get token response
		tokenResp, err := authService.Login(c.UserContext(), input.Username, input.Password)
		if err != nil {
			logger.Error("Authentication failed",
				"error", err,
				"username", input.Username,
				"error_type", fmt.Sprintf("%T", err))
			return middleware.SendError(c, fiber.StatusUnauthorized, "incorrect username or password")
		}

		// Get auth config from the service
		authConfig := authService.GetAuthConfig()

		// Get SameSite value
		var sameSite string
		switch authConfig.Cookie.SameSite {
		case "strict":
			sameSite = "Strict"
		case "none":
			sameSite = "None"
		default:
			sameSite = "Lax"
		}

		// Set HTTP-only cookie with the JWT token using config settings
		// Set HTTP-only cookie with the access token
		accessCookie := fiber.Cookie{
			Name:     authConfig.Cookie.Name,
			Value:    tokenResp.AccessToken,
			Path:     authConfig.Cookie.Path,
			MaxAge:   tokenResp.ExpiresIn,
			HTTPOnly: authConfig.Cookie.HTTPOnly,
			Secure:   authConfig.Cookie.Secure,
			SameSite: sameSite,
		}
		c.Cookie(&accessCookie)

		// Set HTTP-only cookie with the refresh token
		refreshCookie := fiber.Cookie{
			Name:     authConfig.Cookie.Name + "_refresh",
			Value:    tokenResp.RefreshToken,
			Path:     authConfig.Cookie.Path,
			MaxAge:   tokenResp.RefreshExpiresIn,
			HTTPOnly: authConfig.Cookie.HTTPOnly,
			Secure:   authConfig.Cookie.Secure,
			SameSite: sameSite,
		}
		c.Cookie(&refreshCookie)

		accessExpiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		refreshExpiry := time.Now().Add(time.Duration(tokenResp.RefreshExpiresIn) * time.Second)

		logger.Info("Authentication successful",
			"username", input.Username,
			"access_token", maskToken(tokenResp.AccessToken),
			"refresh_token", maskToken(tokenResp.RefreshToken),
			"access_expires_at", accessExpiry.Format(time.RFC3339),
			"refresh_expires_at", refreshExpiry.Format(time.RFC3339))

		// Return the token response matching Python implementation
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
func Logout(authService services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := fmt.Sprintf("%s-%d", c.IP(), time.Now().UnixNano())
		logger := slog.With("request_id", requestID, "ip", c.IP(), "endpoint", "logout")

		username := c.Locals("username")
		logger.Info("Processing logout request", "username", username)

		// Get auth config from the service
		authConfig := authService.GetAuthConfig()

		// Get SameSite value
		var sameSite string
		switch authConfig.Cookie.SameSite {
		case "strict":
			sameSite = "Strict"
		case "none":
			sameSite = "None"
		default:
			sameSite = "Lax"
		}

		// Clear the auth cookie by setting its expiration to the past
		accessCookie := fiber.Cookie{
			Name:     authConfig.Cookie.Name,
			Value:    "",
			Path:     authConfig.Cookie.Path,
			Expires:  time.Now().Add(-time.Hour),
			HTTPOnly: authConfig.Cookie.HTTPOnly,
			Secure:   authConfig.Cookie.Secure,
			SameSite: sameSite,
		}
		c.Cookie(&accessCookie)

		refreshCookie := fiber.Cookie{
			Name:     authConfig.Cookie.Name + "_refresh",
			Value:    "",
			Path:     authConfig.Cookie.Path,
			Expires:  time.Now().Add(-time.Hour),
			HTTPOnly: authConfig.Cookie.HTTPOnly,
			Secure:   authConfig.Cookie.Secure,
			SameSite: sameSite,
		}
		c.Cookie(&refreshCookie)

		logger.Info("Logout successful", "username", username)

		return c.JSON(fiber.Map{
			"message": "Successfully logged out",
		})
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
func Refresh(authService services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := fmt.Sprintf("%s-%d", c.IP(), time.Now().UnixNano())
		logger := slog.With("request_id", requestID, "ip", c.IP(), "endpoint", "refresh")

		logger.Info("Processing token refresh request")

		authConfig := authService.GetAuthConfig()
		refreshToken := c.Cookies(authConfig.Cookie.Name + "_refresh")
		if refreshToken == "" {
			logger.Warn("Refresh attempt with missing token")
			return middleware.SendError(c, fiber.StatusUnauthorized, "missing refresh token")
		}

		logger.Info("Validating refresh token", "token", maskToken(refreshToken))

		username, err := authService.ValidateRefreshToken(refreshToken)
		if err != nil {
			logger.Error("Refresh token validation failed",
				"error", err,
				"token", maskToken(refreshToken),
				"error_type", fmt.Sprintf("%T", err))
			return middleware.SendError(c, fiber.StatusUnauthorized, "invalid refresh token")
		}

		logger.Info("Refresh token valid, generating new tokens", "username", username)

		// Generate new tokens
		accessToken, accessExp, err := authService.GenerateJWT(username)
		if err != nil {
			logger.Error("Failed to generate access token",
				"error", err,
				"username", username,
				"error_type", fmt.Sprintf("%T", err))
			return middleware.SendError(c, fiber.StatusInternalServerError, "could not refresh token")
		}
		refreshTokenNew, refreshExp, err := authService.GenerateRefreshJWT(username)
		if err != nil {
			logger.Error("Failed to generate refresh token",
				"error", err,
				"username", username,
				"error_type", fmt.Sprintf("%T", err))
			return middleware.SendError(c, fiber.StatusInternalServerError, "could not refresh token")
		}

		tokenResp := domain.TokenResponse{
			AccessToken:      accessToken,
			TokenType:        "bearer",
			ExpiresIn:        int(time.Until(accessExp).Seconds()),
			RefreshToken:     refreshTokenNew,
			RefreshExpiresIn: int(time.Until(refreshExp).Seconds()),
		}

		// Set cookies same as login handler
		var sameSite string
		switch authConfig.Cookie.SameSite {
		case "strict":
			sameSite = "Strict"
		case "none":
			sameSite = "None"
		default:
			sameSite = "Lax"
		}

		accessCookie := fiber.Cookie{
			Name:     authConfig.Cookie.Name,
			Value:    tokenResp.AccessToken,
			Path:     authConfig.Cookie.Path,
			MaxAge:   tokenResp.ExpiresIn,
			HTTPOnly: authConfig.Cookie.HTTPOnly,
			Secure:   authConfig.Cookie.Secure,
			SameSite: sameSite,
		}
		c.Cookie(&accessCookie)

		refreshCookie := fiber.Cookie{
			Name:     authConfig.Cookie.Name + "_refresh",
			Value:    tokenResp.RefreshToken,
			Path:     authConfig.Cookie.Path,
			MaxAge:   tokenResp.RefreshExpiresIn,
			HTTPOnly: authConfig.Cookie.HTTPOnly,
			Secure:   authConfig.Cookie.Secure,
			SameSite: sameSite,
		}
		c.Cookie(&refreshCookie)

		logger.Info("Token refresh successful",
			"username", username,
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
func GetCurrentUser(authService services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := fmt.Sprintf("%s-%d", c.IP(), time.Now().UnixNano())
		logger := slog.With("request_id", requestID, "ip", c.IP(), "endpoint", "get_current_user")

		// The middleware has validated the token and set the username in locals
		username := c.Locals("username")
		if username == nil {
			logger.Warn("Attempt to access user info without authentication")
			return middleware.SendError(c, fiber.StatusUnauthorized, "not authenticated")
		}

		logger.Info("Retrieved current user", "username", username)

		// Return the user info matching Python implementation
		return c.JSON(fiber.Map{
			"username": username,
			"scopes":   []string{}, // Empty scopes array
		})
	}
}
