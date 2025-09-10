// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

package services

import (
	"context"
	"fmt"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
)

// PaginatedSearchService provides generic pagination and search functionality
type PaginatedSearchService[DBRecord any, DomainModel any] struct {
	// Database operations
	loadWithSearch     func(ctx context.Context, search utils.SearchParams, limit, offset int) ([]DBRecord, error)
	countWithSearch    func(ctx context.Context, search utils.SearchParams) (int, error)
	loadWithPagination func(ctx context.Context, limit, offset int) ([]DBRecord, error)
	count              func(ctx context.Context) (int, error)

	// Conversion function from DB record to domain model
	toDomainModel func(record DBRecord) DomainModel
}

// NewPaginatedSearchService creates a new paginated search service
func NewPaginatedSearchService[DBRecord any, DomainModel any](
	loadWithSearch func(ctx context.Context, search utils.SearchParams, limit, offset int) ([]DBRecord, error),
	countWithSearch func(ctx context.Context, search utils.SearchParams) (int, error),
	loadWithPagination func(ctx context.Context, limit, offset int) ([]DBRecord, error),
	count func(ctx context.Context) (int, error),
	toDomainModel func(record DBRecord) DomainModel,
) *PaginatedSearchService[DBRecord, DomainModel] {
	return &PaginatedSearchService[DBRecord, DomainModel]{
		loadWithSearch:     loadWithSearch,
		countWithSearch:    countWithSearch,
		loadWithPagination: loadWithPagination,
		count:              count,
		toDomainModel:      toDomainModel,
	}
}

// GetWithPaginationAndSearch retrieves resources with pagination and search support
func (s *PaginatedSearchService[DBRecord, DomainModel]) GetWithPaginationAndSearch(
	ctx context.Context,
	params utils.PaginationAndSearchParams,
) ([]DomainModel, domain.PaginationInfo, error) {
	// Set default values if not provided
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 100 // Default limit
	}

	// Calculate offset
	offset := (params.Page - 1) * params.Limit

	// Get total count for pagination info (with search filters applied)
	totalCount, err := s.countWithSearch(ctx, params.SearchParams)
	if err != nil {
		return nil, domain.PaginationInfo{}, fmt.Errorf("failed to get total count with search: %w", err)
	}

	// Get resources with pagination and search
	dbRecords, err := s.loadWithSearch(ctx, params.SearchParams, params.Limit, offset)
	if err != nil {
		log.Error("Failed to load resources with search", "error", err, "search", params.SearchParams)
		return nil, domain.PaginationInfo{}, fmt.Errorf("failed to load resources with search: %w", err)
	}

	// Convert database records to domain models
	resources := make([]DomainModel, 0, len(dbRecords))
	for _, record := range dbRecords {
		domainModel := s.toDomainModel(record)
		resources = append(resources, domainModel)
	}

	// Calculate pagination info
	paginationInfo := utils.CreatePaginationInfo(params.PaginationParams, totalCount)

	return resources, paginationInfo, nil
}

// GetWithPagination retrieves resources with pagination only (no search)
func (s *PaginatedSearchService[DBRecord, DomainModel]) GetWithPagination(
	ctx context.Context,
	params utils.PaginationParams,
) ([]DomainModel, domain.PaginationInfo, error) {
	// Set default values if not provided
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 100 // Default limit
	}

	// Calculate offset
	offset := (params.Page - 1) * params.Limit

	// Get total count for pagination info
	totalCount, err := s.count(ctx)
	if err != nil {
		return nil, domain.PaginationInfo{}, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get resources with pagination
	dbRecords, err := s.loadWithPagination(ctx, params.Limit, offset)
	if err != nil {
		log.Error("Failed to load resources", "error", err)
		return nil, domain.PaginationInfo{}, fmt.Errorf("failed to load resources: %w", err)
	}

	// Convert database records to domain models
	resources := make([]DomainModel, 0, len(dbRecords))
	for _, record := range dbRecords {
		domainModel := s.toDomainModel(record)
		resources = append(resources, domainModel)
	}

	// Calculate pagination info
	paginationInfo := utils.CreatePaginationInfo(params, totalCount)

	return resources, paginationInfo, nil
}

// GetCount retrieves the total count of resources
func (s *PaginatedSearchService[DBRecord, DomainModel]) GetCount(ctx context.Context) (int, error) {
	return s.count(ctx)
}
