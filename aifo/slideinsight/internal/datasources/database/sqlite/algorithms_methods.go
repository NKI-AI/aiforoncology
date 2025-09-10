// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package sqlite

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Algorithm operations

// CreateAlgorithm creates a new algorithm
func (db *DB) CreateAlgorithm(ctx context.Context, newAlgorithm domain.NewAlgorithm) (domain.Algorithm, error) {
	return db.algorithms.CreateAlgorithm(ctx, newAlgorithm)
}

// GetAlgorithms retrieves algorithms with search/filter and pagination
func (db *DB) GetAlgorithms(ctx context.Context, tenantID int, search utils.SearchParams, pagination utils.PaginationParams) ([]domain.Algorithm, error) {
	return db.algorithms.GetAlgorithms(ctx, tenantID, search, pagination)
}

// GetAllAlgorithms retrieves algorithms from all tenants with search/filter and pagination
func (db *DB) GetAllAlgorithms(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]domain.Algorithm, error) {
	return db.algorithms.GetAllAlgorithms(ctx, search, pagination)
}

// GetAlgorithm retrieves a specific algorithm by ID
func (db *DB) GetAlgorithm(ctx context.Context, algorithmID string) (domain.Algorithm, error) {
	return db.algorithms.GetAlgorithm(ctx, algorithmID)
}

// UpdateAlgorithm updates an existing algorithm
func (db *DB) UpdateAlgorithm(ctx context.Context, algorithmID string, updates domain.AlgorithmUpdates) error {
	return db.algorithms.UpdateAlgorithm(ctx, algorithmID, updates)
}

// DeleteAlgorithm deletes an algorithm
func (db *DB) DeleteAlgorithm(ctx context.Context, algorithmID string) error {
	return db.algorithms.DeleteAlgorithm(ctx, algorithmID)
}

// AlgorithmExists checks if an algorithm exists
func (db *DB) AlgorithmExists(ctx context.Context, algorithmID string) (bool, error) {
	return db.algorithms.AlgorithmExists(ctx, algorithmID)
}

// GetAlgorithmsCount returns the total count of algorithms
func (db *DB) GetAlgorithmsCount(ctx context.Context, tenantID int) (int, error) {
	return db.algorithms.GetAlgorithmsCount(ctx, tenantID)
}

// GetAllAlgorithmsCount returns the total count of algorithms across all tenants
func (db *DB) GetAllAlgorithmsCount(ctx context.Context) (int, error) {
	return db.algorithms.GetAllAlgorithmsCount(ctx)
}

// GetAlgorithmsCountWithSearch returns the count of algorithms matching search criteria
func (db *DB) GetAlgorithmsCountWithSearch(ctx context.Context, tenantID int, search utils.SearchParams) (int, error) {
	return db.algorithms.GetAlgorithmsCountWithSearch(ctx, tenantID, search)
}

// GetAllAlgorithmsCountWithSearch returns the count of algorithms matching search criteria across all tenants
func (db *DB) GetAllAlgorithmsCountWithSearch(ctx context.Context, search utils.SearchParams) (int, error) {
	return db.algorithms.GetAllAlgorithmsCountWithSearch(ctx, search)
}

// Result sink operations

// CreateResultSinks creates result sinks for an algorithm
func (db *DB) CreateResultSinks(ctx context.Context, algorithmID string, sinks []domain.NewResultSink) error {
	return db.algorithms.CreateResultSinks(ctx, algorithmID, sinks)
}

// GetResultSinks retrieves result sinks for an algorithm
func (db *DB) GetResultSinks(ctx context.Context, algorithmID string) ([]domain.ResultSink, error) {
	return db.algorithms.GetResultSinks(ctx, algorithmID)
}

// DeleteResultSinks deletes result sinks for an algorithm
func (db *DB) DeleteResultSinks(ctx context.Context, algorithmID string) error {
	return db.algorithms.DeleteResultSinks(ctx, algorithmID)
}

// Post hook operations

// CreatePostHooks creates post hooks for an algorithm
func (db *DB) CreatePostHooks(ctx context.Context, algorithmID string, hooks []domain.NewPostHook) error {
	return db.algorithms.CreatePostHooks(ctx, algorithmID, hooks)
}

// GetPostHooks retrieves post hooks for an algorithm
func (db *DB) GetPostHooks(ctx context.Context, algorithmID string) ([]domain.PostHook, error) {
	return db.algorithms.GetPostHooks(ctx, algorithmID)
}

// DeletePostHooks deletes post hooks for an algorithm
func (db *DB) DeletePostHooks(ctx context.Context, algorithmID string) error {
	return db.algorithms.DeletePostHooks(ctx, algorithmID)
}

// Algorithm run operations

// CreateAlgorithmRun creates a new algorithm run
func (db *DB) CreateAlgorithmRun(ctx context.Context, newRun domain.NewAlgorithmRun) (domain.AlgorithmRun, error) {
	return db.algorithms.CreateAlgorithmRun(ctx, newRun)
}

// GetAlgorithmRuns retrieves algorithm runs with search/filter and pagination
func (db *DB) GetAlgorithmRuns(ctx context.Context, algorithmID string, search utils.SearchParams, pagination utils.PaginationParams) ([]domain.AlgorithmRun, error) {
	return db.algorithms.GetAlgorithmRuns(ctx, algorithmID, search, pagination)
}

// GetAlgorithmRun retrieves a specific algorithm run by ID
func (db *DB) GetAlgorithmRun(ctx context.Context, runID string) (domain.AlgorithmRun, error) {
	return db.algorithms.GetAlgorithmRun(ctx, runID)
}

// UpdateAlgorithmRun updates an existing algorithm run
func (db *DB) UpdateAlgorithmRun(ctx context.Context, runID string, updates domain.AlgorithmRunUpdates) error {
	return db.algorithms.UpdateAlgorithmRun(ctx, runID, updates)
}

// CancelAlgorithmRun cancels an algorithm run
func (db *DB) CancelAlgorithmRun(ctx context.Context, runID string) error {
	return db.algorithms.CancelAlgorithmRun(ctx, runID)
}

// GetRunsCount returns the total count of runs for an algorithm
func (db *DB) GetRunsCount(ctx context.Context, algorithmID string) (int, error) {
	return db.algorithms.GetRunsCount(ctx, algorithmID)
}

// GetRunsCountWithSearch returns the count of runs matching search criteria
func (db *DB) GetRunsCountWithSearch(ctx context.Context, algorithmID string, search utils.SearchParams) (int, error) {
	return db.algorithms.GetRunsCountWithSearch(ctx, algorithmID, search)
}

// Output operations

// CreateOutput creates a new output
func (db *DB) CreateOutput(ctx context.Context, newOutput domain.NewOutput) (domain.Output, error) {
	return db.algorithms.CreateOutput(ctx, newOutput)
}

// GetOutputs retrieves outputs for a run
func (db *DB) GetOutputs(ctx context.Context, runID string) ([]domain.Output, error) {
	return db.algorithms.GetOutputs(ctx, runID)
}

// GetOutput retrieves a specific output by ID
func (db *DB) GetOutput(ctx context.Context, outputID string) (domain.Output, error) {
	return db.algorithms.GetOutput(ctx, outputID)
}

// DeleteOutput deletes an output
func (db *DB) DeleteOutput(ctx context.Context, outputID string) error {
	return db.algorithms.DeleteOutput(ctx, outputID)
}
