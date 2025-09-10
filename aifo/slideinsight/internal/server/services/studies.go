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

// StudiesService is an interface that defines the methods for the studies service.
// Interface is needed for mocking in tests.
type StudiesService interface {
	GetStudies(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Study, domain.PaginationInfo, error)
	GetStudiesGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Study, domain.PaginationInfo, error)
	GetStudiesCount(ctx context.Context, search utils.SearchParams) (int, error)
	GetStudyByUID(ctx context.Context, studyUID string) (domain.Study, error)
	GetStudyMetadata(ctx context.Context, studyUID string) (domain.StudyMetadata, error)
	SaveStudy(ctx context.Context, newStudy domain.Study) (domain.Study, error)
	UpdateStudy(ctx context.Context, studyUID string, updates domain.StudyUpdates) error
	AddCaseToStudy(ctx context.Context, studyUID string, caseUID string) error
	RemoveCaseFromStudy(ctx context.Context, studyUID string, caseUID string) error
	SoftDeleteStudy(ctx context.Context, studyUID string) error
	RestoreStudy(ctx context.Context, studyUID string) error
	Close()
}

type studiesService struct {
	*BaseService
	db                     ports.Database
	paginatedSearchService *PaginatedSearchService[ports.Study, domain.Study]
	permissionChecker      *SharedPermissionChecker
}

// canViewStudy checks if a user can view a specific study
// This now delegates to the shared permission checker for consistency
func (s *studiesService) canViewStudy(ctx context.Context, userID int, study ports.Study) (bool, error) {
	return s.permissionChecker.CanViewStudy(ctx, userID, study)
}

// studyConversionHelpers provides conversion helpers configured for studies (using RFC3339)
var studyConversionHelpers = DefaultConversionHelpers()

// convertStudyDBToDomain converts a database Study record to a domain Study model using conversion helpers
func convertStudyDBToDomain(record ports.Study) domain.Study {
	return ConvertDBToDomainWithSoftDeletion(
		record,
		studyConversionHelpers,
		convertStudyBase,
		getStudySoftDeletionFields,
		applyStudySoftDeletion,
	)
}

// convertStudyDBToDomainWithCounts converts a database Study record to a domain Study model with case and slide counts
func convertStudyDBToDomainWithCounts(record ports.Study, caseCounts, slideCounts map[string]int) domain.Study {
	study := convertStudyDBToDomain(record)
	study.CaseCount = caseCounts[record.StudyUID]
	study.SlideCount = slideCounts[record.StudyUID]
	return study
}

// convertStudyBase handles the basic study conversion without soft deletion
func convertStudyBase(record ports.Study, helpers *ConversionHelpers) domain.Study {
	return domain.Study{
		ID:          record.ID,
		TenantID:    record.TenantID,
		TenantUID:   record.TenantUID,
		StudyUID:    record.StudyUID,
		CreatorID:   record.CreatorID,
		CreatorUID:  record.CreatorUID,
		Name:        record.Name,
		Description: record.Description,
		Metadata:    helpers.ConvertMetadata(record.Metadata),
		IsPublished: record.IsPublished,
		CaseCount:   0, // Will be filled in by convertStudyDBToDomainWithCounts
		SlideCount:  0, // Will be filled in by convertStudyDBToDomainWithCounts
		CreatedAt:   helpers.FormatTime(record.CreatedAt),
	}
}

// getStudySoftDeletionFields extracts soft deletion fields from a study record
func getStudySoftDeletionFields(record ports.Study) SoftDeletionFields {
	return SoftDeletionFields{
		DeletedAt: record.DeletedAt,
		DeletedBy: record.DeletedBy,
	}
}

// applyStudySoftDeletion applies soft deletion fields to a study domain model
func applyStudySoftDeletion(domainStudy domain.Study, converted ConvertedSoftDeletionFields) domain.Study {
	return ApplySoftDeletion(
		domainStudy,
		converted,
		func(s *domain.Study, deletedAt *string) { s.DeletedAt = deletedAt },
		func(s *domain.Study, deletedBy *int) { s.DeletedBy = deletedBy },
	)
}

// NewStudiesService creates a new StudiesService
func NewStudiesService(db ports.Database) StudiesService {
	permissionChecker := NewSharedPermissionChecker(db)

	// Create the generic paginated search service WITHOUT permission filtering
	// Permission filtering should ONLY be controlled by FilterAccessibleStudies
	paginatedSearchService := NewPaginatedSearchService(
		func(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.Study, error) {
			// Return all studies without any permission filtering
			// Permission filtering is controlled exclusively by FilterAccessibleStudies
			return db.LoadAllStudies(ctx, search, utils.PaginationParams{Page: (offset / limit) + 1, Limit: limit})
		},

		func(ctx context.Context, search utils.SearchParams) (int, error) {
			// Return total count without any permission filtering
			// Permission filtering is controlled exclusively by FilterAccessibleStudies
			return db.GetStudiesCount(ctx, search)
		},
		func(ctx context.Context, limit, offset int) ([]ports.Study, error) {
			// Return all studies without any permission filtering
			// Permission filtering is controlled exclusively by FilterAccessibleStudies
			return db.LoadAllStudies(ctx, utils.SearchParams{}, utils.PaginationParams{Page: (offset / limit) + 1, Limit: limit})
		},
		func(ctx context.Context) (int, error) {
			// Return total count without any permission filtering
			// Permission filtering is controlled exclusively by FilterAccessibleStudies
			return db.GetStudiesCount(ctx, utils.SearchParams{})
		},
		convertStudyDBToDomain,
	)

	return &studiesService{
		BaseService:            NewBaseService(db),
		db:                     db,
		paginatedSearchService: paginatedSearchService,
		permissionChecker:      permissionChecker,
	}
}

// GetStudiesGeneric retrieves studies using simplified filtering logic
func (s *studiesService) GetStudiesGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Study, domain.PaginationInfo, error) {
	// If SQL-based permission filtering is requested, use it exclusively
	if params.SearchParams.ShouldFilterAccessibleStudies() {
		return s.getStudiesWithSQLFiltering(ctx, params)
	}

	// Otherwise, return all studies without any permission filtering
	studies, paginationInfo, err := s.paginatedSearchService.GetWithPaginationAndSearch(ctx, params)
	if err != nil {
		return nil, domain.PaginationInfo{}, err
	}

	// Get case and slide counts efficiently
	caseCounts, slideCounts, err := s.db.GetStudyCaseAndSlideCounts(ctx)
	if err != nil {
		log.Error("Failed to get study case and slide counts", "error", err)
		// Continue without counts rather than failing completely
		caseCounts = make(map[string]int)
		slideCounts = make(map[string]int)
	}

	// Add counts to all studies
	for i := range studies {
		studies[i].CaseCount = caseCounts[studies[i].StudyUID]
		studies[i].SlideCount = slideCounts[studies[i].StudyUID]
	}

	return studies, paginationInfo, nil
}

// getStudiesWithSQLFiltering uses SQL-based permission filtering for efficient access control
func (s *studiesService) getStudiesWithSQLFiltering(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Study, domain.PaginationInfo, error) {
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return nil, domain.PaginationInfo{}, err
	}

	var accessibleStudyIDs []int

	// For super admin, get all study IDs; for regular users, get accessible study IDs
	if authCtx.IsSuperAdmin {
		// Get all study IDs without permission filtering
		allStudies, err := s.db.LoadAllStudies(ctx, utils.SearchParams{}, utils.PaginationParams{Page: 1, Limit: 0})
		if err != nil {
			log.Error("Failed to get all studies for super admin", "error", err)
			return nil, domain.PaginationInfo{}, fmt.Errorf("failed to get all studies: %w", err)
		}
		accessibleStudyIDs = make([]int, len(allStudies))
		for i, study := range allStudies {
			accessibleStudyIDs[i] = study.ID
		}
	} else {
		// Get accessible study IDs using SQL-based filtering
		accessibleStudyIDs, err = s.db.GetAccessibleStudyIDs(ctx, authCtx.CreatorID, "studies.view")
		if err != nil {
			log.Error("Failed to get accessible study IDs", "error", err)
			return nil, domain.PaginationInfo{}, fmt.Errorf("failed to get accessible study IDs: %w", err)
		}
	}

	// If no accessible studies, return empty result
	if len(accessibleStudyIDs) == 0 {
		return []domain.Study{}, utils.CreatePaginationInfo(params.PaginationParams, 0), nil
	}

	// Get studies filtered by accessible IDs with search and pagination
	dbRecords, err := s.db.LoadStudiesByIDs(ctx, accessibleStudyIDs, params.SearchParams, params.PaginationParams)
	if err != nil {
		log.Error("Failed to load studies by IDs", "error", err)
		return nil, domain.PaginationInfo{}, fmt.Errorf("failed to load studies by IDs: %w", err)
	}

	// Get total count for pagination
	totalCount, err := s.db.CountStudiesByIDs(ctx, accessibleStudyIDs, params.SearchParams)
	if err != nil {
		log.Error("Failed to count studies by IDs", "error", err)
		return nil, domain.PaginationInfo{}, fmt.Errorf("failed to count studies by IDs: %w", err)
	}

	// Get case and slide counts efficiently
	caseCounts, slideCounts, err := s.db.GetStudyCaseAndSlideCounts(ctx)
	if err != nil {
		log.Error("Failed to get study case and slide counts", "error", err)
		// Continue without counts rather than failing completely
		caseCounts = make(map[string]int)
		slideCounts = make(map[string]int)
	}

	// Convert to domain objects with counts
	studies := make([]domain.Study, 0, len(dbRecords))
	for _, record := range dbRecords {
		study := convertStudyDBToDomainWithCounts(record, caseCounts, slideCounts)
		studies = append(studies, study)
	}

	paginationInfo := utils.CreatePaginationInfo(params.PaginationParams, totalCount)

	return studies, paginationInfo, nil
}

// GetStudies retrieves all studies from the database with pagination support
func (s *studiesService) GetStudies(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Study, domain.PaginationInfo, error) {
	studies, paginationInfo, err := s.paginatedSearchService.GetWithPaginationAndSearch(ctx, params)
	if err != nil {
		return nil, domain.PaginationInfo{}, err
	}

	// Get case and slide counts efficiently
	caseCounts, slideCounts, err := s.db.GetStudyCaseAndSlideCounts(ctx)
	if err != nil {
		log.Error("Failed to get study case and slide counts", "error", err)
		// Continue without counts rather than failing completely
		caseCounts = make(map[string]int)
		slideCounts = make(map[string]int)
	}

	// Add counts to all studies
	for i := range studies {
		studies[i].CaseCount = caseCounts[studies[i].StudyUID]
		studies[i].SlideCount = slideCounts[studies[i].StudyUID]
	}

	return studies, paginationInfo, nil
}

// GetStudiesCount retrieves the total count of studies from the database
func (s *studiesService) GetStudiesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	count, err := s.db.GetStudiesCount(ctx, search)
	if err != nil {
		log.Error("Failed to get studies count", "error", err)
		return 0, fmt.Errorf("failed to get studies count: %w", err)
	}
	return count, nil
}

func (s *studiesService) GetStudyByUID(ctx context.Context, studyUID string) (domain.Study, error) {
	dbRecord, err := s.db.GetStudyByUID(ctx, studyUID)
	if err != nil {
		log.Error("Failed to get study by UID", "error", err)
		return domain.Study{}, errors.WithDetails(errors.ErrSlideNotFound, "study with UID '%s'", studyUID)
	}

	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		log.Error("Failed to get auth context", "error", err)
		return domain.Study{}, err
	}

	// Use the shared permission checking method for consistency
	allowed := authCtx.IsSuperAdmin
	if !allowed {
		log.Info("Not super admin, checking if user has access to study", "study", dbRecord, "user", authCtx.CreatorID)
		allowed, err = s.canViewStudy(ctx, authCtx.CreatorID, dbRecord)
		log.Info("canViewStudy result", "allowed", allowed, "error", err)
		if err != nil {
			return domain.Study{}, err
		}
	} else {
		log.Info("Super admin, allowing access to study", "study", dbRecord)
	}

	if !allowed {
		return domain.Study{}, errors.ErrInsufficientPermissions
	}

	// Get case and slide counts for this specific study
	caseCounts, slideCounts, err := s.db.GetStudyCaseAndSlideCounts(ctx)
	if err != nil {
		log.Error("Failed to get study case and slide counts", "error", err)
		// Continue without counts rather than failing completely
		caseCounts = make(map[string]int)
		slideCounts = make(map[string]int)
	}

	return convertStudyDBToDomainWithCounts(dbRecord, caseCounts, slideCounts), nil
}

// GetStudyMetadata retrieves the metadata for a specific study
func (s *studiesService) GetStudyMetadata(ctx context.Context, studyUID string) (domain.StudyMetadata, error) {
	// Verify study exists
	dbRecord, err := s.db.GetStudyByUID(ctx, studyUID)
	if err != nil {
		log.Error("Failed to get study for metadata", "error", err)
		return domain.StudyMetadata{}, errors.WithDetails(errors.ErrSlideNotFound, "study with ID '%s'", studyUID)
	}

	// Get case counts for this specific study
	caseCounts, err := s.db.GetStudyCaseCounts(ctx)
	if err != nil {
		log.Error("Failed to load study case counts for metadata", "error", err)
		return domain.StudyMetadata{}, fmt.Errorf("failed to load study case counts: %w", err)
	}

	caseCount := caseCounts[dbRecord.StudyUID] // Will be 0 if not found

	metadata := domain.StudyMetadata{
		StudyUID:  dbRecord.StudyUID,
		CaseCount: caseCount,
	}

	return metadata, nil
}

// SaveStudy saves a study to the database
func (s *studiesService) SaveStudy(ctx context.Context, study domain.Study) (domain.Study, error) {
	// Check if study already exists (for cache invalidation)
	existingStudy := false
	if study.StudyUID != "" {
		_, err := s.db.GetStudyByUID(ctx, study.StudyUID)
		if err == nil {
			existingStudy = true
		}
	}

	log.Info("Saving study", "study", study)

	// Error if study name is empty
	if study.Name == "" {
		return domain.Study{}, errors.WithDetails(errors.ErrInvalidInput, "study name cannot be empty")
	}

	// TODO: How to update study?
	if existingStudy {
		return domain.Study{}, errors.WithDetails(errors.ErrAlreadyExists, "study with ID '%s'", study.StudyUID)
	}

	randomID, err := utils.GenerateFixedShortUID()
	if err != nil {
		return domain.Study{}, errors.WithDetails(errors.ErrInternal, "failed to generate study ID: %v", err)
	}
	study.StudyUID = randomID

	// Use the base service to get authentication context
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return domain.Study{}, err
	}

	// Log the study metadata
	log.Info("Saving study", "study", study)

	dbStudy := ports.NewStudy{
		StudyUID:    study.StudyUID,
		TenantID:    authCtx.TenantID,
		CreatorID:   authCtx.CreatorID,
		Name:        study.Name,
		Description: study.Description,
	}

	err = s.db.CreateStudy(ctx, dbStudy)
	if err != nil {
		log.Error("Failed to save study", "error", err)
		return domain.Study{}, errors.WithDetails(errors.ErrInternal, "failed to save study: %v", err)
	}

	// Return the newly created study with proper domain fields filled
	study.TenantID = authCtx.TenantID
	study.CreatorID = authCtx.CreatorID
	study.CaseCount = 0  // New study has no cases
	study.SlideCount = 0 // New study has no slides

	return study, nil
}

// UpdateStudy updates study information for a study with the specified ID
func (s *studiesService) UpdateStudy(ctx context.Context, studyUID string, updates domain.StudyUpdates) error {
	if studyUID == "" {
		return fmt.Errorf("study ID cannot be empty")
	}

	// Verify study exists
	_, err := s.db.GetStudyByUID(ctx, studyUID)
	if err != nil {
		return fmt.Errorf("study with ID '%s' does not exist", studyUID)
	}

	// Convert domain updates to ports updates
	var metadataBytes *[]byte
	if updates.Metadata != nil {
		bytes := []byte(*updates.Metadata)
		metadataBytes = &bytes
	}

	portsUpdates := ports.StudyUpdates{
		Name:        updates.Name,
		Description: updates.Description,
		Metadata:    metadataBytes,
		IsPublished: updates.IsPublished,
	}

	// Update the study in the database
	err = s.db.UpdateStudy(ctx, studyUID, portsUpdates)
	if err != nil {
		return fmt.Errorf("failed to update study: %w", err)
	}

	return nil
}

// SoftDeleteStudy marks a study as deleted without removing it from the database
func (s *studiesService) SoftDeleteStudy(ctx context.Context, studyUID string) error {
	// Use the base service to get authentication context
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return err
	}

	err = s.db.SoftDeleteStudy(ctx, studyUID, authCtx.CreatorID)
	if err != nil {
		log.Error("Failed to soft delete study", "error", err)
		return errors.WithDetails(errors.ErrInternal, "failed to soft delete study: %v", err)
	}
	return nil
}

// RestoreStudy restores a soft-deleted study
func (s *studiesService) RestoreStudy(ctx context.Context, studyUID string) error {
	err := s.db.RestoreStudy(ctx, studyUID)
	if err != nil {
		log.Error("Failed to restore study", "error", err)
		return errors.WithDetails(errors.ErrInternal, "failed to restore study: %v", err)
	}
	return nil
}

func (s *studiesService) AddCaseToStudy(ctx context.Context, studyUID string, caseUID string) error {
	err := s.db.AddCaseToStudy(ctx, studyUID, caseUID)
	if err != nil {
		log.Error("Failed to add case to study", "error", err)
		return errors.WithDetails(errors.ErrInternal, "failed to add case to study: %v", err)
	}
	return nil
}

func (s *studiesService) RemoveCaseFromStudy(ctx context.Context, studyUID string, caseUID string) error {
	err := s.db.RemoveCaseFromStudy(ctx, studyUID, caseUID)
	if err != nil {
		log.Error("Failed to remove case from study", "error", err)
		return errors.WithDetails(errors.ErrInternal, "failed to remove case from study: %v", err)
	}
	return nil
}

// Close releases all resources held by the service
func (s *studiesService) Close() {
	// no-op
}
