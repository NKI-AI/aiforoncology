// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"errors"
	"os"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	apperrors "aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"aifo.dev/aifo/slideinsight/internal/server/validation"
	"aifo.dev/aifo/slideinsight/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// buildVectorAnnotationListResponse creates a consistent response structure for vector annotations
func buildVectorAnnotationListResponse(slideUID string, vectors []domain.VectorAnnotation) domain.VectorAnnotationList {
	// Ensure annotations is never nil to avoid JSON null serialization
	if vectors == nil {
		vectors = []domain.VectorAnnotation{}
	}
	return domain.VectorAnnotationList{
		SlideUID:    slideUID,
		Annotations: vectors,
	}
}

// GetVectorAnnotations returns a handler function that retrieves all vector annotations
// with optional pagination and search support
// @Summary Get all vector annotations
// @Description Retrieve all vector annotations in the system with pagination and search support
// @Tags vector-annotations
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 25)" minimum(1) maximum(100) example(25)
// @Param q query string false "General search across name, file URI, and slide UID" example("annotation")
// @Param name query string false "Filter by annotation name" example("Vector Annotation")
// @Param file_uri query string false "Filter by file URI" example("file:///vectors/")
// @Param slide_uid query string false "Filter by slide UID" example("slide123")
// @Param actor_type query string false "Filter by actor type (user, model)" example("user")
// @Param sort query string false "Sort field (name, created_at, slide_uid, file_uri)" example("name")
// @Param dir query string false "Sort direction (asc, desc)" example("asc")
// @Success 200 {object} domain.VectorAnnotationsResponse "Paginated list of vector annotations"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/admin/annotations/vector [get]
func GetVectorAnnotations(service services.VectorAnnotationsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		params, err := utils.ParsePaginationAndSearchParamsWithConfig(c, utils.DefaultPaginationOptions(), utils.DefaultVectorAnnotationsSearchConfig())
		if err != nil {
			return err // Error already contains proper status code and message
		}

		// Use search method which handles both search and non-search cases
		vectors, paginationInfo, err := service.GetVectorAnnotationsGeneric(c.UserContext(), params)
		if err != nil {
			log.Error("GetVectorAnnotations failed", "error", err, "search", params.SearchParams)
			return middleware.HandleError(c, err)
		}

		return c.JSON(domain.VectorAnnotationsResponse{
			Annotations: vectors,
			Pagination:  paginationInfo,
		})
	}
}

// GetVectorAnnotationsBySlideUID returns a handler function that retrieves vector annotations for a specific slide
// @Summary Get vector annotations by slide UID
// @Description Retrieve vector annotations for a specific slide
// @Tags vector-annotations
// @Produce json
// @Security BearerAuth
// @Param slideUid path string true "Slide UID to filter vector annotations" example("slide123")
// @Param actor_type query string false "Filter by actor type (user, model)" example("user")
// @Success 200 {object} domain.VectorAnnotationList "List of vector annotations for the slide"
// @Failure 400 {object} domain.ErrorResponse "Bad request - slideUid required"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUid}/annotations/vector [get]
func GetVectorAnnotationsBySlideUID(service services.VectorAnnotationsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input SlideUIDParams
		if err := c.ParamsParser(&input); err != nil {
			log.Warn("GetVectorAnnotationsBySlideUID request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// Parse search parameters to support actor_type filtering
		params, err := utils.ParsePaginationAndSearchParamsWithConfig(c, utils.DefaultPaginationOptions(), utils.DefaultVectorAnnotationsSearchConfig())
		if err != nil {
			return err // Error already contains proper status code and message
		}

		// Add slide_uid filter to the search parameters
		if params.SearchParams.Filters == nil {
			params.SearchParams.Filters = make(map[string]string)
		}
		params.SearchParams.Filters["slide_uid"] = input.SlideUID

		// Use the generic search method which handles filtering
		vectors, _, err := service.GetVectorAnnotationsGeneric(c.UserContext(), params)
		if err != nil {
			log.Error("GetVectorAnnotationsBySlideUID failed", "error", err, "slideUid", input.SlideUID)
			return middleware.HandleError(c, err)
		}

		// Return vector annotations for the specified slide
		return c.JSON(buildVectorAnnotationListResponse(input.SlideUID, vectors))
	}
}

// AddVectorAnnotation returns a handler function that adds a vector annotation and links it to a slide
// @Summary Add a new vector annotation
// @Description Add a new vector annotation and associate it with a slide
// @Tags vector-annotations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slideUid path string false "Slide UID (can be provided in URL or body)" example("slide123")
// @Param vector body domain.VectorAnnotation true "Vector annotation information"
// @Success 201 {object} domain.VectorAnnotation "Created vector annotation"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Slide not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/annotations/vector [post]
// @Router /api/v1/slides/{slideUid}/annotations/vector [post]
func AddVectorAnnotation(service services.VectorAnnotationsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		vector := domain.VectorAnnotation{}
		if err := c.BodyParser(&vector); err != nil {
			log.Warn("AddVectorAnnotation request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request body")
		}

		// Validate the vector annotation struct
		if err := validation.GlobalValidator.ValidateStruct(c, vector); err != nil {
			return err // Error already formatted and sent
		}

		// Check if file exists only if FileURI is provided (not required if GeoJSON is inline)
		if vector.FileURI != "" {
			if _, err := os.Stat(vector.FileURI); os.IsNotExist(err) {
				log.Warn("AddVectorAnnotation request parsing failed", "error", "file does not exist")
				return middleware.SendError(c, fiber.StatusBadRequest, "file does not exist")
			}
		}

		// Save the vector annotation
		createdVector, err := service.SaveVectorAnnotation(c.UserContext(), vector)
		if err != nil {
			log.Error("AddVectorAnnotation failed", "error", err)
			// Check for specific error cases using proper error types
			if errors.Is(err, apperrors.ErrSlideNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "slide not found")
			}
			return middleware.HandleError(c, err)
		}

		// Return 201 Created with the created vector annotation
		c.Status(fiber.StatusCreated)
		return c.JSON(createdVector)
	}
}

// GetVectorAnnotationFile returns a handler for serving GeoJSON file content
// @Summary Get vector annotation GeoJSON file
// @Description Retrieve the GeoJSON file content for a vector annotation
// @Tags vector-annotations
// @Produce application/json
// @Security BearerAuth
// @Param slideUid path string true "Slide UID" example("slide123")
// @Param vectorUid path string true "Vector annotation UID" example("vector456")
// @Success 200 {object} object "GeoJSON content"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 404 {object} domain.ErrorResponse "Vector annotation, slide, or file not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUid}/annotations/vector/{vectorUid}/file [get]
func GetVectorAnnotationFile(service services.VectorAnnotationsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input VectorAnnotationFileParams
		if err := c.ParamsParser(&input); err != nil {
			log.Warn("GetVectorAnnotationFile request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// Get the GeoJSON file content
		geoJSONContent, err := service.GetVectorAnnotationFile(c.UserContext(), input.SlideUID, input.VectorUID)
		if err != nil {
			log.Error("GetVectorAnnotationFile failed", "error", err)
			// Check for specific error cases using proper error types
			if errors.Is(err, apperrors.ErrVectorAnnotationNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "vector annotation not found")
			}
			if apperrors.IsNotFound(err) {
				return middleware.SendError(c, fiber.StatusNotFound, "file not found")
			}
			return middleware.HandleError(c, err)
		}

		c.Set("Content-Type", "application/json")
		return c.Send(geoJSONContent)
	}
}

// UpdateVectorAnnotation returns a handler function that updates an existing vector annotation
// @Summary Update a vector annotation
// @Description Update an existing vector annotation
// @Tags vector-annotations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slideUid path string true "Slide UID" example("slide123")
// @Param vectorUid path string true "Vector annotation UID" example("vector456")
// @Param vector body domain.VectorAnnotation true "Updated vector annotation information"
// @Success 200 {object} domain.VectorAnnotation "Updated vector annotation"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Vector annotation not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUid}/annotations/vector/{vectorUid} [put]
func UpdateVectorAnnotation(service services.VectorAnnotationsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input VectorAnnotationFileParams
		if err := c.ParamsParser(&input); err != nil {
			log.Warn("UpdateVectorAnnotation request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		vector := domain.VectorAnnotation{}
		if err := c.BodyParser(&vector); err != nil {
			log.Warn("UpdateVectorAnnotation request body parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request body")
		}

		// Ensure the slide UID matches the URL parameter
		vector.SlideUID = input.SlideUID

		// Update the vector annotation
		updatedVector, err := service.UpdateVectorAnnotation(c.UserContext(), input.VectorUID, vector)
		if err != nil {
			log.Error("UpdateVectorAnnotation failed", "error", err)
			// Check for specific error cases using proper error types
			if errors.Is(err, apperrors.ErrVectorAnnotationNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "vector annotation not found")
			}
			if errors.Is(err, apperrors.ErrSlideNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "slide not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(updatedVector)
	}
}

// DeleteVectorAnnotation returns a handler function that deletes a vector annotation
// @Summary Delete a vector annotation
// @Description Soft-delete a vector annotation
// @Tags vector-annotations
// @Security BearerAuth
// @Param slideUid path string true "Slide UID" example("slide123")
// @Param vectorUid path string true "Vector annotation UID" example("vector456")
// @Success 204 "Vector annotation deleted successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 404 {object} domain.ErrorResponse "Vector annotation not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUid}/annotations/vector/{vectorUid} [delete]
func DeleteVectorAnnotation(service services.VectorAnnotationsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input VectorAnnotationFileParams
		if err := c.ParamsParser(&input); err != nil {
			log.Warn("DeleteVectorAnnotation request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// Delete the vector annotation
		err := service.DeleteVectorAnnotation(c.UserContext(), input.VectorUID)
		if err != nil {
			log.Error("DeleteVectorAnnotation failed", "error", err)
			// Check for specific error cases using proper error types
			if errors.Is(err, apperrors.ErrVectorAnnotationNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "vector annotation not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

// ImportVectorAnnotationRequest represents the request body for importing vector annotations
type ImportVectorAnnotationRequest struct {
	AllowedLabels []StudyAnnotationLabel `json:"allowedLabels" validate:"required,min=1"`
}

// StudyAnnotationLabel represents a label configuration from study settings
type StudyAnnotationLabel struct {
	Label string `json:"label" validate:"required"`
	Type  string `json:"type" validate:"required,oneof=point box polygon"`
	Color string `json:"color" validate:"required,hexcolor"`
}

// ImportVectorAnnotation returns a handler function that imports a vector annotation to workspace annotations
// @Summary Import vector annotation to workspace
// @Description Import a vector annotation as workspace annotations, matching labels with provided study settings and overwriting existing annotations
// @Tags vector-annotations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slideUid path string true "Slide UID" example("slide123")
// @Param vectorUid path string true "Vector annotation UID" example("vector456")
// @Param body body ImportVectorAnnotationRequest true "Study annotation settings with allowed labels"
// @Success 200 {object} domain.AnnotationImportResult "Import result with converted annotations"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 404 {object} domain.ErrorResponse "Vector annotation not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUid}/annotations/vector/{vectorUid}/import [post]
func ImportVectorAnnotation(vectorService services.VectorAnnotationsService, studiesService services.StudiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var pathParams VectorAnnotationFileParams
		if err := c.ParamsParser(&pathParams); err != nil {
			log.Warn("ImportVectorAnnotation path params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid path parameters")
		}

		// Validate the parsed path parameters
		if err := validation.GlobalValidator.ValidateStruct(c, pathParams); err != nil {
			return err // Error already formatted and sent
		}

		// Parse request body with study annotation settings
		var requestBody ImportVectorAnnotationRequest
		if err := c.BodyParser(&requestBody); err != nil {
			log.Warn("ImportVectorAnnotation body parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request body")
		}

		// Validate the request body
		if err := validation.GlobalValidator.ValidateStruct(c, requestBody); err != nil {
			return err // Error already formatted and sent
		}

		// Convert handler-level labels to service-level labels
		serviceLabels := make([]services.StudyAnnotationLabel, len(requestBody.AllowedLabels))
		for i, label := range requestBody.AllowedLabels {
			serviceLabels[i] = services.StudyAnnotationLabel{
				Label: label.Label,
				Type:  label.Type,
				Color: label.Color,
			}
		}

		// Import the vector annotation with provided study settings
		result, err := vectorService.ImportVectorAnnotationToWorkspace(c.UserContext(), pathParams.SlideUID, pathParams.VectorUID, serviceLabels)
		if err != nil {
			log.Error("ImportVectorAnnotation failed", "error", err, "slideUid", pathParams.SlideUID, "vectorUid", pathParams.VectorUID)
			// Check for specific error cases using proper error types
			if errors.Is(err, apperrors.ErrVectorAnnotationNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "vector annotation not found")
			}
			if errors.Is(err, apperrors.ErrSlideNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "slide not found")
			}
			if errors.Is(err, apperrors.ErrStudyNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "study not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(result)
	}
}
