// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"strconv"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"aifo.dev/aifo/slideinsight/internal/server/validation"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// RoleInput represents input for role creation/update
type RoleInput struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"required,min=1,max=1000"`
}

// RolePermissionAssignmentInput represents input for assigning permissions to roles
type RolePermissionAssignmentInput struct {
	PermissionIDs []int `json:"permission_ids" validate:"required,min=1"`
}

// UserRoleAssignmentInput represents input for assigning roles to users
type UserRoleAssignmentInput struct {
	UserUIDs []string `json:"user_uids" validate:"required,min=1"`
}

// CreateRole handles role creation
// @Summary Create a new role
// @Description Create a new role in the system
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param role body RoleInput true "Role creation request"
// @Success 201 {object} ports.Role "Role created successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/roles [post]
func CreateRole(roleService services.RoleService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "CreateRole")
		log.WithContext(c.UserContext()).Info("Processing role creation request", "requestId", reqInfo.ID)

		var input RoleInput
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("Role creation request parsing failed",
				"error", err, "requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the role input struct
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err
		}

		log.WithContext(c.UserContext()).Info("Creating role",
			"name", input.Name, "description", input.Description, "requestId", reqInfo.ID)

		role, err := roleService.CreateRole(c.UserContext(), input.Name, input.Description)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Role creation failed",
				"error", err, "name", input.Name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Role created successfully",
			"name", input.Name, "id", role.ID, "requestId", reqInfo.ID)

		return c.Status(fiber.StatusCreated).JSON(role)
	}
}

// GetRoles handles retrieving all roles
// @Summary Get all roles
// @Description Retrieve all roles from the system with pagination and search support
// @Tags roles
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 20)" minimum(1) maximum(200) example(20)
// @Param q query string false "General search across name and description" example("admin")
// @Param name query string false "Filter by role name" example("superadmin")
// @Param sort query string false "Sort field (name, created_at)" example("name")
// @Param dir query string false "Sort direction (asc, desc)" example("asc")
// @Success 200 {object} domain.RolesResponse "List of roles with pagination"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/roles [get]
func GetRoles(roleService services.RoleService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "GetRoles")
		log.WithContext(c.UserContext()).Info("Processing get roles request", "requestId", reqInfo.ID)

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
		roles, paginationInfo, err := roleService.GetRolesWithPagination(c.UserContext(), params)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve roles",
				"error", err, "search", params.SearchParams, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Roles retrieved successfully",
			"count", len(roles), "total", paginationInfo.Total, "requestId", reqInfo.ID)

		return c.JSON(domain.RolesResponse{
			Roles:      roles,
			Pagination: paginationInfo,
		})
	}
}

// GetRoleByName handles retrieving a role by its name
// @Summary Get role by name
// @Description Retrieve a role by its name
// @Tags roles
// @Produce json
// @Security BearerAuth
// @Param name path string true "Role name"
// @Success 200 {object} ports.Role "Role retrieved successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid name"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Role not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/roles/{name} [get]
func GetRoleByName(roleService services.RoleService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "GetRoleByName")
		name := c.Params("name")

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Get role by name request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "role name is required")
		}

		log.WithContext(c.UserContext()).Info("Processing get role by name request",
			"name", name, "requestId", reqInfo.ID)

		role, err := roleService.GetRoleByName(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve role",
				"error", err, "name", name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Role retrieved successfully",
			"name", name, "id", role.ID, "requestId", reqInfo.ID)

		return c.JSON(role)
	}
}

// DeleteRole handles role deletion
// @Summary Delete a role
// @Description Delete a role by its name
// @Tags roles
// @Security BearerAuth
// @Param name path string true "Role name"
// @Success 204 "Role deleted successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid name"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Role not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/roles/{name} [delete]
func DeleteRole(roleService services.RoleService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "DeleteRole")
		name := c.Params("name")

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Delete role request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "role name is required")
		}

		log.WithContext(c.UserContext()).Info("Processing delete role request",
			"name", name, "requestId", reqInfo.ID)

		err := roleService.DeleteRole(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to delete role",
				"error", err, "name", name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Role deleted successfully",
			"name", name, "requestId", reqInfo.ID)

		return c.SendStatus(fiber.StatusNoContent)
	}
}

// GetRolePermissions handles retrieving permissions for a role
// @Summary Get role permissions
// @Description Retrieve all permissions assigned to a role
// @Tags roles
// @Produce json
// @Security BearerAuth
// @Param name path string true "Role name"
// @Success 200 {array} ports.Permission "Role permissions retrieved successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid name"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Role not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/roles/{name}/permissions [get]
func GetRolePermissions(roleService services.RoleService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "GetRolePermissions")
		name := c.Params("name")

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Get role permissions request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "role name is required")
		}

		log.WithContext(c.UserContext()).Info("Processing get role permissions request",
			"name", name, "requestId", reqInfo.ID)

		// First get the role to get its ID
		role, err := roleService.GetRoleByName(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve role",
				"error", err, "name", name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		permissions, err := roleService.GetRolePermissions(c.UserContext(), role.ID)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve role permissions",
				"error", err, "name", name, "roleId", role.ID, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Role permissions retrieved successfully",
			"name", name, "roleId", role.ID, "count", len(permissions), "requestId", reqInfo.ID)

		return c.JSON(permissions)
	}
}

// AssignPermissionsToRole handles assigning permissions to a role
// @Summary Assign permissions to role
// @Description Assign multiple permissions to a role
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Role name"
// @Param permissions body RolePermissionAssignmentInput true "Permission assignment request"
// @Success 200 {object} map[string]interface{} "Permissions assigned successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Role not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/roles/{name}/permissions [post]
func AssignPermissionsToRole(roleService services.RoleService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "AssignPermissionsToRole")
		name := c.Params("name")

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Assign permissions to role request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "role name is required")
		}

		var input RolePermissionAssignmentInput
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("Role permission assignment request parsing failed",
				"error", err, "requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err
		}

		log.WithContext(c.UserContext()).Info("Processing assign permissions to role request",
			"name", name, "permissionCount", len(input.PermissionIDs), "requestId", reqInfo.ID)

		// First get the role to get its ID
		role, err := roleService.GetRoleByName(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve role",
				"error", err, "name", name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		var assigned []int
		var errors []string

		for _, permissionID := range input.PermissionIDs {
			err := roleService.AssignPermissionToRole(c.UserContext(), role.ID, permissionID)
			if err != nil {
				log.WithContext(c.UserContext()).Warn("Failed to assign permission to role",
					"error", err, "roleId", role.ID, "permissionId", permissionID, "requestId", reqInfo.ID)
				errors = append(errors, strconv.Itoa(permissionID)+": "+err.Error())
			} else {
				assigned = append(assigned, permissionID)
				log.WithContext(c.UserContext()).Info("Permission assigned to role",
					"roleId", role.ID, "permissionId", permissionID, "requestId", reqInfo.ID)
			}
		}

		result := map[string]interface{}{
			"assigned": assigned,
			"count":    len(assigned),
		}

		if len(errors) > 0 {
			result["errors"] = errors
			result["error_count"] = len(errors)
		}

		log.WithContext(c.UserContext()).Info("Permission assignment to role completed",
			"name", name, "assigned", len(assigned), "errors", len(errors), "requestId", reqInfo.ID)

		return c.JSON(result)
	}
}

// RemovePermissionFromRole handles removing a permission from a role
// @Summary Remove permission from role
// @Description Remove a permission from a role
// @Tags roles
// @Security BearerAuth
// @Param name path string true "Role name"
// @Param permissionId path int true "Permission ID"
// @Success 204 "Permission removed successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Role or permission not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/roles/{name}/permissions/{permissionId} [delete]
func RemovePermissionFromRole(roleService services.RoleService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "RemovePermissionFromRole")
		name := c.Params("name")
		permissionIDStr := c.Params("permissionId")

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Remove permission from role request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "role name is required")
		}

		permissionID, err := strconv.Atoi(permissionIDStr)
		if err != nil || permissionID <= 0 {
			log.WithContext(c.UserContext()).Warn("Remove permission from role request with invalid permission ID",
				"permissionId", permissionIDStr, "requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "valid permission ID is required")
		}

		log.WithContext(c.UserContext()).Info("Processing remove permission from role request",
			"name", name, "permissionId", permissionID, "requestId", reqInfo.ID)

		// First get the role to get its ID
		role, err := roleService.GetRoleByName(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve role",
				"error", err, "name", name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		err = roleService.RemovePermissionFromRole(c.UserContext(), role.ID, permissionID)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to remove permission from role",
				"error", err, "roleId", role.ID, "permissionId", permissionID, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Permission removed from role successfully",
			"name", name, "roleId", role.ID, "permissionId", permissionID, "requestId", reqInfo.ID)

		return c.SendStatus(fiber.StatusNoContent)
	}
}

// GetRoleUsers handles retrieving users assigned to a role
// @Summary Get role users
// @Description Retrieve all users assigned to a role
// @Tags roles
// @Produce json
// @Security BearerAuth
// @Param name path string true "Role name"
// @Success 200 {array} string "Role users retrieved successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid name"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Role not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/roles/{name}/users [get]
func GetRoleUsers(roleService services.RoleService, userService services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "GetRoleUsers")
		name := c.Params("name")

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Get role users request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "role name is required")
		}

		log.WithContext(c.UserContext()).Info("Processing get role users request",
			"name", name, "requestId", reqInfo.ID)

		// First get the role to get its ID
		role, err := roleService.GetRoleByName(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve role",
				"error", err, "name", name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		userRoles, err := roleService.GetRoleUsers(c.UserContext(), role.ID)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve role users",
				"error", err, "name", name, "roleId", role.ID, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		// Extract user IDs and convert to userUIDs
		userUIDs := make([]string, 0, len(userRoles))
		for _, userRole := range userRoles {
			user, err := userService.GetUserByInternalID(c.UserContext(), userRole.UserID)
			if err != nil {
				log.WithContext(c.UserContext()).Warn("Failed to retrieve user",
					"error", err, "userID", userRole.UserID, "requestId", reqInfo.ID)
				continue // Skip this user if we can't find them
			}
			userUIDs = append(userUIDs, user.ShortUID)
		}

		log.WithContext(c.UserContext()).Info("Role users retrieved successfully",
			"name", name, "roleId", role.ID, "count", len(userUIDs), "requestId", reqInfo.ID)

		return c.JSON(userUIDs)
	}
}

// AssignUsersToRole handles assigning users to a role
// @Summary Assign users to role
// @Description Assign multiple users to a role
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Role name"
// @Param users body UserRoleAssignmentInput true "User assignment request"
// @Success 200 {object} map[string]interface{} "Users assigned successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Role not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/roles/{name}/users [post]
func AssignUsersToRole(roleService services.RoleService, userService services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "AssignUsersToRole")
		name := c.Params("name")

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Assign users to role request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "role name is required")
		}

		var input UserRoleAssignmentInput
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("User role assignment request parsing failed",
				"error", err, "requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err
		}

		log.WithContext(c.UserContext()).Info("Processing assign users to role request",
			"name", name, "userCount", len(input.UserUIDs), "requestId", reqInfo.ID)

		// First get the role to get its ID
		role, err := roleService.GetRoleByName(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve role",
				"error", err, "name", name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		var assigned []string
		var errors []string

		for _, userUID := range input.UserUIDs {
			// Convert userUID to internal user ID
			user, err := userService.GetUserByUID(c.UserContext(), userUID)
			if err != nil {
				log.WithContext(c.UserContext()).Warn("Failed to find user by UID",
					"error", err, "userUID", userUID, "requestId", reqInfo.ID)
				errors = append(errors, userUID+": user not found")
				continue
			}

			err = roleService.AssignRoleToUser(c.UserContext(), user.ID, role.ID, nil) // nil for global tenant
			if err != nil {
				log.WithContext(c.UserContext()).Warn("Failed to assign role to user",
					"error", err, "roleId", role.ID, "userUID", userUID, "requestId", reqInfo.ID)
				errors = append(errors, userUID+": "+err.Error())
			} else {
				assigned = append(assigned, userUID)
				log.WithContext(c.UserContext()).Info("Role assigned to user",
					"roleId", role.ID, "userUID", userUID, "requestId", reqInfo.ID)
			}
		}

		result := map[string]interface{}{
			"assigned": assigned,
			"count":    len(assigned),
		}

		if len(errors) > 0 {
			result["errors"] = errors
			result["error_count"] = len(errors)
		}

		log.WithContext(c.UserContext()).Info("User assignment to role completed",
			"name", name, "assigned", len(assigned), "errors", len(errors), "requestId", reqInfo.ID)

		return c.JSON(result)
	}
}

// RemoveUserFromRole handles removing a user from a role
// @Summary Remove user from role
// @Description Remove a user from a role
// @Tags roles
// @Security BearerAuth
// @Param name path string true "Role name"
// @Param userId path string true "User UID"
// @Success 204 "User removed successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Role or user not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/roles/{name}/users/{userId} [delete]
func RemoveUserFromRole(roleService services.RoleService, userService services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "RemoveUserFromRole")
		name := c.Params("name")
		userUID := c.Params("userId") // This is actually userUID now

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Remove user from role request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "role name is required")
		}

		if userUID == "" {
			log.WithContext(c.UserContext()).Warn("Remove user from role request with empty user UID",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "user UID is required")
		}

		log.WithContext(c.UserContext()).Info("Processing remove user from role request",
			"name", name, "userUID", userUID, "requestId", reqInfo.ID)

		// First get the role to get its ID
		role, err := roleService.GetRoleByName(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve role",
				"error", err, "name", name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		// Convert userUID to internal user ID
		user, err := userService.GetUserByUID(c.UserContext(), userUID)
		if err != nil {
			log.WithContext(c.UserContext()).Warn("Failed to find user by UID",
				"error", err, "userUID", userUID, "requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "user not found")
		}

		err = roleService.RemoveRoleFromUser(c.UserContext(), user.ID, role.ID, nil) // nil for global tenant
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to remove role from user",
				"error", err, "roleId", role.ID, "userUID", userUID, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Role removed from user successfully",
			"name", name, "roleId", role.ID, "userUID", userUID, "requestId", reqInfo.ID)

		return c.SendStatus(fiber.StatusNoContent)
	}
}
