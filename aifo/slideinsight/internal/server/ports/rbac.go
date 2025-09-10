package ports

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Permission represents a permission in the system
type Permission struct {
	ID          int    `json:"id"`
	TenantID    int    `json:"tenant_id"` // 0 = system tenant, >0 = regular tenant
	ShortUID    string `json:"short_uid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// Role represents a role in the system
type Role struct {
	ID          int    `json:"id"`
	TenantID    int    `json:"tenant_id"` // 0 = system tenant, >0 = regular tenant
	ShortUID    string `json:"short_uid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// Group represents a group in the system
type Group struct {
	ID          int    `json:"id"`
	TenantID    int    `json:"tenant_id"` // 0 = system tenant, >0 = regular tenant
	ShortUID    string `json:"short_uid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// UserRole represents a user-role assignment
type UserRole struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	RoleID    int    `json:"role_id"`
	TenantID  int    `json:"tenant_id"` // Always matches the role's tenant_id
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ObjectGrant represents an object grant in the database
type ObjectGrant struct {
	ID           int    `json:"id"`
	GranteeType  string `json:"grantee_type"`
	GranteeID    int    `json:"grantee_id"`
	Permission   string `json:"permission"`
	ResourceType string `json:"resource_type"`
	ResourceID   int    `json:"resource_id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// RBACRepository defines methods for role-based access control queries.
type RBACRepository interface {
	// UserHasGlobalRole returns true if the user has the specified role without a tenant scope.
	UserHasGlobalRole(ctx context.Context, userID int, roleName string) (bool, error)

	// UserHasRolePermission returns true if the user has the specified permission through any of their assigned roles.
	UserHasRolePermission(ctx context.Context, userID int, permissionName string) (bool, error)

	// HasObjectGrant returns true if the user (or one of their groups/roles) has
	// the given permission on the specified resource.
	HasObjectGrant(ctx context.Context, userID int, permission, resourceType string, resourceID int) (bool, error)

	// CreateRoleIfNotExists creates a role if it doesn't already exist.
	// Returns the role ID.
	CreateRoleIfNotExists(ctx context.Context, roleName, description string) (int, error)

	// CreateTenantRoleIfNotExists creates a tenant-specific role if it doesn't already exist.
	// Returns the role ID.
	CreateTenantRoleIfNotExists(ctx context.Context, tenantID int, roleName, description string) (int, error)

	// AssignGlobalRoleToUser assigns a global role to a user.
	AssignGlobalRoleToUser(ctx context.Context, userID int, roleName string) error

	// Permission methods
	// CreatePermission creates a new permission in the database.
	// Returns the permission ID.
	CreatePermission(ctx context.Context, name, description string) (int, error)

	// CreatePermissionIfNotExists creates a global permission if it doesn't already exist.
	// Returns the permission ID.
	CreatePermissionIfNotExists(ctx context.Context, name, description string) (int, error)

	// CreateTenantPermissionIfNotExists creates a tenant-specific permission if it doesn't already exist.
	// Returns the permission ID.
	CreateTenantPermissionIfNotExists(ctx context.Context, tenantID int, name, description string) (int, error)

	// GetAllPermissions returns all permissions from the database.
	GetAllPermissions(ctx context.Context) ([]Permission, error)

	// GetPermissionsWithPagination returns permissions with pagination support
	GetPermissionsWithPagination(ctx context.Context, search utils.SearchParams, limit, offset int) ([]Permission, error)

	// GetPermissionCount returns the total count of permissions matching optional search criteria
	GetPermissionCount(ctx context.Context, search utils.SearchParams) (int, error)

	// GetPermissionByName returns a permission by its name.
	GetPermissionByName(ctx context.Context, name string) (*Permission, error)

	// GetTenantPermissionByName returns a tenant-specific permission by its name and tenant ID.
	GetTenantPermissionByName(ctx context.Context, tenantID int, name string) (*Permission, error)

	// DeletePermission deletes a permission by its name
	DeletePermission(ctx context.Context, name string) error

	// DeleteTenantPermission deletes a tenant-specific permission by its name and tenant ID
	DeleteTenantPermission(ctx context.Context, tenantID int, name string) error

	// Role methods
	// CreateRole creates a new role in the database.
	// Returns the role ID.
	CreateRole(ctx context.Context, name, description string) (int, error)

	// GetAllRoles returns all roles from the database.
	GetAllRoles(ctx context.Context) ([]Role, error)

	// GetRolesWithPagination returns roles with pagination support
	GetRolesWithPagination(ctx context.Context, search utils.SearchParams, limit, offset int) ([]Role, error)

	// GetRoleCount returns the total count of roles matching optional search criteria
	GetRoleCount(ctx context.Context, search utils.SearchParams) (int, error)

	// GetRoleByName returns a role by its name.
	GetRoleByName(ctx context.Context, name string) (*Role, error)

	// GetRoleByID returns a role by its ID.
	GetRoleByID(ctx context.Context, roleID int) (*Role, error)

	// DeleteRole deletes a role by its name.
	DeleteRole(ctx context.Context, name string) error

	// Role-Permission methods
	// AssignPermissionToRole assigns a permission to a role.
	AssignPermissionToRole(ctx context.Context, roleID int, permissionID int) error

	// RemovePermissionFromRole removes a permission from a role.
	RemovePermissionFromRole(ctx context.Context, roleID int, permissionID int) error

	// GetRolePermissions returns all permissions assigned to a role.
	GetRolePermissions(ctx context.Context, roleID int) ([]Permission, error)

	// User-Role methods
	// AssignRoleToUser assigns a role to a user (with optional tenant scope).
	AssignRoleToUser(ctx context.Context, userID int, roleID int, tenantID *int) error

	// RemoveRoleFromUser removes a role from a user.
	RemoveRoleFromUser(ctx context.Context, userID int, roleID int, tenantID *int) error

	// GetUserRoles returns all roles assigned to a user.
	GetUserRoles(ctx context.Context, userID int) ([]UserRole, error)

	// GetRoleUsers returns all users assigned to a specific role.
	GetRoleUsers(ctx context.Context, roleID int) ([]UserRole, error)

	// Group methods
	// CreateGroup creates a new group in the database.
	// Returns the group ID.
	CreateGroup(ctx context.Context, name, description string) (int, error)

	// CreateTenantGroup creates a new tenant-specific group in the database.
	// Returns the group ID.
	CreateTenantGroup(ctx context.Context, tenantID int, name, description string) (int, error)

	// GetAllGroups returns all groups from the database.
	GetAllGroups(ctx context.Context) ([]Group, error)

	// GetGroupsWithPagination returns groups with pagination support
	GetGroupsWithPagination(ctx context.Context, search utils.SearchParams, limit, offset int) ([]Group, error)

	// GetGroupCount returns the total count of groups matching optional search criteria
	GetGroupCount(ctx context.Context, search utils.SearchParams) (int, error)

	// GetGroupByName returns a group by its name.
	GetGroupByName(ctx context.Context, name string) (*Group, error)

	// GetGroupByID returns a group by its ID.
	GetGroupByID(ctx context.Context, groupID int) (*Group, error)

	// DeleteGroup deletes a group by its name.
	DeleteGroup(ctx context.Context, name string) error

	// User-Group methods
	// AssignUserToGroup assigns a user to a group.
	AssignUserToGroup(ctx context.Context, userID int, groupID int) error

	// RemoveUserFromGroup removes a user from a group.
	RemoveUserFromGroup(ctx context.Context, userID int, groupID int) error

	// GetUserGroups returns all groups a user belongs to.
	GetUserGroups(ctx context.Context, userID int) ([]Group, error)

	// GetGroupUsers returns all users in a group.
	GetGroupUsers(ctx context.Context, groupID int) ([]int, error) // Returns user IDs

	// Object grant methods
	// CreateObjectGrant creates a new object grant.
	CreateObjectGrant(ctx context.Context, granteeType string, granteeID int, permission, resourceType string, resourceID int) error

	// GetObjectGrants returns all object grants for a specific resource.
	GetObjectGrants(ctx context.Context, resourceType string, resourceID int) ([]ObjectGrant, error)

	// DeleteObjectGrant deletes an object grant.
	DeleteObjectGrant(ctx context.Context, granteeType string, granteeID int, permission, resourceType string, resourceID int) error

	// SQL-based permission filtering methods for efficient bulk operations
	// These methods return only the IDs of resources that a user can access
	// to avoid the expensive post-processing filtering currently used.

	// GetAccessibleStudyIDs returns study IDs that the user can view with the given permission
	GetAccessibleStudyIDs(ctx context.Context, userID int, permission string) ([]int, error)

	// GetAccessibleCaseIDs returns case IDs that the user can view with the given permission
	// This includes cases accessible through direct grants and inherited from parent studies
	GetAccessibleCaseIDs(ctx context.Context, userID int, permission string) ([]int, error)

	// GetAccessibleSlideIDs returns slide IDs that the user can view with the given permission
	// This includes slides accessible through direct grants and inherited from parent cases/studies
	GetAccessibleSlideIDs(ctx context.Context, userID int, permission string) ([]int, error)
}
