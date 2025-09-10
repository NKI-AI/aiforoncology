// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package auth

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"github.com/golang-jwt/jwt/v5"
)

func (s *authService) GenerateJWT(user domain.User) (string, time.Time, error) {
	// Calculate expiration time
	expTime := time.Now().Add(s.jwtConfig.JWTExpirationMinutes)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":    user.Email,                       // Subject (email)
		"tenant": fmt.Sprintf("%d", user.TenantID), // Tenant as string
		"user":   user.ShortUID,                    // User UID
		"scopes": []string{},                       // Empty scopes array
		"exp":    expTime.Unix(),                   // Expiration time
		"iat":    time.Now().Unix(),                // Issued at time
	})

	// Sign the token with our secret
	tokenString, err := token.SignedString(s.jwtConfig.GetJWTSecretBytes())
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expTime, nil
}

func (s *authService) GenerateRefreshJWT(user domain.User) (string, time.Time, error) {
	expTime := time.Now().Add(s.jwtConfig.JWTRefreshExpirationMinutes)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":    user.Email,
		"tenant": fmt.Sprintf("%d", user.TenantID),
		"user":   user.ShortUID,
		"email":  user.Email,
		"exp":    expTime.Unix(),
		"iat":    time.Now().Unix(),
	})

	tokenString, err := token.SignedString(s.jwtConfig.GetJWTSecretBytes())
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expTime, nil
}

// GenerateSwitchedUserJWT generates a JWT for a switched user, including original admin info
func (s *authService) GenerateSwitchedUserJWT(targetUser domain.User, originalUserUID, originalEmail string) (string, time.Time, error) {
	// Calculate expiration time
	expTime := time.Now().Add(s.jwtConfig.JWTExpirationMinutes)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":              targetUser.Email,                       // Subject (target user's email)
		"tenant":           fmt.Sprintf("%d", targetUser.TenantID), // Target user's tenant as string
		"user":             targetUser.ShortUID,                    // Target user's UID
		"scopes":           []string{},                             // Empty scopes array
		"exp":              expTime.Unix(),                         // Expiration time
		"iat":              time.Now().Unix(),                      // Issued at time
		"original_user":    originalUserUID,                        // Original admin's UID
		"original_email":   originalEmail,                          // Original admin's email
		"switched_session": true,                                   // Flag to identify switched sessions
	})

	// Sign the token with our secret
	tokenString, err := token.SignedString(s.jwtConfig.GetJWTSecretBytes())
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expTime, nil
}

func (s *authService) ValidateRefreshToken(ctx context.Context, tokenString string) (domain.User, error) {
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, stderrors.New("unexpected signing method")
		}
		return s.jwtConfig.GetJWTSecretBytes(), nil
	})
	if err != nil || !parsedToken.Valid {
		return domain.User{}, stderrors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return domain.User{}, stderrors.New("invalid token claims")
	}

	email, ok := claims["sub"].(string)
	if !ok {
		return domain.User{}, stderrors.New("invalid email in token")
	}

	// Fetch the complete user object
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, stderrors.New("user not found")
	}

	return user, nil
}
