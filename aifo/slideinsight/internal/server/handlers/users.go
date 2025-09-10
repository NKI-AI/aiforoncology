// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/email"
	apperrors "aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"aifo.dev/aifo/slideinsight/internal/server/validation"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// usersResponseBuilder builds a UsersResponse from users and pagination info
func usersResponseBuilder(users []domain.User, pagination domain.PaginationInfo) domain.UsersResponse {
	return domain.UsersResponse{
		Users:      users,
		Pagination: pagination,
	}
}

// CreateUser creates a new user
// @Summary Create a new user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body UserInput true "User information"
// @Success 201 {object} domain.User "Created user (password field will be empty)"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/users [post]
func CreateUser(service services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input UserInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("CreateUser request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		user := domain.User{
			Email:     input.Email,
			FirstName: input.FirstName,
			LastName:  input.LastName,
			Password:  input.Password,
			TenantUID: input.TenantUID,
		}

		createdUser, err := service.CreateUser(c.UserContext(), user)
		if err != nil {
			log.Error("CreateUser failed", "error", err, "email", user.Email)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(createdUser)
	}
}

// GetUserByUID retrieves a user by UID
// @Summary Get user by UID
// @Description Retrieve user information by UID
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param userUid path string true "User UID" example("abc123")
// @Success 200 {object} domain.User "User information"
// @Failure 400 {object} domain.ErrorResponse "Bad request - user UID required"
// @Failure 404 {object} domain.ErrorResponse "User not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/users/{userUID} [get]
func GetUserByUID(service services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input UserUIDParams
		if err := c.ParamsParser(&input); err != nil {
			log.Warn("GetUserByUID request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		user, err := service.GetUserByUID(c.UserContext(), input.UserUID)
		if err != nil {
			log.Error("GetUserByUID failed", "error", err, "userUID", input.UserUID)
			// Check for specific error cases using proper error types
			if apperrors.IsUserNotFound(err) {
				return middleware.SendError(c, fiber.StatusNotFound, "user not found")
			}
			return middleware.HandleError(c, err)
		}

		// TODO: Return a sanitized version of the user without password
		return c.JSON(user)
	}
}

// UpdateUser updates user information
// @Summary Update user information
// @Description Update user information (excluding password) by user UID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userUid path string true "User UID" example("abc123")
// @Param user body UserUpdateInput true "User update information"
// @Success 200 {object} domain.User "Updated user information"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "User not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/users/{userUID} [put]
func UpdateUser(service services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var pathParams UserUIDParams
		if err := c.ParamsParser(&pathParams); err != nil {
			log.Warn("UpdateUser path params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed path parameters
		if err := validation.GlobalValidator.ValidateStruct(c, pathParams); err != nil {
			return err // Error already formatted and sent
		}

		var input UserUpdateInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("UpdateUser request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		updates := domain.UserUpdates{
			Email:             input.Email,
			FirstName:         input.FirstName,
			LastName:          input.LastName,
			MustResetPassword: input.MustResetPassword,
			IsActive:          input.IsActive,
		}

		err := service.UpdateUserByUID(c.UserContext(), pathParams.UserUID, updates)
		if err != nil {
			log.Error("UpdateUserByUID failed", "error", err, "userUID", pathParams.UserUID)
			// Check for specific error cases using proper error types
			if apperrors.IsUserNotFound(err) {
				return middleware.SendError(c, fiber.StatusNotFound, "user not found")
			}
			return middleware.HandleError(c, err)
		}

		// Get the updated user to return
		updatedUser, err := service.GetUserByUID(c.UserContext(), pathParams.UserUID)
		if err != nil {
			log.Error("GetUserByUID after update failed", "error", err, "userUID", pathParams.UserUID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(updatedUser)
	}
}

// DeleteUser deletes a user
// @Summary Delete user
// @Description Delete a user by user UID after checking for dependencies
// @Tags users
// @Security BearerAuth
// @Param userUid path string true "User UID" example("abc123")
// @Success 204 "User deleted successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - user UID required"
// @Failure 404 {object} domain.ErrorResponse "User not found"
// @Failure 409 {object} domain.ErrorResponse "User has dependencies"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/users/{userUid} [delete]
func DeleteUser(service services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input UserUIDParams
		if err := c.ParamsParser(&input); err != nil {
			log.Warn("DeleteUser request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		err := service.DeleteUser(c.UserContext(), input.UserUID)
		if err != nil {
			log.Error("DeleteUser failed", "error", err, "userUID", input.UserUID)
			// Check for specific error cases using proper error types
			if apperrors.IsUserNotFound(err) {
				return middleware.SendError(c, fiber.StatusNotFound, "user not found")
			}
			// Check for dependency conflicts (when user cannot be deleted due to dependencies)
			if apperrors.IsUserHasDependencies(err) {
				return middleware.SendError(c, fiber.StatusConflict, err.Error())
			}
			return middleware.HandleError(c, err)
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

// GetUsers returns a handler function that retrieves all users with optional search/filter support
// @Summary Get all users
// @Description Retrieve a list of users with optional search/filter and pagination support
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 50)" minimum(1) maximum(200) example(50)
// @Param q query string false "General search across email, first_name, last_name" example("admin")
// @Param email query string false "Filter by email" example("user@example.com")
// @Param first_name query string false "Filter by first name" example("John")
// @Param last_name query string false "Filter by last name" example("Doe")
// @Param is_active query string false "Filter by active status" example("true")
// @Param must_reset_password query string false "Filter by must reset password" example("true")
// @Param sort query string false "Sort field (email, created_at, short_uid)" example("email")
// @Param dir query string false "Sort direction (asc, desc)" example("asc")
// @Success 200 {object} domain.UsersResponse "List of users with pagination"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/users [get]
func GetUsers(service services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		params, err := utils.ParsePaginationAndSearchParamsWithConfig(c, utils.DefaultPaginationOptions(), utils.DefaultUsersSearchConfig())
		if err != nil {
			return err // Error already contains proper status code and message
		}

		// Use search method which handles both search and non-search cases
		users, paginationInfo, err := service.GetUsersGeneric(c.UserContext(), params)
		if err != nil {
			log.Error("GetUsers failed", "error", err, "search", params.SearchParams)
			return middleware.HandleError(c, err)
		}

		return c.JSON(domain.UsersResponse{
			Users:      users,
			Pagination: paginationInfo,
		})
	}
}

// GetUsersCount returns a handler function that retrieves the total count of users
// @Summary Get users count
// @Description Retrieve the total count of users in the system
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} fiber.Map "Total count of users"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/users/count [get]
func GetUsersCount(service services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		count, err := service.GetUserCount(c.UserContext())
		if err != nil {
			log.Error("GetUsersCount failed", "error", err)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{
			"count": count,
		})
	}
}

// SendUserEmail sends an email to a specific user
// @Summary Send email to user
// @Description Send an email to a user using predefined templates
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userUid path string true "User UID" example("abc123")
// @Param emailRequest body SendUserEmailInput true "Email request information"
// @Success 200 {object} fiber.Map "Email sent successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "User not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/users/{userUID}/send-email [post]
func SendUserEmail(userService services.UserService, asyncEmailService email.AsyncEmailService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var pathParams UserUIDParams
		if err := c.ParamsParser(&pathParams); err != nil {
			log.Warn("SendUserEmail path params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed path parameters
		if err := validation.GlobalValidator.ValidateStruct(c, pathParams); err != nil {
			return err // Error already formatted and sent
		}

		var input SendUserEmailInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("SendUserEmail request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// Get the user to send email to
		user, err := userService.GetUserByUID(c.UserContext(), pathParams.UserUID)
		if err != nil {
			log.Error("SendUserEmail - GetUserByUID failed", "error", err, "userUID", pathParams.UserUID)
			if apperrors.IsUserNotFound(err) {
				return middleware.SendError(c, fiber.StatusNotFound, "user not found")
			}
			return middleware.HandleError(c, err)
		}

		// Prepare email data based on template
		emailData := map[string]interface{}{
			"email":     user.Email,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"tenantId":  user.TenantID,
		}

		// Add additional data based on template type
		switch input.Template {
		case "password_reset":
			// For password reset, we'd normally generate a token, but for admin-initiated emails
			// we'll send a generic message asking user to use the forgot password feature
			emailData["token"] = "Please use the 'Forgot Password' feature on the login page"
		case "email_verification":
			// For email verification, similar approach
			emailData["token"] = "Please contact your administrator for email verification"
		}

		// Create email request
		emailRequest := ports.EmailRequest{
			To:       user.Email,
			Template: ports.EmailTemplateType(input.Template),
			Data:     emailData,
			TenantID: user.TenantID,
		}

		// Override subject if provided
		if input.Subject != "" {
			emailRequest.Subject = input.Subject
		}

		// Send the email asynchronously via queue
		err = asyncEmailService.SendEmailAsync(c.UserContext(), emailRequest)
		if err != nil {
			log.Error("SendUserEmail failed", "error", err, "userUID", pathParams.UserUID, "email", user.Email)
			return middleware.SendError(c, fiber.StatusInternalServerError, "failed to send email")
		}

		log.Info("Email sent successfully", "userUID", pathParams.UserUID, "email", user.Email, "template", input.Template)
		return c.JSON(fiber.Map{
			"message":  "Email sent successfully",
			"email":    user.Email,
			"template": input.Template,
		})
	}
}

// GetUserRoles retrieves all roles assigned to a user
// @Summary Get user roles
// @Description Retrieve all roles assigned to a specific user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param userUid path string true "User UID" example("abc123")
// @Success 200 {array} string "User role names"
// @Failure 400 {object} domain.ErrorResponse "Bad request - user UID required"
// @Failure 404 {object} domain.ErrorResponse "User not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/users/{userUID}/roles [get]
func GetUserRoles(userService services.UserService, roleService services.RoleService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params UserUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetUserRoles path params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed path parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		// Get the user to convert UID to internal ID
		user, err := userService.GetUserByUID(c.UserContext(), params.UserUID)
		if err != nil {
			log.Error("GetUserRoles - GetUserByUID failed", "error", err, "userUID", params.UserUID)
			if apperrors.IsUserNotFound(err) {
				return middleware.SendError(c, fiber.StatusNotFound, "user not found")
			}
			return middleware.HandleError(c, err)
		}

		// Get user roles using the role service
		userRoles, err := roleService.GetUserRoles(c.UserContext(), user.ID)
		if err != nil {
			log.Error("GetUserRoles - roleService.GetUserRoles failed", "error", err, "userUID", params.UserUID, "userID", user.ID)
			return middleware.HandleError(c, err)
		}

		// Look up role names by role IDs
		roleNames := make([]string, 0, len(userRoles))
		for _, userRole := range userRoles {
			role, err := roleService.GetRoleByID(c.UserContext(), userRole.RoleID)
			if err != nil {
				log.Error("GetUserRoles - failed to get role by ID", "error", err, "roleID", userRole.RoleID, "userUID", params.UserUID)
				// Skip this role instead of failing entirely
				continue
			}
			roleNames = append(roleNames, role.Name)
		}

		log.Info("Successfully retrieved user roles", "userUID", params.UserUID, "userID", user.ID, "roleCount", len(roleNames))
		return c.JSON(roleNames)
	}
}
