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

// Case represents a case in the database
type Case struct {
	ID         int
	TenantID   int
	TenantUID  string
	CaseUID    string
	CreatorID  int
	CreatorUID string
	Name       string
	Metadata   []byte
	DeletedAt  *time.Time
	DeletedBy  *int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewCase represents a new case to be created to the database
type NewCase struct {
	TenantID  int
	CaseUID   string
	CreatorID int
	Name      string
	Metadata  []byte
}

// CasesRepository defines the interface for case-related database operations
type CasesRepository interface {
	// LoadAllCases retrieves cases from the database with search/filter and pagination support
	LoadAllCases(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]Case, error)

	// GetCasesCount returns the total count of cases matching search criteria
	GetCasesCount(ctx context.Context, search utils.SearchParams) (int, error)

	// CreateCase adds a new case to the database.
	CreateCase(ctx context.Context, newCase NewCase) error

	// GetCaseByUID retrieves a specific case by its ID.
	GetCaseByUID(ctx context.Context, caseUID string) (Case, error)

	// GetCasesByStudyUID retrieves cases by study ID with search/filter and pagination support.
	GetCasesByStudyUID(ctx context.Context, studyUID string, params utils.PaginationAndSearchParams) ([]Case, error)

	// GetCasesByStudyUIDCount returns the total count of cases for a study matching search criteria.
	GetCasesByStudyUIDCount(ctx context.Context, studyUID string, search utils.SearchParams) (int, error)

	// AddCaseToStudy adds an existing case to a study via the study_cases table.
	AddCaseToStudy(ctx context.Context, studyUID string, caseUID string) error

	// RemoveCaseFromStudy removes a case from a study via the study_cases table.
	RemoveCaseFromStudy(ctx context.Context, studyUID string, caseUID string) error

	// SoftDeleteCase marks a case as deleted without removing it from the database.
	SoftDeleteCase(ctx context.Context, caseUID string, deletedBy int) error

	// GetDeletedCases retrieves all soft-deleted cases.
	GetDeletedCases(ctx context.Context) ([]Case, error)

	// RestoreCase restores a soft-deleted case.
	RestoreCase(ctx context.Context, caseUID string) error
}
