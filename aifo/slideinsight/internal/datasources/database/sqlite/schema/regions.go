// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package schema

import "database/sql"

// RegionsSchema handles region-related schema creation
type RegionsSchema struct{}

// CreateTables creates region-related tables
func (s *RegionsSchema) CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
		-- Regions table for storing regions of interest on slides
		-- Regions represent semantic areas like patients, tissue sections, etc.
		-- Similar to vector_annotations but specialized for region management
		CREATE TABLE IF NOT EXISTS regions (
			id             TEXT    PRIMARY KEY,                -- UUID stored as TEXT
			slide_id       INTEGER NOT NULL,
			tenant_id      INTEGER NOT NULL,
			actor_type     TEXT    NOT NULL
							CHECK(actor_type IN ('user','model')),
			actor_id       INTEGER NOT NULL,
			creator_id     INTEGER NOT NULL,
			version        INTEGER NOT NULL DEFAULT 1,
			name           TEXT    NOT NULL,                    -- Required for regions (e.g., "Patient 1", "Tissue Section A")
			region_type    TEXT    NOT NULL DEFAULT 'roi'       -- 'roi', 'patient', 'tissue', 'artifact', 'background'
							CHECK(region_type IN ('roi','patient','tissue','artifact','background','other')),
			geometry_data  TEXT    NOT NULL,                    -- GeoJSON geometry (polygon, rectangle, etc.)
			coordinate_system TEXT NOT NULL DEFAULT 'pixel'    -- 'pixel' or 'physical' coordinates
							CHECK(coordinate_system IN ('pixel','physical')),
			area_pixels    INTEGER,                             -- Cached area in pixels for quick filtering
			area_physical  REAL,                               -- Cached area in physical units (μm²) if applicable
			labels         JSON,                               -- Structured labels (e.g., {"diagnosis": "normal", "confidence": 0.95})
			metadata       JSON,                               -- Additional metadata (e.g., {"patient_id": "P123", "section": "A"})
			mutable        BOOLEAN NOT NULL DEFAULT TRUE,      -- Regions are typically user-editable by default
			visible        BOOLEAN NOT NULL DEFAULT TRUE,      -- Whether region is visible in UI
			style_config   JSON,                               -- UI styling (colors, opacity, stroke width, etc.)
			deleted_at     TIMESTAMP NULL,
			deleted_by     INTEGER   NULL,
			created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(slide_id)   REFERENCES slides(id) ON DELETE CASCADE,
			FOREIGN KEY(tenant_id)  REFERENCES tenants(id) ON DELETE CASCADE,
			FOREIGN KEY(creator_id) REFERENCES users(id)   ON DELETE CASCADE,
			FOREIGN KEY(deleted_by) REFERENCES users(id)   ON DELETE SET NULL
		);
	`)
	return err
}

// CreateTriggers creates region-related triggers
func (s *RegionsSchema) CreateTriggers(db *sql.DB) error {
	_, err := db.Exec(`
		-- Trigger for regions table
		CREATE TRIGGER IF NOT EXISTS regions_updated_at 
		AFTER UPDATE ON regions 
		FOR EACH ROW 
		BEGIN
			UPDATE regions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
	`)
	return err
}

// CreateIndexes creates region-related indexes
func (s *RegionsSchema) CreateIndexes(db *sql.DB) error {
	_, err := db.Exec(`
		-- Index for slide-based region lookups (most common query pattern)
		CREATE INDEX IF NOT EXISTS idx_regions_slide_id ON regions(slide_id);
		
		-- Index for region type filtering
		CREATE INDEX IF NOT EXISTS idx_regions_type ON regions(region_type);
		
		-- Index for tenant-based queries
		CREATE INDEX IF NOT EXISTS idx_regions_tenant_id ON regions(tenant_id);
		
		-- Index for creator-based queries
		CREATE INDEX IF NOT EXISTS idx_regions_creator_id ON regions(creator_id);
		
		-- Composite index for slide + visible regions (common UI query)
		CREATE INDEX IF NOT EXISTS idx_regions_slide_visible ON regions(slide_id, visible) WHERE deleted_at IS NULL;
		
		-- Index for area-based filtering (useful for finding large/small regions)
		CREATE INDEX IF NOT EXISTS idx_regions_area_pixels ON regions(area_pixels) WHERE area_pixels IS NOT NULL;
	`)
	return err
}
