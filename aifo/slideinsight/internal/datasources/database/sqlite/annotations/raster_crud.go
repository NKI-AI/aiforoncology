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
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/gofiber/fiber/v2/log"
)

// RasterCrudService handles basic CRUD operations for raster annotations
type RasterCrudService struct {
	db *sql.DB
}

// NewRasterCrudService creates a new raster CRUD service instance
func NewRasterCrudService(db *sql.DB) *RasterCrudService {
	return &RasterCrudService{db: db}
}

// CreateMask adds a new raster annotation to the database
func (s *RasterCrudService) CreateMask(ctx context.Context, newMask ports.NewMask) error {
	// First get the internal slide ID from the slide UID
	var internalSlideID int
	err := s.db.QueryRow("SELECT id FROM slides WHERE slide_uid = ? AND deleted_at IS NULL", newMask.SlideUID).Scan(&internalSlideID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.WithDetails(errors.ErrMaskSlideNotFound, "slide UID '%s'", newMask.SlideUID)
		}
		return errors.NewDatabaseQueryError("slide ID lookup", err)
	}

	// Set default version if not provided
	version := newMask.Version
	if version == 0 {
		version = 1
	}

	// Set default format if not provided
	format := newMask.Format
	if format == "" {
		format = "tiff" // default format
	}

	// Validate format
	if format != "tiff" && format != "png" {
		return errors.WithDetails(errors.ErrMaskFormatNotSupported, "format '%s': only 'tiff' and 'png' are supported", format)
	}

	// Validate actor_type
	if newMask.ActorType != "user" && newMask.ActorType != "model" {
		return errors.WithDetails(errors.ErrInvalidInput, "invalid actor_type '%s': only 'user' and 'model' are supported", newMask.ActorType)
	}

	// Convert metadata to JSON string if needed
	metadata := newMask.Metadata
	if metadata == "" {
		metadata = "{}" // empty JSON object
	}

	_, err = s.db.Exec(`
		INSERT INTO raster_annotations 
		(tenant_id, actor_type, actor_id, creator_id, slide_id, raster_uid, version, name, file_uri, file_hash, format, labels, metadata, mask_width, mask_height, mask_mpp, mutable) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newMask.TenantID, newMask.ActorType, newMask.ActorID, newMask.CreatorID, internalSlideID, newMask.MaskUID, version, newMask.Name,
		newMask.MaskURI, newMask.FileHash, format, newMask.Labels, metadata, newMask.MaskWidth, newMask.MaskHeight, newMask.MaskMpp, newMask.Mutable)
	if err != nil {
		return errors.NewDatabaseInsertError("raster annotation", err)
	}
	return nil
}

// GetMaskByUID retrieves a specific raster annotation by its mask_id
func (s *RasterCrudService) GetMaskByUID(ctx context.Context, maskUID string) (ports.Mask, error) {
	var mask ports.Mask
	var labels sql.NullString
	var metadata sql.NullString
	var fileHash sql.NullString
	var maskWidth, maskHeight sql.NullInt64
	var maskMpp sql.NullFloat64
	var createdAtStr string
	var deletedAtStr sql.NullString
	var deletedBy sql.NullInt64

	query := `
		SELECT 
			ra.id, ra.actor_type, ra.actor_id, ra.creator_id, ra.slide_id, s.slide_uid, ra.raster_uid, ra.version, ra.name, 
			ra.file_uri, ra.file_hash, ra.format, ra.labels, ra.metadata, ra.mask_width, ra.mask_height, ra.mask_mpp, ra.mutable,
			ra.deleted_at, ra.deleted_by, ra.created_at
		FROM raster_annotations ra
		JOIN slides s ON ra.slide_id = s.id
		WHERE ra.raster_uid = ? AND ra.deleted_at IS NULL AND s.deleted_at IS NULL
	`
	err := s.db.QueryRow(query, maskUID).Scan(
		&mask.ID, &mask.ActorType, &mask.ActorID, &mask.CreatorID, &mask.SlideID, &mask.SlideUID, &mask.MaskUID, &mask.Version,
		&mask.Name, &mask.MaskURI, &fileHash, &mask.Format, &labels, &metadata,
		&maskWidth, &maskHeight, &maskMpp, &mask.Mutable,
		&deletedAtStr, &deletedBy, &createdAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.Mask{}, errors.NewMaskNotFoundError(maskUID)
		}
		return ports.Mask{}, errors.NewDatabaseQueryError("raster annotation", err)
	}

	// Set labels, metadata, and file_hash
	if labels.Valid {
		mask.Labels = labels.String
	}
	if metadata.Valid {
		mask.Metadata = metadata.String
	}
	if fileHash.Valid {
		mask.FileHash = &fileHash.String
	}

	if maskWidth.Valid {
		mask.MaskWidth = int(maskWidth.Int64)
	}

	if maskHeight.Valid {
		mask.MaskHeight = int(maskHeight.Int64)
	}

	if maskMpp.Valid {
		mask.MaskMpp = maskMpp.Float64
	}

	// Parse the created_at timestamp
	if createdAtStr != "" {
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			log.Error("failed to parse created_at timestamp", "error", err)
		} else {
			mask.CreatedAt = createdAt
		}
	}

	// Handle soft deletion fields
	if deletedAtStr.Valid {
		deletedAt, err := time.Parse(time.RFC3339, deletedAtStr.String)
		if err != nil {
			log.Error("failed to parse deleted_at timestamp", "error", err)
		} else {
			mask.DeletedAt = &deletedAt
		}
	}

	if deletedBy.Valid {
		deletedByInt := int(deletedBy.Int64)
		mask.DeletedBy = &deletedByInt
	}

	return mask, nil
}
