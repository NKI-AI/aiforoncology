// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package ports

import (
	"context"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
)

// EmailTemplate represents an email template stored in the database
type EmailTemplate struct {
	ID           int                    `json:"id"`
	TenantID     int                    `json:"tenantId"`
	TenantName   string                 `json:"tenantName,omitempty"`
	TemplateType EmailTemplateType      `json:"templateType"`
	Name         string                 `json:"name"`
	Subject      string                 `json:"subject"`
	BodyText     string                 `json:"bodyText"`
	BodyHTML     string                 `json:"bodyHtml"`
	Variables    map[string]interface{} `json:"variables"`
	IsActive     bool                   `json:"isActive"`
	IsSystem     bool                   `json:"isSystem"`
	CreatedBy    int                    `json:"createdBy"`
	UpdatedBy    *int                   `json:"updatedBy"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

// EmailTemplateType represents the type of email template
type EmailTemplateType string

const (
	EmailTemplateTypePasswordReset     EmailTemplateType = "password_reset"
	EmailTemplateTypeEmailVerification EmailTemplateType = "email_verification"
	EmailTemplateTypeWelcome           EmailTemplateType = "welcome"
)

// NewEmailTemplate represents the data needed to create a new email template
type NewEmailTemplate struct {
	TenantID     int                    `json:"tenantId" validate:"required"`
	TemplateType EmailTemplateType      `json:"templateType" validate:"required"`
	Name         string                 `json:"name" validate:"required,min=1,max=255"`
	Subject      string                 `json:"subject" validate:"required,min=1,max=255"`
	BodyText     string                 `json:"bodyText" validate:"required,min=1"`
	BodyHTML     string                 `json:"bodyHtml" validate:"required,min=1"`
	Variables    map[string]interface{} `json:"variables"`
	IsActive     bool                   `json:"isActive"`
	CreatedBy    int                    `json:"createdBy" validate:"required"`
}

// UpdateEmailTemplate represents the data that can be updated in an email template
type UpdateEmailTemplate struct {
	Name      *string                `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Subject   *string                `json:"subject,omitempty" validate:"omitempty,min=1,max=255"`
	BodyText  *string                `json:"bodyText,omitempty" validate:"omitempty,min=1"`
	BodyHTML  *string                `json:"bodyHtml,omitempty" validate:"omitempty,min=1"`
	Variables map[string]interface{} `json:"variables,omitempty"`
	IsActive  *bool                  `json:"isActive,omitempty"`
	UpdatedBy int                    `json:"updatedBy" validate:"required"`
}

// EmailTemplateVariables defines the available variables for each template type
type EmailTemplateVariables struct {
	TemplateType EmailTemplateType  `json:"templateType"`
	Variables    []TemplateVariable `json:"variables"`
}

// TemplateVariable represents a variable that can be used in email templates
type TemplateVariable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type"` // "string", "url", "datetime", etc.
}

// EmailTemplatesResponse represents the response for email templates list
type EmailTemplatesResponse struct {
	Templates  []EmailTemplate       `json:"templates"`
	Pagination domain.PaginationInfo `json:"pagination"`
}

// RenderedEmail represents a rendered email template
type RenderedEmail struct {
	Subject  string `json:"subject"`
	BodyText string `json:"bodyText"`
	BodyHTML string `json:"bodyHtml"`
}

// EmailTemplateRepository defines the interface for email template data operations
type EmailTemplateRepository interface {
	// GetEmailTemplates retrieves email templates with pagination and filtering
	GetEmailTemplates(ctx context.Context, tenantID int, limit, offset int) ([]EmailTemplate, int, error)

	// GetAllEmailTemplates retrieves email templates from all tenants (for superadmin access)
	GetAllEmailTemplates(ctx context.Context, limit, offset int) ([]EmailTemplate, int, error)

	// GetEmailTemplateByID retrieves a specific email template by ID
	GetEmailTemplateByID(ctx context.Context, id int) (*EmailTemplate, error)

	// GetEmailTemplateByType retrieves a template by tenant and type
	GetEmailTemplateByType(ctx context.Context, tenantID int, templateType EmailTemplateType) (*EmailTemplate, error)

	// CreateEmailTemplate creates a new email template
	CreateEmailTemplate(ctx context.Context, template NewEmailTemplate) (*EmailTemplate, error)

	// UpdateEmailTemplate updates an existing email template
	UpdateEmailTemplate(ctx context.Context, id int, updates UpdateEmailTemplate) (*EmailTemplate, error)

	// DeleteEmailTemplate deletes an email template (soft delete)
	// allowSystemDeletion: if true, allows deletion of system templates (for superadmins)
	DeleteEmailTemplate(ctx context.Context, id int, allowSystemDeletion bool) error

	// CreateDefaultTemplates creates default system templates for a tenant
	CreateDefaultTemplates(ctx context.Context, tenantID int, createdBy int) error
}

// GetDefaultTemplateVariables returns the available variables for each template type
func GetDefaultTemplateVariables() map[EmailTemplateType][]TemplateVariable {
	return map[EmailTemplateType][]TemplateVariable{
		EmailTemplateTypePasswordReset: {
			{Name: "{{.firstName}}", Description: "User's first name", Required: false, Type: "string"},
			{Name: "{{.lastName}}", Description: "User's last name", Required: false, Type: "string"},
			{Name: "{{.email}}", Description: "User's email address", Required: false, Type: "string"},
			{Name: "{{.token}}", Description: "Password reset token", Required: true, Type: "string"},
			{Name: "{{.tenantDomain}}", Description: "Tenant primary domain", Required: false, Type: "string"},
		},
		EmailTemplateTypeEmailVerification: {
			{Name: "{{.firstName}}", Description: "User's first name", Required: false, Type: "string"},
			{Name: "{{.lastName}}", Description: "User's last name", Required: false, Type: "string"},
			{Name: "{{.email}}", Description: "User's email address", Required: false, Type: "string"},
			{Name: "{{.token}}", Description: "Email verification token", Required: true, Type: "string"},
			{Name: "{{.tenantDomain}}", Description: "Tenant primary domain", Required: false, Type: "string"},
		},
		EmailTemplateTypeWelcome: {
			{Name: "{{.firstName}}", Description: "User's first name", Required: false, Type: "string"},
			{Name: "{{.lastName}}", Description: "User's last name", Required: false, Type: "string"},
			{Name: "{{.email}}", Description: "User's email address", Required: false, Type: "string"},
			{Name: "{{.tenantDomain}}", Description: "Tenant primary domain", Required: false, Type: "string"},
		},
	}
}

// GetDefaultTemplates returns the default email templates with placeholders
func GetDefaultTemplates() map[EmailTemplateType]struct {
	Name     string
	Subject  string
	BodyText string
	BodyHTML string
} {
	return map[EmailTemplateType]struct {
		Name     string
		Subject  string
		BodyText string
		BodyHTML string
	}{
		EmailTemplateTypePasswordReset: {
			Name:    "Password Reset Email",
			Subject: "Password Reset Request - SlideInsight",
			BodyText: `Dear {{.firstName}} {{.lastName}},

You have requested a password reset for your SlideInsight account.

Please use the following token to reset your password:

Reset Token: {{.token}}

This token will expire in 1 hour.

If you didn't request this reset, please ignore this email.

Best regards,
The SlideInsight Team`,
			BodyHTML: `<html>
<body>
<h2>Password Reset Request</h2>
<p>Dear {{.firstName}} {{.lastName}},</p>

<p>You have requested a password reset for your SlideInsight account.</p>

<p>Please use the following token to reset your password:</p>

<div style="background-color: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 4px;">
<strong>Reset Token:</strong> <code>{{.token}}</code>
</div>

<p><strong>This token will expire in 1 hour.</strong></p>

<p>If you didn't request this reset, please ignore this email.</p>

<p>Best regards,<br>
The SlideInsight Team</p>
</body>
</html>`,
		},
		EmailTemplateTypeEmailVerification: {
			Name:    "Email Verification",
			Subject: "Please Verify Your Email Address - SlideInsight",
			BodyText: `Dear {{.firstName}} {{.lastName}},

Welcome to SlideInsight! Please verify your email address to complete your registration.

Please use the following token to verify your email:

Verification Token: {{.token}}

This token will expire in 24 hours.

Thank you for joining us!

Best regards,
The SlideInsight Team`,
			BodyHTML: `<html>
<body>
<h2>Welcome to SlideInsight!</h2>
<p>Dear {{.firstName}} {{.lastName}},</p>

<p>Welcome to SlideInsight! Please verify your email address to complete your registration.</p>

<p>Please use the following token to verify your email:</p>

<div style="background-color: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 4px;">
<strong>Verification Token:</strong> <code>{{.token}}</code>
</div>

<p><strong>This token will expire in 24 hours.</strong></p>

<p>Thank you for joining us!</p>

<p>Best regards,<br>
The SlideInsight Team</p>
</body>
</html>`,
		},
		EmailTemplateTypeWelcome: {
			Name:    "Welcome Email",
			Subject: "Welcome to SlideInsight!",
			BodyText: `Dear {{.firstName}} {{.lastName}},

Welcome to SlideInsight! Your account has been successfully created and verified.

You can now start using SlideInsight to manage and analyze your slides.

If you have any questions or need assistance, please don't hesitate to contact our support team.

Best regards,
The SlideInsight Team`,
			BodyHTML: `<html>
<body>
<h2>Welcome to SlideInsight!</h2>
<p>Dear {{.firstName}} {{.lastName}},</p>

<p>Welcome to SlideInsight! Your account has been successfully created and verified.</p>

<p>You can now start using SlideInsight to manage and analyze your slides.</p>

<p>If you have any questions or need assistance, please don't hesitate to contact our support team.</p>

<p>Best regards,<br>
The SlideInsight Team</p>
</body>
</html>`,
		},
	}
}
