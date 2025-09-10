// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package annotations

import (
	"strconv"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/utils"
)

// VectorAnnotationFileParams represents path parameters for vector annotation file requests
type VectorAnnotationFileParams struct {
	SlideUID  string `params:"slideUID" validate:"required"`
	VectorUID string `params:"vectorUID" validate:"required"`
}

// SetupVectorAnnotationRoutes configures all vector annotation related routes
func SetupVectorAnnotationRoutes(protectedRoutes fiber.Router, vectorsService services.VectorAnnotationsService, studiesService services.StudiesService) {
	// Routes for all vector annotations
	// TODO: Make this route protected by a role and make a group
	protectedRoutes.Get("/admin/annotations/vector", handlers.GetVectorAnnotations(vectorsService))

	// Routes for slide-specific vector annotations
	protectedRoutes.Get("/slides/:slideUID/annotations/vector", handlers.GetVectorAnnotationsBySlideUID(vectorsService))
	protectedRoutes.Post("/slides/:slideUID/annotations/vector", handlers.AddVectorAnnotation(vectorsService))
	protectedRoutes.Put("/slides/:slideUID/annotations/vector/:vectorUID", handlers.UpdateVectorAnnotation(vectorsService))
	protectedRoutes.Delete("/slides/:slideUID/annotations/vector/:vectorUID", handlers.DeleteVectorAnnotation(vectorsService))

	// Route for vector annotation file with cache middleware
	protectedRoutes.Get("/slides/:slideUID/annotations/vector/:vectorUID/file",
		cache.New(cache.Config{
			ExpirationGenerator: func(c *fiber.Ctx, cfg *cache.Config) time.Duration {
				newCacheTime, _ := strconv.Atoi(c.GetRespHeader("Cache-Time", "600"))
				return time.Second * time.Duration(newCacheTime)
			},
			KeyGenerator: func(c *fiber.Ctx) string {
				var params VectorAnnotationFileParams
				if err := c.ParamsParser(&params); err != nil {
					// Fallback to path-based key if parsing fails
					return utils.CopyString(c.Path())
				}
				return utils.CopyString(c.Path() + ":" + params.SlideUID + ":" + params.VectorUID)
			},
			CacheHeader: "X-Cache",
		}),
		handlers.GetVectorAnnotationFile(vectorsService))

	// Route for importing vector annotation to workspace
	protectedRoutes.Post("/slides/:slideUID/annotations/vector/:vectorUID/import",
		handlers.ImportVectorAnnotation(vectorsService, studiesService))
}
