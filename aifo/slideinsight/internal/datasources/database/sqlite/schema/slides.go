// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package schema

import "database/sql"

// SlidesSchema handles slide-related schema creation
type SlidesSchema struct{}

// CreateTables creates slide-related tables
func (s *SlidesSchema) CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
		-- Slides table: each slide belongs to exactly one case
		CREATE TABLE IF NOT EXISTS slides (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			case_id       INTEGER NOT NULL,
			slide_uid     TEXT    NOT NULL UNIQUE,
			slide_hash TEXT NULL, -- for deduplication, but should be NOT NULL UNIQUE
			slide_name    TEXT,
			slide_uri     TEXT    NOT NULL,
			slide_width   INTEGER,
			slide_height  INTEGER,
			slide_mpp     REAL,
	        image_type_id TEXT    NOT NULL DEFAULT 'img_type_unspec', -- FK to image_types
			creator_id    INTEGER,
			tenant_id     INTEGER NOT NULL,
			metadata      JSON,
			deleted_at    TIMESTAMP NULL,
			deleted_by    INTEGER NULL,
			created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE,
			FOREIGN KEY (image_type_id) REFERENCES image_types(id) ON DELETE SET DEFAULT,
			FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE SET NULL,
			FOREIGN KEY (deleted_by) REFERENCES users(id) ON DELETE SET NULL,
			FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
		);
	`)
	return err
}

// CreateTriggers creates slide-related triggers
func (s *SlidesSchema) CreateTriggers(db *sql.DB) error {
	_, err := db.Exec(`
		-- Trigger for slides table
		CREATE TRIGGER IF NOT EXISTS slides_updated_at 
		AFTER UPDATE ON slides 
		FOR EACH ROW 
		BEGIN
			UPDATE slides SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
	`)
	return err
}

// CreateIndexes creates slide-related indexes
func (s *SlidesSchema) CreateIndexes(db *sql.DB) error {
	_, err := db.Exec(`
		-- Index for image type lookups
		CREATE INDEX IF NOT EXISTS idx_slides_image_type_id ON slides(image_type_id);
	`)
	return err
}
