// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package permissions

import (
	"aifo.dev/aifo/slideinsight/internal/config"
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// SetupPermissionRoutes configures all permission related routes
func SetupPermissionRoutes(apiRoutes fiber.Router, permissionService services.PermissionService, config config.Config) {
	// Permission routes group
	permissionRoutes := apiRoutes.Group("/v1/permissions")

	// All permission routes require authentication
	permissionRoutes.Use(middleware.Protected(config.Auth))

	// Permission management routes
	permissionRoutes.Post("/", handlers.CreatePermission(permissionService))
	permissionRoutes.Post("/bulk", handlers.CreatePermissionsBulk(permissionService))
	permissionRoutes.Get("/", handlers.GetPermissions(permissionService))
	permissionRoutes.Get("/:name", handlers.GetPermissionByName(permissionService))
	permissionRoutes.Delete("/:name", handlers.DeletePermission(permissionService))
}
