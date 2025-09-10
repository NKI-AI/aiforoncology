// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package email_templates

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/gofiber/fiber/v2/log"
)

// Adapter implements the EmailTemplateRepository interface for SQLite
type Adapter struct {
	db *sql.DB
}

// NewAdapter creates a new email templates adapter
func NewAdapter(db *sql.DB) *Adapter {
	return &Adapter{db: db}
}

// GetEmailTemplates retrieves email templates with pagination and filtering
func (a *Adapter) GetEmailTemplates(ctx context.Context, tenantID int, limit, offset int) ([]ports.EmailTemplate, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM email_templates WHERE tenant_id = ?`
	var totalCount int
	if err := a.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&totalCount); err != nil {
		return nil, 0, errors.NewDatabaseQueryError("email templates count", err)
	}

	// Get templates with pagination, joining with tenants table to get tenant name
	query := `
		SELECT et.id, et.tenant_id, et.template_type, et.name, et.subject, et.body_text, et.body_html, 
		       et.variables, et.is_active, et.is_system, et.created_by, et.updated_by, et.created_at, et.updated_at,
		       t.name as tenant_name
		FROM email_templates et
		LEFT JOIN tenants t ON et.tenant_id = t.id
		WHERE et.tenant_id = ? 
		ORDER BY et.template_type ASC, et.created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := a.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, errors.NewDatabaseQueryError("email templates", err)
	}
	defer rows.Close()

	var templates []ports.EmailTemplate
	for rows.Next() {
		template, err := a.scanEmailTemplateWithTenant(rows)
		if err != nil {
			return nil, 0, err
		}
		templates = append(templates, *template)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.NewDatabaseIterateRowsError("email templates", err)
	}

	return templates, totalCount, nil
}

// GetAllEmailTemplates retrieves email templates from all tenants (for superadmin access)
func (a *Adapter) GetAllEmailTemplates(ctx context.Context, limit, offset int) ([]ports.EmailTemplate, int, error) {
	// Get total count across all tenants
	countQuery := `SELECT COUNT(*) FROM email_templates`
	var totalCount int
	if err := a.db.QueryRowContext(ctx, countQuery).Scan(&totalCount); err != nil {
		return nil, 0, errors.NewDatabaseQueryError("email templates count", err)
	}

	// Get templates with pagination, joining with tenants table to get tenant name
	query := `
		SELECT et.id, et.tenant_id, et.template_type, et.name, et.subject, et.body_text, et.body_html, 
		       et.variables, et.is_active, et.is_system, et.created_by, et.updated_by, et.created_at, et.updated_at,
		       t.name as tenant_name
		FROM email_templates et
		LEFT JOIN tenants t ON et.tenant_id = t.id
		ORDER BY t.name ASC, et.template_type ASC, et.created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := a.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, errors.NewDatabaseQueryError("email templates", err)
	}
	defer rows.Close()

	var templates []ports.EmailTemplate
	for rows.Next() {
		template, err := a.scanEmailTemplateWithTenant(rows)
		if err != nil {
			return nil, 0, err
		}
		templates = append(templates, *template)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.NewDatabaseIterateRowsError("email templates", err)
	}

	return templates, totalCount, nil
}

// GetEmailTemplateByID retrieves a specific email template by ID
func (a *Adapter) GetEmailTemplateByID(ctx context.Context, id int) (*ports.EmailTemplate, error) {
	query := `
		SELECT id, tenant_id, template_type, name, subject, body_text, body_html, 
		       variables, is_active, is_system, created_by, updated_by, created_at, updated_at
		FROM email_templates 
		WHERE id = ?`

	row := a.db.QueryRowContext(ctx, query, id)
	return a.scanEmailTemplate(row)
}

// GetEmailTemplateByType retrieves a template by tenant and type
func (a *Adapter) GetEmailTemplateByType(ctx context.Context, tenantID int, templateType ports.EmailTemplateType) (*ports.EmailTemplate, error) {
	query := `
		SELECT id, tenant_id, template_type, name, subject, body_text, body_html, 
		       variables, is_active, is_system, created_by, updated_by, created_at, updated_at
		FROM email_templates 
		WHERE tenant_id = ? AND template_type = ? AND is_active = 1`

	row := a.db.QueryRowContext(ctx, query, tenantID, string(templateType))
	return a.scanEmailTemplate(row)
}

// CreateEmailTemplate creates a new email template
func (a *Adapter) CreateEmailTemplate(ctx context.Context, template ports.NewEmailTemplate) (*ports.EmailTemplate, error) {
	variablesJSON, err := json.Marshal(template.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal variables: %w", err)
	}

	query := `
		INSERT INTO email_templates (tenant_id, template_type, name, subject, body_text, body_html, variables, is_active, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := a.db.ExecContext(ctx, query,
		template.TenantID,
		string(template.TemplateType),
		template.Name,
		template.Subject,
		template.BodyText,
		template.BodyHTML,
		string(variablesJSON),
		template.IsActive,
		template.CreatedBy,
	)
	if err != nil {
		return nil, errors.NewDatabaseInsertError("email template", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.NewDatabaseInsertError("email template (get ID)", err)
	}

	return a.GetEmailTemplateByID(ctx, int(id))
}

// UpdateEmailTemplate updates an existing email template
func (a *Adapter) UpdateEmailTemplate(ctx context.Context, id int, updates ports.UpdateEmailTemplate) (*ports.EmailTemplate, error) {
	setParts := []string{}
	args := []interface{}{}

	if updates.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *updates.Name)
	}
	if updates.Subject != nil {
		setParts = append(setParts, "subject = ?")
		args = append(args, *updates.Subject)
	}
	if updates.BodyText != nil {
		setParts = append(setParts, "body_text = ?")
		args = append(args, *updates.BodyText)
	}
	if updates.BodyHTML != nil {
		setParts = append(setParts, "body_html = ?")
		args = append(args, *updates.BodyHTML)
	}
	if updates.Variables != nil {
		variablesJSON, err := json.Marshal(updates.Variables)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal variables: %w", err)
		}
		setParts = append(setParts, "variables = ?")
		args = append(args, string(variablesJSON))
	}
	if updates.IsActive != nil {
		setParts = append(setParts, "is_active = ?")
		args = append(args, *updates.IsActive)
	}

	// Always update updated_by
	setParts = append(setParts, "updated_by = ?")
	args = append(args, updates.UpdatedBy)

	if len(setParts) == 1 { // Only updated_by was set
		return a.GetEmailTemplateByID(ctx, id)
	}

	// Build the query using strings.Join for clean comma separation
	query := fmt.Sprintf("UPDATE email_templates SET %s WHERE id = ?", strings.Join(setParts, ", "))
	args = append(args, id)

	_, err := a.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.NewDatabaseUpdateError("email template", err)
	}

	return a.GetEmailTemplateByID(ctx, id)
}

// DeleteEmailTemplate deletes an email template (hard delete)
func (a *Adapter) DeleteEmailTemplate(ctx context.Context, id int, allowSystemDeletion bool) error {
	// Check if it's a system template first
	template, err := a.GetEmailTemplateByID(ctx, id)
	if err != nil {
		return err
	}

	log.Info("🔍 Adapter: Checking template before deletion",
		"templateId", id,
		"templateName", template.Name,
		"isSystem", template.IsSystem,
		"allowSystemDeletion", allowSystemDeletion)

	// Only prevent system template deletion if allowSystemDeletion is false
	if template.IsSystem && !allowSystemDeletion {
		return fmt.Errorf("cannot delete system template")
	}

	log.Info("🗑️ Adapter: Executing DELETE query", "templateId", id)
	query := `DELETE FROM email_templates WHERE id = ?`
	result, err := a.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Error("❌ Adapter: DELETE query failed", "templateId", id, "error", err)
		return errors.NewDatabaseUpdateError("email template (delete)", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error("❌ Adapter: Failed to check rows affected", "templateId", id, "error", err)
		return errors.NewDatabaseUpdateError("email template (check rows affected)", err)
	}

	log.Info("✅ Adapter: DELETE query completed", "templateId", id, "rowsAffected", rowsAffected)

	if rowsAffected == 0 {
		return errors.WithDetails(errors.ErrNotFound, "email template")
	}

	return nil
}

// CreateDefaultTemplates creates default system templates for a tenant
func (a *Adapter) CreateDefaultTemplates(ctx context.Context, tenantID int, createdBy int) error {
	defaultTemplates := ports.GetDefaultTemplates()
	defaultVariables := ports.GetDefaultTemplateVariables()

	for templateType, template := range defaultTemplates {
		variables := defaultVariables[templateType]
		variablesMap := make(map[string]interface{})
		for _, v := range variables {
			variablesMap[v.Name] = map[string]interface{}{
				"description": v.Description,
				"required":    v.Required,
				"type":        v.Type,
			}
		}

		variablesJSON, err := json.Marshal(variablesMap)
		if err != nil {
			return fmt.Errorf("failed to marshal variables for %s: %w", templateType, err)
		}

		query := `
			INSERT OR REPLACE INTO email_templates 
			(tenant_id, template_type, name, subject, body_text, body_html, variables, is_active, is_system, created_by)
			VALUES (?, ?, ?, ?, ?, ?, ?, 1, 1, ?)`

		_, err = a.db.ExecContext(ctx, query,
			tenantID,
			string(templateType),
			template.Name,
			template.Subject,
			template.BodyText,
			template.BodyHTML,
			string(variablesJSON),
			createdBy,
		)
		if err != nil {
			return errors.NewDatabaseInsertError(fmt.Sprintf("default email template %s", templateType), err)
		}
	}

	return nil
}

// scanEmailTemplate scans a database row into an EmailTemplate struct
func (a *Adapter) scanEmailTemplate(scanner interface {
	Scan(dest ...interface{}) error
},
) (*ports.EmailTemplate, error) {
	var template ports.EmailTemplate
	var variablesJSON sql.NullString
	var updatedBy sql.NullInt64
	var createdAtStr, updatedAtStr string

	err := scanner.Scan(
		&template.ID,
		&template.TenantID,
		&template.TemplateType,
		&template.Name,
		&template.Subject,
		&template.BodyText,
		&template.BodyHTML,
		&variablesJSON,
		&template.IsActive,
		&template.IsSystem,
		&template.CreatedBy,
		&updatedBy,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.WithDetails(errors.ErrNotFound, "email template")
		}
		return nil, errors.NewDatabaseScanError("email template", err)
	}

	// Parse timestamps
	if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
		template.CreatedAt = createdAt
	}
	if updatedAt, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
		template.UpdatedAt = updatedAt
	}

	// Handle nullable updated_by
	if updatedBy.Valid {
		updatedByInt := int(updatedBy.Int64)
		template.UpdatedBy = &updatedByInt
	}

	// Parse variables JSON
	if variablesJSON.Valid && variablesJSON.String != "" {
		if err := json.Unmarshal([]byte(variablesJSON.String), &template.Variables); err != nil {
			// If variables JSON is invalid, set to empty map instead of failing
			template.Variables = make(map[string]interface{})
		}
	} else {
		template.Variables = make(map[string]interface{})
	}

	return &template, nil
}

// scanEmailTemplateWithTenant scans a database row into an EmailTemplate struct, including tenant name
func (a *Adapter) scanEmailTemplateWithTenant(scanner interface {
	Scan(dest ...interface{}) error
},
) (*ports.EmailTemplate, error) {
	var template ports.EmailTemplate
	var variablesJSON sql.NullString
	var updatedBy sql.NullInt64
	var createdAtStr, updatedAtStr string
	var tenantName sql.NullString // Use sql.NullString to handle NULL values

	err := scanner.Scan(
		&template.ID,
		&template.TenantID,
		&template.TemplateType,
		&template.Name,
		&template.Subject,
		&template.BodyText,
		&template.BodyHTML,
		&variablesJSON,
		&template.IsActive,
		&template.IsSystem,
		&template.CreatedBy,
		&updatedBy,
		&createdAtStr,
		&updatedAtStr,
		&tenantName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.WithDetails(errors.ErrNotFound, "email template")
		}
		return nil, errors.NewDatabaseScanError("email template", err)
	}

	// Parse timestamps
	if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
		template.CreatedAt = createdAt
	}
	if updatedAt, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
		template.UpdatedAt = updatedAt
	}

	// Handle nullable updated_by
	if updatedBy.Valid {
		updatedByInt := int(updatedBy.Int64)
		template.UpdatedBy = &updatedByInt
	}

	// Parse variables JSON
	if variablesJSON.Valid && variablesJSON.String != "" {
		if err := json.Unmarshal([]byte(variablesJSON.String), &template.Variables); err != nil {
			// If variables JSON is invalid, set to empty map instead of failing
			template.Variables = make(map[string]interface{})
		}
	} else {
		template.Variables = make(map[string]interface{})
	}

	// Handle nullable tenant name
	if tenantName.Valid {
		template.TenantName = tenantName.String
	} else {
		template.TenantName = "Unknown Tenant"
	}

	return &template, nil
}
