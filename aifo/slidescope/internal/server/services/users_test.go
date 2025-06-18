// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package services

import (
	"context"
	"testing"

	"aifo.dev/aifo/slidescope/internal/datasources/database"
	"aifo.dev/aifo/slidescope/internal/server/auth"
	"aifo.dev/aifo/slidescope/internal/server/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateUser_Success(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Setup input
	inputUser := domain.User{
		Username: "testuser",
		Password: "password123",
	}

	// Mock DB behavior
	mockDB.On("GetUserByUsername", mock.Anything, "testuser").Return(database.User{}, assert.AnError) // User doesn't exist
	mockDB.On("CreateUser", mock.Anything, mock.AnythingOfType("database.NewUser")).Return(nil)

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	createdUser, err := service.CreateUser(context.Background(), inputUser)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "testuser", createdUser.Username)
	assert.Empty(t, createdUser.Password, "Password should not be returned")

	// Verify mocks
	mockDB.AssertExpectations(t)
}

func TestCreateUser_EmptyUsername(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Setup input with empty username
	inputUser := domain.User{
		Username: "",
		Password: "password123",
	}

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	_, err := service.CreateUser(context.Background(), inputUser)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username cannot be empty")
}

func TestCreateUser_EmptyPassword(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Setup input with empty password
	inputUser := domain.User{
		Username: "testuser",
		Password: "",
	}

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	_, err := service.CreateUser(context.Background(), inputUser)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

func TestCreateUser_UserAlreadyExists(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Setup input
	inputUser := domain.User{
		Username: "existinguser",
		Password: "password123",
	}

	// Mock DB behavior - user already exists
	mockDB.On("GetUserByUsername", mock.Anything, "existinguser").Return(database.User{
		Username: "existinguser",
		Password: "hashedpassword",
	}, nil)

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	_, err := service.CreateUser(context.Background(), inputUser)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Verify mocks
	mockDB.AssertExpectations(t)
}

func TestCreateUser_DBError(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Setup input
	inputUser := domain.User{
		Username: "testuser",
		Password: "password123",
	}

	// Mock DB behavior
	mockDB.On("GetUserByUsername", mock.Anything, "testuser").Return(database.User{}, assert.AnError) // User doesn't exist
	mockDB.On("CreateUser", mock.Anything, mock.AnythingOfType("database.NewUser")).Return(assert.AnError)

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	_, err := service.CreateUser(context.Background(), inputUser)

	// Assertions
	assert.Error(t, err)

	// Verify mocks
	mockDB.AssertExpectations(t)
}

func TestGetUserByUsername_Success(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Mock DB behavior
	mockDB.On("GetUserByUsername", mock.Anything, "testuser").Return(database.User{
		Username: "testuser",
		Password: "hashedpassword",
	}, nil)

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	user, err := service.GetUserByUsername(context.Background(), "testuser")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "hashedpassword", user.Password)

	// Verify mocks
	mockDB.AssertExpectations(t)
}

func TestGetUserByUsername_UserNotFound(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Mock DB behavior
	mockDB.On("GetUserByUsername", mock.Anything, "nonexistent").Return(database.User{}, assert.AnError)

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	_, err := service.GetUserByUsername(context.Background(), "nonexistent")

	// Assertions
	assert.Error(t, err)

	// Verify mocks
	mockDB.AssertExpectations(t)
}

func TestUpdatePassword_Success(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Mock DB behavior
	mockDB.On("GetUserByUsername", mock.Anything, "testuser").Return(database.User{
		Username: "testuser",
		Password: "oldhash",
	}, nil)
	mockDB.On("UpdateUserPassword", mock.Anything, "testuser", mock.AnythingOfType("string")).Return(nil)

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	err := service.UpdatePassword(context.Background(), "testuser", "newpassword")

	// Assertions
	assert.NoError(t, err)

	// Verify mocks
	mockDB.AssertExpectations(t)

	// Verify that the password was hashed
	call := mockDB.Calls[1]
	hashedPassword := call.Arguments.Get(2).(string)
	assert.True(t, auth.IsHashedPassword(hashedPassword), "Password should be hashed")
}

func TestUpdatePassword_EmptyUsername(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	err := service.UpdatePassword(context.Background(), "", "newpassword")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username cannot be empty")
}

func TestUpdatePassword_EmptyPassword(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	err := service.UpdatePassword(context.Background(), "testuser", "")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

func TestUpdatePassword_UserNotFound(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Mock DB behavior
	mockDB.On("GetUserByUsername", mock.Anything, "nonexistent").Return(database.User{}, assert.AnError)

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	err := service.UpdatePassword(context.Background(), "nonexistent", "newpassword")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")

	// Verify mocks
	mockDB.AssertExpectations(t)
}

func TestUpdatePassword_DBError(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Mock DB behavior
	mockDB.On("GetUserByUsername", mock.Anything, "testuser").Return(database.User{
		Username: "testuser",
		Password: "oldhash",
	}, nil)
	mockDB.On("UpdateUserPassword", mock.Anything, "testuser", mock.AnythingOfType("string")).Return(assert.AnError)

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	err := service.UpdatePassword(context.Background(), "testuser", "newpassword")

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update password")

	// Verify mocks
	mockDB.AssertExpectations(t)
}

func TestVerifyPassword_Success(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Create a real hashed password
	hashedPassword, _ := auth.HashPassword("correctpassword")

	// Mock DB behavior
	mockDB.On("GetUserByUsername", mock.Anything, "testuser").Return(database.User{
		Username: "testuser",
		Password: hashedPassword,
	}, nil)

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	err := service.VerifyPassword(context.Background(), "testuser", "correctpassword")

	// Assertions
	assert.NoError(t, err)

	// Verify mocks
	mockDB.AssertExpectations(t)
}

func TestVerifyPassword_UserNotFound(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Mock DB behavior
	mockDB.On("GetUserByUsername", mock.Anything, "nonexistent").Return(database.User{}, assert.AnError)

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	err := service.VerifyPassword(context.Background(), "nonexistent", "password")

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, auth.ErrPasswordMismatch, err, "Should return generic error for nonexistent user")

	// Verify mocks
	mockDB.AssertExpectations(t)
}

func TestVerifyPassword_IncorrectPassword(t *testing.T) {
	mockDB := new(database.DatabaseMock)

	// Create a real hashed password
	hashedPassword, _ := auth.HashPassword("correctpassword")

	// Mock DB behavior
	mockDB.On("GetUserByUsername", mock.Anything, "testuser").Return(database.User{
		Username: "testuser",
		Password: hashedPassword,
	}, nil)

	// Create the service
	service := NewUserService(mockDB)

	// Call the method
	err := service.VerifyPassword(context.Background(), "testuser", "wrongpassword")

	// Assertions
	assert.Error(t, err)

	// Verify mocks
	mockDB.AssertExpectations(t)
}
