// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package slides

import (
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// SetupSlideRoutes configures all slide related routes
func SetupSlideRoutes(
	protectedRoutes fiber.Router,
	slidesService services.SlidesService,
	casesService services.CasesService,
	masksService services.RasterAnnotationsService,
	vectorsService services.VectorAnnotationsService,
	imageTypesService services.ImageTypesService,
	histogramsService services.SlideHistogramsService,
	stainingProtocolsService services.StainingProtocolsService,
) {
	// Existing slide routes
	protectedRoutes.Get("/slides", handlers.GetSlides(slidesService))
	protectedRoutes.Get("/slides/count", handlers.GetSlidesCount(slidesService))
	protectedRoutes.Get("/slides/:slideUID", handlers.GetSlideByUID(slidesService))
	protectedRoutes.Get("/slides/:slideUID/metadata", handlers.GetSlideMetadata(slidesService))
	protectedRoutes.Get("/slides/:slideUID/annotations", handlers.GetSlideAnnotationsOverview(slidesService, masksService, vectorsService))
	protectedRoutes.Get("/cases/:caseUID/slides", handlers.GetSlidesByCaseUID(slidesService))
	protectedRoutes.Post("/slides", handlers.AddSlide(slidesService, imageTypesService))

	protectedRoutes.Get("/slides/:slideUID/tiles/:z/:x/:y.:format",
		handlers.GetSlideTile(slidesService))

	// Image types routes (system-level)
	protectedRoutes.Get("/image_types", handlers.GetImageTypes(imageTypesService))
	protectedRoutes.Get("/image_types/:id", handlers.GetImageType(imageTypesService))
	protectedRoutes.Post("/image_types", handlers.CreateImageType(imageTypesService))
	protectedRoutes.Put("/image_types/:id", handlers.UpdateImageType(imageTypesService))
	protectedRoutes.Delete("/image_types/:id", handlers.DeleteImageType(imageTypesService))

	// Slide histogram routes (slide-specific)
	protectedRoutes.Get("/slides/:slideUID/histogram", handlers.GetSlideHistogram(histogramsService))
	protectedRoutes.Post("/slides/:slideUID/histogram", handlers.CreateSlideHistogram(histogramsService))
	protectedRoutes.Delete("/slides/:slideUID/histogram", handlers.DeleteSlideHistogram(histogramsService))

	// Slide staining protocol routes (slide-specific)
	protectedRoutes.Get("/slides/:slideUID/staining_protocols", handlers.GetSlideStainingProtocols(stainingProtocolsService))
	protectedRoutes.Get("/slides/:slideUID/staining_protocols/:id", handlers.GetStainingProtocol(stainingProtocolsService))
	protectedRoutes.Post("/slides/:slideUID/staining_protocols", handlers.CreateStainingProtocol(stainingProtocolsService))
	protectedRoutes.Put("/slides/:slideUID/staining_protocols/:id", handlers.UpdateStainingProtocol(stainingProtocolsService))
	protectedRoutes.Delete("/slides/:slideUID/staining_protocols/:id", handlers.DeleteStainingProtocol(stainingProtocolsService))
}
