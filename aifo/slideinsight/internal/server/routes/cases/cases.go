// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package cases

import (
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// SetupCaseRoutes configures all case related routes
func SetupCaseRoutes(protectedRoutes fiber.Router, casesService services.CasesService, appService services.ApplicationService, imageTypesService services.ImageTypesService) {
	protectedRoutes.Get("/cases", handlers.GetCases(casesService))
	protectedRoutes.Get("/cases/count", handlers.GetCasesCount(casesService))
	protectedRoutes.Get("/cases/:caseUID", handlers.GetCaseByUID(casesService))
	protectedRoutes.Post("/cases", handlers.CreateCase(casesService))

	protectedRoutes.Get("/studies/:studyUID/cases", handlers.GetCasesByStudyUID(casesService))
	protectedRoutes.Get("/studies/:studyUID/cases/:caseUID/neighbors", handlers.GetCaseNeighborsByStudyUID(casesService))

	protectedRoutes.Post("/cases/:caseUID/slides", handlers.AddSlideToCase(casesService, imageTypesService))

	protectedRoutes.Delete("/cases/:caseUID", handlers.SoftDeleteCase(casesService))
	protectedRoutes.Get("/cases/deleted", handlers.GetDeletedCases(casesService))
	protectedRoutes.Post("/cases/:caseUID/restore", handlers.RestoreCase(casesService))
}
