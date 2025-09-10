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

// Study-related methods that delegate to the studies adapter

// CreateStudy adds a new study to the database
func (db *DB) CreateStudy(ctx context.Context, newStudy ports.NewStudy) error {
	return db.studies.CreateStudy(ctx, newStudy)
}

// GetStudyByUID retrieves a specific study by its ID
func (db *DB) GetStudyByUID(ctx context.Context, studyUID string) (ports.Study, error) {
	return db.studies.GetStudyByUID(ctx, studyUID)
}

// GetStudyByInternalID retrieves a specific study by its internal database ID
func (db *DB) GetStudyByInternalID(ctx context.Context, studyID int) (ports.Study, error) {
	return db.studies.GetStudyByInternalID(ctx, studyID)
}

// UpdateStudy updates study information for a study with the specified ID
func (db *DB) UpdateStudy(ctx context.Context, studyUID string, updates ports.StudyUpdates) error {
	return db.studies.UpdateStudy(ctx, studyUID, updates)
}

// LoadAllStudies retrieves studies from the database with search/filter and pagination support
func (db *DB) LoadAllStudies(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Study, error) {
	return db.studies.LoadAllStudies(ctx, search, pagination)
}

// GetStudiesCount returns the total count of studies matching search criteria
func (db *DB) GetStudiesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	return db.studies.GetStudiesCount(ctx, search)
}

// GetStudyIDByShortUID retrieves the internal study ID by its short UID
func (db *DB) GetStudyIDByShortUID(ctx context.Context, studyUID string) (int, error) {
	return db.studies.GetStudyIDByShortUID(ctx, studyUID)
}

// GetStudyCaseCounts retrieves case counts for all studies
func (db *DB) GetStudyCaseCounts(ctx context.Context) (map[string]int, error) {
	return db.studies.GetStudyCaseCounts(ctx)
}

// GetStudySlideCounts retrieves slide counts for all studies
func (db *DB) GetStudySlideCounts(ctx context.Context) (map[string]int, error) {
	return db.studies.GetStudySlideCounts(ctx)
}

// GetStudyCaseAndSlideCounts retrieves both case and slide counts for all studies efficiently
func (db *DB) GetStudyCaseAndSlideCounts(ctx context.Context) (map[string]int, map[string]int, error) {
	return db.studies.GetStudyCaseAndSlideCounts(ctx)
}

// SoftDeleteStudy marks a study as deleted without removing it from the database
func (db *DB) SoftDeleteStudy(ctx context.Context, studyUID string, deletedBy int) error {
	return db.studies.SoftDeleteStudy(ctx, studyUID, deletedBy)
}

// GetDeletedStudies retrieves all soft-deleted studies
func (db *DB) GetDeletedStudies(ctx context.Context) ([]ports.Study, error) {
	return db.studies.GetDeletedStudies(ctx)
}

// RestoreStudy restores a soft-deleted study
func (db *DB) RestoreStudy(ctx context.Context, studyUID string) error {
	return db.studies.RestoreStudy(ctx, studyUID)
}

// LoadStudiesByIDs retrieves studies filtered by a list of IDs with search/filter and pagination support
func (db *DB) LoadStudiesByIDs(ctx context.Context, studyIDs []int, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Study, error) {
	return db.studies.LoadStudiesByIDs(ctx, studyIDs, search, pagination)
}

// CountStudiesByIDs returns the total count of studies in the ID list matching search criteria
func (db *DB) CountStudiesByIDs(ctx context.Context, studyIDs []int, search utils.SearchParams) (int, error) {
	return db.studies.CountStudiesByIDs(ctx, studyIDs, search)
}
