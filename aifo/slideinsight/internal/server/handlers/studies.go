// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	apperrors "aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"aifo.dev/aifo/slideinsight/internal/server/validation"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// studiesResponseBuilder builds a StudiesResponse from studies and pagination info
func studiesResponseBuilder(studies []domain.Study, pagination domain.PaginationInfo) domain.StudiesResponse {
	return domain.StudiesResponse{
		Studies:    studies,
		Pagination: pagination,
	}
}

// CreateStudy creates a new study
// @Summary Create a new study
// @Description Create a new study
// @Tags studies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param study body StudiesInput true "Study credentials"
// @Success 201 {object} domain.Study "Created study"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/studies [post]
func CreateStudy(service services.StudiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input StudiesInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("CreateTenant request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		study := domain.Study{
			StudyUID:    input.StudyUID,
			Name:        input.Name,
			Description: input.Description,
		}

		createdStudy, err := service.SaveStudy(c.UserContext(), study)
		if err != nil {
			log.Error("CreateStudy failed", "error", err, "name", study.Name)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(createdStudy)
	}
}

// GetStudyByUID retrieves a study by ID
// @Summary Get study by ID
// @Description Retrieve study information by ID
// @Tags studies
// @Produce json
// @Security BearerAuth
// @Param studyUid path string true "Study ID" example("Xyxz2234")
// @Success 200 {object} domain.Study "Study information"
// @Failure 400 {object} domain.ErrorResponse "Bad request - study ID required"
// @Failure 404 {object} domain.ErrorResponse "Study not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/studies/{studyUid} [get]
func GetStudyByUID(service services.StudiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params StudyUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetStudyByUID request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		study, err := service.GetStudyByUID(c.UserContext(), params.StudyUID)
		if err != nil {
			log.Error("GetStudyByUID failed", "error", err, "studyUID", params.StudyUID)
			// Check for specific error cases using proper error types
			if errors.Is(err, apperrors.ErrStudyNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "study not found")
			}
			return middleware.HandleError(c, err)
		}
		return c.JSON(study)
	}
}

// GetStudyMetadata retrieves study metadata by ID
// @Summary Get study metadata by ID
// @Description Retrieve study metadata including case count by ID
// @Tags studies
// @Produce json
// @Security BearerAuth
// @Param studyUid path string true "Study ID" example("Xyxz2234")
// @Success 200 {object} domain.StudyMetadata "Study metadata"
// @Failure 400 {object} domain.ErrorResponse "Bad request - study UID required"
// @Failure 404 {object} domain.ErrorResponse "Study not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/studies/{studyUid}/metadata [get]
func GetStudyMetadata(service services.StudiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params StudyUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetStudyMetadata request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		metadata, err := service.GetStudyMetadata(c.UserContext(), params.StudyUID)
		if err != nil {
			log.Error("GetStudyMetadata failed", "error", err, "studyUID", params.StudyUID)
			// Check for specific error cases using proper error types
			if errors.Is(err, apperrors.ErrStudyNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "study not found")
			}
			return middleware.HandleError(c, err)
		}
		return c.JSON(metadata)
	}
}

// GetStudyMetadataField retrieves study metadata field as JSON by ID
// @Summary Get study metadata field by ID
// @Description Retrieve study metadata field (JSON) by ID
// @Tags studies
// @Produce json
// @Security BearerAuth
// @Param studyUid path string true "Study ID" example("Xyxz2234")
// @Success 200 {object} fiber.Map "Study metadata field"
// @Failure 400 {object} domain.ErrorResponse "Bad request - study UID required"
// @Failure 404 {object} domain.ErrorResponse "Study not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/studies/{studyUid}/metadata-field [get]
func GetStudyMetadataField(service services.StudiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params StudyUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetStudyMetadataField request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		study, err := service.GetStudyByUID(c.UserContext(), params.StudyUID)
		if err != nil {
			log.Error("GetStudyMetadataField failed", "error", err, "studyUID", params.StudyUID)
			// Check for specific error cases using proper error types
			if errors.Is(err, apperrors.ErrStudyNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "study not found")
			}
			return middleware.HandleError(c, err)
		}

		// Return metadata as proper JSON object (already parsed as json.RawMessage)
		var metadata interface{} = map[string]interface{}{}
		if len(study.Metadata) > 0 {
			if err := json.Unmarshal(study.Metadata, &metadata); err != nil {
				// If parsing fails, return empty object
				log.Warn("Invalid JSON in study metadata", "studyUID", params.StudyUID, "metadata", string(study.Metadata))
				metadata = map[string]interface{}{}
			}
		}

		return c.JSON(fiber.Map{
			"metadata": metadata,
		})
	}
}

// GetStudies returns a handler function that retrieves all studies with optional search/filter support
// @Summary Get all studies
// @Description Retrieve a list of studies with optional search/filter and pagination support
// @Tags studies
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 25)" minimum(1) maximum(100) example(25)
// @Param q query string false "General search across name, description, and metadata" example("DCIS study")
// @Param name query string false "Filter by study name" example("study-001")
// @Param is_published query string false "Filter by published status" example("true")
// @Param sort query string false "Sort field (name, created_at, short_uid)" example("name")
// @Param dir query string false "Sort direction (asc, desc)" example("asc")
// @Param filterAccessibleStudies query bool false "Enable SQL-based permission filtering for efficiency" example(true)
// @Success 200 {object} domain.StudiesResponse "List of studies with pagination"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/studies [get]
func GetStudies(service services.StudiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		params, err := utils.ParsePaginationAndSearchParamsWithConfig(c, utils.DefaultPaginationOptions(), utils.DefaultStudiesSearchConfig())
		if err != nil {
			return err // Error already contains proper status code and message
		}

		// Use search method which handles both search and non-search cases
		studies, paginationInfo, err := service.GetStudiesGeneric(c.UserContext(), params)
		if err != nil {
			log.Error("GetStudies failed", "error", err, "search", params.SearchParams)
			return middleware.HandleError(c, err)
		}

		return c.JSON(domain.StudiesResponse{
			Studies:    studies,
			Pagination: paginationInfo,
		})
	}
}

// GetStudiesCount returns a handler function that retrieves the total count of studies
// @Summary Get studies count
// @Description Retrieve the total count of studies in the system
// @Tags studies
// @Produce json
// @Security BearerAuth
// @Success 200 {object} fiber.Map "Total count of studies"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/studies/count [get]
func GetStudiesCount(service services.StudiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		count, err := service.GetStudiesCount(c.UserContext(), utils.SearchParams{})
		if err != nil {
			log.Error("GetStudiesCount failed", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get studies count",
			})
		}

		return c.JSON(fiber.Map{
			"count": count,
		})
	}
}

// UpdateStudy updates study information
// @Summary Update study information
// @Description Update study information by study ID
// @Tags studies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param studyUid path string true "Study ID" example("Xyxz2234")
// @Param study body StudyUpdateInput true "Study update information"
// @Success 200 {object} domain.Study "Updated study information"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Study not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/studies/{studyUid} [put]
func UpdateStudy(service services.StudiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params StudyUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("UpdateStudy request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		var input StudyUpdateInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("UpdateStudy request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		updates := domain.StudyUpdates{
			Name:        input.Name,
			Description: input.Description,
			Metadata:    input.Metadata,
			IsPublished: input.IsPublished,
		}

		err := service.UpdateStudy(c.UserContext(), params.StudyUID, updates)
		if err != nil {
			log.Error("UpdateStudy failed", "error", err, "studyUID", params.StudyUID)
			// Check for specific error cases using proper error types
			if errors.Is(err, apperrors.ErrStudyNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "study not found")
			}
			return middleware.HandleError(c, err)
		}

		// Get the updated study to return
		updatedStudy, err := service.GetStudyByUID(c.UserContext(), params.StudyUID)
		if err != nil {
			log.Error("GetStudyByUID after update failed", "error", err, "studyUID", params.StudyUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.JSON(updatedStudy)
	}
}

// UpdateStudyMetadataField updates study metadata field
// @Summary Update study metadata field
// @Description Update study metadata field (JSON) by study ID
// @Tags studies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param studyUid path string true "Study ID" example("Xyxz2234")
// @Param metadata body fiber.Map true "Study metadata object"
// @Success 200 {object} fiber.Map "Updated metadata"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Study not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/studies/{studyUid}/metadata-field [post]
func UpdateStudyMetadataField(service services.StudiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params StudyUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("UpdateStudyMetadataField request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		var input map[string]interface{}
		if err := c.BodyParser(&input); err != nil {
			log.Warn("UpdateStudyMetadataField request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Convert the metadata to JSON bytes
		metadataBytes, err := json.Marshal(input)
		if err != nil {
			log.Warn("UpdateStudyMetadataField JSON marshaling failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid JSON")
		}
		metadataRaw := json.RawMessage(metadataBytes)

		updates := domain.StudyUpdates{
			Metadata: &metadataRaw,
		}

		err = service.UpdateStudy(c.UserContext(), params.StudyUID, updates)
		if err != nil {
			log.Error("UpdateStudyMetadataField failed", "error", err, "studyUID", params.StudyUID)
			// Check for specific error cases using proper error types
			if errors.Is(err, apperrors.ErrStudyNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "study not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{
			"message":  "Metadata updated successfully",
			"metadata": input,
		})
	}
}

// SoftDeleteStudy marks a study as deleted
// @Summary Soft delete a study
// @Description Mark a study as deleted without removing it from the database
// @Tags studies
// @Produce json
// @Security BearerAuth
// @Param studyUid path string true "Study ID" example("Xyxz2234")
// @Success 200 {object} fiber.Map "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request - study UID required"
// @Failure 404 {object} domain.ErrorResponse "Study not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/studies/{studyUid} [delete]
func SoftDeleteStudy(service services.StudiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params StudyUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("SoftDeleteStudy request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		err := service.SoftDeleteStudy(c.UserContext(), params.StudyUID)
		if err != nil {
			log.Error("SoftDeleteStudy failed", "error", err, "studyUID", params.StudyUID)
			// Check for specific error cases using proper error types
			if errors.Is(err, apperrors.ErrStudyNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "study not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{
			"message": "Study deleted successfully",
		})
	}
}

// RestoreStudy restores a soft-deleted study
// @Summary Restore a soft-deleted study
// @Description Restore a soft-deleted study
// @Tags studies
// @Produce json
// @Security BearerAuth
// @Param studyUid path string true "Study ID" example("Xyxz2234")
// @Success 200 {object} fiber.Map "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request - study UID required"
// @Failure 404 {object} domain.ErrorResponse "Study not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/studies/{studyUid}/restore [post]
func RestoreStudy(service services.StudiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params StudyUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("RestoreStudy request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		err := service.RestoreStudy(c.UserContext(), params.StudyUID)
		if err != nil {
			log.Error("RestoreStudy failed", "error", err, "studyUID", params.StudyUID)
			// Check for specific error cases using proper error types
			if errors.Is(err, apperrors.ErrStudyNotFound) {
				return middleware.SendError(c, fiber.StatusNotFound, "study not found")
			}
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{
			"message": "Study restored successfully",
		})
	}
}

// ExplainStudyAccess explains why the current user has access to a specific study
// @Summary Explain study access permissions
// @Description Explain in detail why the current user has or doesn't have access to a specific study
// @Tags studies
// @Produce json
// @Security BearerAuth
// @Param studyUid path string true "Study ID" example("Xyxz2234")
// @Param permission query string false "Permission to check (default: studies.view)" example("studies.view")
// @Success 200 {object} domain.PermissionExplanation "Permission explanation"
// @Failure 400 {object} domain.ErrorResponse "Bad request - study UID required"
// @Failure 401 {object} domain.ErrorResponse "Authentication required"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/studies/{studyUid}/access-explanation [get]
func ExplainStudyAccess(db ports.Database) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params StudyUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("ExplainStudyAccess request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		// Get the permission to check from query parameter (default to "studies.view")
		permission := c.Query("permission", "studies.view")

		// Get the principal from the JWT middleware
		principal := c.Locals("principal")
		if principal == nil {
			log.Warn("ExplainStudyAccess attempted without authentication")
			return middleware.SendError(c, fiber.StatusUnauthorized, "authentication required")
		}

		p, ok := principal.(*middleware.Principal)
		if !ok || p == nil {
			log.Warn("Invalid principal object in ExplainStudyAccess")
			return middleware.SendError(c, fiber.StatusUnauthorized, "invalid authentication")
		}

		// Get the user ID
		userID, err := db.GetUserIDByUID(c.UserContext(), p.UserUID)
		if err != nil {
			log.Error("Failed to get user ID for access explanation",
				"error", err,
				"userUID", p.UserUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "failed to get user information")
		}

		// Generate the permission explanation
		explanation, err := explainStudyPermission(c.UserContext(), db, userID, p.UserUID, permission, params.StudyUID)
		if err != nil {
			log.Error("Failed to generate permission explanation",
				"error", err,
				"userID", userID,
				"permission", permission,
				"studyUID", params.StudyUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "failed to generate permission explanation")
		}

		return c.JSON(explanation)
	}
}

// AddCaseToStudy adds a case to a study using StudiesService
// @Summary Add a case to a study
// @Description Add an existing case to a study
// @Tags studies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param studyUid path string true "Study ID" example("Xyxz2234")
// @Param request body AddCaseToStudyInput true "Case to add to study"
// @Success 200 {object} fiber.Map "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Study or case not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/studies/{studyUid}/cases [post]
func AddCaseToStudy(studiesService services.StudiesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params StudyUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("AddCaseToStudy path parameter parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		var input AddCaseToStudyInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("AddCaseToStudy request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		err := studiesService.AddCaseToStudy(c.UserContext(), params.StudyUID, input.CaseUID)
		if err != nil {
			log.Error("AddCaseToStudy failed", "error", err, "studyUID", params.StudyUID, "caseUID", input.CaseUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Case added to study"})
	}
}

// explainStudyPermission provides a detailed explanation of why a user has or doesn't have access to a study
func explainStudyPermission(ctx context.Context, db ports.Database, userID int, userUID, permission, studyUID string) (domain.PermissionExplanation, error) {
	explanation := domain.PermissionExplanation{
		UserUID:         userUID,
		Permission:      permission,
		ResourceType:    "study",
		ResourceUID:     studyUID,
		HasAccess:       false,
		ChecksPerformed: []domain.PermissionCheck{},
	}

	// Step 1: Check role-based permission first (takes precedence)
	roleCheck := domain.PermissionCheck{
		CheckType:   "role_based",
		Description: fmt.Sprintf("Checking if user has role-based permission '%s' (global precedence)", permission),
	}

	hasRolePermission, err := db.UserHasRolePermission(ctx, userID, permission)
	if err != nil {
		return explanation, fmt.Errorf("failed to check role permission: %w", err)
	}

	roleCheck.Result = hasRolePermission
	if hasRolePermission {
		roleCheck.GrantingEntity = &domain.PermissionGrantResource{
			ResourceType: "role",
			ResourceName: "User's assigned roles",
		}
		explanation.HasAccess = true
		explanation.GrantType = "role_based_grant"
		explanation.Message = fmt.Sprintf("User has access to study '%s' through role-based permission '%s'. Role-based permissions take precedence over object-specific permissions.", studyUID, permission)
	}
	explanation.ChecksPerformed = append(explanation.ChecksPerformed, roleCheck)

	// If role permission granted, we can stop here
	if hasRolePermission {
		return explanation, nil
	}

	// Step 2: Check direct study permission
	studyID, err := db.GetStudyIDByUID(ctx, studyUID)
	if err != nil {
		return explanation, fmt.Errorf("failed to get study ID: %w", err)
	}

	// Get study details for better explanation
	study, err := db.GetStudyByUID(ctx, studyUID)
	studyName := studyUID
	if err == nil {
		// Study found, use the actual name
		studyName = study.Name
	}
	// If error occurred, we'll continue with studyUID as the name

	directCheck := domain.PermissionCheck{
		CheckType:    "direct_object",
		ResourceType: "study",
		ResourceUID:  studyUID,
		ResourceName: studyName,
		Description:  fmt.Sprintf("Checking direct permission '%s' on study '%s'", permission, studyName),
	}

	hasDirectPermission, err := db.HasObjectGrant(ctx, userID, permission, "study", studyID)
	if err != nil {
		return explanation, fmt.Errorf("failed to check direct study permission: %w", err)
	}

	directCheck.Result = hasDirectPermission
	if hasDirectPermission {
		directCheck.GrantingEntity = &domain.PermissionGrantResource{
			ResourceType: "study",
			ResourceUID:  studyUID,
			ResourceName: studyName,
		}
		explanation.HasAccess = true
		explanation.GrantType = "direct_object_grant"
		explanation.GrantingResource = &domain.PermissionGrantResource{
			ResourceType: "study",
			ResourceUID:  studyUID,
			ResourceName: studyName,
		}
		explanation.Message = fmt.Sprintf("User has direct access to study '%s' with permission '%s'.", studyName, permission)
	}
	explanation.ChecksPerformed = append(explanation.ChecksPerformed, directCheck)

	// If no access granted, provide denial message
	if !explanation.HasAccess {
		explanation.GrantType = "access_denied"
		explanation.Message = fmt.Sprintf("User does not have access to study '%s'. No role-based permission '%s' found and no direct object grant on the study.", studyName, permission)
	}

	return explanation, nil
}
