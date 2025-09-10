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
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
)

// CasesService is an interface that defines the methods for the cases service.
// Interface is needed for mocking in tests.
type CasesService interface {
	GetCases(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Case, domain.PaginationInfo, error)
	GetCasesGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Case, domain.PaginationInfo, error)
	GetCasesByStudyUID(ctx context.Context, studyUID string, params utils.PaginationAndSearchParams) ([]domain.Case, domain.PaginationInfo, error)
	GetCaseNeighborsByStudyUID(ctx context.Context, studyUID string, caseUID string, searchParams utils.SearchParams) (domain.CaseNeighborsResponse, error)
	GetCasesCount(ctx context.Context) (int, error)
	GetCaseByUID(ctx context.Context, caseUID string) (domain.Case, error)
	SaveCase(ctx context.Context, newCase domain.Case) (domain.Case, error)
	CreateCaseAndAssignToStudy(ctx context.Context, newCase domain.Case, studyUID string, tenantID, creatorID int) (domain.Case, error)
	AddCaseToStudy(ctx context.Context, studyUID string, caseUID string) error
	RemoveCaseFromStudy(ctx context.Context, studyUID string, caseUID string) error
	SoftDeleteCase(ctx context.Context, caseUID string, deletedBy int) error
	GetDeletedCases(ctx context.Context) ([]domain.Case, error)
	RestoreCase(ctx context.Context, caseUID string) error
	AddSlideToCase(ctx context.Context, caseUID string, slide domain.Slide) (domain.Slide, error)
	Close()
}

type casesService struct {
	*BaseService
	db            ports.Database
	slidesService SlidesService
	// New generic pagination service
	paginatedSearchService *PaginatedSearchService[ports.Case, domain.Case]
	permissionChecker      *SharedPermissionChecker
}

// canViewCase checks if a user can view a specific case
// This now delegates to the shared permission checker for consistency
func (s *casesService) canViewCase(ctx context.Context, userID int, c ports.Case) (bool, error) {
	return s.permissionChecker.CanViewCase(ctx, userID, c)
}

// caseConversionHelpers provides conversion helpers configured for cases
var caseConversionHelpers = WithTimeFormat("2006-01-02T15:04:05Z")

// convertCaseDBToDomain converts a database Case record to a domain Case model using conversion helpers
func convertCaseDBToDomain(record ports.Case) domain.Case {
	return ConvertDBToDomainWithSoftDeletion(
		record,
		caseConversionHelpers,
		convertCaseBase,
		getCaseSoftDeletionFields,
		applyCaseSoftDeletion,
	)
}

// convertCaseBase handles the basic case conversion without soft deletion
func convertCaseBase(record ports.Case, helpers *ConversionHelpers) domain.Case {
	return domain.Case{
		ID:         record.ID,
		TenantID:   record.TenantID,
		TenantUID:  record.TenantUID,
		CaseUID:    record.CaseUID,
		CreatorID:  record.CreatorID,
		CreatorUID: record.CreatorUID,
		Name:       record.Name,
		Metadata:   helpers.ConvertMetadata(record.Metadata),
		CreatedAt:  helpers.FormatTime(record.CreatedAt),
		UpdatedAt:  helpers.FormatTime(record.UpdatedAt),
	}
}

// getCaseSoftDeletionFields extracts soft deletion fields from a case record
func getCaseSoftDeletionFields(record ports.Case) SoftDeletionFields {
	return SoftDeletionFields{
		DeletedAt: record.DeletedAt,
		DeletedBy: record.DeletedBy,
	}
}

// applyCaseSoftDeletion applies soft deletion fields to a case domain model
func applyCaseSoftDeletion(domainCase domain.Case, converted ConvertedSoftDeletionFields) domain.Case {
	return ApplySoftDeletion(
		domainCase,
		converted,
		func(c *domain.Case, deletedAt *string) { c.DeletedAt = deletedAt },
		func(c *domain.Case, deletedBy *int) { c.DeletedBy = deletedBy },
	)
}

// NewCasesService creates a new CasesService
func NewCasesService(db ports.Database, slidesService SlidesService) CasesService {
	base := NewBaseService(db)
	permissionChecker := NewSharedPermissionChecker(db)

	// Create the generic paginated search service
	paginatedSearchService := NewPaginatedSearchService(
		func(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.Case, error) {
			// Skip tenant filtering for cases to allow cross-tenant access via object grants
			// Cases can inherit permissions from parent studies across tenants
			records, err := db.LoadAllCases(ctx, search, utils.PaginationParams{Page: (offset / limit) + 1, Limit: limit})
			if err != nil {
				return nil, err
			}
			return FilterByPermission(ctx, base, records, func(c context.Context, uid int, r ports.Case) (bool, error) {
				return permissionChecker.CanViewCase(c, uid, r)
			})
		},

		func(ctx context.Context, search utils.SearchParams) (int, error) {
			// Skip tenant filtering for cases to allow cross-tenant access via object grants
			// Cases can inherit permissions from parent studies across tenants
			total, err := db.GetCasesCount(ctx, search)
			if err != nil {
				return 0, err
			}
			records, err := db.LoadAllCases(ctx, search, utils.PaginationParams{Page: 1, Limit: total})
			if err != nil {
				return 0, err
			}
			return CountByPermission(ctx, base, records, func(c context.Context, uid int, r ports.Case) (bool, error) {
				return permissionChecker.CanViewCase(c, uid, r)
			})
		},
		func(ctx context.Context, limit, offset int) ([]ports.Case, error) {
			// Skip tenant filtering for cases to allow cross-tenant access via object grants
			// Cases can inherit permissions from parent studies across tenants
			records, err := db.LoadAllCases(ctx, utils.SearchParams{}, utils.PaginationParams{Page: (offset / limit) + 1, Limit: limit})
			if err != nil {
				return nil, err
			}
			return FilterByPermission(ctx, base, records, func(c context.Context, uid int, r ports.Case) (bool, error) {
				return permissionChecker.CanViewCase(c, uid, r)
			})
		},
		func(ctx context.Context) (int, error) {
			// Skip tenant filtering for cases to allow cross-tenant access via object grants
			// Cases can inherit permissions from parent studies across tenants
			records, err := db.LoadAllCases(ctx, utils.SearchParams{}, utils.PaginationParams{Page: 1, Limit: 0})
			if err != nil {
				return 0, err
			}
			return CountByPermission(ctx, base, records, func(c context.Context, uid int, r ports.Case) (bool, error) {
				return permissionChecker.CanViewCase(c, uid, r)
			})
		},
		convertCaseDBToDomain,
	)

	return &casesService{
		BaseService:            NewBaseService(db),
		db:                     db,
		slidesService:          slidesService,
		paginatedSearchService: paginatedSearchService,
		permissionChecker:      permissionChecker,
	}
}

// GetCasesGeneric retrieves cases using the new generic pattern
func (s *casesService) GetCasesGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Case, domain.PaginationInfo, error) {
	cases, paginationInfo, err := s.paginatedSearchService.GetWithPaginationAndSearch(ctx, params)
	if err != nil {
		return nil, domain.PaginationInfo{}, err
	}

	// Enhance each case with slide count
	for i := range cases {
		slides, err := s.slidesService.GetSlidesByCaseUID(ctx, cases[i].CaseUID)
		if err != nil {
			log.Error("Failed to get slides for case", "error", err, "caseUID", cases[i].CaseUID)
			// Continue without slide count rather than failing completely
		} else {
			cases[i].SlideCount = len(slides)
		}
	}

	return cases, paginationInfo, nil
}

// GetCases retrieves cases from the database with pagination and search support
func (s *casesService) GetCases(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Case, domain.PaginationInfo, error) {
	cases, paginationInfo, err := s.paginatedSearchService.GetWithPaginationAndSearch(ctx, params)
	if err != nil {
		return nil, domain.PaginationInfo{}, err
	}

	// Enhance each case with slide count
	for i := range cases {
		slides, err := s.slidesService.GetSlidesByCaseUID(ctx, cases[i].CaseUID)
		if err != nil {
			log.Error("Failed to get slides for case", "error", err, "caseUID", cases[i].CaseUID)
			// Continue without slide count rather than failing completely
		} else {
			cases[i].SlideCount = len(slides)
		}
	}

	return cases, paginationInfo, nil
}

// GetCaseByUID retrieves a specific case by its short_uid
func (s *casesService) GetCaseByUID(ctx context.Context, caseUID string) (domain.Case, error) {
	dbRecord, err := s.db.GetCaseByUID(ctx, caseUID)
	if err != nil {
		log.Error("Failed to get case by ID", "error", err)
		return domain.Case{}, errors.WithDetails(errors.ErrSlideNotFound, "case with ID '%s'", caseUID)
	}

	// Convert to domain model
	domainCase := convertCaseDBToDomain(dbRecord)

	// Enhance with slide count
	slides, err := s.slidesService.GetSlidesByCaseUID(ctx, caseUID)
	if err != nil {
		log.Error("Failed to get slides for case", "error", err, "caseUID", caseUID)
		// Continue without slide count rather than failing completely
	} else {
		domainCase.SlideCount = len(slides)
	}

	return domainCase, nil
}

// SaveCase saves a case to the database
func (s *casesService) SaveCase(ctx context.Context, newCase domain.Case) (domain.Case, error) {
	existingCase := false
	if newCase.CaseUID != "" {
		_, err := s.db.GetCaseByUID(ctx, newCase.CaseUID)
		if err == nil {
			existingCase = true
		}
	}

	log.Info("Saving case", "case", newCase)

	// Error if case name is empty
	if newCase.Name == "" {
		return domain.Case{}, errors.WithDetails(errors.ErrInvalidInput, "case name cannot be empty")
	}

	// TODO: How to update case?
	if existingCase {
		return domain.Case{}, errors.WithDetails(errors.ErrAlreadyExists, "case with ID '%s'", newCase.CaseUID)
	}

	shortUID, err := s.GenerateShortUID()
	if err != nil {
		return domain.Case{}, err
	}

	newCase.CaseUID = shortUID

	// Use the WithAuthContext helper for cleaner code
	err = s.WithAuthContext(ctx, func(authCtx *AuthContext) error {
		log.Info("Saving case", "case", newCase)

		dbCase := ports.NewCase{
			TenantID:  authCtx.TenantID,
			CaseUID:   shortUID,
			CreatorID: authCtx.CreatorID,
			Name:      newCase.Name,
			Metadata:  []byte(newCase.Metadata),
		}

		return s.db.CreateCase(ctx, dbCase)
	})
	if err != nil {
		log.Error("Failed to save case", "error", err)
		return domain.Case{}, errors.WithDetails(errors.ErrInternal, "failed to save case: %v", err)
	}

	return newCase, nil
}

// Close releases all resources held by the service
func (s *casesService) Close() {
	// no-op
}

// GetCasesByStudyUID retrieves all cases that belong to a specific study
func (s *casesService) GetCasesByStudyUID(ctx context.Context, studyUID string, params utils.PaginationAndSearchParams) ([]domain.Case, domain.PaginationInfo, error) {
	// Get cases with search and pagination - database layer handles all filtering and pagination now
	dbRecords, err := s.db.GetCasesByStudyUID(ctx, studyUID, params)
	if err != nil {
		log.Error("Failed to load cases by study UID", "error", err, "studyUID", studyUID)
		return nil, domain.PaginationInfo{}, fmt.Errorf("failed to load cases by study UID: %w", err)
	}

	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return nil, domain.PaginationInfo{}, err
	}

	// Apply permission filtering using the shared permission checker for consistency
	if !authCtx.IsSuperAdmin {
		filtered := make([]ports.Case, 0, len(dbRecords))
		for _, r := range dbRecords {
			// Use the shared permission checker to ensure consistency with object grants
			allowed, err := s.permissionChecker.CanViewCase(ctx, authCtx.CreatorID, r)
			if err != nil {
				return nil, domain.PaginationInfo{}, err
			}
			if allowed {
				filtered = append(filtered, r)
			}
		}
		dbRecords = filtered
	}

	// Get total count with search filters
	totalCount, err := s.db.GetCasesByStudyUIDCount(ctx, studyUID, params.SearchParams)
	if err != nil {
		log.Error("Failed to get cases count by study UID", "error", err, "studyUID", studyUID)
		return nil, domain.PaginationInfo{}, fmt.Errorf("failed to get cases count by study UID: %w", err)
	}

	// Adjust total count for permission filtering if not super admin
	if !authCtx.IsSuperAdmin {
		totalCount = len(dbRecords)
	}

	// Convert database records to domain records
	cases := make([]domain.Case, 0, len(dbRecords))
	for _, record := range dbRecords {
		domainCase := convertCaseDBToDomain(record)

		// Enhance with slide count
		slides, err := s.slidesService.GetSlidesByCaseUID(ctx, record.CaseUID)
		if err != nil {
			log.Error("Failed to get slides for case", "error", err, "caseUID", record.CaseUID)
			// Continue without slide count rather than failing completely
		} else {
			domainCase.SlideCount = len(slides)
		}

		cases = append(cases, domainCase)
	}

	// Calculate pagination info using the utility function
	paginationInfo := utils.CreatePaginationInfo(params.PaginationParams, totalCount)

	return cases, paginationInfo, nil
}

// AddCaseToStudy adds an existing case to a study via the study_cases table
func (s *casesService) AddCaseToStudy(ctx context.Context, studyUID string, caseUID string) error {
	err := s.db.AddCaseToStudy(ctx, studyUID, caseUID)
	if err != nil {
		log.Error("Failed to add case to study", "error", err, "studyUID", studyUID, "caseUID", caseUID)
		return fmt.Errorf("failed to add case to study: %w", err)
	}
	return nil
}

// CreateCaseAndAssignToStudy creates a new case and assigns it to a study
func (s *casesService) CreateCaseAndAssignToStudy(ctx context.Context, newCase domain.Case, studyUID string, tenantID, creatorID int) (domain.Case, error) {
	shortUID, err := utils.GenerateFixedShortUID()
	if err != nil {
		return domain.Case{}, err
	}

	log.Info("Saving case", "case", newCase)

	dbCase := ports.NewCase{
		TenantID:  tenantID,
		CaseUID:   shortUID,
		CreatorID: creatorID,
		Name:      newCase.Name,
		Metadata:  []byte(newCase.Metadata),
	}

	newCase.CaseUID = shortUID

	log.Info("Saving case", "case", newCase)

	// Create the case
	if err := s.db.CreateCase(ctx, dbCase); err != nil {
		log.Error("Failed to save case", "error", err)
		return domain.Case{}, err
	}

	// Add the case to the study
	if err := s.db.AddCaseToStudy(ctx, studyUID, shortUID); err != nil {
		log.Error("Failed to add case to study", "error", err)
		return domain.Case{}, err
	}

	return domain.Case{
		ID:        0, // Will be set by database
		TenantID:  tenantID,
		CaseUID:   shortUID,
		CreatorID: creatorID,
		Name:      newCase.Name,
		Metadata:  newCase.Metadata,
		CreatedAt: "", // Will be set by database
	}, nil
}

// RemoveCaseFromStudy removes a case from a study
func (s *casesService) RemoveCaseFromStudy(ctx context.Context, studyUID string, caseUID string) error {
	err := s.db.RemoveCaseFromStudy(ctx, studyUID, caseUID)
	if err != nil {
		log.Error("Failed to remove case from study", "error", err, "studyUID", studyUID, "caseUID", caseUID)
		return fmt.Errorf("failed to remove case from study: %w", err)
	}
	return nil
}

// SoftDeleteCase marks a case as deleted
func (s *casesService) SoftDeleteCase(ctx context.Context, caseUID string, deletedBy int) error {
	err := s.db.SoftDeleteCase(ctx, caseUID, deletedBy)
	if err != nil {
		log.Error("Failed to soft delete case", "error", err, "caseUID", caseUID)
		return fmt.Errorf("failed to soft delete case: %w", err)
	}
	return nil
}

// GetDeletedCases retrieves all deleted cases
func (s *casesService) GetDeletedCases(ctx context.Context) ([]domain.Case, error) {
	dbRecords, err := s.db.GetDeletedCases(ctx)
	if err != nil {
		log.Error("Failed to load deleted cases", "error", err)
		return nil, fmt.Errorf("failed to load deleted cases: %w", err)
	}

	cases := make([]domain.Case, 0, len(dbRecords))
	for _, record := range dbRecords {
		domainCase := convertCaseDBToDomain(record)

		// Enhance with slide count
		slides, err := s.slidesService.GetSlidesByCaseUID(ctx, record.CaseUID)
		if err != nil {
			log.Error("Failed to get slides for deleted case", "error", err, "caseUID", record.CaseUID)
			// Continue without slide count rather than failing completely
		} else {
			domainCase.SlideCount = len(slides)
		}

		cases = append(cases, domainCase)
	}

	return cases, nil
}

// RestoreCase restores a deleted case
func (s *casesService) RestoreCase(ctx context.Context, caseUID string) error {
	err := s.db.RestoreCase(ctx, caseUID)
	if err != nil {
		log.Error("Failed to restore case", "error", err, "caseUID", caseUID)
		return fmt.Errorf("failed to restore case: %w", err)
	}
	return nil
}

// GetCasesCount retrieves the total count of cases from the database
func (s *casesService) GetCasesCount(ctx context.Context) (int, error) {
	return s.db.GetCasesCount(ctx, utils.SearchParams{})
}

// GetCaseNeighborsByStudyUID retrieves case neighbors for navigation
func (s *casesService) GetCaseNeighborsByStudyUID(ctx context.Context, studyUID string, caseUID string, searchParams utils.SearchParams) (domain.CaseNeighborsResponse, error) {
	// Get all cases for the study with search filters (no pagination - set limit to 0 to get all)
	params := utils.PaginationAndSearchParams{
		PaginationParams: utils.PaginationParams{
			Page:  1,
			Limit: 0, // 0 means no pagination limit
		},
		SearchParams: searchParams,
	}

	dbRecords, err := s.db.GetCasesByStudyUID(ctx, studyUID, params)
	if err != nil {
		log.Error("Failed to load cases by study UID for neighbors", "error", err, "studyUID", studyUID)
		return domain.CaseNeighborsResponse{}, fmt.Errorf("failed to load cases by study UID: %w", err)
	}

	// Find the current case position in the filtered list
	currentIndex := -1
	for i, record := range dbRecords {
		if record.CaseUID == caseUID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		log.Warn("Case not found in study", "caseUID", caseUID, "studyUID", studyUID)
		return domain.CaseNeighborsResponse{}, fmt.Errorf("case with ID '%s' not found in study '%s'", caseUID, studyUID)
	}

	// Build the neighbors response
	response := domain.CaseNeighborsResponse{
		Number: currentIndex + 1, // 1-indexed position
		Count:  len(dbRecords),
	}

	// Set previous case if exists
	if currentIndex > 0 {
		prevRecord := dbRecords[currentIndex-1]
		response.Prev = &domain.CaseNeighbor{
			CaseUID: prevRecord.CaseUID,
			Name:    prevRecord.Name,
		}
	}

	// Set next case if exists
	if currentIndex < len(dbRecords)-1 {
		nextRecord := dbRecords[currentIndex+1]
		response.Next = &domain.CaseNeighbor{
			CaseUID: nextRecord.CaseUID,
			Name:    nextRecord.Name,
		}
	}

	return response, nil
}

// AddSlideToCase adds a slide to a case
func (s *casesService) AddSlideToCase(ctx context.Context, caseUID string, slide domain.Slide) (domain.Slide, error) {
	// First verify the case exists and get its internal ID
	case_, err := s.GetCaseByUID(ctx, caseUID)
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
