// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package handlers

import (
	"log/slog"
	"strings"

	"aifo.dev/aifo/slidescope/internal/server/domain"
	"aifo.dev/aifo/slidescope/internal/server/middleware"
	"aifo.dev/aifo/slidescope/internal/server/services"
	"github.com/gofiber/fiber/v2"
)

// UserInput represents the user creation request payload
type UserInput struct {
	Username string `json:"username" example:"testuser" validate:"required"`
	Password string `json:"password" example:"password123" validate:"required"`
}

// CreateUser creates a new user account
// @Summary Create a new user
// @Description Create a new user account with username and password
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body UserInput true "User credentials"
// @Success 201 {object} domain.User "Created user (password field will be empty)"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/users [post]
func CreateUser(service services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input UserInput
		if err := c.BodyParser(&input); err != nil {
			slog.Warn("CreateUser request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		if input.Username == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "username is required")
		}

		if input.Password == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "password is required")
		}

		user := domain.User{
			Username: input.Username,
			Password: input.Password,
		}

		createdUser, err := service.CreateUser(c.UserContext(), user)
		if err != nil {
			slog.Error("CreateUser failed", "error", err, "username", user.Username)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(createdUser)
	}
}

// GetUserByUsername retrieves a user by username
// @Summary Get user by username
// @Description Retrieve user information by username
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param username path string true "Username" example("testuser")
// @Success 200 {object} domain.User "User information"
// @Failure 400 {object} domain.ErrorResponse "Bad request - username required"
// @Failure 404 {object} domain.ErrorResponse "User not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/users/{username} [get]
func GetUserByUsername(service services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		username := c.Params("username")
		if username == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "username is required")
		}

		user, err := service.GetUserByUsername(c.UserContext(), username)
		if err != nil {
			slog.Error("GetUserByUsername failed", "error", err, "username", username)
			if strings.Contains(err.Error(), "user with username") && strings.Contains(err.Error(), "not found") {
				return middleware.SendError(c, fiber.StatusNotFound, "user not found")
			}
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		// TODO: Return a sanitized version of the user without password
		return c.JSON(user)
	}
}
