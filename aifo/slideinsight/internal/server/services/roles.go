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

// RoleService interface defines role-related operations
type RoleService interface {
	CreateRole(ctx context.Context, name, description string) (*ports.Role, error)
	GetAllRoles(ctx context.Context) ([]ports.Role, error)
	GetRolesWithPagination(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Role, domain.PaginationInfo, error)
	GetRoleByName(ctx context.Context, name string) (*ports.Role, error)
	GetRoleByID(ctx context.Context, roleID int) (*ports.Role, error)
	DeleteRole(ctx context.Context, name string) error

	// Role-Permission methods
	AssignPermissionToRole(ctx context.Context, roleID int, permissionID int) error
	RemovePermissionFromRole(ctx context.Context, roleID int, permissionID int) error
	GetRolePermissions(ctx context.Context, roleID int) ([]ports.Permission, error)

	// User-Role methods
	AssignRoleToUser(ctx context.Context, userID int, roleID int, tenantID *int) error
	RemoveRoleFromUser(ctx context.Context, userID int, roleID int, tenantID *int) error
	GetUserRoles(ctx context.Context, userID int) ([]ports.UserRole, error)
	GetRoleUsers(ctx context.Context, roleID int) ([]ports.UserRole, error)

	Close()
}

type roleService struct {
	db ports.RBACRepository
	// Generic pagination and search service
	paginatedSearchService *PaginatedSearchService[ports.Role, domain.Role]
}

// convertRoleDBToDomain converts a database Role record to a domain Role model
func convertRoleDBToDomain(record ports.Role) domain.Role {
	return domain.Role{
		Name:        record.Name,
		TenantID:    record.TenantID,
		ShortUID:    record.ShortUID,
		DisplayName: record.Name, // Use name as display name for now
		Description: record.Description,
		CreatedAt:   record.CreatedAt, // Already a string in the database model
		UpdatedAt:   record.UpdatedAt, // Already a string in the database model
	}
}

// NewRoleService creates a new role service
func NewRoleService(db ports.RBACRepository) RoleService {
	// Create the generic paginated search service
	paginatedSearchService := NewPaginatedSearchService(
		db.GetRolesWithPagination,
		db.GetRoleCount,
		func(ctx context.Context, limit, offset int) ([]ports.Role, error) {
			return db.GetRolesWithPagination(ctx, utils.SearchParams{}, limit, offset)
		},
		func(ctx context.Context) (int, error) {
			return db.GetRoleCount(ctx, utils.SearchParams{})
		},
		convertRoleDBToDomain,
	)

	return &roleService{
		db:                     db,
		paginatedSearchService: paginatedSearchService,
	}
}

// GetRolesWithPagination uses the generic search pattern
func (s *roleService) GetRolesWithPagination(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Role, domain.PaginationInfo, error) {
	return s.paginatedSearchService.GetWithPaginationAndSearch(ctx, params)
}

// CreateRole creates a new role in the system
func (s *roleService) CreateRole(ctx context.Context, name, description string) (*ports.Role, error) {
	if name == "" {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "role name cannot be empty")
	}

	log.Info("Creating role", "name", name, "description", description)

	// Create the role
	roleID, err := s.db.CreateRole(ctx, name, description)
	if err != nil {
		log.Error("Failed to create role", "error", err, "name", name)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to create role: %v", err)
	}

	// Retrieve the created role to return it
	role, err := s.db.GetRoleByName(ctx, name)
	if err != nil {
		log.Error("Failed to retrieve created role", "error", err, "name", name)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve role: %v", err)
	}

	log.Info("Role created successfully", "id", roleID, "name", name)
	return role, nil
}

// GetAllRoles retrieves all roles from the system
func (s *roleService) GetAllRoles(ctx context.Context) ([]ports.Role, error) {
	log.Info("Retrieving all roles")

	roles, err := s.db.GetAllRoles(ctx)
	if err != nil {
		log.Error("Failed to retrieve roles", "error", err)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve roles: %v", err)
	}

	log.Info("Successfully retrieved roles", "count", len(roles))
	return roles, nil
}

// GetRoleByName retrieves a role by its name
func (s *roleService) GetRoleByName(ctx context.Context, name string) (*ports.Role, error) {
	if name == "" {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "role name cannot be empty")
	}

	log.Info("Retrieving role by name", "name", name)

	role, err := s.db.GetRoleByName(ctx, name)
	if err != nil {
		log.Error("Failed to retrieve role", "error", err, "name", name)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve role: %v", err)
	}

	if role == nil {
		log.Warn("Role not found", "name", name)
		return nil, errors.WithDetails(errors.ErrNotFound, "role not found: %s", name)
	}

	log.Info("Role found", "name", name, "id", role.ID)
	return role, nil
}

// GetRoleByID retrieves a role by its ID
func (s *roleService) GetRoleByID(ctx context.Context, roleID int) (*ports.Role, error) {
	if roleID <= 0 {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "role ID must be positive")
	}

	log.Info("Retrieving role by ID", "roleID", roleID)

	role, err := s.db.GetRoleByID(ctx, roleID)
	if err != nil {
		log.Error("Failed to retrieve role", "error", err, "roleID", roleID)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve role: %v", err)
	}

	if role == nil {
		log.Warn("Role not found", "roleID", roleID)
		return nil, errors.WithDetails(errors.ErrNotFound, "role not found with ID: %d", roleID)
	}

	log.Info("Role found", "roleID", roleID, "name", role.Name)
	return role, nil
}

// DeleteRole deletes a role by its name
func (s *roleService) DeleteRole(ctx context.Context, name string) error {
	if name == "" {
		return errors.WithDetails(errors.ErrInvalidInput, "role name cannot be empty")
	}

	log.Info("Deleting role", "name", name)

	err := s.db.DeleteRole(ctx, name)
	if err != nil {
		log.Error("Failed to delete role", "error", err, "name", name)
		return errors.WithDetails(errors.ErrInternal, "failed to delete role: %v", err)
	}

	log.Info("Role deleted successfully", "name", name)
	return nil
}

// Role-Permission methods

// AssignPermissionToRole assigns a permission to a role
func (s *roleService) AssignPermissionToRole(ctx context.Context, roleID int, permissionID int) error {
	if roleID <= 0 || permissionID <= 0 {
		return errors.WithDetails(errors.ErrInvalidInput, "role ID and permission ID must be positive")
	}

	log.Info("Assigning permission to role", "roleID", roleID, "permissionID", permissionID)

	err := s.db.AssignPermissionToRole(ctx, roleID, permissionID)
	if err != nil {
		log.Error("Failed to assign permission to role", "error", err, "roleID", roleID, "permissionID", permissionID)
		return errors.WithDetails(errors.ErrInternal, "failed to assign permission to role: %v", err)
	}

	log.Info("Permission assigned to role successfully", "roleID", roleID, "permissionID", permissionID)
	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (s *roleService) RemovePermissionFromRole(ctx context.Context, roleID int, permissionID int) error {
	if roleID <= 0 || permissionID <= 0 {
		return errors.WithDetails(errors.ErrInvalidInput, "role ID and permission ID must be positive")
	}

	log.Info("Removing permission from role", "roleID", roleID, "permissionID", permissionID)

	err := s.db.RemovePermissionFromRole(ctx, roleID, permissionID)
	if err != nil {
		log.Error("Failed to remove permission from role", "error", err, "roleID", roleID, "permissionID", permissionID)
		return errors.WithDetails(errors.ErrInternal, "failed to remove permission from role: %v", err)
	}

	log.Info("Permission removed from role successfully", "roleID", roleID, "permissionID", permissionID)
	return nil
}

// GetRolePermissions retrieves all permissions assigned to a role
func (s *roleService) GetRolePermissions(ctx context.Context, roleID int) ([]ports.Permission, error) {
	if roleID <= 0 {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "role ID must be positive")
	}

	log.Info("Retrieving permissions for role", "roleID", roleID)

	permissions, err := s.db.GetRolePermissions(ctx, roleID)
	if err != nil {
		log.Error("Failed to retrieve role permissions", "error", err, "roleID", roleID)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve role permissions: %v", err)
	}

	log.Info("Successfully retrieved role permissions", "roleID", roleID, "count", len(permissions))
	return permissions, nil
}

// User-Role methods

// AssignRoleToUser assigns a role to a user
func (s *roleService) AssignRoleToUser(ctx context.Context, userID int, roleID int, tenantID *int) error {
	if userID <= 0 || roleID <= 0 {
		return errors.WithDetails(errors.ErrInvalidInput, "user ID and role ID must be positive")
	}

	log.Info("Assigning role to user", "userID", userID, "roleID", roleID, "tenantID", tenantID)

	err := s.db.AssignRoleToUser(ctx, userID, roleID, tenantID)
	if err != nil {
		log.Error("Failed to assign role to user", "error", err, "userID", userID, "roleID", roleID)
		return errors.WithDetails(errors.ErrInternal, "failed to assign role to user: %v", err)
	}

	log.Info("Role assigned to user successfully", "userID", userID, "roleID", roleID)
	return nil
}

// RemoveRoleFromUser removes a role from a user
func (s *roleService) RemoveRoleFromUser(ctx context.Context, userID int, roleID int, tenantID *int) error {
	if userID <= 0 || roleID <= 0 {
		return errors.WithDetails(errors.ErrInvalidInput, "user ID and role ID must be positive")
	}

	log.Info("Removing role from user", "userID", userID, "roleID", roleID, "tenantID", tenantID)

	err := s.db.RemoveRoleFromUser(ctx, userID, roleID, tenantID)
	if err != nil {
		log.Error("Failed to remove role from user", "error", err, "userID", userID, "roleID", roleID)
		return errors.WithDetails(errors.ErrInternal, "failed to remove role from user: %v", err)
	}

	log.Info("Role removed from user successfully", "userID", userID, "roleID", roleID)
	return nil
}

// GetUserRoles retrieves all roles assigned to a user
func (s *roleService) GetUserRoles(ctx context.Context, userID int) ([]ports.UserRole, error) {
	if userID <= 0 {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "user ID must be positive")
	}

	log.Info("Retrieving roles for user", "userID", userID)

	userRoles, err := s.db.GetUserRoles(ctx, userID)
	if err != nil {
		log.Error("Failed to retrieve user roles", "error", err, "userID", userID)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve user roles: %v", err)
	}

	log.Info("Successfully retrieved user roles", "userID", userID, "count", len(userRoles))
	return userRoles, nil
}

// GetRoleUsers retrieves all users assigned to a specific role
func (s *roleService) GetRoleUsers(ctx context.Context, roleID int) ([]ports.UserRole, error) {
	if roleID <= 0 {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "role ID must be positive")
	}

	log.Info("Retrieving users for role", "roleID", roleID)

	roleUsers, err := s.db.GetRoleUsers(ctx, roleID)
	if err != nil {
		log.Error("Failed to retrieve role users", "error", err, "roleID", roleID)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve role users: %v", err)
	}

	log.Info("Successfully retrieved role users", "roleID", roleID, "count", len(roleUsers))
	return roleUsers, nil
}

// Close closes the role service
func (s *roleService) Close() {
	// no-op
}
