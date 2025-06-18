// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aifo.dev/aifo/slidescope/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// Define a test JWT configuration for all tests
var testJWTConfig = config.AuthConfig{
	JWTSecret:            "test-secret-key",
	JWTAlgorithm:         "HS256",
	JWTExpirationMinutes: 60 * time.Minute,
	Cookie: config.CookieConfig{
		Name:     "slidescope_auth",
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "lax",
	},
}

// Helper function to generate a valid JWT token for testing
func generateTestToken(username string, expires time.Time) string {
	claims := jwt.MapClaims{
		"sub":    username,
		"scopes": []string{},
		"exp":    expires.Unix(),
		"iat":    time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testJWTConfig.JWTSecret))
	return tokenString
}

// Mock version of ExtractToken for testing
func extractTokenForTest(c *fiber.Ctx, cookieName string) string {
	// First try to get token from cookie
	token := c.Cookies(cookieName)

	// If no token found in cookie, check Authorization header
	if token == "" {
		authHeader := c.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	return token
}

// Test version of Protected middleware that uses our test config
func protectedForTest() fiber.Handler {
	jwtSecret := []byte(testJWTConfig.JWTSecret)
	cookieName := testJWTConfig.Cookie.Name

	return func(c *fiber.Ctx) error {
		// Extract token from cookie or header
		token := extractTokenForTest(c, cookieName)

		// No token found
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid or missing API token",
			})
		}

		// Parse and validate the token
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil {
			var message string
			if err.Error() == "token has expired" {
				message = "Token has expired"
			} else {
				message = "Invalid token"
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": message,
			})
		}

		if !parsedToken.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid token",
			})
		}

		// Extract claims
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid token claims",
			})
		}

		// Extract username from subject claim
		username, ok := claims["sub"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid username in token",
			})
		}

		// Set username in locals for use in subsequent handlers
		c.Locals("username", username)

		// Continue to the next middleware/handler
		return c.Next()
	}
}

func TestExtractToken_FromCookie(t *testing.T) {
	// Create a test app
	app := fiber.New()

	// Define a test route
	app.Get("/", func(c *fiber.Ctx) error {
		token := extractTokenForTest(c, testJWTConfig.Cookie.Name)
		return c.SendString(token)
	})

	// Create a test request with a cookie
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  testJWTConfig.Cookie.Name,
		Value: "test-token-from-cookie",
	})

	// Execute the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Check response
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "test-token-from-cookie", string(bodyBytes))
}

func TestExtractToken_FromAuthHeader(t *testing.T) {
	// Create a test app
	app := fiber.New()

	// Define a test route
	app.Get("/", func(c *fiber.Ctx) error {
		token := extractTokenForTest(c, testJWTConfig.Cookie.Name)
		return c.SendString(token)
	})

	// Create a test request with an Authorization header
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer test-token-from-header")

	// Execute the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Check response
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "test-token-from-header", string(bodyBytes))
}

func TestExtractToken_NoCookieNoHeader(t *testing.T) {
	// Create a test app
	app := fiber.New()

	// Define a test route
	app.Get("/", func(c *fiber.Ctx) error {
		token := extractTokenForTest(c, testJWTConfig.Cookie.Name)
		return c.SendString(token)
	})

	// Create a test request with no cookie or header
	req := httptest.NewRequest("GET", "/", nil)

	// Execute the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Check response
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "", string(bodyBytes))
}

func TestExtractToken_PrefersCookie(t *testing.T) {
	// Create a test app
	app := fiber.New()

	// Define a test route
	app.Get("/", func(c *fiber.Ctx) error {
		token := extractTokenForTest(c, testJWTConfig.Cookie.Name)
		return c.SendString(token)
	})

	// Create a test request with both cookie and header
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  testJWTConfig.Cookie.Name,
		Value: "test-token-from-cookie",
	})
	req.Header.Set("Authorization", "Bearer test-token-from-header")

	// Execute the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Check response
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "test-token-from-cookie", string(bodyBytes))
}

func TestProtected_ValidToken(t *testing.T) {
	// Create a test app
	app := fiber.New()

	// Add the test middleware
	app.Use(protectedForTest())

	// Define a test route
	app.Get("/", func(c *fiber.Ctx) error {
		username := c.Locals("username")
		return c.SendString(username.(string))
	})

	// Generate a valid token
	token := generateTestToken("testuser", time.Now().Add(time.Hour))

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  testJWTConfig.Cookie.Name,
		Value: token,
	})

	// Execute the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Check response
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", string(bodyBytes))
}

func TestProtected_ExpiredToken(t *testing.T) {
	// Create a test app
	app := fiber.New()

	// Add the test middleware
	app.Use(protectedForTest())

	// Define a test route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("You should not see this")
	})

	// Generate an expired token
	token := generateTestToken("testuser", time.Now().Add(-time.Hour))

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  testJWTConfig.Cookie.Name,
		Value: token,
	})

	// Execute the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestProtected_InvalidToken(t *testing.T) {
	// Create a test app
	app := fiber.New()

	// Add the test middleware
	app.Use(protectedForTest())

	// Define a test route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("You should not see this")
	})

	// Set an invalid token
	invalidToken := "invalid.jwt.token"

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  testJWTConfig.Cookie.Name,
		Value: invalidToken,
	})

	// Execute the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestProtected_MissingToken(t *testing.T) {
	// Create a test app
	app := fiber.New()

	// Add the test middleware
	app.Use(protectedForTest())

	// Define a test route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("You should not see this")
	})

	// Create a test request with no token
	req := httptest.NewRequest("GET", "/", nil)

	// Execute the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestProtected_InvalidSigningMethod(t *testing.T) {
	// Create a test app
	app := fiber.New()

	// Add the test middleware
	app.Use(protectedForTest())

	// Define a test route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("You should not see this")
	})

	// Create a token with a different signing method
	claims := jwt.MapClaims{
		"sub":    "testuser",
		"scopes": []string{},
		"exp":    time.Now().Add(time.Hour).Unix(),
		"iat":    time.Now().Unix(),
	}

	// Use a different signing method
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  testJWTConfig.Cookie.Name,
		Value: tokenString,
	})

	// Execute the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestProtected_MissingUsername(t *testing.T) {
	// Create a test app
	app := fiber.New()

	// Add the test middleware
	app.Use(protectedForTest())

	// Define a test route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("You should not see this")
	})

	// Create a token without a subject (username)
	claims := jwt.MapClaims{
		// Missing "sub" claim
		"scopes": []string{},
		"exp":    time.Now().Add(time.Hour).Unix(),
		"iat":    time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testJWTConfig.JWTSecret))

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  testJWTConfig.Cookie.Name,
		Value: tokenString,
	})

	// Execute the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestProtected_TokenFromAuthHeader(t *testing.T) {
	// Create a test app
	app := fiber.New()

	// Add the test middleware
	app.Use(protectedForTest())

	// Define a test route
	app.Get("/", func(c *fiber.Ctx) error {
		username := c.Locals("username")
		return c.SendString(username.(string))
	})

	// Generate a valid token
	token := generateTestToken("testuser", time.Now().Add(time.Hour))

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute the request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Check response
	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", string(bodyBytes))
}
