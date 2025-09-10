// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/gofiber/fiber/v2/log"
)

// EmailTemplateService provides business logic for email template operations
type EmailTemplateService interface {
	// GetEmailTemplates retrieves email templates with pagination and filtering
	GetEmailTemplates(ctx context.Context, limit, offset int) ([]ports.EmailTemplate, domain.PaginationInfo, error)

	// GetEmailTemplateByID retrieves a specific email template by ID
	GetEmailTemplateByID(ctx context.Context, id int) (*ports.EmailTemplate, error)

	// CreateEmailTemplate creates a new email template
	CreateEmailTemplate(ctx context.Context, template ports.NewEmailTemplate) (*ports.EmailTemplate, error)

	// UpdateEmailTemplate updates an existing email template
	UpdateEmailTemplate(ctx context.Context, id int, updates ports.UpdateEmailTemplate) (*ports.EmailTemplate, error)

	// DeleteEmailTemplate deletes an email template
	DeleteEmailTemplate(ctx context.Context, id int) error

	// GetTemplateVariables returns available variables for each template type
	GetTemplateVariables(ctx context.Context) (map[ports.EmailTemplateType][]ports.TemplateVariable, error)

	// CreateDefaultTemplatesForTenant creates default templates for a tenant
	CreateDefaultTemplatesForTenant(ctx context.Context, tenantID int) error

	// RenderTemplate renders a template with provided data
	RenderTemplate(ctx context.Context, templateType ports.EmailTemplateType, data map[string]interface{}) (*ports.RenderedEmail, error)
}

// emailTemplateService implements EmailTemplateService
type emailTemplateService struct {
	*BaseService // Inherit from BaseService to get auth context methods
	repo         ports.EmailTemplateRepository
	userID       int // Current user ID for authorization and audit trails
}

// NewEmailTemplateService creates a new email template service
func NewEmailTemplateService(repo ports.EmailTemplateRepository, userID int) EmailTemplateService {
	return &emailTemplateService{
		BaseService: NewBaseService(repo.(ports.Database)), // Assuming repo is the database
		repo:        repo,
		userID:      userID,
	}
}

// GetEmailTemplates retrieves email templates with pagination and filtering
func (s *emailTemplateService) GetEmailTemplates(ctx context.Context, limit, offset int) ([]ports.EmailTemplate, domain.PaginationInfo, error) {
	// Get authentication context to check if user is superadmin
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return nil, domain.PaginationInfo{}, fmt.Errorf("failed to get auth context: %w", err)
	}

	var templates []ports.EmailTemplate
	var totalCount int

	if authCtx.IsSuperAdmin {
		// Superadmin can see all templates across all tenants
		templates, totalCount, err = s.repo.GetAllEmailTemplates(ctx, limit, offset)
		if err != nil {
			return nil, domain.PaginationInfo{}, fmt.Errorf("failed to get all email templates: %w", err)
		}
	} else {
		// Regular users see only templates for their tenant
		templates, totalCount, err = s.repo.GetEmailTemplates(ctx, authCtx.TenantID, limit, offset)
		if err != nil {
			return nil, domain.PaginationInfo{}, fmt.Errorf("failed to get email templates: %w", err)
		}
	}

	// Calculate pagination info
	totalPages := 0
	if totalCount > 0 && limit > 0 {
		totalPages = (totalCount + limit - 1) / limit
	}

	pagination := domain.PaginationInfo{
		Page:       (offset / limit) + 1,
		Limit:      limit,
		Total:      totalCount,
		TotalPages: totalPages,
		HasNext:    offset+limit < totalCount,
		HasPrev:    offset > 0,
	}

	return templates, pagination, nil
}

// GetEmailTemplateByID retrieves a specific email template by ID
func (s *emailTemplateService) GetEmailTemplateByID(ctx context.Context, id int) (*ports.EmailTemplate, error) {
	template, err := s.repo.GetEmailTemplateByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get email template: %w", err)
	}

	// Get authentication context to check if user is superadmin
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth context: %w", err)
	}

	// Superadmins can access templates from any tenant
	if !authCtx.IsSuperAdmin {
		// Regular users can only access templates from their own tenant
		if template.TenantID != authCtx.TenantID {
			return nil, fmt.Errorf("template not found")
		}
	}

	return template, nil
}

// CreateEmailTemplate creates a new email template
func (s *emailTemplateService) CreateEmailTemplate(ctx context.Context, template ports.NewEmailTemplate) (*ports.EmailTemplate, error) {
	// Get authentication context
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth context: %w", err)
	}

	// Set tenant ID and created by
	template.TenantID = authCtx.TenantID
	template.CreatedBy = s.userID

	// Validate that variables are properly formatted
	if err := s.validateTemplateVariables(template.BodyText, template.BodyHTML, template.TemplateType); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	// Check if a template with this type already exists for this tenant
	existingTemplate, err := s.repo.GetEmailTemplateByType(ctx, authCtx.TenantID, template.TemplateType)
	if err == nil && existingTemplate != nil {
		return nil, fmt.Errorf("template of type %s already exists for this tenant", template.TemplateType)
	}

	createdTemplate, err := s.repo.CreateEmailTemplate(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("failed to create email template: %w", err)
	}

	return createdTemplate, nil
}

// UpdateEmailTemplate updates an existing email template
func (s *emailTemplateService) UpdateEmailTemplate(ctx context.Context, id int, updates ports.UpdateEmailTemplate) (*ports.EmailTemplate, error) {
	// Check that the template exists and belongs to current tenant
	template, err := s.GetEmailTemplateByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get authentication context to check if user is superadmin
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth context: %w", err)
	}

	// Check if it's a system template and prevent certain modifications (unless user is superadmin)
	if template.IsSystem && !authCtx.IsSuperAdmin {
		// For system templates, only allow name changes for non-superadmin users
		if updates.BodyText != nil || updates.BodyHTML != nil || updates.Subject != nil {
			return nil, fmt.Errorf("cannot modify content of system templates")
		}
	}

	// Set updated by
	updates.UpdatedBy = s.userID

	// Validate template content if being updated
	if updates.BodyText != nil || updates.BodyHTML != nil {
		bodyText := template.BodyText
		bodyHTML := template.BodyHTML
		if updates.BodyText != nil {
			bodyText = *updates.BodyText
		}
		if updates.BodyHTML != nil {
			bodyHTML = *updates.BodyHTML
		}
		if err := s.validateTemplateVariables(bodyText, bodyHTML, template.TemplateType); err != nil {
			return nil, fmt.Errorf("template validation failed: %w", err)
		}
	}

	updatedTemplate, err := s.repo.UpdateEmailTemplate(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update email template: %w", err)
	}

	return updatedTemplate, nil
}

// DeleteEmailTemplate deletes an email template
func (s *emailTemplateService) DeleteEmailTemplate(ctx context.Context, id int) error {
	// Check that the template exists and belongs to current tenant
	template, err := s.GetEmailTemplateByID(ctx, id)
	if err != nil {
		return err
	}

	// Get authentication context to check if user is superadmin
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth context: %w", err)
	}

	log.Info("🔍 Checking template deletion permissions",
		"templateId", id,
		"templateName", template.Name,
		"isSystem", template.IsSystem,
		"isSuperAdmin", authCtx.IsSuperAdmin,
		"tenantId", template.TenantID)

	// Superadmins can delete system templates, regular users cannot
	allowSystemDeletion := authCtx.IsSuperAdmin

	// For system templates, only superadmins can delete them
	if template.IsSystem && !authCtx.IsSuperAdmin {
		return fmt.Errorf("system templates can only be deleted by superadmins")
	}

	log.Info("🗑️ Proceeding with template deletion",
		"templateId", id,
		"allowSystemDeletion", allowSystemDeletion)

	if err := s.repo.DeleteEmailTemplate(ctx, id, allowSystemDeletion); err != nil {
		return fmt.Errorf("failed to delete email template: %w", err)
	}

	log.Info("✅ Template deletion completed at service layer", "templateId", id)
	return nil
}

// GetTemplateVariables returns available variables for each template type
func (s *emailTemplateService) GetTemplateVariables(ctx context.Context) (map[ports.EmailTemplateType][]ports.TemplateVariable, error) {
	return ports.GetDefaultTemplateVariables(), nil
}

// CreateDefaultTemplatesForTenant creates default templates for a tenant
func (s *emailTemplateService) CreateDefaultTemplatesForTenant(ctx context.Context, tenantID int) error {
	// Get authentication context to verify permissions
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth context: %w", err)
	}

	// Only superadmins can create templates for other tenants
	if !authCtx.IsSuperAdmin && tenantID != authCtx.TenantID {
		return fmt.Errorf("insufficient permissions to create templates for other tenants")
	}

	if err := s.repo.CreateDefaultTemplates(ctx, tenantID, s.userID); err != nil {
		return fmt.Errorf("failed to create default templates: %w", err)
	}
	return nil
}

// RenderTemplate renders a template with provided data
func (s *emailTemplateService) RenderTemplate(ctx context.Context, templateType ports.EmailTemplateType, data map[string]interface{}) (*ports.RenderedEmail, error) {
	// Get authentication context
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth context: %w", err)
	}

	// Get the template
	emailTemplate, err := s.repo.GetEmailTemplateByType(ctx, authCtx.TenantID, templateType)
	if err != nil {
		return nil, fmt.Errorf("failed to get email template: %w", err)
	}

	// Render subject
	subjectTmpl, err := template.New("subject").Parse(emailTemplate.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subject template: %w", err)
	}
	var subjectBuf strings.Builder
	if err := subjectTmpl.Execute(&subjectBuf, data); err != nil {
		return nil, fmt.Errorf("failed to render subject: %w", err)
	}

	// Render text body
	textTmpl, err := template.New("text").Parse(emailTemplate.BodyText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse text template: %w", err)
	}
	var textBuf strings.Builder
	if err := textTmpl.Execute(&textBuf, data); err != nil {
		return nil, fmt.Errorf("failed to render text body: %w", err)
	}

	// Render HTML body
	htmlTmpl, err := template.New("html").Parse(emailTemplate.BodyHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML template: %w", err)
	}
	var htmlBuf strings.Builder
	if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
		return nil, fmt.Errorf("failed to render HTML body: %w", err)
	}

	return &ports.RenderedEmail{
		Subject:  subjectBuf.String(),
		BodyText: textBuf.String(),
		BodyHTML: htmlBuf.String(),
	}, nil
}

// validateTemplateVariables validates that template variables are properly formatted
func (s *emailTemplateService) validateTemplateVariables(bodyText, bodyHTML string, templateType ports.EmailTemplateType) error {
	// Get expected variables for this template type
	expectedVars := ports.GetDefaultTemplateVariables()[templateType]

	// Check that all required variables are present in both text and HTML versions
	for _, variable := range expectedVars {
		if variable.Required {
			if !strings.Contains(bodyText, variable.Name) {
				return fmt.Errorf("required variable %s missing from text body", variable.Name)
			}
			if !strings.Contains(bodyHTML, variable.Name) {
				return fmt.Errorf("required variable %s missing from HTML body", variable.Name)
			}
		}
	}

	// Try to parse templates to ensure they're valid Go templates
	if _, err := template.New("test-text").Parse(bodyText); err != nil {
		return fmt.Errorf("invalid text template syntax: %w", err)
	}
	if _, err := template.New("test-html").Parse(bodyHTML); err != nil {
		return fmt.Errorf("invalid HTML template syntax: %w", err)
	}

	return nil
}
