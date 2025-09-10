// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"errors"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	apperrors "aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"aifo.dev/aifo/slideinsight/internal/server/validation"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"

	"github.com/gofiber/fiber/v2"
)

// buildRasterAnnotationListResponse creates a consistent response structure for raster annotations
func buildRasterAnnotationListResponse(slideUID string, masks []domain.Mask) domain.MaskList {
	return domain.MaskList{
		SlideUID: slideUID,
		Masks:    masks,
	}
}

// GetRasterAnnotations returns a handler function that retrieves all raster annotations
// with optional pagination and search support
// @Summary Get all raster annotations
// @Description Retrieve all raster annotations in the system with pagination and search support
// @Tags raster-annotations
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 25)" minimum(1) maximum(100) example(25)
// @Param q query string false "General search across name, file URI, and slide UID" example("annotation")
// @Param name query string false "Filter by annotation name" example("Raster Annotation")
// @Param file_uri query string false "Filter by file URI" example("file:///raster/")
// @Param slide_uid query string false "Filter by slide UID" example("slide123")
// @Param actor_type query string false "Filter by actor type (user, model)" example("user")
// @Param sort query string false "Sort field (name, created_at, slide_uid, file_uri)" example("name")
// @Param dir query string false "Sort direction (asc, desc)" example("asc")
// @Success 200 {object} domain.RasterAnnotationsResponse "Paginated list of raster annotations"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/admin/annotations/raster [get]
func GetRasterAnnotations(service services.RasterAnnotationsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		params, err := utils.ParsePaginationAndSearchParamsWithConfig(c, utils.DefaultPaginationOptions(), utils.DefaultRasterAnnotationsSearchConfig())
		if err != nil {
			return err // Error already contains proper status code and message
		}

		// Use search method which handles both search and non-search cases
		masks, paginationInfo, err := service.GetRasterAnnotationsGeneric(c.UserContext(), params)
		if err != nil {
			log.Error("GetRasterAnnotations failed", "error", err, "search", params.SearchParams)
			return middleware.HandleError(c, err)
		}

		return c.JSON(domain.RasterAnnotationsResponse{
			Annotations: masks,
			Pagination:  paginationInfo,
		})
	}
}

// GetRasterAnnotationsBySlideUID returns a handler function that retrieves raster annotations for a specific slide
// @Summary Get raster annotations by slide UID
// @Description Retrieve raster annotations for a specific slide
// @Tags raster-annotations
// @Produce json
// @Security BearerAuth
// @Param slideUid path string true "Slide UID to filter raster annotations" example("slide123")
// @Param actor_type query string false "Filter by actor type (user, model)" example("user")
// @Success 200 {object} domain.MaskList "List of raster annotations for the slide"
// @Failure 400 {object} domain.ErrorResponse "Bad request - slideUid required"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUid}/annotations/raster [get]
func GetRasterAnnotationsBySlideUID(service services.RasterAnnotationsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input SlideUIDParams
		if err := c.ParamsParser(&input); err != nil {
			log.Warn("GetRasterAnnotationsBySlideUID request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// Parse search parameters to support actor_type filtering
		params, err := utils.ParsePaginationAndSearchParamsWithConfig(c, utils.DefaultPaginationOptions(), utils.DefaultRasterAnnotationsSearchConfig())
		if err != nil {
			return err // Error already contains proper status code and message
		}

		// Add slide_uid filter to the search parameters
		if params.SearchParams.Filters == nil {
			params.SearchParams.Filters = make(map[string]string)
		}
		params.SearchParams.Filters["slide_uid"] = input.SlideUID

		// Use the generic search method which handles filtering
		masks, _, err := service.GetRasterAnnotationsGeneric(c.UserContext(), params)
		if err != nil {
			log.Error("GetRasterAnnotationsBySlideUID failed", "error", err, "slideUid", input.SlideUID)
			return middleware.HandleError(c, err)
		}

		// Return masks for the specified slide
		return c.JSON(buildRasterAnnotationListResponse(input.SlideUID, masks))
	}
}

// AddRasterAnnotation returns a handler function that adds a mask and links it to a slide
// @Summary Add a new raster annotation
// @Description Add a new raster annotation and associate it with a slide
// @Tags raster-annotations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slideUid path string true "Slide UID" example("slide123")
// @Param mask body domain.Mask true "Raster annotation information"
// @Success 201 {object} domain.Mask "Created raster annotation"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Slide not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUid}/annotations/raster [post]
func AddRasterAnnotation(service services.RasterAnnotationsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input SlideUIDParams
		if err := c.ParamsParser(&input); err != nil {
			log.Warn("AddRasterAnnotation request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		mask := domain.Mask{}
		if err := c.BodyParser(&mask); err != nil {
			log.Warn("AddRasterAnnotation request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request body")
		}

		// Set slide UID from URL parameter
		mask.SlideUID = input.SlideUID

		// Validate the mask struct using modern validation
		if err := validation.GlobalValidator.ValidateStruct(c, mask); err != nil {
			return err // Error already formatted and sent
		}

		// Save the mask
		createdMask, err := service.SaveMask(c.UserContext(), mask)
		if err != nil {
			log.Error("AddRasterAnnotation failed", "error", err)
			// Check for specific error cases using proper error types
			if errors.Is(err, apperrors.ErrSlideNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "slide not found")
			}
			return middleware.HandleError(c, err)
		}

		// Return 201 Created with the created mask
		c.Status(fiber.StatusCreated)
		return c.JSON(createdMask)
	}
}

// GetMaskTile returns a handler for retrieving a specific mask tile in XYZ format
// @Summary Get mask tile
// @Description Retrieve a specific tile from a mask in XYZ tile format
// @Tags raster-annotations
// @Produce image/png
// @Security BearerAuth
// @Param slide_uid path string true "Slide ID" example("slide123")
// @Param mask_id path string true "Mask ID" example("mask456")
// @Param z path int true "Zoom level" example(0)
// @Param x path int true "Tile X coordinate" example(0)
// @Param y path int true "Tile Y coordinate" example(0)
// @Param format path string true "Image format (only png supported)" Enums(png) example("png")
// @Success 200 {file} binary "Mask tile image data"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 404 {object} domain.ErrorResponse "Mask, slide, or tile not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slide_uid}/annotations/raster/{mask_id}/tile/{z}/{x}/{y}.{format} [get]
func GetMaskTile(service services.RasterAnnotationsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input MaskTileParams
		if err := c.ParamsParser(&input); err != nil {
			log.Warn("GetMaskTile request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// Get the mask tile
		tile, err := service.GetMaskTile(c.UserContext(), input.SlideUID, input.MaskUID, input.Z, input.X, input.Y)
		if err != nil {
			log.Error("GetMaskTile failed", "error", err)
			// Check for specific error cases using proper error types
			if errors.Is(err, apperrors.ErrMaskNotFound) || errors.Is(err, apperrors.ErrSlideNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "mask or slide not found")
			}
			if apperrors.IsOutOfBounds(err) {
				return middleware.SendError(c, fiber.StatusNotFound, "tile not found")
			}
			return middleware.HandleError(c, err)
		}

		// Set appropriate headers
		c.Set("Content-Type", tile.ContentType)
		// TODO: This is not a good idea to do here, we should also use cache middleware.
		c.Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours

		// Send the image data
		return c.Send(tile.Image)
	}
}
