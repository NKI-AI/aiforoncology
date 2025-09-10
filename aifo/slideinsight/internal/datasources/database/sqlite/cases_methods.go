// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package sqlite

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Case-related methods that delegate to the cases adapter

// CreateCase adds a new case to the database
func (db *DB) CreateCase(ctx context.Context, newCase ports.NewCase) error {
	return db.cases.CreateCase(ctx, newCase)
}

// LoadAllCases retrieves cases from the database with search/filter and pagination support
func (db *DB) LoadAllCases(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Case, error) {
	return db.cases.LoadAllCases(ctx, search, pagination)
}

// GetCaseByUID retrieves a specific case by its ID
func (db *DB) GetCaseByUID(ctx context.Context, caseUID string) (ports.Case, error) {
	return db.cases.GetCaseByUID(ctx, caseUID)
}

// AddCaseToStudy adds an existing case to a study via the study_cases table
func (db *DB) AddCaseToStudy(ctx context.Context, studyUID string, caseUID string) error {
	return db.cases.AddCaseToStudy(ctx, studyUID, caseUID)
}

// RemoveCaseFromStudy removes a case from a study via the study_cases table
func (db *DB) RemoveCaseFromStudy(ctx context.Context, studyUID string, caseUID string) error {
	return db.cases.RemoveCaseFromStudy(ctx, studyUID, caseUID)
}

// SoftDeleteCase marks a case as deleted without removing it from the database
func (db *DB) SoftDeleteCase(ctx context.Context, caseUID string, deletedBy int) error {
	return db.cases.SoftDeleteCase(ctx, caseUID, deletedBy)
}

// GetDeletedCases retrieves all soft-deleted cases
func (db *DB) GetDeletedCases(ctx context.Context) ([]ports.Case, error) {
	return db.cases.GetDeletedCases(ctx)
}

// RestoreCase restores a soft-deleted case
func (db *DB) RestoreCase(ctx context.Context, caseUID string) error {
	return db.cases.RestoreCase(ctx, caseUID)
}

// GetCasesCount returns the total count of cases matching search criteria
func (db *DB) GetCasesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	return db.cases.GetCasesCount(ctx, search)
}

// GetCasesByStudyUID retrieves cases by study ID with search/filter and pagination support
func (db *DB) GetCasesByStudyUID(ctx context.Context, studyUID string, params utils.PaginationAndSearchParams) ([]ports.Case, error) {
	return db.cases.GetCasesByStudyUID(ctx, studyUID, params)
}

// GetCasesByStudyUIDCount returns the total count of cases for a study matching search criteria
func (db *DB) GetCasesByStudyUIDCount(ctx context.Context, studyUID string, search utils.SearchParams) (int, error) {
	return db.cases.GetCasesByStudyUIDCount(ctx, studyUID, search)
}
