// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package users

import (
	"aifo.dev/aifo/slideinsight/internal/server/email"
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// SetupUserRoutes configures all user related routes
func SetupUserRoutes(protectedRoutes fiber.Router, userService services.UserService, roleService services.RoleService, asyncEmailService email.AsyncEmailService) {
	protectedRoutes.Get("/users", handlers.GetUsers(userService))
	protectedRoutes.Get("/users/count", handlers.GetUsersCount(userService))
	protectedRoutes.Get("/users/:userUID", handlers.GetUserByUID(userService))
	protectedRoutes.Get("/users/:userUID/roles", handlers.GetUserRoles(userService, roleService))
	protectedRoutes.Post("/users", handlers.CreateUser(userService))
	protectedRoutes.Put("/users/:userUID", handlers.UpdateUser(userService))
	protectedRoutes.Delete("/users/:userUID", handlers.DeleteUser(userService))
	protectedRoutes.Post("/users/:userUID/send-email", handlers.SendUserEmail(userService, asyncEmailService))
}
