// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package handlers

import (
	"log/slog"
	"strings"

	"aifo.dev/aifo/slidescope/internal/server/domain"
	"aifo.dev/aifo/slidescope/internal/server/middleware"
	"aifo.dev/aifo/slidescope/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// GetMasks returns a handler function that retrieves masks
// If slide_id is in the URL path, only returns masks for that slide
// @Summary Get masks
// @Description Retrieve masks, optionally filtered by slide ID
// @Tags masks
// @Produce json
// @Security BearerAuth
// @Param slide_id path string false "Slide ID to filter masks" example("slide123")
// @Success 200 {object} domain.MaskList "List of masks"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/raster [get]
// @Router /api/v1/slides/{slide_id}/raster [get]
func GetMasks(service services.MasksService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get all masks from the service
		masks, err := service.GetMasks(c.UserContext())
		if err != nil {
			slog.Error("GetMasks failed", "error", err)
			return middleware.HandleError(c, err)
		}

		// If a slide_id is provided in the path, filter masks for that slide only
		slideID := c.Params("slide_id")
		if slideID != "" {
			filteredMasks := []domain.Mask{}
			for _, mask := range masks {
				if mask.SlideID == slideID {
					filteredMasks = append(filteredMasks, mask)
				}
			}

			// Return only masks for the specified slide
			return c.JSON(domain.MaskList{
				SlideID: slideID,
				Masks:   filteredMasks,
			})
		}

		// Return all masks if no slide_id was specified
		return c.JSON(domain.MaskList{
			SlideID: "",
			Masks:   masks,
		})
	}
}

// AddMask returns a handler function that adds a mask and links it to a slide
// @Summary Add a new mask
// @Description Add a new mask and associate it with a slide
// @Tags masks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slide_id path string false "Slide ID (can be provided in URL or body)" example("slide123")
// @Param mask body domain.Mask true "Mask information"
// @Success 201 {object} domain.Mask "Created mask"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Slide not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/raster [post]
// @Router /api/v1/slides/{slide_id}/raster [post]
func AddMask(service services.MasksService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		mask := domain.Mask{}
		if err := c.BodyParser(&mask); err != nil {
			slog.Warn("AddMask request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request body")
		}

		// Extract slide ID from URL parameter if not provided in body
		if mask.SlideID == "" {
			mask.SlideID = c.Params("slide_id")
		}

		// Validation
		if mask.MaskURI == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "maskUri is required")
		}

		if mask.SlideID == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "slideId is required")
		}

		// Force generation of a new ID
		mask.MaskID = ""

		// Save the mask
		createdMask, err := service.SaveMask(c.UserContext(), mask)
		if err != nil {
			slog.Error("AddMask failed", "error", err)
			// Check for specific error cases
			if strings.Contains(err.Error(), "slide with ID") && strings.Contains(err.Error(), "does not exist") {
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
// @Tags masks
// @Produce image/png
// @Security BearerAuth
// @Param slide_id path string true "Slide ID" example("slide123")
// @Param mask_id path string true "Mask ID" example("mask456")
// @Param z path int true "Zoom level" example(0)
// @Param x path int true "Tile X coordinate" example(0)
// @Param y path int true "Tile Y coordinate" example(0)
// @Param format path string true "Image format (only png supported)" Enums(png) example("png")
// @Success 200 {file} binary "Mask tile image data"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 404 {object} domain.ErrorResponse "Mask, slide, or tile not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slide_id}/raster/{mask_id}/tile/{z}/{x}/{y}.{format} [get]
func GetMaskTile(service services.MasksService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		slideID := c.Params("slide_id")
		maskID := c.Params("mask_id")
		z, err := c.ParamsInt("z")
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid z parameter")
		}
		x, err := c.ParamsInt("x")
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid x parameter")
		}
		y, err := c.ParamsInt("y")
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid y parameter")
		}
		format := c.Params("format")
		if format != "png" {
			return middleware.SendError(c, fiber.StatusBadRequest, "only png format is supported for masks")
		}

		// Get the mask tile
		tile, err := service.GetMaskTile(c.UserContext(), slideID, maskID, z, x, y)
		if err != nil {
			slog.Error("GetMaskTile failed", "error", err)
			// Check for specific error cases
			if strings.Contains(err.Error(), "no masks found for slide") {
				return middleware.SendError(c, fiber.StatusNotFound, "mask or slide not found")
			}
			if strings.Contains(err.Error(), "tile coordinates out of bounds") {
				return middleware.SendError(c, fiber.StatusNotFound, "tile not found")
			}
			return middleware.HandleError(c, err)
		}

		// Set appropriate headers
		c.Set("Content-Type", tile.ContentType)
		c.Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours

		// Send the image data
		return c.Send(tile.Image)
	}
}
