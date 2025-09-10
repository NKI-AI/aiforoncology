// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package queue

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
)

// EmailTaskType represents the type identifier for email tasks
const EmailTaskType = "email"

// EmailTask implements the Task interface for sending emails
type EmailTask struct {
	id           string
	emailService ports.EmailService
	request      ports.EmailRequest
	maxRetries   int
	retryDelay   time.Duration
}

// generateUUID creates a UUID v4 using crypto/rand
func generateUUID() string {
	uuid := make([]byte, 16)
	rand.Read(uuid)

	// Set version (4) and variant bits according to RFC 4122
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant 10

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// NewEmailTask creates a new email task
func NewEmailTask(emailService ports.EmailService, request ports.EmailRequest) *EmailTask {
	return &EmailTask{
		id:           generateUUID(),
		emailService: emailService,
		request:      request,
		maxRetries:   3,                // Default to 3 retries
		retryDelay:   30 * time.Second, // Default to 30 second delay between retries
	}
}

// NewEmailTaskWithRetryConfig creates a new email task with custom retry configuration
func NewEmailTaskWithRetryConfig(emailService ports.EmailService, request ports.EmailRequest, maxRetries int, retryDelay time.Duration) *EmailTask {
	return &EmailTask{
		id:           generateUUID(),
		emailService: emailService,
		request:      request,
		maxRetries:   maxRetries,
		retryDelay:   retryDelay,
	}
}

// ID returns the unique identifier for this task
func (e *EmailTask) ID() string {
	return e.id
}

// Type returns the task type
func (e *EmailTask) Type() string {
	return EmailTaskType
}

// Execute sends the email using the email service
func (e *EmailTask) Execute(ctx context.Context) error {
	return e.emailService.SendEmail(ctx, e.request)
}

// MaxRetries returns the maximum number of retry attempts
func (e *EmailTask) MaxRetries() int {
	return e.maxRetries
}

// RetryDelay returns the delay between retry attempts
func (e *EmailTask) RetryDelay() time.Duration {
	return e.retryDelay
}

// String returns a human-readable description of the task
func (e *EmailTask) String() string {
	return fmt.Sprintf("EmailTask(id=%s, to=%s, template=%s)",
		e.id, e.request.To, e.request.Template)
}

// EmailTaskBuilder provides a fluent interface for building email tasks
type EmailTaskBuilder struct {
	emailService ports.EmailService
	request      ports.EmailRequest
	maxRetries   int
	retryDelay   time.Duration
}

// NewEmailTaskBuilder creates a new email task builder
func NewEmailTaskBuilder(emailService ports.EmailService) *EmailTaskBuilder {
	return &EmailTaskBuilder{
		emailService: emailService,
		maxRetries:   3,
		retryDelay:   30 * time.Second,
	}
}

// To sets the recipient email address
func (b *EmailTaskBuilder) To(email string) *EmailTaskBuilder {
	b.request.To = email
	return b
}

// Template sets the email template
func (b *EmailTaskBuilder) Template(template ports.EmailTemplateType) *EmailTaskBuilder {
	b.request.Template = template
	return b
}

// Data sets the template data
func (b *EmailTaskBuilder) Data(data map[string]interface{}) *EmailTaskBuilder {
	b.request.Data = data
	return b
}

// MaxRetries sets the maximum number of retry attempts
func (b *EmailTaskBuilder) MaxRetries(maxRetries int) *EmailTaskBuilder {
	b.maxRetries = maxRetries
	return b
}

// RetryDelay sets the delay between retry attempts
func (b *EmailTaskBuilder) RetryDelay(delay time.Duration) *EmailTaskBuilder {
	b.retryDelay = delay
	return b
}

// Build creates the email task
func (b *EmailTaskBuilder) Build() *EmailTask {
	return NewEmailTaskWithRetryConfig(b.emailService, b.request, b.maxRetries, b.retryDelay)
}

// BuildPasswordResetTask creates a password reset email task
func (b *EmailTaskBuilder) BuildPasswordResetTask(email, token string) *EmailTask {
	return b.To(email).
		Template(ports.EmailTemplateTypePasswordReset).
		Data(map[string]interface{}{
			"email": email,
			"token": token,
		}).
		Build()
}

// BuildEmailVerificationTask creates an email verification task
func (b *EmailTaskBuilder) BuildEmailVerificationTask(email, token string) *EmailTask {
	return b.To(email).
		Template(ports.EmailTemplateTypeEmailVerification).
		Data(map[string]interface{}{
			"email": email,
			"token": token,
		}).
		Build()
}

// BuildWelcomeTask creates a welcome email task
func (b *EmailTaskBuilder) BuildWelcomeTask(email string) *EmailTask {
	return b.To(email).
		Template(ports.EmailTemplateTypeWelcome).
		Data(map[string]interface{}{
			"email": email,
		}).
		Build()
}
