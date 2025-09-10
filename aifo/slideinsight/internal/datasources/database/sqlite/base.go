// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite/algorithms"
	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite/annotations"
	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite/cases"
	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite/email_templates"
	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite/image_types"
	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite/notifications"
	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite/rbac"
	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite/regions"
	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite/schema"
	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite/settings"
	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite/slides"
	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite/studies"
	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite/tenants"
	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite/users"
	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// DB implements the Database interface using SQLite
type DB struct {
	db                *sql.DB
	algorithms        *algorithms.Adapter
	cases             *cases.Adapter
	annotations       *annotations.Adapter
	users             *users.Adapter
	studies           *studies.Adapter
	tenants           *tenants.Adapter
	slides            *slides.Adapter
	rbac              *rbac.Adapter
	regions           *regions.Adapter
	notifications     *notifications.Adapter
	emailTemplates    *email_templates.Adapter
	settings          *settings.Adapter
	imageTypes        *image_types.ImageTypesAdapter
	slideHistograms   *image_types.SlideHistogramsAdapter
	stainingProtocols *image_types.StainingProtocolsAdapter
}

// NewDB creates a new SQLite database instance
func NewDB(databaseURL string) (ports.Database, error) {
	// Strip the sqlite:// prefix if present
	dbPath := databaseURL
	if strings.HasPrefix(dbPath, "sqlite://") {
		dbPath = strings.TrimPrefix(dbPath, "sqlite://")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open SQLite database: %w", err)
	}

	// Initialize the database schema using the new modular approach
	if err := schema.InitializeSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize SQLite schema: %w", err)
	}

	sqliteDB := &DB{
		db:                db,
		algorithms:        algorithms.NewAdapter(db),
		cases:             cases.NewAdapter(db),
		annotations:       annotations.NewAdapter(db),
		users:             users.NewAdapter(db),
		studies:           studies.NewAdapter(db),
		tenants:           tenants.NewAdapter(db),
		slides:            slides.NewAdapter(db),
		rbac:              rbac.NewAdapter(db),
		regions:           regions.NewAdapter(db),
		notifications:     notifications.NewAdapter(db),
		emailTemplates:    email_templates.NewAdapter(db),
		settings:          settings.NewAdapter(db),
		imageTypes:        image_types.NewImageTypesAdapter(db),
		slideHistograms:   image_types.NewSlideHistogramsAdapter(db),
		stainingProtocols: image_types.NewStainingProtocolsAdapter(db),
	}

	return sqliteDB, nil
}

// CloseConnections closes all open connections to the database
func (db *DB) CloseConnections() {
	if db.db != nil {
		db.db.Close()
	}
}

// UserHasGlobalRole proxies the RBAC adapter call.
func (db *DB) UserHasGlobalRole(ctx context.Context, userID int, roleName string) (bool, error) {
	return db.rbac.UserHasGlobalRole(ctx, userID, roleName)
}

func (db *DB) HasObjectGrant(ctx context.Context, userID int, permission, resourceType string, resourceID int) (bool, error) {
	return db.rbac.HasObjectGrant(ctx, userID, permission, resourceType, resourceID)
}

// CreateRoleIfNotExists proxies the RBAC adapter call.
func (db *DB) CreateRoleIfNotExists(ctx context.Context, roleName, description string) (int, error) {
	return db.rbac.CreateRoleIfNotExists(ctx, roleName, description)
}

// CreateTenantRoleIfNotExists proxies the RBAC adapter call.
func (db *DB) CreateTenantRoleIfNotExists(ctx context.Context, tenantID int, roleName, description string) (int, error) {
	return db.rbac.CreateTenantRoleIfNotExists(ctx, tenantID, roleName, description)
}

// AssignGlobalRoleToUser proxies the RBAC adapter call.
func (db *DB) AssignGlobalRoleToUser(ctx context.Context, userID int, roleName string) error {
	return db.rbac.AssignGlobalRoleToUser(ctx, userID, roleName)
}

// CreatePermission proxies the RBAC adapter call.
func (db *DB) CreatePermission(ctx context.Context, name, description string) (int, error) {
	return db.rbac.CreatePermission(ctx, name, description)
}

// CreatePermissionIfNotExists proxies the RBAC adapter call.
func (db *DB) CreatePermissionIfNotExists(ctx context.Context, name, description string) (int, error) {
	return db.rbac.CreatePermissionIfNotExists(ctx, name, description)
}

// CreateTenantPermissionIfNotExists proxies the RBAC adapter call.
func (db *DB) CreateTenantPermissionIfNotExists(ctx context.Context, tenantID int, name, description string) (int, error) {
	return db.rbac.CreateTenantPermissionIfNotExists(ctx, tenantID, name, description)
}

// GetAllPermissions proxies the RBAC adapter call.
func (db *DB) GetAllPermissions(ctx context.Context) ([]ports.Permission, error) {
	return db.rbac.GetAllPermissions(ctx)
}

// GetPermissionsWithPagination proxies the RBAC adapter call.
func (db *DB) GetPermissionsWithPagination(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.Permission, error) {
	return db.rbac.GetPermissionsWithPagination(ctx, search, limit, offset)
}

// GetPermissionCount proxies the RBAC adapter call.
func (db *DB) GetPermissionCount(ctx context.Context, search utils.SearchParams) (int, error) {
	return db.rbac.GetPermissionCount(ctx, search)
}

// GetPermissionByName proxies the RBAC adapter call.
func (db *DB) GetPermissionByName(ctx context.Context, name string) (*ports.Permission, error) {
	return db.rbac.GetPermissionByName(ctx, name)
}

// GetTenantPermissionByName proxies the RBAC adapter call.
func (db *DB) GetTenantPermissionByName(ctx context.Context, tenantID int, name string) (*ports.Permission, error) {
	return db.rbac.GetTenantPermissionByName(ctx, tenantID, name)
}

// DeletePermission proxies the RBAC adapter call.
func (db *DB) DeletePermission(ctx context.Context, name string) error {
	return db.rbac.DeletePermission(ctx, name)
}

// DeleteTenantPermission proxies the RBAC adapter call.
func (db *DB) DeleteTenantPermission(ctx context.Context, tenantID int, name string) error {
	return db.rbac.DeleteTenantPermission(ctx, tenantID, name)
}

// CreateObjectGrant proxies the RBAC adapter call.
func (db *DB) CreateObjectGrant(ctx context.Context, granteeType string, granteeID int, permission, resourceType string, resourceID int) error {
	return db.rbac.CreateObjectGrant(ctx, granteeType, granteeID, permission, resourceType, resourceID)
}

// GetObjectGrants proxies the RBAC adapter call.
func (db *DB) GetObjectGrants(ctx context.Context, resourceType string, resourceID int) ([]ports.ObjectGrant, error) {
	return db.rbac.GetObjectGrants(ctx, resourceType, resourceID)
}

// DeleteObjectGrant proxies the RBAC adapter call.
func (db *DB) DeleteObjectGrant(ctx context.Context, granteeType string, granteeID int, permission, resourceType string, resourceID int) error {
	return db.rbac.DeleteObjectGrant(ctx, granteeType, granteeID, permission, resourceType, resourceID)
}

// SQL-based permission filtering methods for efficient bulk operations

// GetAccessibleStudyIDs proxies the RBAC adapter call.
func (db *DB) GetAccessibleStudyIDs(ctx context.Context, userID int, permission string) ([]int, error) {
	return db.rbac.GetAccessibleStudyIDs(ctx, userID, permission)
}

// GetAccessibleCaseIDs proxies the RBAC adapter call.
func (db *DB) GetAccessibleCaseIDs(ctx context.Context, userID int, permission string) ([]int, error) {
	return db.rbac.GetAccessibleCaseIDs(ctx, userID, permission)
}

// GetAccessibleSlideIDs proxies the RBAC adapter call.
func (db *DB) GetAccessibleSlideIDs(ctx context.Context, userID int, permission string) ([]int, error) {
	return db.rbac.GetAccessibleSlideIDs(ctx, userID, permission)
}

// Role methods - proxy to RBAC adapter

// CreateRole proxies the RBAC adapter call.
func (db *DB) CreateRole(ctx context.Context, name, description string) (int, error) {
	return db.rbac.CreateRole(ctx, name, description)
}

// GetAllRoles proxies the RBAC adapter call.
func (db *DB) GetAllRoles(ctx context.Context) ([]ports.Role, error) {
	return db.rbac.GetAllRoles(ctx)
}

// GetRolesWithPagination proxies the RBAC adapter call.
func (db *DB) GetRolesWithPagination(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.Role, error) {
	return db.rbac.GetRolesWithPagination(ctx, search, limit, offset)
}

// GetRoleCount proxies the RBAC adapter call.
func (db *DB) GetRoleCount(ctx context.Context, search utils.SearchParams) (int, error) {
	return db.rbac.GetRoleCount(ctx, search)
}

// GetRoleByName proxies the RBAC adapter call.
func (db *DB) GetRoleByName(ctx context.Context, name string) (*ports.Role, error) {
	return db.rbac.GetRoleByName(ctx, name)
}

// GetRoleByID proxies the RBAC adapter call.
func (db *DB) GetRoleByID(ctx context.Context, roleID int) (*ports.Role, error) {
	return db.rbac.GetRoleByID(ctx, roleID)
}

// DeleteRole proxies the RBAC adapter call.
func (db *DB) DeleteRole(ctx context.Context, name string) error {
	return db.rbac.DeleteRole(ctx, name)
}

// Role-Permission methods - proxy to RBAC adapter

// AssignPermissionToRole proxies the RBAC adapter call.
func (db *DB) AssignPermissionToRole(ctx context.Context, roleID int, permissionID int) error {
	return db.rbac.AssignPermissionToRole(ctx, roleID, permissionID)
}

// RemovePermissionFromRole proxies the RBAC adapter call.
func (db *DB) RemovePermissionFromRole(ctx context.Context, roleID int, permissionID int) error {
	return db.rbac.RemovePermissionFromRole(ctx, roleID, permissionID)
}

// GetRolePermissions proxies the RBAC adapter call.
func (db *DB) GetRolePermissions(ctx context.Context, roleID int) ([]ports.Permission, error) {
	return db.rbac.GetRolePermissions(ctx, roleID)
}

// User-Role methods - proxy to RBAC adapter

// AssignRoleToUser proxies the RBAC adapter call.
func (db *DB) AssignRoleToUser(ctx context.Context, userID int, roleID int, tenantID *int) error {
	return db.rbac.AssignRoleToUser(ctx, userID, roleID, tenantID)
}

// RemoveRoleFromUser proxies the RBAC adapter call.
func (db *DB) RemoveRoleFromUser(ctx context.Context, userID int, roleID int, tenantID *int) error {
	return db.rbac.RemoveRoleFromUser(ctx, userID, roleID, tenantID)
}

// GetUserRoles proxies the RBAC adapter call.
func (db *DB) GetUserRoles(ctx context.Context, userID int) ([]ports.UserRole, error) {
	return db.rbac.GetUserRoles(ctx, userID)
}

// GetRoleUsers proxies the RBAC adapter call.
func (db *DB) GetRoleUsers(ctx context.Context, roleID int) ([]ports.UserRole, error) {
	return db.rbac.GetRoleUsers(ctx, roleID)
}

// Group methods - proxy to RBAC adapter

// CreateGroup proxies the RBAC adapter call.
func (db *DB) CreateGroup(ctx context.Context, name, description string) (int, error) {
	return db.rbac.CreateGroup(ctx, name, description)
}

// CreateTenantGroup proxies the RBAC adapter call.
func (db *DB) CreateTenantGroup(ctx context.Context, tenantID int, name, description string) (int, error) {
	return db.rbac.CreateTenantGroup(ctx, tenantID, name, description)
}

// GetAllGroups proxies the RBAC adapter call.
func (db *DB) GetAllGroups(ctx context.Context) ([]ports.Group, error) {
	return db.rbac.GetAllGroups(ctx)
}

// GetGroupsWithPagination proxies the RBAC adapter call.
func (db *DB) GetGroupsWithPagination(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.Group, error) {
	return db.rbac.GetGroupsWithPagination(ctx, search, limit, offset)
}

// GetGroupCount proxies the RBAC adapter call.
func (db *DB) GetGroupCount(ctx context.Context, search utils.SearchParams) (int, error) {
	return db.rbac.GetGroupCount(ctx, search)
}

// GetGroupByName proxies the RBAC adapter call.
func (db *DB) GetGroupByName(ctx context.Context, name string) (*ports.Group, error) {
	return db.rbac.GetGroupByName(ctx, name)
}

// GetGroupByID proxies the RBAC adapter call.
func (db *DB) GetGroupByID(ctx context.Context, groupID int) (*ports.Group, error) {
	return db.rbac.GetGroupByID(ctx, groupID)
}

// DeleteGroup proxies the RBAC adapter call.
func (db *DB) DeleteGroup(ctx context.Context, name string) error {
	return db.rbac.DeleteGroup(ctx, name)
}

// User-Group methods - proxy to RBAC adapter

// AssignUserToGroup proxies the RBAC adapter call.
func (db *DB) AssignUserToGroup(ctx context.Context, userID int, groupID int) error {
	return db.rbac.AssignUserToGroup(ctx, userID, groupID)
}

// RemoveUserFromGroup proxies the RBAC adapter call.
func (db *DB) RemoveUserFromGroup(ctx context.Context, userID int, groupID int) error {
	return db.rbac.RemoveUserFromGroup(ctx, userID, groupID)
}

// GetUserGroups proxies the RBAC adapter call.
func (db *DB) GetUserGroups(ctx context.Context, userID int) ([]ports.Group, error) {
	return db.rbac.GetUserGroups(ctx, userID)
}

// GetGroupUsers proxies the RBAC adapter call.
func (db *DB) GetGroupUsers(ctx context.Context, groupID int) ([]int, error) {
	return db.rbac.GetGroupUsers(ctx, groupID)
}

// EmailTemplateRepository methods - proxy to email templates adapter

// GetEmailTemplates retrieves email templates with pagination and filtering
func (db *DB) GetEmailTemplates(ctx context.Context, tenantID int, limit, offset int) ([]ports.EmailTemplate, int, error) {
	return db.emailTemplates.GetEmailTemplates(ctx, tenantID, limit, offset)
}

// GetAllEmailTemplates retrieves email templates from all tenants (for superadmin access)
func (db *DB) GetAllEmailTemplates(ctx context.Context, limit, offset int) ([]ports.EmailTemplate, int, error) {
	return db.emailTemplates.GetAllEmailTemplates(ctx, limit, offset)
}

// GetEmailTemplateByID retrieves a specific email template by ID
func (db *DB) GetEmailTemplateByID(ctx context.Context, id int) (*ports.EmailTemplate, error) {
	return db.emailTemplates.GetEmailTemplateByID(ctx, id)
}

// GetEmailTemplateByType retrieves a template by tenant and type
func (db *DB) GetEmailTemplateByType(ctx context.Context, tenantID int, templateType ports.EmailTemplateType) (*ports.EmailTemplate, error) {
	return db.emailTemplates.GetEmailTemplateByType(ctx, tenantID, templateType)
}

// CreateEmailTemplate creates a new email template
func (db *DB) CreateEmailTemplate(ctx context.Context, template ports.NewEmailTemplate) (*ports.EmailTemplate, error) {
	return db.emailTemplates.CreateEmailTemplate(ctx, template)
}

// UpdateEmailTemplate updates an existing email template
func (db *DB) UpdateEmailTemplate(ctx context.Context, id int, updates ports.UpdateEmailTemplate) (*ports.EmailTemplate, error) {
	return db.emailTemplates.UpdateEmailTemplate(ctx, id, updates)
}

// DeleteEmailTemplate deletes an email template (soft delete)
func (db *DB) DeleteEmailTemplate(ctx context.Context, id int, allowSystemDeletion bool) error {
	return db.emailTemplates.DeleteEmailTemplate(ctx, id, allowSystemDeletion)
}

// CreateDefaultTemplates creates default system templates for a tenant
func (db *DB) CreateDefaultTemplates(ctx context.Context, tenantID int, createdBy int) error {
	return db.emailTemplates.CreateDefaultTemplates(ctx, tenantID, createdBy)
}

// Permission-related helper methods for middleware

// GetCaseIDByUID gets the internal case ID from a case UID
func (db *DB) GetCaseIDByUID(ctx context.Context, caseUID string) (int, error) {
	var id int
	err := db.db.QueryRowContext(ctx, "SELECT id FROM cases WHERE short_uid = ? AND deleted_at IS NULL", caseUID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("case with UID '%s' not found", caseUID)
		}
		return 0, fmt.Errorf("failed to get case ID: %w", err)
	}
	return id, nil
}

// GetStudyIDByUID gets the internal study ID from a study UID (proxies to studies adapter)
func (db *DB) GetStudyIDByUID(ctx context.Context, studyUID string) (int, error) {
	return db.studies.GetStudyIDByShortUID(ctx, studyUID)
}

// GetSlideIDByUID gets the internal slide ID from a slide UID
func (db *DB) GetSlideIDByUID(ctx context.Context, slideUID string) (int, error) {
	var id int
	err := db.db.QueryRowContext(ctx, "SELECT id FROM slides WHERE slide_uid = ? AND deleted_at IS NULL", slideUID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("slide with UID '%s' not found", slideUID)
		}
		return 0, fmt.Errorf("failed to get slide ID: %w", err)
	}
	return id, nil
}

// GetCaseStudyRelations gets all study IDs that contain the given case
func (db *DB) GetCaseStudyRelations(ctx context.Context, caseID int) ([]int, error) {
	rows, err := db.db.QueryContext(ctx, "SELECT study_id FROM study_cases WHERE case_id = ?", caseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query case study relations: %w", err)
	}
	defer rows.Close()

	var studyIDs []int
	for rows.Next() {
		var studyID int
		if err := rows.Scan(&studyID); err != nil {
			return nil, fmt.Errorf("failed to scan study ID: %w", err)
		}
		studyIDs = append(studyIDs, studyID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over study relations: %w", err)
	}

	return studyIDs, nil
}

// GetSlideCaseRelation gets the case ID that contains the given slide
func (db *DB) GetSlideCaseRelation(ctx context.Context, slideID int) (int, error) {
	var caseID int
	err := db.db.QueryRowContext(ctx, "SELECT case_id FROM slides WHERE id = ? AND deleted_at IS NULL", slideID).Scan(&caseID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("slide with ID %d not found", slideID)
		}
		return 0, fmt.Errorf("failed to get slide case relation: %w", err)
	}
	return caseID, nil
}

// GetUserIDByUID gets the internal user ID from a user UID (for permission middleware)
func (db *DB) GetUserIDByUID(ctx context.Context, userUID string) (int, error) {
	var id int
	err := db.db.QueryRowContext(ctx, "SELECT id FROM users WHERE short_uid = ?", userUID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("user with UID '%s' not found", userUID)
		}
		return 0, fmt.Errorf("failed to get user ID: %w", err)
	}
	return id, nil
}

// UserHasRolePermission proxies the RBAC adapter call.
func (db *DB) UserHasRolePermission(ctx context.Context, userID int, permissionName string) (bool, error) {
	return db.rbac.UserHasRolePermission(ctx, userID, permissionName)
}

// Regions adapter proxy methods
func (db *DB) LoadAllRegions(ctx context.Context) ([]ports.Region, error) {
	return db.regions.LoadAllRegions(ctx)
}

func (db *DB) GetRegionsGeneric(ctx context.Context, params utils.PaginationAndSearchParams, searchParams ports.RegionSearchParams) ([]ports.Region, domain.PaginationInfo, error) {
	return db.regions.GetRegionsGeneric(ctx, params, searchParams)
}

func (db *DB) GetRegionsBySlideUID(ctx context.Context, slideUID string) ([]ports.Region, error) {
	return db.regions.GetRegionsBySlideUID(ctx, slideUID)
}

func (db *DB) CreateRegion(ctx context.Context, newRegion ports.NewRegion) error {
	return db.regions.CreateRegion(ctx, newRegion)
}

func (db *DB) GetRegionByID(ctx context.Context, regionID string) (ports.Region, error) {
	return db.regions.GetRegionByID(ctx, regionID)
}

func (db *DB) UpdateRegion(ctx context.Context, regionID string, updates ports.UpdateRegion) error {
	return db.regions.UpdateRegion(ctx, regionID, updates)
}

func (db *DB) SoftDeleteRegion(ctx context.Context, regionID string, deletedBy int) error {
	return db.regions.SoftDeleteRegion(ctx, regionID, deletedBy)
}

func (db *DB) GetDeletedRegions(ctx context.Context) ([]ports.Region, error) {
	return db.regions.GetDeletedRegions(ctx)
}

func (db *DB) RestoreRegion(ctx context.Context, regionID string) error {
	return db.regions.RestoreRegion(ctx, regionID)
}

func (db *DB) BulkCreateRegions(ctx context.Context, newRegions []ports.NewRegion) error {
	return db.regions.BulkCreateRegions(ctx, newRegions)
}

func (db *DB) BulkUpdateRegions(ctx context.Context, updates map[string]ports.UpdateRegion) error {
	return db.regions.BulkUpdateRegions(ctx, updates)
}

func (db *DB) BulkDeleteRegions(ctx context.Context, regionIDs []string, deletedBy int) error {
	return db.regions.BulkDeleteRegions(ctx, regionIDs, deletedBy)
}
