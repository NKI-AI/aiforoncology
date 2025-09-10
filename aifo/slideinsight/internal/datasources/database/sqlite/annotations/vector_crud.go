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
	"strings"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/gofiber/fiber/v2/log"
)

// VectorCrudService handles basic CRUD operations for vector annotations
type VectorCrudService struct {
	db *sql.DB
}

// NewVectorCrudService creates a new vector CRUD service instance
func NewVectorCrudService(db *sql.DB) *VectorCrudService {
	return &VectorCrudService{db: db}
}

// CreateVectorAnnotation adds a new vector annotation to the database
func (s *VectorCrudService) CreateVectorAnnotation(ctx context.Context, newVector ports.NewVectorAnnotation) error {
	// First get the internal slide ID from the slide UID
	var internalSlideID int
	err := s.db.QueryRow("SELECT id FROM slides WHERE slide_uid = ? AND deleted_at IS NULL", newVector.SlideUID).Scan(&internalSlideID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("cannot create vector annotation: slide with UID '%s' does not exist", newVector.SlideUID)
		}
		return errors.NewDatabaseQueryError("slide ID lookup", err)
	}

	// Set default version if not provided
	version := newVector.Version
	if version == 0 {
		version = 1
	}

	// Validate actor_type
	if newVector.ActorType != "user" && newVector.ActorType != "model" {
		return fmt.Errorf("invalid actor_type '%s': only 'user' and 'model' are supported", newVector.ActorType)
	}

	// Convert metadata to JSON string if needed
	metadata := newVector.Metadata
	if metadata == "" {
		metadata = "{}" // empty JSON object
	}

	_, err = s.db.Exec(`
		INSERT INTO vector_annotations 
		(tenant_id, actor_type, actor_id, creator_id, slide_id, vector_uid, version, name, file_uri, data_blob, labels, metadata, mutable) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newVector.TenantID, newVector.ActorType, newVector.ActorID, newVector.CreatorID, internalSlideID, newVector.VectorUID, version, newVector.Name,
		newVector.FileURI, newVector.DataBlob, newVector.Labels, metadata, newVector.Mutable)
	if err != nil {
		return errors.NewDatabaseInsertError("vector annotation", err)
	}
	return nil
}

// GetVectorAnnotationByUID retrieves a specific vector annotation by its vector_id
func (s *VectorCrudService) GetVectorAnnotationByUID(ctx context.Context, vectorUID string) (ports.VectorAnnotation, error) {
	var vector ports.VectorAnnotation
	var createdAtStr string
	var deletedAtStr sql.NullString
	var deletedBy sql.NullInt64
	var labels sql.NullString
	var metadata sql.NullString

	query := `
		SELECT 
			va.id, va.actor_type, va.actor_id, va.creator_id, va.slide_id, s.slide_uid, va.vector_uid, va.version, va.name, 
			va.file_uri, va.data_blob, va.labels, va.metadata, va.mutable, va.deleted_at, va.deleted_by, va.created_at, va.updated_at
		FROM vector_annotations va
		JOIN slides s ON va.slide_id = s.id
		WHERE va.vector_uid = ? AND va.deleted_at IS NULL AND s.deleted_at IS NULL
	`
	var updatedAtStr string
	var dataBlob sql.NullString
	err := s.db.QueryRow(query, vectorUID).Scan(
		&vector.ID, &vector.ActorType, &vector.ActorID, &vector.CreatorID, &vector.SlideID, &vector.SlideUID, &vector.VectorUID, &vector.Version,
		&vector.Name, &vector.FileURI, &dataBlob, &labels, &metadata, &vector.Mutable,
		&deletedAtStr, &deletedBy, &createdAtStr, &updatedAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.VectorAnnotation{}, fmt.Errorf("vector annotation with UID '%s' not found", vectorUID)
		}
		return ports.VectorAnnotation{}, errors.NewDatabaseQueryError("vector annotation", err)
	}

	// Set labels, metadata, and data_blob
	if labels.Valid {
		vector.Labels = labels.String
	}
	if metadata.Valid {
		vector.Metadata = metadata.String
	}
	if dataBlob.Valid {
		vector.DataBlob = dataBlob.String
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

	// Parse the updated_at timestamp
	if updatedAtStr != "" {
		updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			log.Error("failed to parse updated_at timestamp", "error", err)
		} else {
			vector.UpdatedAt = updatedAt
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

	return vector, nil
}

// UpdateVectorAnnotation updates an existing vector annotation
func (s *VectorCrudService) UpdateVectorAnnotation(ctx context.Context, vectorUID string, updates ports.UpdateVectorAnnotation) error {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}

	if updates.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *updates.Name)
	}

	if updates.FileURI != nil {
		setParts = append(setParts, "file_uri = ?")
		args = append(args, *updates.FileURI)
	}

	if updates.DataBlob != nil {
		setParts = append(setParts, "data_blob = ?")
		args = append(args, *updates.DataBlob)
	}

	if updates.Labels != nil {
		setParts = append(setParts, "labels = ?")
		args = append(args, *updates.Labels)
	}

	if updates.Metadata != nil {
		setParts = append(setParts, "metadata = ?")
		args = append(args, *updates.Metadata)
	}

	if updates.Mutable != nil {
		setParts = append(setParts, "mutable = ?")
		args = append(args, *updates.Mutable)
	}

	// If no updates provided, return early
	if len(setParts) == 0 {
		return nil
	}

	// Add updated_at timestamp (will be handled by trigger, but we can set it explicitly)
	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")

	// Add vector_uid to args for WHERE clause
	args = append(args, vectorUID)

	query := fmt.Sprintf(`
		UPDATE vector_annotations 
		SET %s 
		WHERE vector_uid = ? AND deleted_at IS NULL`,
		strings.Join(setParts, ", "))

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return errors.NewDatabaseUpdateError("vector annotation", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseUpdateError("vector annotation rows affected check", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vector annotation with UID '%s' not found or already deleted", vectorUID)
	}

	return nil
}
