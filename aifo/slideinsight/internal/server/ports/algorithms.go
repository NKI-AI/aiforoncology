// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package ports

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// AlgorithmsRepository defines the interface for algorithm-related database operations
type AlgorithmsRepository interface {
	// Algorithm operations
	CreateAlgorithm(ctx context.Context, newAlgorithm domain.NewAlgorithm) (domain.Algorithm, error)
	GetAlgorithms(ctx context.Context, tenantID int, search utils.SearchParams, pagination utils.PaginationParams) ([]domain.Algorithm, error)
	GetAllAlgorithms(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]domain.Algorithm, error)
	GetAlgorithm(ctx context.Context, algorithmID string) (domain.Algorithm, error)
	UpdateAlgorithm(ctx context.Context, algorithmID string, updates domain.AlgorithmUpdates) error
	DeleteAlgorithm(ctx context.Context, algorithmID string) error
	AlgorithmExists(ctx context.Context, algorithmID string) (bool, error)
	GetAlgorithmsCount(ctx context.Context, tenantID int) (int, error)
	GetAllAlgorithmsCount(ctx context.Context) (int, error)
	GetAlgorithmsCountWithSearch(ctx context.Context, tenantID int, search utils.SearchParams) (int, error)
	GetAllAlgorithmsCountWithSearch(ctx context.Context, search utils.SearchParams) (int, error)

	// Result sink operations
	CreateResultSinks(ctx context.Context, algorithmID string, sinks []domain.NewResultSink) error
	GetResultSinks(ctx context.Context, algorithmID string) ([]domain.ResultSink, error)
	DeleteResultSinks(ctx context.Context, algorithmID string) error

	// Post hook operations
	CreatePostHooks(ctx context.Context, algorithmID string, hooks []domain.NewPostHook) error
	GetPostHooks(ctx context.Context, algorithmID string) ([]domain.PostHook, error)
	DeletePostHooks(ctx context.Context, algorithmID string) error

	// Algorithm run operations
	CreateAlgorithmRun(ctx context.Context, newRun domain.NewAlgorithmRun) (domain.AlgorithmRun, error)
	GetAlgorithmRuns(ctx context.Context, algorithmID string, search utils.SearchParams, pagination utils.PaginationParams) ([]domain.AlgorithmRun, error)
	GetAlgorithmRun(ctx context.Context, runID string) (domain.AlgorithmRun, error)
	UpdateAlgorithmRun(ctx context.Context, runID string, updates domain.AlgorithmRunUpdates) error
	CancelAlgorithmRun(ctx context.Context, runID string) error
	GetRunsCount(ctx context.Context, algorithmID string) (int, error)
	GetRunsCountWithSearch(ctx context.Context, algorithmID string, search utils.SearchParams) (int, error)

	// Output operations
	CreateOutput(ctx context.Context, newOutput domain.NewOutput) (domain.Output, error)
	GetOutputs(ctx context.Context, runID string) ([]domain.Output, error)
	GetOutput(ctx context.Context, outputID string) (domain.Output, error)
	DeleteOutput(ctx context.Context, outputID string) error
}
