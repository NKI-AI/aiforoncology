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

// GroupInput represents input for group creation/update
type GroupInput struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"required,min=1,max=1000"`
}

// UserGroupAssignmentInput represents input for assigning users to groups
type UserGroupAssignmentInput struct {
	UserUIDs []string `json:"user_uids" validate:"required,min=1"`
}

// CreateGroup handles group creation
// @Summary Create a new group
// @Description Create a new group in the system
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param group body GroupInput true "Group creation request"
// @Success 201 {object} ports.Group "Group created successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/groups [post]
func CreateGroup(groupService services.GroupService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "CreateGroup")
		log.WithContext(c.UserContext()).Info("Processing group creation request", "requestId", reqInfo.ID)

		var input GroupInput
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("Group creation request parsing failed",
				"error", err, "requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the group input struct
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err
		}

		log.WithContext(c.UserContext()).Info("Creating group",
			"name", input.Name, "description", input.Description, "requestId", reqInfo.ID)

		group, err := groupService.CreateGroup(c.UserContext(), input.Name, input.Description)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Group creation failed",
				"error", err, "name", input.Name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Group created successfully",
			"name", input.Name, "id", group.ID, "requestId", reqInfo.ID)

		return c.Status(fiber.StatusCreated).JSON(group)
	}
}

// GetGroups handles retrieving all groups
// @Summary Get all groups
// @Description Retrieve all groups from the system with pagination and search support
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 20)" minimum(1) maximum(200) example(20)
// @Param q query string false "General search across name and description" example("admin")
// @Param name query string false "Filter by group name" example("superadmins")
// @Param sort query string false "Sort field (name, created_at)" example("name")
// @Param dir query string false "Sort direction (asc, desc)" example("asc")
// @Success 200 {object} domain.GroupsResponse "List of groups with pagination"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/groups [get]
func GetGroups(groupService services.GroupService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "GetGroups")
		log.WithContext(c.UserContext()).Info("Processing get groups request", "requestId", reqInfo.ID)

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
		groups, paginationInfo, err := groupService.GetGroupsWithPagination(c.UserContext(), params)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve groups",
				"error", err, "search", params.SearchParams, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Groups retrieved successfully",
			"count", len(groups), "total", paginationInfo.Total, "requestId", reqInfo.ID)

		return c.JSON(domain.GroupsResponse{
			Groups:     groups,
			Pagination: paginationInfo,
		})
	}
}

// GetGroupByName handles retrieving a group by its name
// @Summary Get group by name
// @Description Retrieve a group by its name
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param name path string true "Group name"
// @Success 200 {object} ports.Group "Group retrieved successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid name"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Group not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/groups/{name} [get]
func GetGroupByName(groupService services.GroupService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "GetGroupByName")
		name := c.Params("name")

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Get group by name request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "group name is required")
		}

		log.WithContext(c.UserContext()).Info("Processing get group by name request",
			"name", name, "requestId", reqInfo.ID)

		group, err := groupService.GetGroupByName(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve group",
				"error", err, "name", name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Group retrieved successfully",
			"name", name, "id", group.ID, "requestId", reqInfo.ID)

		return c.JSON(group)
	}
}

// DeleteGroup handles group deletion
// @Summary Delete a group
// @Description Delete a group by its name
// @Tags groups
// @Security BearerAuth
// @Param name path string true "Group name"
// @Success 204 "Group deleted successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid name"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Group not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/groups/{name} [delete]
func DeleteGroup(groupService services.GroupService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "DeleteGroup")
		name := c.Params("name")

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Delete group request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "group name is required")
		}

		log.WithContext(c.UserContext()).Info("Processing delete group request",
			"name", name, "requestId", reqInfo.ID)

		err := groupService.DeleteGroup(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to delete group",
				"error", err, "name", name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Group deleted successfully",
			"name", name, "requestId", reqInfo.ID)

		return c.SendStatus(fiber.StatusNoContent)
	}
}

// GetGroupUsers handles retrieving users in a group
// @Summary Get group users
// @Description Retrieve all users in a group
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param name path string true "Group name"
// @Success 200 {array} int "Group users retrieved successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid name"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Group not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/groups/{name}/users [get]
func GetGroupUsers(groupService services.GroupService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "GetGroupUsers")
		name := c.Params("name")

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Get group users request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "group name is required")
		}

		log.WithContext(c.UserContext()).Info("Processing get group users request",
			"name", name, "requestId", reqInfo.ID)

		// First get the group to get its ID
		group, err := groupService.GetGroupByName(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve group",
				"error", err, "name", name, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		userIDs, err := groupService.GetGroupUsers(c.UserContext(), group.ID)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve group users",
				"error", err, "name", name, "groupId", group.ID, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("Group users retrieved successfully",
			"name", name, "groupId", group.ID, "count", len(userIDs), "requestId", reqInfo.ID)

		return c.JSON(userIDs)
	}
}

// AssignUsersToGroup handles assigning users to a group
// @Summary Assign users to group
// @Description Assign multiple users to a group
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Group name"
// @Param users body UserGroupAssignmentInput true "User assignment request"
// @Success 200 {object} map[string]interface{} "Users assigned successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Group not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/groups/{name}/users [post]
func AssignUsersToGroup(groupService services.GroupService, userService services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "AssignUsersToGroup")
		name := c.Params("name")

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Assign users to group request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "group name is required")
		}

		var input UserGroupAssignmentInput
		if err := c.BodyParser(&input); err != nil {
			log.WithContext(c.UserContext()).Warn("User group assignment request parsing failed",
				"error", err, "requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err
		}

		log.WithContext(c.UserContext()).Info("Processing assign users to group request",
			"name", name, "userCount", len(input.UserUIDs), "requestId", reqInfo.ID)

		// First get the group to get its ID
		group, err := groupService.GetGroupByName(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve group",
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

			err = groupService.AssignUserToGroup(c.UserContext(), user.ID, group.ID)
			if err != nil {
				log.WithContext(c.UserContext()).Warn("Failed to assign user to group",
					"error", err, "groupId", group.ID, "userUID", userUID, "requestId", reqInfo.ID)
				errors = append(errors, userUID+": "+err.Error())
			} else {
				assigned = append(assigned, userUID)
				log.WithContext(c.UserContext()).Info("User assigned to group",
					"groupId", group.ID, "userUID", userUID, "requestId", reqInfo.ID)
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

		log.WithContext(c.UserContext()).Info("User assignment to group completed",
			"name", name, "assigned", len(assigned), "errors", len(errors), "requestId", reqInfo.ID)

		return c.JSON(result)
	}
}

// RemoveUserFromGroup handles removing a user from a group
// @Summary Remove user from group
// @Description Remove a user from a group
// @Tags groups
// @Security BearerAuth
// @Param name path string true "Group name"
// @Param userId path string true "User UID"
// @Success 204 "User removed successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Group or user not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/groups/{name}/users/{userId} [delete]
func RemoveUserFromGroup(groupService services.GroupService, userService services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqInfo := getRequestInfo(c, "RemoveUserFromGroup")
		name := c.Params("name")
		userUID := c.Params("userId") // This is actually userUID now

		if name == "" {
			log.WithContext(c.UserContext()).Warn("Remove user from group request with empty name",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "group name is required")
		}

		if userUID == "" {
			log.WithContext(c.UserContext()).Warn("Remove user from group request with empty user UID",
				"requestId", reqInfo.ID)
			return middleware.SendError(c, fiber.StatusBadRequest, "user UID is required")
		}

		log.WithContext(c.UserContext()).Info("Processing remove user from group request",
			"name", name, "userUID", userUID, "requestId", reqInfo.ID)

		// First get the group to get its ID
		group, err := groupService.GetGroupByName(c.UserContext(), name)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to retrieve group",
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

		err = groupService.RemoveUserFromGroup(c.UserContext(), user.ID, group.ID)
		if err != nil {
			log.WithContext(c.UserContext()).Error("Failed to remove user from group",
				"error", err, "groupId", group.ID, "userUID", userUID, "requestId", reqInfo.ID)
			return middleware.HandleError(c, err)
		}

		log.WithContext(c.UserContext()).Info("User removed from group successfully",
			"name", name, "groupId", group.ID, "userUID", userUID, "requestId", reqInfo.ID)

		return c.SendStatus(fiber.StatusNoContent)
	}
}
