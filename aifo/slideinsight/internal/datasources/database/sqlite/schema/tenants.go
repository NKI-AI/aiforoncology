// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package schema

import "database/sql"

// TenantsSchema handles tenant-related schema creation
type TenantsSchema struct{}

// CreateTables creates tenant-related tables
func (s *TenantsSchema) CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
		-- Tenants table
		CREATE TABLE IF NOT EXISTS tenants (
		id               INTEGER PRIMARY KEY AUTOINCREMENT,
		short_uid        TEXT    NOT NULL UNIQUE,
		name             TEXT    NOT NULL,
		description      TEXT    DEFAULT '',
		status           TEXT    NOT NULL
							DEFAULT 'active'
							CHECK(status IN ('active','inactive','suspended','pending')),
		deactivated_at   TIMESTAMP NULL,
		deactivated_by   INTEGER   NULL,
		created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- Tenant domains table for domain-based tenant identification
		CREATE TABLE IF NOT EXISTS tenant_domains (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id  INTEGER NOT NULL,
			domain     TEXT    NOT NULL UNIQUE,
			is_verified BOOLEAN DEFAULT FALSE,
			is_primary BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		return err
	}

	// Immediately create the system tenant with ID=0 to ensure it exists for foreign key constraints
	// This is required because RBAC tables will reference tenant_id=0 for system-level permissions and roles
	_, err = db.Exec(`
		INSERT OR IGNORE INTO tenants (id, short_uid, name, description, status)
		VALUES (0, 'system', 'System', 'System tenant for platform operations', 'active');
	`)
	return err
}

// CreateTriggers creates tenant-related triggers
func (s *TenantsSchema) CreateTriggers(db *sql.DB) error {
	_, err := db.Exec(`
		-- Trigger for tenants table
		CREATE TRIGGER IF NOT EXISTS tenants_updated_at 
		AFTER UPDATE ON tenants 
		FOR EACH ROW 
		BEGIN
			UPDATE tenants SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

		-- Trigger for tenant_domains table
		CREATE TRIGGER IF NOT EXISTS tenant_domains_updated_at 
		AFTER UPDATE ON tenant_domains 
		FOR EACH ROW 
		BEGIN
			UPDATE tenant_domains SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
	`)
	return err
}

// CreateIndexes creates tenant-related indexes
func (s *TenantsSchema) CreateIndexes(db *sql.DB) error {
	_, err := db.Exec(`
		-- Indexes for tenant domains (for domain-based tenant lookup)
		CREATE INDEX IF NOT EXISTS idx_tenant_domains_domain ON tenant_domains(domain);
		CREATE INDEX IF NOT EXISTS idx_tenant_domains_tenant_id ON tenant_domains(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_tenant_domains_verified ON tenant_domains(is_verified);
		CREATE INDEX IF NOT EXISTS idx_tenant_domains_primary ON tenant_domains(tenant_id, is_primary);
	`)
	return err
}
