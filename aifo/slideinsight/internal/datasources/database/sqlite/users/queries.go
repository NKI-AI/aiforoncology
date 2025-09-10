// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package users

import (
	"strings"

	"aifo.dev/aifo/slideinsight/internal/utils"
)

// buildUsersSearchWhereClause builds WHERE conditions and arguments for users search queries
func buildUsersSearchWhereClause(search utils.SearchParams, searchableFields []string) ([]string, []interface{}) {
	var whereConditions []string
	var args []interface{}

	// General search across specified searchable fields
	if search.Query != "" {
		// Build LIKE conditions for each searchable field
		var fieldConditions []string
		for _, field := range searchableFields {
			switch field {
			case "short_uid":
				fieldConditions = append(fieldConditions, "u.short_uid LIKE ?")
				args = append(args, "%"+search.Query+"%")
			case "email":
				fieldConditions = append(fieldConditions, "u.email LIKE ?")
				args = append(args, "%"+search.Query+"%")
			case "first_name":
				fieldConditions = append(fieldConditions, "u.first_name LIKE ?")
				args = append(args, "%"+search.Query+"%")
			case "last_name":
				fieldConditions = append(fieldConditions, "u.last_name LIKE ?")
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
		case "email":
			whereConditions = append(whereConditions, "u.email LIKE ?")
			args = append(args, "%"+value+"%")
		case "first_name":
			whereConditions = append(whereConditions, "u.first_name LIKE ?")
			args = append(args, "%"+value+"%")
		case "last_name":
			whereConditions = append(whereConditions, "u.last_name LIKE ?")
			args = append(args, "%"+value+"%")
		case "is_active":
			// Convert string to boolean for database query
			if value == "true" || value == "1" {
				whereConditions = append(whereConditions, "u.is_active = ?")
				args = append(args, true)
			} else if value == "false" || value == "0" {
				whereConditions = append(whereConditions, "u.is_active = ?")
				args = append(args, false)
			}
		case "must_reset_password":
			// Convert string to boolean for database query
			if value == "true" || value == "1" {
				whereConditions = append(whereConditions, "u.must_reset_password = ?")
				args = append(args, true)
			} else if value == "false" || value == "0" {
				whereConditions = append(whereConditions, "u.must_reset_password = ?")
				args = append(args, false)
			}
		default:
			// Skip unknown filter fields
			continue
		}
	}

	return whereConditions, args
}

// buildUsersBaseQuery returns the base query for loading users with joins
func buildUsersBaseQuery() string {
	return `
		SELECT u.id, u.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, u.short_uid, u.email, u.first_name, u.last_name, u.password, u.must_reset_password, u.is_active, u.email_verified, u.password_changed_at, u.created_at, u.updated_at 
		FROM users u
		LEFT JOIN tenants t ON u.tenant_id = t.id`
}

// buildUsersCountQuery returns the base query for counting users
func buildUsersCountQuery() string {
	return `
		SELECT COUNT(*) 
		FROM users u
		LEFT JOIN tenants t ON u.tenant_id = t.id`
}

// validateSortDir ensures sort direction is valid to prevent SQL injection
func validateSortDir(sortDir string) string {
	dir := strings.ToUpper(sortDir)
	if dir != "ASC" && dir != "DESC" {
		return "ASC"
	}
	return dir
}
