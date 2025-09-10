// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package roles

import (
	"aifo.dev/aifo/slideinsight/internal/config"
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// SetupRoleRoutes configures all role related routes
func SetupRoleRoutes(apiRoutes fiber.Router, roleService services.RoleService, userService services.UserService, config config.Config) {
	// Role routes group
	roleRoutes := apiRoutes.Group("/v1/roles")

	// All role routes require authentication
	roleRoutes.Use(middleware.Protected(config.Auth))

	// Basic role management routes
	roleRoutes.Post("/", handlers.CreateRole(roleService))
	roleRoutes.Get("/", handlers.GetRoles(roleService))
	roleRoutes.Get("/:name", handlers.GetRoleByName(roleService))
	roleRoutes.Delete("/:name", handlers.DeleteRole(roleService))

	// Role-permission management routes
	roleRoutes.Get("/:name/permissions", handlers.GetRolePermissions(roleService))
	roleRoutes.Post("/:name/permissions", handlers.AssignPermissionsToRole(roleService))
	roleRoutes.Delete("/:name/permissions/:permissionId", handlers.RemovePermissionFromRole(roleService))

	// Role-user management routes
	roleRoutes.Get("/:name/users", handlers.GetRoleUsers(roleService, userService))
	roleRoutes.Post("/:name/users", handlers.AssignUsersToRole(roleService, userService))
	roleRoutes.Delete("/:name/users/:userId", handlers.RemoveUserFromRole(roleService, userService))
}
