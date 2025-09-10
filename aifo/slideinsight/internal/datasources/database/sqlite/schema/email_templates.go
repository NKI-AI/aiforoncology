// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package schema

import "database/sql"

// EmailTemplatesSchema handles email template-related schema creation
type EmailTemplatesSchema struct{}

// CreateTables creates email template-related tables
func (s *EmailTemplatesSchema) CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
		-- Email templates table
		CREATE TABLE IF NOT EXISTS email_templates (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id      INTEGER NOT NULL,
			template_type  TEXT    NOT NULL, -- 'password_reset', 'email_verification', 'welcome', etc.
			name           TEXT    NOT NULL,
			subject        TEXT    NOT NULL,
			body_text      TEXT    NOT NULL, -- Plain text version
			body_html      TEXT    NOT NULL, -- HTML version
			variables      JSON,             -- Available variables for this template
			is_active      BOOLEAN DEFAULT TRUE,
			is_system      BOOLEAN DEFAULT FALSE, -- System templates can't be deleted
			created_by     INTEGER NOT NULL,
			updated_by     INTEGER,
			created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (tenant_id, template_type), -- One template per type per tenant
			FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
			FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
		);
	`)
	return err
}

// CreateTriggers creates email template-related triggers
func (s *EmailTemplatesSchema) CreateTriggers(db *sql.DB) error {
	_, err := db.Exec(`
		-- Trigger for email_templates table
		CREATE TRIGGER IF NOT EXISTS email_templates_updated_at 
		AFTER UPDATE ON email_templates 
		FOR EACH ROW 
		BEGIN
			UPDATE email_templates SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

		-- Trigger to update updated_by when template is modified
		CREATE TRIGGER IF NOT EXISTS email_templates_updated_by 
		AFTER UPDATE ON email_templates 
		FOR EACH ROW 
		WHEN OLD.body_text != NEW.body_text OR OLD.body_html != NEW.body_html OR OLD.subject != NEW.subject
		BEGIN
			UPDATE email_templates SET updated_by = NEW.updated_by WHERE id = NEW.id;
		END;
	`)
	return err
}

// CreateIndexes creates email template-related indexes
func (s *EmailTemplatesSchema) CreateIndexes(db *sql.DB) error {
	_, err := db.Exec(`
		-- Indexes for email_templates
		CREATE INDEX IF NOT EXISTS idx_email_templates_tenant_id ON email_templates(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_email_templates_template_type ON email_templates(template_type);
		CREATE INDEX IF NOT EXISTS idx_email_templates_is_active ON email_templates(is_active);
		CREATE INDEX IF NOT EXISTS idx_email_templates_is_system ON email_templates(is_system);
		CREATE INDEX IF NOT EXISTS idx_email_templates_created_by ON email_templates(created_by);
		
		-- Composite index for common queries
		CREATE INDEX IF NOT EXISTS idx_email_templates_tenant_type_active ON email_templates(tenant_id, template_type, is_active);
	`)
	return err
}
