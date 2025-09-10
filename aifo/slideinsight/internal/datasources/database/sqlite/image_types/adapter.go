// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package image_types

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// ImageTypesAdapter provides image types operations
type ImageTypesAdapter struct {
	db *sql.DB
}

// NewImageTypesAdapter creates a new image types adapter
func NewImageTypesAdapter(db *sql.DB) *ImageTypesAdapter {
	return &ImageTypesAdapter{db: db}
}

// LoadAllImageTypes retrieves image types from the database with search/filter and pagination support
func (a *ImageTypesAdapter) LoadAllImageTypes(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.ImageType, error) {
	whereConditions, args := a.buildImageTypesSearchWhereClause(search)

	query := `
		SELECT id, tenant_id, type_uid, name, description, category, requires_histogram, metadata_schema, is_active, created_at, updated_at
		FROM image_types`

	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	query += " ORDER BY name ASC"

	// Add pagination
	limit := pagination.Limit
	offset := (pagination.Page - 1) * pagination.Limit
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query image types: %w", err)
	}
	defer rows.Close()

	var imageTypes []ports.ImageType
	for rows.Next() {
		var it ports.ImageType
		var createdAt, updatedAt time.Time
		var description, metadataSchema sql.NullString

		err := rows.Scan(
			&it.ID, &it.TenantID, &it.TypeUID, &it.Name, &description, &it.Category,
			&it.RequiresHistogram, &metadataSchema, &it.IsActive, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image type: %w", err)
		}

		if description.Valid {
			it.Description = description.String
		}
		if metadataSchema.Valid {
			it.MetadataSchema = metadataSchema.String
		}
		it.CreatedAt = createdAt
		it.UpdatedAt = updatedAt

		imageTypes = append(imageTypes, it)
	}

	return imageTypes, nil
}

// GetImageTypesCount returns the total count of image types matching search criteria
func (a *ImageTypesAdapter) GetImageTypesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	whereConditions, args := a.buildImageTypesSearchWhereClause(search)

	query := "SELECT COUNT(*) FROM image_types"
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	var count int
	err := a.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count image types: %w", err)
	}

	return count, nil
}

// GetImageTypeByID retrieves a specific image type by its ID
func (a *ImageTypesAdapter) GetImageTypeByID(ctx context.Context, id string) (ports.ImageType, error) {
	query := `
		SELECT id, tenant_id, type_uid, name, description, category, requires_histogram, metadata_schema, is_active, created_at, updated_at
		FROM image_types WHERE id = ?`

	var it ports.ImageType
	var createdAt, updatedAt time.Time
	var description, metadataSchema sql.NullString

	err := a.db.QueryRowContext(ctx, query, id).Scan(
		&it.ID, &it.TenantID, &it.TypeUID, &it.Name, &description, &it.Category,
		&it.RequiresHistogram, &metadataSchema, &it.IsActive, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.ImageType{}, fmt.Errorf("image type not found")
		}
		return ports.ImageType{}, fmt.Errorf("failed to get image type: %w", err)
	}

	if description.Valid {
		it.Description = description.String
	}
	if metadataSchema.Valid {
		it.MetadataSchema = metadataSchema.String
	}
	it.CreatedAt = createdAt
	it.UpdatedAt = updatedAt

	return it, nil
}

// CreateImageType adds a new image type to the database
func (a *ImageTypesAdapter) CreateImageType(ctx context.Context, imageType ports.NewImageType) error {
	query := `
		INSERT INTO image_types (id, tenant_id, type_uid, name, description, category, requires_histogram, metadata_schema, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := a.db.ExecContext(ctx, query,
		imageType.ID, imageType.TenantID, imageType.TypeUID, imageType.Name,
		imageType.Description, imageType.Category, imageType.RequiresHistogram,
		imageType.MetadataSchema, imageType.IsActive,
	)
	if err != nil {
		return fmt.Errorf("failed to create image type: %w", err)
	}

	return nil
}

// UpdateImageType updates an existing image type
func (a *ImageTypesAdapter) UpdateImageType(ctx context.Context, id string, updates ports.ImageTypeUpdates) error {
	setParts := []string{}
	args := []interface{}{}

	if updates.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *updates.Name)
	}
	if updates.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *updates.Description)
	}
	if updates.Category != nil {
		setParts = append(setParts, "category = ?")
		args = append(args, *updates.Category)
	}
	if updates.RequiresHistogram != nil {
		setParts = append(setParts, "requires_histogram = ?")
		args = append(args, *updates.RequiresHistogram)
	}
	if updates.MetadataSchema != nil {
		setParts = append(setParts, "metadata_schema = ?")
		args = append(args, *updates.MetadataSchema)
	}
	if updates.IsActive != nil {
		setParts = append(setParts, "is_active = ?")
		args = append(args, *updates.IsActive)
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := "UPDATE image_types SET " + strings.Join(setParts, ", ") + " WHERE id = ?"

	_, err := a.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update image type: %w", err)
	}

	return nil
}

// DeleteImageType removes an image type from the database
func (a *ImageTypesAdapter) DeleteImageType(ctx context.Context, id string) error {
	query := "DELETE FROM image_types WHERE id = ?"

	_, err := a.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete image type: %w", err)
	}

	return nil
}

// ImageTypeExists checks if an image type with the given ID already exists
func (a *ImageTypesAdapter) ImageTypeExists(ctx context.Context, id string) (bool, error) {
	query := "SELECT COUNT(*) FROM image_types WHERE id = ?"

	var count int
	err := a.db.QueryRowContext(ctx, query, id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check image type existence: %w", err)
	}

	return count > 0, nil
}

// buildImageTypesSearchWhereClause builds WHERE conditions and arguments for image types search queries
func (a *ImageTypesAdapter) buildImageTypesSearchWhereClause(search utils.SearchParams) ([]string, []interface{}) {
	var whereConditions []string
	var args []interface{}

	// General search across name, type_uid, and description
	if search.Query != "" {
		whereConditions = append(whereConditions, "(name LIKE ? OR type_uid LIKE ? OR description LIKE ?)")
		searchTerm := "%" + search.Query + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	// Add specific field filters
	for field, value := range search.Filters {
		switch field {
		case "category":
			whereConditions = append(whereConditions, "category = ?")
			args = append(args, value)
		case "is_active":
			whereConditions = append(whereConditions, "is_active = ?")
			args = append(args, value)
		case "requires_histogram":
			whereConditions = append(whereConditions, "requires_histogram = ?")
			args = append(args, value)
		}
	}

	return whereConditions, args
}

// SlideHistogramsAdapter provides slide histograms operations
type SlideHistogramsAdapter struct {
	db *sql.DB
}

// NewSlideHistogramsAdapter creates a new slide histograms adapter
func NewSlideHistogramsAdapter(db *sql.DB) *SlideHistogramsAdapter {
	return &SlideHistogramsAdapter{db: db}
}

// GetHistogramsBySlideUID retrieves all histograms for a given slide
func (a *SlideHistogramsAdapter) GetHistogramsBySlideUID(ctx context.Context, slideUID string) ([]ports.SlideHistogram, error) {
	// First get the slide's internal ID
	var slideID int
	err := a.db.QueryRowContext(ctx, "SELECT id FROM slides WHERE slide_uid = ?", slideUID).Scan(&slideID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("slide not found")
		}
		return nil, fmt.Errorf("failed to get slide ID: %w", err)
	}

	query := `
		SELECT id, slide_id, channel_index, channel_name, bin_count, min_value, max_value, histogram_data, metadata, created_at, updated_at
		FROM slide_histograms WHERE slide_id = ? ORDER BY channel_index`

	rows, err := a.db.QueryContext(ctx, query, slideID)
	if err != nil {
		return nil, fmt.Errorf("failed to query slide histograms: %w", err)
	}
	defer rows.Close()

	var histograms []ports.SlideHistogram
	for rows.Next() {
		var h ports.SlideHistogram
		var createdAt, updatedAt time.Time
		var channelName, metadata sql.NullString

		err := rows.Scan(
			&h.ID, &h.SlideID, &h.ChannelIndex, &channelName, &h.BinCount,
			&h.MinValue, &h.MaxValue, &h.HistogramData, &metadata, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan slide histogram: %w", err)
		}

		if channelName.Valid {
			h.ChannelName = channelName.String
		}
		if metadata.Valid {
			h.Metadata = metadata.String
		}
		h.CreatedAt = createdAt
		h.UpdatedAt = updatedAt

		histograms = append(histograms, h)
	}

	return histograms, nil
}

// GetHistogramByID retrieves a specific histogram by its ID
func (a *SlideHistogramsAdapter) GetHistogramByID(ctx context.Context, id string) (ports.SlideHistogram, error) {
	query := `
		SELECT id, slide_id, channel_index, channel_name, bin_count, min_value, max_value, histogram_data, metadata, created_at, updated_at
		FROM slide_histograms WHERE id = ?`

	var h ports.SlideHistogram
	var createdAt, updatedAt time.Time
	var channelName, metadata sql.NullString

	err := a.db.QueryRowContext(ctx, query, id).Scan(
		&h.ID, &h.SlideID, &h.ChannelIndex, &channelName, &h.BinCount,
		&h.MinValue, &h.MaxValue, &h.HistogramData, &metadata, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.SlideHistogram{}, fmt.Errorf("slide histogram not found")
		}
		return ports.SlideHistogram{}, fmt.Errorf("failed to get slide histogram: %w", err)
	}

	if channelName.Valid {
		h.ChannelName = channelName.String
	}
	if metadata.Valid {
		h.Metadata = metadata.String
	}
	h.CreatedAt = createdAt
	h.UpdatedAt = updatedAt

	return h, nil
}

// CreateHistogram adds a new histogram to the database
func (a *SlideHistogramsAdapter) CreateHistogram(ctx context.Context, histogram ports.NewSlideHistogram) error {
	// First get the slide's internal ID
	var slideID int
	err := a.db.QueryRowContext(ctx, "SELECT id FROM slides WHERE slide_uid = ?", histogram.SlideUID).Scan(&slideID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("slide not found")
		}
		return fmt.Errorf("failed to get slide ID: %w", err)
	}

	query := `
		INSERT INTO slide_histograms (id, slide_id, channel_index, channel_name, bin_count, min_value, max_value, histogram_data, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = a.db.ExecContext(ctx, query,
		histogram.ID, slideID, histogram.ChannelIndex, histogram.ChannelName,
		histogram.BinCount, histogram.MinValue, histogram.MaxValue, histogram.HistogramData, histogram.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create slide histogram: %w", err)
	}

	return nil
}

// UpdateHistogram updates an existing histogram
func (a *SlideHistogramsAdapter) UpdateHistogram(ctx context.Context, id string, histogram ports.NewSlideHistogram) error {
	query := `
		UPDATE slide_histograms 
		SET channel_index = ?, channel_name = ?, bin_count = ?, min_value = ?, max_value = ?, histogram_data = ?, metadata = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	_, err := a.db.ExecContext(ctx, query,
		histogram.ChannelIndex, histogram.ChannelName, histogram.BinCount,
		histogram.MinValue, histogram.MaxValue, histogram.HistogramData, histogram.Metadata, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update slide histogram: %w", err)
	}

	return nil
}

// DeleteHistogram removes a histogram from the database
func (a *SlideHistogramsAdapter) DeleteHistogram(ctx context.Context, id string) error {
	query := "DELETE FROM slide_histograms WHERE id = ?"

	_, err := a.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete slide histogram: %w", err)
	}

	return nil
}

// DeleteHistogramsBySlideUID removes all histograms for a given slide
func (a *SlideHistogramsAdapter) DeleteHistogramsBySlideUID(ctx context.Context, slideUID string) error {
	// First get the slide's internal ID
	var slideID int
	err := a.db.QueryRowContext(ctx, "SELECT id FROM slides WHERE slide_uid = ?", slideUID).Scan(&slideID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("slide not found")
		}
		return fmt.Errorf("failed to get slide ID: %w", err)
	}

	query := "DELETE FROM slide_histograms WHERE slide_id = ?"

	_, err = a.db.ExecContext(ctx, query, slideID)
	if err != nil {
		return fmt.Errorf("failed to delete slide histograms: %w", err)
	}

	return nil
}

// StainingProtocolsAdapter provides staining protocols operations
type StainingProtocolsAdapter struct {
	db *sql.DB
}

// NewStainingProtocolsAdapter creates a new staining protocols adapter
func NewStainingProtocolsAdapter(db *sql.DB) *StainingProtocolsAdapter {
	return &StainingProtocolsAdapter{db: db}
}

// GetProtocolsBySlideUID retrieves all staining protocols for a given slide
func (a *StainingProtocolsAdapter) GetProtocolsBySlideUID(ctx context.Context, slideUID string) ([]ports.StainingProtocol, error) {
	// First get the slide's internal ID
	var slideID int
	err := a.db.QueryRowContext(ctx, "SELECT id FROM slides WHERE slide_uid = ?", slideUID).Scan(&slideID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("slide not found")
		}
		return nil, fmt.Errorf("failed to get slide ID: %w", err)
	}

	query := `
		SELECT id, slide_id, stain_name, stain_type, concentration, incubation_time, antibody_info, excitation_nm, emission_nm, metadata, created_at, updated_at
		FROM slide_staining_protocols WHERE slide_id = ? ORDER BY stain_name`

	rows, err := a.db.QueryContext(ctx, query, slideID)
	if err != nil {
		return nil, fmt.Errorf("failed to query staining protocols: %w", err)
	}
	defer rows.Close()

	var protocols []ports.StainingProtocol
	for rows.Next() {
		var p ports.StainingProtocol
		var createdAt, updatedAt time.Time
		var concentration, incubationTime, antibodyInfo, metadata sql.NullString
		var excitationNm, emissionNm sql.NullInt32

		err := rows.Scan(
			&p.ID, &p.SlideID, &p.StainName, &p.StainType, &concentration, &incubationTime,
			&antibodyInfo, &excitationNm, &emissionNm, &metadata, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan staining protocol: %w", err)
		}

		if concentration.Valid {
			p.Concentration = concentration.String
		}
		if incubationTime.Valid {
			p.IncubationTime = incubationTime.String
		}
		if antibodyInfo.Valid {
			p.AntibodyInfo = antibodyInfo.String
		}
		if excitationNm.Valid {
			excitation := int(excitationNm.Int32)
			p.ExcitationNm = &excitation
		}
		if emissionNm.Valid {
			emission := int(emissionNm.Int32)
			p.EmissionNm = &emission
		}
		if metadata.Valid {
			p.Metadata = metadata.String
		}
		p.CreatedAt = createdAt
		p.UpdatedAt = updatedAt

		protocols = append(protocols, p)
	}

	return protocols, nil
}

// GetProtocolByID retrieves a specific staining protocol by its ID
func (a *StainingProtocolsAdapter) GetProtocolByID(ctx context.Context, id string) (ports.StainingProtocol, error) {
	query := `
		SELECT id, slide_id, stain_name, stain_type, concentration, incubation_time, antibody_info, excitation_nm, emission_nm, metadata, created_at, updated_at
		FROM slide_staining_protocols WHERE id = ?`

	var p ports.StainingProtocol
	var createdAt, updatedAt time.Time
	var concentration, incubationTime, antibodyInfo, metadata sql.NullString
	var excitationNm, emissionNm sql.NullInt32

	err := a.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.SlideID, &p.StainName, &p.StainType, &concentration, &incubationTime,
		&antibodyInfo, &excitationNm, &emissionNm, &metadata, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.StainingProtocol{}, fmt.Errorf("staining protocol not found")
		}
		return ports.StainingProtocol{}, fmt.Errorf("failed to get staining protocol: %w", err)
	}

	if concentration.Valid {
		p.Concentration = concentration.String
	}
	if incubationTime.Valid {
		p.IncubationTime = incubationTime.String
	}
	if antibodyInfo.Valid {
		p.AntibodyInfo = antibodyInfo.String
	}
	if excitationNm.Valid {
		excitation := int(excitationNm.Int32)
		p.ExcitationNm = &excitation
	}
	if emissionNm.Valid {
		emission := int(emissionNm.Int32)
		p.EmissionNm = &emission
	}
	if metadata.Valid {
		p.Metadata = metadata.String
	}
	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt

	return p, nil
}

// CreateProtocol adds a new staining protocol to the database
func (a *StainingProtocolsAdapter) CreateProtocol(ctx context.Context, protocol ports.NewStainingProtocol) error {
	// First get the slide's internal ID
	var slideID int
	err := a.db.QueryRowContext(ctx, "SELECT id FROM slides WHERE slide_uid = ?", protocol.SlideUID).Scan(&slideID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("slide not found")
		}
		return fmt.Errorf("failed to get slide ID: %w", err)
	}

	query := `
		INSERT INTO slide_staining_protocols (id, slide_id, stain_name, stain_type, concentration, incubation_time, antibody_info, excitation_nm, emission_nm, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = a.db.ExecContext(ctx, query,
		protocol.ID, slideID, protocol.StainName, protocol.StainType,
		protocol.Concentration, protocol.IncubationTime, protocol.AntibodyInfo,
		protocol.ExcitationNm, protocol.EmissionNm, protocol.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create staining protocol: %w", err)
	}

	return nil
}

// UpdateProtocol updates an existing staining protocol
func (a *StainingProtocolsAdapter) UpdateProtocol(ctx context.Context, id string, updates ports.StainingProtocolUpdates) error {
	setParts := []string{}
	args := []interface{}{}

	if updates.StainName != nil {
		setParts = append(setParts, "stain_name = ?")
		args = append(args, *updates.StainName)
	}
	if updates.StainType != nil {
		setParts = append(setParts, "stain_type = ?")
		args = append(args, *updates.StainType)
	}
	if updates.Concentration != nil {
		setParts = append(setParts, "concentration = ?")
		args = append(args, *updates.Concentration)
	}
	if updates.IncubationTime != nil {
		setParts = append(setParts, "incubation_time = ?")
		args = append(args, *updates.IncubationTime)
	}
	if updates.AntibodyInfo != nil {
		setParts = append(setParts, "antibody_info = ?")
		args = append(args, *updates.AntibodyInfo)
	}
	if updates.ExcitationNm != nil {
		setParts = append(setParts, "excitation_nm = ?")
		args = append(args, *updates.ExcitationNm)
	}
	if updates.EmissionNm != nil {
		setParts = append(setParts, "emission_nm = ?")
		args = append(args, *updates.EmissionNm)
	}
	if updates.Metadata != nil {
		setParts = append(setParts, "metadata = ?")
		args = append(args, *updates.Metadata)
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := "UPDATE slide_staining_protocols SET " + strings.Join(setParts, ", ") + " WHERE id = ?"

	_, err := a.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update staining protocol: %w", err)
	}

	return nil
}

// DeleteProtocol removes a staining protocol from the database
func (a *StainingProtocolsAdapter) DeleteProtocol(ctx context.Context, id string) error {
	query := "DELETE FROM slide_staining_protocols WHERE id = ?"

	_, err := a.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete staining protocol: %w", err)
	}

	return nil
}
