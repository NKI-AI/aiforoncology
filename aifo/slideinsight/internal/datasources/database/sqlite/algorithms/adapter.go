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

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Adapter provides a unified interface for all algorithm operations
type Adapter struct {
	crud   *CrudService
	search *SearchService
}

// NewAdapter creates a new algorithms adapter
func NewAdapter(db *sql.DB) *Adapter {
	return &Adapter{
		crud:   NewCrudService(db),
		search: NewSearchService(db),
	}
}

// Algorithm operations

// CreateAlgorithm creates a new algorithm
func (a *Adapter) CreateAlgorithm(ctx context.Context, newAlgorithm domain.NewAlgorithm) (domain.Algorithm, error) {
	return a.crud.CreateAlgorithm(ctx, newAlgorithm)
}

// GetAlgorithms retrieves algorithms with search/filter and pagination
func (a *Adapter) GetAlgorithms(ctx context.Context, tenantID int, search utils.SearchParams, pagination utils.PaginationParams) ([]domain.Algorithm, error) {
	return a.search.GetAlgorithms(ctx, tenantID, search, pagination)
}

// GetAllAlgorithms retrieves algorithms from all tenants with search/filter and pagination
func (a *Adapter) GetAllAlgorithms(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]domain.Algorithm, error) {
	return a.search.GetAllAlgorithms(ctx, search, pagination)
}

// GetAlgorithm retrieves a specific algorithm by ID
func (a *Adapter) GetAlgorithm(ctx context.Context, algorithmID string) (domain.Algorithm, error) {
	return a.crud.GetAlgorithm(ctx, algorithmID)
}

// UpdateAlgorithm updates an existing algorithm
func (a *Adapter) UpdateAlgorithm(ctx context.Context, algorithmID string, updates domain.AlgorithmUpdates) error {
	return a.crud.UpdateAlgorithm(ctx, algorithmID, updates)
}

// DeleteAlgorithm deletes an algorithm
func (a *Adapter) DeleteAlgorithm(ctx context.Context, algorithmID string) error {
	return a.crud.DeleteAlgorithm(ctx, algorithmID)
}

// AlgorithmExists checks if an algorithm exists
func (a *Adapter) AlgorithmExists(ctx context.Context, algorithmID string) (bool, error) {
	return a.crud.AlgorithmExists(ctx, algorithmID)
}

// GetAlgorithmsCount returns the total count of algorithms for a tenant
func (a *Adapter) GetAlgorithmsCount(ctx context.Context, tenantID int) (int, error) {
	return a.search.GetAlgorithmsCount(ctx, tenantID)
}

// GetAllAlgorithmsCount returns the total count of algorithms across all tenants
func (a *Adapter) GetAllAlgorithmsCount(ctx context.Context) (int, error) {
	return a.search.GetAllAlgorithmsCount(ctx)
}

// GetAlgorithmsCountWithSearch returns the count of algorithms matching search criteria
func (a *Adapter) GetAlgorithmsCountWithSearch(ctx context.Context, tenantID int, search utils.SearchParams) (int, error) {
	return a.search.GetAlgorithmsCountWithSearch(ctx, tenantID, search)
}

// GetAllAlgorithmsCountWithSearch returns the count of algorithms matching search criteria across all tenants
func (a *Adapter) GetAllAlgorithmsCountWithSearch(ctx context.Context, search utils.SearchParams) (int, error) {
	return a.search.GetAllAlgorithmsCountWithSearch(ctx, search)
}

// Result sink operations

// CreateResultSinks creates result sinks for an algorithm
func (a *Adapter) CreateResultSinks(ctx context.Context, algorithmID string, sinks []domain.NewResultSink) error {
	return a.crud.CreateResultSinks(ctx, algorithmID, sinks)
}

// GetResultSinks retrieves result sinks for an algorithm
func (a *Adapter) GetResultSinks(ctx context.Context, algorithmID string) ([]domain.ResultSink, error) {
	return a.crud.GetResultSinks(ctx, algorithmID)
}

// DeleteResultSinks deletes result sinks for an algorithm
func (a *Adapter) DeleteResultSinks(ctx context.Context, algorithmID string) error {
	return a.crud.DeleteResultSinks(ctx, algorithmID)
}

// Post hook operations

// CreatePostHooks creates post hooks for an algorithm
func (a *Adapter) CreatePostHooks(ctx context.Context, algorithmID string, hooks []domain.NewPostHook) error {
	return a.crud.CreatePostHooks(ctx, algorithmID, hooks)
}

// GetPostHooks retrieves post hooks for an algorithm
func (a *Adapter) GetPostHooks(ctx context.Context, algorithmID string) ([]domain.PostHook, error) {
	return a.crud.GetPostHooks(ctx, algorithmID)
}

// DeletePostHooks deletes post hooks for an algorithm
func (a *Adapter) DeletePostHooks(ctx context.Context, algorithmID string) error {
	return a.crud.DeletePostHooks(ctx, algorithmID)
}

// Algorithm run operations

// CreateAlgorithmRun creates a new algorithm run
func (a *Adapter) CreateAlgorithmRun(ctx context.Context, newRun domain.NewAlgorithmRun) (domain.AlgorithmRun, error) {
	return a.crud.CreateAlgorithmRun(ctx, newRun)
}

// GetAlgorithmRuns retrieves algorithm runs with search/filter and pagination
func (a *Adapter) GetAlgorithmRuns(ctx context.Context, algorithmID string, search utils.SearchParams, pagination utils.PaginationParams) ([]domain.AlgorithmRun, error) {
	return a.search.GetAlgorithmRuns(ctx, algorithmID, search, pagination)
}

// GetAlgorithmRun retrieves a specific algorithm run by ID
func (a *Adapter) GetAlgorithmRun(ctx context.Context, runID string) (domain.AlgorithmRun, error) {
	return a.crud.GetAlgorithmRun(ctx, runID)
}

// UpdateAlgorithmRun updates an existing algorithm run
func (a *Adapter) UpdateAlgorithmRun(ctx context.Context, runID string, updates domain.AlgorithmRunUpdates) error {
	return a.crud.UpdateAlgorithmRun(ctx, runID, updates)
}

// CancelAlgorithmRun cancels an algorithm run
func (a *Adapter) CancelAlgorithmRun(ctx context.Context, runID string) error {
	return a.crud.CancelAlgorithmRun(ctx, runID)
}

// GetRunsCount returns the total count of runs for an algorithm
func (a *Adapter) GetRunsCount(ctx context.Context, algorithmID string) (int, error) {
	return a.search.GetRunsCount(ctx, algorithmID)
}

// GetRunsCountWithSearch returns the count of runs matching search criteria
func (a *Adapter) GetRunsCountWithSearch(ctx context.Context, algorithmID string, search utils.SearchParams) (int, error) {
	return a.search.GetRunsCountWithSearch(ctx, algorithmID, search)
}

// Output operations

// CreateOutput creates a new output
func (a *Adapter) CreateOutput(ctx context.Context, newOutput domain.NewOutput) (domain.Output, error) {
	return a.crud.CreateOutput(ctx, newOutput)
}

// GetOutputs retrieves outputs for a run
func (a *Adapter) GetOutputs(ctx context.Context, runID string) ([]domain.Output, error) {
	return a.crud.GetOutputs(ctx, runID)
}

// GetOutput retrieves a specific output by ID
func (a *Adapter) GetOutput(ctx context.Context, outputID string) (domain.Output, error) {
	return a.crud.GetOutput(ctx, outputID)
}

// DeleteOutput deletes an output
func (a *Adapter) DeleteOutput(ctx context.Context, outputID string) error {
	return a.crud.DeleteOutput(ctx, outputID)
}
