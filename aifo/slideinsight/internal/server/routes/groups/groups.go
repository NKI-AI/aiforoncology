// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package groups

import (
	"aifo.dev/aifo/slideinsight/internal/config"
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// SetupGroupRoutes configures all group related routes
func SetupGroupRoutes(apiRoutes fiber.Router, groupService services.GroupService, userService services.UserService, config config.Config) {
	// Group routes group
	groupRoutes := apiRoutes.Group("/v1/groups")

	// All group routes require authentication
	groupRoutes.Use(middleware.Protected(config.Auth))

	// Basic group management routes
	groupRoutes.Post("/", handlers.CreateGroup(groupService))
	groupRoutes.Get("/", handlers.GetGroups(groupService))
	groupRoutes.Get("/:name", handlers.GetGroupByName(groupService))
	groupRoutes.Delete("/:name", handlers.DeleteGroup(groupService))

	// Group-user management routes
	groupRoutes.Get("/:name/users", handlers.GetGroupUsers(groupService))
	groupRoutes.Post("/:name/users", handlers.AssignUsersToGroup(groupService, userService))
	groupRoutes.Delete("/:name/users/:userId", handlers.RemoveUserFromGroup(groupService, userService))
}
