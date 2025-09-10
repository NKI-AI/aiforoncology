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
	"strings"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
)

// RasterSearchService handles raster annotation search and listing operations
type RasterSearchService struct {
	db *sql.DB
}

// NewRasterSearchService creates a new raster search service instance
func NewRasterSearchService(db *sql.DB) *RasterSearchService {
	return &RasterSearchService{db: db}
}

// LoadAllMasks retrieves all raster annotations from the database (without pagination)
func (s *RasterSearchService) LoadAllMasks(ctx context.Context) ([]ports.Mask, error) {
	query := buildRasterAnnotationsBaseQuery()

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("raster annotations", err)
	}
	defer rows.Close()

	var masks []ports.Mask
	for rows.Next() {
		mask, err := s.scanRasterAnnotation(rows)
		if err != nil {
			return nil, err
		}
		masks = append(masks, mask)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("raster annotations", err)
	}

	return masks, nil
}

// GetRasterAnnotationsGeneric retrieves raster annotations with pagination and search support
func (s *RasterSearchService) GetRasterAnnotationsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]ports.Mask, domain.PaginationInfo, error) {
	baseQuery := buildRasterAnnotationsBaseQuery()

	// Build WHERE clause based on search parameters
	searchableFields := []string{"name", "file_uri", "slide_uid"} // Define which fields are searchable
	whereConditions, args := buildRasterAnnotationsSearchWhereClause(params.SearchParams, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	// Get total count for pagination
	totalCount, err := s.getRasterAnnotationsCount(ctx, params.SearchParams)
	if err != nil {
		return nil, domain.PaginationInfo{}, err
	}

	// Add ordering
	orderBy := "ra.created_at DESC" // Default ordering (with table alias)
	if params.SearchParams.HasSort() {
		safeSortDir := validateSortDir(params.SearchParams.SortDir)
		switch params.SearchParams.SortBy {
		case "name":
			orderBy = "ra.name " + safeSortDir
		case "created_at", "createdAt":
			orderBy = "ra.created_at " + safeSortDir
		case "slide_uid", "slideId", "slideUid":
			orderBy = "s.slide_uid " + safeSortDir
		case "file_uri", "fileUri":
			orderBy = "ra.file_uri " + safeSortDir
		case "format":
			orderBy = "ra.format " + safeSortDir
		default:
			// Keep default ordering for unknown sort fields
		}
	}
	baseQuery += " ORDER BY " + orderBy

	// Add pagination
	if params.PaginationParams.Limit > 0 {
		offset := params.PaginationParams.CalculateOffset()
		baseQuery += " LIMIT ? OFFSET ?"
		args = append(args, params.PaginationParams.Limit, offset)
	}

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, domain.PaginationInfo{}, errors.NewDatabaseQueryError("raster annotations", err)
	}
	defer rows.Close()

	var masks []ports.Mask
	for rows.Next() {
		mask, err := s.scanRasterAnnotation(rows)
		if err != nil {
			return nil, domain.PaginationInfo{}, err
		}
		masks = append(masks, mask)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.PaginationInfo{}, errors.NewDatabaseIterateRowsError("raster annotations", err)
	}

	// Calculate pagination info
	paginationInfo := utils.CreatePaginationInfo(params.PaginationParams, totalCount)

	return masks, paginationInfo, nil
}

// getRasterAnnotationsCount returns the total count of raster annotations matching search criteria
func (s *RasterSearchService) getRasterAnnotationsCount(ctx context.Context, search utils.SearchParams) (int, error) {
	baseQuery := buildRasterAnnotationsCountQuery()

	// Build WHERE clause based on search parameters (same logic as GetRasterAnnotationsGeneric)
	searchableFields := []string{"name", "file_uri", "slide_uid"} // Define which fields are searchable
	whereConditions, args := buildRasterAnnotationsSearchWhereClause(search, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	var count int
	err := s.db.QueryRow(baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, errors.NewDatabaseQueryError("raster annotation count", err)
	}
	return count, nil
}

// scanRasterAnnotation scans a row into a Mask struct
func (s *RasterSearchService) scanRasterAnnotation(rows *sql.Rows) (ports.Mask, error) {
	var mask ports.Mask
	var labels sql.NullString
	var metadata sql.NullString
	var fileHash sql.NullString
	var maskWidth, maskHeight sql.NullInt64
	var maskMpp sql.NullFloat64
	var createdAtStr string
	var deletedAtStr sql.NullString
	var deletedBy sql.NullInt64

	if err := rows.Scan(
		&mask.ID, &mask.ActorType, &mask.ActorID, &mask.CreatorID, &mask.SlideID, &mask.SlideUID, &mask.MaskUID, &mask.Version,
		&mask.Name, &mask.MaskURI, &fileHash, &mask.Format, &labels, &metadata,
		&maskWidth, &maskHeight, &maskMpp, &mask.Mutable,
		&deletedAtStr, &deletedBy, &createdAtStr,
	); err != nil {
		return ports.Mask{}, errors.NewDatabaseScanError("raster annotation", err)
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
