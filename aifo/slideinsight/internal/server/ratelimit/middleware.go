// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package ratelimit

import (
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	authService "aifo.dev/aifo/slideinsight/internal/server/services/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// RateLimited creates a rate limiting middleware for authentication endpoints
func RateLimited(authService authService.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract IP and username (if available in request body)
		ipAddress := c.IP()
		var username string

		// Try to extract username from request body for better rate limiting
		if c.Method() == "POST" {
			type loginRequest struct {
				Username string `json:"username"`
			}
			var req loginRequest
			if err := c.BodyParser(&req); err == nil {
				username = req.Username
			}
			// Reset body for next handler
			c.Request().SetBody(c.Body())
		}

		// Check rate limiting
		rateLimitResult, err := authService.CheckRateLimit(c.UserContext(), ipAddress, username)
		if err != nil {
			log.Error("Rate limit check failed", "error", err, "ip", ipAddress, "username", username)
			return middleware.SendError(c, fiber.StatusInternalServerError, "service temporarily unavailable")
		}

		if !rateLimitResult.Allowed {
			log.Warn("Request blocked by rate limiter",
				"ip", ipAddress,
				"username", username,
				"reason", rateLimitResult.Reason,
				"attempts_left", rateLimitResult.AttemptsLeft,
				"total_attempts", rateLimitResult.TotalAttempts)

			// Record the rate-limited attempt
			authService.RecordAuthAttempt(c.UserContext(), ipAddress, username, false, "rate_limited")

			return middleware.SendError(c, fiber.StatusTooManyRequests, rateLimitResult.Reason)
		}

		// Store rate limit info in context for handlers to use
		c.Locals("rate_limit_result", rateLimitResult)
		c.Locals("client_ip", ipAddress)
		if username != "" {
			c.Locals("attempted_username", username)
		}

		return c.Next()
	}
}
