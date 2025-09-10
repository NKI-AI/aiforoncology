// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package annotations

import (
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
)

// SetupRasterAnnotationRoutes configures all raster annotation related routes
func SetupRasterAnnotationRoutes(protectedRoutes fiber.Router, rasterAnnotationsService services.RasterAnnotationsService) {
	// Routes for all raster annotations
	// TODO: Make this route protected by a role and make a group
	protectedRoutes.Get("/admin/annotations/raster", handlers.GetRasterAnnotations(rasterAnnotationsService))

	// Routes for slide-specific raster annotations
	protectedRoutes.Get("/slides/:slideUID/annotations/raster", handlers.GetRasterAnnotationsBySlideUID(rasterAnnotationsService))
	protectedRoutes.Post("/slides/:slideUID/annotations/raster", handlers.AddRasterAnnotation(rasterAnnotationsService))
	protectedRoutes.Get("/slides/:slideUID/annotations/raster/:maskUID/tiles/:z/:x/:y.:format",
		compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}),
		handlers.GetMaskTile(rasterAnnotationsService),
	)
}
