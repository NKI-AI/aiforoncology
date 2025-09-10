// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package auth

import (
	"context"
	"fmt"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
)

const (
	// Rate limiting windows
	ShortWindow  = 15 * time.Minute // 15 minutes
	MediumWindow = 1 * time.Hour    // 1 hour
	LongWindow   = 24 * time.Hour   // 24 hours

	// Rate limiting thresholds
	MaxAttemptsShort  = 5  // 5 attempts per 15 minutes
	MaxAttemptsMedium = 10 // 10 attempts per hour
	MaxAttemptsLong   = 50 // 50 attempts per day

	// Account lockout
	AccountLockoutThreshold = 10               // Lock account after 10 failed attempts
	AccountLockoutDuration  = 30 * time.Minute // Lock for 30 minutes
)

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed       bool
	Reason        string
	ResetTime     time.Time
	AttemptsLeft  int
	TotalAttempts int
}

// RateLimiter handles authentication rate limiting
type RateLimiter struct {
	db ports.Database
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(db ports.Database) *RateLimiter {
	return &RateLimiter{
		db: db,
	}
}

// CheckRateLimit checks if an authentication attempt should be allowed
func (rl *RateLimiter) CheckRateLimit(ctx context.Context, ipAddress, email string) (RateLimitResult, error) {
	now := time.Now()

	// Check IP-based rate limiting first (prevents distributed attacks)
	ipResult, err := rl.checkIPRateLimit(ctx, ipAddress, now)
	if err != nil {
		return RateLimitResult{}, fmt.Errorf("failed to check IP rate limit: %w", err)
	}
	if !ipResult.Allowed {
		return ipResult, nil
	}

	// Check email-based rate limiting (prevents account-specific attacks)
	if email != "" {
		userResult, err := rl.checkUserRateLimit(ctx, email, now)
		if err != nil {
			return RateLimitResult{}, fmt.Errorf("failed to check user rate limit: %w", err)
		}
		if !userResult.Allowed {
			return userResult, nil
		}
	}

	return RateLimitResult{Allowed: true}, nil
}

// RecordAttempt records an authentication attempt
func (rl *RateLimiter) RecordAttempt(ctx context.Context, ipAddress, email string, success bool, failReason string) error {
	return rl.db.RecordAuthAttempt(ctx, ipAddress, email, success, failReason)
}

// checkIPRateLimit checks rate limiting for an IP address
func (rl *RateLimiter) checkIPRateLimit(ctx context.Context, ipAddress string, now time.Time) (RateLimitResult, error) {
	// Check short window (15 minutes)
	shortAttempts, err := rl.db.GetRecentAuthAttempts(ctx, ipAddress, now.Add(-ShortWindow))
	if err != nil {
		return RateLimitResult{}, err
	}

	failedShort := countFailedAttempts(shortAttempts)
	if failedShort >= MaxAttemptsShort {
		return RateLimitResult{
			Allowed:       false,
			Reason:        fmt.Sprintf("Too many failed attempts from this IP address. Try again in %d minutes.", int(ShortWindow.Minutes())),
			ResetTime:     now.Add(ShortWindow),
			AttemptsLeft:  0,
			TotalAttempts: failedShort,
		}, nil
	}

	// Check medium window (1 hour)
	mediumAttempts, err := rl.db.GetRecentAuthAttempts(ctx, ipAddress, now.Add(-MediumWindow))
	if err != nil {
		return RateLimitResult{}, err
	}

	failedMedium := countFailedAttempts(mediumAttempts)
	if failedMedium >= MaxAttemptsMedium {
		return RateLimitResult{
			Allowed:       false,
			Reason:        fmt.Sprintf("Too many failed attempts from this IP address. Try again in %d minutes.", int(MediumWindow.Minutes())),
			ResetTime:     now.Add(MediumWindow),
			AttemptsLeft:  0,
			TotalAttempts: failedMedium,
		}, nil
	}

	// Check long window (24 hours)
	longAttempts, err := rl.db.GetRecentAuthAttempts(ctx, ipAddress, now.Add(-LongWindow))
	if err != nil {
		return RateLimitResult{}, err
	}

	failedLong := countFailedAttempts(longAttempts)
	if failedLong >= MaxAttemptsLong {
		return RateLimitResult{
			Allowed:       false,
			Reason:        fmt.Sprintf("Too many failed attempts from this IP address. Try again in %d hours.", int(LongWindow.Hours())),
			ResetTime:     now.Add(LongWindow),
			AttemptsLeft:  0,
			TotalAttempts: failedLong,
		}, nil
	}

	// Calculate attempts left (use the most restrictive window)
	attemptsLeft := MaxAttemptsShort - failedShort
	if mediumLeft := MaxAttemptsMedium - failedMedium; mediumLeft < attemptsLeft {
		attemptsLeft = mediumLeft
	}
	if longLeft := MaxAttemptsLong - failedLong; longLeft < attemptsLeft {
		attemptsLeft = longLeft
	}

	return RateLimitResult{
		Allowed:       true,
		AttemptsLeft:  attemptsLeft,
		TotalAttempts: failedShort,
	}, nil
}

// checkUserRateLimit checks rate limiting for a specific email
func (rl *RateLimiter) checkUserRateLimit(ctx context.Context, email string, now time.Time) (RateLimitResult, error) {
	// Check account lockout window
	lockoutAttempts, err := rl.db.GetRecentAuthAttemptsForUser(ctx, email, now.Add(-AccountLockoutDuration))
	if err != nil {
		return RateLimitResult{}, err
	}

	failedLockout := countFailedAttempts(lockoutAttempts)
	if failedLockout >= AccountLockoutThreshold {
		return RateLimitResult{
			Allowed:       false,
			Reason:        fmt.Sprintf("Account temporarily locked due to too many failed attempts. Try again in %d minutes.", int(AccountLockoutDuration.Minutes())),
			ResetTime:     now.Add(AccountLockoutDuration),
			AttemptsLeft:  0,
			TotalAttempts: failedLockout,
		}, nil
	}

	// Check short window for user
	shortAttempts, err := rl.db.GetRecentAuthAttemptsForUser(ctx, email, now.Add(-ShortWindow))
	if err != nil {
		return RateLimitResult{}, err
	}

	failedShort := countFailedAttempts(shortAttempts)
	if failedShort >= MaxAttemptsShort {
		return RateLimitResult{
			Allowed:       false,
			Reason:        fmt.Sprintf("Too many failed attempts for this account. Try again in %d minutes.", int(ShortWindow.Minutes())),
			ResetTime:     now.Add(ShortWindow),
			AttemptsLeft:  0,
			TotalAttempts: failedShort,
		}, nil
	}

	attemptsLeft := AccountLockoutThreshold - failedLockout
	if shortLeft := MaxAttemptsShort - failedShort; shortLeft < attemptsLeft {
		attemptsLeft = shortLeft
	}

	return RateLimitResult{
		Allowed:       true,
		AttemptsLeft:  attemptsLeft,
		TotalAttempts: failedShort,
	}, nil
}

// countFailedAttempts counts the number of failed attempts in a list
func countFailedAttempts(attempts []ports.AuthAttempt) int {
	count := 0
	for _, attempt := range attempts {
		if !attempt.Success {
			count++
		}
	}
	return count
}

// CleanupOldAttempts removes old authentication attempts to keep the database clean
func (rl *RateLimiter) CleanupOldAttempts(ctx context.Context) error {
	// Keep attempts for 7 days for analysis
	cleanupTime := time.Now().Add(-7 * 24 * time.Hour)
	return rl.db.CleanupOldAuthAttempts(ctx, cleanupTime)
}

// GetRateLimitStatus provides current rate limit status for monitoring
func (rl *RateLimiter) GetRateLimitStatus(ctx context.Context, ipAddress, email string) (map[string]interface{}, error) {
	now := time.Now()
	status := make(map[string]interface{})

	// IP status
	if ipAddress != "" {
		shortAttempts, err := rl.db.GetRecentAuthAttempts(ctx, ipAddress, now.Add(-ShortWindow))
		if err != nil {
			return nil, err
		}
		status["ip_failed_attempts_15min"] = countFailedAttempts(shortAttempts)
		status["ip_limit_15min"] = MaxAttemptsShort
	}

	// User status
	if email != "" {
		shortAttempts, err := rl.db.GetRecentAuthAttemptsForUser(ctx, email, now.Add(-ShortWindow))
		if err != nil {
			return nil, err
		}
		lockoutAttempts, err := rl.db.GetRecentAuthAttemptsForUser(ctx, email, now.Add(-AccountLockoutDuration))
		if err != nil {
			return nil, err
		}
		status["user_failed_attempts_15min"] = countFailedAttempts(shortAttempts)
		status["user_failed_attempts_30min"] = countFailedAttempts(lockoutAttempts)
		status["user_limit_15min"] = MaxAttemptsShort
		status["account_lockout_threshold"] = AccountLockoutThreshold
	}

	return status, nil
}
