// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// GetEmailTemplates returns a handler function that retrieves email templates with pagination
// @Summary Get email templates
// @Description Retrieve email templates for the current tenant with pagination support
// @Tags email-templates,admin
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of templates to return" default(20)
// @Param offset query int false "Number of templates to skip" default(0)
// @Success 200 {object} ports.EmailTemplatesResponse "Paginated list of email templates"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/admin/system/email-templates [get]
func GetEmailTemplates(templateService services.EmailTemplateService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse query parameters
		limit := c.QueryInt("limit", 20)
		offset := c.QueryInt("offset", 0)

		// Validate parameters
		if limit <= 0 || limit > 100 {
			return middleware.SendError(c, fiber.StatusBadRequest, "limit must be between 1 and 100")
		}
		if offset < 0 {
			return middleware.SendError(c, fiber.StatusBadRequest, "offset must be non-negative")
		}

		templates, pagination, err := templateService.GetEmailTemplates(c.UserContext(), limit, offset)
		if err != nil {
			log.Error("Failed to get email templates", "error", err)
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		response := ports.EmailTemplatesResponse{
			Templates:  templates,
			Pagination: pagination,
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   response,
		})
	}
}

// GetEmailTemplateByID returns a handler function that retrieves a specific email template
// @Summary Get email template by ID
// @Description Retrieve a specific email template by its ID
// @Tags email-templates,admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Template ID"
// @Success 200 {object} ports.EmailTemplate "Email template details"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid ID"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Template not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/admin/system/email-templates/{id} [get]
func GetEmailTemplateByID(templateService services.EmailTemplateService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid template ID")
		}

		template, err := templateService.GetEmailTemplateByID(c.UserContext(), id)
		if err != nil {
			log.Error("Failed to get email template", "error", err, "id", id)
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   template,
		})
	}
}

// CreateEmailTemplate returns a handler function that creates a new email template
// @Summary Create email template
// @Description Create a new email template for the current tenant
// @Tags email-templates,admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param template body ports.NewEmailTemplate true "Email template data"
// @Success 201 {object} ports.EmailTemplate "Created email template"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid data"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 409 {object} domain.ErrorResponse "Template type already exists"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/admin/system/email-templates [post]
func CreateEmailTemplate(templateService services.EmailTemplateService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var newTemplate ports.NewEmailTemplate
		if err := c.BodyParser(&newTemplate); err != nil {
			log.Error("Failed to parse request body", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid request body")
		}

		// Basic validation
		if newTemplate.Name == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "Template name is required")
		}
		if newTemplate.Subject == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "Template subject is required")
		}
		if newTemplate.BodyText == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "Template text body is required")
		}
		if newTemplate.BodyHTML == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "Template HTML body is required")
		}

		template, err := templateService.CreateEmailTemplate(c.UserContext(), newTemplate)
		if err != nil {
			log.Error("Failed to create email template", "error", err)
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status": "success",
			"data":   template,
		})
	}
}

// UpdateEmailTemplate returns a handler function that updates an existing email template
// @Summary Update email template
// @Description Update an existing email template
// @Tags email-templates,admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Template ID"
// @Param updates body ports.UpdateEmailTemplate true "Template updates"
// @Success 200 {object} ports.EmailTemplate "Updated email template"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid data"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Template not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/admin/system/email-templates/{id} [put]
func UpdateEmailTemplate(templateService services.EmailTemplateService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid template ID")
		}

		var updates ports.UpdateEmailTemplate
		if err := c.BodyParser(&updates); err != nil {
			log.Error("Failed to parse request body", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid request body")
		}

		template, err := templateService.UpdateEmailTemplate(c.UserContext(), id, updates)
		if err != nil {
			log.Error("Failed to update email template", "error", err, "id", id)
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   template,
		})
	}
}

// DeleteEmailTemplate returns a handler function that deletes an email template
// @Summary Delete email template
// @Description Delete an email template (hard delete)
// @Tags email-templates,admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Template ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid ID"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 404 {object} domain.ErrorResponse "Template not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/admin/system/email-templates/{id} [delete]
func DeleteEmailTemplate(templateService services.EmailTemplateService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid template ID")
		}

		log.Info("🗑️ Starting email template deletion", "templateId", id)

		if err := templateService.DeleteEmailTemplate(c.UserContext(), id); err != nil {
			log.Error("Failed to delete email template", "error", err, "id", id)
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		log.Info("✅ Email template deleted successfully", "templateId", id)

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Email template deleted successfully",
		})
	}
}

// GetTemplateVariables returns a handler function that retrieves available template variables
// @Summary Get template variables
// @Description Get available variables for each email template type
// @Tags email-templates,admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string][]ports.TemplateVariable "Available variables by template type"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/admin/system/email-templates/variables [get]
func GetTemplateVariables(templateService services.EmailTemplateService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		variables, err := templateService.GetTemplateVariables(c.UserContext())
		if err != nil {
			log.Error("Failed to get template variables", "error", err)
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   variables,
		})
	}
}

// CreateDefaultTemplates returns a handler function that creates default templates for the current tenant
// @Summary Create default templates
// @Description Create default email templates for the current tenant
// @Tags email-templates,admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Success message"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/admin/system/email-templates/defaults [post]
func CreateDefaultTemplates(templateService services.EmailTemplateService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get tenant ID from context
		tenantID, err := middleware.GetTenantIDFromContext(c.UserContext())
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "Failed to get tenant ID")
		}

		if err := templateService.CreateDefaultTemplatesForTenant(c.UserContext(), tenantID); err != nil {
			log.Error("Failed to create default templates", "error", err, "tenantID", tenantID)
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Default email templates created successfully",
		})
	}
}

// PreviewTemplate returns a handler function that previews a template with sample data
// @Summary Preview email template
// @Description Preview an email template with sample data
// @Tags email-templates,admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param templateType query string true "Template type" Enums(password_reset,email_verification,welcome)
// @Param data body map[string]interface{} false "Template data for preview"
// @Success 200 {object} ports.RenderedEmail "Rendered email preview"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid template type"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/admin/system/email-templates/preview [post]
func PreviewTemplate(templateService services.EmailTemplateService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		templateTypeStr := c.Query("templateType")
		if templateTypeStr == "" {
			return middleware.SendError(c, fiber.StatusBadRequest, "templateType query parameter is required")
		}

		templateType := ports.EmailTemplateType(templateTypeStr)

		// Validate template type
		switch templateType {
		case ports.EmailTemplateTypePasswordReset, ports.EmailTemplateTypeEmailVerification, ports.EmailTemplateTypeWelcome:
			// Valid types
		default:
			return middleware.SendError(c, fiber.StatusBadRequest, "Invalid template type")
		}

		// Parse data from request body, use defaults if not provided
		var data map[string]interface{}
		if err := c.BodyParser(&data); err != nil {
			// Use default sample data if parsing fails
			data = getSampleData(templateType)
		}

		// Ensure we have some sample data
		if data == nil {
			data = getSampleData(templateType)
		}

		rendered, err := templateService.RenderTemplate(c.UserContext(), templateType, data)
		if err != nil {
			log.Error("Failed to render template", "error", err, "templateType", templateType)
			return middleware.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   rendered,
		})
	}
}

// getSampleData returns sample data for template preview
func getSampleData(templateType ports.EmailTemplateType) map[string]interface{} {
	baseData := map[string]interface{}{
		"firstName":    "John",
		"lastName":     "Doe",
		"email":        "john.doe@example.com",
		"tenantDomain": "example.com",
	}

	switch templateType {
	case ports.EmailTemplateTypePasswordReset, ports.EmailTemplateTypeEmailVerification:
		baseData["token"] = "ABC123XYZ789"
	}

	return baseData
}
