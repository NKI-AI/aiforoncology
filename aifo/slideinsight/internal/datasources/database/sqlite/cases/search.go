// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package cases

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
)

// SearchService handles search and filtering operations for cases
type SearchService struct {
	db *sql.DB
}

// NewSearchService creates a new search service instance
func NewSearchService(db *sql.DB) *SearchService {
	return &SearchService{db: db}
}

// validateSortDir ensures sort direction is valid to prevent SQL injection
func validateSortDir(sortDir string) string {
	dir := strings.ToUpper(sortDir)
	if dir != "ASC" && dir != "DESC" {
		return "ASC"
	}
	return dir
}

// LoadAllCases retrieves cases from the database with search/filter and pagination support
func (s *SearchService) LoadAllCases(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Case, error) {
	baseQuery := buildCasesBaseQuery()

	// Build WHERE clause based on search parameters
	searchableFields := []string{"name", "metadata"} // Define which fields are searchable
	whereConditions, args := buildSearchWhereClause(search, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	// Add ordering with safe sort direction validation
	orderBy := "c.created_at DESC" // Default ordering (with table alias)
	if search.HasSort() {
		safeSortDir := validateSortDir(search.SortDir)
		switch search.SortBy {
		case "name":
			orderBy = "c.name " + safeSortDir
		case "created_at", "createdAt":
			orderBy = "c.created_at " + safeSortDir
		case "short_uid", "shortId", "caseUid":
			orderBy = "c.short_uid " + safeSortDir
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
		return nil, fmt.Errorf("failed to query cases table: %w", err)
	}
	defer rows.Close()

	var cases []ports.Case
	for rows.Next() {
		var case_ ports.Case
		var createdAtStr string
		var updatedAtStr string
		var deletedAtStr sql.NullString
		var deletedBy sql.NullInt64
		if err := rows.Scan(
			&case_.ID,
			&case_.TenantID,
			&case_.TenantUID,
			&case_.CaseUID,
			&case_.CreatorID,
			&case_.CreatorUID,
			&case_.Name,
			&case_.Metadata,
			&deletedAtStr,
			&deletedBy,
			&createdAtStr,
			&updatedAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan case row: %w", err)
		}

		// Parse the created_at timestamp
		if createdAtStr != "" {
			createdAt, err := time.Parse(time.RFC3339, createdAtStr)
			if err != nil {
				log.Error("failed to parse created_at timestamp", "error", err)
			} else {
				case_.CreatedAt = createdAt
			}
		}

		// Parse the updated_at timestamp
		if updatedAtStr != "" {
			updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
			if err != nil {
				log.Error("failed to parse updated_at timestamp", "error", err)
			} else {
				case_.UpdatedAt = updatedAt
			}
		}

		// Handle soft deletion fields
		if deletedAtStr.Valid {
			deletedAt, err := time.Parse(time.RFC3339, deletedAtStr.String)
			if err != nil {
				log.Error("failed to parse deleted_at timestamp", "error", err)
			} else {
				case_.DeletedAt = &deletedAt
			}
		}

		if deletedBy.Valid {
			deletedByInt := int(deletedBy.Int64)
			case_.DeletedBy = &deletedByInt
		}

		cases = append(cases, case_)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over case rows: %w", err)
	}

	return cases, nil
}

// GetCasesCount returns the total count of cases matching search criteria
func (s *SearchService) GetCasesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	baseQuery := buildCasesCountQuery()

	// Build WHERE clause based on search parameters (same logic as LoadAllCases)
	searchableFields := []string{"name", "metadata"} // Define which fields are searchable
	whereConditions, args := buildSearchWhereClause(search, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	var count int
	err := s.db.QueryRow(baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count cases with search: %w", err)
	}
	return count, nil
}

// GetCasesByStudyUID retrieves cases by study ID with search/filter and pagination support
func (s *SearchService) GetCasesByStudyUID(ctx context.Context, studyUID string, params utils.PaginationAndSearchParams) ([]ports.Case, error) {
	// Set default values if not provided
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 100 // Default limit
	}

	baseQuery := buildCasesByStudyBaseQuery()

	// Start with study ID as first argument
	args := []interface{}{studyUID}

	// Build WHERE clause based on search parameters
	searchableFields := []string{"name", "metadata"} // Define which fields are searchable
	whereConditions, searchArgs := buildSearchWhereClause(params.SearchParams, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
		args = append(args, searchArgs...)
	}

	// Add ordering with safe sort direction validation
	orderBy := "c.created_at DESC" // Default ordering
	if params.SearchParams.HasSort() {
		safeSortDir := validateSortDir(params.SearchParams.SortDir)
		switch params.SearchParams.SortBy {
		case "name":
			orderBy = "c.name " + safeSortDir
		case "created_at", "createdAt":
			orderBy = "c.created_at " + safeSortDir
		case "short_uid", "shortId", "caseUid":
			orderBy = "c.short_uid " + safeSortDir
		default:
			// Keep default ordering for unknown sort fields
		}
	}
	baseQuery += " ORDER BY " + orderBy

	// Add pagination using utility method
	if params.Limit > 0 {
		offset := params.CalculateOffset()
		baseQuery += " LIMIT ? OFFSET ?"
		args = append(args, params.Limit, offset)
	}

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query cases by study ID with search: %w", err)
	}
	defer rows.Close()

	var cases []ports.Case
	for rows.Next() {
		var case_ ports.Case
		var createdAtStr string
		var updatedAtStr string
		var deletedAtStr sql.NullString
		var deletedBy sql.NullInt64
		if err := rows.Scan(
			&case_.ID,
			&case_.TenantID,
			&case_.TenantUID,
			&case_.CaseUID,
			&case_.CreatorID,
			&case_.CreatorUID,
			&case_.Name,
			&case_.Metadata,
			&deletedAtStr,
			&deletedBy,
			&createdAtStr,
			&updatedAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan case row: %w", err)
		}

		// Parse the created_at timestamp
		if createdAtStr != "" {
			createdAt, err := time.Parse(time.RFC3339, createdAtStr)
			if err != nil {
				log.Error("failed to parse created_at timestamp", "error", err)
			} else {
				case_.CreatedAt = createdAt
			}
		}

		// Parse the updated_at timestamp
		if updatedAtStr != "" {
			updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
			if err != nil {
				log.Error("failed to parse updated_at timestamp", "error", err)
			} else {
				case_.UpdatedAt = updatedAt
			}
		}

		// Handle soft deletion fields
		if deletedAtStr.Valid {
			deletedAt, err := time.Parse(time.RFC3339, deletedAtStr.String)
			if err != nil {
				log.Error("failed to parse deleted_at timestamp", "error", err)
			} else {
				case_.DeletedAt = &deletedAt
			}
		}

		if deletedBy.Valid {
			deletedByInt := int(deletedBy.Int64)
			case_.DeletedBy = &deletedByInt
		}

		cases = append(cases, case_)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over case rows: %w", err)
	}

	return cases, nil
}

// GetCasesByStudyUIDCount returns the total count of cases for a study matching search criteria
func (s *SearchService) GetCasesByStudyUIDCount(ctx context.Context, studyUID string, search utils.SearchParams) (int, error) {
	baseQuery := buildCasesByStudyCountQuery()

	// Start with study ID as first argument
	args := []interface{}{studyUID}

	// Build WHERE clause based on search parameters (same logic as GetCasesByStudyUID)
	searchableFields := []string{"name", "metadata"} // Define which fields are searchable
	whereConditions, searchArgs := buildSearchWhereClause(search, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
		args = append(args, searchArgs...)
	}

	var count int
	err := s.db.QueryRow(baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count cases by study ID with search: %w", err)
	}
	return count, nil
}
