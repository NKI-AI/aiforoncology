// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package studies

import (
	"strings"

	"aifo.dev/aifo/slideinsight/internal/utils"
)

// buildStudiesSearchWhereClause builds WHERE conditions and arguments for studies search queries
func buildStudiesSearchWhereClause(search utils.SearchParams, searchableFields []string) ([]string, []interface{}) {
	var whereConditions []string
	var args []interface{}

	// General search across specified searchable fields
	if search.Query != "" {
		// Build LIKE conditions for each searchable field
		var fieldConditions []string
		for _, field := range searchableFields {
			switch field {
			case "name":
				fieldConditions = append(fieldConditions, "s.name LIKE ?")
				args = append(args, "%"+search.Query+"%")
			case "description":
				fieldConditions = append(fieldConditions, "s.description LIKE ?")
				args = append(args, "%"+search.Query+"%")
			case "metadata":
				fieldConditions = append(fieldConditions, "s.metadata LIKE ?")
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
		case "name", "searchName":
			whereConditions = append(whereConditions, "s.name LIKE ?")
			args = append(args, "%"+value+"%")
		case "description", "searchDescription":
			whereConditions = append(whereConditions, "s.description LIKE ?")
			args = append(args, "%"+value+"%")
		case "status", "searchStatus":
			// Handle frontend status values for publication status
			if value == "published" {
				whereConditions = append(whereConditions, "s.is_published = ?")
				args = append(args, true)
			} else if value == "draft" {
				whereConditions = append(whereConditions, "s.is_published = ?")
				args = append(args, false)
			}
		case "is_published":
			// Convert string to boolean for database query (legacy support)
			if value == "true" || value == "1" {
				whereConditions = append(whereConditions, "s.is_published = ?")
				args = append(args, true)
			} else if value == "false" || value == "0" {
				whereConditions = append(whereConditions, "s.is_published = ?")
				args = append(args, false)
			}
		case "tenant_id":
			whereConditions = append(whereConditions, "s.tenant_id = ?")
			args = append(args, value)
		default:
			// Skip unknown filter fields
			continue
		}
	}

	return whereConditions, args
}

// buildStudiesBaseQuery returns the base query for loading studies with joins
func buildStudiesBaseQuery() string {
	return `
		SELECT s.id, s.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, s.short_uid as study_uid, s.creator_id, COALESCE(u.short_uid, '') as creator_uid, s.name, s.description, s.metadata, s.is_published, s.deleted_at, s.deleted_by, s.created_at 
		FROM studies s
		LEFT JOIN tenants t ON s.tenant_id = t.id
		LEFT JOIN users u ON s.creator_id = u.id
		WHERE s.deleted_at IS NULL`
}

// buildStudiesCountQuery returns the base query for counting studies
func buildStudiesCountQuery() string {
	return `
		SELECT COUNT(*) 
		FROM studies s
		LEFT JOIN tenants t ON s.tenant_id = t.id
		LEFT JOIN users u ON s.creator_id = u.id
		WHERE s.deleted_at IS NULL`
}

// validateSortDir ensures sort direction is valid to prevent SQL injection
func validateSortDir(sortDir string) string {
	dir := strings.ToUpper(sortDir)
	if dir != "ASC" && dir != "DESC" {
		return "ASC"
	}
	return dir
}
