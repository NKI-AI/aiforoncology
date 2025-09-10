// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package settings

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// SearchService provides search and pagination operations for settings
type SearchService struct {
	db *sql.DB
}

// NewSearchService creates a new search service
func NewSearchService(db *sql.DB) *SearchService {
	return &SearchService{db: db}
}

// GetSettings retrieves settings with search and pagination
func (s *SearchService) GetSettings(ctx context.Context, tenantID int, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Setting, error) {
	baseQuery := `
		SELECT id, tenant_id, key, value_type, value, created_at, updated_at
		FROM settings 
		WHERE tenant_id = ?
	`

	args := []interface{}{tenantID}
	conditions := []string{}

	// Add search conditions
	if search.Query != "" {
		conditions = append(conditions, "(key LIKE ? OR value LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(search.Filters) > 0 {
		for field, value := range search.Filters {
			switch field {
			case "key", "value_type":
				conditions = append(conditions, fmt.Sprintf("%s = ?", field))
				args = append(args, value)
			}
		}
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Add ordering
	if search.SortBy != "" {
		direction := "ASC"
		if search.SortDir == "desc" {
			direction = "DESC"
		}
		baseQuery += fmt.Sprintf(" ORDER BY %s %s", search.SortBy, direction)
	} else {
		baseQuery += " ORDER BY key ASC"
	}

	// Add pagination
	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, pagination.Limit, pagination.CalculateOffset())

	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query settings: %w", err)
	}
	defer rows.Close()

	var settings []ports.Setting
	for rows.Next() {
		setting, err := s.scanSetting(rows)
		if err != nil {
			return nil, err
		}
		settings = append(settings, *setting)
	}

	// Ensure we return an empty slice instead of nil
	if settings == nil {
		settings = []ports.Setting{}
	}

	return settings, nil
}

// GetAllSettings retrieves settings from all tenants with pagination and filtering
func (s *SearchService) GetAllSettings(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Setting, error) {
	baseQuery := `
		SELECT id, tenant_id, key, value_type, value, created_at, updated_at
		FROM settings 
		WHERE 1=1
	`

	args := []interface{}{}
	conditions := []string{}

	// Add search conditions
	if search.Query != "" {
		conditions = append(conditions, "(key LIKE ? OR value LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(search.Filters) > 0 {
		for field, value := range search.Filters {
			switch field {
			case "key", "value_type":
				conditions = append(conditions, fmt.Sprintf("%s = ?", field))
				args = append(args, value)
			case "tenant_id":
				conditions = append(conditions, "tenant_id = ?")
				args = append(args, value)
			}
		}
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Add ordering
	if search.SortBy != "" {
		direction := "ASC"
		if search.SortDir == "desc" {
			direction = "DESC"
		}
		baseQuery += fmt.Sprintf(" ORDER BY %s %s", search.SortBy, direction)
	} else {
		baseQuery += " ORDER BY tenant_id ASC, key ASC"
	}

	// Add pagination
	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, pagination.Limit, pagination.CalculateOffset())

	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query all settings: %w", err)
	}
	defer rows.Close()

	var settings []ports.Setting
	for rows.Next() {
		setting, err := s.scanSetting(rows)
		if err != nil {
			return nil, err
		}
		settings = append(settings, *setting)
	}

	// Ensure we return an empty slice instead of nil
	if settings == nil {
		settings = []ports.Setting{}
	}

	return settings, nil
}

// GetSettingsCount returns the total count of settings for a tenant
func (s *SearchService) GetSettingsCount(ctx context.Context, tenantID int) (int, error) {
	query := "SELECT COUNT(*) FROM settings WHERE tenant_id = ?"

	var count int
	err := s.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get settings count: %w", err)
	}

	return count, nil
}

// GetAllSettingsCount returns the total count of settings across all tenants
func (s *SearchService) GetAllSettingsCount(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM settings"
	var count int
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count all settings: %w", err)
	}
	return count, nil
}

// GetSettingsCountWithSearch returns the count of settings matching search criteria
func (s *SearchService) GetSettingsCountWithSearch(ctx context.Context, tenantID int, search utils.SearchParams) (int, error) {
	baseQuery := "SELECT COUNT(*) FROM settings WHERE tenant_id = ?"
	args := []interface{}{tenantID}
	conditions := []string{}

	if search.Query != "" {
		conditions = append(conditions, "(key LIKE ? OR value LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(search.Filters) > 0 {
		for field, value := range search.Filters {
			switch field {
			case "key", "value_type":
				conditions = append(conditions, fmt.Sprintf("%s = ?", field))
				args = append(args, value)
			}
		}
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	var count int
	err := s.db.QueryRowContext(ctx, baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get settings count with search: %w", err)
	}

	return count, nil
}

// GetAllSettingsCountWithSearch returns the count of settings matching search criteria across all tenants
func (s *SearchService) GetAllSettingsCountWithSearch(ctx context.Context, search utils.SearchParams) (int, error) {
	baseQuery := "SELECT COUNT(*) FROM settings WHERE 1=1"
	args := []interface{}{}
	conditions := []string{}

	if search.Query != "" {
		conditions = append(conditions, "(key LIKE ? OR value LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(search.Filters) > 0 {
		for field, value := range search.Filters {
			switch field {
			case "key", "value_type":
				conditions = append(conditions, fmt.Sprintf("%s = ?", field))
				args = append(args, value)
			case "tenant_id":
				conditions = append(conditions, "tenant_id = ?")
				args = append(args, value)
			}
		}
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	var count int
	err := s.db.QueryRowContext(ctx, baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count settings with search: %w", err)
	}
	return count, nil
}

// scanSetting scans a row into a Setting struct
func (s *SearchService) scanSetting(rows *sql.Rows) (*ports.Setting, error) {
	var setting ports.Setting
	var createdAt, updatedAt string

	err := rows.Scan(&setting.ID, &setting.TenantID, &setting.Key, &setting.ValueType,
		&setting.Value, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan setting row: %w", err)
	}

	if setting.CreatedAt, err = parseTimestamp(createdAt); err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	if setting.UpdatedAt, err = parseTimestamp(updatedAt); err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &setting, nil
}
