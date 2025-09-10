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

// tenantsResponseBuilder builds a TenantsResponse from tenants and pagination info
func tenantsResponseBuilder(tenants []domain.Tenant, pagination domain.PaginationInfo) domain.TenantsResponse {
	return domain.TenantsResponse{
		Tenants:    tenants,
		Pagination: pagination,
	}
}

// CreateTenant creates a new tenant
// @Summary Create a new tenant
// @Description Create a new tenant
// @Tags tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenant body TenantsInput true "Tenant information"
// @Success 201 {object} domain.Tenant "Created tenant"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 409 {object} domain.ErrorResponse "Conflict - tenant already exists"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/tenants [post]
func CreateTenant(service services.TenantsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input TenantsInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("CreateTenant request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request body")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		tenant := domain.Tenant{
			Name:        input.Name,
			Description: input.Description,
		}

		createdTenant, err := service.SaveTenant(c.UserContext(), tenant)
		if err != nil {
			log.Error("CreateTenant failed", "error", err, "name", tenant.Name)
			return middleware.HandleError(c, err)
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(createdTenant)
	}
}

// GetTenantByUID retrieves a tenant by ID (using path parameter)
// @Summary Get tenant by ID
// @Description Retrieve tenant information by ID
// @Tags tenants
// @Produce json
// @Security BearerAuth
// @Param tenantUid path string true "Tenant UID" example("Xyxz2234")
// @Success 200 {object} domain.Tenant "Tenant information"
// @Failure 400 {object} domain.ErrorResponse "Bad request - tenant UID required"
// @Failure 404 {object} domain.ErrorResponse "Tenant not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/tenants/{tentantUid} [get]
func GetTenantByUID(service services.TenantsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params TenantUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetTenantByUID params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid path parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		tenant, err := service.GetTenantByUID(c.UserContext(), params.TenantUID)
		if err != nil {
			log.Error("GetTenantByUID failed", "error", err, "tenantUID", params.TenantUID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(tenant)
	}
}

// UpdateTenant updates tenant information
// @Summary Update tenant information
// @Description Update tenant information by tenant UID
// @Tags tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantUid path string true "Tenant UID" example("Xyxz2234")
// @Param tenant body TenantUpdateInput true "Tenant update information"
// @Success 200 {object} domain.Tenant "Updated tenant information"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Tenant not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/tenants/{tenantUid} [put]
func UpdateTenant(service services.TenantsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params TenantUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("UpdateTenant params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid path parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		var input TenantUpdateInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("UpdateTenant request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		updates := domain.TenantUpdates{
			Name:        input.Name,
			Description: input.Description,
			Status:      input.Status,
		}

		err := service.UpdateTenant(c.UserContext(), params.TenantUID, updates)
		if err != nil {
			log.Error("UpdateTenant failed", "error", err, "tenantUID", params.TenantUID)
			return middleware.HandleError(c, err)
		}

		// Get the updated tenant to return
		updatedTenant, err := service.GetTenantByUID(c.UserContext(), params.TenantUID)
		if err != nil {
			log.Error("GetTenantByUID after update failed", "error", err, "tenantUID", params.TenantUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.JSON(updatedTenant)
	}
}

// GetTenants returns a handler using the new generic pattern
// @Summary Get all tenants (generic)
// @Description Retrieve a list of tenants using the new generic pagination and search pattern
// @Tags tenants
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 20)" minimum(1) maximum(50) example(20)
// @Param q query string false "General search across name and short_uid" example("acme")
// @Param name query string false "Filter by tenant name" example("acme-corp")
// @Param sort query string false "Sort field (name, created_at, short_uid)" example("name")
// @Param dir query string false "Sort direction (asc, desc)" example("asc")
// @Success 200 {object} domain.TenantsResponse "List of tenants with pagination"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/tenants/generic [get]
func GetTenants(service services.TenantsService) fiber.Handler {
	return utils.GetPaginatedResourcesHandler(
		utils.DefaultTenantsSearchConfig(),
		service.GetTenantsGeneric,
		tenantsResponseBuilder,
	)
}

// GetTenantsCount returns a handler function that retrieves the total count of tenants
// @Summary Get tenants count
// @Description Retrieve the total count of tenants in the system
// @Tags tenants
// @Produce json
// @Security BearerAuth
// @Success 200 {object} fiber.Map "Total count of tenants"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/tenants/count [get]
func GetTenantsCount(service services.TenantsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		count, err := service.GetTenantsCount(c.UserContext())
		if err != nil {
			log.Error("GetTenantsCount failed", "error", err)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{
			"count": count,
		})
	}
}

// GetTenantsCountGeneric returns a generic handler for tenant count using the new pattern
// @Summary Get tenants count (generic)
// @Description Retrieve the total count of tenants using the generic pattern
// @Tags tenants
// @Produce json
// @Security BearerAuth
// @Success 200 {object} fiber.Map "Total count of tenants"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/tenants/count/generic [get]
func GetTenantsCountGeneric(service services.TenantsService) fiber.Handler {
	return utils.GetCountHandler(service.GetTenantsCount)
}

// Domain Management Handlers

// GetTenantDomains retrieves all domains for a specific tenant
// @Summary Get tenant domains
// @Description Retrieve all domains associated with a tenant
// @Tags tenants
// @Produce json
// @Security BearerAuth
// @Param tenantUid path string true "Tenant UID" example("Xyxz2234")
// @Success 200 {object} domain.TenantDomainsResponse "List of tenant domains"
// @Failure 400 {object} domain.ErrorResponse "Bad request - tenant ID required"
// @Failure 404 {object} domain.ErrorResponse "Tenant not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/tenants/{tenantUid}/domains [get]
func GetTenantDomains(service services.TenantsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params TenantUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetTenantDomains params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid path parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		domains, err := service.GetTenantDomains(c.UserContext(), params.TenantUID)
		if err != nil {
			log.Error("GetTenantDomains failed", "error", err, "tenantUID", params.TenantUID)
			return middleware.HandleError(c, err)
		}

		response := domain.TenantDomainsResponse{
			Domains: domains,
		}
		return c.JSON(response)
	}
}

// AddTenantDomain adds a new domain to a tenant
// @Summary Add domain to tenant
// @Description Add a new domain to a tenant for domain-based registration
// @Tags tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantUid path string true "Tenant ID" example("Xyxz2234")
// @Param domain body AddDomainInput true "Domain information"
// @Success 201 {object} fiber.Map "Domain added successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Tenant not found"
// @Failure 409 {object} domain.ErrorResponse "Domain already exists"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/tenants/{tenantUid}/domains [post]
func AddTenantDomain(service services.TenantsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params TenantUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("AddTenantDomain params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid path parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		var input AddDomainInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("AddTenantDomain request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request body")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		request := domain.NewTenantDomainRequest{
			Domain:    input.Domain,
			IsPrimary: input.IsPrimary,
		}

		err := service.AddTenantDomain(c.UserContext(), params.TenantUID, request)
		if err != nil {
			log.Error("AddTenantDomain failed", "error", err, "tenantUID", params.TenantUID, "domain", input.Domain)
			return middleware.HandleError(c, err)
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(fiber.Map{"message": "Domain added successfully"})
	}
}

// UpdateTenantDomain updates domain verification or primary status
// @Summary Update tenant domain
// @Description Update domain verification status or primary flag
// @Tags tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantUid path string true "Tenant UID" example("Xyxz2234")
// @Param domainUid path int true "Domain UID" example(123)
// @Param updates body UpdateDomainInput true "Domain updates"
// @Success 200 {object} fiber.Map "Domain updated successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Domain not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/tenants/{tenantUid}/domains/{domainUid} [put]
func UpdateTenantDomain(service services.TenantsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params TenantUIDDomainUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("UpdateTenantDomain params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid path parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		var input UpdateDomainInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("UpdateTenantDomain request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request body")
		}

		updates := domain.TenantDomainUpdates{
			IsVerified: input.IsVerified,
			IsPrimary:  input.IsPrimary,
		}

		err := service.UpdateTenantDomain(c.UserContext(), params.DomainUID, updates)
		if err != nil {
			log.Error("UpdateTenantDomain failed", "error", err, "tenantUID", params.TenantUID, "domainUID", params.DomainUID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{"message": "Domain updated successfully"})
	}
}

// RemoveTenantDomain removes a domain from a tenant
// @Summary Remove tenant domain
// @Description Remove a domain from a tenant
// @Tags tenants
// @Produce json
// @Security BearerAuth
// @Param tenantUid path string true "Tenant UID" example("Xyxz2234")
// @Param domainUid path int true "Domain UID" example(123)
// @Success 200 {object} fiber.Map "Domain removed successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 404 {object} domain.ErrorResponse "Domain not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/tenants/{tenantUid}/domains/{domainUid} [delete]
func RemoveTenantDomain(service services.TenantsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params TenantUIDDomainUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("RemoveTenantDomain params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid path parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		err := service.RemoveTenantDomain(c.UserContext(), params.DomainUID)
		if err != nil {
			log.Error("RemoveTenantDomain failed", "error", err, "tenantUID", params.TenantUID, "domainUID", params.DomainUID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{"message": "Domain removed successfully"})
	}
}

// DeleteTenant permanently deletes a tenant
// @Summary Delete tenant
// @Description Permanently delete a tenant and all associated data
// @Tags tenants
// @Produce json
// @Security BearerAuth
// @Param tenantUid path string true "Tenant UID" example("Xyxz2234")
// @Success 200 {object} fiber.Map "Tenant deleted successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - tenant ID required"
// @Failure 404 {object} domain.ErrorResponse "Tenant not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/tenants/{tenantUid} [delete]
func DeleteTenant(service services.TenantsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params TenantUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("DeleteTenant params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid path parameters")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		err := service.DeleteTenant(c.UserContext(), params.TenantUID)
		if err != nil {
			log.Error("DeleteTenant failed", "error", err, "tenantUID", params.TenantUID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{"message": "Tenant deleted successfully"})
	}
}
