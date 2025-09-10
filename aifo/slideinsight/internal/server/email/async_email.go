// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package email

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/queue"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/gofiber/fiber/v2/log"
)

// AsyncEmailService provides a unified interface for sending emails asynchronously
// with automatic fallback to synchronous sending if the queue is unavailable
type AsyncEmailService interface {
	// SendEmailAsync sends an email asynchronously via the queue, with fallback to direct sending
	SendEmailAsync(ctx context.Context, request ports.EmailRequest) error

	// Helper methods for common email types
	SendPasswordResetEmailAsync(ctx context.Context, email, token string) error
	SendEmailVerificationEmailAsync(ctx context.Context, email, token string) error
	SendWelcomeEmailAsync(ctx context.Context, email string) error
}

type asyncEmailService struct {
	emailService ports.EmailService
	queueManager *queue.QueueManager
}

// NewAsyncEmailService creates a new async email service
func NewAsyncEmailService(emailService ports.EmailService, queueManager *queue.QueueManager) AsyncEmailService {
	return &asyncEmailService{
		emailService: emailService,
		queueManager: queueManager,
	}
}

// SendEmailAsync sends an email asynchronously via the queue, with fallback to direct sending
func (s *asyncEmailService) SendEmailAsync(ctx context.Context, request ports.EmailRequest) error {
	if s.queueManager == nil {
		log.Warn("Queue manager not available, sending email directly",
			"template", request.Template, "email", request.To)
		return s.emailService.SendEmail(ctx, request)
	}

	// Submit the full email request to the queue to preserve all data including tenantId
	if err := s.queueManager.SubmitEmailTask(request); err != nil {
		log.Error("Failed to queue email task", "error", err, "email", request.To, "template", request.Template)
		// Fallback to direct sending
		return s.emailService.SendEmail(ctx, request)
	}

	log.Info("Email queued successfully", "email", request.To, "template", request.Template)
	return nil
}

// SendPasswordResetEmailAsync sends a password reset email asynchronously
func (s *asyncEmailService) SendPasswordResetEmailAsync(ctx context.Context, email, token string) error {
	// Extract tenant ID from context
	tenantID := 0
	if ctxTenantID := ctx.Value("tenantId"); ctxTenantID != nil {
		if tid, ok := ctxTenantID.(int); ok {
			tenantID = tid
		}
	}

	request := ports.EmailRequest{
		To:       email,
		Template: ports.EmailTemplateTypePasswordReset,
		TenantID: tenantID,
		Data: map[string]interface{}{
			"email":    email,
			"token":    token,
			"tenantId": tenantID,
		},
	}
	return s.SendEmailAsync(ctx, request)
}

// SendEmailVerificationEmailAsync sends an email verification email asynchronously
func (s *asyncEmailService) SendEmailVerificationEmailAsync(ctx context.Context, email, token string) error {
	// Extract tenant ID from context
	tenantID := 0
	if ctxTenantID := ctx.Value("tenantId"); ctxTenantID != nil {
		if tid, ok := ctxTenantID.(int); ok {
			tenantID = tid
		}
	}

	request := ports.EmailRequest{
		To:       email,
		Template: ports.EmailTemplateTypeEmailVerification,
		TenantID: tenantID,
		Data: map[string]interface{}{
			"email":    email,
			"token":    token,
			"tenantId": tenantID,
		},
	}
	return s.SendEmailAsync(ctx, request)
}

// SendWelcomeEmailAsync sends a welcome email asynchronously
func (s *asyncEmailService) SendWelcomeEmailAsync(ctx context.Context, email string) error {
	// Extract tenant ID from context
	tenantID := 0
	if ctxTenantID := ctx.Value("tenantId"); ctxTenantID != nil {
		if tid, ok := ctxTenantID.(int); ok {
			tenantID = tid
		}
	}

	request := ports.EmailRequest{
		To:       email,
		Template: ports.EmailTemplateTypeWelcome,
		TenantID: tenantID,
		Data: map[string]interface{}{
			"email":    email,
			"tenantId": tenantID,
		},
	}
	return s.SendEmailAsync(ctx, request)
}
