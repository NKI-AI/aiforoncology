// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package system

import (
	"aifo.dev/aifo/slideinsight/internal/config"
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// SetupEmailTemplateRoutes configures all email template related routes
func SetupEmailTemplateRoutes(apiRoutes fiber.Router, templateService services.EmailTemplateService, config config.Config) {
	// Email template routes group
	templateRoutes := apiRoutes.Group("/v1/admin/system/email-templates")

	// All email template routes require authentication
	templateRoutes.Use(middleware.Protected(config.Auth))

	// Utility routes - these must come BEFORE parameterized routes
	templateRoutes.Get("/variables", handlers.GetTemplateVariables(templateService))
	templateRoutes.Post("/defaults", handlers.CreateDefaultTemplates(templateService))
	templateRoutes.Post("/preview", handlers.PreviewTemplate(templateService))

	// Email template management routes - parameterized routes come last
	templateRoutes.Get("/", handlers.GetEmailTemplates(templateService))
	templateRoutes.Post("/", handlers.CreateEmailTemplate(templateService))
	templateRoutes.Get("/:id", handlers.GetEmailTemplateByID(templateService))
	templateRoutes.Put("/:id", handlers.UpdateEmailTemplate(templateService))
	templateRoutes.Delete("/:id", handlers.DeleteEmailTemplate(templateService))
}
