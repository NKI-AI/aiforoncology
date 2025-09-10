// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

// SlideInsight API
//
//	@title			SlideInsight API
//	@version		1.0
//	@description	A digital pathology slide management and viewing API
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	API Support
//	@contact.url	http://www.example.com/support
//	@contact.email	support@example.com
//
//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT
//
//	@host		localhost:8080
//	@BasePath	/api/v1
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"aifo.dev/aifo/slideinsight/internal/config"
	"aifo.dev/aifo/slideinsight/internal/datasources"
	"aifo.dev/aifo/slideinsight/internal/datasources/database"
	"aifo.dev/aifo/slideinsight/internal/server"
	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse command-line flags
	configFile := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, warnings, err := config.LoadFromFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Merge with environment variables
	cfg = config.MergeWithEnv(cfg)

	// Log configuration warnings
	for _, warning := range warnings {
		slog.Warn("Configuration warning", "field", warning.Field, "message", warning.Message)
	}

	// Initialize logger based on configuration
	initLogger(cfg.Logging.Level)

	// Connect to database
	db, err := database.NewDatabase(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	// Initialize the system in the correct order
	if err := initializeSystem(ctx, db); err != nil {
		log.Fatalf("Failed to initialize system: %v", err)
	}

	defer db.CloseConnections()

	// Create and start server
	app := server.NewServer(ctx, cfg, &datasources.DataSources{DB: db})
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	slog.Info("Starting server", "address", serverAddr)
	log.Fatal(app.Listen(serverAddr))
}

// initLogger configures the global logger based on the specified level
func initLogger(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// initializeSystem initializes the system in the correct dependency order
func initializeSystem(ctx context.Context, db ports.Database) error {
	slog.Info("Initializing SlideInsight system using Google Cloud IAM style tenant isolation...")

	// Step 1: Initialize system tenant (ID=0)
	if err := initializeSystemTenant(ctx, db); err != nil {
		return fmt.Errorf("failed to initialize system tenant: %w", err)
	}
	slog.Info("✓ System tenant (ID=0) initialized")

	// Step 2: Initialize default permissions in system tenant
	if err := initializeSystemPermissions(ctx, db); err != nil {
		return fmt.Errorf("failed to initialize system permissions: %w", err)
	}
	slog.Info("✓ System permissions initialized")

	// Step 3: Initialize default roles in system tenant
	if err := initializeSystemRoles(ctx, db); err != nil {
		return fmt.Errorf("failed to initialize system roles: %w", err)
	}
	slog.Info("✓ System roles initialized")

	// Step 4: Initialize default tenant
	defaultTenantUID, tenantID, err := initializeDefaultTenant(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to initialize default tenant: %w", err)
	}
	slog.Info("✓ Default tenant initialized", "tenant_uid", defaultTenantUID, "tenant_id", tenantID)

	// Step 5: Initialize tenant-specific permissions for default tenant
	if err := initializeTenantPermissions(ctx, db, tenantID); err != nil {
		return fmt.Errorf("failed to initialize tenant permissions: %w", err)
	}
	slog.Info("✓ Tenant permissions initialized for default tenant")

	// Step 6: Initialize tenant-specific roles for default tenant
	if err := initializeTenantRoles(ctx, db, tenantID); err != nil {
		return fmt.Errorf("failed to initialize tenant roles: %w", err)
	}
	slog.Info("✓ Tenant roles initialized for default tenant")

	// Step 7: Initialize default user
	adminUserID, err := initializeDefaultUser(ctx, db, defaultTenantUID)
	if err != nil {
		return fmt.Errorf("failed to initialize default user: %w", err)
	}
	slog.Info("✓ Default admin user initialized", "user_id", adminUserID)

	// Step 8: Assign superadmin role in system tenant
	if err := assignSystemSuperAdminRole(ctx, db, adminUserID); err != nil {
		return fmt.Errorf("failed to assign system superadmin role: %w", err)
	}
	slog.Info("✓ System superadmin role assigned to default user")

	// Step 9: Assign admin role in default tenant
	if err := assignTenantAdminRole(ctx, db, adminUserID, tenantID); err != nil {
		return fmt.Errorf("failed to assign tenant admin role: %w", err)
	}
	slog.Info("✓ Tenant admin role assigned to default user", "tenant_id", tenantID)

	// Step 10: Initialize default email templates (now that user exists)
	if err := initializeDefaultEmailTemplates(ctx, db, adminUserID); err != nil {
		return fmt.Errorf("failed to initialize default email templates: %w", err)
	}
	slog.Info("✓ Default email templates initialized")

	slog.Info("🎉 SlideInsight system initialization completed successfully!")
	return nil
}

// initializeSystemTenant creates the special system tenant (ID=0) for platform operations
func initializeSystemTenant(ctx context.Context, db ports.Database) error {
	// Check if system tenant already exists by checking for tenant with ID=0
	// The schema initialization should have already created this
	existingTenant, err := db.GetTenantByUID(ctx, "system")
	if err == nil {
		// System tenant already exists, verify it has ID=0
		if existingTenant.ID == 0 {
			slog.Info("System tenant already exists", "tenant_id", existingTenant.ID, "tenant_uid", existingTenant.TenantUID)
			return nil
		} else {
			return fmt.Errorf("system tenant exists but has wrong ID: expected 0, got %d", existingTenant.ID)
		}
	}

	// If we get here, the system tenant doesn't exist, which means schema initialization failed
	// This should not happen in normal operation since the schema creation now includes system tenant creation
	return fmt.Errorf("system tenant not found - this indicates a database schema initialization problem: %w", err)
}

// initializeSystemPermissions creates system-level permissions in the system tenant
func initializeSystemPermissions(ctx context.Context, db ports.Database) error {
	systemPermissions := []struct {
		name        string
		description string
	}{
		// Platform administration permissions (system tenant only)
		{"platform.admin", "Full platform administration access across all tenants"},
		{"platform.tenant.create", "Create new tenants"},
		{"platform.tenant.delete", "Delete tenants"},
		{"platform.tenant.manage", "Manage tenant settings and users"},
		{"platform.system.view_logs", "View system-wide logs"},
		{"platform.system.manage_settings", "Manage platform-wide settings"},
		{"platform.billing.manage", "Manage billing and subscriptions"},
		{"platform.support", "Platform support operations"},

		// Cross-tenant permissions (for system admins)
		{"cross_tenant.users.view", "View users across all tenants"},
		{"cross_tenant.studies.view", "View studies across all tenants"},
		{"cross_tenant.analytics.view", "View analytics across all tenants"},
	}

	// Create permissions in system tenant (tenant_id=0)
	for _, perm := range systemPermissions {
		_, err := db.CreateTenantPermissionIfNotExists(ctx, 0, perm.name, perm.description)
		if err != nil {
			slog.Warn("Failed to create system permission (may already exist)", "permission", perm.name, "error", err)
		} else {
			slog.Debug("Created system permission", "name", perm.name, "description", perm.description)
		}
	}

	return nil
}

// initializeSystemRoles creates system-level roles in the system tenant
func initializeSystemRoles(ctx context.Context, db ports.Database) error {
	// Create system roles in system tenant (tenant_id=0)
	systemRoles := []struct {
		name        string
		description string
		permissions []string
	}{
		{
			name:        "superadmin",
			description: "Super administrator with full platform access",
			permissions: []string{
				"platform.admin",
				"platform.tenant.create",
				"platform.tenant.delete",
				"platform.tenant.manage",
				"platform.system.view_logs",
				"platform.system.manage_settings",
				"cross_tenant.users.view",
				"cross_tenant.studies.view",
				"cross_tenant.analytics.view",
			},
		},
		{
			name:        "platform_admin",
			description: "Platform administrator with tenant management access",
			permissions: []string{
				"platform.tenant.create",
				"platform.tenant.manage",
				"platform.system.view_logs",
				"cross_tenant.users.view",
				"cross_tenant.analytics.view",
			},
		},
		{
			name:        "support",
			description: "Platform support with limited access",
			permissions: []string{
				"platform.support",
				"platform.system.view_logs",
				"cross_tenant.users.view",
			},
		},
	}

	for _, roleSpec := range systemRoles {
		roleID, err := db.CreateTenantRoleIfNotExists(ctx, 0, roleSpec.name, roleSpec.description)
		if err != nil {
			return fmt.Errorf("failed to create system role %s: %w", roleSpec.name, err)
		}

		// Assign permissions to role
		for _, permissionName := range roleSpec.permissions {
			permission, err := db.GetTenantPermissionByName(ctx, 0, permissionName)
			if err != nil {
				slog.Warn("Failed to find system permission for role assignment", "permission", permissionName, "role", roleSpec.name, "error", err)
				continue
			}
			if permission == nil {
				slog.Warn("System permission not found for role assignment", "permission", permissionName, "role", roleSpec.name)
				continue
			}

			err = db.AssignPermissionToRole(ctx, roleID, permission.ID)
			if err != nil {
				slog.Warn("Failed to assign system permission to role", "permission", permissionName, "role", roleSpec.name, "error", err)
			}
		}

		slog.Debug("Created system role with permissions", "role", roleSpec.name, "role_id", roleID, "permissions", len(roleSpec.permissions))
	}

	return nil
}

// initializeDefaultTenant creates a default tenant if none exists (returns UID and internal ID)
func initializeDefaultTenant(ctx context.Context, db ports.Database) (string, int, error) {
	tenantsService := services.NewTenantsService(db)
	defer tenantsService.Close()

	pagination := utils.PaginationParams{Page: 1, Limit: 100}
	tenants, _, err := tenantsService.GetTenants(ctx, pagination)
	if err != nil {
		return "", 0, fmt.Errorf("failed to load tenants: %w", err)
	}

	// Filter out the system tenant (ID=0) from regular tenants
	// We need to check the internal database records to get the ID
	var regularTenants []domain.Tenant
	for _, tenant := range tenants {
		// Get the database record to check the internal ID
		dbTenant, err := db.GetTenantByUID(ctx, tenant.TenantUID)
		if err != nil {
			slog.Warn("Failed to retrieve tenant details for filtering", "tenant_uid", tenant.TenantUID, "error", err)
			continue // Skip tenants we can't fetch
		}
		if dbTenant.ID != 0 { // Skip system tenant
			regularTenants = append(regularTenants, tenant)
		}
	}

	if len(regularTenants) == 0 {
		slog.Info("No regular tenants found, creating default tenant")
		tenant := domain.Tenant{
			Name: "default",
		}
		createdTenant, err := tenantsService.SaveTenant(ctx, tenant)
		if err != nil {
			return "", 0, fmt.Errorf("failed to create default tenant: %w", err)
		}

		// Get the database record to get the internal ID
		dbTenant, err := db.GetTenantByUID(ctx, createdTenant.TenantUID)
		if err != nil {
			return "", 0, fmt.Errorf("failed to retrieve created tenant: %w", err)
		}

		return createdTenant.TenantUID, dbTenant.ID, nil
	}

	// Use the first existing regular tenant
	firstTenant := regularTenants[0]

	// Get the database record to get the internal ID
	dbTenant, err := db.GetTenantByUID(ctx, firstTenant.TenantUID)
	if err != nil {
		return "", 0, fmt.Errorf("failed to retrieve existing tenant: %w", err)
	}

	slog.Info("Using existing tenant", "tenant_uid", firstTenant.TenantUID, "tenant_id", dbTenant.ID)
	return firstTenant.TenantUID, dbTenant.ID, nil
}

// initializeTenantPermissions creates standard permissions for a specific tenant
func initializeTenantPermissions(ctx context.Context, db ports.Database, tenantID int) error {
	tenantPermissions := []struct {
		name        string
		description string
	}{
		// Study permissions
		{"studies.view", "View study details and content"},
		{"studies.create", "Create new studies"},
		{"studies.edit", "Edit existing studies"},
		{"studies.delete", "Delete studies"},
		{"studies.add_case", "Add cases to studies"},
		{"studies.modify_case", "Modify or remove cases in studies"},
		{"studies.annotate_case", "Create and edit annotations on study cases"},

		// Case permissions
		{"cases.view", "View case details and content"},
		{"cases.create", "Create new cases"},
		{"cases.edit", "Edit existing cases"},
		{"cases.delete", "Delete cases"},
		{"cases.add_slide", "Add slides to cases"},
		{"cases.modify_slide", "Modify or remove slides in cases"},

		// Slide permissions
		{"slides.view", "View slide content"},
		{"slides.create", "Upload new slides"},
		{"slides.edit", "Edit slide metadata"},
		{"slides.delete", "Delete slides"},
		{"slides.annotate", "Create and edit slide annotations"},

		// Annotation permissions
		{"annotations.view", "View annotations"},
		{"annotations.create", "Create new annotations"},
		{"annotations.edit", "Edit existing annotations"},
		{"annotations.delete", "Delete annotations"},

		// Tenant user management permissions
		{"tenant.users.view", "View tenant users"},
		{"tenant.users.create", "Create tenant user accounts"},
		{"tenant.users.edit", "Edit tenant user accounts"},
		{"tenant.users.delete", "Delete tenant user accounts"},
		{"tenant.users.manage_permissions", "Manage tenant user permissions"},

		// Tenant administration permissions
		{"tenant.admin", "Full tenant administration access"},
		{"tenant.settings.view", "View tenant settings"},
		{"tenant.settings.edit", "Edit tenant settings"},
	}

	for _, perm := range tenantPermissions {
		_, err := db.CreateTenantPermissionIfNotExists(ctx, tenantID, perm.name, perm.description)
		if err != nil {
			slog.Warn("Failed to create tenant permission (may already exist)", "tenant_id", tenantID, "permission", perm.name, "error", err)
		} else {
			slog.Debug("Created tenant permission", "tenant_id", tenantID, "name", perm.name, "description", perm.description)
		}
	}

	return nil
}

// initializeTenantRoles creates default roles for a specific tenant
func initializeTenantRoles(ctx context.Context, db ports.Database, tenantID int) error {
	tenantRoles := []struct {
		name        string
		description string
		permissions []string
	}{
		{
			name:        "admin",
			description: "Tenant administrator with full access to tenant resources",
			permissions: []string{
				"tenant.admin",
				"tenant.settings.view",
				"tenant.settings.edit",
				"tenant.users.view",
				"tenant.users.create",
				"tenant.users.edit",
				"tenant.users.delete",
				"tenant.users.manage_permissions",
				"studies.view", "studies.create", "studies.edit", "studies.delete",
				"studies.add_case", "studies.modify_case", "studies.annotate_case",
				"cases.view", "cases.create", "cases.edit", "cases.delete",
				"cases.add_slide", "cases.modify_slide",
				"slides.view", "slides.create", "slides.edit", "slides.delete", "slides.annotate",
				"annotations.view", "annotations.create", "annotations.edit", "annotations.delete",
			},
		},
		{
			name:        "researcher",
			description: "Researcher with access to studies, cases, and annotations",
			permissions: []string{
				"studies.view", "studies.create", "studies.edit",
				"studies.add_case", "studies.modify_case", "studies.annotate_case",
				"cases.view", "cases.create", "cases.edit",
				"cases.add_slide", "cases.modify_slide",
				"slides.view", "slides.create", "slides.edit", "slides.annotate",
				"annotations.view", "annotations.create", "annotations.edit", "annotations.delete",
			},
		},
		{
			name:        "viewer",
			description: "Read-only access to tenant content",
			permissions: []string{
				"studies.view",
				"cases.view",
				"slides.view",
				"annotations.view",
			},
		},
	}

	for _, roleSpec := range tenantRoles {
		roleID, err := db.CreateTenantRoleIfNotExists(ctx, tenantID, roleSpec.name, roleSpec.description)
		if err != nil {
			return fmt.Errorf("failed to create tenant role %s: %w", roleSpec.name, err)
		}

		// Assign permissions to role
		for _, permissionName := range roleSpec.permissions {
			permission, err := db.GetTenantPermissionByName(ctx, tenantID, permissionName)
			if err != nil {
				slog.Warn("Failed to find tenant permission for role assignment", "tenant_id", tenantID, "permission", permissionName, "role", roleSpec.name, "error", err)
				continue
			}
			if permission == nil {
				slog.Warn("Tenant permission not found for role assignment", "tenant_id", tenantID, "permission", permissionName, "role", roleSpec.name)
				continue
			}

			err = db.AssignPermissionToRole(ctx, roleID, permission.ID)
			if err != nil {
				slog.Warn("Failed to assign tenant permission to role", "tenant_id", tenantID, "permission", permissionName, "role", roleSpec.name, "error", err)
			}
		}

		slog.Debug("Created tenant role with permissions", "tenant_id", tenantID, "role", roleSpec.name, "role_id", roleID, "permissions", len(roleSpec.permissions))
	}

	return nil
}

// initializeDefaultUser creates the default admin user if no users exist
func initializeDefaultUser(ctx context.Context, db ports.Database, defaultTenantUID string) (int, error) {
	usersService := services.NewUserService(db)
	defer usersService.Close()

	usersPagination := utils.PaginationParams{Page: 1, Limit: 100}
	users, _, err := usersService.GetUsers(ctx, usersPagination)
	if err != nil {
		return 0, fmt.Errorf("failed to load users: %w", err)
	}

	if len(users) == 0 {
		slog.Info("No users found, creating default admin user")
		user := domain.User{
			TenantUID:         defaultTenantUID,
			Email:             "admin@slideinsight.net",
			Password:          "admin",
			FirstName:         "Admin",
			LastName:          "User",
			IsActive:          true,
			MustResetPassword: true,
		}
		createdUser, err := usersService.CreateUser(ctx, user)
		if err != nil {
			return 0, fmt.Errorf("failed to create default user: %w", err)
		}

		return createdUser.ID, nil
	}

	slog.Info("Users already exist, using first user as admin", "user_count", len(users))
	return users[0].ID, nil
}

// assignSystemSuperAdminRole assigns the superadmin role from system tenant to the given user
func assignSystemSuperAdminRole(ctx context.Context, db ports.Database, userID int) error {
	// For system tenant roles, we search by name and hope it finds the system tenant one first
	// Since GetRoleByName searches system tenant (ID=0) first, this should work
	role, err := db.GetRoleByName(ctx, "superadmin")
	if err != nil {
		return fmt.Errorf("failed to get system superadmin role: %w", err)
	}
	if role == nil {
		return fmt.Errorf("system superadmin role not found")
	}

	// Verify this is a system tenant role (tenant_id=0)
	if role.TenantID != 0 {
		return fmt.Errorf("expected superadmin role from system tenant (ID=0), but got tenant_id=%d", role.TenantID)
	}

	// Assign role from system tenant
	return db.AssignRoleToUser(ctx, userID, role.ID, &role.TenantID)
}

// assignTenantAdminRole assigns the tenant admin role to the given user
func assignTenantAdminRole(ctx context.Context, db ports.Database, userID int, tenantID int) error {
	// Search for admin role - this is tricky because GetRoleByName searches system tenant first
	// We need to get all roles and find the one for our tenant
	allRoles, err := db.GetAllRoles(ctx)
	if err != nil {
		return fmt.Errorf("failed to get roles: %w", err)
	}

	var adminRole *ports.Role
	for _, role := range allRoles {
		if role.Name == "admin" && role.TenantID == tenantID {
			adminRole = &role
			break
		}
	}

	if adminRole == nil {
		return fmt.Errorf("tenant admin role not found for tenant %d", tenantID)
	}

	// Assign role from specific tenant
	return db.AssignRoleToUser(ctx, userID, adminRole.ID, &adminRole.TenantID)
}

// initializeDefaultEmailTemplates ensures that default email templates exist for all tenants
func initializeDefaultEmailTemplates(ctx context.Context, db ports.Database, adminUserID int) error {
	// Get all tenants
	allTenants, err := db.LoadAllTenants(ctx, utils.SearchParams{}, utils.PaginationParams{Page: 1, Limit: 1000})
	if err != nil {
		return fmt.Errorf("failed to load tenants: %w", err)
	}

	for _, tenant := range allTenants {
		// Check if email templates already exist for this tenant
		templates, _, err := db.GetEmailTemplates(ctx, tenant.ID, 1, 0)
		if err != nil {
			slog.Warn("Failed to check existing email templates for tenant", "tenant_id", tenant.ID, "error", err)
			continue
		}

		// If no templates exist, create default ones
		if len(templates) == 0 {
			err = db.CreateDefaultTemplates(ctx, tenant.ID, adminUserID)
			if err != nil {
				slog.Warn("Failed to create default email templates for tenant", "tenant_id", tenant.ID, "error", err)
			} else {
				slog.Info("Created default email templates for tenant", "tenant_id", tenant.ID, "tenant_name", tenant.Name)
			}
		}
	}

	return nil
}
