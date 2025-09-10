// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package schema

import "database/sql"

// ImageTypesSchema handles image type and staining-related schema creation
type ImageTypesSchema struct{}

// CreateTables creates image type–related tables
func (s *ImageTypesSchema) CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
        -- Image types lookup table (system-level, tenant_id = 0)
        CREATE TABLE IF NOT EXISTS image_types (
            id                TEXT      PRIMARY KEY,                       -- UUID stored as TEXT
            tenant_id         INTEGER   NOT NULL DEFAULT 0,               -- 0 = system-level
            type_uid          TEXT      NOT NULL UNIQUE,                  -- e.g. 'brightfield_he'
            name              TEXT      NOT NULL,                         -- e.g. 'Brightfield H&E'
            description       TEXT,                                       -- optional longer description
            category          TEXT      NOT NULL                          -- one of brightfield, fluorescence, other
                                 CHECK(category IN ('brightfield','fluorescence','other')),
            requires_histogram BOOLEAN  NOT NULL DEFAULT FALSE,            -- flag for heavy histogram data
            metadata_schema   JSON,                                       -- JSON schema for per‑type metadata
            is_active         BOOLEAN  NOT NULL DEFAULT TRUE,
            created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );

        -- Slide histograms table for types that need big arrays
        CREATE TABLE IF NOT EXISTS slide_histograms (
            id               TEXT      PRIMARY KEY,                       -- UUID
            slide_id         INTEGER   NOT NULL,                         -- FK to slides.id
            channel_index    INTEGER   NOT NULL DEFAULT 0,
            channel_name     TEXT,
            bin_count        INTEGER   NOT NULL,
            min_value        REAL      NOT NULL,
            max_value        REAL      NOT NULL,
            histogram_data   BLOB      NOT NULL,                         -- compressed binary counts
            metadata         JSON,                                       -- any extra per‑channel fields
            created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(slide_id, channel_index),
            FOREIGN KEY (slide_id) REFERENCES slides(id) ON DELETE CASCADE
        );

        -- Slide staining protocols table
        CREATE TABLE IF NOT EXISTS slide_staining_protocols (
            id               TEXT      PRIMARY KEY,                       -- UUID
            slide_id         INTEGER   NOT NULL,                         -- FK to slides.id
            stain_name       TEXT      NOT NULL,                         -- 'Hematoxylin', 'DAPI', etc.
            stain_type       TEXT      NOT NULL                          -- primary, counterstain, fluorophore, other
                                 CHECK(stain_type IN ('primary','counterstain','fluorophore','other')),
            concentration    TEXT,                                       -- e.g. '1:1000'
            incubation_time  TEXT,                                       -- e.g. '30m'
            antibody_info    JSON,                                       -- clone, supplier, etc.
            excitation_nm    INTEGER,                                    -- for fluorophores
            emission_nm      INTEGER,                                    -- for fluorophores
            metadata         JSON,                                       -- any extra protocol data
            created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (slide_id) REFERENCES slides(id) ON DELETE CASCADE
        );
    `)
	if err != nil {
		return err
	}

	// Insert the QuPath‑standard image types
	_, err = db.Exec(`
        INSERT OR IGNORE INTO image_types (id, type_uid, name, description, category, requires_histogram) VALUES
        ('img_type_bf_he',    'brightfield_he',    'Brightfield H&E',    'Hematoxylin and Eosin stained brightfield image', 'brightfield', FALSE),
        ('img_type_bf_hdab',  'brightfield_hdab',  'Brightfield H-DAB',  'Hematoxylin and DAB (IHC) brightfield image',     'brightfield', FALSE),
        ('img_type_bf_other', 'brightfield_other', 'Brightfield Other',  'Other brightfield methods',                      'brightfield', FALSE),
        ('img_type_fluor',    'fluorescence',      'Fluorescence',       'Fluorescence microscopy image',                  'fluorescence', TRUE),
        ('img_type_other',    'other',             'Other',              'Other imaging modalities',                       'other', FALSE),
        ('img_type_unspec',   'unspecified',       'Unspecified',        'Image type not specified',                       'other', FALSE);
    `)
	return err
}

// CreateTriggers sets up updated_at triggers
func (s *ImageTypesSchema) CreateTriggers(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TRIGGER IF NOT EXISTS image_types_updated_at 
        AFTER UPDATE ON image_types 
        FOR EACH ROW 
        BEGIN
            UPDATE image_types SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
        END;

        CREATE TRIGGER IF NOT EXISTS slide_histograms_updated_at 
        AFTER UPDATE ON slide_histograms 
        FOR EACH ROW 
        BEGIN
            UPDATE slide_histograms SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
        END;

        CREATE TRIGGER IF NOT EXISTS slide_staining_protocols_updated_at 
        AFTER UPDATE ON slide_staining_protocols 
        FOR EACH ROW 
        BEGIN
            UPDATE slide_staining_protocols SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
        END;
    `)
	return err
}

// CreateIndexes adds performance‑boosting indexes
func (s *ImageTypesSchema) CreateIndexes(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_image_types_type_uid     ON image_types(type_uid);
        CREATE INDEX IF NOT EXISTS idx_image_types_category     ON image_types(category);
        CREATE INDEX IF NOT EXISTS idx_image_types_tenant_id    ON image_types(tenant_id);
        CREATE INDEX IF NOT EXISTS idx_image_types_active       ON image_types(is_active);

        CREATE INDEX IF NOT EXISTS idx_slide_histograms_slide   ON slide_histograms(slide_id);
        CREATE INDEX IF NOT EXISTS idx_slide_histograms_channel ON slide_histograms(slide_id, channel_index);

        CREATE INDEX IF NOT EXISTS idx_slide_staining_slide     ON slide_staining_protocols(slide_id);
        CREATE INDEX IF NOT EXISTS idx_slide_staining_name      ON slide_staining_protocols(stain_name);
        CREATE INDEX IF NOT EXISTS idx_slide_staining_type      ON slide_staining_protocols(stain_type);
    `)
	return err
}
