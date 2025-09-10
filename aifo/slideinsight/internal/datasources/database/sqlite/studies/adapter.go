// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package studies

import (
	"context"
	"database/sql"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Adapter provides a unified interface for all study operations
type Adapter struct {
	crud       *CrudService
	search     *SearchService
	relations  *RelationsService
	softDelete *SoftDeleteService
}

// NewAdapter creates a new studies adapter
func NewAdapter(db *sql.DB) *Adapter {
	return &Adapter{
		crud:       NewCrudService(db),
		search:     NewSearchService(db),
		relations:  NewRelationsService(db),
		softDelete: NewSoftDeleteService(db),
	}
}

// Basic CRUD operations

// CreateStudy adds a new study to the database
func (a *Adapter) CreateStudy(ctx context.Context, newStudy ports.NewStudy) error {
	return a.crud.CreateStudy(ctx, newStudy)
}

// GetStudyByUID retrieves a specific study by its ID
func (a *Adapter) GetStudyByUID(ctx context.Context, studyUID string) (ports.Study, error) {
	return a.crud.GetStudyByUID(ctx, studyUID)
}

// GetStudyByInternalID retrieves a specific study by its internal database ID
func (a *Adapter) GetStudyByInternalID(ctx context.Context, studyID int) (ports.Study, error) {
	return a.crud.GetStudyByInternalID(ctx, studyID)
}

// UpdateStudy updates study information for a study with the specified ID
func (a *Adapter) UpdateStudy(ctx context.Context, studyUID string, updates ports.StudyUpdates) error {
	return a.crud.UpdateStudy(ctx, studyUID, updates)
}

// Search and listing operations

// LoadAllStudies retrieves studies from the database with search/filter and pagination support
func (a *Adapter) LoadAllStudies(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Study, error) {
	return a.search.LoadAllStudies(ctx, search, pagination)
}

// GetStudiesCount returns the total count of studies matching search criteria
func (a *Adapter) GetStudiesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	return a.search.GetStudiesCount(ctx, search)
}

// LoadStudiesByIDs retrieves studies filtered by a list of IDs with search/filter and pagination support
func (a *Adapter) LoadStudiesByIDs(ctx context.Context, studyIDs []int, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Study, error) {
	return a.search.LoadStudiesByIDs(ctx, studyIDs, search, pagination)
}

// CountStudiesByIDs returns the total count of studies in the ID list matching search criteria
func (a *Adapter) CountStudiesByIDs(ctx context.Context, studyIDs []int, search utils.SearchParams) (int, error) {
	return a.search.CountStudiesByIDs(ctx, studyIDs, search)
}

// Relations operations

// GetStudyIDByShortUID retrieves the internal study ID by its short UID
func (a *Adapter) GetStudyIDByShortUID(ctx context.Context, studyUID string) (int, error) {
	return a.relations.GetStudyIDByShortUID(ctx, studyUID)
}

// GetStudyCaseCounts retrieves case counts for all studies
func (a *Adapter) GetStudyCaseCounts(ctx context.Context) (map[string]int, error) {
	return a.relations.GetStudyCaseCounts(ctx)
}

// GetStudySlideCounts retrieves slide counts for all studies
func (a *Adapter) GetStudySlideCounts(ctx context.Context) (map[string]int, error) {
	return a.relations.GetStudySlideCounts(ctx)
}

// GetStudyCaseAndSlideCounts retrieves both case and slide counts for all studies efficiently
func (a *Adapter) GetStudyCaseAndSlideCounts(ctx context.Context) (map[string]int, map[string]int, error) {
	return a.relations.GetStudyCaseAndSlideCounts(ctx)
}

// Soft delete operations

// SoftDeleteStudy marks a study as deleted without removing it from the database
func (a *Adapter) SoftDeleteStudy(ctx context.Context, studyUID string, deletedBy int) error {
	return a.softDelete.SoftDeleteStudy(ctx, studyUID, deletedBy)
}

// GetDeletedStudies retrieves all soft-deleted studies
func (a *Adapter) GetDeletedStudies(ctx context.Context) ([]ports.Study, error) {
	return a.softDelete.GetDeletedStudies(ctx)
}

// RestoreStudy restores a soft-deleted study
func (a *Adapter) RestoreStudy(ctx context.Context, studyUID string) error {
	return a.softDelete.RestoreStudy(ctx, studyUID)
}
