// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package middleware

import (
	"fmt"
	"log/slog"
	"strings"

	"aifo.dev/aifo/slidescope/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// ExtractToken extracts JWT token from the request (from cookie or Authorization header)
func ExtractToken(c *fiber.Ctx, authConfig config.AuthConfig) string {
	// First try to get token from cookie
	token := c.Cookies(authConfig.Cookie.Name)

	// If no token found in cookie, check Authorization header
	if token == "" {
		authHeader := c.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 {
				token = parts[1]
			}
		}
	}

	return token
}

// Protected middleware verifies the JWT token from the cookie or Authorization header
func Protected(authConfig config.AuthConfig) fiber.Handler {
	jwtSecret := authConfig.GetJWTSecretBytes()

	return func(c *fiber.Ctx) error {
		// Extract token from cookie or header
		token := ExtractToken(c, authConfig)

		// No token found
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid or missing API token",
			})
		}

		// Parse and validate the token
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})
		if err != nil {
			slog.Warn("JWT validation failed", "error", err)
			var message string
			if err.Error() == "token has expired" {
				message = "Token has expired"
			} else {
				message = "Invalid token"
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": message,
			})
		}

		if !parsedToken.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid token",
			})
		}

		// Extract claims
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid token claims",
			})
		}

		// Extract username from subject claim
		username, ok := claims["sub"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid username in token",
			})
		}

		// Set username in locals for use in subsequent handlers
		c.Locals("username", username)

		// Continue to the next middleware/handler
		return c.Next()
	}
}
