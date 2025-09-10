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

// RasterSoftDeleteService handles raster annotation soft deletion operations
type RasterSoftDeleteService struct {
	db *sql.DB
}

// NewRasterSoftDeleteService creates a new raster soft delete service instance
func NewRasterSoftDeleteService(db *sql.DB) *RasterSoftDeleteService {
	return &RasterSoftDeleteService{db: db}
}

// SoftDeleteMask marks a mask as deleted without removing it from the database
func (s *RasterSoftDeleteService) SoftDeleteMask(ctx context.Context, maskUID string, deletedBy int) error {
	result, err := s.db.Exec("UPDATE raster_annotations SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ? WHERE raster_uid = ? AND deleted_at IS NULL", deletedBy, maskUID)
	if err != nil {
		return errors.NewDatabaseUpdateError("raster annotation soft delete", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("mask with UID '%s' not found or already deleted", maskUID)
	}

	return nil
}

// GetDeletedMasks retrieves all soft-deleted masks
func (s *RasterSoftDeleteService) GetDeletedMasks(ctx context.Context) ([]ports.Mask, error) {
	query := `
		SELECT 
			ra.id, ra.actor_type, ra.actor_id, ra.creator_id, ra.slide_id, s.slide_uid, ra.raster_uid, ra.version, ra.name, 
			ra.file_uri, ra.format, ra.labels, ra.metadata, ra.mask_width, ra.mask_height, ra.mask_mpp,
			ra.deleted_at, ra.deleted_by, ra.created_at
		FROM raster_annotations ra
		JOIN slides s ON ra.slide_id = s.id
		WHERE ra.deleted_at IS NOT NULL`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("deleted raster annotations", err)
	}
	defer rows.Close()

	var masks []ports.Mask
	for rows.Next() {
		var mask ports.Mask
		var labels sql.NullString
		var metadata sql.NullString
		var maskWidth, maskHeight sql.NullInt64
		var maskMpp sql.NullFloat64
		var createdAtStr string
		var deletedAtStr sql.NullString
		var deletedBy sql.NullInt64

		if err := rows.Scan(
			&mask.ID, &mask.ActorType, &mask.ActorID, &mask.CreatorID, &mask.SlideID, &mask.SlideUID, &mask.MaskUID, &mask.Version,
			&mask.Name, &mask.MaskURI, &mask.Format, &labels, &metadata,
			&maskWidth, &maskHeight, &maskMpp,
			&deletedAtStr, &deletedBy, &createdAtStr,
		); err != nil {
			return nil, errors.NewDatabaseScanError("deleted raster annotation", err)
		}

		// Set labels and metadata
		if labels.Valid {
			mask.Labels = labels.String
		}
		if metadata.Valid {
			mask.Metadata = metadata.String
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

		masks = append(masks, mask)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("deleted raster annotations", err)
	}

	return masks, nil
}

// RestoreMask restores a soft-deleted mask
func (s *RasterSoftDeleteService) RestoreMask(ctx context.Context, maskUID string) error {
	result, err := s.db.Exec("UPDATE raster_annotations SET deleted_at = NULL, deleted_by = NULL WHERE raster_uid = ? AND deleted_at IS NOT NULL", maskUID)
	if err != nil {
		return errors.NewDatabaseUpdateError("raster annotation restore", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("mask with UID '%s' not found or not deleted", maskUID)
	}

	return nil
}
