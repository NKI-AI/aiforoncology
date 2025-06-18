// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"aifo.dev/aifo/slidescope/internal/server/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var usersRoute = "/api/v1/users"

// UserServiceMock mocks the UserService interface for testing
type UserServiceMock struct {
	mock.Mock
}

func (m *UserServiceMock) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return domain.User{}, args.Error(1)
	}
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *UserServiceMock) GetUserByUsername(ctx context.Context, username string) (domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return domain.User{}, args.Error(1)
	}
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *UserServiceMock) UpdatePassword(ctx context.Context, username, newPassword string) error {
	args := m.Called(ctx, username, newPassword)
	return args.Error(0)
}

func (m *UserServiceMock) VerifyPassword(ctx context.Context, username, password string) error {
	args := m.Called(ctx, username, password)
	return args.Error(0)
}

func TestCreateUser_Success(t *testing.T) {
	mockService := new(UserServiceMock)

	// Setup input and mock behavior
	inputUser := domain.User{
		Username: "newuser",
		Password: "password123",
	}

	returnedUser := domain.User{
		Username: "newuser",
		Password: "", // Password should not be returned
	}

	mockService.On("CreateUser", mock.Anything, inputUser).Return(returnedUser, nil)

	// Setup app with handler
	app := fiber.New()
	app.Post(usersRoute, CreateUser(mockService))

	// Create request
	requestBody := `{"username":"newuser","password":"password123"}`
	req := httptest.NewRequest("POST", usersRoute, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Test
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	// Check response body
	var response domain.User
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	assert.Equal(t, "newuser", response.Username)
	assert.Empty(t, response.Password)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestCreateUser_EmptyUsername(t *testing.T) {
	mockService := new(UserServiceMock)

	// Setup app with handler
	app := fiber.New()
	app.Post(usersRoute, CreateUser(mockService))

	// Create request with empty username
	requestBody := `{"username":"","password":"password123"}`
	req := httptest.NewRequest("POST", usersRoute, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Test
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "username is required", errorResp.Error)
}

func TestCreateUser_EmptyPassword(t *testing.T) {
	mockService := new(UserServiceMock)

	// Setup app with handler
	app := fiber.New()
	app.Post(usersRoute, CreateUser(mockService))

	// Create request with empty password
	requestBody := `{"username":"newuser","password":""}`
	req := httptest.NewRequest("POST", usersRoute, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Test
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "password is required", errorResp.Error)
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	mockService := new(UserServiceMock)

	// Setup app with handler
	app := fiber.New()
	app.Post(usersRoute, CreateUser(mockService))

	// Create request with invalid JSON
	requestBody := `{"username":newuser,"password":"password123"}` // Missing quotes around newuser
	req := httptest.NewRequest("POST", usersRoute, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Test
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "invalid request", errorResp.Error)
}

func TestCreateUser_ServiceError(t *testing.T) {
	mockService := new(UserServiceMock)

	// Setup input and mock behavior
	inputUser := domain.User{
		Username: "newuser",
		Password: "password123",
	}

	mockService.On("CreateUser", mock.Anything, inputUser).Return(domain.User{}, assert.AnError)

	// Setup app with handler
	app := fiber.New()
	app.Post(usersRoute, CreateUser(mockService))

	// Create request
	requestBody := `{"username":"newuser","password":"password123"}`
	req := httptest.NewRequest("POST", usersRoute, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Test
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "internal error", errorResp.Error)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestGetUserByUsername_Success(t *testing.T) {
	mockService := new(UserServiceMock)

	// Setup mock behavior
	returnedUser := domain.User{
		Username: "existinguser",
		Password: "hashedpassword", // Note: In a real scenario, this might need sanitization
	}

	mockService.On("GetUserByUsername", mock.Anything, "existinguser").Return(returnedUser, nil)

	// Setup app with handler
	app := fiber.New()
	app.Get(usersRoute+"/:username", GetUserByUsername(mockService))

	// Test
	resp, err := app.Test(httptest.NewRequest("GET", usersRoute+"/existinguser", nil))
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Check response body
	var response domain.User
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	assert.Equal(t, "existinguser", response.Username)
	assert.Equal(t, "hashedpassword", response.Password) // In real app, this should be sanitized before returning

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	mockService := new(UserServiceMock)

	// Setup mock behavior - user not found
	mockService.On("GetUserByUsername", mock.Anything, "nonexistent").Return(
		domain.User{},
		fmt.Errorf("failed to get user: user with username '%s' not found", "nonexistent"),
	)

	// Setup app with handler
	app := fiber.New()
	app.Get(usersRoute+"/:username", GetUserByUsername(mockService))

	// Test
	resp, err := app.Test(httptest.NewRequest("GET", usersRoute+"/nonexistent", nil))
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "user not found", errorResp.Error)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestGetUserByUsername_ServiceError(t *testing.T) {
	mockService := new(UserServiceMock)

	// Setup mock behavior - general service error
	mockService.On("GetUserByUsername", mock.Anything, "existinguser").Return(domain.User{}, assert.AnError)

	// Setup app with handler
	app := fiber.New()
	app.Get(usersRoute+"/:username", GetUserByUsername(mockService))

	// Test
	resp, err := app.Test(httptest.NewRequest("GET", usersRoute+"/existinguser", nil))
	assert.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "internal error", errorResp.Error)

	// Verify mocks
	mockService.AssertExpectations(t)
}

func TestGetUserByUsername_EmptyUsername(t *testing.T) {
	mockService := new(UserServiceMock)

	// Setup app with handler
	app := fiber.New()
	app.Get(usersRoute+"/:username", GetUserByUsername(mockService))

	// Test with empty username - this is a bit artificial since the router wouldn't match this
	resp, err := app.Test(httptest.NewRequest("GET", usersRoute+"/", nil))
	assert.Nil(t, err)
	assert.Equal(t, 404, resp.StatusCode) // Fiber returns 404 for missing route params
}
