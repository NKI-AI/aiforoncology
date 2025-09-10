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
	"strings"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
)

// SearchService handles study search and listing operations
type SearchService struct {
	db *sql.DB
}

// NewSearchService creates a new search service instance
func NewSearchService(db *sql.DB) *SearchService {
	return &SearchService{db: db}
}

// LoadAllStudies retrieves studies from the database with search/filter and pagination support
func (s *SearchService) LoadAllStudies(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Study, error) {
	baseQuery := buildStudiesBaseQuery()

	// Build WHERE clause based on search parameters
	searchableFields := []string{"name", "description", "metadata"} // Define which fields are searchable
	whereConditions, args := buildStudiesSearchWhereClause(search, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	// Add ordering
	orderBy := "s.created_at DESC" // Default ordering (with table alias)
	if search.HasSort() {
		safeSortDir := validateSortDir(search.SortDir)
		switch search.SortBy {
		case "name":
			orderBy = "s.name " + safeSortDir
		case "created_at", "createdAt":
			orderBy = "s.created_at " + safeSortDir
		case "short_uid", "shortId", "studyUid":
			orderBy = "s.short_uid " + safeSortDir
		default:
			// Keep default ordering for unknown sort fields
		}
	}
	baseQuery += " ORDER BY " + orderBy

	// Add pagination
	if pagination.Limit > 0 {
		offset := pagination.CalculateOffset()
		baseQuery += " LIMIT ? OFFSET ?"
		args = append(args, pagination.Limit, offset)
	}

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("studies", err)
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
			return nil, errors.NewDatabaseScanError("study", err)
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
		return nil, errors.NewDatabaseIterateRowsError("studies", err)
	}

	return studies, nil
}

// GetStudiesCount returns the total count of studies matching search criteria
func (s *SearchService) GetStudiesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	baseQuery := buildStudiesCountQuery()

	// Build WHERE clause based on search parameters (same logic as LoadAllStudies)
	searchableFields := []string{"name", "description", "metadata"} // Define which fields are searchable
	whereConditions, args := buildStudiesSearchWhereClause(search, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	var count int
	err := s.db.QueryRow(baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, errors.NewDatabaseQueryError("study count", err)
	}
	return count, nil
}

// LoadStudiesByIDs retrieves studies filtered by a list of IDs with search/filter and pagination support
func (s *SearchService) LoadStudiesByIDs(ctx context.Context, studyIDs []int, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Study, error) {
	if len(studyIDs) == 0 {
		return []ports.Study{}, nil
	}

	baseQuery := buildStudiesBaseQuery()

	// Build WHERE clause based on search parameters
	searchableFields := []string{"name", "description", "metadata"}
	whereConditions, args := buildStudiesSearchWhereClause(search, searchableFields)

	// Add ID filtering
	placeholders := make([]string, len(studyIDs))
	for i, id := range studyIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	whereConditions = append(whereConditions, "s.id IN ("+strings.Join(placeholders, ",")+")")

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	// Add ordering
	orderBy := "s.created_at DESC" // Default ordering (with table alias)
	if search.HasSort() {
		safeSortDir := validateSortDir(search.SortDir)
		switch search.SortBy {
		case "name":
			orderBy = "s.name " + safeSortDir
		case "created_at", "createdAt":
			orderBy = "s.created_at " + safeSortDir
		case "short_uid", "shortId", "studyUid":
			orderBy = "s.short_uid " + safeSortDir
		default:
			// Keep default ordering for unknown sort fields
		}
	}
	baseQuery += " ORDER BY " + orderBy

	// Add pagination
	if pagination.Limit > 0 {
		offset := pagination.CalculateOffset()
		baseQuery += " LIMIT ? OFFSET ?"
		args = append(args, pagination.Limit, offset)
	}

	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("studies", err)
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
			return nil, errors.NewDatabaseScanError("study", err)
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
		return nil, errors.NewDatabaseIterateRowsError("studies", err)
	}

	log.Info("LoadStudiesByIDs completed", "filtered_ids_count", len(studyIDs), "results_count", len(studies))
	return studies, nil
}

// CountStudiesByIDs returns the total count of studies in the ID list matching search criteria
func (s *SearchService) CountStudiesByIDs(ctx context.Context, studyIDs []int, search utils.SearchParams) (int, error) {
	if len(studyIDs) == 0 {
		return 0, nil
	}

	baseQuery := buildStudiesCountQuery()

	// Build WHERE clause based on search parameters
	searchableFields := []string{"name", "description", "metadata"}
	whereConditions, args := buildStudiesSearchWhereClause(search, searchableFields)

	// Add ID filtering
	placeholders := make([]string, len(studyIDs))
	for i, id := range studyIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	whereConditions = append(whereConditions, "s.id IN ("+strings.Join(placeholders, ",")+")")

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	var count int
	err := s.db.QueryRowContext(ctx, baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, errors.NewDatabaseQueryError("studies count", err)
	}

	return count, nil
}
