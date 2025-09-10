// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package schema

import "database/sql"

// AlgorithmsSchema handles algorithm-related schema creation
type AlgorithmsSchema struct{}

// CreateTables creates algorithm-related tables
func (s *AlgorithmsSchema) CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
		-- 1. Algorithms registry
		CREATE TABLE IF NOT EXISTS algorithms (
			id                    TEXT      PRIMARY KEY,                         -- system-generated UUID
			tenant_id             INT      NOT NULL,                            -- FK into tenants
			name                  TEXT      NOT NULL,                            -- human-readable
			description           TEXT,
			version               TEXT      NOT NULL,                            -- SemVer
			endpoint_url          TEXT      NOT NULL,
			http_method           TEXT      NOT NULL DEFAULT 'POST',
			execution_mode        TEXT      NOT NULL CHECK(execution_mode IN ('BATCH','STREAM')),
			progress_transport    TEXT      NOT NULL CHECK(progress_transport IN ('WEBSOCKET','SSE'))
										DEFAULT 'WEBSOCKET',               -- per-algorithm setting
			metadata              JSON,                                         -- SQLite JSON1
			created_at            DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at            DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(tenant_id)  REFERENCES tenants(id) ON DELETE CASCADE
		);

		-- 2. Result sinks (callbacks, storage, etc.)
		CREATE TABLE IF NOT EXISTS result_sinks (
			id           TEXT      PRIMARY KEY,
			algorithm_id TEXT      NOT NULL REFERENCES algorithms(id) ON DELETE CASCADE,
			type         TEXT      NOT NULL,        -- e.g. "CALLBACK", "STORAGE"
			config       JSON      NOT NULL,
			created_at   DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- 3. Post-actions/hooks
		CREATE TABLE IF NOT EXISTS post_hooks (
			id           TEXT      PRIMARY KEY,
			algorithm_id TEXT      NOT NULL REFERENCES algorithms(id) ON DELETE CASCADE,
			type         TEXT      NOT NULL,        -- e.g. "VIEWER_OVERLAY"
			config       JSON      NOT NULL,
			created_at   DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- 4. Runs / operations
		CREATE TABLE IF NOT EXISTS algorithm_runs (
			id              TEXT      PRIMARY KEY,                       -- run UID (system-generated)
			algorithm_id    TEXT      NOT NULL REFERENCES algorithms(id),
			case_id         TEXT,                                       -- optional case-level run
			image_ids       JSON,                                       -- ["img1","img2"] if case or image-level
			regions         JSON,                                       -- per-image ROIs: { "img1":[{x,y,w,h},…], … }
			parameters      JSON,
			execution_mode  TEXT      NOT NULL CHECK(execution_mode IN ('BATCH','STREAM')),
			status          TEXT      NOT NULL CHECK(status IN ('QUEUED','RUNNING','SUCCEEDED','FAILED')),
			progress        INTEGER   NOT NULL DEFAULT 0,               -- 0–100
			result_uri      TEXT,
			error_info      JSON,
			created_at      DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at      DATETIME,
			finished_at     DATETIME,
			updated_at      DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- 5. Outputs catalog (artifacts)
		CREATE TABLE IF NOT EXISTS outputs (
			id            TEXT      PRIMARY KEY,
			run_id        TEXT      NOT NULL REFERENCES algorithm_runs(id) ON DELETE CASCADE,
			name          TEXT      NOT NULL,        -- e.g. "tissue_mask"
			uri           TEXT      NOT NULL,
			metadata      JSON,
			created_at    DATETIME  NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

// CreateTriggers creates algorithm-related triggers
func (s *AlgorithmsSchema) CreateTriggers(db *sql.DB) error {
	_, err := db.Exec(`
		-- Trigger for algorithms table
		CREATE TRIGGER IF NOT EXISTS algorithms_updated_at 
		AFTER UPDATE ON algorithms 
		FOR EACH ROW 
		BEGIN
			UPDATE algorithms SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

		-- Trigger for algorithm_runs table
		CREATE TRIGGER IF NOT EXISTS algorithm_runs_updated_at 
		AFTER UPDATE ON algorithm_runs 
		FOR EACH ROW 
		BEGIN
			UPDATE algorithm_runs SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
	`)
	return err
}

// CreateIndexes creates algorithm-related indexes
func (s *AlgorithmsSchema) CreateIndexes(db *sql.DB) error {
	_, err := db.Exec(`
		-- Indexes for algorithms table
		CREATE INDEX IF NOT EXISTS idx_algorithms_tenant_id ON algorithms(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_algorithms_name ON algorithms(name);
		CREATE INDEX IF NOT EXISTS idx_algorithms_version ON algorithms(version);

		-- Indexes for result_sinks table
		CREATE INDEX IF NOT EXISTS idx_result_sinks_algorithm_id ON result_sinks(algorithm_id);
		CREATE INDEX IF NOT EXISTS idx_result_sinks_type ON result_sinks(type);

		-- Indexes for post_hooks table
		CREATE INDEX IF NOT EXISTS idx_post_hooks_algorithm_id ON post_hooks(algorithm_id);
		CREATE INDEX IF NOT EXISTS idx_post_hooks_type ON post_hooks(type);

		-- Indexes for algorithm_runs table
		CREATE INDEX IF NOT EXISTS idx_algorithm_runs_algorithm_id ON algorithm_runs(algorithm_id);
		CREATE INDEX IF NOT EXISTS idx_algorithm_runs_case_id ON algorithm_runs(case_id);
		CREATE INDEX IF NOT EXISTS idx_algorithm_runs_status ON algorithm_runs(status);
		CREATE INDEX IF NOT EXISTS idx_algorithm_runs_created_at ON algorithm_runs(created_at);

		-- Indexes for outputs table
		CREATE INDEX IF NOT EXISTS idx_outputs_run_id ON outputs(run_id);
		CREATE INDEX IF NOT EXISTS idx_outputs_name ON outputs(name);
	`)
	return err
}
