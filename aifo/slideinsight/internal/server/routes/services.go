// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package routes

import (
	"aifo.dev/aifo/slideinsight/internal/queue"
	"aifo.dev/aifo/slideinsight/internal/server/email"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	authService "aifo.dev/aifo/slideinsight/internal/server/services/auth"
)

// ServiceContainer holds all services needed for route setup
type ServiceContainer struct {
	Database                 ports.Database
	SlidesService            services.SlidesService
	MasksService             services.RasterAnnotationsService
	VectorAnnotationsService services.VectorAnnotationsService
	RegionsService           services.RegionsService
	TenantService            services.TenantsService
	UserService              services.UserService
	StudiesService           services.StudiesService
	AuthService              authService.AuthService
	CasesService             services.CasesService
	ApplicationService       services.ApplicationService
	PermissionService        services.PermissionService
	RoleService              services.RoleService
	GroupService             services.GroupService
	ObjectGrantService       services.ObjectGrantService
	NotificationService      services.NotificationService
	EmailService             ports.EmailService
	AsyncEmailService        email.AsyncEmailService
	EmailTemplateService     services.EmailTemplateService
	AlgorithmsService        services.AlgorithmsService
	SettingsService          *services.SettingsService
	ImageTypesService        services.ImageTypesService
	SlideHistogramsService   services.SlideHistogramsService
	StainingProtocolsService services.StainingProtocolsService
	QueueManager             *queue.QueueManager
}

// GetSlidesService returns the slides service for cache monitoring
func (sc *ServiceContainer) GetSlidesService() services.SlidesService {
	return sc.SlidesService
}

// GetRasterAnnotationsService returns the raster annotations service for cache monitoring
func (sc *ServiceContainer) GetRasterAnnotationsService() services.RasterAnnotationsService {
	return sc.MasksService
}
