// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package objectgrants

import (
	"aifo.dev/aifo/slideinsight/internal/config"
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// SetupObjectGrantRoutes configures all object grant related routes
func SetupObjectGrantRoutes(apiRoutes fiber.Router, objectGrantService services.ObjectGrantService, config config.Config) {
	// Object grant routes group
	objectGrantRoutes := apiRoutes.Group("/v1/object-grants")

	// All object grant routes require authentication
	objectGrantRoutes.Use(middleware.Protected(config.Auth))

	// Object grant management routes
	objectGrantRoutes.Post("/", handlers.CreateObjectGrant(objectGrantService))
	objectGrantRoutes.Get("/:resource_type/:resource_id", handlers.GetObjectGrants(objectGrantService))
	objectGrantRoutes.Delete("/:resource_type/:resource_id", handlers.DeleteObjectGrant(objectGrantService))
}
