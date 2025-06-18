// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package handlers

import (
	"log/slog"
	"strconv"
	"strings"

	"aifo.dev/aifo/slidescope/internal/server/domain"
	"aifo.dev/aifo/slidescope/internal/server/middleware"
	"aifo.dev/aifo/slidescope/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// SlideInput represents the slide creation request payload
type SlideInput struct {
	SlideID   string `json:"slideId" example:"slide123"`
	SlideName string `json:"slideName" example:"Sample Slide"`
	SlideURI  string `json:"slideUri" example:"file:///path/to/slide.svs" validate:"required"`
}

// GetSlides returns a handler function that retrieves all slides
// @Summary Get all slides
// @Description Retrieve a list of all available slides
// @Tags slides
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.SlidesResponse "List of slides"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides [get]
func GetSlides(service services.SlidesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		slides, err := service.GetSlides(c.UserContext())
		if err != nil {
			slog.Error("GetSlides failed", "error", err)
			return middleware.HandleError(c, err)
		}

		return c.JSON(domain.SlidesResponse{
			Slides: slides,
		})
	}
}

// GetSlideByID returns a handler function that retrieves a specific slide by its ID
// @Summary Get slide by ID
// @Description Retrieve a specific slide by its unique identifier
// @Tags slides
// @Produce json
// @Security BearerAuth
// @Param slide_id path string true "Slide ID" example("slide123")
// @Success 200 {object} domain.Slide "Slide information"
// @Failure 400 {object} domain.ErrorResponse "Bad request - slide_id required"
// @Failure 404 {object} domain.ErrorResponse "Slide not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slide_id} [get]
func GetSlideByID(service services.SlidesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		slideID := c.Params("slide_id")
		if slideID == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "slide_id is required")
		}

		slide, err := service.GetSlideByID(c.UserContext(), slideID)
		if err != nil {
			slog.Error("GetSlideByID failed", "error", err, "slide_id", slideID)
			if strings.Contains(err.Error(), "slide with ID") && strings.Contains(err.Error(), "not found") {
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
// @Param slide_id path string true "Slide ID" example("slide123")
// @Success 200 {object} domain.SlideMetadata "Slide metadata"
// @Failure 400 {object} domain.ErrorResponse "Bad request - slide_id required"
// @Failure 404 {object} domain.ErrorResponse "Slide not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slide_id}/metadata [get]
func GetSlideMetadata(service services.SlidesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		slideID := c.Params("slide_id")
		if slideID == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "slide_id is required")
		}

		metadata, err := service.GetSlideMetadata(c.UserContext(), slideID)
		if err != nil {
			slog.Error("GetSlideMetadata failed", "error", err, "slide_id", slideID)
			if strings.Contains(err.Error(), "slide with ID") && strings.Contains(err.Error(), "not found") {
				return middleware.SendError(c, fiber.StatusNotFound, "slide not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(metadata)
	}
}

// GetSlideTile handles requests for slide tiles
// @Summary Get slide tile
// @Description Retrieve a specific tile from a slide in XYZ tile format
// @Tags slides
// @Produce image/jpeg,image/png
// @Security BearerAuth
// @Param slide_id path string true "Slide ID" example("slide123")
// @Param z path int true "Zoom level" example(0)
// @Param x path int true "Tile X coordinate" example(0)
// @Param y path int true "Tile Y coordinate" example(0)
// @Param format path string true "Image format" Enums(jpeg, png) example("jpeg")
// @Success 200 {file} binary "Tile image data"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 404 {object} domain.ErrorResponse "Slide or tile not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slide_id}/tile/{z}/{x}/{y}.{format} [get]
func GetSlideTile(service services.SlidesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get parameters from URL
		slideID := c.Params("slide_id")
		if slideID == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "slide_id is required")
		}

		zStr := c.Params("z")
		xStr := c.Params("x")
		yStr := c.Params("y")
		format := c.Params("format")

		// Validate and convert parameters
		z, err := strconv.Atoi(zStr)
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid zoom level")
		}

		x, err := strconv.Atoi(xStr)
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid x coordinate")
		}

		y, err := strconv.Atoi(yStr)
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid y coordinate")
		}

		if format == "" {
			format = "jpeg" // Default format
		}

		// Get the tile
		tile, err := service.GetSlideTile(c.UserContext(), slideID, z, x, y, format)
		if err != nil {
			slog.Error("GetSlideTile failed", "error", err, "slide_id", slideID, "z", z, "x", x, "y", y)
			if strings.Contains(err.Error(), "slide with ID") && strings.Contains(err.Error(), "not found") {
				return middleware.SendError(c, fiber.StatusNotFound, "slide not found")
			}
			if strings.Contains(err.Error(), "unsupported image format") {
				return middleware.SendError(c, fiber.StatusBadRequest, err.Error())
			}
			if strings.Contains(err.Error(), "tile coordinates out of bounds") {
				return middleware.SendError(c, fiber.StatusNotFound, "tile not found")
			}
			return middleware.HandleError(c, err)
		}

		// Set response headers
		c.Set("Content-Type", tile.ContentType)
		c.Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours

		// Send the tile image
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
// @Param slide body SlideInput true "Slide information"
// @Success 201 {object} domain.Slide "Created slide"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides [post]
func AddSlide(service services.SlidesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input SlideInput
		if err := c.BodyParser(&input); err != nil {
			slog.Warn("AddSlide request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate
		if input.SlideURI == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "slideUri is required")
		}

		slide := domain.Slide{
			SlideID:   input.SlideID,
			SlideName: input.SlideName,
			SlideURI:  input.SlideURI,
		}

		createdSlide, err := service.SaveSlide(c.UserContext(), slide)
		if err != nil {
			slog.Error("AddSlide failed", "error", err)
			return middleware.HandleError(c, err)
		}

		// Return 201 Created with the slide data
		c.Status(fiber.StatusCreated)
		return c.JSON(createdSlide)
	}
}
