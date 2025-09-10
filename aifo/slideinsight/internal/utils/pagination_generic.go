// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

// Package utils provides generic pagination and search utilities for the SlideInsight application.
//
// This package extends the basic pagination utilities with Go generics to provide
// reusable patterns for paginated and searchable resource APIs.
package utils

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ResponseBuilder is a function type that builds a response from resources and pagination info
type ResponseBuilder[T any, R any] func(resources []T, pagination domain.PaginationInfo) R

// ServiceFunc represents a service function that returns paginated resources with search
type ServiceFunc[T any] func(ctx context.Context, params PaginationAndSearchParams) ([]T, domain.PaginationInfo, error)

// PaginationOnlyServiceFunc represents a service function that returns paginated resources without search
type PaginationOnlyServiceFunc[T any] func(ctx context.Context, params PaginationParams) ([]T, domain.PaginationInfo, error)

// CountServiceFunc represents a service function that returns total count
type CountServiceFunc func(ctx context.Context) (int, error)

// sendError sends a JSON error response using fiber
func sendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"error":   "Request failed",
		"message": message,
	})
}

// GetPaginatedResourcesHandler creates a generic handler for paginated resources with search support
func GetPaginatedResourcesHandler[T any, R any](
	searchConfig SearchConfig,
	serviceFunc ServiceFunc[T],
	responseBuilder ResponseBuilder[T, R],
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		params, err := ParsePaginationAndSearchParamsWithConfig(c, DefaultPaginationOptions(), searchConfig)
		if err != nil {
			return err // Error already contains proper status code and message
		}

		resources, paginationInfo, err := serviceFunc(c.UserContext(), params)
		if err != nil {
			// Use structured logging instead of fiber log
			logger := GetLoggerFromFiber(c)
			logger.Error("GetPaginatedResources failed",
				zap.Error(err),
				zap.Any("search", params.SearchParams),
			)
			return sendError(c, fiber.StatusInternalServerError, "failed to load resources")
		}

		return c.JSON(responseBuilder(resources, paginationInfo))
	}
}

// GetPaginatedOnlyHandler creates a generic handler for pagination-only resources (no search)
func GetPaginatedOnlyHandler[T any, R any](
	paginationOpts PaginationOptions,
	serviceFunc PaginationOnlyServiceFunc[T],
	responseBuilder ResponseBuilder[T, R],
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		params, err := ParsePaginationParamsWithOptions(c, paginationOpts)
		if err != nil {
			return err // Error already contains proper status code and message
		}

		resources, paginationInfo, err := serviceFunc(c.UserContext(), params)
		if err != nil {
			// Use structured logging instead of fiber log
			logger := GetLoggerFromFiber(c)
			logger.Error("GetPaginatedOnly failed", zap.Error(err))
			return sendError(c, fiber.StatusInternalServerError, "failed to load resources")
		}

		return c.JSON(responseBuilder(resources, paginationInfo))
	}
}

// GetCountHandler creates a generic handler for count endpoints
func GetCountHandler(serviceFunc CountServiceFunc) fiber.Handler {
	return func(c *fiber.Ctx) error {
		count, err := serviceFunc(c.UserContext())
		if err != nil {
			// Use structured logging instead of fiber log
			logger := GetLoggerFromFiber(c)
			logger.Error("GetCount failed", zap.Error(err))
			return sendError(c, fiber.StatusInternalServerError, "failed to get count")
		}

		return c.JSON(fiber.Map{
			"count": count,
		})
	}
}

// PaginatedServiceManager provides a fluent interface for configuring paginated services
type PaginatedServiceManager[T any] struct {
	searchConfig       SearchConfig
	paginationOpts     PaginationOptions
	serviceFunc        ServiceFunc[T]
	paginationOnlyFunc PaginationOnlyServiceFunc[T]
	countFunc          CountServiceFunc
}

// NewPaginatedServiceManager creates a new service manager
func NewPaginatedServiceManager[T any]() *PaginatedServiceManager[T] {
	return &PaginatedServiceManager[T]{
		paginationOpts: DefaultPaginationOptions(),
	}
}

// WithSearchConfig sets the search configuration
func (m *PaginatedServiceManager[T]) WithSearchConfig(config SearchConfig) *PaginatedServiceManager[T] {
	m.searchConfig = config
	return m
}

// WithPaginationOptions sets pagination options
func (m *PaginatedServiceManager[T]) WithPaginationOptions(opts PaginationOptions) *PaginatedServiceManager[T] {
	m.paginationOpts = opts
	return m
}

// WithSearchService sets the service function that supports search
func (m *PaginatedServiceManager[T]) WithSearchService(serviceFunc ServiceFunc[T]) *PaginatedServiceManager[T] {
	m.serviceFunc = serviceFunc
	return m
}

// WithPaginationOnlyService sets the service function for pagination-only
func (m *PaginatedServiceManager[T]) WithPaginationOnlyService(serviceFunc PaginationOnlyServiceFunc[T]) *PaginatedServiceManager[T] {
	m.paginationOnlyFunc = serviceFunc
	return m
}

// WithCountService sets the count service function
func (m *PaginatedServiceManager[T]) WithCountService(countFunc CountServiceFunc) *PaginatedServiceManager[T] {
	m.countFunc = countFunc
	return m
}

// GetSearchHandler returns a handler with search support using any response type
func (m *PaginatedServiceManager[T]) GetSearchHandler(responseBuilder ResponseBuilder[T, any]) fiber.Handler {
	if m.serviceFunc == nil {
		panic("search service function not set")
	}
	return GetPaginatedResourcesHandler(m.searchConfig, m.serviceFunc, responseBuilder)
}

// GetPaginationHandler returns a handler with pagination-only support using any response type
func (m *PaginatedServiceManager[T]) GetPaginationHandler(responseBuilder ResponseBuilder[T, any]) fiber.Handler {
	if m.paginationOnlyFunc == nil {
		panic("pagination-only service function not set")
	}
	return GetPaginatedOnlyHandler(m.paginationOpts, m.paginationOnlyFunc, responseBuilder)
}

// GetCountHandler returns a count handler
func (m *PaginatedServiceManager[T]) GetCountHandler() fiber.Handler {
	if m.countFunc == nil {
		panic("count service function not set")
	}
	return GetCountHandler(m.countFunc)
}

// Resource-specific pagination options

// DefaultUsersOptions returns pagination options optimized for users
func DefaultUsersOptions() PaginationOptions {
	return PaginationOptions{
		DefaultPage:  1,
		DefaultLimit: 50,  // Smaller default for user listings
		MaxLimit:     200, // Lower max limit for user data
	}
}

// DefaultStudiesOptions returns pagination options optimized for studies
func DefaultStudiesOptions() PaginationOptions {
	return PaginationOptions{
		DefaultPage:  1,
		DefaultLimit: 25,  // Smaller default for potentially complex study data
		MaxLimit:     100, // Conservative max limit
	}
}

// DefaultTenantsOptions returns pagination options optimized for tenants
func DefaultTenantsOptions() PaginationOptions {
	return PaginationOptions{
		DefaultPage:  1,
		DefaultLimit: 20, // Small default - typically few tenants
		MaxLimit:     50, // Small max limit
	}
}
