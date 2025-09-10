// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package schema

import "database/sql"

// AnnotationsSchema handles annotation-related schema creation
type AnnotationsSchema struct{}

// CreateTables creates annotation-related tables
func (s *AnnotationsSchema) CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
		-- Raster annotations
		-- TODO: Add hashes for deduplication

		-- Raster annotations (polymorphic actor)
		CREATE TABLE IF NOT EXISTS raster_annotations (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		slide_id       INTEGER NOT NULL,
		tenant_id      INTEGER NOT NULL,
		actor_type     TEXT    NOT NULL
						CHECK(actor_type IN ('user','model')),
		actor_id       INTEGER NOT NULL,
		version        INTEGER NOT NULL DEFAULT 1,
		raster_uid     TEXT    NOT NULL UNIQUE,
		creator_id     INTEGER NOT NULL,
		name           TEXT,
		file_uri       TEXT    NOT NULL,
		file_hash      TEXT    UNIQUE,  -- Hash for deduplication, nullable and unique
		format         TEXT    CHECK(format IN ('tiff','png')),
		labels         JSON,
		metadata       JSON,    -- generic metadata or model‐specific info
		mask_width     INTEGER,
		mask_height    INTEGER,
		mask_mpp       REAL,
		mutable        BOOLEAN NOT NULL DEFAULT FALSE,  -- Whether the annotation can be modified
		deleted_at     TIMESTAMP NULL,
		deleted_by     INTEGER   NULL,
		created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(slide_id)   REFERENCES slides(id) ON DELETE CASCADE,
		FOREIGN KEY(tenant_id)  REFERENCES tenants(id) ON DELETE CASCADE,
		FOREIGN KEY(creator_id) REFERENCES users(id)   ON DELETE CASCADE,
		FOREIGN KEY(deleted_by) REFERENCES users(id)   ON DELETE SET NULL
		);

		-- Vector annotations (polymorphic actor)
		CREATE TABLE IF NOT EXISTS vector_annotations (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		slide_id       INTEGER NOT NULL,
		tenant_id      INTEGER NOT NULL,
		actor_type     TEXT    NOT NULL
						CHECK(actor_type IN ('user','model')),
		actor_id       INTEGER NOT NULL,
		creator_id     INTEGER NOT NULL,
		vector_uid     TEXT    NOT NULL UNIQUE,
		version        INTEGER NOT NULL DEFAULT 1,
		name           TEXT,
		file_uri       TEXT,    -- Made optional, can be NULL if data_blob is used
		data_blob      TEXT,    -- Inline data (JSON, GeoJSON, etc.) as alternative to file_uri
		format         TEXT    CHECK(format IN ('protobuf','geojson','zarr')),
		labels         JSON,
		metadata       JSON,
		mutable        BOOLEAN NOT NULL DEFAULT FALSE,  -- Whether the annotation can be modified
		deleted_at     TIMESTAMP NULL,
		deleted_by     INTEGER   NULL,
		created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(slide_id)   REFERENCES slides(id) ON DELETE CASCADE,
		FOREIGN KEY(tenant_id)  REFERENCES tenants(id) ON DELETE CASCADE,
		FOREIGN KEY(creator_id) REFERENCES users(id)   ON DELETE CASCADE,
		FOREIGN KEY(deleted_by) REFERENCES users(id)   ON DELETE SET NULL,
		CHECK(file_uri IS NOT NULL OR data_blob IS NOT NULL) -- At least one must be provided
		);
	`)
	return err
}

// CreateTriggers creates annotation-related triggers
func (s *AnnotationsSchema) CreateTriggers(db *sql.DB) error {
	_, err := db.Exec(`
		-- Trigger for raster_annotations table
		CREATE TRIGGER IF NOT EXISTS raster_annotations_updated_at 
		AFTER UPDATE ON raster_annotations 
		FOR EACH ROW 
		BEGIN
			UPDATE raster_annotations SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

		-- Trigger for vector_annotations table
		CREATE TRIGGER IF NOT EXISTS vector_annotations_updated_at 
		AFTER UPDATE ON vector_annotations 
		FOR EACH ROW 
		BEGIN
			UPDATE vector_annotations SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
	`)
	return err
}

// CreateIndexes creates annotation-related indexes
func (s *AnnotationsSchema) CreateIndexes(db *sql.DB) error {
	// Annotations don't need special indexes beyond the foreign keys
	return nil
}
