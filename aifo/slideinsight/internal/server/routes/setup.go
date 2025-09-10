// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package routes

import (
	"aifo.dev/aifo/slideinsight/internal/config"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	algorithmroutes "aifo.dev/aifo/slideinsight/internal/server/routes/algorithms"
	annoroutes "aifo.dev/aifo/slideinsight/internal/server/routes/annotations"
	authroutes "aifo.dev/aifo/slideinsight/internal/server/routes/auth"
	caseroutes "aifo.dev/aifo/slideinsight/internal/server/routes/cases"
	grouproutes "aifo.dev/aifo/slideinsight/internal/server/routes/groups"
	notificationroutes "aifo.dev/aifo/slideinsight/internal/server/routes/notifications"
	objectgrantroutes "aifo.dev/aifo/slideinsight/internal/server/routes/objectgrants"
	permissionroutes "aifo.dev/aifo/slideinsight/internal/server/routes/permissions"
	regionsroutes "aifo.dev/aifo/slideinsight/internal/server/routes/regions"
	roleroutes "aifo.dev/aifo/slideinsight/internal/server/routes/roles"
	settingsroutes "aifo.dev/aifo/slideinsight/internal/server/routes/settings"
	slidesroutes "aifo.dev/aifo/slideinsight/internal/server/routes/slides"
	studyroutes "aifo.dev/aifo/slideinsight/internal/server/routes/studies"
	emailtemplatesroutes "aifo.dev/aifo/slideinsight/internal/server/routes/system"
	systemroutes "aifo.dev/aifo/slideinsight/internal/server/routes/system"
	tenantsroutes "aifo.dev/aifo/slideinsight/internal/server/routes/tenants"
	userroutes "aifo.dev/aifo/slideinsight/internal/server/routes/users"
	"github.com/gofiber/fiber/v2"
)

// SetupAllRoutes configures all API routes using domain-specific services where possible
// and ApplicationService only for cross-domain operations
func SetupAllRoutes(apiRoutes fiber.Router, services *ServiceContainer, config config.Config) {
	// Setup authentication routes (auth service is still needed for JWT operations)
	authroutes.SetupAuthRoutes(apiRoutes, services.AuthService, services.Database, config)

	// Create protected routes group
	protectedRoutes := apiRoutes.Group("/v1")
	protectedRoutes.Use(middleware.Protected(config.Auth))

	// Setup all protected routes using domain-specific services
	userroutes.SetupUserRoutes(protectedRoutes, services.UserService, services.RoleService, services.AsyncEmailService)
	tenantsroutes.SetupTenantRoutes(protectedRoutes, services.TenantService)
	caseroutes.SetupCaseRoutes(protectedRoutes, services.CasesService, services.ApplicationService, services.ImageTypesService)
	studyroutes.SetupStudyRoutes(protectedRoutes, services.StudiesService, services.CasesService, services.SlidesService, services.Database, services.ImageTypesService)
	slidesroutes.SetupSlideRoutes(protectedRoutes, services.SlidesService, services.CasesService, services.MasksService, services.VectorAnnotationsService, services.ImageTypesService, services.SlideHistogramsService, services.StainingProtocolsService)
	annoroutes.SetupRasterAnnotationRoutes(protectedRoutes, services.MasksService)
	annoroutes.SetupVectorAnnotationRoutes(protectedRoutes, services.VectorAnnotationsService, services.StudiesService)
	regionsroutes.SetupRegionRoutes(protectedRoutes, services.RegionsService)

	// Setup settings routes
	settingsroutes.SetupSettingsRoutes(protectedRoutes, services.SettingsService)

	// Setup notification routes
	notificationroutes.SetupNotificationRoutes(protectedRoutes, services.NotificationService, services.UserService, config)

	// Setup permission routes
	permissionroutes.SetupPermissionRoutes(apiRoutes, services.PermissionService, config)

	// Setup role routes
	roleroutes.SetupRoleRoutes(apiRoutes, services.RoleService, services.UserService, config)

	// Setup group route
	grouproutes.SetupGroupRoutes(apiRoutes, services.GroupService, services.UserService, config)

	// Setup object grant routes
	objectgrantroutes.SetupObjectGrantRoutes(apiRoutes, services.ObjectGrantService, config)

	// Setup system routes
	systemroutes.SetupSystemRoutes(apiRoutes, services.QueueManager, config)

	// Setup email template routes
	emailtemplatesroutes.SetupEmailTemplateRoutes(apiRoutes, services.EmailTemplateService, config)

	// Setup algorithm routes
	algorithmroutes.SetupAlgorithmRoutes(protectedRoutes, services.AlgorithmsService)
}
