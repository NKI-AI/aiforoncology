// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package regions

import (
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// SetupRegionRoutes configures all region related routes
func SetupRegionRoutes(
	protectedRoutes fiber.Router,
	regionsService services.RegionsService,
) {
	// Slide-specific region routes (most common pattern)
	protectedRoutes.Get("/slides/:slideUID/regions", handlers.GetRegionsBySlideUID(regionsService))
	protectedRoutes.Post("/slides/:slideUID/regions", handlers.CreateRegion(regionsService))
	protectedRoutes.Post("/slides/:slideUID/regions/bulk", handlers.BulkCreateRegions(regionsService))
	protectedRoutes.Delete("/slides/:slideUID/regions/:regionID", handlers.DeleteRegionBySlideUID(regionsService))

	// Individual region operations (using region ID)
	protectedRoutes.Get("/regions/:regionID", handlers.GetRegionByID(regionsService))
	protectedRoutes.Put("/regions/:regionID", handlers.UpdateRegion(regionsService))
	protectedRoutes.Delete("/regions/:regionID", handlers.DeleteRegion(regionsService))

	// Bulk operations on regions
	protectedRoutes.Put("/regions/bulk", handlers.BulkUpdateRegions(regionsService))
	protectedRoutes.Delete("/regions/bulk", handlers.BulkDeleteRegions(regionsService))

	// Region search and filtering
	protectedRoutes.Get("/regions", handlers.GetRegions(regionsService))

	// Region statistics
	protectedRoutes.Get("/slides/:slideUID/regions/statistics", handlers.GetRegionStatistics(regionsService))

	// Soft delete management
	protectedRoutes.Get("/regions/deleted", handlers.GetDeletedRegions(regionsService))
	protectedRoutes.Post("/regions/:regionID/restore", handlers.RestoreRegion(regionsService))
}
