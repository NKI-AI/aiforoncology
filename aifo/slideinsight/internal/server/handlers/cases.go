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
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// casesResponseBuilder builds a CasesResponse from cases and pagination info
func casesResponseBuilder(cases []domain.Case, pagination domain.PaginationInfo) domain.CasesResponse {
	return domain.CasesResponse{
		Cases:      cases,
		Pagination: pagination,
	}
}

// CreateCase creates a new case
// @Summary Create a new case
// @Description Create a new case
// @Tags cases
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param case body CasesInput true "Case credentials"
// @Success 201 {object} domain.Case "Created case"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/cases [post]
func CreateCase(service services.CasesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input CasesInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("CreateCase request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		case_ := domain.Case{
			Name: input.Name,
		}

		// The service layer will handle populating TenantID, CreatorID, ShortUID, etc.
		// from the user context and generate necessary IDs
		createdCase, err := service.SaveCase(c.UserContext(), case_)
		if err != nil {
			log.Error("CreateCase failed", "error", err, "name", input.Name)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(createdCase)
	}
}

// GetCases retrieves cases with annotation counts using ApplicationService
// @Summary Get all cases with annotation information
// @Description Retrieve all cases with slide counts and annotation counts included
// @Tags cases
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 100)" minimum(1) maximum(1000) example(100)
// @Param q query string false "General search across case name and metadata" example("sample case")
// @Param name query string false "Filter by case name" example("case-001")
// @Success 200 {object} domain.CasesResponse "List of cases with annotation information"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/cases [get]
func GetCases(casesService services.CasesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse pagination and search parameters
		params, err := utils.ParsePaginationAndSearchParams(c)
		if err != nil {
			return err // Error already contains proper status code and message
		}

		cases, paginationInfo, err := casesService.GetCases(c.UserContext(), params)
		if err != nil {
			log.Error("GetCases failed", "error", err)
			return middleware.HandleError(c, err)
		}

		return c.JSON(domain.CasesResponse{
			Cases:      cases,
			Pagination: paginationInfo,
		})
	}
}

// GetCaseByUID retrieves a specific case with annotation counts using ApplicationService
// @Summary Get case by ID with annotation information
// @Description Retrieve a specific case by its unique identifier with slide counts and annotation counts included
// @Tags cases
// @Produce json
// @Security BearerAuth
// @Param caseUid path string true "Case UID" example("Abcd1234")
// @Success 200 {object} domain.Case "Case information with annotation counts"
// @Failure 400 {object} domain.ErrorResponse "Bad request - case ID required"
// @Failure 404 {object} domain.ErrorResponse "Case not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{caseUid} [get]
func GetCaseByUID(casesService services.CasesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params CaseUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetCaseByUID request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		case_, err := casesService.GetCaseByUID(c.UserContext(), params.CaseUID)
		if err != nil {
			log.Error("GetCaseByUID failed", "error", err, "caseUID", params.CaseUID)
			if errors.Is(err, apperrors.ErrCaseNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "case not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(case_)
	}
}

// GetCasesByStudyUID retrieves cases by study ID with annotation counts using CasesService
// @Summary Get cases by study ID with annotation information
// @Description Retrieve all cases that belong to a specific study with slide counts and annotation counts included
// @Tags cases
// @Produce json
// @Security BearerAuth
// @Param studyUid path string true "Study UID" example("Xyxz2234")
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 100)" minimum(1) maximum(1000) example(100)
// @Param q query string false "General search across case name and metadata" example("sample case")
// @Param name query string false "Filter by case name" example("case-001")
// @Success 200 {object} domain.CasesResponse "List of cases for the study with annotation information"
// @Failure 400 {object} domain.ErrorResponse "Bad request - study ID required"
// @Failure 404 {object} domain.ErrorResponse "Study not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/studies/{studyUid}/cases [get]
func GetCasesByStudyUID(casesService services.CasesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params StudyUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetCasesByStudyUID request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		// Parse pagination and search parameters
		queryParams, err := utils.ParsePaginationAndSearchParams(c)
		if err != nil {
			return err // Error already contains proper status code and message
		}

		cases, paginationInfo, err := casesService.GetCasesByStudyUID(c.UserContext(), params.StudyUID, queryParams)
		if err != nil {
			log.Error("GetCasesByStudyUID failed", "error", err, "studyUID", params.StudyUID)
			if errors.Is(err, apperrors.ErrStudyNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "study not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(domain.CasesResponse{
			Cases:      cases,
			Pagination: paginationInfo,
		})
	}
}

// GetCasesCount returns a handler function that retrieves the total count of cases
// @Summary Get cases count
// @Description Retrieve the total count of cases in the system
// @Tags cases
// @Produce json
// @Security BearerAuth
// @Success 200 {object} fiber.Map "Total count of cases"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/cases/count [get]
func GetCasesCount(service services.CasesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		count, err := service.GetCasesCount(c.UserContext())
		if err != nil {
			log.Error("GetCasesCount failed", "error", err)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{
			"count": count,
		})
	}
}

// GetCaseNeighborsByStudyUID returns a handler function that retrieves case neighbors for navigation
// @Summary Get case neighbors for navigation
// @Description Retrieve the previous and next cases in a study for navigation purposes, honoring search parameters
// @Tags cases
// @Produce json
// @Security BearerAuth
// @Param studyUid path string true "Study UID" example("Xyxz2234")
// @Param caseUid path string true "Case UID" example("Abcd1234")
// @Param q query string false "General search across case name and metadata" example("sample case")
// @Param name query string false "Filter by case name" example("case-001")
// @Success 200 {object} domain.CaseNeighborsResponse "Case neighbors information"
// @Failure 400 {object} domain.ErrorResponse "Bad request - study ID or case ID required"
// @Failure 404 {object} domain.ErrorResponse "Case not found in study"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/studies/{studyUid}/cases/{caseUid}/neighbors [get]
func GetCaseNeighborsByStudyUID(service services.CasesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params StudyUIDCaseUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetCaseNeighborsByStudyUID request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		// Parse search parameters (but not pagination since it doesn't make sense for neighbors)
		searchParams := utils.ParseSearchParams(c, utils.DefaultCasesSearchConfig())

		neighbors, err := service.GetCaseNeighborsByStudyUID(c.UserContext(), params.StudyUID, params.CaseUID, searchParams)
		if err != nil {
			log.Error("GetCaseNeighborsByStudyUID failed", "error", err, "studyUID", params.StudyUID, "caseUID", params.CaseUID)
			if errors.Is(err, apperrors.ErrCaseNotFound) || errors.Is(err, apperrors.ErrStudyNotFound) || apperrors.IsNotFound(err) {
				return middleware.SendError(c, fiber.StatusNotFound, "case not found in study")
			}
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.JSON(neighbors)
	}
}

// AddSlideToCase adds a slide to a case using CasesService
// @Summary Add a new slide to a case
// @Description Add a new slide and associate it with a case
// @Tags slides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param caseUid path string true "Case UID" example("Abcd1234")
// @Param slide body SlideCreationInput true "Slide information"
// @Success 201 {object} domain.Slide "Created slide"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Case not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{caseUid}/slides [post]
func AddSlideToCase(casesService services.CasesService, imageTypesService services.ImageTypesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input SlideCreationInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("AddSlideToCase request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		var params CaseUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("AddSlideToCase request parsing failed", "error", err)
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
			CaseUID:     params.CaseUID,
		}

		createdSlide, err := casesService.AddSlideToCase(c.UserContext(), params.CaseUID, slide)
		if err != nil {
			log.Error("AddSlideToCase failed", "error", err)
			if errors.Is(err, apperrors.ErrCaseNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "case not found")
			}
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

// SoftDeleteCase soft deletes a case using CasesService
// @Summary Soft delete a case
// @Description Mark a case as deleted without removing it from the database
// @Tags cases
// @Produce json
// @Security BearerAuth
// @Param caseUid path string true "Case UID" example("Abcd1234")
// @Success 200 {object} fiber.Map "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request - case ID required"
// @Failure 404 {object} domain.ErrorResponse "Case not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{caseUid} [delete]
func SoftDeleteCase(casesService services.CasesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params CaseUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("SoftDeleteCase request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		err := casesService.SoftDeleteCase(c.UserContext(), params.CaseUID, 1) // TODO: Get actual user ID from context
		if err != nil {
			log.Error("SoftDeleteCase failed", "error", err, "caseUID", params.CaseUID)
			if errors.Is(err, apperrors.ErrCaseNotFound) || errors.Is(err, apperrors.ErrCaseAlreadyDeleted) {
				return middleware.SendError(c, fiber.StatusNotFound, err.Error())
			}
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Case deleted successfully"})
	}
}

// GetDeletedCases retrieves all soft-deleted cases using CasesService
// @Summary Get deleted cases
// @Description Retrieve all soft-deleted cases
// @Tags cases
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.CasesResponse "List of deleted cases"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/cases/deleted [get]
func GetDeletedCases(casesService services.CasesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		cases, err := casesService.GetDeletedCases(c.UserContext())
		if err != nil {
			log.Error("GetDeletedCases failed", "error", err)
			return middleware.HandleError(c, err)
		}
		return c.JSON(domain.CasesResponse{
			Cases: cases,
		})
	}
}

// RestoreCase restores a soft-deleted case using CasesService
// @Summary Restore a deleted case
// @Description Restore a soft-deleted case
// @Tags cases
// @Produce json
// @Security BearerAuth
// @Param caseUid path string true "Case UID" example("Abcd1234")
// @Success 200 {object} fiber.Map "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request - case ID required"
// @Failure 404 {object} domain.ErrorResponse "Case not found or not deleted"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{caseUid}/restore [post]
func RestoreCase(casesService services.CasesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params CaseUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("RestoreCase request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		err := casesService.RestoreCase(c.UserContext(), params.CaseUID)
		if err != nil {
			log.Error("RestoreCase failed", "error", err, "caseUID", params.CaseUID)
			if errors.Is(err, apperrors.ErrCaseNotFound) || errors.Is(err, apperrors.ErrCaseNotDeleted) {
				return middleware.SendError(c, fiber.StatusNotFound, err.Error())
			}
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Case restored successfully"})
	}
}

// UpdateCase updates an existing case using CasesService
// @Summary Update a case
// @Description Update an existing case's information
// @Tags cases
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param caseUid path string true "Case UID" example("Abcd1234")
// @Param case body UpdateCaseInput true "Case update information"
// @Success 200 {object} domain.Case "Updated case"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Case not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/cases/{caseUid} [put]
func UpdateCase(casesService services.CasesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input UpdateCaseInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("UpdateCase request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		var params CaseUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("UpdateCase request parsing failed", "error", err)
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

		// Get the existing case
		existingCase, err := casesService.GetCaseByUID(c.UserContext(), params.CaseUID)
		if err != nil {
			log.Error("UpdateCase - failed to get existing case", "error", err, "caseUID", params.CaseUID)
			if errors.Is(err, apperrors.ErrCaseNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "case not found")
			}
			return middleware.HandleError(c, err)
		}

		// Apply updates to the existing case
		updatedCase := existingCase
		if input.Name != nil {
			updatedCase.Name = *input.Name
		}
		if input.Metadata != nil {
			updatedCase.Metadata = *input.Metadata
		}

		// Save the updated case
		savedCase, err := casesService.SaveCase(c.UserContext(), updatedCase)
		if err != nil {
			log.Error("UpdateCase failed", "error", err, "caseUID", params.CaseUID)
			if apperrors.IsInvalidInput(err) {
				return middleware.SendError(c, fiber.StatusBadRequest, err.Error())
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(savedCase)
	}
}
