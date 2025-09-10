// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package ports

import "context"

// EmailRequest represents a request to send an email
type EmailRequest struct {
	To       string                 `json:"to"`
	Subject  string                 `json:"subject"`
	Template EmailTemplateType      `json:"template"`
	Data     map[string]interface{} `json:"data"`
	TenantID int                    `json:"tenantId,omitempty"`
}

// EmailService defines the interface for sending emails
type EmailService interface {
	// SendEmail sends an email using the specified template and data
	SendEmail(ctx context.Context, request EmailRequest) error

	// SendPasswordResetEmail sends a password reset email
	SendPasswordResetEmail(ctx context.Context, email, token string) error

	// SendEmailVerificationEmail sends an email verification email
	SendEmailVerificationEmail(ctx context.Context, email, token string) error

	// SendWelcomeEmail sends a welcome email to new users
	SendWelcomeEmail(ctx context.Context, email string) error
}
