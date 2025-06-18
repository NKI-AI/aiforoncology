// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"aifo.dev/aifo/slidescope/internal/config"
	"aifo.dev/aifo/slidescope/internal/datasources/database"
	"aifo.dev/aifo/slidescope/internal/server/domain"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (domain.TokenResponse, error)
	GenerateJWT(username string) (string, time.Time, error)
	GenerateRefreshJWT(username string) (string, time.Time, error)
	ValidateRefreshToken(token string) (string, error)
	GetAuthConfig() config.AuthConfig
}

type authService struct {
	db          database.Database
	userService UserService
	jwtConfig   config.AuthConfig
}

func NewAuthService(db database.Database, jwtConfig config.AuthConfig) AuthService {
	return &authService{
		db:          db,
		userService: NewUserService(db),
		jwtConfig:   jwtConfig,
	}
}

func (s *authService) Login(ctx context.Context, username, password string) (domain.TokenResponse, error) {
	// Verify the user's credentials
	slog.Info("Verifying password", "username", username)
	err := s.userService.VerifyPassword(ctx, username, password)
	if err != nil {
		return domain.TokenResponse{}, errors.New("invalid credentials")
	}

	// Generate JWT token
	// Generate access and refresh tokens
	accessToken, accessExpiry, err := s.GenerateJWT(username)
	if err != nil {
		return domain.TokenResponse{}, err
	}

	refreshToken, refreshExpiry, err := s.GenerateRefreshJWT(username)
	if err != nil {
		return domain.TokenResponse{}, err
	}

	// Calculate expiration in seconds
	expiresIn := int(time.Until(accessExpiry).Seconds())
	refreshExpiresIn := int(time.Until(refreshExpiry).Seconds())

	// Return token response matching Python implementation
	return domain.TokenResponse{
		AccessToken:      accessToken,
		TokenType:        "bearer",
		ExpiresIn:        expiresIn,
		RefreshToken:     refreshToken,
		RefreshExpiresIn: refreshExpiresIn,
	}, nil
}

func (s *authService) GenerateRefreshJWT(username string) (string, time.Time, error) {
	expTime := time.Now().Add(s.jwtConfig.JWTRefreshExpirationMinutes)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": expTime.Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, err := token.SignedString(s.jwtConfig.GetJWTSecretBytes())
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expTime, nil
}

func (s *authService) ValidateRefreshToken(tokenString string) (string, error) {
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtConfig.GetJWTSecretBytes(), nil
	})
	if err != nil || !parsedToken.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	username, ok := claims["sub"].(string)
	if !ok {
		return "", errors.New("invalid username in token")
	}
	return username, nil
}

func (s *authService) GenerateJWT(username string) (string, time.Time, error) {
	// Calculate expiration time
	expTime := time.Now().Add(s.jwtConfig.JWTExpirationMinutes)

	// Create a new token with claims that match Python implementation
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":    username,          // Subject (username)
		"scopes": []string{},        // Empty scopes array (matching Python)
		"exp":    expTime.Unix(),    // Expiration time
		"iat":    time.Now().Unix(), // Issued at time
	})

	// Sign the token with our secret
	tokenString, err := token.SignedString(s.jwtConfig.GetJWTSecretBytes())
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expTime, nil
}

// GetAuthConfig returns the authentication configuration
func (s *authService) GetAuthConfig() config.AuthConfig {
	return s.jwtConfig
}
