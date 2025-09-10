// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"aifo.dev/aifo/slideinsight/internal/server/validation"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// CreatePermission handles permission creation
// @Summary Create a new permission
// @Description Create a new permission in the system
// @Tags permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param permission body PermissionInput true "Permission creation request"
// @Success 201 {object} ports.Permission "Permission created successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/permissions [post]
func CreatePermission(permissionService services.PermissionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "CreatePermission")
		log.WithContext(c.UserContext()).Info("Processing permission creation request", "requestId", reqInfo.ID)

		var input PermissionInput
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("Permission creation request parsing failed",
				"error", err, "requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the permission input struct using modern validation
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		log.WithContext(c.UserContext()).Info("Creating permission",
			"name", input.Name, "description", input.Description, "requestId", reqInfo.ID)

		permission, err := permissionService.CreatePermission(c.UserContext(), input.Name, input.Description)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Permission creation failed",
				"error", err, "name", input.Name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Permission created successfully",
			"name", input.Name, "id", permission.ID, "requestId", reqInfo.ID)

		return c.Status(fiber.StatusCreated).JSON(permission)
	}
}

// GetPermissions handles retrieving all permissions with pagination support
// @Summary Get all permissions
// @Description Retrieve permissions from the system with optional search/filter and pagination support
// @Tags permissions
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 20)" minimum(1) maximum(200) example(20)
// @Param q query string false "General search across name and description" example("view")
// @Param name query string false "Filter by permission name" example("studies.view")
// @Param category query string false "Filter by category" example("studies")
// @Param sort query string false "Sort field (name, created_at)" example("name")
// @Param dir query string false "Sort direction (asc, desc)" example("asc")
// @Success 200 {object} domain.PermissionsResponse "List of permissions with pagination"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/permissions [get]
func GetPermissions(permissionService services.PermissionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "GetPermissions")
		log.WithContext(c.UserContext()).Info("Processing get permissions request", "requestId", reqInfo.ID)

		// Parse pagination and search parameters
		params, err := utils.ParsePaginationAndSearchParamsWithConfig(c, utils.DefaultPaginationOptions(), utils.SearchConfig{
			SearchableFields: []string{"name", "description"},
			SortableFields:   []string{"name", "created_at", "updated_at"},
		})
		if err != nil {
			log.WithContext(c.UserContext()).Warn("Failed to parse pagination parameters",
				"error", err, "requestId", reqInfo.ID)
			return err // Error already contains proper status code and message
		}

		// Use search method which handles both search and non-search cases
		permissions, paginationInfo, err := permissionService.GetPermissionsWithPagination(c.UserContext(), params)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve permissions",
				"error", err, "search", params.SearchParams, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Permissions retrieved successfully",
			"count", len(permissions), "total", paginationInfo.Total, "requestId", reqInfo.ID)

		return c.JSON(domain.PermissionsResponse{
			Permissions: permissions,
			Pagination:  paginationInfo,
		})
	}
}

// GetPermissionByName handles retrieving a permission by its name
// @Summary Get permission by name
// @Description Retrieve a permission by its name
// @Tags permissions
// @Produce json
// @Security BearerAuth
// @Param name path string true "Permission name"
// @Success 200 {object} ports.Permission "Permission retrieved successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid name"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Permission not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/permissions/{name} [get]
func GetPermissionByName(permissionService services.PermissionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "GetPermissionByName")
		name := c.Params("name")

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Get permission by name request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "permission name is required")
		}

		log.WithContext(c.UserContext()).Info("Processing get permission by name request",
			"name", name, "requestId", reqInfo.ID)

		permission, err := permissionService.GetPermissionByName(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve permission",
				"error", err, "name", name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Permission retrieved successfully",
			"name", name, "id", permission.ID, "requestId", reqInfo.ID)

		return c.JSON(permission)
	}
}

// CreatePermissionsBulk handles bulk permission creation
// @Summary Create multiple permissions
// @Description Create multiple permissions in a single request
// @Tags permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param permissions body []PermissionInput true "List of permissions to create"
// @Success 201 {array} ports.Permission "Permissions created successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/permissions/bulk [post]
func CreatePermissionsBulk(permissionService services.PermissionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "CreatePermissionsBulk")
		log.WithContext(c.UserContext()).Info("Processing bulk permission creation request", "requestId", reqInfo.ID)

		var input []PermissionInput
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("Bulk permission creation request parsing failed",
				"error", err, "requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		if len(input) == 0 {
			log.WithContext(c.UserContext()).Warn("Bulk permission creation request with empty list",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "at least one permission must be provided")
		}

		// Validate each permission input struct
		for i, perm := range input {
			if perm.Name == "" {
				log.WithContext(c.UserContext()).Warn("Empty permission name in bulk request",
					"index", i, "requestId", reqInfo.ID)
				return middleware.SendError(c, fiber.StatusBadRequest, "permission name cannot be empty")
			}
			if perm.Description == "" {
				log.WithContext(c.UserContext()).Warn("Empty permission description in bulk request",
					"index", i, "name", perm.Name, "requestId", reqInfo.ID)
				return middleware.SendError(c, fiber.StatusBadRequest, "permission description cannot be empty")
			}
		}

		log.WithContext(c.UserContext()).Info("Creating bulk permissions",
			"count", len(input), "requestId", reqInfo.ID)

		var createdPermissions []interface{}
		var errors []string

		for _, perm := range input {
			permission, err := permissionService.CreatePermission(c.UserContext(), perm.Name, perm.Description)
			if err != nil {
				log.WithContext(c.UserContext()).Warn("Failed to create permission in bulk",
					"error", err, "name", perm.Name, "requestId", reqInfo.ID)
				errors = append(errors, perm.Name+": "+err.Error())
			} else {
				createdPermissions = append(createdPermissions, permission)
				log.WithContext(c.UserContext()).Info("Permission created in bulk",
					"name", perm.Name, "id", permission.ID, "requestId", reqInfo.ID)
			}
		}

		result := map[string]interface{}{
			"created": createdPermissions,
			"count":   len(createdPermissions),
		}

		if len(errors) > 0 {
			result["errors"] = errors
			result["error_count"] = len(errors)
			log.WithContext(c.UserContext()).Warn("Bulk permission creation completed with errors",
				"created", len(createdPermissions), "errors", len(errors), "requestId", reqInfo.ID)
		} else {
			log.WithContext(c.UserContext()).Info("Bulk permission creation completed successfully",
				"created", len(createdPermissions), "requestId", reqInfo.ID)
		}

		return c.Status(fiber.StatusCreated).JSON(result)
	}
}

// DeletePermission handles permission deletion
// @Summary Delete a permission
// @Description Delete a permission by its name
// @Tags permissions
// @Security BearerAuth
// @Param name path string true "Permission name"
// @Success 204 "Permission deleted successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid name"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Permission not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/permissions/{name} [delete]
func DeletePermission(permissionService services.PermissionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "DeletePermission")
		name := c.Params("name")

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Delete permission request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "permission name is required")
		}

		log.WithContext(c.UserContext()).Info("Processing delete permission request",
			"name", name, "requestId", reqInfo.ID)

		err := permissionService.DeletePermission(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to delete permission",
				"error", err, "name", name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Permission deleted successfully",
			"name", name, "requestId", reqInfo.ID)

		return c.SendStatus(fiber.StatusNoContent)
	}
}
