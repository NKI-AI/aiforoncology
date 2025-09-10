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
	"encoding/json"
	"fmt"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	apperrors "aifo.dev/aifo/slideinsight/internal/server/errors"
	"github.com/google/uuid"
)

// CrudService provides CRUD operations for algorithms
type CrudService struct {
	db *sql.DB
}

// NewCrudService creates a new CRUD service
func NewCrudService(db *sql.DB) *CrudService {
	return &CrudService{db: db}
}

// Algorithm operations

// CreateAlgorithm creates a new algorithm in the database
func (s *CrudService) CreateAlgorithm(ctx context.Context, newAlgorithm domain.NewAlgorithm) (domain.Algorithm, error) {
	// Set defaults
	httpMethod := newAlgorithm.HTTPMethod
	if httpMethod == "" {
		httpMethod = "POST"
	}

	progressTransport := newAlgorithm.ProgressTransport
	if progressTransport == "" {
		progressTransport = "WEBSOCKET"
	}

	// Generate UUID for the algorithm
	algorithmID := uuid.New().String()

	query := `
		INSERT INTO algorithms (id, tenant_id, name, description, version, endpoint_url, 
		                       http_method, execution_mode, progress_transport, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		algorithmID, newAlgorithm.TenantID, newAlgorithm.Name, newAlgorithm.Description,
		newAlgorithm.Version, newAlgorithm.EndpointURL, httpMethod,
		newAlgorithm.ExecutionMode, progressTransport, newAlgorithm.Metadata)
	if err != nil {
		return domain.Algorithm{}, fmt.Errorf("failed to create algorithm: %w", err)
	}

	// Return the created algorithm
	return s.GetAlgorithm(ctx, algorithmID)
}

// GetAlgorithm retrieves a single algorithm by ID
func (s *CrudService) GetAlgorithm(ctx context.Context, algorithmID string) (domain.Algorithm, error) {
	query := `
		SELECT id, tenant_id, name, description, version, endpoint_url,
		       http_method, execution_mode, progress_transport, metadata,
		       created_at, updated_at
		FROM algorithms 
		WHERE id = ?
	`

	row := s.db.QueryRowContext(ctx, query, algorithmID)

	var alg domain.Algorithm
	var description, metadata sql.NullString
	var createdAt, updatedAt string

	err := row.Scan(&alg.ID, &alg.TenantID, &alg.Name, &description, &alg.Version,
		&alg.EndpointURL, &alg.HTTPMethod, &alg.ExecutionMode, &alg.ProgressTransport,
		&metadata, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Algorithm{}, apperrors.ErrAlgorithmNotFound
		}
		return domain.Algorithm{}, fmt.Errorf("failed to get algorithm: %w", err)
	}

	alg.Description = description.String
	alg.Metadata = metadata.String

	if alg.CreatedAt, err = parseTimestamp(createdAt); err != nil {
		return domain.Algorithm{}, fmt.Errorf("failed to parse created_at: %w", err)
	}
	if alg.UpdatedAt, err = parseTimestamp(updatedAt); err != nil {
		return domain.Algorithm{}, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return alg, nil
}

// UpdateAlgorithm updates an existing algorithm
func (s *CrudService) UpdateAlgorithm(ctx context.Context, algorithmID string, updates domain.AlgorithmUpdates) error {
	setParts := []string{}
	args := []interface{}{}

	if updates.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *updates.Name)
	}
	if updates.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *updates.Description)
	}
	if updates.Version != nil {
		setParts = append(setParts, "version = ?")
		args = append(args, *updates.Version)
	}
	if updates.EndpointURL != nil {
		setParts = append(setParts, "endpoint_url = ?")
		args = append(args, *updates.EndpointURL)
	}
	if updates.HTTPMethod != nil {
		setParts = append(setParts, "http_method = ?")
		args = append(args, *updates.HTTPMethod)
	}
	if updates.ExecutionMode != nil {
		setParts = append(setParts, "execution_mode = ?")
		args = append(args, *updates.ExecutionMode)
	}
	if updates.ProgressTransport != nil {
		setParts = append(setParts, "progress_transport = ?")
		args = append(args, *updates.ProgressTransport)
	}
	if updates.Metadata != nil {
		setParts = append(setParts, "metadata = ?")
		args = append(args, *updates.Metadata)
	}

	if len(setParts) == 0 {
		return nil // No updates to apply
	}

	query := fmt.Sprintf("UPDATE algorithms SET %s WHERE id = ?", strings.Join(setParts, ", "))
	args = append(args, algorithmID)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update algorithm: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperrors.ErrAlgorithmNotFound
	}

	return nil
}

// DeleteAlgorithm deletes an algorithm
func (s *CrudService) DeleteAlgorithm(ctx context.Context, algorithmID string) error {
	query := "DELETE FROM algorithms WHERE id = ?"

	result, err := s.db.ExecContext(ctx, query, algorithmID)
	if err != nil {
		return fmt.Errorf("failed to delete algorithm: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperrors.ErrAlgorithmNotFound
	}

	return nil
}

// AlgorithmExists checks if an algorithm exists
func (s *CrudService) AlgorithmExists(ctx context.Context, algorithmID string) (bool, error) {
	query := "SELECT 1 FROM algorithms WHERE id = ? LIMIT 1"

	var exists int
	err := s.db.QueryRowContext(ctx, query, algorithmID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check algorithm existence: %w", err)
	}

	return true, nil
}

// Result sink operations

// CreateResultSinks creates result sinks for an algorithm
func (s *CrudService) CreateResultSinks(ctx context.Context, algorithmID string, sinks []domain.NewResultSink) error {
	if len(sinks) == 0 {
		return nil
	}

	query := "INSERT INTO result_sinks (id, algorithm_id, type, config) VALUES (?, ?, ?, ?)"

	for _, sink := range sinks {
		sinkID := uuid.New().String()
		_, err := s.db.ExecContext(ctx, query, sinkID, algorithmID, sink.Type, sink.Config)
		if err != nil {
			return fmt.Errorf("failed to create result sink: %w", err)
		}
	}

	return nil
}

// GetResultSinks retrieves result sinks for an algorithm
func (s *CrudService) GetResultSinks(ctx context.Context, algorithmID string) ([]domain.ResultSink, error) {
	query := `
		SELECT id, algorithm_id, type, config, created_at
		FROM result_sinks 
		WHERE algorithm_id = ?
		ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, algorithmID)
	if err != nil {
		return nil, fmt.Errorf("failed to query result sinks: %w", err)
	}
	defer rows.Close()

	var sinks []domain.ResultSink
	for rows.Next() {
		var sink domain.ResultSink
		var createdAt string

		err := rows.Scan(&sink.ID, &sink.AlgorithmID, &sink.Type, &sink.Config, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result sink: %w", err)
		}

		if sink.CreatedAt, err = parseTimestamp(createdAt); err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		sinks = append(sinks, sink)
	}

	return sinks, nil
}

// DeleteResultSinks deletes all result sinks for an algorithm
func (s *CrudService) DeleteResultSinks(ctx context.Context, algorithmID string) error {
	query := "DELETE FROM result_sinks WHERE algorithm_id = ?"

	_, err := s.db.ExecContext(ctx, query, algorithmID)
	if err != nil {
		return fmt.Errorf("failed to delete result sinks: %w", err)
	}

	return nil
}

// Post hook operations

// CreatePostHooks creates post hooks for an algorithm
func (s *CrudService) CreatePostHooks(ctx context.Context, algorithmID string, hooks []domain.NewPostHook) error {
	if len(hooks) == 0 {
		return nil
	}

	query := "INSERT INTO post_hooks (id, algorithm_id, type, config) VALUES (?, ?, ?, ?)"

	for _, hook := range hooks {
		hookID := uuid.New().String()
		_, err := s.db.ExecContext(ctx, query, hookID, algorithmID, hook.Type, hook.Config)
		if err != nil {
			return fmt.Errorf("failed to create post hook: %w", err)
		}
	}

	return nil
}

// GetPostHooks retrieves post hooks for an algorithm
func (s *CrudService) GetPostHooks(ctx context.Context, algorithmID string) ([]domain.PostHook, error) {
	query := `
		SELECT id, algorithm_id, type, config, created_at
		FROM post_hooks 
		WHERE algorithm_id = ?
		ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, algorithmID)
	if err != nil {
		return nil, fmt.Errorf("failed to query post hooks: %w", err)
	}
	defer rows.Close()

	var hooks []domain.PostHook
	for rows.Next() {
		var hook domain.PostHook
		var createdAt string

		err := rows.Scan(&hook.ID, &hook.AlgorithmID, &hook.Type, &hook.Config, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post hook: %w", err)
		}

		if hook.CreatedAt, err = parseTimestamp(createdAt); err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		hooks = append(hooks, hook)
	}

	return hooks, nil
}

// DeletePostHooks deletes all post hooks for an algorithm
func (s *CrudService) DeletePostHooks(ctx context.Context, algorithmID string) error {
	query := "DELETE FROM post_hooks WHERE algorithm_id = ?"

	_, err := s.db.ExecContext(ctx, query, algorithmID)
	if err != nil {
		return fmt.Errorf("failed to delete post hooks: %w", err)
	}

	return nil
}

// Algorithm run operations

// CreateAlgorithmRun creates a new algorithm run
func (s *CrudService) CreateAlgorithmRun(ctx context.Context, newRun domain.NewAlgorithmRun) (domain.AlgorithmRun, error) {
	// Generate UUID for the run
	runID := uuid.New().String()

	// Get algorithm to determine execution mode if not specified
	algorithm, err := s.GetAlgorithm(ctx, newRun.AlgorithmID)
	if err != nil {
		return domain.AlgorithmRun{}, fmt.Errorf("failed to get algorithm: %w", err)
	}

	executionMode := algorithm.ExecutionMode
	if newRun.ExecutionMode != nil {
		executionMode = *newRun.ExecutionMode
	}

	// Convert image IDs to JSON
	var imageIDsJSON string
	if len(newRun.ImageIDs) > 0 {
		imageIDsBytes, err := json.Marshal(newRun.ImageIDs)
		if err != nil {
			return domain.AlgorithmRun{}, fmt.Errorf("failed to marshal image IDs: %w", err)
		}
		imageIDsJSON = string(imageIDsBytes)
	}

	query := `
		INSERT INTO algorithm_runs (id, algorithm_id, case_id, image_ids, regions, 
		                           parameters, execution_mode, status, progress)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'QUEUED', 0)
	`

	_, err = s.db.ExecContext(ctx, query,
		runID, newRun.AlgorithmID, newRun.CaseID, imageIDsJSON,
		newRun.Regions, newRun.Parameters, executionMode)
	if err != nil {
		return domain.AlgorithmRun{}, fmt.Errorf("failed to create algorithm run: %w", err)
	}

	// Create associated result sinks and post hooks
	if len(newRun.ResultSinks) > 0 {
		err = s.CreateResultSinks(ctx, newRun.AlgorithmID, newRun.ResultSinks)
		if err != nil {
			return domain.AlgorithmRun{}, fmt.Errorf("failed to create result sinks: %w", err)
		}
	}

	if len(newRun.PostHooks) > 0 {
		err = s.CreatePostHooks(ctx, newRun.AlgorithmID, newRun.PostHooks)
		if err != nil {
			return domain.AlgorithmRun{}, fmt.Errorf("failed to create post hooks: %w", err)
		}
	}

	// Return the created run
	return s.GetAlgorithmRun(ctx, runID)
}

// GetAlgorithmRun retrieves a specific algorithm run by ID
func (s *CrudService) GetAlgorithmRun(ctx context.Context, runID string) (domain.AlgorithmRun, error) {
	query := `
		SELECT id, algorithm_id, case_id, image_ids, regions, parameters,
		       execution_mode, status, progress, result_uri, error_info,
		       created_at, started_at, finished_at, updated_at
		FROM algorithm_runs 
		WHERE id = ?
	`

	row := s.db.QueryRowContext(ctx, query, runID)

	var run domain.AlgorithmRun
	var caseID, imageIDs, regions, parameters, resultURI, errorInfo sql.NullString
	var startedAt, finishedAt sql.NullString
	var createdAt, updatedAt string

	err := row.Scan(&run.ID, &run.AlgorithmID, &caseID, &imageIDs, &regions, &parameters,
		&run.ExecutionMode, &run.Status, &run.Progress, &resultURI, &errorInfo,
		&createdAt, &startedAt, &finishedAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.AlgorithmRun{}, apperrors.ErrRunNotFound
		}
		return domain.AlgorithmRun{}, fmt.Errorf("failed to get algorithm run: %w", err)
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
		return domain.AlgorithmRun{}, fmt.Errorf("failed to parse created_at: %w", err)
	}
	if run.UpdatedAt, err = parseTimestamp(updatedAt); err != nil {
		return domain.AlgorithmRun{}, fmt.Errorf("failed to parse updated_at: %w", err)
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

	return run, nil
}

// UpdateAlgorithmRun updates an existing algorithm run
func (s *CrudService) UpdateAlgorithmRun(ctx context.Context, runID string, updates domain.AlgorithmRunUpdates) error {
	setParts := []string{}
	args := []interface{}{}

	if updates.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *updates.Status)
	}
	if updates.Progress != nil {
		setParts = append(setParts, "progress = ?")
		args = append(args, *updates.Progress)
	}
	if updates.ResultURI != nil {
		setParts = append(setParts, "result_uri = ?")
		args = append(args, *updates.ResultURI)
	}
	if updates.ErrorInfo != nil {
		setParts = append(setParts, "error_info = ?")
		args = append(args, *updates.ErrorInfo)
	}
	if updates.StartedAt != nil {
		setParts = append(setParts, "started_at = ?")
		args = append(args, updates.StartedAt.Format("2006-01-02 15:04:05"))
	}
	if updates.FinishedAt != nil {
		setParts = append(setParts, "finished_at = ?")
		args = append(args, updates.FinishedAt.Format("2006-01-02 15:04:05"))
	}

	if len(setParts) == 0 {
		return nil // No updates to apply
	}

	query := fmt.Sprintf("UPDATE algorithm_runs SET %s WHERE id = ?", strings.Join(setParts, ", "))
	args = append(args, runID)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update algorithm run: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperrors.ErrRunNotFound
	}

	return nil
}

// CancelAlgorithmRun cancels an algorithm run
func (s *CrudService) CancelAlgorithmRun(ctx context.Context, runID string) error {
	query := "UPDATE algorithm_runs SET status = 'FAILED' WHERE id = ? AND status IN ('QUEUED', 'RUNNING')"

	result, err := s.db.ExecContext(ctx, query, runID)
	if err != nil {
		return fmt.Errorf("failed to cancel algorithm run: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperrors.ErrRunNotFound
	}

	return nil
}

// Output operations

// CreateOutput creates a new output
func (s *CrudService) CreateOutput(ctx context.Context, newOutput domain.NewOutput) (domain.Output, error) {
	outputID := uuid.New().String()

	query := "INSERT INTO outputs (id, run_id, name, uri, metadata) VALUES (?, ?, ?, ?, ?)"

	_, err := s.db.ExecContext(ctx, query, outputID, newOutput.RunID, newOutput.Name, newOutput.URI, newOutput.Metadata)
	if err != nil {
		return domain.Output{}, fmt.Errorf("failed to create output: %w", err)
	}

	return s.GetOutput(ctx, outputID)
}

// GetOutputs retrieves outputs for a run
func (s *CrudService) GetOutputs(ctx context.Context, runID string) ([]domain.Output, error) {
	query := `
		SELECT id, run_id, name, uri, metadata, created_at
		FROM outputs 
		WHERE run_id = ?
		ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to query outputs: %w", err)
	}
	defer rows.Close()

	var outputs []domain.Output
	for rows.Next() {
		var output domain.Output
		var metadata sql.NullString
		var createdAt string

		err := rows.Scan(&output.ID, &output.RunID, &output.Name, &output.URI, &metadata, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan output: %w", err)
		}

		output.Metadata = metadata.String

		if output.CreatedAt, err = parseTimestamp(createdAt); err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
}

// GetOutput retrieves a specific output by ID
func (s *CrudService) GetOutput(ctx context.Context, outputID string) (domain.Output, error) {
	query := `
		SELECT id, run_id, name, uri, metadata, created_at
		FROM outputs 
		WHERE id = ?
	`

	row := s.db.QueryRowContext(ctx, query, outputID)

	var output domain.Output
	var metadata sql.NullString
	var createdAt string

	err := row.Scan(&output.ID, &output.RunID, &output.Name, &output.URI, &metadata, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Output{}, apperrors.ErrOutputNotFound
		}
		return domain.Output{}, fmt.Errorf("failed to get output: %w", err)
	}

	output.Metadata = metadata.String

	if output.CreatedAt, err = parseTimestamp(createdAt); err != nil {
		return domain.Output{}, fmt.Errorf("failed to parse created_at: %w", err)
	}

	return output, nil
}

// DeleteOutput deletes an output
func (s *CrudService) DeleteOutput(ctx context.Context, outputID string) error {
	query := "DELETE FROM outputs WHERE id = ?"

	result, err := s.db.ExecContext(ctx, query, outputID)
	if err != nil {
		return fmt.Errorf("failed to delete output: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperrors.ErrOutputNotFound
	}

	return nil
}
