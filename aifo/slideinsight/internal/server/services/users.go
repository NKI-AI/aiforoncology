// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/auth"
	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

type UserService interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserByUID(ctx context.Context, userUID string) (domain.User, error)
	GetUserByInternalID(ctx context.Context, userID int) (domain.User, error)
	GetInternalUserByEmail(ctx context.Context, email string) (ports.User, error)
	GetUsers(ctx context.Context, pagination utils.PaginationParams) ([]domain.User, domain.PaginationInfo, error)
	GetUsersGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.User, domain.PaginationInfo, error)
	UpdatePassword(ctx context.Context, email, newPassword string) error
	UpdateUser(ctx context.Context, email string, updates domain.UserUpdates) error
	UpdateUserByUID(ctx context.Context, userUID string, updates domain.UserUpdates) error
	DeleteUser(ctx context.Context, userUID string) error
	VerifyPassword(ctx context.Context, email, password string) error
	VerifyPasswordByEmail(ctx context.Context, email, password string) error
	GetUserCount(ctx context.Context) (int, error)
	DeactivateUser(ctx context.Context, email string) error
	ActivateUser(ctx context.Context, email string) error
	Close()
}

type userService struct {
	db ports.Database
	// Generic pagination and search service
	paginatedSearchService *PaginatedSearchService[ports.User, domain.User]
}

// userConversionHelpers provides conversion helpers configured for users (using RFC3339)
var userConversionHelpers = DefaultConversionHelpers()

// convertUserDBToDomain converts a database User record to a domain User model using conversion helpers
func convertUserDBToDomain(record ports.User) domain.User {
	return ConvertDBToDomain(
		record,
		userConversionHelpers,
		convertUserBase,
	)
}

// convertUserBase handles the user conversion with password hiding and multiple time fields
func convertUserBase(record ports.User, helpers *ConversionHelpers) domain.User {
	return domain.User{
		ID:                record.ID,
		TenantID:          record.TenantID,
		TenantUID:         record.TenantUID,
		ShortUID:          record.ShortUID,
		Email:             record.Email,
		FirstName:         record.FirstName,
		LastName:          record.LastName,
		Password:          "", // Don't expose password in listings
		MustResetPassword: record.MustResetPassword,
		IsActive:          record.IsActive,
		EmailVerified:     record.EmailVerified,
		CreatedAt:         helpers.FormatTime(record.CreatedAt),
		UpdatedAt:         helpers.FormatTime(record.UpdatedAt),
	}
}

func NewUserService(db ports.Database) UserService {
	// Create the generic paginated search service
	paginatedSearchService := NewPaginatedSearchService(
		db.LoadAllUsers,
		db.GetUserCount,
		func(ctx context.Context, limit, offset int) ([]ports.User, error) {
			return db.LoadAllUsers(ctx, utils.SearchParams{}, limit, offset)
		},
		func(ctx context.Context) (int, error) {
			return db.GetUserCount(ctx, utils.SearchParams{})
		},
		convertUserDBToDomain,
	)

	return &userService{
		db:                     db,
		paginatedSearchService: paginatedSearchService,
	}
}

// GetUsersGeneric uses the generic search pattern
func (s *userService) GetUsersGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.User, domain.PaginationInfo, error) {
	return s.paginatedSearchService.GetWithPaginationAndSearch(ctx, params)
}

// GetUsers retrieves all users from the database with pagination support
func (s *userService) GetUsers(ctx context.Context, pagination utils.PaginationParams) ([]domain.User, domain.PaginationInfo, error) {
	// Convert PaginationParams to PaginationAndSearchParams for generic service
	params := utils.PaginationAndSearchParams{
		PaginationParams: pagination,
		SearchParams:     utils.SearchParams{}, // Empty search params for simple pagination
	}
	return s.paginatedSearchService.GetWithPaginationAndSearch(ctx, params)
}

func (s *userService) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	if user.Email == "" {
		return domain.User{}, errors.ErrEmailEmpty
	}

	if user.Password == "" {
		return domain.User{}, errors.ErrPasswordEmpty
	}

	if user.TenantUID == "" {
		return domain.User{}, errors.ErrTenantRequired
	}

	// Check if user already exists by email
	_, err := s.db.GetUserByEmail(ctx, user.Email)
	if err == nil {
		return domain.User{}, errors.NewUserAlreadyExistsError(user.Email)
	}

	// Look up tenant by UID to get the internal tenant ID and validate it exists
	dbTenant, err := s.db.GetTenantByUID(ctx, user.TenantUID)
	if err != nil {
		return domain.User{}, errors.NewTenantNotFoundError(user.TenantUID)
	}

	// Generate a new UID for the user
	userUID, err := utils.GenerateFixedShortUID()
	if err != nil {
		return domain.User{}, errors.NewUIDGenerationError(err)
	}

	// Only hash the password if it's not already hashed
	hashedPassword := user.Password
	if !auth.IsHashedPassword(user.Password) {
		var err error
		hashedPassword, err = auth.HashPassword(user.Password)
		if err != nil {
			return domain.User{}, errors.NewPasswordHashError(err)
		}
	}

	dbUser := ports.NewUser{
		TenantID:          dbTenant.ID, // Use the internal tenant ID from database
		Email:             user.Email,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		ShortUID:          userUID,
		Password:          hashedPassword,
		MustResetPassword: user.MustResetPassword,
		IsActive:          user.IsActive,
	}

	err = s.db.CreateUser(ctx, dbUser)
	if err != nil {
		return domain.User{}, err
	}

	// Fetch the created user to get the assigned ID and other fields
	createdUser, err := s.db.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return domain.User{}, errors.WithDetails(errors.ErrInternal, "failed to retrieve created user: %v", err)
	}

	// Return user with proper ID but without exposing the password
	return domain.User{
		ID:                createdUser.ID,
		TenantID:          createdUser.TenantID,
		TenantUID:         createdUser.TenantUID,
		ShortUID:          createdUser.ShortUID,
		Email:             createdUser.Email,
		FirstName:         createdUser.FirstName,
		LastName:          createdUser.LastName,
		Password:          "", // Don't return the password
		MustResetPassword: createdUser.MustResetPassword,
		IsActive:          createdUser.IsActive,
		EmailVerified:     createdUser.EmailVerified,
		CreatedAt:         createdUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         createdUser.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	utils.InfoWithKeyValue(ctx, "Getting user by email", "email", email)
	dbUser, err := s.db.GetUserByEmail(ctx, email)
	utils.InfoWithKeyValue(ctx, "Got user by email", "email", email, "user_id", dbUser.ID)
	if err != nil {
		return domain.User{}, errors.WithDetails(errors.ErrUserNotFound, "failed to get user: %v", err)
	}

	return domain.User{
		ID:                dbUser.ID,
		TenantID:          dbUser.TenantID,
		TenantUID:         dbUser.TenantUID,
		ShortUID:          dbUser.ShortUID,
		Email:             dbUser.Email,
		FirstName:         dbUser.FirstName,
		LastName:          dbUser.LastName,
		Password:          dbUser.Password,
		MustResetPassword: dbUser.MustResetPassword,
		IsActive:          dbUser.IsActive,
		EmailVerified:     dbUser.EmailVerified,
		CreatedAt:         dbUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         dbUser.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *userService) GetUserByUID(ctx context.Context, userUID string) (domain.User, error) {
	utils.InfoWithKeyValue(ctx, "Getting user by UID", "userUID", userUID)
	dbUser, err := s.db.GetUserByUID(ctx, userUID)
	if err != nil {
		return domain.User{}, errors.WithDetails(errors.ErrUserNotFound, "failed to get user: %v", err)
	}

	return domain.User{
		ID:                dbUser.ID,
		TenantID:          dbUser.TenantID,
		TenantUID:         dbUser.TenantUID,
		ShortUID:          dbUser.ShortUID,
		Email:             dbUser.Email,
		FirstName:         dbUser.FirstName,
		LastName:          dbUser.LastName,
		Password:          dbUser.Password,
		MustResetPassword: dbUser.MustResetPassword,
		IsActive:          dbUser.IsActive,
		EmailVerified:     dbUser.EmailVerified,
		CreatedAt:         dbUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         dbUser.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *userService) GetUserByInternalID(ctx context.Context, userID int) (domain.User, error) {
	utils.InfoWithKeyValue(ctx, "Getting user by internal ID", "userID", userID)
	dbUser, err := s.db.GetUserByInternalID(ctx, userID)
	if err != nil {
		return domain.User{}, errors.WithDetails(errors.ErrUserNotFound, "failed to get user: %v", err)
	}

	return domain.User{
		ID:                dbUser.ID,
		TenantID:          dbUser.TenantID,
		TenantUID:         dbUser.TenantUID,
		ShortUID:          dbUser.ShortUID,
		Email:             dbUser.Email,
		FirstName:         dbUser.FirstName,
		LastName:          dbUser.LastName,
		Password:          dbUser.Password,
		MustResetPassword: dbUser.MustResetPassword,
		IsActive:          dbUser.IsActive,
		EmailVerified:     dbUser.EmailVerified,
		CreatedAt:         dbUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         dbUser.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *userService) GetInternalUserByEmail(ctx context.Context, email string) (ports.User, error) {
	return s.db.GetUserByEmail(ctx, email)
}

func (s *userService) UpdatePassword(ctx context.Context, email, newPassword string) error {
	if email == "" {
		return errors.ErrEmailEmpty
	}

	if newPassword == "" {
		return errors.ErrPasswordEmpty
	}

	// Verify user exists
	_, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		return errors.NewUserNotFoundByEmailError(email)
	}

	// Hash the new password
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return errors.NewPasswordHashError(err)
	}

	// Update the password in the database
	err = s.db.UpdateUserPassword(ctx, email, hashedPassword)
	if err != nil {
		return errors.WithDetails(errors.ErrDatabaseUpdate, "failed to update password: %v", err)
	}

	return nil
}

func (s *userService) UpdateUser(ctx context.Context, email string, updates domain.UserUpdates) error {
	if email == "" {
		return errors.ErrEmailEmpty
	}

	// Verify user exists
	_, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		return errors.NewUserNotFoundByEmailError(email)
	}

	// Convert domain updates to ports updates
	portsUpdates := ports.UserUpdates{
		Email:             updates.Email,
		FirstName:         updates.FirstName,
		LastName:          updates.LastName,
		MustResetPassword: updates.MustResetPassword,
		IsActive:          updates.IsActive,
		EmailVerified:     updates.EmailVerified,
	}

	// Update the user in the database
	err = s.db.UpdateUser(ctx, email, portsUpdates)
	if err != nil {
		return errors.WithDetails(errors.ErrDatabaseUpdate, "failed to update user: %v", err)
	}

	return nil
}

func (s *userService) UpdateUserByUID(ctx context.Context, userUID string, updates domain.UserUpdates) error {
	if userUID == "" {
		return errors.ErrUserUIDEmpty
	}

	// Verify user exists
	_, err := s.db.GetUserByUID(ctx, userUID)
	if err != nil {
		return errors.NewUserNotFoundByUIDError(userUID)
	}

	// Convert domain updates to ports updates
	portsUpdates := ports.UserUpdates{
		Email:             updates.Email,
		FirstName:         updates.FirstName,
		LastName:          updates.LastName,
		MustResetPassword: updates.MustResetPassword,
		IsActive:          updates.IsActive,
		EmailVerified:     updates.EmailVerified,
	}

	// Update the user in the database
	err = s.db.UpdateUserByUID(ctx, userUID, portsUpdates)
	if err != nil {
		return errors.WithDetails(errors.ErrDatabaseUpdate, "failed to update user: %v", err)
	}

	return nil
}

func (s *userService) DeleteUser(ctx context.Context, userUID string) error {
	if userUID == "" {
		return errors.ErrUserUIDEmpty
	}

	// Verify user exists
	_, err := s.db.GetUserByUID(ctx, userUID)
	if err != nil {
		return errors.NewUserNotFoundByUIDError(userUID)
	}

	// Delete the user from the database
	err = s.db.DeleteUser(ctx, userUID)
	if err != nil {
		return errors.WithDetails(errors.ErrDatabaseDelete, "failed to delete user: %v", err)
	}

	return nil
}

func (s *userService) VerifyPassword(ctx context.Context, email, password string) error {
	// Get the user with the hashed password
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		// Return generic error to prevent user enumeration
		return auth.ErrPasswordMismatch
	}

	// Verify the password
	err = auth.VerifyPassword(user.Password, password)
	if err != nil {
		return err
	}

	return nil
}

func (s *userService) VerifyPasswordByEmail(ctx context.Context, email, password string) error {
	// Get the user with the hashed password
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		// Return generic error to prevent user enumeration
		return auth.ErrPasswordMismatch
	}

	// Verify the password
	err = auth.VerifyPassword(user.Password, password)
	if err != nil {
		return err
	}

	return nil
}

func (s *userService) GetUserCount(ctx context.Context) (int, error) {
	return s.db.GetUserCount(ctx, utils.SearchParams{})
}

func (s *userService) DeactivateUser(ctx context.Context, email string) error {
	return s.db.DeactivateUser(ctx, email)
}

func (s *userService) ActivateUser(ctx context.Context, email string) error {
	return s.db.ActivateUser(ctx, email)
}

func (s *userService) Close() {
	// no-op
}
