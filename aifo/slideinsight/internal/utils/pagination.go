// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

// Package utils provides common utilities for the SlideInsight application.
//
// # Pagination Utilities
//
// This package provides reusable pagination parameter parsing and validation
// that can be used across all handlers that need paginated results.
//
// Basic usage example:
//
//	func GetItems(service ItemService) fiber.Handler {
//		return func(c *fiber.Ctx) error {
//			// Parse pagination parameters with default values (page=1, limit=100, max=1000)
//			pagination, err := utils.ParsePaginationParams(c)
//			if err != nil {
//				return err // Error already contains proper status code and message
//			}
//
//			items, paginationInfo, err := service.GetItems(c.UserContext(), pagination.Page, pagination.Limit)
//			if err != nil {
//				return middleware.HandleError(c, err)
//			}
//
//			return c.JSON(ItemsResponse{
//				Items:      items,
//				Pagination: paginationInfo,
//			})
//		}
//	}
//
// Custom pagination limits example:
//
//	func GetLargeDataset(service DataService) fiber.Handler {
//		return func(c *fiber.Ctx) error {
//			// Use custom pagination limits for performance-sensitive endpoints
//			customOpts := utils.PaginationOptions{
//				DefaultPage:  1,
//				DefaultLimit: 25,  // Smaller default for heavy operations
//				MaxLimit:     100, // Lower max limit
//			}
//			pagination, err := utils.ParsePaginationParamsWithOptions(c, customOpts)
//			if err != nil {
//				return err
//			}
//
//			// Use pagination.Page and pagination.Limit as needed
//			// ...
//		}
//	}
package utils

import (
	"fmt"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"github.com/gofiber/fiber/v2"
)

// PaginationParams represents pagination query parameters
type PaginationParams struct {
	Page  int `json:"page" query:"page" example:"1" validate:"min=1"`
	Limit int `json:"limit" query:"limit" example:"100" validate:"min=1,max=1000"`
}

// PermissionFilterParams represents permission-based filtering parameters
type PermissionFilterParams struct {
	FilterAccessibleStudies bool `json:"filterAccessibleStudies,omitempty" query:"filterAccessibleStudies"`
	FilterAccessibleCases   bool `json:"filterAccessibleCases,omitempty" query:"filterAccessibleCases"`
	FilterAccessibleSlides  bool `json:"filterAccessibleSlides,omitempty" query:"filterAccessibleSlides"`
}

// SearchParams represents search/filter query parameters
type SearchParams struct {
	Query             string                 `json:"query,omitempty" query:"q" example:"search term"` // General search across multiple fields
	Filters           map[string]string      `json:"filters,omitempty"`                               // Dynamic filters based on configuration
	SortBy            string                 `json:"sortBy,omitempty" query:"sort" example:"name"`    // Sort field
	SortDir           string                 `json:"sortDir,omitempty" query:"dir" example:"asc"`     // Sort direction (asc/desc)
	PermissionFilters PermissionFilterParams `json:"permissionFilters,omitempty"`                     // Permission-based filters
}

// SearchConfig defines which fields are searchable for a specific resource
type SearchConfig struct {
	SearchableFields []string // Fields that can be searched with ?q=
	FilterableFields []string // Fields that can be filtered with ?fieldname=value
	SortableFields   []string // Fields that can be sorted
}

// DefaultCasesSearchConfig returns the search configuration for cases
func DefaultCasesSearchConfig() SearchConfig {
	return SearchConfig{
		SearchableFields: []string{"name", "metadata"},
		FilterableFields: []string{"name", "status", "has_vector_annotations", "has_raster_annotations"},
		SortableFields:   []string{"name", "created_at", "createdAt", "short_uid", "shortId"},
	}
}

// DefaultSlidesSearchConfig returns the search configuration for slides
func DefaultSlidesSearchConfig() SearchConfig {
	return SearchConfig{
		SearchableFields: []string{"slide_name", "slide_uri"},
		FilterableFields: []string{"slide_name"},
		SortableFields:   []string{"slide_name", "created_at", "createdAt", "slide_id", "slideId"},
	}
}

// DefaultTenantsSearchConfig returns the search configuration for tenants
func DefaultTenantsSearchConfig() SearchConfig {
	return SearchConfig{
		SearchableFields: []string{"name"},
		FilterableFields: []string{"name", "status"},
		SortableFields:   []string{"name", "created_at", "createdAt"},
	}
}

// DefaultStudiesSearchConfig returns the search configuration for studies
func DefaultStudiesSearchConfig() SearchConfig {
	return SearchConfig{
		SearchableFields: []string{"name", "description", "metadata"},
		FilterableFields: []string{"name", "description", "is_published", "searchName", "searchDescription", "searchStatus"},
		SortableFields:   []string{"name", "created_at", "createdAt", "short_uid", "shortId"},
	}
}

// DefaultUsersSearchConfig returns the search configuration for users
func DefaultUsersSearchConfig() SearchConfig {
	return SearchConfig{
		SearchableFields: []string{"short_uid", "email", "first_name", "last_name"},
		FilterableFields: []string{"email", "first_name", "last_name", "is_active", "must_reset_password"},
		SortableFields:   []string{"created_at", "short_uid"},
	}
}

// DefaultVectorAnnotationsSearchConfig returns the search configuration for vector annotations
func DefaultVectorAnnotationsSearchConfig() SearchConfig {
	return SearchConfig{
		SearchableFields: []string{"name", "file_uri", "slide_uid"},
		FilterableFields: []string{"name", "file_uri", "slide_uid", "actor_type"},
		SortableFields:   []string{"name", "created_at", "slide_uid", "file_uri"},
	}
}

// DefaultRasterAnnotationsSearchConfig returns the search configuration for raster annotations
func DefaultRasterAnnotationsSearchConfig() SearchConfig {
	return SearchConfig{
		SearchableFields: []string{"name", "file_uri", "slide_uid"},
		FilterableFields: []string{"name", "file_uri", "slide_uid", "actor_type", "format"},
		SortableFields:   []string{"name", "created_at", "createdAt", "slide_uid", "slideUid", "file_uri", "fileUri", "format"},
	}
}

// PaginationAndSearchParams combines pagination and search parameters
type PaginationAndSearchParams struct {
	PaginationParams
	SearchParams
}

// PaginationOptions represents configurable pagination limits
type PaginationOptions struct {
	DefaultPage  int
	DefaultLimit int
	MaxLimit     int
}

// DefaultPaginationOptions returns sensible default pagination options
func DefaultPaginationOptions() PaginationOptions {
	return PaginationOptions{
		DefaultPage:  1,
		DefaultLimit: 100,
		MaxLimit:     1000,
	}
}

// ParsePaginationParams parses and validates pagination parameters from query string
func ParsePaginationParams(c *fiber.Ctx) (PaginationParams, error) {
	return ParsePaginationParamsWithOptions(c, DefaultPaginationOptions())
}

// ParsePaginationParamsWithOptions parses and validates pagination parameters with custom options
func ParsePaginationParamsWithOptions(c *fiber.Ctx, opts PaginationOptions) (PaginationParams, error) {
	params := PaginationParams{
		Page:  c.QueryInt("page", opts.DefaultPage),
		Limit: c.QueryInt("limit", opts.DefaultLimit),
	}

	// Validate pagination parameters
	if params.Page < 1 {
		return params, fiber.NewError(fiber.StatusBadRequest, "page must be greater than 0")
	}
	if params.Limit < 1 || params.Limit > opts.MaxLimit {
		return params, fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("limit must be between 1 and %d", opts.MaxLimit))
	}

	return params, nil
}

// ParseSearchParams parses search/filter parameters from query string with configuration
func ParseSearchParams(c *fiber.Ctx, config SearchConfig) SearchParams {
	search := SearchParams{
		Query:   strings.TrimSpace(c.Query("q")),
		Filters: make(map[string]string),
		SortBy:  strings.TrimSpace(c.Query("sort")),
		SortDir: strings.TrimSpace(c.Query("dir", "asc")), // Default to ascending
		PermissionFilters: PermissionFilterParams{
			FilterAccessibleStudies: c.QueryBool("filterAccessibleStudies", false),
			FilterAccessibleCases:   c.QueryBool("filterAccessibleCases", false),
			FilterAccessibleSlides:  c.QueryBool("filterAccessibleSlides", false),
		},
	}

	// Parse dynamic filters based on configuration
	for _, field := range config.FilterableFields {
		if value := strings.TrimSpace(c.Query(field)); value != "" {
			search.Filters[field] = value
		}
	}

	// Validate sort field against allowed fields
	if search.SortBy != "" {
		validSort := false
		for _, field := range config.SortableFields {
			if search.SortBy == field {
				validSort = true
				break
			}
		}
		if !validSort {
			search.SortBy = "" // Reset to default if invalid
		}
	}

	// Validate sort direction
	if search.SortDir != "" && search.SortDir != "asc" && search.SortDir != "desc" {
		search.SortDir = "asc" // Default to asc if invalid
	}

	return search
}

// ParsePaginationAndSearchParams parses both pagination and search parameters with default cases config
func ParsePaginationAndSearchParams(c *fiber.Ctx) (PaginationAndSearchParams, error) {
	return ParsePaginationAndSearchParamsWithConfig(c, DefaultPaginationOptions(), DefaultCasesSearchConfig())
}

// ParsePaginationAndSearchParamsWithConfig parses both pagination and search parameters with custom options
func ParsePaginationAndSearchParamsWithConfig(c *fiber.Ctx, paginationOpts PaginationOptions, searchConfig SearchConfig) (PaginationAndSearchParams, error) {
	pagination, err := ParsePaginationParamsWithOptions(c, paginationOpts)
	if err != nil {
		return PaginationAndSearchParams{}, err
	}

	search := ParseSearchParams(c, searchConfig)

	return PaginationAndSearchParams{
		PaginationParams: pagination,
		SearchParams:     search,
	}, nil
}

// HasSearch returns true if any search parameters are provided
func (s SearchParams) HasSearch() bool {
	return s.Query != "" || len(s.Filters) > 0
}

// HasSort returns true if sorting parameters are provided
func (s SearchParams) HasSort() bool {
	return s.SortBy != ""
}

// HasPermissionFilters returns true if any permission filters are enabled
func (s SearchParams) HasPermissionFilters() bool {
	return s.PermissionFilters.FilterAccessibleStudies ||
		s.PermissionFilters.FilterAccessibleCases ||
		s.PermissionFilters.FilterAccessibleSlides
}

// ShouldFilterAccessibleStudies returns true if study permission filtering is enabled
func (s SearchParams) ShouldFilterAccessibleStudies() bool {
	return s.PermissionFilters.FilterAccessibleStudies
}

// ShouldFilterAccessibleCases returns true if case permission filtering is enabled
func (s SearchParams) ShouldFilterAccessibleCases() bool {
	return s.PermissionFilters.FilterAccessibleCases
}

// ShouldFilterAccessibleSlides returns true if slide permission filtering is enabled
func (s SearchParams) ShouldFilterAccessibleSlides() bool {
	return s.PermissionFilters.FilterAccessibleSlides
}

// CalculateOffset calculates the offset from pagination parameters
func (p PaginationParams) CalculateOffset() int {
	return (p.Page - 1) * p.Limit
}

// CreatePaginationInfo creates a PaginationInfo struct from pagination parameters and total count
func CreatePaginationInfo(pagination PaginationParams, totalCount int) domain.PaginationInfo {
	totalPages := (totalCount + pagination.Limit - 1) / pagination.Limit // Ceiling division
	if totalPages < 1 {
		totalPages = 1
	}

	return domain.PaginationInfo{
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		Total:      totalCount,
		TotalPages: totalPages,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1,
	}
}

// SearchConfigResponse represents the API response for search configuration
type SearchConfigResponse struct {
	Resource         string   `json:"resource"`
	SearchableFields []string `json:"searchableFields"`
	FilterableFields []string `json:"filterableFields"`
	SortableFields   []string `json:"sortableFields"`
}

// GetAllSearchConfigs returns search configurations for all resources
func GetAllSearchConfigs() map[string]SearchConfigResponse {
	return map[string]SearchConfigResponse{
		"cases": {
			Resource:         "cases",
			SearchableFields: DefaultCasesSearchConfig().SearchableFields,
			FilterableFields: DefaultCasesSearchConfig().FilterableFields,
			SortableFields:   DefaultCasesSearchConfig().SortableFields,
		},
		"slides": {
			Resource:         "slides",
			SearchableFields: DefaultSlidesSearchConfig().SearchableFields,
			FilterableFields: DefaultSlidesSearchConfig().FilterableFields,
			SortableFields:   DefaultSlidesSearchConfig().SortableFields,
		},
		"studies": {
			Resource:         "studies",
			SearchableFields: DefaultStudiesSearchConfig().SearchableFields,
			FilterableFields: DefaultStudiesSearchConfig().FilterableFields,
			SortableFields:   DefaultStudiesSearchConfig().SortableFields,
		},
		"tenants": {
			Resource:         "tenants",
			SearchableFields: DefaultTenantsSearchConfig().SearchableFields,
			FilterableFields: DefaultTenantsSearchConfig().FilterableFields,
			SortableFields:   DefaultTenantsSearchConfig().SortableFields,
		},
		"users": {
			Resource:         "users",
			SearchableFields: DefaultUsersSearchConfig().SearchableFields,
			FilterableFields: DefaultUsersSearchConfig().FilterableFields,
			SortableFields:   DefaultUsersSearchConfig().SortableFields,
		},
		"vector_annotations": {
			Resource:         "vector_annotations",
			SearchableFields: DefaultVectorAnnotationsSearchConfig().SearchableFields,
			FilterableFields: DefaultVectorAnnotationsSearchConfig().FilterableFields,
			SortableFields:   DefaultVectorAnnotationsSearchConfig().SortableFields,
		},
		"raster_annotations": {
			Resource:         "raster_annotations",
			SearchableFields: DefaultRasterAnnotationsSearchConfig().SearchableFields,
			FilterableFields: DefaultRasterAnnotationsSearchConfig().FilterableFields,
			SortableFields:   DefaultRasterAnnotationsSearchConfig().SortableFields,
		},
	}
}

// GetSearchConfigForResource returns the search configuration for a specific resource
func GetSearchConfigForResource(resource string) (SearchConfigResponse, bool) {
	configs := GetAllSearchConfigs()
	config, exists := configs[resource]
	return config, exists
}
