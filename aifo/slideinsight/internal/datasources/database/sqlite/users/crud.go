// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package users

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
)

// CrudService handles basic CRUD operations for users
type CrudService struct {
	db *sql.DB
}

// NewCrudService creates a new CRUD service instance
func NewCrudService(db *sql.DB) *CrudService {
	return &CrudService{db: db}
}

// CreateUser adds a new user to the database
func (s *CrudService) CreateUser(ctx context.Context, newUser ports.NewUser) error {
	_, err := s.db.Exec("INSERT INTO users (tenant_id, short_uid, email, first_name, last_name, password, must_reset_password, is_active, email_verified) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		newUser.TenantID, newUser.ShortUID, newUser.Email, newUser.FirstName, newUser.LastName, newUser.Password, newUser.MustResetPassword, newUser.IsActive, newUser.EmailVerified)
	if err != nil {
		return errors.NewDatabaseInsertError("user", err)
	}
	return nil
}

// GetUserByEmail retrieves a specific user by email address
func (s *CrudService) GetUserByEmail(ctx context.Context, email string) (ports.User, error) {
	var user ports.User

	query := `
		SELECT u.id, u.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, u.short_uid, u.email, u.first_name, u.last_name, u.password, u.must_reset_password, u.is_active, u.email_verified, u.password_changed_at, u.created_at, u.updated_at 
		FROM users u
		LEFT JOIN tenants t ON u.tenant_id = t.id
		WHERE u.email = ?`

	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.TenantID, &user.TenantUID, &user.ShortUID, &user.Email, &user.FirstName, &user.LastName, &user.Password, &user.MustResetPassword, &user.IsActive, &user.EmailVerified, &user.PasswordChangedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.User{}, errors.NewUserNotFoundByEmailError(email)
		}
		return ports.User{}, errors.NewDatabaseQueryError("user", err)
	}
	return user, nil
}

// GetUserByUID retrieves a specific user by its UID
func (s *CrudService) GetUserByUID(ctx context.Context, userUID string) (ports.User, error) {
	var user ports.User

	query := `
		SELECT u.id, u.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, u.short_uid, u.email, u.first_name, u.last_name, u.password, u.must_reset_password, u.is_active, u.email_verified, u.password_changed_at, u.created_at, u.updated_at 
		FROM users u
		LEFT JOIN tenants t ON u.tenant_id = t.id
		WHERE u.short_uid = ?`

	err := s.db.QueryRow(query, userUID).Scan(
		&user.ID, &user.TenantID, &user.TenantUID, &user.ShortUID, &user.Email, &user.FirstName, &user.LastName, &user.Password, &user.MustResetPassword, &user.IsActive, &user.EmailVerified, &user.PasswordChangedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.User{}, errors.NewUserNotFoundByUIDError(userUID)
		}
		return ports.User{}, errors.NewDatabaseQueryError("user", err)
	}
	return user, nil
}

// GetUserByInternalID retrieves a specific user by its internal database ID
func (s *CrudService) GetUserByInternalID(ctx context.Context, userID int) (ports.User, error) {
	var user ports.User

	query := `
		SELECT u.id, u.tenant_id, COALESCE(t.short_uid, '') as tenant_uid, u.short_uid, u.email, u.first_name, u.last_name, u.password, u.must_reset_password, u.is_active, u.email_verified, u.password_changed_at, u.created_at, u.updated_at 
		FROM users u
		LEFT JOIN tenants t ON u.tenant_id = t.id
		WHERE u.id = ?`

	err := s.db.QueryRow(query, userID).Scan(
		&user.ID, &user.TenantID, &user.TenantUID, &user.ShortUID, &user.Email, &user.FirstName, &user.LastName, &user.Password, &user.MustResetPassword, &user.IsActive, &user.EmailVerified, &user.PasswordChangedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.User{}, errors.WithDetails(errors.ErrUserNotFound, "user with internal ID %d not found", userID)
		}
		return ports.User{}, errors.NewDatabaseQueryError("user", err)
	}
	return user, nil
}

// UpdateUserPassword updates the password for a user with the specified email
func (s *CrudService) UpdateUserPassword(ctx context.Context, email string, hashedPassword string) error {
	result, err := s.db.Exec("UPDATE users SET password = ? WHERE email = ?", hashedPassword, email)
	if err != nil {
		return errors.NewDatabaseUpdateError("user password", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.NewUserNotFoundByEmailError(email)
	}

	return nil
}

// UpdateUser updates user information (excluding password) for a user with the specified email
func (s *CrudService) UpdateUser(ctx context.Context, email string, updates ports.UserUpdates) error {
	var setParts []string
	var args []interface{}

	// Build SET clause dynamically based on provided updates
	if updates.Email != nil {
		setParts = append(setParts, "email = ?")
		args = append(args, *updates.Email)
	}
	if updates.FirstName != nil {
		setParts = append(setParts, "first_name = ?")
		args = append(args, *updates.FirstName)
	}
	if updates.LastName != nil {
		setParts = append(setParts, "last_name = ?")
		args = append(args, *updates.LastName)
	}
	if updates.MustResetPassword != nil {
		setParts = append(setParts, "must_reset_password = ?")
		args = append(args, *updates.MustResetPassword)
	}
	if updates.IsActive != nil {
		setParts = append(setParts, "is_active = ?")
		args = append(args, *updates.IsActive)
	}
	if updates.EmailVerified != nil {
		setParts = append(setParts, "email_verified = ?")
		args = append(args, *updates.EmailVerified)
	}

	if len(setParts) == 0 {
		return errors.ErrNoFieldsToUpdate
	}

	// Add email to args for WHERE clause
	args = append(args, email)

	query := fmt.Sprintf("UPDATE users SET %s WHERE email = ?", strings.Join(setParts, ", "))
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return errors.NewDatabaseUpdateError("user", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.NewUserNotFoundByEmailError(email)
	}

	return nil
}

// UpdateUserByUID updates user information (excluding password) for a user with the specified UID
func (s *CrudService) UpdateUserByUID(ctx context.Context, userUID string, updates ports.UserUpdates) error {
	var setParts []string
	var args []interface{}

	// Build SET clause dynamically based on provided updates
	if updates.Email != nil {
		setParts = append(setParts, "email = ?")
		args = append(args, *updates.Email)
	}
	if updates.FirstName != nil {
		setParts = append(setParts, "first_name = ?")
		args = append(args, *updates.FirstName)
	}
	if updates.LastName != nil {
		setParts = append(setParts, "last_name = ?")
		args = append(args, *updates.LastName)
	}
	if updates.MustResetPassword != nil {
		setParts = append(setParts, "must_reset_password = ?")
		args = append(args, *updates.MustResetPassword)
	}
	if updates.IsActive != nil {
		setParts = append(setParts, "is_active = ?")
		args = append(args, *updates.IsActive)
	}
	if updates.EmailVerified != nil {
		setParts = append(setParts, "email_verified = ?")
		args = append(args, *updates.EmailVerified)
	}

	if len(setParts) == 0 {
		return errors.ErrNoFieldsToUpdate
	}

	// Add userUID to args for WHERE clause
	args = append(args, userUID)

	query := fmt.Sprintf("UPDATE users SET %s WHERE short_uid = ?", strings.Join(setParts, ", "))
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return errors.NewDatabaseUpdateError("user", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.NewUserNotFoundByUIDError(userUID)
	}

	return nil
}

// CheckUserDependencies verifies if a user has any dependent entities that would prevent deletion
func (s *CrudService) CheckUserDependencies(ctx context.Context, userUID string) error {
	// First get the user ID
	var userID int
	err := s.db.QueryRow("SELECT id FROM users WHERE short_uid = ? AND deactivated_at IS NULL", userUID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NewUserNotFoundByUIDError(userUID)
		}
		return errors.NewDatabaseQueryError("user", err)
	}

	// Check for studies created by this user
	var studyCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM studies WHERE creator_id = ? AND deleted_at IS NULL", userID).Scan(&studyCount)
	if err != nil {
		return errors.NewDatabaseQueryError("user studies", err)
	}
	if studyCount > 0 {
		return errors.NewUserHasDependenciesError(userUID, fmt.Sprintf("user has %d active studies", studyCount))
	}

	// Check for cases created by this user
	var caseCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM cases WHERE creator_id = ? AND deleted_at IS NULL", userID).Scan(&caseCount)
	if err != nil {
		return errors.NewDatabaseQueryError("user cases", err)
	}
	if caseCount > 0 {
		return errors.NewUserHasDependenciesError(userUID, fmt.Sprintf("user has %d active cases", caseCount))
	}

	// Check for vector annotations created by this user
	var vectorAnnotationCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM vector_annotations WHERE creator_id = ? AND deleted_at IS NULL", userID).Scan(&vectorAnnotationCount)
	if err != nil {
		return errors.NewDatabaseQueryError("user vector annotations", err)
	}
	if vectorAnnotationCount > 0 {
		return errors.NewUserHasDependenciesError(userUID, fmt.Sprintf("user has %d active vector annotations", vectorAnnotationCount))
	}

	// Check for raster annotations created by this user
	var rasterAnnotationCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM raster_annotations WHERE creator_id = ? AND deleted_at IS NULL", userID).Scan(&rasterAnnotationCount)
	if err != nil {
		return errors.NewDatabaseQueryError("user raster annotations", err)
	}
	if rasterAnnotationCount > 0 {
		return errors.NewUserHasDependenciesError(userUID, fmt.Sprintf("user has %d active raster annotations", rasterAnnotationCount))
	}

	return nil
}

// DeleteUser removes a user from the database after checking for dependencies
func (s *CrudService) DeleteUser(ctx context.Context, userUID string) error {
	// First check if the user has any dependencies
	if err := s.CheckUserDependencies(ctx, userUID); err != nil {
		return err
	}

	// If no dependencies, proceed with deletion
	result, err := s.db.Exec("DELETE FROM users WHERE short_uid = ? AND deactivated_at IS NULL", userUID)
	if err != nil {
		return errors.NewDatabaseDeleteError("user", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.NewUserNotFoundByUIDError(userUID)
	}

	return nil
}
