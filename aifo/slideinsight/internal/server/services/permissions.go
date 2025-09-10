// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
)

// PermissionService interface defines permission-related operations
type PermissionService interface {
	CreatePermission(ctx context.Context, name, description string) (*ports.Permission, error)
	CreateTenantPermission(ctx context.Context, tenantID int, name, description string) (*ports.Permission, error)
	GetAllPermissions(ctx context.Context) ([]ports.Permission, error)
	GetPermissionsWithPagination(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Permission, domain.PaginationInfo, error)
	GetPermissionByName(ctx context.Context, name string) (*ports.Permission, error)
	DeletePermission(ctx context.Context, name string) error
	Close()
}

type permissionService struct {
	db ports.RBACRepository
	// Generic pagination and search service
	paginatedSearchService *PaginatedSearchService[ports.Permission, domain.Permission]
}

// convertPermissionDBToDomain converts a database Permission record to a domain Permission model
func convertPermissionDBToDomain(record ports.Permission) domain.Permission {
	return domain.Permission{
		TenantID:    record.TenantID,
		Name:        record.Name,
		DisplayName: record.Name, // Use name as display name for now
		Description: record.Description,
		Category:    "",               // Not available in current schema
		CreatedAt:   record.CreatedAt, // Already a string in the database model
		UpdatedAt:   record.UpdatedAt, // Already a string in the database model
	}
}

// NewPermissionService creates a new permission service
func NewPermissionService(db ports.RBACRepository) PermissionService {
	// Create the paginated search service for permissions
	paginatedSearchService := NewPaginatedSearchService(
		// Data fetcher function with search
		func(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.Permission, error) {
			return db.GetPermissionsWithPagination(ctx, search, limit, offset)
		},
		// Count fetcher function with search
		func(ctx context.Context, search utils.SearchParams) (int, error) {
			return db.GetPermissionCount(ctx, search)
		},
		// Data fetcher function without search (pagination only)
		func(ctx context.Context, limit, offset int) ([]ports.Permission, error) {
			return db.GetPermissionsWithPagination(ctx, utils.SearchParams{}, limit, offset)
		},
		// Count fetcher function without search
		func(ctx context.Context) (int, error) {
			return db.GetPermissionCount(ctx, utils.SearchParams{})
		},
		// Converter function
		convertPermissionDBToDomain,
	)

	return &permissionService{
		db:                     db,
		paginatedSearchService: paginatedSearchService,
	}
}

// GetPermissionsWithPagination uses the generic search pattern
func (s *permissionService) GetPermissionsWithPagination(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Permission, domain.PaginationInfo, error) {
	return s.paginatedSearchService.GetWithPaginationAndSearch(ctx, params)
}

// CreatePermission creates a new global permission
func (s *permissionService) CreatePermission(ctx context.Context, name, description string) (*ports.Permission, error) {
	if name == "" {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "permission name cannot be empty")
	}

	log.Info("Creating permission", "name", name, "description", description)

	// Create the permission
	permissionID, err := s.db.CreatePermission(ctx, name, description)
	if err != nil {
		log.Error("Failed to create permission", "error", err, "name", name)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to create permission: %v", err)
	}

	// Retrieve the created permission to return it
	permission, err := s.db.GetPermissionByName(ctx, name)
	if err != nil {
		log.Error("Failed to retrieve created permission", "error", err, "name", name)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve permission: %v", err)
	}

	log.Info("Permission created successfully", "id", permissionID, "name", name)
	return permission, nil
}

// CreateTenantPermission creates a new tenant-specific permission
func (s *permissionService) CreateTenantPermission(ctx context.Context, tenantID int, name, description string) (*ports.Permission, error) {
	if name == "" {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "permission name cannot be empty")
	}

	log.Info("Creating tenant permission", "tenantID", tenantID, "name", name, "description", description)

	// Create the tenant-specific permission
	permissionID, err := s.db.CreateTenantPermissionIfNotExists(ctx, tenantID, name, description)
	if err != nil {
		log.Error("Failed to create tenant permission", "error", err, "tenantID", tenantID, "name", name)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to create tenant permission: %v", err)
	}

	// Retrieve the created permission to return it
	permission, err := s.db.GetTenantPermissionByName(ctx, tenantID, name)
	if err != nil {
		log.Error("Failed to retrieve created tenant permission", "error", err, "tenantID", tenantID, "name", name)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve tenant permission: %v", err)
	}

	log.Info("Tenant permission created successfully", "id", permissionID, "tenantID", tenantID, "name", name)
	return permission, nil
}

// GetAllPermissions retrieves all permissions from the system
func (s *permissionService) GetAllPermissions(ctx context.Context) ([]ports.Permission, error) {
	log.Info("Retrieving all permissions")

	permissions, err := s.db.GetAllPermissions(ctx)
	if err != nil {
		log.Error("Failed to retrieve permissions", "error", err)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve permissions: %v", err)
	}

	log.Info("Successfully retrieved permissions", "count", len(permissions))
	return permissions, nil
}

// GetPermissionByName retrieves a permission by its name
func (s *permissionService) GetPermissionByName(ctx context.Context, name string) (*ports.Permission, error) {
	if name == "" {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "permission name cannot be empty")
	}

	log.Info("Retrieving permission by name", "name", name)

	permission, err := s.db.GetPermissionByName(ctx, name)
	if err != nil {
		log.Error("Failed to retrieve permission", "error", err, "name", name)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve permission: %v", err)
	}

	if permission == nil {
		log.Warn("Permission not found", "name", name)
		return nil, errors.WithDetails(errors.ErrNotFound, "permission not found: %s", name)
	}

	log.Info("Permission found", "name", name, "id", permission.ID)
	return permission, nil
}

// DeletePermission deletes a permission by its name
func (s *permissionService) DeletePermission(ctx context.Context, name string) error {
	if name == "" {
		return errors.WithDetails(errors.ErrInvalidInput, "permission name cannot be empty")
	}

	log.Info("Deleting permission", "name", name)

	err := s.db.DeletePermission(ctx, name)
	if err != nil {
		log.Error("Failed to delete permission", "error", err, "name", name)
		return errors.WithDetails(errors.ErrInternal, "failed to delete permission: %v", err)
	}

	log.Info("Permission deleted successfully", "name", name)
	return nil
}

// Close closes the permission service
func (s *permissionService) Close() {
	// no-op
}
