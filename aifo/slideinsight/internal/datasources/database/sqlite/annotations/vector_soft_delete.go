// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package annotations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/gofiber/fiber/v2/log"
)

// VectorSoftDeleteService handles vector annotation soft deletion operations
type VectorSoftDeleteService struct {
	db *sql.DB
}

// NewVectorSoftDeleteService creates a new vector soft delete service instance
func NewVectorSoftDeleteService(db *sql.DB) *VectorSoftDeleteService {
	return &VectorSoftDeleteService{db: db}
}

// SoftDeleteVectorAnnotation marks a vector annotation as deleted without removing it from the database
func (s *VectorSoftDeleteService) SoftDeleteVectorAnnotation(ctx context.Context, vectorUID string, deletedBy int) error {
	result, err := s.db.Exec("UPDATE vector_annotations SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ? WHERE vector_uid = ? AND deleted_at IS NULL", deletedBy, vectorUID)
	if err != nil {
		return errors.NewDatabaseUpdateError("vector annotation soft delete", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vector annotation with UID '%s' not found or already deleted", vectorUID)
	}

	return nil
}

// GetDeletedVectorAnnotations retrieves all soft-deleted vector annotations
func (s *VectorSoftDeleteService) GetDeletedVectorAnnotations(ctx context.Context) ([]ports.VectorAnnotation, error) {
	query := `
		SELECT 
			va.id, va.actor_type, va.actor_id, va.creator_id, va.slide_id, s.slide_uid, va.vector_uid, va.version, va.name, 
			va.file_uri, va.labels, va.metadata, va.deleted_at, va.deleted_by, va.created_at
		FROM vector_annotations va
		JOIN slides s ON va.slide_id = s.id
		WHERE va.deleted_at IS NOT NULL`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("deleted vector annotations", err)
	}
	defer rows.Close()

	var vectors []ports.VectorAnnotation
	for rows.Next() {
		var vector ports.VectorAnnotation
		var createdAtStr string
		var deletedAtStr sql.NullString
		var deletedBy sql.NullInt64
		var labels sql.NullString
		var metadata sql.NullString

		if err := rows.Scan(
			&vector.ID, &vector.ActorType, &vector.ActorID, &vector.CreatorID, &vector.SlideID, &vector.SlideUID, &vector.VectorUID, &vector.Version,
			&vector.Name, &vector.FileURI, &labels, &metadata,
			&deletedAtStr, &deletedBy, &createdAtStr,
		); err != nil {
			return nil, errors.NewDatabaseScanError("deleted vector annotation", err)
		}

		// Set labels and metadata
		if labels.Valid {
			vector.Labels = labels.String
		}
		if metadata.Valid {
			vector.Metadata = metadata.String
		}

		// Parse the created_at timestamp
		if createdAtStr != "" {
			createdAt, err := time.Parse(time.RFC3339, createdAtStr)
			if err != nil {
				log.Error("failed to parse created_at timestamp", "error", err)
			} else {
				vector.CreatedAt = createdAt
			}
		}

		// Handle soft deletion fields
		if deletedAtStr.Valid {
			deletedAt, err := time.Parse(time.RFC3339, deletedAtStr.String)
			if err != nil {
				log.Error("failed to parse deleted_at timestamp", "error", err)
			} else {
				vector.DeletedAt = &deletedAt
			}
		}

		if deletedBy.Valid {
			deletedByInt := int(deletedBy.Int64)
			vector.DeletedBy = &deletedByInt
		}

		vectors = append(vectors, vector)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("deleted vector annotations", err)
	}

	return vectors, nil
}

// RestoreVectorAnnotation restores a soft-deleted vector annotation
func (s *VectorSoftDeleteService) RestoreVectorAnnotation(ctx context.Context, vectorUID string) error {
	result, err := s.db.Exec("UPDATE vector_annotations SET deleted_at = NULL, deleted_by = NULL WHERE vector_uid = ? AND deleted_at IS NOT NULL", vectorUID)
	if err != nil {
		return errors.NewDatabaseUpdateError("vector annotation restore", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vector annotation with UID '%s' not found or not deleted", vectorUID)
	}

	return nil
}
