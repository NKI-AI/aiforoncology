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
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/gofiber/fiber/v2/log"
)

// SoftDeleteService handles study soft deletion operations
type SoftDeleteService struct {
	db *sql.DB
}

// NewSoftDeleteService creates a new soft delete service instance
func NewSoftDeleteService(db *sql.DB) *SoftDeleteService {
	return &SoftDeleteService{db: db}
}

// SoftDeleteStudy marks a study as deleted without removing it from the database
func (s *SoftDeleteService) SoftDeleteStudy(ctx context.Context, studyUID string, deletedBy int) error {
	result, err := s.db.Exec("UPDATE studies SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ? WHERE short_uid = ? AND deleted_at IS NULL", deletedBy, studyUID)
	if err != nil {
		return errors.NewDatabaseUpdateError("study soft delete", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.NewStudyAlreadyDeletedError(studyUID)
	}

	return nil
}

// GetDeletedStudies retrieves all soft-deleted studies
func (s *SoftDeleteService) GetDeletedStudies(ctx context.Context) ([]ports.Study, error) {
	query := `
		SELECT s.id, s.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, s.short_uid as study_uid, s.creator_id, COALESCE(u.short_uid, '') as creator_uid, s.name, s.description, s.metadata, s.is_published, s.deleted_at, s.deleted_by, s.created_at 
		FROM studies s
		LEFT JOIN tenants t ON s.tenant_id = t.id
		LEFT JOIN users u ON s.creator_id = u.id
		WHERE s.deleted_at IS NOT NULL`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("deleted studies", err)
	}
	defer rows.Close()

	var studies []ports.Study
	for rows.Next() {
		var study ports.Study
		var createdAtStr string
		var deletedAtStr sql.NullString
		var deletedBy sql.NullInt64
		if err := rows.Scan(
			&study.ID,
			&study.TenantID,
			&study.TenantUID,
			&study.StudyUID,
			&study.CreatorID,
			&study.CreatorUID,
			&study.Name,
			&study.Description,
			&study.Metadata,
			&study.IsPublished,
			&deletedAtStr,
			&deletedBy,
			&createdAtStr); err != nil {
			return nil, errors.NewDatabaseScanError("deleted study", err)
		}

		// Parse the created_at timestamp
		if createdAtStr != "" {
			createdAt, err := time.Parse(time.RFC3339, createdAtStr)
			if err != nil {
				log.Error("failed to parse created_at timestamp", "error", err)
			} else {
				study.CreatedAt = createdAt
			}
		}

		// Handle soft deletion fields
		if deletedAtStr.Valid {
			deletedAt, err := time.Parse(time.RFC3339, deletedAtStr.String)
			if err != nil {
				log.Error("failed to parse deleted_at timestamp", "error", err)
			} else {
				study.DeletedAt = &deletedAt
			}
		}

		if deletedBy.Valid {
			deletedByInt := int(deletedBy.Int64)
			study.DeletedBy = &deletedByInt
		}

		studies = append(studies, study)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("deleted studies", err)
	}

	return studies, nil
}

// RestoreStudy restores a soft-deleted study
func (s *SoftDeleteService) RestoreStudy(ctx context.Context, studyUID string) error {
	result, err := s.db.Exec("UPDATE studies SET deleted_at = NULL, deleted_by = NULL WHERE short_uid = ? AND deleted_at IS NOT NULL", studyUID)
	if err != nil {
		return errors.NewDatabaseUpdateError("study restore", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.NewStudyNotDeletedError(studyUID)
	}

	return nil
}
