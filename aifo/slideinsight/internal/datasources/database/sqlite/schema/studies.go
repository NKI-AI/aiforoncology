// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package schema

import "database/sql"

// StudiesSchema handles study-related schema creation
type StudiesSchema struct{}

// CreateTables creates study-related tables
func (s *StudiesSchema) CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
		-- Studies table
		CREATE TABLE IF NOT EXISTS studies (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id     INTEGER NOT NULL,
			short_uid     TEXT    NOT NULL UNIQUE,
			creator_id    INTEGER NOT NULL,
			name          TEXT    NOT NULL,
			description   TEXT,
			metadata      JSON,
			is_published  BOOLEAN DEFAULT FALSE,
			deleted_at    TIMESTAMP NULL,
			deleted_by    INTEGER NULL,
			created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (tenant_id)  REFERENCES tenants(id) ON DELETE CASCADE,
			FOREIGN KEY (creator_id) REFERENCES users(id)   ON DELETE CASCADE,
			FOREIGN KEY (deleted_by) REFERENCES users(id)   ON DELETE SET NULL
		);
	`)
	return err
}

// CreateTriggers creates study-related triggers
func (s *StudiesSchema) CreateTriggers(db *sql.DB) error {
	_, err := db.Exec(`
		-- Trigger for studies table
		CREATE TRIGGER IF NOT EXISTS studies_updated_at 
		AFTER UPDATE ON studies 
		FOR EACH ROW 
		BEGIN
			UPDATE studies SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
	`)
	return err
}

// CreateIndexes creates study-related indexes
func (s *StudiesSchema) CreateIndexes(db *sql.DB) error {
	// Studies don't need special indexes beyond the foreign keys
	return nil
}
