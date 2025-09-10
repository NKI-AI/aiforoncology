// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package users

import (
	"context"
	"database/sql"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// SearchService handles user search and listing operations
type SearchService struct {
	db *sql.DB
}

// NewSearchService creates a new search service instance
func NewSearchService(db *sql.DB) *SearchService {
	return &SearchService{db: db}
}

// LoadAllUsers retrieves users from the database with optional search/filter and pagination support
func (s *SearchService) LoadAllUsers(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.User, error) {
	baseQuery := buildUsersBaseQuery()

	// Build WHERE clause based on search parameters
	searchableFields := []string{"short_uid", "email", "first_name", "last_name"} // Define which fields are searchable
	whereConditions, args := buildUsersSearchWhereClause(search, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Add ordering
	orderBy := "u.created_at DESC" // Default ordering
	if search.HasSort() {
		// Prevent SQL injection by validating sort direction
		dir := strings.ToUpper(search.SortDir)
		if dir != "ASC" && dir != "DESC" {
			dir = "ASC"
		}
		switch search.SortBy {
		case "email":
			orderBy = "u.email " + dir
		case "created_at", "createdAt":
			orderBy = "u.created_at " + dir
		case "updated_at", "updatedAt":
			orderBy = "u.updated_at " + dir
		case "short_uid", "shortId", "userUid":
			orderBy = "u.short_uid " + dir
		default:
			// Keep default ordering for unknown sort fields
		}
	}

	baseQuery += " ORDER BY " + orderBy

	// Add pagination
	if limit > 0 {
		baseQuery += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("users", err)
	}
	defer rows.Close()

	var users []ports.User
	for rows.Next() {
		var user ports.User
		if err := rows.Scan(&user.ID, &user.TenantID, &user.TenantUID, &user.ShortUID, &user.Email, &user.FirstName, &user.LastName, &user.Password, &user.MustResetPassword, &user.IsActive, &user.EmailVerified, &user.PasswordChangedAt, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, errors.NewDatabaseScanError("user", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("users", err)
	}

	return users, nil
}

// GetUserCount returns the total count of users matching optional search criteria
func (s *SearchService) GetUserCount(ctx context.Context, search utils.SearchParams) (int, error) {
	baseQuery := buildUsersCountQuery()

	// Build WHERE clause based on search parameters (same logic as LoadAllUsers)
	searchableFields := []string{"short_uid", "email", "first_name", "last_name"} // Define which fields are searchable
	whereConditions, args := buildUsersSearchWhereClause(search, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	var count int
	err := s.db.QueryRow(baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, errors.NewDatabaseQueryError("user count", err)
	}
	return count, nil
}
