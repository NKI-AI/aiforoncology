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
	"fmt"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/gofiber/fiber/v2/log"
)

// StudyRelationService handles case-study relationship operations
type StudyRelationService struct {
	db *sql.DB
}

// NewStudyRelationService creates a new study relation service instance
func NewStudyRelationService(db *sql.DB) *StudyRelationService {
	return &StudyRelationService{db: db}
}

// GetCasesByStudyUID retrieves all cases that belong to a specific study
func (s *StudyRelationService) GetCasesByStudyUID(ctx context.Context, studyUID string) ([]ports.Case, error) {
	query := `
		SELECT c.id, c.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, c.short_uid as case_uid, c.creator_id, COALESCE(u.short_uid, '') as creator_uid, c.name, c.metadata, c.deleted_at, c.deleted_by, c.created_at 
		FROM cases c
		LEFT JOIN tenants t ON c.tenant_id = t.id
		LEFT JOIN users u ON c.creator_id = u.id
		INNER JOIN study_cases sc ON c.id = sc.case_id
		INNER JOIN studies s ON sc.study_id = s.id
		WHERE s.short_uid = ? AND c.deleted_at IS NULL`

	rows, err := s.db.Query(query, studyUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cases by study ID: %w", err)
	}
	defer rows.Close()

	var cases []ports.Case
	for rows.Next() {
		var case_ ports.Case
		var createdAtStr string
		var deletedAtStr sql.NullString
		var deletedBy sql.NullInt64
		if err := rows.Scan(
			&case_.ID,
			&case_.TenantID,
			&case_.TenantUID,
			&case_.CaseUID,
			&case_.CreatorID,
			&case_.CreatorUID,
			&case_.Name,
			&case_.Metadata,
			&deletedAtStr,
			&deletedBy,
			&createdAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan case row: %w", err)
		}

		// Parse the created_at timestamp
		if createdAtStr != "" {
			createdAt, err := time.Parse(time.RFC3339, createdAtStr)
			if err != nil {
				log.Error("failed to parse created_at timestamp", "error", err)
			} else {
				case_.CreatedAt = createdAt
			}
		}

		// Handle soft deletion fields
		if deletedAtStr.Valid {
			deletedAt, err := time.Parse(time.RFC3339, deletedAtStr.String)
			if err != nil {
				log.Error("failed to parse deleted_at timestamp", "error", err)
			} else {
				case_.DeletedAt = &deletedAt
			}
		}

		if deletedBy.Valid {
			deletedByInt := int(deletedBy.Int64)
			case_.DeletedBy = &deletedByInt
		}

		cases = append(cases, case_)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over case rows: %w", err)
	}

	return cases, nil
}

// AddCaseToStudy adds an existing case to a study via the study_cases table
func (s *StudyRelationService) AddCaseToStudy(ctx context.Context, studyUID string, caseUID string) error {
	// First, get the study ID and tenant_id from the study short UID
	var studyPK, studyTenantID int

	err := s.db.QueryRow("SELECT id, tenant_id FROM studies WHERE short_uid = ?", studyUID).Scan(&studyPK, &studyTenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("study with ID '%s' not found", studyUID)
		}
		return fmt.Errorf("failed to get study: %w", err)
	}

	// Get the case ID from the case short UID
	var casePK int
	err = s.db.QueryRow("SELECT id FROM cases WHERE short_uid = ?", caseUID).Scan(&casePK)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("case with UID '%s' not found", caseUID)
		}
		return fmt.Errorf("failed to get case: %w", err)
	}

	// Insert into study_cases table with tenant_id
	_, err = s.db.Exec("INSERT INTO study_cases (study_id, case_id, tenant_id) VALUES (?, ?, ?)", studyPK, casePK, studyTenantID)
	if err != nil {
		return fmt.Errorf("failed to add case to study: %w", err)
	}

	return nil
}

// RemoveCaseFromStudy removes a case from a study via the study_cases table
func (s *StudyRelationService) RemoveCaseFromStudy(ctx context.Context, studyUID string, caseUID string) error {
	// First, get the study ID and case ID from their short UIDs
	var studyPK, casePK int

	err := s.db.QueryRow("SELECT id FROM studies WHERE short_uid = ?", studyUID).Scan(&studyPK)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("study with ID '%s' not found", studyUID)
		}
		return fmt.Errorf("failed to get study: %w", err)
	}

	err = s.db.QueryRow("SELECT id FROM cases WHERE short_uid = ?", caseUID).Scan(&casePK)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("case with UID '%s' not found", caseUID)
		}
		return fmt.Errorf("failed to get case: %w", err)
	}

	// Delete from study_cases table
	result, err := s.db.Exec("DELETE FROM study_cases WHERE study_id = ? AND case_id = ?", studyPK, casePK)
	if err != nil {
		return fmt.Errorf("failed to remove case from study: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("case with UID '%s' is not associated with study '%s'", caseUID, studyUID)
	}

	return nil
}
