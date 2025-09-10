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

// VectorSearchService handles vector annotation search and listing operations
type VectorSearchService struct {
	db *sql.DB
}

// NewVectorSearchService creates a new vector search service instance
func NewVectorSearchService(db *sql.DB) *VectorSearchService {
	return &VectorSearchService{db: db}
}

// LoadAllVectorAnnotations retrieves all vector annotations from the database (without pagination)
func (s *VectorSearchService) LoadAllVectorAnnotations(ctx context.Context) ([]ports.VectorAnnotation, error) {
	query := buildVectorAnnotationsBaseQuery()

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("vector annotations", err)
	}
	defer rows.Close()

	var vectors []ports.VectorAnnotation
	for rows.Next() {
		vector, err := s.scanVectorAnnotation(rows)
		if err != nil {
			return nil, err
		}
		vectors = append(vectors, vector)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("vector annotations", err)
	}

	return vectors, nil
}

// GetVectorAnnotationsGeneric retrieves vector annotations with pagination and search support
func (s *VectorSearchService) GetVectorAnnotationsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]ports.VectorAnnotation, domain.PaginationInfo, error) {
	baseQuery := buildVectorAnnotationsBaseQuery()

	// Build WHERE clause based on search parameters
	searchableFields := []string{"name", "file_uri", "slide_uid"} // Define which fields are searchable
	whereConditions, args := buildVectorAnnotationsSearchWhereClause(params.SearchParams, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	// Get total count for pagination
	totalCount, err := s.getVectorAnnotationsCount(ctx, params.SearchParams)
	if err != nil {
		return nil, domain.PaginationInfo{}, err
	}

	// Add ordering
	orderBy := "va.created_at DESC" // Default ordering (with table alias)
	if params.SearchParams.HasSort() {
		safeSortDir := validateSortDir(params.SearchParams.SortDir)
		switch params.SearchParams.SortBy {
		case "name":
			orderBy = "va.name " + safeSortDir
		case "created_at", "createdAt":
			orderBy = "va.created_at " + safeSortDir
		case "slide_uid", "slideId", "slideUid":
			orderBy = "s.slide_uid " + safeSortDir
		case "file_uri", "fileUri":
			orderBy = "va.file_uri " + safeSortDir
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
		return nil, domain.PaginationInfo{}, errors.NewDatabaseQueryError("vector annotations", err)
	}
	defer rows.Close()

	var vectors []ports.VectorAnnotation
	for rows.Next() {
		vector, err := s.scanVectorAnnotation(rows)
		if err != nil {
			return nil, domain.PaginationInfo{}, err
		}
		vectors = append(vectors, vector)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.PaginationInfo{}, errors.NewDatabaseIterateRowsError("vector annotations", err)
	}

	// Calculate pagination info
	paginationInfo := utils.CreatePaginationInfo(params.PaginationParams, totalCount)

	return vectors, paginationInfo, nil
}

// getVectorAnnotationsCount returns the total count of vector annotations matching search criteria
func (s *VectorSearchService) getVectorAnnotationsCount(ctx context.Context, search utils.SearchParams) (int, error) {
	baseQuery := buildVectorAnnotationsCountQuery()

	// Build WHERE clause based on search parameters (same logic as GetVectorAnnotationsGeneric)
	searchableFields := []string{"name", "file_uri", "slide_uid"} // Define which fields are searchable
	whereConditions, args := buildVectorAnnotationsSearchWhereClause(search, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	var count int
	err := s.db.QueryRow(baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, errors.NewDatabaseQueryError("vector annotation count", err)
	}
	return count, nil
}

// scanVectorAnnotation scans a row into a VectorAnnotation struct
func (s *VectorSearchService) scanVectorAnnotation(rows *sql.Rows) (ports.VectorAnnotation, error) {
	var vector ports.VectorAnnotation
	var createdAtStr string
	var updatedAtStr string
	var deletedAtStr sql.NullString
	var deletedBy sql.NullInt64
	var labels sql.NullString
	var metadata sql.NullString
	var dataBlob sql.NullString

	if err := rows.Scan(
		&vector.ID, &vector.ActorType, &vector.ActorID, &vector.CreatorID, &vector.SlideID, &vector.SlideUID, &vector.VectorUID, &vector.Version,
		&vector.Name, &vector.FileURI, &dataBlob, &labels, &metadata, &vector.Mutable,
		&deletedAtStr, &deletedBy, &createdAtStr, &updatedAtStr,
	); err != nil {
		return ports.VectorAnnotation{}, errors.NewDatabaseScanError("vector annotation", err)
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
