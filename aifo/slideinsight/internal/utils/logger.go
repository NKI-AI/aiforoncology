// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package utils

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// GetLoggerFromContext retrieves the structured zap logger from context
// Use this for hot paths where performance matters
func GetLoggerFromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value("logger").(*zap.Logger); ok && logger != nil {
		return logger
	}
	// Fallback to global logger
	return zap.L()
}

// GetSugarFromContext retrieves the sugar logger from context
// Use this for non-hot paths where convenience matters
func GetSugarFromContext(ctx context.Context) *zap.SugaredLogger {
	if sugar, ok := ctx.Value("sugar").(*zap.SugaredLogger); ok && sugar != nil {
		return sugar
	}
	// Fallback to global logger sugar
	return zap.L().Sugar()
}

// GetLoggerFromFiber retrieves the structured zap logger from fiber context
// Use this for hot paths where performance matters
func GetLoggerFromFiber(c *fiber.Ctx) *zap.Logger {
	if logger, ok := c.Locals("logger").(*zap.Logger); ok && logger != nil {
		return logger
	}
	// Try from user context
	return GetLoggerFromContext(c.UserContext())
}

// GetSugarFromFiber retrieves the sugar logger from fiber context
// Use this for non-hot paths where convenience matters
func GetSugarFromFiber(c *fiber.Ctx) *zap.SugaredLogger {
	if sugar, ok := c.Locals("sugar").(*zap.SugaredLogger); ok && sugar != nil {
		return sugar
	}
	// Try from user context
	return GetSugarFromContext(c.UserContext())
}

// LogInfo logs an info message with structured fields for hot paths
func LogInfo(ctx context.Context, msg string, fields ...zap.Field) {
	GetLoggerFromContext(ctx).Info(msg, fields...)
}

// LogWarn logs a warning message with structured fields for hot paths
func LogWarn(ctx context.Context, msg string, fields ...zap.Field) {
	GetLoggerFromContext(ctx).Warn(msg, fields...)
}

// LogError logs an error message with structured fields for hot paths
func LogError(ctx context.Context, msg string, fields ...zap.Field) {
	GetLoggerFromContext(ctx).Error(msg, fields...)
}

// LogDebug logs a debug message with structured fields for hot paths
func LogDebug(ctx context.Context, msg string, fields ...zap.Field) {
	GetLoggerFromContext(ctx).Debug(msg, fields...)
}

// InfoWithKeyValue logs an info message with key-value pairs for non-hot paths
func InfoWithKeyValue(ctx context.Context, msg string, keysAndValues ...interface{}) {
	GetSugarFromContext(ctx).Infow(msg, keysAndValues...)
}

// WarnWithKeyValue logs a warning message with key-value pairs for non-hot paths
func WarnWithKeyValue(ctx context.Context, msg string, keysAndValues ...interface{}) {
	GetSugarFromContext(ctx).Warnw(msg, keysAndValues...)
}

// ErrorWithKeyValue logs an error message with key-value pairs for non-hot paths
func ErrorWithKeyValue(ctx context.Context, msg string, keysAndValues ...interface{}) {
	GetSugarFromContext(ctx).Errorw(msg, keysAndValues...)
}

// DebugWithKeyValue logs a debug message with key-value pairs for non-hot paths
func DebugWithKeyValue(ctx context.Context, msg string, keysAndValues ...interface{}) {
	GetSugarFromContext(ctx).Debugw(msg, keysAndValues...)
}

// Convenience functions for when you don't have access to context
// These should be used sparingly and only when context is not available

// Info logs an info message without context (use sparingly)
func Info(msg string, fields ...zap.Field) {
	zap.L().Info(msg, fields...)
}

// Warn logs a warning message without context (use sparingly)
func Warn(msg string, fields ...zap.Field) {
	zap.L().Warn(msg, fields...)
}

// Error logs an error message without context (use sparingly)
func Error(msg string, fields ...zap.Field) {
	zap.L().Error(msg, fields...)
}

// Debug logs a debug message without context (use sparingly)
func Debug(msg string, fields ...zap.Field) {
	zap.L().Debug(msg, fields...)
}
