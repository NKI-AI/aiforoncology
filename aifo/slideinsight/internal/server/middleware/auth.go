// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package middleware

import (
	"fmt"
	"strconv"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
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
			log.Warn("JWT validation failed", "error", err)
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

		email, ok := claims["sub"].(string)
		if !ok {
			log.Warn("Invalid email in token", "token", token)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid email in token",
			})
		}

		userUID, ok := claims["user"].(string)
		if !ok {
			log.Warn("Invalid user id in token", "token", token)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid user id in token",
			})
		}

		tenantUID, ok := claims["tenant"].(string)
		if !ok {
			log.Warn("Invalid tenant in token", "token", token)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid tenant in token",
			})
		}

		// Parse tenant ID from the tenant claim (it's stored as string representation of ID)
		tenantID, err := strconv.Atoi(tenantUID)
		if err != nil {
			log.Warn("Invalid tenant ID format in token", "tenant", tenantUID, "error", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid tenant ID in token",
			})
		}

		p := &Principal{
			Email:     email,
			UserUID:   userUID,
			TenantUID: tenantUID, // Keep for backward compatibility
			TenantID:  tenantID,  // Add parsed tenant ID
		}

		// Check if this is a switched user session
		if switchedSession, ok := claims["switched_session"].(bool); ok && switchedSession {
			// Extract original user information
			if originalUserUID, ok := claims["original_user"].(string); ok {
				p.OriginalUserUID = originalUserUID
			}
			if originalEmail, ok := claims["original_email"].(string); ok {
				p.OriginalEmail = originalEmail
			}
		}

		// attach to both Fiber Locals and context.Context
		c.Locals("principal", p)
		c.SetUserContext(WithPrincipal(c.UserContext(), p))

		// Continue to the next middleware/handler
		return c.Next()
	}
}
