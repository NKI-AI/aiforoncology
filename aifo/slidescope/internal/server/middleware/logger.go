// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package middleware

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

// RequestIDKey is the context key used to store and retrieve the request ID
type contextKey string

const RequestIDContextKey = contextKey("requestid")

// WithRequestID returns a middleware that adds the request ID to the slog logger context
func WithRequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the request ID from the context - this assumes requestid middleware is used first
		requestID := c.Locals("requestid")
		if requestID == nil {
			// Fallback to the header in case Locals is not set
			requestID = c.Get(fiber.HeaderXRequestID)
		}

		if requestID != "" {
			// Create a new logger with the request ID as an attribute
			requestLogger := slog.With("request_id", requestID)

			// Create a new context with the enhanced logger
			ctx := context.WithValue(c.UserContext(), RequestIDContextKey, requestLogger)

			// Set the context back to Fiber
			c.SetUserContext(ctx)
		}

		return c.Next()
	}
}

// GetLogger retrieves the logger with request ID from the context
// This can be used anywhere you have access to the context
func GetLogger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(RequestIDContextKey).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return slog.Default()
}
