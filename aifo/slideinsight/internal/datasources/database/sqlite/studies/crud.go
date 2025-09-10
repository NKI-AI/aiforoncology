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
	"fmt"
	"strings"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/gofiber/fiber/v2/log"
)

// CrudService handles basic CRUD operations for studies
type CrudService struct {
	db *sql.DB
}

// NewCrudService creates a new CRUD service instance
func NewCrudService(db *sql.DB) *CrudService {
	return &CrudService{db: db}
}

// CreateStudy adds a new study to the database
func (s *CrudService) CreateStudy(ctx context.Context, newStudy ports.NewStudy) error {
	_, err := s.db.Exec("INSERT INTO studies (tenant_id, short_uid, creator_id, name, description, is_published) VALUES (?, ?, ?, ?, ?, ?)",
		newStudy.TenantID, newStudy.StudyUID, newStudy.CreatorID, newStudy.Name, newStudy.Description, false)
	if err != nil {
		return errors.NewDatabaseInsertError("study", err)
	}
	return nil
}

// GetStudyByUID retrieves a specific study by its ID
func (s *CrudService) GetStudyByUID(ctx context.Context, studyUID string) (ports.Study, error) {
	var study ports.Study
	var createdAtStr string
	var deletedAtStr sql.NullString
	var deletedBy sql.NullInt64

	query := `
		SELECT s.id, s.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, s.short_uid as study_uid, s.creator_id, COALESCE(u.short_uid, '') as creator_uid, s.name, s.description, s.metadata, s.is_published, s.deleted_at, s.deleted_by, s.created_at 
		FROM studies s
		LEFT JOIN tenants t ON s.tenant_id = t.id
		LEFT JOIN users u ON s.creator_id = u.id
		WHERE s.short_uid = ? AND s.deleted_at IS NULL`

	err := s.db.QueryRow(query, studyUID).Scan(
		&study.ID, &study.TenantID, &study.TenantUID, &study.StudyUID, &study.CreatorID, &study.CreatorUID, &study.Name, &study.Description, &study.Metadata, &study.IsPublished, &deletedAtStr, &deletedBy, &createdAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.Study{}, errors.NewStudyNotFoundError(studyUID)
		}
		return ports.Study{}, errors.NewDatabaseQueryError("study", err)
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

	return study, nil
}

// GetStudyByInternalID retrieves a specific study by its internal database ID
func (s *CrudService) GetStudyByInternalID(ctx context.Context, studyID int) (ports.Study, error) {
	var study ports.Study
	var createdAtStr string
	var deletedAtStr sql.NullString
	var deletedBy sql.NullInt64

	query := `
		SELECT s.id, s.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, s.short_uid as study_uid, s.creator_id, COALESCE(u.short_uid, '') as creator_uid, s.name, s.description, s.metadata, s.is_published, s.deleted_at, s.deleted_by, s.created_at 
		FROM studies s
		LEFT JOIN tenants t ON s.tenant_id = t.id
		LEFT JOIN users u ON s.creator_id = u.id
		WHERE s.id = ? AND s.deleted_at IS NULL`

	err := s.db.QueryRow(query, studyID).Scan(
		&study.ID, &study.TenantID, &study.TenantUID, &study.StudyUID, &study.CreatorID, &study.CreatorUID, &study.Name, &study.Description, &study.Metadata, &study.IsPublished, &deletedAtStr, &deletedBy, &createdAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.Study{}, errors.WithDetails(errors.ErrStudyNotFound, "study with internal ID %d not found", studyID)
		}
		return ports.Study{}, errors.NewDatabaseQueryError("study", err)
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

	return study, nil
}

// UpdateStudy updates study information for a study with the specified ID
func (s *CrudService) UpdateStudy(ctx context.Context, studyUID string, updates ports.StudyUpdates) error {
	var setParts []string
	var args []interface{}

	// Build SET clause dynamically based on provided updates
	if updates.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *updates.Name)
	}
	if updates.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *updates.Description)
	}
	if updates.Metadata != nil {
		setParts = append(setParts, "metadata = ?")
		args = append(args, *updates.Metadata)
	}
	if updates.IsPublished != nil {
		setParts = append(setParts, "is_published = ?")
		args = append(args, *updates.IsPublished)
	}

	if len(setParts) == 0 {
		return errors.ErrNoFieldsToUpdate
	}

	// Add studyUID to args for WHERE clause
	args = append(args, studyUID)

	query := fmt.Sprintf("UPDATE studies SET %s WHERE short_uid = ? AND deleted_at IS NULL", strings.Join(setParts, ", "))
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return errors.NewDatabaseUpdateError("study", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.WithDetails(errors.ErrStudyNotFound, "study with ID '%s' not found or is deleted", studyUID)
	}

	return nil
}
