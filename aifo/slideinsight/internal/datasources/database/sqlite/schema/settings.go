// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.

package schema

import "database/sql"

// SettingsSchema handles settings table creation
type SettingsSchema struct{}

// CreateTables creates the settings table
func (s *SettingsSchema) CreateTables(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS settings (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id   INTEGER        NOT NULL DEFAULT 0,        -- 0 = global
			key         TEXT           NOT NULL,
			value_type  TEXT           NOT NULL
			                   CHECK(value_type IN ('boolean','number','string','json')),
			value       JSON           NOT NULL,                   -- JSON stored as TEXT under the hood
			created_at  DATETIME       NOT NULL DEFAULT (CURRENT_TIMESTAMP),
			updated_at  DATETIME       NOT NULL DEFAULT (CURRENT_TIMESTAMP),
			UNIQUE(tenant_id, key),
			FOREIGN KEY(tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
		);
	`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	// Initialize with default settings
	initQuery := `
		INSERT OR IGNORE INTO settings (tenant_id, key, value_type, value)
		VALUES (0, 'enable_registration', 'boolean', 'false');
	`

	_, err = db.Exec(initQuery)
	return err
}

// CreateTriggers creates triggers for the settings table
func (s *SettingsSchema) CreateTriggers(db *sql.DB) error {
	query := `
		CREATE TRIGGER IF NOT EXISTS trg_settings_updated_at
		BEFORE UPDATE ON settings
		FOR EACH ROW
		BEGIN
			UPDATE settings
			   SET updated_at = CURRENT_TIMESTAMP
			 WHERE id = OLD.id;
		END;
	`

	_, err := db.Exec(query)
	return err
}

// CreateIndexes creates indexes for the settings table
func (s *SettingsSchema) CreateIndexes(db *sql.DB) error {
	queries := []string{
		"CREATE INDEX IF NOT EXISTS idx_settings_tenant_key ON settings(tenant_id, key);",
		"CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(key);",
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}
