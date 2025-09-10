// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package ports

import (
	"context"
	"time"

	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Study represents a study in the database
type Study struct {
	ID          int
	TenantID    int
	TenantUID   string
	StudyUID    string
	CreatorID   int
	CreatorUID  string
	Name        string
	Description string
	Metadata    []byte
	IsPublished bool
	DeletedAt   *time.Time
	DeletedBy   *int
	CreatedAt   time.Time
}

// NewStudy represents a new study to be created to the database
type NewStudy struct {
	TenantID    int
	StudyUID    string
	CreatorID   int
	Name        string
	Description string
	Metadata    []byte
	IsPublished bool
}

// StudyUpdates represents fields that can be updated for an existing study
type StudyUpdates struct {
	Name        *string
	Description *string
	Metadata    *[]byte
	IsPublished *bool
}

// StudiesRepository defines the interface for study-related database operations
type StudiesRepository interface {
	// LoadAllStudies retrieves studies from the database with search/filter and pagination support
	LoadAllStudies(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]Study, error)

	// GetStudiesCount returns the total count of studies matching search criteria
	GetStudiesCount(ctx context.Context, search utils.SearchParams) (int, error)

	// CreateStudy adds a new study to the database.
	CreateStudy(ctx context.Context, newStudy NewStudy) error

	// GetStudyByUID retrieves a specific study by its ID.
	GetStudyByUID(ctx context.Context, studyUID string) (Study, error)

	// GetStudyByInternalID retrieves a specific study by its internal database ID.
	GetStudyByInternalID(ctx context.Context, studyID int) (Study, error)

	// UpdateStudy updates study information for a study with the specified ID.
	UpdateStudy(ctx context.Context, studyUID string, updates StudyUpdates) error

	// GetStudyIDByShortUID retrieves the internal study ID by its short UID.
	GetStudyIDByShortUID(ctx context.Context, studyUID string) (int, error)

	// GetStudyCaseCounts retrieves case counts for all studies.
	GetStudyCaseCounts(ctx context.Context) (map[string]int, error)

	// GetStudySlideCounts retrieves slide counts for all studies.
	GetStudySlideCounts(ctx context.Context) (map[string]int, error)

	// GetStudyCaseAndSlideCounts retrieves both case and slide counts for all studies efficiently.
	GetStudyCaseAndSlideCounts(ctx context.Context) (map[string]int, map[string]int, error)

	// SoftDeleteStudy marks a study as deleted without removing it from the database.
	SoftDeleteStudy(ctx context.Context, studyUID string, deletedBy int) error

	// GetDeletedStudies retrieves all soft-deleted studies.
	GetDeletedStudies(ctx context.Context) ([]Study, error)

	// RestoreStudy restores a soft-deleted study.
	RestoreStudy(ctx context.Context, studyUID string) error

	// SQL-based permission filtering support
	// LoadStudiesByIDs retrieves studies filtered by a list of IDs with search/filter and pagination support
	LoadStudiesByIDs(ctx context.Context, studyIDs []int, search utils.SearchParams, pagination utils.PaginationParams) ([]Study, error)

	// CountStudiesByIDs returns the total count of studies in the ID list matching search criteria
	CountStudiesByIDs(ctx context.Context, studyIDs []int, search utils.SearchParams) (int, error)
}
