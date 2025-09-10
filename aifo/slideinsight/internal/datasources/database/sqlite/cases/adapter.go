// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package cases

import (
	"context"
	"database/sql"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Adapter provides a unified interface for all case operations
type Adapter struct {
	crud           *CaseService
	search         *SearchService
	studyRelations *StudyRelationService
}

// NewAdapter creates a new cases adapter
func NewAdapter(db *sql.DB) *Adapter {
	return &Adapter{
		crud:           NewCaseService(db),
		search:         NewSearchService(db),
		studyRelations: NewStudyRelationService(db),
	}
}

// CreateCase adds a new case to the database
func (a *Adapter) CreateCase(ctx context.Context, newCase ports.NewCase) error {
	return a.crud.CreateCase(ctx, newCase)
}

// LoadAllCases retrieves cases from the database with search/filter and pagination support
func (a *Adapter) LoadAllCases(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Case, error) {
	return a.search.LoadAllCases(ctx, search, pagination)
}

// GetCaseByUID retrieves a specific case by its ID
func (a *Adapter) GetCaseByUID(ctx context.Context, caseUID string) (ports.Case, error) {
	return a.crud.GetCaseByUID(ctx, caseUID)
}

// AddCaseToStudy adds an existing case to a study via the study_cases table
func (a *Adapter) AddCaseToStudy(ctx context.Context, studyUID string, caseUID string) error {
	return a.studyRelations.AddCaseToStudy(ctx, studyUID, caseUID)
}

// RemoveCaseFromStudy removes a case from a study via the study_cases table
func (a *Adapter) RemoveCaseFromStudy(ctx context.Context, studyUID string, caseUID string) error {
	return a.studyRelations.RemoveCaseFromStudy(ctx, studyUID, caseUID)
}

// SoftDeleteCase marks a case as deleted without removing it from the database
func (a *Adapter) SoftDeleteCase(ctx context.Context, caseUID string, deletedBy int) error {
	return a.crud.SoftDeleteCase(ctx, caseUID, deletedBy)
}

// GetDeletedCases retrieves all soft-deleted cases
func (a *Adapter) GetDeletedCases(ctx context.Context) ([]ports.Case, error) {
	return a.crud.GetDeletedCases(ctx)
}

// RestoreCase restores a soft-deleted case
func (a *Adapter) RestoreCase(ctx context.Context, caseUID string) error {
	return a.crud.RestoreCase(ctx, caseUID)
}

// GetCasesCount returns the total count of cases matching search criteria
func (a *Adapter) GetCasesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	return a.search.GetCasesCount(ctx, search)
}

// GetCasesByStudyUID retrieves cases by study ID with search/filter and pagination support
func (a *Adapter) GetCasesByStudyUID(ctx context.Context, studyUID string, params utils.PaginationAndSearchParams) ([]ports.Case, error) {
	return a.search.GetCasesByStudyUID(ctx, studyUID, params)
}

// GetCasesByStudyUIDCount returns the total count of cases for a study matching search criteria
func (a *Adapter) GetCasesByStudyUIDCount(ctx context.Context, studyUID string, search utils.SearchParams) (int, error) {
	return a.search.GetCasesByStudyUIDCount(ctx, studyUID, search)
}
