package rbac

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Constants for tenant isolation (Google Cloud IAM style)
const (
	SystemTenantID = 0 // System tenant for platform operations
)

// Adapter provides RBAC related operations.
type Adapter struct {
	db *sql.DB
}

// NewAdapter creates a new rbac adapter.
func NewAdapter(db *sql.DB) *Adapter {
	return &Adapter{db: db}
}

// UserHasSystemRole checks if the user has the given role in the system tenant (tenant_id=0).
// This replaces the old UserHasGlobalRole function.
func (a *Adapter) UserHasSystemRole(ctx context.Context, userID int, roleName string) (bool, error) {
	query := `
		SELECT 1
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = ?
		  AND r.name = ?
		  AND r.tenant_id = ?
		LIMIT 1`

	row := a.db.QueryRowContext(ctx, query, userID, roleName, SystemTenantID)

	var tmp int
	err := row.Scan(&tmp)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// UserHasGlobalRole is kept for backward compatibility but now delegates to UserHasSystemRole
func (a *Adapter) UserHasGlobalRole(ctx context.Context, userID int, roleName string) (bool, error) {
	return a.UserHasSystemRole(ctx, userID, roleName)
}

// UserHasRolePermission checks if the user has the specified permission through any of their assigned roles.
// This method checks across all tenants the user has roles in.
func (a *Adapter) UserHasRolePermission(ctx context.Context, userID int, permissionName string) (bool, error) {
	query := `
		SELECT 1
		FROM user_roles ur
		JOIN role_permissions rp ON ur.role_id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE ur.user_id = ? AND p.name = ?
		LIMIT 1`

	row := a.db.QueryRowContext(ctx, query, userID, permissionName)

	var tmp int
	err := row.Scan(&tmp)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateRoleIfNotExists creates a role in the specified tenant if it doesn't already exist.
// Returns the role ID.
func (a *Adapter) CreateRoleIfNotExists(ctx context.Context, roleName, description string) (int, error) {
	// For backward compatibility, create in system tenant
	return a.CreateTenantRoleIfNotExists(ctx, SystemTenantID, roleName, description)
}

// CreateTenantRoleIfNotExists creates a tenant-specific role if it doesn't already exist.
// Returns the role ID.
func (a *Adapter) CreateTenantRoleIfNotExists(ctx context.Context, tenantID int, roleName, description string) (int, error) {
	// First check if role exists
	var roleID int
	checkQuery := `SELECT id FROM roles WHERE name = ? AND tenant_id = ?`
	err := a.db.QueryRowContext(ctx, checkQuery, roleName, tenantID).Scan(&roleID)
	if err == nil {
		// Role already exists
		return roleID, nil
	}
	if err != sql.ErrNoRows {
		// Real error
		return 0, err
	}

	// Generate short UID
	shortUID, err := utils.GenerateFixedShortUID()
	if err != nil {
		return 0, err
	}

	// Role doesn't exist, create it
	insertQuery := `
		INSERT INTO roles (tenant_id, short_uid, name, description)
		VALUES (?, ?, ?, ?)
		RETURNING id`

	err = a.db.QueryRowContext(ctx, insertQuery, tenantID, shortUID, roleName, description).Scan(&roleID)
	if err != nil {
		return 0, err
	}

	return roleID, nil
}

// AssignGlobalRoleToUser assigns a system role to a user.
func (a *Adapter) AssignGlobalRoleToUser(ctx context.Context, userID int, roleName string) error {
	return a.AssignTenantRoleToUser(ctx, userID, SystemTenantID, roleName)
}

// AssignTenantRoleToUser assigns a role from the specified tenant to a user.
func (a *Adapter) AssignTenantRoleToUser(ctx context.Context, userID int, tenantID int, roleName string) error {
	// TENANT ISOLATION SECURITY CHECK
	// Users can only be assigned roles from:
	// 1. Their home tenant
	// 2. System tenant (ID=0) for platform operations
	// This prevents cross-tenant privilege escalation
	if tenantID != SystemTenantID { // If not system tenant, validate home tenant
		// Get the user's home tenant
		var userTenantID int
		userQuery := `SELECT tenant_id FROM users WHERE id = ?`
		err := a.db.QueryRowContext(ctx, userQuery, userID).Scan(&userTenantID)
		if err != nil {
			return fmt.Errorf("failed to get user's home tenant: %w", err)
		}

		// Reject if trying to assign role from different tenant
		if tenantID != userTenantID {
			return fmt.Errorf("tenant isolation violation: user (tenant_id=%d) cannot be assigned role from tenant_id=%d", userTenantID, tenantID)
		}
	}

	// First get the role ID
	var roleID int
	roleQuery := `SELECT id FROM roles WHERE name = ? AND tenant_id = ?`
	err := a.db.QueryRowContext(ctx, roleQuery, roleName, tenantID).Scan(&roleID)
	if err != nil {
		return err
	}

	// Check if assignment already exists
	checkQuery := `
		SELECT 1 FROM user_roles 
		WHERE user_id = ? AND role_id = ? AND tenant_id = ?
		LIMIT 1`

	var tmp int
	err = a.db.QueryRowContext(ctx, checkQuery, userID, roleID, tenantID).Scan(&tmp)
	if err == nil {
		// Assignment already exists
		return nil
	}
	if err != sql.ErrNoRows {
		// Real error
		return err
	}

	// Create the assignment
	assignQuery := `
		INSERT INTO user_roles (user_id, role_id, tenant_id)
		VALUES (?, ?, ?)`

	_, err = a.db.ExecContext(ctx, assignQuery, userID, roleID, tenantID)
	return err
}

// CreatePermission creates a new permission in the system tenant.
// Returns the permission ID.
func (a *Adapter) CreatePermission(ctx context.Context, name, description string) (int, error) {
	return a.CreateTenantPermissionIfNotExists(ctx, SystemTenantID, name, description)
}

// CreatePermissionIfNotExists creates a permission in the system tenant if it doesn't already exist.
// Returns the permission ID.
func (a *Adapter) CreatePermissionIfNotExists(ctx context.Context, name, description string) (int, error) {
	return a.CreateTenantPermissionIfNotExists(ctx, SystemTenantID, name, description)
}

// CreateTenantPermissionIfNotExists creates a tenant-specific permission if it doesn't already exist.
// Returns the permission ID.
func (a *Adapter) CreateTenantPermissionIfNotExists(ctx context.Context, tenantID int, name, description string) (int, error) {
	// First check if permission exists
	var permissionID int
	checkQuery := `SELECT id FROM permissions WHERE name = ? AND tenant_id = ?`
	err := a.db.QueryRowContext(ctx, checkQuery, name, tenantID).Scan(&permissionID)
	if err == nil {
		// Permission already exists
		return permissionID, nil
	}
	if err != sql.ErrNoRows {
		// Real error
		return 0, err
	}

	// Generate short UID
	shortUID, err := utils.GenerateFixedShortUID()
	if err != nil {
		return 0, err
	}

	// Permission doesn't exist, create it
	insertQuery := `
		INSERT INTO permissions (tenant_id, short_uid, name, description)
		VALUES (?, ?, ?, ?)
		RETURNING id`

	err = a.db.QueryRowContext(ctx, insertQuery, tenantID, shortUID, name, description).Scan(&permissionID)
	if err != nil {
		return 0, err
	}

	return permissionID, nil
}

// GetAllPermissions returns all permissions from all tenants (system + regular).
func (a *Adapter) GetAllPermissions(ctx context.Context) ([]ports.Permission, error) {
	query := `
		SELECT id, tenant_id, short_uid, name, COALESCE(description, '') as description, created_at, updated_at
		FROM permissions
		ORDER BY tenant_id, name`

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []ports.Permission
	for rows.Next() {
		var p ports.Permission
		err := rows.Scan(&p.ID, &p.TenantID, &p.ShortUID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// GetPermissionsWithPagination returns permissions with pagination support
func (a *Adapter) GetPermissionsWithPagination(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.Permission, error) {
	baseQuery := `
		SELECT id, tenant_id, short_uid, name, COALESCE(description, '') as description, created_at, updated_at
		FROM permissions`

	// Build WHERE clause based on search parameters
	whereConditions := []string{}
	args := []interface{}{}

	// General search across name and description
	if search.Query != "" {
		whereConditions = append(whereConditions, "(name LIKE ? OR description LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Specific field searches
	if nameFilter := search.Filters["name"]; nameFilter != "" {
		whereConditions = append(whereConditions, "name LIKE ?")
		args = append(args, "%"+nameFilter+"%")
	}

	// Tenant filter
	if tenantFilter := search.Filters["tenant_id"]; tenantFilter != "" {
		if tenantFilter == "system" {
			whereConditions = append(whereConditions, "tenant_id = ?")
			args = append(args, SystemTenantID)
		} else {
			whereConditions = append(whereConditions, "tenant_id = ?")
			args = append(args, tenantFilter)
		}
	}

	if categoryFilter := search.Filters["category"]; categoryFilter != "" {
		// For now, we don't have a category field in the database, but we can prepare for it
		// This will be a no-op until the database schema is updated
		whereConditions = append(whereConditions, "1=1") // placeholder
	}

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Add ordering
	orderBy := "tenant_id, name ASC" // Default ordering
	if search.HasSort() {
		// Prevent SQL injection by validating sort direction
		dir := strings.ToUpper(search.SortDir)
		if dir != "ASC" && dir != "DESC" {
			dir = "ASC"
		}
		switch search.SortBy {
		case "name":
			orderBy = "name " + dir
		case "tenant_id":
			orderBy = "tenant_id " + dir + ", name ASC"
		case "created_at", "createdAt":
			orderBy = "created_at " + dir
		case "updated_at", "updatedAt":
			orderBy = "updated_at " + dir
		default:
			// Keep default ordering for unknown sort fields
		}
	}

	baseQuery += " ORDER BY " + orderBy

	// Add pagination
	if limit > 0 {
		baseQuery += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}

	rows, err := a.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []ports.Permission
	for rows.Next() {
		var p ports.Permission
		err := rows.Scan(&p.ID, &p.TenantID, &p.ShortUID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// GetPermissionCount returns the total count of permissions matching optional search criteria
func (a *Adapter) GetPermissionCount(ctx context.Context, search utils.SearchParams) (int, error) {
	baseQuery := `SELECT COUNT(*) FROM permissions`

	// Build WHERE clause based on search parameters (same logic as GetPermissionsWithPagination)
	whereConditions := []string{}
	args := []interface{}{}

	// General search across name and description
	if search.Query != "" {
		whereConditions = append(whereConditions, "(name LIKE ? OR description LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Specific field searches
	if nameFilter := search.Filters["name"]; nameFilter != "" {
		whereConditions = append(whereConditions, "name LIKE ?")
		args = append(args, "%"+nameFilter+"%")
	}

	// Tenant filter
	if tenantFilter := search.Filters["tenant_id"]; tenantFilter != "" {
		if tenantFilter == "system" {
			whereConditions = append(whereConditions, "tenant_id = ?")
			args = append(args, SystemTenantID)
		} else {
			whereConditions = append(whereConditions, "tenant_id = ?")
			args = append(args, tenantFilter)
		}
	}

	if categoryFilter := search.Filters["category"]; categoryFilter != "" {
		// For now, we don't have a category field in the database, but we can prepare for it
		// This will be a no-op until the database schema is updated
		whereConditions = append(whereConditions, "1=1") // placeholder
	}

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	var count int
	err := a.db.QueryRowContext(ctx, baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// DeletePermission deletes a permission by its name (from system tenant by default)
func (a *Adapter) DeletePermission(ctx context.Context, name string) error {
	return a.DeleteTenantPermission(ctx, SystemTenantID, name)
}

// DeleteTenantPermission deletes a tenant-specific permission by its name and tenant ID
func (a *Adapter) DeleteTenantPermission(ctx context.Context, tenantID int, name string) error {
	deleteQuery := `DELETE FROM permissions WHERE name = ? AND tenant_id = ?`
	result, err := a.db.ExecContext(ctx, deleteQuery, name, tenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetPermissionByName returns a permission by its name (searches system tenant first, then all tenants).
func (a *Adapter) GetPermissionByName(ctx context.Context, name string) (*ports.Permission, error) {
	query := `
		SELECT id, tenant_id, short_uid, name, COALESCE(description, '') as description, created_at, updated_at
		FROM permissions
		WHERE name = ?
		ORDER BY tenant_id ASC
		LIMIT 1`

	var p ports.Permission
	err := a.db.QueryRowContext(ctx, query, name).Scan(
		&p.ID, &p.TenantID, &p.ShortUID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// GetTenantPermissionByName returns a tenant-specific permission by its name and tenant ID.
func (a *Adapter) GetTenantPermissionByName(ctx context.Context, tenantID int, name string) (*ports.Permission, error) {
	query := `
		SELECT id, tenant_id, short_uid, name, COALESCE(description, '') as description, created_at, updated_at
		FROM permissions
		WHERE name = ? AND tenant_id = ?`

	var p ports.Permission
	err := a.db.QueryRowContext(ctx, query, name, tenantID).Scan(
		&p.ID, &p.TenantID, &p.ShortUID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// GetRolePermissions returns all permissions assigned to a role.
func (a *Adapter) GetRolePermissions(ctx context.Context, roleID int) ([]ports.Permission, error) {
	query := `
		SELECT p.id, p.tenant_id, p.short_uid, p.name, COALESCE(p.description, '') as description, p.created_at, p.updated_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = ?
		ORDER BY p.tenant_id, p.name`

	rows, err := a.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []ports.Permission
	for rows.Next() {
		var p ports.Permission
		err := rows.Scan(&p.ID, &p.TenantID, &p.ShortUID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// Role methods

// CreateRole creates a new role in the system tenant.
// Returns the role ID.
func (a *Adapter) CreateRole(ctx context.Context, name, description string) (int, error) {
	return a.CreateTenantRoleIfNotExists(ctx, SystemTenantID, name, description)
}

// GetAllRoles returns all roles from the database.
func (a *Adapter) GetAllRoles(ctx context.Context) ([]ports.Role, error) {
	query := `
		SELECT id, tenant_id, short_uid, name, COALESCE(description, '') as description, created_at, updated_at
		FROM roles
		ORDER BY tenant_id, name`

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []ports.Role
	for rows.Next() {
		var r ports.Role
		err := rows.Scan(&r.ID, &r.TenantID, &r.ShortUID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

// GetRolesWithPagination returns roles with pagination support
func (a *Adapter) GetRolesWithPagination(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.Role, error) {
	baseQuery := `
		SELECT id, tenant_id, short_uid, name, COALESCE(description, '') as description, created_at, updated_at
		FROM roles`

	var whereClause string
	var args []interface{}

	// Build WHERE clause based on search parameters
	var conditions []string

	// Handle general search query
	if search.Query != "" {
		conditions = append(conditions, "(name LIKE ? OR description LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Handle specific filters
	if name, exists := search.Filters["name"]; exists && name != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+name+"%")
	}

	// Tenant filter
	if tenantFilter := search.Filters["tenant_id"]; tenantFilter != "" {
		if tenantFilter == "system" {
			conditions = append(conditions, "tenant_id = ?")
			args = append(args, SystemTenantID)
		} else {
			conditions = append(conditions, "tenant_id = ?")
			args = append(args, tenantFilter)
		}
	}

	// Combine conditions
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Handle sorting
	orderClause := " ORDER BY tenant_id, name" // Default sort
	if search.HasSort() {
		direction := "ASC"
		if search.SortDir == "desc" {
			direction = "DESC"
		}

		switch search.SortBy {
		case "name":
			orderClause = fmt.Sprintf(" ORDER BY name %s", direction)
		case "tenant_id":
			orderClause = fmt.Sprintf(" ORDER BY tenant_id %s, name ASC", direction)
		case "created_at":
			orderClause = fmt.Sprintf(" ORDER BY created_at %s", direction)
		case "updated_at":
			orderClause = fmt.Sprintf(" ORDER BY updated_at %s", direction)
		default:
			orderClause = " ORDER BY tenant_id, name ASC" // Fallback
		}
	}

	// Add pagination
	limitClause := ""
	if limit > 0 {
		limitClause = fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}

	query := baseQuery + whereClause + orderClause + limitClause

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []ports.Role
	for rows.Next() {
		var r ports.Role
		err := rows.Scan(&r.ID, &r.TenantID, &r.ShortUID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

// GetRoleCount returns the total count of roles matching optional search criteria
func (a *Adapter) GetRoleCount(ctx context.Context, search utils.SearchParams) (int, error) {
	baseQuery := "SELECT COUNT(*) FROM roles"

	var whereClause string
	var args []interface{}

	// Build WHERE clause based on search parameters
	var conditions []string

	// Handle general search query
	if search.Query != "" {
		conditions = append(conditions, "(name LIKE ? OR description LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Handle specific filters
	if name, exists := search.Filters["name"]; exists && name != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+name+"%")
	}

	// Tenant filter
	if tenantFilter := search.Filters["tenant_id"]; tenantFilter != "" {
		if tenantFilter == "system" {
			conditions = append(conditions, "tenant_id = ?")
			args = append(args, SystemTenantID)
		} else {
			conditions = append(conditions, "tenant_id = ?")
			args = append(args, tenantFilter)
		}
	}

	// Combine conditions
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := baseQuery + whereClause

	var count int
	err := a.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetRoleByName returns a role by its name (searches system tenant first, then all tenants).
func (a *Adapter) GetRoleByName(ctx context.Context, name string) (*ports.Role, error) {
	query := `
		SELECT id, tenant_id, short_uid, name, COALESCE(description, '') as description, created_at, updated_at
		FROM roles
		WHERE name = ?
		ORDER BY tenant_id ASC
		LIMIT 1`

	var r ports.Role
	err := a.db.QueryRowContext(ctx, query, name).Scan(
		&r.ID, &r.TenantID, &r.ShortUID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// GetRoleByID returns a role by its ID.
func (a *Adapter) GetRoleByID(ctx context.Context, roleID int) (*ports.Role, error) {
	query := `
		SELECT id, tenant_id, short_uid, name, COALESCE(description, '') as description, created_at, updated_at
		FROM roles
		WHERE id = ?`

	var r ports.Role
	err := a.db.QueryRowContext(ctx, query, roleID).Scan(
		&r.ID, &r.TenantID, &r.ShortUID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// DeleteRole deletes a role by its name (from system tenant by default).
func (a *Adapter) DeleteRole(ctx context.Context, name string) error {
	deleteQuery := `DELETE FROM roles WHERE name = ? AND tenant_id = ?`
	result, err := a.db.ExecContext(ctx, deleteQuery, name, SystemTenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Role-Permission methods

// AssignPermissionToRole assigns a permission to a role.
func (a *Adapter) AssignPermissionToRole(ctx context.Context, roleID int, permissionID int) error {
	// Check if assignment already exists
	checkQuery := `
		SELECT 1 FROM role_permissions 
		WHERE role_id = ? AND permission_id = ?
		LIMIT 1`

	var tmp int
	err := a.db.QueryRowContext(ctx, checkQuery, roleID, permissionID).Scan(&tmp)
	if err == nil {
		// Assignment already exists
		return nil
	}
	if err != sql.ErrNoRows {
		// Real error
		return err
	}

	// Create the assignment
	insertQuery := `
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES (?, ?)`

	_, err = a.db.ExecContext(ctx, insertQuery, roleID, permissionID)
	return err
}

// RemovePermissionFromRole removes a permission from a role.
func (a *Adapter) RemovePermissionFromRole(ctx context.Context, roleID int, permissionID int) error {
	deleteQuery := `
		DELETE FROM role_permissions 
		WHERE role_id = ? AND permission_id = ?`

	result, err := a.db.ExecContext(ctx, deleteQuery, roleID, permissionID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// User-Role methods

// AssignRoleToUser assigns a role to a user with proper tenant context.
func (a *Adapter) AssignRoleToUser(ctx context.Context, userID int, roleID int, tenantID *int) error {
	// For compatibility with the interface, we handle the nullable tenantID
	// but in our new system, we need to get the role's tenant_id
	var actualTenantID int
	if tenantID != nil {
		actualTenantID = *tenantID
	} else {
		// Get the role's tenant_id
		var roleTenantID int
		roleQuery := `SELECT tenant_id FROM roles WHERE id = ?`
		err := a.db.QueryRowContext(ctx, roleQuery, roleID).Scan(&roleTenantID)
		if err != nil {
			return err
		}
		actualTenantID = roleTenantID
	}

	// TENANT ISOLATION SECURITY CHECK
	// Users can only be assigned roles from:
	// 1. Their home tenant
	// 2. System tenant (ID=0) for platform operations
	// This prevents cross-tenant privilege escalation
	if actualTenantID != SystemTenantID { // If not system tenant, validate home tenant
		// Get the user's home tenant
		var userTenantID int
		userQuery := `SELECT tenant_id FROM users WHERE id = ?`
		err := a.db.QueryRowContext(ctx, userQuery, userID).Scan(&userTenantID)
		if err != nil {
			return fmt.Errorf("failed to get user's home tenant: %w", err)
		}

		// Reject if trying to assign role from different tenant
		if actualTenantID != userTenantID {
			return fmt.Errorf("tenant isolation violation: user (tenant_id=%d) cannot be assigned role from tenant_id=%d", userTenantID, actualTenantID)
		}
	}

	// Check if assignment already exists
	checkQuery := `
		SELECT 1 FROM user_roles 
		WHERE user_id = ? AND role_id = ? AND tenant_id = ?
		LIMIT 1`

	var tmp int
	err := a.db.QueryRowContext(ctx, checkQuery, userID, roleID, actualTenantID).Scan(&tmp)
	if err == nil {
		// Assignment already exists
		return nil
	}
	if err != sql.ErrNoRows {
		// Real error
		return err
	}

	// Create the assignment
	insertQuery := `
		INSERT INTO user_roles (user_id, role_id, tenant_id)
		VALUES (?, ?, ?)`

	_, err = a.db.ExecContext(ctx, insertQuery, userID, roleID, actualTenantID)
	return err
}

// RemoveRoleFromUser removes a role from a user.
func (a *Adapter) RemoveRoleFromUser(ctx context.Context, userID int, roleID int, tenantID *int) error {
	// For compatibility with the interface, we handle the nullable tenantID
	var actualTenantID int
	if tenantID != nil {
		actualTenantID = *tenantID
	} else {
		// Get the role's tenant_id
		var roleTenantID int
		roleQuery := `SELECT tenant_id FROM roles WHERE id = ?`
		err := a.db.QueryRowContext(ctx, roleQuery, roleID).Scan(&roleTenantID)
		if err != nil {
			return err
		}
		actualTenantID = roleTenantID
	}

	// NOTE: For role removal, we're more permissive to allow cleanup of invalid assignments
	// But we still log suspicious cross-tenant removals for security auditing
	if actualTenantID != SystemTenantID {
		// Get the user's home tenant for validation (but don't block removal)
		var userTenantID int
		userQuery := `SELECT tenant_id FROM users WHERE id = ?`
		err := a.db.QueryRowContext(ctx, userQuery, userID).Scan(&userTenantID)
		if err == nil && actualTenantID != userTenantID {
			// Log suspicious cross-tenant role removal but allow it for cleanup
			// In a production system, you might want to log this to an audit trail
			fmt.Printf("WARNING: Cross-tenant role removal - user (tenant_id=%d) removing role from tenant_id=%d\n", userTenantID, actualTenantID)
		}
	}

	deleteQuery := `
		DELETE FROM user_roles 
		WHERE user_id = ? AND role_id = ? AND tenant_id = ?`

	result, err := a.db.ExecContext(ctx, deleteQuery, userID, roleID, actualTenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetUserRoles returns all roles assigned to a user.
func (a *Adapter) GetUserRoles(ctx context.Context, userID int) ([]ports.UserRole, error) {
	query := `
		SELECT id, user_id, role_id, tenant_id, created_at, updated_at
		FROM user_roles
		WHERE user_id = ?
		ORDER BY tenant_id, role_id`

	rows, err := a.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userRoles []ports.UserRole
	for rows.Next() {
		var ur ports.UserRole
		err := rows.Scan(&ur.ID, &ur.UserID, &ur.RoleID, &ur.TenantID, &ur.CreatedAt, &ur.UpdatedAt)
		if err != nil {
			return nil, err
		}
		userRoles = append(userRoles, ur)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userRoles, nil
}

// GetRoleUsers returns all users assigned to a specific role.
func (a *Adapter) GetRoleUsers(ctx context.Context, roleID int) ([]ports.UserRole, error) {
	query := `
		SELECT id, user_id, role_id, tenant_id, created_at, updated_at
		FROM user_roles
		WHERE role_id = ?
		ORDER BY user_id`

	rows, err := a.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userRoles []ports.UserRole
	for rows.Next() {
		var ur ports.UserRole
		err := rows.Scan(&ur.ID, &ur.UserID, &ur.RoleID, &ur.TenantID, &ur.CreatedAt, &ur.UpdatedAt)
		if err != nil {
			return nil, err
		}
		userRoles = append(userRoles, ur)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userRoles, nil
}

// Group methods

// CreateGroup creates a new group in the system tenant.
// Returns the group ID.
func (a *Adapter) CreateGroup(ctx context.Context, name, description string) (int, error) {
	return a.CreateTenantGroup(ctx, SystemTenantID, name, description)
}

// CreateTenantGroup creates a new tenant-specific group in the database.
// Returns the group ID.
func (a *Adapter) CreateTenantGroup(ctx context.Context, tenantID int, name, description string) (int, error) {
	// First check if group exists
	var groupID int
	checkQuery := `SELECT id FROM groups WHERE name = ? AND tenant_id = ?`
	err := a.db.QueryRowContext(ctx, checkQuery, name, tenantID).Scan(&groupID)
	if err == nil {
		// Group already exists, return error since this is explicit creation
		return 0, sql.ErrNoRows // We can define a custom error later
	}
	if err != sql.ErrNoRows {
		// Real error
		return 0, err
	}

	// Generate short UID
	shortUID, err := utils.GenerateFixedShortUID()
	if err != nil {
		return 0, err
	}

	// Group doesn't exist, create it
	insertQuery := `
		INSERT INTO groups (tenant_id, short_uid, name, description)
		VALUES (?, ?, ?, ?)
		RETURNING id`

	err = a.db.QueryRowContext(ctx, insertQuery, tenantID, shortUID, name, description).Scan(&groupID)
	if err != nil {
		return 0, err
	}

	return groupID, nil
}

// GetAllGroups returns all groups from the database.
func (a *Adapter) GetAllGroups(ctx context.Context) ([]ports.Group, error) {
	query := `
		SELECT id, tenant_id, short_uid, name, COALESCE(description, '') as description, created_at, updated_at
		FROM groups
		ORDER BY tenant_id, name`

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []ports.Group
	for rows.Next() {
		var g ports.Group
		err := rows.Scan(&g.ID, &g.TenantID, &g.ShortUID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

// GetGroupsWithPagination returns groups with pagination support
func (a *Adapter) GetGroupsWithPagination(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.Group, error) {
	baseQuery := `
		SELECT id, tenant_id, short_uid, name, COALESCE(description, '') as description, created_at, updated_at
		FROM groups`

	var whereClause string
	var args []interface{}

	// Build WHERE clause based on search parameters
	var conditions []string

	// Handle general search query
	if search.Query != "" {
		conditions = append(conditions, "(name LIKE ? OR description LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Handle specific filters
	if name, exists := search.Filters["name"]; exists && name != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+name+"%")
	}

	// Tenant filter
	if tenantFilter := search.Filters["tenant_id"]; tenantFilter != "" {
		if tenantFilter == "system" {
			conditions = append(conditions, "tenant_id = ?")
			args = append(args, SystemTenantID)
		} else {
			conditions = append(conditions, "tenant_id = ?")
			args = append(args, tenantFilter)
		}
	}

	// Combine conditions
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Handle sorting
	orderClause := " ORDER BY tenant_id, name" // Default sort
	if search.HasSort() {
		direction := "ASC"
		if search.SortDir == "desc" {
			direction = "DESC"
		}

		switch search.SortBy {
		case "name":
			orderClause = fmt.Sprintf(" ORDER BY name %s", direction)
		case "tenant_id":
			orderClause = fmt.Sprintf(" ORDER BY tenant_id %s, name ASC", direction)
		case "created_at":
			orderClause = fmt.Sprintf(" ORDER BY created_at %s", direction)
		case "updated_at":
			orderClause = fmt.Sprintf(" ORDER BY updated_at %s", direction)
		default:
			orderClause = " ORDER BY tenant_id, name ASC" // Fallback
		}
	}

	// Add pagination
	limitClause := ""
	if limit > 0 {
		limitClause = fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}

	query := baseQuery + whereClause + orderClause + limitClause

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []ports.Group
	for rows.Next() {
		var g ports.Group
		err := rows.Scan(&g.ID, &g.TenantID, &g.ShortUID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

// GetGroupCount returns the total count of groups matching optional search criteria
func (a *Adapter) GetGroupCount(ctx context.Context, search utils.SearchParams) (int, error) {
	baseQuery := `SELECT COUNT(*) FROM groups`

	var whereClause string
	var args []interface{}

	// Build WHERE clause based on search parameters
	var conditions []string

	// Handle general search query
	if search.Query != "" {
		conditions = append(conditions, "(name LIKE ? OR description LIKE ?)")
		searchPattern := "%" + search.Query + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Handle specific filters
	if name, exists := search.Filters["name"]; exists && name != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+name+"%")
	}

	// Tenant filter
	if tenantFilter := search.Filters["tenant_id"]; tenantFilter != "" {
		if tenantFilter == "system" {
			conditions = append(conditions, "tenant_id = ?")
			args = append(args, SystemTenantID)
		} else {
			conditions = append(conditions, "tenant_id = ?")
			args = append(args, tenantFilter)
		}
	}

	// Combine conditions
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := baseQuery + whereClause

	var count int
	err := a.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetGroupByName returns a group by its name (searches system tenant first, then all tenants).
func (a *Adapter) GetGroupByName(ctx context.Context, name string) (*ports.Group, error) {
	query := `
		SELECT id, tenant_id, short_uid, name, COALESCE(description, '') as description, created_at, updated_at
		FROM groups
		WHERE name = ?
		ORDER BY tenant_id ASC
		LIMIT 1`

	var g ports.Group
	err := a.db.QueryRowContext(ctx, query, name).Scan(
		&g.ID, &g.TenantID, &g.ShortUID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &g, nil
}

// GetGroupByID returns a group by its ID.
func (a *Adapter) GetGroupByID(ctx context.Context, groupID int) (*ports.Group, error) {
	query := `
		SELECT id, tenant_id, short_uid, name, COALESCE(description, '') as description, created_at, updated_at
		FROM groups
		WHERE id = ?`

	var g ports.Group
	err := a.db.QueryRowContext(ctx, query, groupID).Scan(
		&g.ID, &g.TenantID, &g.ShortUID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &g, nil
}

// DeleteGroup deletes a group by its name (from system tenant by default).
func (a *Adapter) DeleteGroup(ctx context.Context, name string) error {
	deleteQuery := `DELETE FROM groups WHERE name = ? AND tenant_id = ?`
	result, err := a.db.ExecContext(ctx, deleteQuery, name, SystemTenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// User-Group methods

// AssignUserToGroup assigns a user to a group.
func (a *Adapter) AssignUserToGroup(ctx context.Context, userID int, groupID int) error {
	// Check if assignment already exists
	checkQuery := `
		SELECT 1 FROM user_groups 
		WHERE user_id = ? AND group_id = ?
		LIMIT 1`

	var tmp int
	err := a.db.QueryRowContext(ctx, checkQuery, userID, groupID).Scan(&tmp)
	if err == nil {
		// Assignment already exists
		return nil
	}
	if err != sql.ErrNoRows {
		// Real error
		return err
	}

	// Create the assignment
	insertQuery := `
		INSERT INTO user_groups (user_id, group_id)
		VALUES (?, ?)`

	_, err = a.db.ExecContext(ctx, insertQuery, userID, groupID)
	return err
}

// RemoveUserFromGroup removes a user from a group.
func (a *Adapter) RemoveUserFromGroup(ctx context.Context, userID int, groupID int) error {
	deleteQuery := `
		DELETE FROM user_groups 
		WHERE user_id = ? AND group_id = ?`

	result, err := a.db.ExecContext(ctx, deleteQuery, userID, groupID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetUserGroups returns all groups a user belongs to.
func (a *Adapter) GetUserGroups(ctx context.Context, userID int) ([]ports.Group, error) {
	query := `
		SELECT g.id, g.tenant_id, g.short_uid, g.name, COALESCE(g.description, '') as description, g.created_at, g.updated_at
		FROM groups g
		JOIN user_groups ug ON g.id = ug.group_id
		WHERE ug.user_id = ?
		ORDER BY g.tenant_id, g.name`

	rows, err := a.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []ports.Group
	for rows.Next() {
		var g ports.Group
		err := rows.Scan(&g.ID, &g.TenantID, &g.ShortUID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

// GetGroupUsers returns all users in a group.
func (a *Adapter) GetGroupUsers(ctx context.Context, groupID int) ([]int, error) {
	query := `
		SELECT user_id
		FROM user_groups
		WHERE group_id = ?
		ORDER BY user_id`

	rows, err := a.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var userID int
		err := rows.Scan(&userID)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userIDs, nil
}

// Object grant methods

// HasObjectGrant checks if a user, one of their groups, or one of their roles
// has the specified permission on the given resource. This provides basic
// per-object ACL lookups.
func (a *Adapter) HasObjectGrant(ctx context.Context, userID int, permission, resourceType string, resourceID int) (bool, error) {
	query := `
		SELECT 1
		FROM object_grants og
		WHERE og.permission = ?
		  AND og.resource_type = ?
		  AND og.resource_id = ?
		  AND (
			(og.grantee_type = 'user'  AND og.grantee_id = ?)
		 OR (og.grantee_type = 'group' AND og.grantee_id IN (
			SELECT group_id FROM user_groups WHERE user_id = ?
		 ))
		 OR (og.grantee_type = 'role'  AND og.grantee_id IN (
			SELECT role_id FROM user_roles WHERE user_id = ?
		 ))
		  )
		LIMIT 1`

	row := a.db.QueryRowContext(ctx, query,
		permission, resourceType, resourceID,
		userID, userID, userID,
	)

	var tmp int
	err := row.Scan(&tmp)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateObjectGrant creates a new object grant.
func (a *Adapter) CreateObjectGrant(ctx context.Context, granteeType string, granteeID int, permission, resourceType string, resourceID int) error {
	// Check if the grant already exists
	checkQuery := `
		SELECT 1 FROM object_grants 
		WHERE grantee_type = ? AND grantee_id = ? AND permission = ? AND resource_type = ? AND resource_id = ?
		LIMIT 1`

	var tmp int
	err := a.db.QueryRowContext(ctx, checkQuery, granteeType, granteeID, permission, resourceType, resourceID).Scan(&tmp)
	if err == nil {
		// Grant already exists
		return nil
	}
	if err != sql.ErrNoRows {
		// Real error
		return err
	}

	// Create the grant
	insertQuery := `
		INSERT INTO object_grants (grantee_type, grantee_id, permission, resource_type, resource_id)
		VALUES (?, ?, ?, ?, ?)`

	_, err = a.db.ExecContext(ctx, insertQuery, granteeType, granteeID, permission, resourceType, resourceID)
	return err
}

// GetObjectGrants returns all object grants for a specific resource.
func (a *Adapter) GetObjectGrants(ctx context.Context, resourceType string, resourceID int) ([]ports.ObjectGrant, error) {
	query := `
		SELECT id, grantee_type, grantee_id, permission, resource_type, resource_id, created_at, updated_at
		FROM object_grants
		WHERE resource_type = ? AND resource_id = ?
		ORDER BY grantee_type, grantee_id, permission`

	rows, err := a.db.QueryContext(ctx, query, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []ports.ObjectGrant
	for rows.Next() {
		var g ports.ObjectGrant
		err := rows.Scan(&g.ID, &g.GranteeType, &g.GranteeID, &g.Permission, &g.ResourceType, &g.ResourceID, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			return nil, err
		}
		grants = append(grants, g)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return grants, nil
}

// DeleteObjectGrant deletes an object grant.
func (a *Adapter) DeleteObjectGrant(ctx context.Context, granteeType string, granteeID int, permission, resourceType string, resourceID int) error {
	deleteQuery := `
		DELETE FROM object_grants 
		WHERE grantee_type = ? AND grantee_id = ? AND permission = ? AND resource_type = ? AND resource_id = ?`

	result, err := a.db.ExecContext(ctx, deleteQuery, granteeType, granteeID, permission, resourceType, resourceID)
	if err != nil {
		return err
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		// No grant was found to delete - this might not be an error depending on requirements
		// For now, we'll treat it as success (idempotent operation)
		return nil
	}

	return nil
}

// SQL-based permission filtering methods for efficient bulk operations

// GetAccessibleStudyIDs returns study IDs that the user can view with the given permission
func (a *Adapter) GetAccessibleStudyIDs(ctx context.Context, userID int, permission string) ([]int, error) {
	query := `
		SELECT DISTINCT s.id
		FROM studies s
		WHERE s.deleted_at IS NULL
		  AND (
			-- Check if user has role-based permission (takes precedence)
			EXISTS (
				SELECT 1
				FROM user_roles ur
				JOIN role_permissions rp ON ur.role_id = rp.role_id
				JOIN permissions p ON rp.permission_id = p.id
				WHERE ur.user_id = ? AND p.name = ?
			)
			OR
			-- Check if user has direct object grant on study
			EXISTS (
				SELECT 1
				FROM object_grants og
				WHERE og.permission = ?
				  AND og.resource_type = 'study'
				  AND og.resource_id = s.id
				  AND (
					(og.grantee_type = 'user' AND og.grantee_id = ?)
					OR (og.grantee_type = 'group' AND og.grantee_id IN (
						SELECT group_id FROM user_groups WHERE user_id = ?
					))
					OR (og.grantee_type = 'role' AND og.grantee_id IN (
						SELECT role_id FROM user_roles WHERE user_id = ?
					))
				  )
			)
		  )`

	rows, err := a.db.QueryContext(ctx, query, userID, permission, permission, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studyIDs []int
	for rows.Next() {
		var studyID int
		if err := rows.Scan(&studyID); err != nil {
			return nil, err
		}
		studyIDs = append(studyIDs, studyID)
	}

	return studyIDs, rows.Err()
}

// GetAccessibleCaseIDs returns case IDs that the user can view with the given permission
// This includes cases accessible through direct grants and inherited from parent studies
func (a *Adapter) GetAccessibleCaseIDs(ctx context.Context, userID int, permission string) ([]int, error) {
	query := `
		SELECT DISTINCT c.id
		FROM cases c
		WHERE c.deleted_at IS NULL
		  AND (
			-- Check if user has role-based permission (takes precedence)
			EXISTS (
				SELECT 1
				FROM user_roles ur
				JOIN role_permissions rp ON ur.role_id = rp.role_id
				JOIN permissions p ON rp.permission_id = p.id
				WHERE ur.user_id = ? AND p.name = ?
			)
			OR
			-- Check if user has direct object grant on case
			EXISTS (
				SELECT 1
				FROM object_grants og
				WHERE og.permission = ?
				  AND og.resource_type = 'case'
				  AND og.resource_id = c.id
				  AND (
					(og.grantee_type = 'user' AND og.grantee_id = ?)
					OR (og.grantee_type = 'group' AND og.grantee_id IN (
						SELECT group_id FROM user_groups WHERE user_id = ?
					))
					OR (og.grantee_type = 'role' AND og.grantee_id IN (
						SELECT role_id FROM user_roles WHERE user_id = ?
					))
				  )
			)
			OR
			-- Check if user has permission on parent studies (inheritance)
			EXISTS (
				SELECT 1
				FROM study_cases sc
				JOIN object_grants og ON og.resource_type = 'study' AND og.resource_id = sc.study_id
				WHERE sc.case_id = c.id
				  AND og.permission IN (?, 'studies.view')  -- Allow both case permission and studies.view
				  AND (
					(og.grantee_type = 'user' AND og.grantee_id = ?)
					OR (og.grantee_type = 'group' AND og.grantee_id IN (
						SELECT group_id FROM user_groups WHERE user_id = ?
					))
					OR (og.grantee_type = 'role' AND og.grantee_id IN (
						SELECT role_id FROM user_roles WHERE user_id = ?
					))
				  )
			)
		  )`

	rows, err := a.db.QueryContext(ctx, query, userID, permission, permission, userID, userID, userID, permission, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caseIDs []int
	for rows.Next() {
		var caseID int
		if err := rows.Scan(&caseID); err != nil {
			return nil, err
		}
		caseIDs = append(caseIDs, caseID)
	}

	return caseIDs, rows.Err()
}

// GetAccessibleSlideIDs returns slide IDs that the user can view with the given permission
// This includes slides accessible through direct grants and inherited from parent cases/studies
func (a *Adapter) GetAccessibleSlideIDs(ctx context.Context, userID int, permission string) ([]int, error) {
	query := `
		SELECT DISTINCT s.id
		FROM slides s
		WHERE s.deleted_at IS NULL
		  AND (
			-- Check if user has role-based permission (takes precedence)
			EXISTS (
				SELECT 1
				FROM user_roles ur
				JOIN role_permissions rp ON ur.role_id = rp.role_id
				JOIN permissions p ON rp.permission_id = p.id
				WHERE ur.user_id = ? AND p.name = ?
			)
			OR
			-- Check if user has direct object grant on slide
			EXISTS (
				SELECT 1
				FROM object_grants og
				WHERE og.permission = ?
				  AND og.resource_type = 'slide'
				  AND og.resource_id = s.id
				  AND (
					(og.grantee_type = 'user' AND og.grantee_id = ?)
					OR (og.grantee_type = 'group' AND og.grantee_id IN (
						SELECT group_id FROM user_groups WHERE user_id = ?
					))
					OR (og.grantee_type = 'role' AND og.grantee_id IN (
						SELECT role_id FROM user_roles WHERE user_id = ?
					))
				  )
			)
			OR
			-- Check if user has permission on parent case (inheritance)
			EXISTS (
				SELECT 1
				FROM object_grants og
				WHERE og.permission IN (?, 'cases.view')  -- Allow both slide permission and cases.view
				  AND og.resource_type = 'case'
				  AND og.resource_id = s.case_id
				  AND (
					(og.grantee_type = 'user' AND og.grantee_id = ?)
					OR (og.grantee_type = 'group' AND og.grantee_id IN (
						SELECT group_id FROM user_groups WHERE user_id = ?
					))
					OR (og.grantee_type = 'role' AND og.grantee_id IN (
						SELECT role_id FROM user_roles WHERE user_id = ?
					))
				  )
			)
			OR
			-- Check if user has permission on parent studies (inheritance)
			EXISTS (
				SELECT 1
				FROM study_cases sc
				JOIN object_grants og ON og.resource_type = 'study' AND og.resource_id = sc.study_id
				WHERE sc.case_id = s.case_id
				  AND og.permission IN (?, 'studies.view')  -- Allow both slide permission and studies.view
				  AND (
					(og.grantee_type = 'user' AND og.grantee_id = ?)
					OR (og.grantee_type = 'group' AND og.grantee_id IN (
						SELECT group_id FROM user_groups WHERE user_id = ?
					))
					OR (og.grantee_type = 'role' AND og.grantee_id IN (
						SELECT role_id FROM user_roles WHERE user_id = ?
					))
				  )
			)
		  )`

	rows, err := a.db.QueryContext(ctx, query, userID, permission, permission, userID, userID, userID, permission, userID, userID, userID, permission, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slideIDs []int
	for rows.Next() {
		var slideID int
		if err := rows.Scan(&slideID); err != nil {
			return nil, err
		}
		slideIDs = append(slideIDs, slideID)
	}

	return slideIDs, rows.Err()
}
