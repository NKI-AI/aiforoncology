// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package auth

import (
	"aifo.dev/aifo/slideinsight/internal/config"
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/server/ratelimit"
	authService "aifo.dev/aifo/slideinsight/internal/server/services/auth"

	"github.com/gofiber/fiber/v2"
)

// SetupAuthRoutes configures all authentication related routes
func SetupAuthRoutes(apiRoutes fiber.Router, authService authService.AuthService, db ports.Database, config config.Config) {
	// Auth routes group
	authRoutes := apiRoutes.Group("/v1/auth")

	// Public authentication routes with rate limiting
	authRoutes.Post("/login", ratelimit.RateLimited(authService), handlers.Login(authService))
	authRoutes.Post("/logout", handlers.Logout(authService))
	authRoutes.Post("/refresh", handlers.Refresh(authService))

	// Public password reset routes with rate limiting
	authRoutes.Post("/reset-password", ratelimit.RateLimited(authService), handlers.ResetPassword(authService))
	authRoutes.Post("/reset-password/confirm", handlers.ResetPasswordConfirm(authService))

	// Public forced password change route (for users who must reset password before getting JWT)
	authRoutes.Post("/forced-change-password", ratelimit.RateLimited(authService), handlers.ForcedChangePassword(authService))

	// Public user registration routes with rate limiting
	authRoutes.Post("/register", ratelimit.RateLimited(authService), handlers.RegisterUser(authService))
	authRoutes.Get("/verify-email", handlers.VerifyEmail(authService))
	authRoutes.Post("/resend-verification", ratelimit.RateLimited(authService), handlers.ResendVerification(authService))

	// Protected auth routes
	authRoutes.Get("/me", middleware.Protected(config.Auth), handlers.GetCurrentUser(authService, db))
	authRoutes.Post("/change-password", middleware.Protected(config.Auth), handlers.ChangePassword(authService))

	// User switching routes (admin only)
	authRoutes.Post("/switch-user", middleware.Protected(config.Auth), handlers.SwitchUser(authService, db))
	authRoutes.Post("/switch-back", middleware.Protected(config.Auth), handlers.SwitchBack(authService, db))
}
