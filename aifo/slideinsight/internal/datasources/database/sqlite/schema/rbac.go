// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package schema

import "database/sql"

// RBACSchema handles role-based access control schema creation
type RBACSchema struct{}

// CreateTables creates RBAC-related tables
func (s *RBACSchema) CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
		-- Roles table - Pure tenant isolation (Google Cloud IAM style)
		CREATE TABLE IF NOT EXISTS roles (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		tenant_id    INTEGER NOT NULL,                  -- 0 for system tenant, >0 for regular tenants
		short_uid    TEXT    NOT NULL UNIQUE,
		name         TEXT    NOT NULL,                 -- e.g. 'admin', 'researcher', 'viewer', 'superadmin'
		description  TEXT,
		created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		UNIQUE(tenant_id, name)                        -- Allow same role name per tenant, but unique within tenant
		);

		-- Permissions table - Pure tenant isolation
		CREATE TABLE IF NOT EXISTS permissions (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		tenant_id    INTEGER NOT NULL,                  -- 0 for system tenant, >0 for regular tenants  
		short_uid    TEXT    NOT NULL UNIQUE,
		name         TEXT    NOT NULL,                 -- e.g. 'studies.view', 'cases.edit', 'platform.admin'
		description  TEXT,
		created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		UNIQUE(tenant_id, name)                        -- Allow same permission name per tenant, but unique within tenant
		);

		-- Groups table - Pure tenant isolation
		CREATE TABLE IF NOT EXISTS groups (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		tenant_id    INTEGER NOT NULL,                  -- 0 for system tenant, >0 for regular tenants
		short_uid    TEXT    NOT NULL UNIQUE,
		name         TEXT    NOT NULL,                 -- e.g. 'administrators', 'researchers', 'viewers'
		description  TEXT,
		created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		UNIQUE(tenant_id, name)                        -- Allow same group name per tenant, but unique within tenant
		);

		-- 1) Link roles ↔ permissions
		CREATE TABLE IF NOT EXISTS role_permissions (
		role_id       INTEGER NOT NULL,
		permission_id INTEGER NOT NULL,
		created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (role_id, permission_id),
		FOREIGN KEY (role_id)       REFERENCES roles(id)       ON DELETE CASCADE,
		FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
		);

		-- 2) Assign roles to users (always with tenant context)
		CREATE TABLE IF NOT EXISTS user_roles (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id     INTEGER NOT NULL,
		role_id     INTEGER NOT NULL,
		tenant_id   INTEGER NOT NULL,                   -- Matches the role's tenant_id for consistency
		created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id)   REFERENCES users(id)   ON DELETE CASCADE,
		FOREIGN KEY (role_id)   REFERENCES roles(id)   ON DELETE CASCADE,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
		);

		-- 3) (Optional) Group users together
		CREATE TABLE IF NOT EXISTS user_groups (
		user_id    INTEGER NOT NULL,
		group_id   INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, group_id),
		FOREIGN KEY (user_id)  REFERENCES users(id)  ON DELETE CASCADE,
		FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
		);

		-- 4) Per-object grants (polymorphic ACL)
		CREATE TABLE IF NOT EXISTS object_grants (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		grantee_type   TEXT    NOT NULL
						CHECK(grantee_type IN ('user','group','role')),
		grantee_id     INTEGER NOT NULL,
		permission     TEXT    NOT NULL,
		resource_type  TEXT    NOT NULL
						CHECK(resource_type IN ('study','case','slide')),
		resource_id    INTEGER NOT NULL,
		created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		-- enforce the FK to users/groups/roles in app logic or via triggers if you like
		);
	`)
	return err
}

// CreateTriggers creates RBAC-related triggers
func (s *RBACSchema) CreateTriggers(db *sql.DB) error {
	_, err := db.Exec(`
		-- Trigger to update updated_at timestamps for roles
		CREATE TRIGGER IF NOT EXISTS roles_updated_at 
		AFTER UPDATE ON roles 
		FOR EACH ROW 
		BEGIN
			UPDATE roles SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

		-- Trigger to update updated_at timestamps for permissions
		CREATE TRIGGER IF NOT EXISTS permissions_updated_at 
		AFTER UPDATE ON permissions 
		FOR EACH ROW 
		BEGIN
			UPDATE permissions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

		-- Trigger to update updated_at timestamps for groups
		CREATE TRIGGER IF NOT EXISTS groups_updated_at 
		AFTER UPDATE ON groups 
		FOR EACH ROW 
		BEGIN
			UPDATE groups SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

		-- role_permissions.updated_at
		CREATE TRIGGER IF NOT EXISTS role_permissions_updated_at
		AFTER UPDATE ON role_permissions
		FOR EACH ROW
		BEGIN
		UPDATE role_permissions
			SET updated_at = CURRENT_TIMESTAMP
		WHERE role_id = NEW.role_id
			AND permission_id = NEW.permission_id;
		END;

		-- user_roles.updated_at
		CREATE TRIGGER IF NOT EXISTS user_roles_updated_at
		AFTER UPDATE ON user_roles
		FOR EACH ROW
		BEGIN
		UPDATE user_roles
			SET updated_at = CURRENT_TIMESTAMP
		WHERE id = NEW.id;
		END;

		-- user_groups.updated_at
		CREATE TRIGGER IF NOT EXISTS user_groups_updated_at
		AFTER UPDATE ON user_groups
		FOR EACH ROW
		BEGIN
		UPDATE user_groups
			SET updated_at = CURRENT_TIMESTAMP
		WHERE user_id  = NEW.user_id
			AND group_id = NEW.group_id;
		END;

		-- object_grants.updated_at
		CREATE TRIGGER IF NOT EXISTS object_grants_updated_at
		AFTER UPDATE ON object_grants
		FOR EACH ROW
		BEGIN
		UPDATE object_grants
			SET updated_at = CURRENT_TIMESTAMP
		WHERE id = NEW.id;
		END;
	`)
	return err
}

// CreateIndexes creates RBAC-related indexes
func (s *RBACSchema) CreateIndexes(db *sql.DB) error {
	_, err := db.Exec(`
		-- Role↔Permission lookups
		CREATE INDEX IF NOT EXISTS idx_role_permissions_role_perm
		ON role_permissions(role_id, permission_id);

		-- User↔Role lookups (and scoped by tenant)
		CREATE INDEX IF NOT EXISTS idx_user_roles_user_tenant ON user_roles(user_id, tenant_id);
		CREATE INDEX IF NOT EXISTS idx_user_roles_role_tenant ON user_roles(role_id, tenant_id);

		-- User↔Group lookups
		CREATE INDEX IF NOT EXISTS idx_user_groups_group_user ON user_groups(group_id, user_id);

		-- Object grants by permission+resource, for fast ACL checks
		CREATE INDEX IF NOT EXISTS idx_objgrants_perm_res ON object_grants(permission, resource_type, resource_id);

		CREATE INDEX IF NOT EXISTS idx_objgrants_grantee ON object_grants(grantee_type, grantee_id);
		
		CREATE INDEX IF NOT EXISTS idx_objgrants_resource ON object_grants(resource_type, resource_id);

		-- Permission lookups with tenant scoping
		CREATE INDEX IF NOT EXISTS idx_permissions_tenant_name ON permissions(tenant_id, name);
		CREATE INDEX IF NOT EXISTS idx_permissions_tenant_id ON permissions(tenant_id);
	`)
	return err
}
