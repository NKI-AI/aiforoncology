// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
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

// settingsResponseBuilder builds a SettingsResponse from settings and pagination info
func settingsResponseBuilder(settings []domain.Setting, pagination domain.PaginationInfo) domain.SettingsResponse {
	return domain.SettingsResponse{
		Settings:   settings,
		Pagination: pagination,
	}
}

// GetSettings returns a handler using the generic pattern
// @Summary Get all settings (generic)
// @Description Retrieve a list of settings using the generic pagination and search pattern
// @Tags settings
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 20)" minimum(1) maximum(50) example(20)
// @Param q query string false "General search across key and value" example("enable")
// @Param key query string false "Filter by setting key" example("enable_registration")
// @Param value_type query string false "Filter by value type" example("boolean")
// @Param tenant_id query string false "Filter by tenant ID" example("1")
// @Param sort query string false "Sort field (key, created_at, tenant_id)" example("key")
// @Param dir query string false "Sort direction (asc, desc)" example("asc")
// @Success 200 {object} domain.SettingsResponse "List of settings with pagination"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/settings [get]
func GetSettings(service *services.SettingsService) fiber.Handler {
	return utils.GetPaginatedResourcesHandler(
		utils.SearchConfig{
			SearchableFields: []string{"key", "value"},
			FilterableFields: []string{"key", "value_type", "tenant_id"},
			SortableFields:   []string{"key", "created_at", "tenant_id", "value_type"},
		},
		service.GetSettingsGeneric,
		settingsResponseBuilder,
	)
}

// GetSetting retrieves a specific setting by tenant ID and key
// @Summary Get setting by tenant and key
// @Description Retrieve setting information by tenant ID and key
// @Tags settings
// @Produce json
// @Security BearerAuth
// @Param tenant_id path int true "Tenant ID" example(1)
// @Param key path string true "Setting key" example("enable_registration")
// @Success 200 {object} domain.Setting "Setting information"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 404 {object} domain.ErrorResponse "Setting not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/settings/{tenant_id}/{key} [get]
func GetSetting(service *services.SettingsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params struct {
			TenantID int    `params:"tenant_id" validate:"required"`
			Key      string `params:"key" validate:"required"`
		}

		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetSetting params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid path parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		setting, err := service.GetSetting(c.UserContext(), params.TenantID, params.Key)
		if err != nil {
			log.Error("GetSetting failed", "error", err, "tenantID", params.TenantID, "key", params.Key)
			return middleware.HandleError(c, err)
		}

		return c.JSON(setting)
	}
}

// CreateSetting creates a new setting
// @Summary Create a new setting
// @Description Create a new setting
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param setting body SettingInput true "Setting information"
// @Success 201 {object} domain.Setting "Created setting"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 409 {object} domain.ErrorResponse "Conflict - setting already exists"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/settings [post]
func CreateSetting(service *services.SettingsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input SettingInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("CreateSetting request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request body")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		newSetting := domain.NewSetting{
			TenantID:  input.TenantID,
			Key:       input.Key,
			ValueType: input.ValueType,
			Value:     input.Value,
		}

		setting, err := service.CreateSetting(c.UserContext(), newSetting)
		if err != nil {
			log.Error("CreateSetting failed", "error", err, "key", newSetting.Key)
			return middleware.HandleError(c, err)
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(setting)
	}
}

// UpdateSetting updates an existing setting
// @Summary Update setting information
// @Description Update setting information by tenant ID and key
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenant_id path int true "Tenant ID" example(1)
// @Param key path string true "Setting key" example("enable_registration")
// @Param setting body SettingUpdateInput true "Setting update information"
// @Success 200 {object} domain.Setting "Updated setting information"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Setting not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/settings/{tenant_id}/{key} [put]
func UpdateSetting(service *services.SettingsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var pathParams struct {
			TenantID int    `params:"tenant_id" validate:"required"`
			Key      string `params:"key" validate:"required"`
		}

		if err := c.ParamsParser(&pathParams); err != nil {
			log.Warn("UpdateSetting path params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid path parameters")
		}

		// Validate the parsed path parameters
		if err := validation.GlobalValidator.ValidateStruct(c, pathParams); err != nil {
			return err // Error already formatted and sent
		}

		var input SettingUpdateInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("UpdateSetting request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request body")
		}

		updates := domain.SettingUpdates{
			ValueType: input.ValueType,
			Value:     input.Value,
		}

		setting, err := service.UpdateSetting(c.UserContext(), pathParams.TenantID, pathParams.Key, updates)
		if err != nil {
			log.Error("UpdateSetting failed", "error", err, "tenantID", pathParams.TenantID, "key", pathParams.Key)
			return middleware.HandleError(c, err)
		}

		return c.JSON(setting)
	}
}

// DeleteSetting deletes a setting
// @Summary Delete setting
// @Description Delete a setting by tenant ID and key
// @Tags settings
// @Security BearerAuth
// @Param tenant_id path int true "Tenant ID" example(1)
// @Param key path string true "Setting key" example("enable_registration")
// @Success 204 "Setting deleted successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 404 {object} domain.ErrorResponse "Setting not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/settings/{tenant_id}/{key} [delete]
func DeleteSetting(service *services.SettingsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params struct {
			TenantID int    `params:"tenant_id" validate:"required"`
			Key      string `params:"key" validate:"required"`
		}

		if err := c.ParamsParser(&params); err != nil {
			log.Warn("DeleteSetting params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid path parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		err := service.DeleteSetting(c.UserContext(), params.TenantID, params.Key)
		if err != nil {
			log.Error("DeleteSetting failed", "error", err, "tenantID", params.TenantID, "key", params.Key)
			return middleware.HandleError(c, err)
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

// GetSettingsCount returns a handler function that retrieves the total count of settings
// @Summary Get settings count
// @Description Retrieve the total count of settings in the system
// @Tags settings
// @Produce json
// @Security BearerAuth
// @Success 200 {object} fiber.Map "Total count of settings"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/settings/count [get]
func GetSettingsCount(service *services.SettingsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		count, err := service.GetSettingsCount(c.UserContext())
		if err != nil {
			log.Error("GetSettingsCount failed", "error", err)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{"count": count})
	}
}
