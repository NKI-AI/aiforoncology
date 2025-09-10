// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package algorithms

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// SearchService provides search and pagination operations for algorithms
type SearchService struct {
	db *sql.DB
}

// NewSearchService creates a new search service
func NewSearchService(db *sql.DB) *SearchService {
	return &SearchService{db: db}
}

// Algorithm search operations

// GetAlgorithms retrieves algorithms with search and pagination
func (s *SearchService) GetAlgorithms(ctx context.Context, tenantID int, search utils.SearchParams, pagination utils.PaginationParams) ([]domain.Algorithm, error) {
	baseQuery := `
		SELECT id, tenant_id, name, description, version, endpoint_url,
		       http_method, execution_mode, progress_transport, metadata,
		       created_at, updated_at
		FROM algorithms 
		WHERE tenant_id = ?
	`

	args := []interface{}{tenantID}
	conditions := []string{}

	// Add search conditions
	if search.Query != "" {
		conditions = append(conditions, "(name LIKE ? OR description LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(search.Filters) > 0 {
		for field, value := range search.Filters {
			switch field {
			case "name", "version", "execution_mode", "progress_transport":
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
		baseQuery += " ORDER BY created_at DESC"
	}

	// Add pagination
	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, pagination.Limit, pagination.CalculateOffset())

	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query algorithms: %w", err)
	}
	defer rows.Close()

	var algorithms []domain.Algorithm
	for rows.Next() {
		var alg domain.Algorithm
		var description, metadata sql.NullString
		var createdAt, updatedAt string

		err := rows.Scan(&alg.ID, &alg.TenantID, &alg.Name, &description, &alg.Version,
			&alg.EndpointURL, &alg.HTTPMethod, &alg.ExecutionMode, &alg.ProgressTransport,
			&metadata, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan algorithm: %w", err)
		}

		alg.Description = description.String
		alg.Metadata = metadata.String

		if alg.CreatedAt, err = parseTimestamp(createdAt); err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}
		if alg.UpdatedAt, err = parseTimestamp(updatedAt); err != nil {
			return nil, fmt.Errorf("failed to parse updated_at: %w", err)
		}

		algorithms = append(algorithms, alg)
	}

	// Ensure we return an empty slice instead of nil
	if algorithms == nil {
		algorithms = []domain.Algorithm{}
	}

	return algorithms, nil
}

// GetAlgorithmsCount returns the total count of algorithms for a tenant
func (s *SearchService) GetAlgorithmsCount(ctx context.Context, tenantID int) (int, error) {
	query := "SELECT COUNT(*) FROM algorithms WHERE tenant_id = ?"

	var count int
	err := s.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get algorithms count: %w", err)
	}

	return count, nil
}

// GetAlgorithmsCountWithSearch returns the count of algorithms matching search criteria
func (s *SearchService) GetAlgorithmsCountWithSearch(ctx context.Context, tenantID int, search utils.SearchParams) (int, error) {
	baseQuery := "SELECT COUNT(*) FROM algorithms WHERE tenant_id = ?"
	args := []interface{}{tenantID}
	conditions := []string{}

	if search.Query != "" {
		conditions = append(conditions, "(name LIKE ? OR description LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(search.Filters) > 0 {
		for field, value := range search.Filters {
			switch field {
			case "name", "version", "execution_mode", "progress_transport":
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
		return 0, fmt.Errorf("failed to get algorithms count with search: %w", err)
	}

	return count, nil
}

// Algorithm run search operations

// GetAlgorithmRuns retrieves algorithm runs with search and pagination
func (s *SearchService) GetAlgorithmRuns(ctx context.Context, algorithmID string, search utils.SearchParams, pagination utils.PaginationParams) ([]domain.AlgorithmRun, error) {
	baseQuery := `
		SELECT id, algorithm_id, case_id, image_ids, regions, parameters,
		       execution_mode, status, progress, result_uri, error_info,
		       created_at, started_at, finished_at, updated_at
		FROM algorithm_runs 
		WHERE algorithm_id = ?
	`

	args := []interface{}{algorithmID}
	conditions := []string{}

	// Add search conditions
	if search.Query != "" {
		conditions = append(conditions, "(status LIKE ? OR case_id LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(search.Filters) > 0 {
		for field, value := range search.Filters {
			switch field {
			case "status", "execution_mode", "case_id":
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
		baseQuery += " ORDER BY created_at DESC"
	}

	// Add pagination
	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, pagination.Limit, pagination.CalculateOffset())

	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query algorithm runs: %w", err)
	}
	defer rows.Close()

	var runs []domain.AlgorithmRun
	for rows.Next() {
		var run domain.AlgorithmRun
		var caseID, imageIDs, regions, parameters, resultURI, errorInfo sql.NullString
		var startedAt, finishedAt sql.NullString
		var createdAt, updatedAt string

		err := rows.Scan(&run.ID, &run.AlgorithmID, &caseID, &imageIDs, &regions, &parameters,
			&run.ExecutionMode, &run.Status, &run.Progress, &resultURI, &errorInfo,
			&createdAt, &startedAt, &finishedAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan algorithm run: %w", err)
		}

		// Handle nullable fields
		if caseID.Valid {
			run.CaseID = &caseID.String
		}
		run.ImageIDs = imageIDs.String
		run.Regions = regions.String
		run.Parameters = parameters.String
		if resultURI.Valid {
			run.ResultURI = &resultURI.String
		}
		if errorInfo.Valid {
			run.ErrorInfo = &errorInfo.String
		}

		// Parse timestamps
		if run.CreatedAt, err = parseTimestamp(createdAt); err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}
		if run.UpdatedAt, err = parseTimestamp(updatedAt); err != nil {
			return nil, fmt.Errorf("failed to parse updated_at: %w", err)
		}
		if startedAt.Valid {
			if parsedStartedAt, err := parseTimestamp(startedAt.String); err == nil {
				run.StartedAt = &parsedStartedAt
			}
		}
		if finishedAt.Valid {
			if parsedFinishedAt, err := parseTimestamp(finishedAt.String); err == nil {
				run.FinishedAt = &parsedFinishedAt
			}
		}

		runs = append(runs, run)
	}

	// Ensure we return an empty slice instead of nil
	if runs == nil {
		runs = []domain.AlgorithmRun{}
	}

	return runs, nil
}

// GetRunsCount returns the total count of runs for an algorithm
func (s *SearchService) GetRunsCount(ctx context.Context, algorithmID string) (int, error) {
	query := "SELECT COUNT(*) FROM algorithm_runs WHERE algorithm_id = ?"

	var count int
	err := s.db.QueryRowContext(ctx, query, algorithmID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get runs count: %w", err)
	}

	return count, nil
}

// GetRunsCountWithSearch returns the count of runs matching search criteria
func (s *SearchService) GetRunsCountWithSearch(ctx context.Context, algorithmID string, search utils.SearchParams) (int, error) {
	baseQuery := "SELECT COUNT(*) FROM algorithm_runs WHERE algorithm_id = ?"
	args := []interface{}{algorithmID}
	conditions := []string{}

	if search.Query != "" {
		conditions = append(conditions, "(status LIKE ? OR case_id LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(search.Filters) > 0 {
		for field, value := range search.Filters {
			switch field {
			case "status", "execution_mode", "case_id":
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
		return 0, fmt.Errorf("failed to get runs count with search: %w", err)
	}

	return count, nil
}

// GetAllAlgorithms retrieves algorithms from all tenants with pagination and filtering
func (s *SearchService) GetAllAlgorithms(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]domain.Algorithm, error) {
	baseQuery := `
		SELECT a.id, a.tenant_id, a.name, a.description, a.version, a.endpoint_url,
		       a.http_method, a.execution_mode, a.progress_transport, a.metadata,
		       a.created_at, a.updated_at, t.name as tenant_name
		FROM algorithms a
		LEFT JOIN tenants t ON a.tenant_id = t.id
		WHERE 1=1
	`

	args := []interface{}{}
	conditions := []string{}

	// Add search conditions
	if search.Query != "" {
		conditions = append(conditions, "(a.name LIKE ? OR a.description LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(search.Filters) > 0 {
		for field, value := range search.Filters {
			switch field {
			case "name", "version", "execution_mode", "progress_transport":
				conditions = append(conditions, fmt.Sprintf("a.%s = ?", field))
				args = append(args, value)
			case "tenant_id":
				conditions = append(conditions, "a.tenant_id = ?")
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
		// Handle sorting by tenant name
		if search.SortBy == "tenant_name" {
			baseQuery += fmt.Sprintf(" ORDER BY t.name %s", direction)
		} else {
			baseQuery += fmt.Sprintf(" ORDER BY a.%s %s", search.SortBy, direction)
		}
	} else {
		baseQuery += " ORDER BY t.name ASC, a.created_at DESC"
	}

	// Add pagination
	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, pagination.Limit, pagination.CalculateOffset())

	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query all algorithms: %w", err)
	}
	defer rows.Close()

	var algorithms []domain.Algorithm
	for rows.Next() {
		algorithm, err := s.scanAlgorithmWithTenant(rows)
		if err != nil {
			return nil, err
		}
		algorithms = append(algorithms, *algorithm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate algorithm rows: %w", err)
	}

	// Ensure we return an empty slice instead of nil
	if algorithms == nil {
		algorithms = []domain.Algorithm{}
	}

	return algorithms, nil
}

// GetAllAlgorithmsCount returns the total count of algorithms across all tenants
func (s *SearchService) GetAllAlgorithmsCount(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM algorithms"
	var count int
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count all algorithms: %w", err)
	}
	return count, nil
}

// GetAllAlgorithmsCountWithSearch returns the count of algorithms matching search criteria across all tenants
func (s *SearchService) GetAllAlgorithmsCountWithSearch(ctx context.Context, search utils.SearchParams) (int, error) {
	baseQuery := "SELECT COUNT(*) FROM algorithms a WHERE 1=1"
	args := []interface{}{}
	conditions := []string{}

	// Add search conditions
	if search.Query != "" {
		conditions = append(conditions, "(a.name LIKE ? OR a.description LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(search.Filters) > 0 {
		for field, value := range search.Filters {
			switch field {
			case "name", "version", "execution_mode", "progress_transport":
				conditions = append(conditions, fmt.Sprintf("a.%s = ?", field))
				args = append(args, value)
			case "tenant_id":
				conditions = append(conditions, "a.tenant_id = ?")
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
		return 0, fmt.Errorf("failed to count algorithms with search: %w", err)
	}
	return count, nil
}

// scanAlgorithmWithTenant scans a row into an Algorithm struct including tenant name
func (s *SearchService) scanAlgorithmWithTenant(rows *sql.Rows) (*domain.Algorithm, error) {
	var alg domain.Algorithm
	var description, metadata, tenantName sql.NullString
	var createdAt, updatedAt string

	err := rows.Scan(&alg.ID, &alg.TenantID, &alg.Name, &description, &alg.Version,
		&alg.EndpointURL, &alg.HTTPMethod, &alg.ExecutionMode, &alg.ProgressTransport,
		&metadata, &createdAt, &updatedAt, &tenantName)
	if err != nil {
		return nil, fmt.Errorf("failed to scan algorithm row: %w", err)
	}

	alg.Description = description.String
	alg.Metadata = metadata.String
	alg.TenantName = tenantName.String

	if alg.CreatedAt, err = parseTimestamp(createdAt); err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	if alg.UpdatedAt, err = parseTimestamp(updatedAt); err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &alg, nil
}

// parseTimestamp flexibly parses timestamps in multiple formats
func parseTimestamp(timestamp string) (time.Time, error) {
	if timestamp == "" {
		return time.Time{}, nil
	}

	// Try RFC3339 format first (2006-01-02T15:04:05Z)
	if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
		return t, nil
	}

	// Try SQLite datetime format (2006-01-02 15:04:05)
	if t, err := time.Parse("2006-01-02 15:04:05", timestamp); err == nil {
		return t, nil
	}

	// Try ISO8601 with timezone offset (2006-01-02T15:04:05Z07:00)
	if t, err := time.Parse(time.RFC3339Nano, timestamp); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp '%s'", timestamp)
}
