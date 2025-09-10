// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package regions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Adapter provides a unified interface for all region operations
type Adapter struct {
	db *sql.DB
}

// NewAdapter creates a new regions adapter
func NewAdapter(db *sql.DB) *Adapter {
	return &Adapter{
		db: db,
	}
}

// TODO: Implement all these methods properly. For now, they return stub responses to prevent panics.

// LoadAllRegions retrieves all regions from the database
func (a *Adapter) LoadAllRegions(ctx context.Context) ([]ports.Region, error) {
	// TODO: Implement proper database query
	return []ports.Region{}, nil
}

// GetRegionsGeneric retrieves regions with pagination and search support
func (a *Adapter) GetRegionsGeneric(ctx context.Context, params utils.PaginationAndSearchParams, searchParams ports.RegionSearchParams) ([]ports.Region, domain.PaginationInfo, error) {
	// TODO: Implement proper database query with search and pagination
	return []ports.Region{}, domain.PaginationInfo{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      0,
		TotalPages: 0,
		HasNext:    false,
		HasPrev:    false,
	}, nil
}

// GetRegionsBySlideUID retrieves all regions for a specific slide
func (a *Adapter) GetRegionsBySlideUID(ctx context.Context, slideUID string) ([]ports.Region, error) {
	query := `
		SELECT 
			r.id, r.actor_type, r.actor_id, r.creator_id, r.slide_id, s.slide_uid, r.version,
			r.name, r.region_type, r.geometry_data, r.coordinate_system, r.area_pixels, r.area_physical,
			r.labels, r.metadata, r.mutable, r.visible, r.style_config,
			r.deleted_at, r.deleted_by, r.created_at, r.updated_at
		FROM regions r
		JOIN slides s ON r.slide_id = s.id
		WHERE s.slide_uid = ? AND r.deleted_at IS NULL AND s.deleted_at IS NULL
		ORDER BY r.created_at ASC
	`

	rows, err := a.db.QueryContext(ctx, query, slideUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query regions for slide %s: %w", slideUID, err)
	}
	defer rows.Close()

	var regions []ports.Region
	for rows.Next() {
		var region ports.Region
		var deletedAt, deletedBy sql.NullString
		var areaPixels sql.NullInt64
		var areaPhysical sql.NullFloat64

		err := rows.Scan(
			&region.ID, &region.ActorType, &region.ActorID, &region.CreatorID, &region.SlideID, &region.SlideUID,
			&region.Version, &region.Name, &region.RegionType, &region.GeometryData, &region.CoordinateSystem,
			&areaPixels, &areaPhysical, &region.Labels, &region.Metadata, &region.Mutable, &region.Visible,
			&region.StyleConfig, &deletedAt, &deletedBy, &region.CreatedAt, &region.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan region: %w", err)
		}

		// Handle nullable fields
		if areaPixels.Valid {
			pixels := int(areaPixels.Int64)
			region.AreaPixels = &pixels
		}
		if areaPhysical.Valid {
			region.AreaPhysical = &areaPhysical.Float64
		}

		regions = append(regions, region)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over regions: %w", err)
	}

	return regions, nil
}

// CreateRegion adds a new region to the database
func (a *Adapter) CreateRegion(ctx context.Context, newRegion ports.NewRegion) error {
	// First get the internal slide ID from the slide UID
	var internalSlideID int
	err := a.db.QueryRowContext(ctx, "SELECT id FROM slides WHERE slide_uid = ? AND deleted_at IS NULL", newRegion.SlideUID).Scan(&internalSlideID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("cannot create region: slide with UID '%s' does not exist", newRegion.SlideUID)
		}
		return fmt.Errorf("failed to get slide ID: %w", err)
	}

	// Set default version if not provided
	version := newRegion.Version
	if version == 0 {
		version = 1
	}

	// Validate actor_type
	if newRegion.ActorType != "user" && newRegion.ActorType != "model" {
		return fmt.Errorf("invalid actor_type '%s': only 'user' and 'model' are supported", newRegion.ActorType)
	}

	// Validate region_type
	validRegionTypes := []string{"roi", "patient", "tissue", "artifact", "background", "other"}
	isValidType := false
	for _, validType := range validRegionTypes {
		if newRegion.RegionType == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("invalid region_type '%s': must be one of %v", newRegion.RegionType, validRegionTypes)
	}

	// Validate coordinate_system
	if newRegion.CoordinateSystem != "pixel" && newRegion.CoordinateSystem != "physical" {
		return fmt.Errorf("invalid coordinate_system '%s': must be 'pixel' or 'physical'", newRegion.CoordinateSystem)
	}

	// Convert metadata to JSON string if needed
	metadata := newRegion.Metadata
	if metadata == "" {
		metadata = "{}" // empty JSON object
	}

	labels := newRegion.Labels
	if labels == "" {
		labels = "{}" // empty JSON object
	}

	styleConfig := newRegion.StyleConfig
	if styleConfig == "" {
		styleConfig = "{}" // empty JSON object
	}

	query := `
		INSERT INTO regions 
		(id, slide_id, tenant_id, actor_type, actor_id, creator_id, version, name, region_type, 
		 geometry_data, coordinate_system, area_pixels, area_physical, labels, metadata, 
		 mutable, visible, style_config) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = a.db.ExecContext(ctx, query,
		newRegion.ID, internalSlideID, newRegion.TenantID, newRegion.ActorType, newRegion.ActorID,
		newRegion.CreatorID, version, newRegion.Name, newRegion.RegionType,
		newRegion.GeometryData, newRegion.CoordinateSystem, newRegion.AreaPixels, newRegion.AreaPhysical,
		labels, metadata, newRegion.Mutable, newRegion.Visible, styleConfig)
	if err != nil {
		return fmt.Errorf("failed to create region: %w", err)
	}

	return nil
}

// GetRegionByID retrieves a specific region by its ID
func (a *Adapter) GetRegionByID(ctx context.Context, regionID string) (ports.Region, error) {
	query := `
		SELECT 
			r.id, r.actor_type, r.actor_id, r.creator_id, r.slide_id, s.slide_uid, r.version,
			r.name, r.region_type, r.geometry_data, r.coordinate_system, r.area_pixels, r.area_physical,
			r.labels, r.metadata, r.mutable, r.visible, r.style_config,
			r.deleted_at, r.deleted_by, r.created_at, r.updated_at
		FROM regions r
		JOIN slides s ON r.slide_id = s.id
		WHERE r.id = ? AND r.deleted_at IS NULL AND s.deleted_at IS NULL
	`

	var region ports.Region
	var deletedAt, deletedBy sql.NullString
	var areaPixels sql.NullInt64
	var areaPhysical sql.NullFloat64

	err := a.db.QueryRowContext(ctx, query, regionID).Scan(
		&region.ID, &region.ActorType, &region.ActorID, &region.CreatorID, &region.SlideID, &region.SlideUID,
		&region.Version, &region.Name, &region.RegionType, &region.GeometryData, &region.CoordinateSystem,
		&areaPixels, &areaPhysical, &region.Labels, &region.Metadata, &region.Mutable, &region.Visible,
		&region.StyleConfig, &deletedAt, &deletedBy, &region.CreatedAt, &region.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.Region{}, fmt.Errorf("region with ID '%s' not found", regionID)
		}
		return ports.Region{}, fmt.Errorf("failed to query region by ID %s: %w", regionID, err)
	}

	// Handle nullable fields
	if areaPixels.Valid {
		pixels := int(areaPixels.Int64)
		region.AreaPixels = &pixels
	}
	if areaPhysical.Valid {
		region.AreaPhysical = &areaPhysical.Float64
	}

	return region, nil
}

// UpdateRegion updates an existing region
func (a *Adapter) UpdateRegion(ctx context.Context, regionID string, updates ports.UpdateRegion) error {
	// Build dynamic update query based on provided fields
	setParts := []string{}
	args := []interface{}{}

	if updates.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *updates.Name)
	}
	if updates.RegionType != nil {
		// Validate region_type
		validRegionTypes := []string{"roi", "patient", "tissue", "artifact", "background", "other"}
		isValidType := false
		for _, validType := range validRegionTypes {
			if *updates.RegionType == validType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			return fmt.Errorf("invalid region_type '%s': must be one of %v", *updates.RegionType, validRegionTypes)
		}
		setParts = append(setParts, "region_type = ?")
		args = append(args, *updates.RegionType)
	}
	if updates.GeometryData != nil {
		setParts = append(setParts, "geometry_data = ?")
		args = append(args, *updates.GeometryData)
	}
	if updates.CoordinateSystem != nil {
		// Validate coordinate_system
		if *updates.CoordinateSystem != "pixel" && *updates.CoordinateSystem != "physical" {
			return fmt.Errorf("invalid coordinate_system '%s': must be 'pixel' or 'physical'", *updates.CoordinateSystem)
		}
		setParts = append(setParts, "coordinate_system = ?")
		args = append(args, *updates.CoordinateSystem)
	}
	if updates.AreaPixels != nil {
		setParts = append(setParts, "area_pixels = ?")
		args = append(args, *updates.AreaPixels)
	}
	if updates.AreaPhysical != nil {
		setParts = append(setParts, "area_physical = ?")
		args = append(args, *updates.AreaPhysical)
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
	if updates.Visible != nil {
		setParts = append(setParts, "visible = ?")
		args = append(args, *updates.Visible)
	}
	if updates.StyleConfig != nil {
		setParts = append(setParts, "style_config = ?")
		args = append(args, *updates.StyleConfig)
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Always update the updated_at timestamp
	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")

	// Add the regionID to args for WHERE clause
	args = append(args, regionID)

	query := fmt.Sprintf(`
		UPDATE regions 
		SET %s
		WHERE id = ? AND deleted_at IS NULL
	`, strings.Join(setParts, ", "))

	result, err := a.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update region: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("region with ID '%s' not found or already deleted", regionID)
	}

	return nil
}

// SoftDeleteRegion marks a region as deleted without removing it from the database
func (a *Adapter) SoftDeleteRegion(ctx context.Context, regionID string, deletedBy int) error {
	query := `
		UPDATE regions 
		SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := a.db.ExecContext(ctx, query, deletedBy, regionID)
	if err != nil {
		return fmt.Errorf("failed to soft delete region: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("region with ID '%s' not found or already deleted", regionID)
	}

	return nil
}

// GetDeletedRegions retrieves all soft-deleted regions
func (a *Adapter) GetDeletedRegions(ctx context.Context) ([]ports.Region, error) {
	// TODO: Implement proper database query for soft-deleted regions
	return []ports.Region{}, nil
}

// RestoreRegion restores a soft-deleted region
func (a *Adapter) RestoreRegion(ctx context.Context, regionID string) error {
	// TODO: Implement proper restore logic
	return fmt.Errorf("regions database implementation not yet complete - RestoreRegion not implemented")
}

// BulkCreateRegions creates multiple regions in a single transaction
func (a *Adapter) BulkCreateRegions(ctx context.Context, newRegions []ports.NewRegion) error {
	if len(newRegions) == 0 {
		return nil
	}

	// Start a transaction for bulk insert
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare the insert statement
	query := `
		INSERT INTO regions 
		(id, slide_id, tenant_id, actor_type, actor_id, creator_id, version, name, region_type, 
		 geometry_data, coordinate_system, area_pixels, area_physical, labels, metadata, 
		 mutable, visible, style_config) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare bulk insert statement: %w", err)
	}
	defer stmt.Close()

	// Insert each region
	for i, newRegion := range newRegions {
		// Get the internal slide ID from the slide UID
		var internalSlideID int
		err := tx.QueryRowContext(ctx, "SELECT id FROM slides WHERE slide_uid = ? AND deleted_at IS NULL", newRegion.SlideUID).Scan(&internalSlideID)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("cannot create region %d: slide with UID '%s' does not exist", i, newRegion.SlideUID)
			}
			return fmt.Errorf("failed to get slide ID for region %d: %w", i, err)
		}

		// Set defaults
		version := newRegion.Version
		if version == 0 {
			version = 1
		}

		metadata := newRegion.Metadata
		if metadata == "" {
			metadata = "{}"
		}

		labels := newRegion.Labels
		if labels == "" {
			labels = "{}"
		}

		styleConfig := newRegion.StyleConfig
		if styleConfig == "" {
			styleConfig = "{}"
		}

		// Execute the insert
		_, err = stmt.ExecContext(ctx,
			newRegion.ID, internalSlideID, newRegion.TenantID, newRegion.ActorType, newRegion.ActorID,
			newRegion.CreatorID, version, newRegion.Name, newRegion.RegionType,
			newRegion.GeometryData, newRegion.CoordinateSystem, newRegion.AreaPixels, newRegion.AreaPhysical,
			labels, metadata, newRegion.Mutable, newRegion.Visible, styleConfig)
		if err != nil {
			return fmt.Errorf("failed to insert region %d: %w", i, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk insert transaction: %w", err)
	}

	return nil
}

// BulkUpdateRegions updates multiple regions in a single transaction
func (a *Adapter) BulkUpdateRegions(ctx context.Context, updates map[string]ports.UpdateRegion) error {
	if len(updates) == 0 {
		return nil
	}

	// Start a transaction for bulk update
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update each region
	for regionID, regionUpdates := range updates {
		err := a.updateRegionInTx(ctx, tx, regionID, regionUpdates)
		if err != nil {
			return fmt.Errorf("failed to update region %s: %w", regionID, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk update transaction: %w", err)
	}

	return nil
}

// Helper method to update a region within a transaction
func (a *Adapter) updateRegionInTx(ctx context.Context, tx *sql.Tx, regionID string, updates ports.UpdateRegion) error {
	// Build dynamic update query (same logic as UpdateRegion but using tx)
	setParts := []string{}
	args := []interface{}{}

	if updates.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *updates.Name)
	}
	if updates.RegionType != nil {
		setParts = append(setParts, "region_type = ?")
		args = append(args, *updates.RegionType)
	}
	if updates.GeometryData != nil {
		setParts = append(setParts, "geometry_data = ?")
		args = append(args, *updates.GeometryData)
	}
	if updates.CoordinateSystem != nil {
		setParts = append(setParts, "coordinate_system = ?")
		args = append(args, *updates.CoordinateSystem)
	}
	if updates.AreaPixels != nil {
		setParts = append(setParts, "area_pixels = ?")
		args = append(args, *updates.AreaPixels)
	}
	if updates.AreaPhysical != nil {
		setParts = append(setParts, "area_physical = ?")
		args = append(args, *updates.AreaPhysical)
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
	if updates.Visible != nil {
		setParts = append(setParts, "visible = ?")
		args = append(args, *updates.Visible)
	}
	if updates.StyleConfig != nil {
		setParts = append(setParts, "style_config = ?")
		args = append(args, *updates.StyleConfig)
	}

	if len(setParts) == 0 {
		return nil // Nothing to update
	}

	// Always update the updated_at timestamp
	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, regionID)

	query := fmt.Sprintf(`
		UPDATE regions 
		SET %s
		WHERE id = ? AND deleted_at IS NULL
	`, strings.Join(setParts, ", "))

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update region: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("region with ID '%s' not found or already deleted", regionID)
	}

	return nil
}

// BulkDeleteRegions soft-deletes multiple regions in a single transaction
func (a *Adapter) BulkDeleteRegions(ctx context.Context, regionIDs []string, deletedBy int) error {
	if len(regionIDs) == 0 {
		return nil
	}

	// Start a transaction for bulk delete
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare the soft delete statement
	query := `
		UPDATE regions 
		SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted_at IS NULL
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare bulk delete statement: %w", err)
	}
	defer stmt.Close()

	// Delete each region
	for _, regionID := range regionIDs {
		result, err := stmt.ExecContext(ctx, deletedBy, regionID)
		if err != nil {
			return fmt.Errorf("failed to delete region %s: %w", regionID, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected for region %s: %w", regionID, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("region with ID '%s' not found or already deleted", regionID)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk delete transaction: %w", err)
	}

	return nil
}
