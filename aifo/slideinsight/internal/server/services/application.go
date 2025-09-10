// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"
	"fmt"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
)

// ApplicationService handles all operations exposed to the routes layer
// This follows hexagonal architecture principles by providing a single interface
// for the infrastructure layer (routes/handlers) to interact with the application core
type ApplicationService interface {
	// User operations
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUserByUID(ctx context.Context, userUID string) (domain.User, error)
	GetUsers(ctx context.Context, pagination utils.PaginationParams) ([]domain.User, domain.PaginationInfo, error)

	// Tenant operations
	GetTenants(ctx context.Context, pagination utils.PaginationParams) ([]domain.Tenant, domain.PaginationInfo, error)
	GetTenantsCount(ctx context.Context) (int, error)
	GetTenantByUID(ctx context.Context, tenantUID string) (domain.Tenant, error)
	CreateTenant(ctx context.Context, tenant domain.Tenant) (domain.Tenant, error)

	// Case operations
	GetCases(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Case, domain.PaginationInfo, error)
	GetCasesByStudyUID(ctx context.Context, studyUID string, params utils.PaginationAndSearchParams) ([]domain.Case, domain.PaginationInfo, error)
	GetCasesCount(ctx context.Context) (int, error)
	GetCaseByUID(ctx context.Context, caseUID string) (domain.Case, error)
	CreateCase(ctx context.Context, newCase domain.Case) (domain.Case, error)
	AddCaseToStudy(ctx context.Context, studyUID string, caseUID string) error
	CreateCaseAndAssignToStudy(ctx context.Context, newCase domain.Case, studyUID string, tenantID, creatorID int) (domain.Case, error)
	SoftDeleteCase(ctx context.Context, caseUID string, deletedBy int) error
	GetDeletedCases(ctx context.Context) ([]domain.Case, error)
	RestoreCase(ctx context.Context, caseUID string) error
	GetCasesByStudyUIDWithAnnotations(ctx context.Context, studyUID string, params utils.PaginationAndSearchParams) ([]domain.Case, domain.PaginationInfo, error)
	GetCaseByUIDWithAnnotations(ctx context.Context, caseUID string) (domain.Case, error)

	// Study operations
	GetStudies(ctx context.Context, pagination utils.PaginationParams) ([]domain.Study, domain.PaginationInfo, error)
	GetStudiesCount(ctx context.Context, search utils.SearchParams) (int, error)
	GetStudyByUID(ctx context.Context, studyUID string) (domain.Study, error)
	CreateStudy(ctx context.Context, study domain.Study) (domain.Study, error)

	// Slide operations
	GetSlides(ctx context.Context, pagination utils.PaginationParams) ([]domain.Slide, domain.PaginationInfo, error)
	GetSlidesCount(ctx context.Context) (int, error)
	GetSlideByUID(ctx context.Context, slideUID string) (domain.Slide, error)
	GetSlidesByCaseUID(ctx context.Context, caseUID string) ([]domain.Slide, error)
	GetSlideMetadata(ctx context.Context, slideUID string) (domain.SlideMetadata, error)
	GetSlideTile(ctx context.Context, slideUID string, z, x, y int, format string, quality int) (domain.SlideTile, error)
	CreateSlide(ctx context.Context, newSlide domain.Slide) (domain.Slide, error)

	// Mask operations
	GetRasterAnnotations(ctx context.Context) ([]domain.Mask, error)
	AddRasterAnnotation(ctx context.Context, mask domain.Mask) (domain.Mask, error)
	GetMaskTile(ctx context.Context, slideID, maskUID string, z, x, y int) (domain.MaskTile, error)

	// Cross-domain operations
	RemoveCaseFromStudy(ctx context.Context, studyUID string, caseUID string) error
	AddSlideToCase(ctx context.Context, caseUID string, slide domain.Slide) (domain.Slide, error)

	Close()
}

type applicationService struct {
	db                       ports.Database
	slidesService            SlidesService
	casesService             CasesService
	masksService             RasterAnnotationsService
	vectorAnnotationsService VectorAnnotationsService
	studiesService           StudiesService
	tenantsService           TenantsService
	usersService             UserService
}

// NewApplicationService creates a new ApplicationService
func NewApplicationService(
	db ports.Database,
	slidesService SlidesService,
	casesService CasesService,
	masksService RasterAnnotationsService,
	vectorAnnotationsService VectorAnnotationsService,
	studiesService StudiesService,
	tenantsService TenantsService,
	usersService UserService,
) ApplicationService {
	return &applicationService{
		db:                       db,
		slidesService:            slidesService,
		casesService:             casesService,
		masksService:             masksService,
		vectorAnnotationsService: vectorAnnotationsService,
		studiesService:           studiesService,
		tenantsService:           tenantsService,
		usersService:             usersService,
	}
}

// User operations - delegate to domain service
func (s *applicationService) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	return s.usersService.CreateUser(ctx, user)
}

func (s *applicationService) GetUserByUID(ctx context.Context, userUID string) (domain.User, error) {
	return s.usersService.GetUserByUID(ctx, userUID)
}

func (s *applicationService) GetUsers(ctx context.Context, pagination utils.PaginationParams) ([]domain.User, domain.PaginationInfo, error) {
	return s.usersService.GetUsers(ctx, pagination)
}

// Tenant operations - delegate to domain service
func (s *applicationService) GetTenants(ctx context.Context, pagination utils.PaginationParams) ([]domain.Tenant, domain.PaginationInfo, error) {
	return s.tenantsService.GetTenants(ctx, pagination)
}

func (s *applicationService) GetTenantsCount(ctx context.Context) (int, error) {
	return s.tenantsService.GetTenantsCount(ctx)
}

func (s *applicationService) GetTenantByUID(ctx context.Context, tenantUID string) (domain.Tenant, error) {
	return s.tenantsService.GetTenantByUID(ctx, tenantUID)
}

func (s *applicationService) CreateTenant(ctx context.Context, tenant domain.Tenant) (domain.Tenant, error) {
	return s.tenantsService.SaveTenant(ctx, tenant)
}

func (s *applicationService) GetCases(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Case, domain.PaginationInfo, error) {
	return s.casesService.GetCases(ctx, params)
}

func (s *applicationService) GetCasesByStudyUID(ctx context.Context, studyUID string, params utils.PaginationAndSearchParams) ([]domain.Case, domain.PaginationInfo, error) {
	return s.casesService.GetCasesByStudyUID(ctx, studyUID, params)
}

func (s *applicationService) GetCasesCount(ctx context.Context) (int, error) {
	return s.casesService.GetCasesCount(ctx)
}

func (s *applicationService) GetCaseByUID(ctx context.Context, caseUID string) (domain.Case, error) {
	return s.casesService.GetCaseByUID(ctx, caseUID)
}

func (s *applicationService) CreateCase(ctx context.Context, newCase domain.Case) (domain.Case, error) {
	return s.casesService.SaveCase(ctx, newCase)
}

func (s *applicationService) AddCaseToStudy(ctx context.Context, studyUID string, caseUID string) error {
	return s.studiesService.AddCaseToStudy(ctx, studyUID, caseUID)
}

func (s *applicationService) CreateCaseAndAssignToStudy(ctx context.Context, newCase domain.Case, studyUID string, tenantID, creatorID int) (domain.Case, error) {
	return s.casesService.CreateCaseAndAssignToStudy(ctx, newCase, studyUID, tenantID, creatorID)
}

func (s *applicationService) SoftDeleteCase(ctx context.Context, caseUID string, deletedBy int) error {
	return s.casesService.SoftDeleteCase(ctx, caseUID, deletedBy)
}

func (s *applicationService) GetDeletedCases(ctx context.Context) ([]domain.Case, error) {
	return s.casesService.GetDeletedCases(ctx)
}

func (s *applicationService) RestoreCase(ctx context.Context, caseUID string) error {
	return s.casesService.RestoreCase(ctx, caseUID)
}

// Study operations - delegate to domain service
func (s *applicationService) GetStudies(ctx context.Context, pagination utils.PaginationParams) ([]domain.Study, domain.PaginationInfo, error) {
	// Convert PaginationParams to PaginationAndSearchParams with empty search
	params := utils.PaginationAndSearchParams{
		PaginationParams: pagination,
		SearchParams:     utils.SearchParams{},
	}
	return s.studiesService.GetStudies(ctx, params)
}

func (s *applicationService) GetStudiesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	return s.studiesService.GetStudiesCount(ctx, search)
}

func (s *applicationService) GetStudyByUID(ctx context.Context, studyUID string) (domain.Study, error) {
	return s.studiesService.GetStudyByUID(ctx, studyUID)
}

func (s *applicationService) CreateStudy(ctx context.Context, study domain.Study) (domain.Study, error) {
	return s.studiesService.SaveStudy(ctx, study)
}

// Slide operations - delegate to domain service
func (s *applicationService) GetSlides(ctx context.Context, pagination utils.PaginationParams) ([]domain.Slide, domain.PaginationInfo, error) {
	return s.slidesService.GetSlides(ctx, pagination)
}

func (s *applicationService) GetSlidesCount(ctx context.Context) (int, error) {
	return s.slidesService.GetSlidesCount(ctx)
}

func (s *applicationService) GetSlideByUID(ctx context.Context, slideUID string) (domain.Slide, error) {
	return s.slidesService.GetSlideByUID(ctx, slideUID)
}

func (s *applicationService) GetSlidesByCaseUID(ctx context.Context, caseUID string) ([]domain.Slide, error) {
	return s.slidesService.GetSlidesByCaseUID(ctx, caseUID)
}

func (s *applicationService) GetSlideMetadata(ctx context.Context, slideUID string) (domain.SlideMetadata, error) {
	return s.slidesService.GetSlideMetadata(ctx, slideUID)
}

func (s *applicationService) GetSlideTile(ctx context.Context, slideUID string, z, x, y int, format string, quality int) (domain.SlideTile, error) {
	return s.slidesService.GetSlideTile(ctx, slideUID, z, x, y, format, quality)
}

func (s *applicationService) CreateSlide(ctx context.Context, newSlide domain.Slide) (domain.Slide, error) {
	return s.slidesService.SaveSlide(ctx, newSlide)
}

// Mask operations - delegate to domain service
func (s *applicationService) GetRasterAnnotations(ctx context.Context) ([]domain.Mask, error) {
	return s.masksService.GetRasterAnnotations(ctx)
}

func (s *applicationService) AddRasterAnnotation(ctx context.Context, mask domain.Mask) (domain.Mask, error) {
	return s.masksService.SaveMask(ctx, mask)
}

func (s *applicationService) GetMaskTile(ctx context.Context, slideUID, maskUID string, z, x, y int) (domain.MaskTile, error) {
	return s.masksService.GetMaskTile(ctx, slideUID, maskUID, z, x, y)
}

// Cross-domain operations - orchestrate between services
func (s *applicationService) RemoveCaseFromStudy(ctx context.Context, studyUID string, caseUID string) error {
	return s.studiesService.RemoveCaseFromStudy(ctx, studyUID, caseUID)
}

func (s *applicationService) AddSlideToCase(ctx context.Context, caseUID string, slide domain.Slide) (domain.Slide, error) {
	// First verify the case exists and get its internal ID
	case_, err := s.casesService.GetCaseByUID(ctx, caseUID)
	if err != nil {
		return domain.Slide{}, err
	}

	// Check current slide count for this case before adding a new one
	currentSlides, err := s.slidesService.GetSlidesByCaseUID(ctx, caseUID)
	if err != nil {
		return domain.Slide{}, fmt.Errorf("failed to check existing slides for case: %w", err)
	}

	const maxSlidesPerCase = 100
	if len(currentSlides) >= maxSlidesPerCase {
		return domain.Slide{}, fmt.Errorf("case already has maximum number of slides (%d)", maxSlidesPerCase)
	}

	// Set the case's internal ID on the slide
	slide.CaseID = case_.ID

	// Add the slide
	return s.slidesService.SaveSlide(ctx, slide)
}

func (s *applicationService) GetCasesByStudyUIDWithAnnotations(ctx context.Context, studyUID string, params utils.PaginationAndSearchParams) ([]domain.Case, domain.PaginationInfo, error) {
	// Get basic cases for study first
	cases, paginationInfo, err := s.casesService.GetCasesByStudyUID(ctx, studyUID, params)
	if err != nil {
		return nil, domain.PaginationInfo{}, err
	}

	// Enhance each case with slide and annotation counts
	for i := range cases {
		if err := s.enhanceCaseWithCounts(ctx, &cases[i]); err != nil {
			// Log error but continue with other cases
			log.Error("Failed to get annotation counts for case", "case_id", cases[i].CaseUID, "error", err)
		}
	}

	// Apply annotation filtering if requested
	if annotationFilter, exists := params.SearchParams.Filters["annotations"]; exists && annotationFilter != "" {
		filteredCases := make([]domain.Case, 0)
		for _, case_ := range cases {
			switch annotationFilter {
			case "with-annotations":
				if case_.AnnotationCount > 0 {
					filteredCases = append(filteredCases, case_)
				}
			case "without-annotations":
				if case_.AnnotationCount == 0 {
					filteredCases = append(filteredCases, case_)
				}
			default:
				// Include all cases for any other value
				filteredCases = append(filteredCases, case_)
			}
		}
		cases = filteredCases

		// Update pagination info to reflect filtered results
		paginationInfo.Total = len(cases)
		paginationInfo.TotalPages = (len(cases) + paginationInfo.Limit - 1) / paginationInfo.Limit
		if paginationInfo.TotalPages < 1 {
			paginationInfo.TotalPages = 1
		}
		paginationInfo.HasNext = paginationInfo.Page < paginationInfo.TotalPages
		paginationInfo.HasPrev = paginationInfo.Page > 1
	}

	return cases, paginationInfo, nil
}

func (s *applicationService) GetCaseByUIDWithAnnotations(ctx context.Context, caseUID string) (domain.Case, error) {
	// Get basic case first
	case_, err := s.casesService.GetCaseByUID(ctx, caseUID)
	if err != nil {
		return domain.Case{}, err
	}

	// Enhance with slide and annotation counts
	if err := s.enhanceCaseWithCounts(ctx, &case_); err != nil {
		log.Error("Failed to get annotation counts for case", "case_id", caseUID, "error", err)
		// Return case without annotation counts rather than failing
	}

	return case_, nil
}

// enhanceCaseWithCounts adds slide count, annotation count, and slides with annotations count to a case
func (s *applicationService) enhanceCaseWithCounts(ctx context.Context, case_ *domain.Case) error {
	// Get slides for this case
	slides, err := s.slidesService.GetSlidesByCaseUID(ctx, case_.CaseUID)
	if err != nil {
		return fmt.Errorf("failed to get slides for case %s: %w", case_.CaseUID, err)
	}

	case_.SlideCount = len(slides)

	if len(slides) == 0 {
		// No slides, so no annotations
		case_.AnnotationCount = 0
		case_.SlidesWithAnnotations = 0
		return nil
	}

	// Calculate annotation counts across all slides
	totalAnnotations := 0
	slidesWithAnnotations := 0

	for _, slide := range slides {
		slideAnnotationCount := 0

		// Get raster annotations for this slide
		rasterAnnotations, err := s.masksService.GetRasterAnnotationsForSlide(ctx, slide.SlideUID)
		if err != nil {
			// Log error but continue
			log.Error("Failed to get raster annotations for slide", "slide_id", slide.SlideUID, "error", err)
		} else {
			slideAnnotationCount += len(rasterAnnotations)
		}

		// Get vector annotations for this slide
		vectorAnnotations, err := s.vectorAnnotationsService.GetVectorAnnotationsForSlide(ctx, slide.SlideUID)
		if err != nil {
			// Log error but continue
			log.Error("Failed to get vector annotations for slide", "slide_id", slide.SlideUID, "error", err)
		} else {
			slideAnnotationCount += len(vectorAnnotations)
		}

		totalAnnotations += slideAnnotationCount

		// If this slide has any annotations, count it
		if slideAnnotationCount > 0 {
			slidesWithAnnotations++
		}
	}

	case_.AnnotationCount = totalAnnotations
	case_.SlidesWithAnnotations = slidesWithAnnotations

	return nil
}

func (s *applicationService) Close() {
	// Domain services are managed elsewhere, no need to close here
}
