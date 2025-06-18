// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package middleware

import (
	"aifo.dev/aifo/slidescope/internal/server/domain"
	"aifo.dev/aifo/slidescope/internal/server/errors"
	"github.com/gofiber/fiber/v2"
)

// SendError is a helper function to send a standardized error response
func SendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(domain.ErrorResponse{
		Error: message,
	})
}

// HandleError maps domain errors to HTTP responses
func HandleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.IsNotFound(err):
		return SendError(c, fiber.StatusNotFound, err.Error())
	case errors.IsInvalidInput(err):
		return SendError(c, fiber.StatusBadRequest, err.Error())
	case errors.IsUnauthorized(err):
		return SendError(c, fiber.StatusUnauthorized, err.Error())
	case errors.IsOutOfBounds(err):
		return SendError(c, fiber.StatusNotFound, "resource not found")
	case errors.IsAlreadyExists(err):
		return SendError(c, fiber.StatusConflict, err.Error())
	default:
		return SendError(c, fiber.StatusInternalServerError, "internal error")
	}
}
