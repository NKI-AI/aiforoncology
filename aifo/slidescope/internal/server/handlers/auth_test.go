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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aifo.dev/aifo/slidescope/internal/config"
	"aifo.dev/aifo/slidescope/internal/server/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var authRoute = "/api/v1/auth"

// TestCookieConfig defines the cookie configuration for testing
var testCookieConfig = config.CookieConfig{
	Name:     "slidescope_auth",
	Path:     "/",
	HTTPOnly: true,
	Secure:   false,
	SameSite: "lax",
}

// TestAuthConfig defines the auth configuration for testing
var testAuthConfig = config.AuthConfig{
	JWTSecret:            "test-secret",
	JWTAlgorithm:         "HS256",
	JWTExpirationMinutes: 60 * time.Minute,
	Cookie:               testCookieConfig,
}

// AuthServiceMock is a mock implementation of the AuthService interface for testing
type AuthServiceMock struct {
	mock.Mock
}

func (m *AuthServiceMock) Login(ctx context.Context, username, password string) (domain.TokenResponse, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return domain.TokenResponse{}, args.Error(1)
	}
	return args.Get(0).(domain.TokenResponse), args.Error(1)
}

func (m *AuthServiceMock) GenerateJWT(username string) (string, time.Time, error) {
	args := m.Called(username)
	return args.String(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *AuthServiceMock) GenerateRefreshJWT(username string) (string, time.Time, error) {
	args := m.Called(username)
	return args.String(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *AuthServiceMock) ValidateRefreshToken(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

func (m *AuthServiceMock) GetAuthConfig() config.AuthConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return testAuthConfig
	}
	return args.Get(0).(config.AuthConfig)
}

func (m *AuthServiceMock) ValidateToken(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

func TestLogin_Success(t *testing.T) {
	mockService := new(AuthServiceMock)

	// Mock service behavior
	mockToken := domain.TokenResponse{
		AccessToken: "jwt.token.here",
		TokenType:   "bearer",
		ExpiresIn:   3600,
	}
	mockService.On("Login", mock.Anything, "testuser", "password123").Return(mockToken, nil)
	mockService.On("GetAuthConfig").Return(testAuthConfig)

	app := fiber.New()
	app.Post(authRoute+"/login", Login(mockService))

	// Create request
	loginBody := `{"username":"testuser","password":"password123"}`
	req := httptest.NewRequest("POST", authRoute+"/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")

	// Test
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Check for auth cookie
	cookies := resp.Cookies()
	var authCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "slidescope_auth" {
			authCookie = cookie
			break
		}
	}
	assert.NotNil(t, authCookie)
	assert.Equal(t, "jwt.token.here", authCookie.Value)
	assert.True(t, authCookie.HttpOnly)
	assert.Equal(t, "/", authCookie.Path)

	// Check response body
	var response domain.TokenResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	assert.Equal(t, "jwt.token.here", response.AccessToken)
	assert.Equal(t, "bearer", response.TokenType)
	assert.Equal(t, 3600, response.ExpiresIn)

	// Verify all mocks
	mockService.AssertExpectations(t)
}

func TestLogin_EmptyUsername(t *testing.T) {
	mockService := new(AuthServiceMock)
	mockService.On("GetAuthConfig").Return(testAuthConfig)

	app := fiber.New()
	app.Post(authRoute+"/login", Login(mockService))

	// Create request with empty username
	loginBody := `{"username":"","password":"password123"}`
	req := httptest.NewRequest("POST", authRoute+"/login", bytes.NewBufferString(loginBody))
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

func TestLogin_EmptyPassword(t *testing.T) {
	mockService := new(AuthServiceMock)
	mockService.On("GetAuthConfig").Return(testAuthConfig)

	app := fiber.New()
	app.Post(authRoute+"/login", Login(mockService))

	// Create request with empty password
	loginBody := `{"username":"testuser","password":""}`
	req := httptest.NewRequest("POST", authRoute+"/login", bytes.NewBufferString(loginBody))
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

func TestLogin_InvalidCredentials(t *testing.T) {
	mockService := new(AuthServiceMock)

	// Mock service behavior - authentication fails
	mockService.On("Login", mock.Anything, "testuser", "wrong_password").Return(domain.TokenResponse{}, assert.AnError)
	// GetAuthConfig is not called when authentication fails, so we don't set an expectation for it

	app := fiber.New()
	app.Post(authRoute+"/login", Login(mockService))

	// Create request
	loginBody := `{"username":"testuser","password":"wrong_password"}`
	req := httptest.NewRequest("POST", authRoute+"/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")

	// Test
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "incorrect username or password", errorResp.Error)

	// Verify all mocks
	mockService.AssertExpectations(t)
}

func TestLogin_InvalidJSON(t *testing.T) {
	mockService := new(AuthServiceMock)
	mockService.On("GetAuthConfig").Return(testAuthConfig)

	app := fiber.New()
	app.Post(authRoute+"/login", Login(mockService))

	// Create request with invalid JSON
	loginBody := `{"username":testuser,"password":"password123"}` // Missing quotes around testuser
	req := httptest.NewRequest("POST", authRoute+"/login", bytes.NewBufferString(loginBody))
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

func TestLogout(t *testing.T) {
	mockService := new(AuthServiceMock)
	mockService.On("GetAuthConfig").Return(testAuthConfig)

	app := fiber.New()
	app.Post(authRoute+"/logout", Logout(mockService))

	// Test
	resp, err := app.Test(httptest.NewRequest("POST", authRoute+"/logout", nil))
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Check for cleared auth cookie
	cookies := resp.Cookies()
	var authCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "slidescope_auth" {
			authCookie = cookie
			break
		}
	}
	assert.NotNil(t, authCookie)
	assert.Equal(t, "", authCookie.Value)
	assert.True(t, authCookie.HttpOnly)
	assert.Equal(t, "/", authCookie.Path)
	assert.True(t, authCookie.Expires.Before(time.Now()), "Cookie should expire in the past")

	// Check response body
	var response map[string]string
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	assert.Equal(t, "Successfully logged out", response["message"])
}

func TestGetCurrentUser_Success(t *testing.T) {
	mockService := new(AuthServiceMock)
	mockService.On("GetAuthConfig").Return(testAuthConfig)

	app := fiber.New()

	// Apply middleware globally before adding routes
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("username", "testuser")
		return c.Next()
	})

	// Add route after middleware
	app.Get(authRoute+"/me", GetCurrentUser(mockService))

	// Test
	resp, err := app.Test(httptest.NewRequest("GET", authRoute+"/me", nil))
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Check response body
	var response struct {
		Username string   `json:"username"`
		Scopes   []string `json:"scopes"`
	}
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &response)
	assert.Nil(t, err)
	assert.Equal(t, "testuser", response.Username)
	assert.Empty(t, response.Scopes)
}

func TestGetCurrentUser_NotAuthenticated(t *testing.T) {
	mockService := new(AuthServiceMock)
	mockService.On("GetAuthConfig").Return(testAuthConfig)

	app := fiber.New()
	app.Get(authRoute+"/me", GetCurrentUser(mockService))

	// Test without setting username in locals
	resp, err := app.Test(httptest.NewRequest("GET", authRoute+"/me", nil))
	assert.Nil(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	// Check response body
	var errorResp domain.ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &errorResp)
	assert.Nil(t, err)
	assert.Equal(t, "not authenticated", errorResp.Error)
}
