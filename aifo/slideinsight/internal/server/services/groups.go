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

// GroupService interface defines group-related operations
type GroupService interface {
	CreateGroup(ctx context.Context, name, description string) (*ports.Group, error)
	GetAllGroups(ctx context.Context) ([]ports.Group, error)
	GetGroupsWithPagination(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Group, domain.PaginationInfo, error)
	GetGroupByName(ctx context.Context, name string) (*ports.Group, error)
	GetGroupByID(ctx context.Context, groupID int) (*ports.Group, error)
	DeleteGroup(ctx context.Context, name string) error

	// User-Group methods
	AssignUserToGroup(ctx context.Context, userID int, groupID int) error
	RemoveUserFromGroup(ctx context.Context, userID int, groupID int) error
	GetUserGroups(ctx context.Context, userID int) ([]ports.Group, error)
	GetGroupUsers(ctx context.Context, groupID int) ([]int, error)

	Close()
}

type groupService struct {
	db ports.RBACRepository
	// Generic pagination and search service
	paginatedSearchService *PaginatedSearchService[ports.Group, domain.Group]
}

// convertGroupDBToDomain converts a database Group record to a domain Group model
func convertGroupDBToDomain(record ports.Group) domain.Group {
	return domain.Group{
		Name:        record.Name,
		TenantID:    record.TenantID,
		ShortUID:    record.ShortUID,
		DisplayName: record.Name, // Use name as display name for now
		Description: record.Description,
		CreatedAt:   record.CreatedAt, // Already a string in the database model
		UpdatedAt:   record.UpdatedAt, // Already a string in the database model
	}
}

// NewGroupService creates a new group service
func NewGroupService(db ports.RBACRepository) GroupService {
	// Create the generic paginated search service
	paginatedSearchService := NewPaginatedSearchService(
		db.GetGroupsWithPagination,
		db.GetGroupCount,
		func(ctx context.Context, limit, offset int) ([]ports.Group, error) {
			return db.GetGroupsWithPagination(ctx, utils.SearchParams{}, limit, offset)
		},
		func(ctx context.Context) (int, error) {
			return db.GetGroupCount(ctx, utils.SearchParams{})
		},
		convertGroupDBToDomain,
	)

	return &groupService{
		db:                     db,
		paginatedSearchService: paginatedSearchService,
	}
}

// CreateGroup creates a new group in the system
func (s *groupService) CreateGroup(ctx context.Context, name, description string) (*ports.Group, error) {
	if name == "" {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "group name cannot be empty")
	}

	log.Info("Creating group", "name", name, "description", description)

	// Create the group
	groupID, err := s.db.CreateGroup(ctx, name, description)
	if err != nil {
		log.Error("Failed to create group", "error", err, "name", name)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to create group: %v", err)
	}

	// Retrieve the created group to return it
	group, err := s.db.GetGroupByName(ctx, name)
	if err != nil {
		log.Error("Failed to retrieve created group", "error", err, "name", name)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve group: %v", err)
	}

	log.Info("Group created successfully", "id", groupID, "name", name)
	return group, nil
}

// GetAllGroups retrieves all groups from the system
func (s *groupService) GetAllGroups(ctx context.Context) ([]ports.Group, error) {
	log.Info("Retrieving all groups")

	groups, err := s.db.GetAllGroups(ctx)
	if err != nil {
		log.Error("Failed to retrieve groups", "error", err)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve groups: %v", err)
	}

	log.Info("Successfully retrieved groups", "count", len(groups))
	return groups, nil
}

// GetGroupsWithPagination uses the generic search pattern
func (s *groupService) GetGroupsWithPagination(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Group, domain.PaginationInfo, error) {
	return s.paginatedSearchService.GetWithPaginationAndSearch(ctx, params)
}

// GetGroupByName retrieves a group by its name
func (s *groupService) GetGroupByName(ctx context.Context, name string) (*ports.Group, error) {
	if name == "" {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "group name cannot be empty")
	}

	log.Info("Retrieving group by name", "name", name)

	group, err := s.db.GetGroupByName(ctx, name)
	if err != nil {
		log.Error("Failed to retrieve group", "error", err, "name", name)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve group: %v", err)
	}

	if group == nil {
		log.Warn("Group not found", "name", name)
		return nil, errors.WithDetails(errors.ErrNotFound, "group not found: %s", name)
	}

	log.Info("Group found", "name", name, "id", group.ID)
	return group, nil
}

// GetGroupByID retrieves a group by its ID
func (s *groupService) GetGroupByID(ctx context.Context, groupID int) (*ports.Group, error) {
	if groupID <= 0 {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "group ID must be positive")
	}

	log.Info("Retrieving group by ID", "groupID", groupID)

	group, err := s.db.GetGroupByID(ctx, groupID)
	if err != nil {
		log.Error("Failed to retrieve group", "error", err, "groupID", groupID)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve group: %v", err)
	}

	if group == nil {
		log.Warn("Group not found", "groupID", groupID)
		return nil, errors.WithDetails(errors.ErrNotFound, "group not found with ID: %d", groupID)
	}

	log.Info("Group found", "groupID", groupID, "name", group.Name)
	return group, nil
}

// DeleteGroup deletes a group by its name
func (s *groupService) DeleteGroup(ctx context.Context, name string) error {
	if name == "" {
		return errors.WithDetails(errors.ErrInvalidInput, "group name cannot be empty")
	}

	log.Info("Deleting group", "name", name)

	err := s.db.DeleteGroup(ctx, name)
	if err != nil {
		log.Error("Failed to delete group", "error", err, "name", name)
		return errors.WithDetails(errors.ErrInternal, "failed to delete group: %v", err)
	}

	log.Info("Group deleted successfully", "name", name)
	return nil
}

// User-Group methods

// AssignUserToGroup assigns a user to a group
func (s *groupService) AssignUserToGroup(ctx context.Context, userID int, groupID int) error {
	if userID <= 0 || groupID <= 0 {
		return errors.WithDetails(errors.ErrInvalidInput, "user ID and group ID must be positive")
	}

	log.Info("Assigning user to group", "userID", userID, "groupID", groupID)

	err := s.db.AssignUserToGroup(ctx, userID, groupID)
	if err != nil {
		log.Error("Failed to assign user to group", "error", err, "userID", userID, "groupID", groupID)
		return errors.WithDetails(errors.ErrInternal, "failed to assign user to group: %v", err)
	}

	log.Info("User assigned to group successfully", "userID", userID, "groupID", groupID)
	return nil
}

// RemoveUserFromGroup removes a user from a group
func (s *groupService) RemoveUserFromGroup(ctx context.Context, userID int, groupID int) error {
	if userID <= 0 || groupID <= 0 {
		return errors.WithDetails(errors.ErrInvalidInput, "user ID and group ID must be positive")
	}

	log.Info("Removing user from group", "userID", userID, "groupID", groupID)

	err := s.db.RemoveUserFromGroup(ctx, userID, groupID)
	if err != nil {
		log.Error("Failed to remove user from group", "error", err, "userID", userID, "groupID", groupID)
		return errors.WithDetails(errors.ErrInternal, "failed to remove user from group: %v", err)
	}

	log.Info("User removed from group successfully", "userID", userID, "groupID", groupID)
	return nil
}

// GetUserGroups retrieves all groups a user belongs to
func (s *groupService) GetUserGroups(ctx context.Context, userID int) ([]ports.Group, error) {
	if userID <= 0 {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "user ID must be positive")
	}

	log.Info("Retrieving groups for user", "userID", userID)

	groups, err := s.db.GetUserGroups(ctx, userID)
	if err != nil {
		log.Error("Failed to retrieve user groups", "error", err, "userID", userID)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve user groups: %v", err)
	}

	log.Info("Successfully retrieved user groups", "userID", userID, "count", len(groups))
	return groups, nil
}

// GetGroupUsers retrieves all users in a group
func (s *groupService) GetGroupUsers(ctx context.Context, groupID int) ([]int, error) {
	if groupID <= 0 {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "group ID must be positive")
	}

	log.Info("Retrieving users for group", "groupID", groupID)

	userIDs, err := s.db.GetGroupUsers(ctx, groupID)
	if err != nil {
		log.Error("Failed to retrieve group users", "error", err, "groupID", groupID)
		return nil, errors.WithDetails(errors.ErrInternal, "failed to retrieve group users: %v", err)
	}

	log.Info("Successfully retrieved group users", "groupID", groupID, "count", len(userIDs))
	return userIDs, nil
}

// Close closes the group service
func (s *groupService) Close() {
	// no-op
}
