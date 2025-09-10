// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// GetSearchConfigs returns search configurations for all resources
// @Summary Get search configurations
// @Description Retrieve search configurations for all resources to help frontend build search forms
// @Tags utils
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]utils.SearchConfigResponse "Search configurations for all resources"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/search-configs [get]
func GetSearchConfigs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		configs := utils.GetAllSearchConfigs()
		return c.JSON(configs)
	}
}

// GetSearchConfigForResource returns search configuration for a specific resource
// @Summary Get search configuration for resource
// @Description Retrieve search configuration for a specific resource type
// @Tags utils
// @Produce json
// @Security BearerAuth
// @Param resource path string true "Resource name" Enums(cases,slides,studies,tenants,users) example("cases")
// @Success 200 {object} utils.SearchConfigResponse "Search configuration for the resource"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid resource"
// @Failure 404 {object} domain.ErrorResponse "Resource configuration not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/search-configs/{resource} [get]
func GetSearchConfigForResource() fiber.Handler {
	return func(c *fiber.Ctx) error {
		resource := c.Params("resource")
		if resource == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "resource parameter is required")
		}

		config, exists := utils.GetSearchConfigForResource(resource)
		if !exists {
			log.Warn("Search configuration not found for resource", "resource", resource)
			return middleware.SendError(c, fiber.StatusNotFound, "search configuration not found for resource: "+resource)
		}

		return c.JSON(config)
	}
}

// GetPaginationConfig returns pagination configuration information
// @Summary Get pagination configuration
// @Description Retrieve pagination configuration details for the application
// @Tags utils
// @Produce json
// @Security BearerAuth
// @Success 200 {object} fiber.Map "Pagination configuration"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/pagination-config [get]
func GetPaginationConfig() fiber.Handler {
	return func(c *fiber.Ctx) error {
		config := fiber.Map{
			"default": utils.DefaultPaginationOptions(),
			"users":   utils.DefaultUsersOptions(),
			"studies": utils.DefaultStudiesOptions(),
			"tenants": utils.DefaultTenantsOptions(),
		}
		return c.JSON(config)
	}
}
