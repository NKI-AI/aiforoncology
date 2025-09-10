// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package settings

import (
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// SetupSettingsRoutes configures all settings related routes
// TODO: Don't add tenant_it to the routes, maybe add a tenant_id to the service?
// TODO: Need to think about it.
func SetupSettingsRoutes(protectedRoutes fiber.Router, settingsService *services.SettingsService) {
	protectedRoutes.Get("/settings", handlers.GetSettings(settingsService))
	protectedRoutes.Get("/settings/count", handlers.GetSettingsCount(settingsService))
	protectedRoutes.Get("/settings/:tenant_id/:key", handlers.GetSetting(settingsService))
	protectedRoutes.Post("/settings", handlers.CreateSetting(settingsService))
	protectedRoutes.Put("/settings/:tenant_id/:key", handlers.UpdateSetting(settingsService))
	protectedRoutes.Delete("/settings/:tenant_id/:key", handlers.DeleteSetting(settingsService))
}
