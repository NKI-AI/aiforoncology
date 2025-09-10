// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package tenants

import (
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// SetupTenantRoutes configures all tenant related routes
func SetupTenantRoutes(protectedRoutes fiber.Router, tenantsService services.TenantsService) {
	// Protected tenant routes
	protectedRoutes.Get("/tenants", handlers.GetTenants(tenantsService))
	protectedRoutes.Get("/tenants/count", handlers.GetTenantsCount(tenantsService))
	protectedRoutes.Get("/tenants/:tenantUID", handlers.GetTenantByUID(tenantsService))
	protectedRoutes.Post("/tenants", handlers.CreateTenant(tenantsService))
	protectedRoutes.Put("/tenants/:tenantUID", handlers.UpdateTenant(tenantsService))
	protectedRoutes.Delete("/tenants/:tenantUID", handlers.DeleteTenant(tenantsService))

	// Domain management routes
	protectedRoutes.Get("/tenants/:tenantUID/domains", handlers.GetTenantDomains(tenantsService))
	protectedRoutes.Post("/tenants/:tenantUID/domains", handlers.AddTenantDomain(tenantsService))
	protectedRoutes.Put("/tenants/:tenantUID/domains/:domainUID", handlers.UpdateTenantDomain(tenantsService))
	protectedRoutes.Delete("/tenants/:tenantUID/domains/:domainUID", handlers.RemoveTenantDomain(tenantsService))
}
