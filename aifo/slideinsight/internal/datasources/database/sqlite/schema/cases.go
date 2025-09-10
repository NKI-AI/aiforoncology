// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package schema

import "database/sql"

// CasesSchema handles case-related schema creation
type CasesSchema struct{}

// CreateTables creates case-related tables
func (s *CasesSchema) CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
		-- Cases table: each case can be linked to studies via study_cases table
		CREATE TABLE IF NOT EXISTS cases (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id     INTEGER NOT NULL,
			creator_id    INTEGER NOT NULL,
			short_uid     TEXT    NOT NULL UNIQUE,
			name          TEXT    NOT NULL,
			metadata      JSON,
			deleted_at    TIMESTAMP NULL,
			deleted_by    INTEGER NULL,
			created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (creator_id) REFERENCES users(id)   ON DELETE CASCADE,
			FOREIGN KEY (tenant_id)  REFERENCES tenants(id) ON DELETE CASCADE,
			FOREIGN KEY (deleted_by) REFERENCES users(id)   ON DELETE SET NULL
		);

		CREATE TABLE IF NOT EXISTS study_cases (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			study_id     INTEGER NOT NULL,
			tenant_id    INTEGER NOT NULL,
			case_id      INTEGER NOT NULL,
			created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(study_id, case_id),
			FOREIGN KEY (study_id) REFERENCES studies(id) ON DELETE CASCADE,
			FOREIGN KEY (case_id)  REFERENCES cases(id)   ON DELETE CASCADE,
			FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
		);
	`)
	return err
}

// CreateTriggers creates case-related triggers
func (s *CasesSchema) CreateTriggers(db *sql.DB) error {
	_, err := db.Exec(`
		-- Trigger for cases table
		CREATE TRIGGER IF NOT EXISTS cases_updated_at 
		AFTER UPDATE ON cases 
		FOR EACH ROW 
		BEGIN
			UPDATE cases SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

		-- Trigger for study_cases table
		CREATE TRIGGER IF NOT EXISTS study_cases_updated_at 
		AFTER UPDATE ON study_cases 
		FOR EACH ROW 
		BEGIN
			UPDATE study_cases SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
	`)
	return err
}

// CreateIndexes creates case-related indexes
func (s *CasesSchema) CreateIndexes(db *sql.DB) error {
	// Cases don't need special indexes beyond the foreign keys
	return nil
}
