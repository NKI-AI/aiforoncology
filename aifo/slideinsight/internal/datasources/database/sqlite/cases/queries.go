// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package cases

import (
	"strings"

	"aifo.dev/aifo/slideinsight/internal/utils"
)

// buildSearchWhereClause builds WHERE conditions and arguments for search queries
func buildSearchWhereClause(search utils.SearchParams, searchableFields []string) ([]string, []interface{}) {
	var whereConditions []string
	var args []interface{}

	// General search across specified searchable fields
	if search.Query != "" {
		// Build LIKE conditions for each searchable field
		var fieldConditions []string
		for _, field := range searchableFields {
			switch field {
			case "name":
				fieldConditions = append(fieldConditions, "c.name LIKE ?")
				args = append(args, "%"+search.Query+"%")
			case "metadata":
				fieldConditions = append(fieldConditions, "CAST(c.metadata AS TEXT) LIKE ?")
				args = append(args, "%"+search.Query+"%")
			}
		}
		if len(fieldConditions) > 0 {
			whereConditions = append(whereConditions, "("+strings.Join(fieldConditions, " OR ")+")")
		}
	}

	// Add specific field filters
	for field, value := range search.Filters {
		switch field {
		case "name":
			whereConditions = append(whereConditions, "c.name LIKE ?")
			args = append(args, "%"+value+"%")
		case "status":
			// Note: status field would need to be added to cases table schema
			// For now, we'll skip unknown fields gracefully
			continue
		case "has_vector_annotations":
			if value == "true" {
				whereConditions = append(whereConditions, "va.id IS NOT NULL")
			} else if value == "false" {
				whereConditions = append(whereConditions, "va.id IS NULL")
			}
		case "has_raster_annotations":
			if value == "true" {
				whereConditions = append(whereConditions, "ra.id IS NOT NULL")
			} else if value == "false" {
				whereConditions = append(whereConditions, "ra.id IS NULL")
			}
		case "tenant_id":
			whereConditions = append(whereConditions, "c.tenant_id = ?")
			args = append(args, value)
		default:
			// Skip unknown filter fields
			continue
		}
	}

	return whereConditions, args
}

// buildCasesBaseQuery returns the base query for loading cases with joins
func buildCasesBaseQuery() string {
	return `
		SELECT DISTINCT c.id, c.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, c.short_uid as case_uid, c.creator_id, COALESCE(u.short_uid, '') as creator_uid, c.name, c.metadata, c.deleted_at, c.deleted_by, c.created_at, c.updated_at 
		FROM cases c
		LEFT JOIN tenants t ON c.tenant_id = t.id
		LEFT JOIN users u ON c.creator_id = u.id
		LEFT JOIN slides s ON c.id = s.case_id AND s.deleted_at IS NULL
		LEFT JOIN vector_annotations va ON s.id = va.slide_id AND va.deleted_at IS NULL
		LEFT JOIN raster_annotations ra ON s.id = ra.slide_id AND ra.deleted_at IS NULL
		WHERE c.deleted_at IS NULL`
}

// buildCasesByStudyBaseQuery returns the base query for loading cases by study ID
func buildCasesByStudyBaseQuery() string {
	return `
		SELECT DISTINCT c.id, c.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, c.short_uid as case_uid, c.creator_id, COALESCE(u.short_uid, '') as creator_uid, c.name, c.metadata, c.deleted_at, c.deleted_by, c.created_at, c.updated_at 
		FROM cases c
		LEFT JOIN tenants t ON c.tenant_id = t.id
		LEFT JOIN users u ON c.creator_id = u.id
		INNER JOIN study_cases sc ON c.id = sc.case_id
		INNER JOIN studies s ON sc.study_id = s.id
		LEFT JOIN slides sl ON c.id = sl.case_id AND sl.deleted_at IS NULL
		LEFT JOIN vector_annotations va ON sl.id = va.slide_id AND va.deleted_at IS NULL
		LEFT JOIN raster_annotations ra ON sl.id = ra.slide_id AND ra.deleted_at IS NULL
		WHERE s.short_uid = ? AND c.deleted_at IS NULL`
}

// buildCasesCountQuery returns the base query for counting cases
func buildCasesCountQuery() string {
	return `
		SELECT COUNT(DISTINCT c.id) 
		FROM cases c
		LEFT JOIN tenants t ON c.tenant_id = t.id
		LEFT JOIN users u ON c.creator_id = u.id
		LEFT JOIN slides s ON c.id = s.case_id AND s.deleted_at IS NULL
		LEFT JOIN vector_annotations va ON s.id = va.slide_id AND va.deleted_at IS NULL
		LEFT JOIN raster_annotations ra ON s.id = ra.slide_id AND ra.deleted_at IS NULL
		WHERE c.deleted_at IS NULL`
}

// buildCasesByStudyCountQuery returns the base query for counting cases by study
func buildCasesByStudyCountQuery() string {
	return `
		SELECT COUNT(DISTINCT c.id) 
		FROM cases c
		INNER JOIN study_cases sc ON c.id = sc.case_id
		INNER JOIN studies s ON sc.study_id = s.id
		LEFT JOIN slides sl ON c.id = sl.case_id AND sl.deleted_at IS NULL
		LEFT JOIN vector_annotations va ON sl.id = va.slide_id AND va.deleted_at IS NULL
		LEFT JOIN raster_annotations ra ON sl.id = ra.slide_id AND ra.deleted_at IS NULL
		WHERE s.short_uid = ? AND c.deleted_at IS NULL`
}
