// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package studies

import (
	"aifo.dev/aifo/slideinsight/internal/server/handlers"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/server/services"

	"github.com/gofiber/fiber/v2"
)

// SetupStudyRoutes configures all study related routes with proper RESTful structure and permissions
func SetupStudyRoutes(protectedRoutes fiber.Router, studiesService services.StudiesService, casesService services.CasesService, slidesService services.SlidesService, db ports.Database, imageTypesService services.ImageTypesService) {
	// Study-level routes
	protectedRoutes.Get("/studies", handlers.GetStudies(studiesService))
	protectedRoutes.Get("/studies/count", handlers.GetStudiesCount(studiesService))
	protectedRoutes.Post("/studies", handlers.CreateStudy(studiesService))

	// Individual study routes (require studies.view permission)
	protectedRoutes.Get("/studies/:studyUID",
		middleware.RequireStudyView(db),
		handlers.GetStudyByUID(studiesService))

	protectedRoutes.Get("/studies/:studyUID/metadata",
		middleware.RequireStudyView(db),
		handlers.GetStudyMetadata(studiesService))

	// Metadata field routes (study settings)
	protectedRoutes.Get("/studies/:studyUID/metadata-field",
		middleware.RequireStudyView(db),
		handlers.GetStudyMetadataField(studiesService))

	protectedRoutes.Post("/studies/:studyUID/metadata-field",
		middleware.RequireStudyEdit(db),
		handlers.UpdateStudyMetadataField(studiesService))

	protectedRoutes.Put("/studies/:studyUID",
		middleware.RequireStudyEdit(db),
		handlers.UpdateStudy(studiesService))

	protectedRoutes.Delete("/studies/:studyUID",
		middleware.RequireStudyEdit(db),
		handlers.SoftDeleteStudy(studiesService))

	protectedRoutes.Post("/studies/:studyUID/restore",
		middleware.RequireStudyEdit(db),
		handlers.RestoreStudy(studiesService))

	// Access explanation endpoint (requires authentication but no specific permission)
	protectedRoutes.Get("/studies/:studyUID/access-explanation",
		handlers.ExplainStudyAccess(db))

	// Case routes within studies (RESTful structure)
	protectedRoutes.Get("/studies/:studyUID/cases",
		middleware.RequireStudyView(db),
		handlers.GetCasesByStudyUID(casesService))

	protectedRoutes.Post("/studies/:studyUID/cases",
		middleware.RequireStudyEdit(db),
		handlers.AddCaseToStudy(studiesService))

	// Individual case routes within studies (require cases.view permission with inheritance)
	protectedRoutes.Get("/studies/:studyUID/cases/:caseUID",
		middleware.RequireCaseView(db),
		handlers.GetCaseByUID(casesService))

	protectedRoutes.Put("/studies/:studyUID/cases/:caseUID",
		middleware.RequireCaseEdit(db),
		handlers.UpdateCase(casesService))

	protectedRoutes.Delete("/studies/:studyUID/cases/:caseUID",
		middleware.RequireCaseEdit(db),
		handlers.SoftDeleteCase(casesService))

	// Slide routes within studies and cases (RESTful structure)
	protectedRoutes.Get("/studies/:studyUID/cases/:caseUID/slides",
		middleware.RequireCaseView(db),
		handlers.GetSlidesByCaseUID(slidesService))

	protectedRoutes.Post("/studies/:studyUID/cases/:caseUID/slides",
		middleware.RequireCaseEdit(db),
		handlers.CreateSlide(slidesService, imageTypesService))

	// Individual slide routes within studies and cases (require slides.view permission with inheritance)
	protectedRoutes.Get("/studies/:studyUID/cases/:caseUID/slides/:slideUID",
		middleware.RequireSlideView(db),
		handlers.GetSlideByUID(slidesService))

	protectedRoutes.Put("/studies/:studyUID/cases/:caseUID/slides/:slideUID",
		middleware.RequireSlideEdit(db),
		handlers.UpdateSlide(slidesService, imageTypesService))

	protectedRoutes.Delete("/studies/:studyUID/cases/:caseUID/slides/:slideUID",
		middleware.RequireSlideEdit(db),
		handlers.SoftDeleteSlide(slidesService))

	// Slide metadata and tiles (require slides.view permission with inheritance)
	protectedRoutes.Get("/studies/:studyUID/cases/:caseUID/slides/:slideUID/metadata",
		middleware.RequireSlideView(db),
		handlers.GetSlideMetadata(slidesService))

	protectedRoutes.Get("/studies/:studyUID/cases/:caseUID/slides/:slideUID/tiles/:z/:x/:y.:format",
		middleware.RequireSlideView(db),
		handlers.GetSlideTile(slidesService))
}
