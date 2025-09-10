// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// AuthContext holds the authenticated user's tenant and creator information
type AuthContext struct {
	TenantID     int
	CreatorID    int
	Email        string
	IsSuperAdmin bool
}

// BaseService provides common functionality for all services
type BaseService struct {
	db          ports.Database
	userService UserService
}

// NewBaseService creates a new base service
func NewBaseService(db ports.Database) *BaseService {
	return &BaseService{
		db:          db,
		userService: NewUserService(db),
	}
}

// GetAuthContext extracts authentication context from the request context
func (s *BaseService) GetAuthContext(ctx context.Context) (*AuthContext, error) {
	p := middleware.FromContext(ctx)
	if p == nil {
		return nil, errors.WithDetails(errors.ErrInvalidInput, "not authenticated")
	}

	// Get the user ID from the UserService using the email from the principal
	user, err := s.userService.GetInternalUserByEmail(ctx, p.Email)
	if err != nil {
		return nil, errors.WithDetails(errors.ErrInternal, "failed to get user: %v", err)
	}

	isSuper, err := s.db.UserHasGlobalRole(ctx, user.ID, "superadmin")
	if err != nil {
		return nil, errors.WithDetails(errors.ErrInternal, "failed to check user role: %v", err)
	}

	return &AuthContext{
		TenantID:     user.TenantID,
		CreatorID:    user.ID,
		Email:        p.Email,
		IsSuperAdmin: isSuper,
	}, nil
}

// GetAuthContextForSoftDelete extracts authentication context specifically for soft delete operations
func (s *BaseService) GetAuthContextForSoftDelete(ctx context.Context) (int, error) {
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return 0, err
	}
	return authCtx.CreatorID, nil
}

// WithAuthContext is a helper that executes a function with authentication context
// This eliminates the need to manually extract auth context in every service method
func (s *BaseService) WithAuthContext(ctx context.Context, fn func(*AuthContext) error) error {
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return err
	}
	return fn(authCtx)
}

// CreateEntityWithAuth is a generic helper for creating entities that need TenantID and CreatorID
// T is the domain type, N is the ports NewEntity type
type EntityCreator[T any, N any] interface {
	CreateEntity(ctx context.Context, entity N) error
}

// CreateWithAuth creates an entity with automatic authentication context injection
func CreateWithAuth[T any, N any](
	ctx context.Context,
	s *BaseService,
	creator EntityCreator[T, N],
	buildEntity func(*AuthContext) N,
) error {
	return s.WithAuthContext(ctx, func(authCtx *AuthContext) error {
		entity := buildEntity(authCtx)
		return creator.CreateEntity(ctx, entity)
	})
}

// GenerateShortUID is a convenience method for generating UIDs
func (s *BaseService) GenerateShortUID() (string, error) {
	return utils.GenerateFixedShortUID()
}

// GetDB returns the database instance for services that need direct access
func (s *BaseService) GetDB() ports.Database {
	return s.db
}

// FilterByPermission filters records based on the result of a permission check function.
// The check function receives the authenticated user ID and must return true if the user
// is allowed to access the record.
func FilterByPermission[T any](ctx context.Context, baseService *BaseService, records []T, check func(context.Context, int, T) (bool, error)) ([]T, error) {
	authCtx, err := baseService.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	if authCtx.IsSuperAdmin {
		return records, nil
	}
	filtered := make([]T, 0, len(records))
	for _, r := range records {
		ok, err := check(ctx, authCtx.CreatorID, r)
		if err != nil {
			return nil, err
		}
		if ok {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

// CountByPermission returns the number of records the authenticated user is allowed to access.
func CountByPermission[T any](ctx context.Context, baseService *BaseService, records []T, check func(context.Context, int, T) (bool, error)) (int, error) {
	filtered, err := FilterByPermission(ctx, baseService, records, check)
	if err != nil {
		return 0, err
	}
	return len(filtered), nil
}
