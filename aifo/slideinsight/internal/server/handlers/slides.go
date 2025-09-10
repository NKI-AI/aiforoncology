// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"errors"
	"strconv"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	apperrors "aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"aifo.dev/aifo/slideinsight/internal/server/validation"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// slidesResponseBuilder builds a SlidesResponse from slides and pagination info
func slidesResponseBuilder(slides []domain.Slide, pagination domain.PaginationInfo) domain.SlidesResponse {
	return domain.SlidesResponse{
		Slides:     slides,
		Pagination: pagination,
	}
}

// GetSlides returns a handler function that retrieves all slides with optional search/filter support
// @Summary Get all slides
// @Description Retrieve a list of slides with optional search/filter and pagination support
// @Tags slides
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 100)" minimum(1) maximum(1000) example(100)
// @Param q query string false "General search across slide name and URI" example("sample slide")
// @Param slide_name query string false "Filter by slide name" example("slide-001")
// @Param sort query string false "Sort field (slide_name, created_at, slide_id)" example("slide_name")
// @Param dir query string false "Sort direction (asc, desc)" example("asc")
// @Success 200 {object} domain.SlidesResponse "List of slides with pagination"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides [get]
func GetSlides(service services.SlidesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		params, err := utils.ParsePaginationAndSearchParams(c)
		if err != nil {
			return err // The error already contains the proper status code and message
		}

		// Use search method which handles both search and non-search cases
		slides, paginationInfo, err := service.GetSlidesGeneric(c.UserContext(), params)
		if err != nil {
			log.Error("GetSlides failed", "error", err, "search", params.SearchParams)
			return middleware.HandleError(c, err)
		}

		return c.JSON(domain.SlidesResponse{
			Slides:     slides,
			Pagination: paginationInfo,
		})
	}
}

// GetSlidesCount returns a handler function that retrieves the total count of slides
// @Summary Get slides count
// @Description Retrieve the total count of slides in the system
// @Tags slides
// @Produce json
// @Security BearerAuth
// @Success 200 {object} fiber.Map "Total count of slides"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/count [get]
func GetSlidesCount(service services.SlidesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		count, err := service.GetSlidesCount(c.UserContext())
		if err != nil {
			log.Error("GetSlidesCount failed", "error", err)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{
			"count": count,
		})
	}
}

func GetSlidesByCaseUID(service services.SlidesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		caseUID := c.Params("caseUID")
		slides, err := service.GetSlidesByCaseUID(c.UserContext(), caseUID)
		if err != nil {
			log.Error("GetSlidesByCaseUID failed", "error", err)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{
			"slides": slides,
		})
	}
}

// GetSlideByUID returns a handler function that retrieves a specific slide by its ID
// @Summary Get slide by ID
// @Description Retrieve a specific slide by its unique identifier
// @Tags slides
// @Produce json
// @Security BearerAuth
// @Param slideUid path string true "Slide UID" example("slide123")
// @Success 200 {object} domain.Slide "Slide information"
// @Failure 400 {object} domain.ErrorResponse "Bad request - slideUid required"
// @Failure 404 {object} domain.ErrorResponse "Slide not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUid} [get]
func GetSlideByUID(service services.SlidesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input SlideUIDParams
		if err := c.ParamsParser(&input); err != nil {
			log.Warn("GetSlideByUID request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		slide, err := service.GetSlideByUID(c.UserContext(), input.SlideUID)
		if err != nil {
			log.Error("GetSlideByUID failed", "error", err, "slideUID", input.SlideUID)
			if errors.Is(err, apperrors.ErrSlideNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "slide not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(slide)
	}
}

// GetSlideMetadata returns a handler function that retrieves detailed metadata for a specific slide
// @Summary Get slide metadata
// @Description Retrieve detailed metadata and properties for a specific slide
// @Tags slides
// @Produce json
// @Security BearerAuth
// @Param slideUid path string true "Slide ID" example("slide123")
// @Success 200 {object} domain.SlideMetadata "Slide metadata"
// @Failure 400 {object} domain.ErrorResponse "Bad request - slideUid required"
// @Failure 404 {object} domain.ErrorResponse "Slide not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUid}/metadata [get]
func GetSlideMetadata(service services.SlidesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input SlideUIDParams
		if err := c.ParamsParser(&input); err != nil {
			log.Warn("GetSlideMetadata request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		metadata, err := service.GetSlideMetadata(c.UserContext(), input.SlideUID)
		if err != nil {
			log.Error("GetSlideMetadata failed", "error", err, "slideUid", input.SlideUID)
			if errors.Is(err, apperrors.ErrSlideNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "slide not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(metadata)
	}
}

// GetSlideTile handles requests for slide tiles
// @Summary Get slide tile
// @Description Retrieve a specific tile from a slide in XYZ tile format, or JXL (JPEG XL) format
// @Tags slides
// @Produce image/jpeg,image/png,image/jxl,application/octet-stream
// @Security BearerAuth
// @Param slideUid path string true "Slide ID" example("slide123")
// @Param z path int true "Zoom level" example(0)
// @Param x path int true "Tile X coordinate" example(0)
// @Param y path int true "Tile Y coordinate" example(0)
// @Param format path string true "Image format" Enums(jpg, png, jxl) example("jpg")
// @Param q query int false "Quality (1-100). Supported for JPEG, and JXL formats. PNG format will return error if set to non-100 value. Default is 75 for JXL, and JPEG." minimum(1) maximum(100) example(75)
// @Success 200 {file} binary "Tile image data, or JXL compressed data"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 404 {object} domain.ErrorResponse "Slide or tile not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUid}/tiles/{z}/{x}/{y}.{format} [get]
func GetSlideTile(service services.SlidesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input SlideTileParams
		if err := c.ParamsParser(&input); err != nil {
			log.Warn("GetSlideTile request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Manually parse quality query parameter since c.ParamsParser() only handles path params
		qualityStr := c.Query("q")
		if qualityStr != "" {
			if quality, err := strconv.Atoi(qualityStr); err == nil {
				if quality >= 1 && quality <= 100 {
					input.Quality = quality
				} else {
					log.Warn("Invalid quality parameter value", "quality", quality, "slideUID", input.SlideUID)
					return middleware.SendError(c, fiber.StatusBadRequest, "quality parameter must be between 1 and 100")
				}
			} else {
				log.Warn("Invalid quality parameter format", "qualityStr", qualityStr, "slideUID", input.SlideUID)
				return middleware.SendError(c, fiber.StatusBadRequest, "quality parameter must be a valid integer")
			}
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// Get the tile
		tile, err := service.GetSlideTile(c.UserContext(), input.SlideUID, input.Z, input.X, input.Y, input.Format, input.Quality)
		if err != nil {
			log.Error("GetSlideTile failed", "error", err, "slideUid", input.SlideUID, "z", input.Z, "x", input.X, "y", input.Y)
			if errors.Is(err, apperrors.ErrSlideNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "slide not found")
			}
			if errors.Is(err, apperrors.ErrInvalidFormat) {
				return middleware.SendError(c, fiber.StatusBadRequest, err.Error())
			}
			if apperrors.IsOutOfBounds(err) {
				return middleware.SendError(c, fiber.StatusNotFound, "tile not found")
			}
			return middleware.HandleError(c, err)
		}

		// Set the content type header explicitly
		if input.Format == "jpg" {
			c.Set("Content-Type", "image/jpeg")
		} else if input.Format == "png" {
			c.Set("Content-Type", "image/png")
		} else if input.Format == "jxl" {
			c.Set("Content-Type", "application/octet-stream")
		}

		// Set status code explicitly
		c.Status(fiber.StatusOK)

		return c.Send(tile.Image)
	}
}

// AddSlide returns a handler function that adds a slide
// @Summary Add a new slide
// @Description Add a new slide to the system
// @Tags slides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slide body SlideCreationInput true "Slide information"
// @Success 201 {object} domain.Slide "Created slide"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides [post]
func AddSlide(service services.SlidesService, imageTypesService services.ImageTypesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input SlideCreationInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("AddSlide request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate
		if input.SlideURI == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "slideUri is required")
		}

		// Validate image type ID if provided
		if input.ImageTypeId != "" {
			_, err := imageTypesService.GetImageTypeByID(c.UserContext(), input.ImageTypeId)
			if err != nil {
				log.Warn("Invalid image type ID provided", "imageTypeId", input.ImageTypeId, "error", err)
				return middleware.SendError(c, fiber.StatusBadRequest, "invalid image type ID")
			}
		}

		slide := domain.Slide{
			SlideUID:    input.SlideUID,
			SlideName:   input.SlideName,
			SlideURI:    input.SlideURI,
			ImageTypeId: input.ImageTypeId,
		}

		createdSlide, err := service.SaveSlide(c.UserContext(), slide)
		if err != nil {
			log.Error("AddSlide failed", "error", err)
			return middleware.HandleError(c, err)
		}

		// Return 201 Created with the slide data
		c.Status(fiber.StatusCreated)
		return c.JSON(createdSlide)
	}
}

// GetSlideAnnotationsOverview returns a handler function that provides an overview of annotations for a slide
// @Summary Get slide annotations overview
// @Description Get URLs and counts for both raster and vector annotations for a specific slide
// @Tags slides
// @Produce json
// @Security BearerAuth
// @Param slideUid path string true "Slide ID" example("slide123")
// @Success 200 {object} domain.SlideAnnotationsOverview "Slide annotations overview"
// @Failure 400 {object} domain.ErrorResponse "Bad request - slideUid required"
// @Failure 404 {object} domain.ErrorResponse "Slide not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUid}/annotations [get]
func GetSlideAnnotationsOverview(slidesService services.SlidesService, masksService services.RasterAnnotationsService, vectorsService services.VectorAnnotationsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input SlideUIDParams
		if err := c.ParamsParser(&input); err != nil {
			log.Warn("GetSlideAnnotationsOverview request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// First verify that the slide exists
		_, err := slidesService.GetSlideByUID(c.UserContext(), input.SlideUID)
		if err != nil {
			log.Error("GetSlideAnnotationsOverview - slide lookup failed", "error", err, "slideUid", input.SlideUID)
			if errors.Is(err, apperrors.ErrSlideNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "slide not found")
			}
			return middleware.HandleError(c, err)
		}

		// Get raster annotations for this slide
		rasterAnnotations, err := masksService.GetRasterAnnotationsForSlide(c.UserContext(), input.SlideUID)
		if err != nil {
			log.Error("GetSlideAnnotationsOverview - failed to get raster annotations", "error", err, "slideUid", input.SlideUID)
			return middleware.HandleError(c, err)
		}

		// Get vector annotations for this slide
		vectorAnnotations, err := vectorsService.GetVectorAnnotationsForSlide(c.UserContext(), input.SlideUID)
		if err != nil {
			log.Error("GetSlideAnnotationsOverview - failed to get vector annotations", "error", err, "slideUid", input.SlideUID)
			return middleware.HandleError(c, err)
		}

		// Build the response with URLs and counts
		overview := domain.SlideAnnotationsOverview{
			SlideUID:    input.SlideUID,
			RasterURL:   "/api/v1/slides/" + input.SlideUID + "/annotations/raster",
			VectorURL:   "/api/v1/slides/" + input.SlideUID + "/annotations/vector",
			RasterCount: len(rasterAnnotations),
			VectorCount: len(vectorAnnotations),
		}

		return c.JSON(overview)
	}
}

// UpdateSlide updates an existing slide using SlidesService
// @Summary Update a slide
// @Description Update an existing slide's information
// @Tags slides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slideUid path string true "Slide UID" example("slide123")
// @Param slide body UpdateSlideInput true "Slide update information"
// @Success 200 {object} domain.Slide "Updated slide"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Slide not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUid} [put]
func UpdateSlide(slidesService services.SlidesService, imageTypesService services.ImageTypesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input UpdateSlideInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("UpdateSlide request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		var params SlideUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("UpdateSlide request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		// Validate image type ID if provided in update
		if input.ImageTypeId != nil && *input.ImageTypeId != "" {
			_, err := imageTypesService.GetImageTypeByID(c.UserContext(), *input.ImageTypeId)
			if err != nil {
				log.Warn("Invalid image type ID provided in update", "imageTypeId", *input.ImageTypeId, "error", err)
				return middleware.SendError(c, fiber.StatusBadRequest, "invalid image type ID")
			}
		}

		// Get the existing slide
		existingSlide, err := slidesService.GetSlideByUID(c.UserContext(), params.SlideUID)
		if err != nil {
			log.Error("UpdateSlide - failed to get existing slide", "error", err, "slideUID", params.SlideUID)
			if errors.Is(err, apperrors.ErrSlideNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "slide not found")
			}
			return middleware.HandleError(c, err)
		}

		// Apply updates to the existing slide
		updatedSlide := existingSlide
		if input.SlideName != nil {
			updatedSlide.SlideName = *input.SlideName
		}
		if input.SlideURI != nil {
			updatedSlide.SlideURI = *input.SlideURI
		}
		if input.ImageTypeId != nil {
			updatedSlide.ImageTypeId = *input.ImageTypeId
		}

		// Save the updated slide
		savedSlide, err := slidesService.SaveSlide(c.UserContext(), updatedSlide)
		if err != nil {
			log.Error("UpdateSlide failed", "error", err, "slideUID", params.SlideUID)
			if apperrors.IsInvalidInput(err) {
				return middleware.SendError(c, fiber.StatusBadRequest, err.Error())
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(savedSlide)
	}
}

// SoftDeleteSlide soft deletes a slide using SlidesService
// @Summary Soft delete a slide
// @Description Mark a slide as deleted without removing it from the database
// @Tags slides
// @Produce json
// @Security BearerAuth
// @Param slideUid path string true "Slide UID" example("slide123")
// @Success 200 {object} fiber.Map "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request - slide ID required"
// @Failure 404 {object} domain.ErrorResponse "Slide not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUid} [delete]
func SoftDeleteSlide(slidesService services.SlidesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params SlideUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("SoftDeleteSlide request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		// TODO: Get actual user ID from context
		err := slidesService.SoftDeleteSlide(c.UserContext(), params.SlideUID, 1)
		if err != nil {
			log.Error("SoftDeleteSlide failed", "error", err, "slideUID", params.SlideUID)
			if errors.Is(err, apperrors.ErrSlideNotFound) || errors.Is(err, apperrors.ErrSlideAlreadyDeleted) {
				return middleware.SendError(c, fiber.StatusNotFound, err.Error())
			}
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Slide deleted successfully"})
	}
}

// CreateSlide creates a new slide using SlidesService
// @Summary Create a new slide
// @Description Create a new slide in the system
// @Tags slides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slide body SlideCreationInput true "Slide creation information"
// @Success 201 {object} domain.Slide "Created slide"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides [post]
func CreateSlide(slidesService services.SlidesService, imageTypesService services.ImageTypesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input SlideCreationInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("CreateSlide request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// Validate image type ID if provided
		if input.ImageTypeId != "" {
			_, err := imageTypesService.GetImageTypeByID(c.UserContext(), input.ImageTypeId)
			if err != nil {
				log.Warn("Invalid image type ID provided", "imageTypeId", input.ImageTypeId, "error", err)
				return middleware.SendError(c, fiber.StatusBadRequest, "invalid image type ID")
			}
		}

		slide := domain.Slide{
			SlideUID:    input.SlideUID,
			SlideName:   input.SlideName,
			SlideURI:    input.SlideURI,
			ImageTypeId: input.ImageTypeId,
		}

		createdSlide, err := slidesService.SaveSlide(c.UserContext(), slide)
		if err != nil {
			log.Error("CreateSlide failed", "error", err)
			if apperrors.IsInvalidInput(err) {
				return middleware.SendError(c, fiber.StatusBadRequest, err.Error())
			}
			return middleware.HandleError(c, err)
		}

		// Return 201 Created with the slide data
		c.Status(fiber.StatusCreated)
		return c.JSON(createdSlide)
	}
}
