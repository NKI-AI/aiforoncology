// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package annotations

import (
	"strings"

	"aifo.dev/aifo/slideinsight/internal/utils"
)

// buildVectorAnnotationsSearchWhereClause builds WHERE conditions and arguments for vector annotations search queries
func buildVectorAnnotationsSearchWhereClause(search utils.SearchParams, searchableFields []string) ([]string, []interface{}) {
	var whereConditions []string
	var args []interface{}

	// General search across specified searchable fields
	if search.Query != "" {
		// Build LIKE conditions for each searchable field
		var fieldConditions []string
		for _, field := range searchableFields {
			switch field {
			case "name":
				fieldConditions = append(fieldConditions, "va.name LIKE ?")
				args = append(args, "%"+search.Query+"%")
			case "file_uri":
				fieldConditions = append(fieldConditions, "va.file_uri LIKE ?")
				args = append(args, "%"+search.Query+"%")
			case "slide_uid":
				fieldConditions = append(fieldConditions, "s.slide_uid LIKE ?")
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
			whereConditions = append(whereConditions, "va.name LIKE ?")
			args = append(args, "%"+value+"%")
		case "file_uri":
			whereConditions = append(whereConditions, "va.file_uri LIKE ?")
			args = append(args, "%"+value+"%")
		case "slide_uid":
			whereConditions = append(whereConditions, "s.slide_uid = ?")
			args = append(args, value)
		case "actor_type":
			whereConditions = append(whereConditions, "va.actor_type = ?")
			args = append(args, value)
		default:
			// Skip unknown filter fields
			continue
		}
	}

	return whereConditions, args
}

// buildVectorAnnotationsBaseQuery returns the base query for loading vector annotations with joins
func buildVectorAnnotationsBaseQuery() string {
	return `
		SELECT 
			va.id, va.actor_type, va.actor_id, va.creator_id, va.slide_id, s.slide_uid, va.vector_uid, va.version, va.name, 
			va.file_uri, va.data_blob, va.labels, va.metadata, va.mutable, va.deleted_at, va.deleted_by, va.created_at, va.updated_at
		FROM vector_annotations va
		JOIN slides s ON va.slide_id = s.id
		WHERE va.deleted_at IS NULL AND s.deleted_at IS NULL`
}

// buildVectorAnnotationsCountQuery returns the base query for counting vector annotations
func buildVectorAnnotationsCountQuery() string {
	return `
		SELECT COUNT(*) 
		FROM vector_annotations va
		JOIN slides s ON va.slide_id = s.id
		WHERE va.deleted_at IS NULL AND s.deleted_at IS NULL`
}

// buildRasterAnnotationsSearchWhereClause builds WHERE conditions and arguments for raster annotations search queries
func buildRasterAnnotationsSearchWhereClause(search utils.SearchParams, searchableFields []string) ([]string, []interface{}) {
	var whereConditions []string
	var args []interface{}

	// General search across specified searchable fields
	if search.Query != "" {
		// Build LIKE conditions for each searchable field
		var fieldConditions []string
		for _, field := range searchableFields {
			switch field {
			case "name":
				fieldConditions = append(fieldConditions, "ra.name LIKE ?")
				args = append(args, "%"+search.Query+"%")
			case "file_uri":
				fieldConditions = append(fieldConditions, "ra.file_uri LIKE ?")
				args = append(args, "%"+search.Query+"%")
			case "slide_uid":
				fieldConditions = append(fieldConditions, "s.slide_uid LIKE ?")
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
			whereConditions = append(whereConditions, "ra.name LIKE ?")
			args = append(args, "%"+value+"%")
		case "file_uri":
			whereConditions = append(whereConditions, "ra.file_uri LIKE ?")
			args = append(args, "%"+value+"%")
		case "slide_uid":
			whereConditions = append(whereConditions, "s.slide_uid = ?")
			args = append(args, value)
		case "actor_type":
			whereConditions = append(whereConditions, "ra.actor_type = ?")
			args = append(args, value)
		case "format":
			whereConditions = append(whereConditions, "ra.format = ?")
			args = append(args, value)
		default:
			// Skip unknown filter fields
			continue
		}
	}

	return whereConditions, args
}

// buildRasterAnnotationsBaseQuery returns the base query for loading raster annotations with joins
func buildRasterAnnotationsBaseQuery() string {
	return `
		SELECT 
			ra.id, ra.actor_type, ra.actor_id, ra.creator_id, ra.slide_id, s.slide_uid, ra.raster_uid, ra.version, ra.name, 
			ra.file_uri, ra.file_hash, ra.format, ra.labels, ra.metadata, ra.mask_width, ra.mask_height, ra.mask_mpp, ra.mutable,
			ra.deleted_at, ra.deleted_by, ra.created_at
		FROM raster_annotations ra
		JOIN slides s ON ra.slide_id = s.id
		WHERE ra.deleted_at IS NULL AND s.deleted_at IS NULL`
}

// buildRasterAnnotationsCountQuery returns the base query for counting raster annotations
func buildRasterAnnotationsCountQuery() string {
	return `
		SELECT COUNT(*) 
		FROM raster_annotations ra
		JOIN slides s ON ra.slide_id = s.id
		WHERE ra.deleted_at IS NULL AND s.deleted_at IS NULL`
}

// validateSortDir ensures sort direction is valid to prevent SQL injection
func validateSortDir(sortDir string) string {
	dir := strings.ToUpper(sortDir)
	if dir != "ASC" && dir != "DESC" {
		return "ASC"
	}
	return dir
}
