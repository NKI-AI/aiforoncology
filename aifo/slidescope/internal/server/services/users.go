// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package services

import (
	"context"
	"fmt"

	"aifo.dev/aifo/slidescope/internal/datasources/database"
	"aifo.dev/aifo/slidescope/internal/server/auth"
	"aifo.dev/aifo/slidescope/internal/server/domain"
)

type UserService interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (domain.User, error)
	UpdatePassword(ctx context.Context, username, newPassword string) error
	VerifyPassword(ctx context.Context, username, password string) error
}

type userService struct {
	db database.Database
}

func NewUserService(db database.Database) UserService {
	return &userService{db: db}
}

func (s *userService) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	if user.Username == "" {
		return domain.User{}, fmt.Errorf("username cannot be empty")
	}

	if user.Password == "" {
		return domain.User{}, fmt.Errorf("password cannot be empty")
	}

	// Check if user already exists
	_, err := s.db.GetUserByUsername(ctx, user.Username)
	if err == nil {
		return domain.User{}, fmt.Errorf("user with username '%s' already exists", user.Username)
	}

	// Only hash the password if it's not already hashed
	hashedPassword := user.Password
	if !auth.IsHashedPassword(user.Password) {
		var err error
		hashedPassword, err = auth.HashPassword(user.Password)
		if err != nil {
			return domain.User{}, fmt.Errorf("failed to hash password: %w", err)
		}
	}

	dbUser := database.NewUser{
		Username: user.Username,
		Password: hashedPassword,
	}

	err = s.db.CreateUser(ctx, dbUser)
	if err != nil {
		return domain.User{}, err
	}

	// Return user without exposing the actual hashed password
	return domain.User{
		Username: user.Username,
		Password: "", // Don't return the password
	}, nil
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (domain.User, error) {
	dbUser, err := s.db.GetUserByUsername(ctx, username)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return domain.User{
		Username: dbUser.Username,
		Password: dbUser.Password,
	}, nil
}

func (s *userService) UpdatePassword(ctx context.Context, username, newPassword string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if newPassword == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Verify user exists
	_, err := s.db.GetUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("user with username '%s' does not exist", username)
	}

	// Hash the new password
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update the password in the database
	err = s.db.UpdateUserPassword(ctx, username, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *userService) VerifyPassword(ctx context.Context, username, password string) error {
	// Get the user with the hashed password
	user, err := s.GetUserByUsername(ctx, username)
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
