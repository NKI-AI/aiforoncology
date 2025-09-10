// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
)

// TenantsService is an interface that defines the methods for the tenants service.
// Interface is needed for mocking in tests.
type TenantsService interface {
	GetTenants(ctx context.Context, pagination utils.PaginationParams) ([]domain.Tenant, domain.PaginationInfo, error)
	GetTenantsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Tenant, domain.PaginationInfo, error)
	GetTenantsCount(ctx context.Context) (int, error)
	GetTenantByUID(ctx context.Context, tenantUID string) (domain.Tenant, error)
	SaveTenant(ctx context.Context, newTenant domain.Tenant) (domain.Tenant, error)
	UpdateTenant(ctx context.Context, tenantUID string, updates domain.TenantUpdates) error
	DeleteTenant(ctx context.Context, tenantUID string) error

	// Domain management methods
	GetTenantDomains(ctx context.Context, tenantUID string) ([]domain.TenantDomain, error)
	AddTenantDomain(ctx context.Context, tenantUID string, request domain.NewTenantDomainRequest) error
	UpdateTenantDomain(ctx context.Context, domainID int, updates domain.TenantDomainUpdates) error
	RemoveTenantDomain(ctx context.Context, domainID int) error

	Close()
}

type tenantsService struct {
	*BaseService
	// Generic pagination and search service
	paginatedSearchService *PaginatedSearchService[ports.Tenant, domain.Tenant]
}

// tenantConversionHelpers provides conversion helpers configured for tenants (using RFC3339)
var tenantConversionHelpers = DefaultConversionHelpers()

// convertTenantDBToDomain converts a database Tenant record to a domain Tenant model using conversion helpers
func convertTenantDBToDomain(record ports.Tenant) domain.Tenant {
	return ConvertDBToDomain(
		record,
		tenantConversionHelpers,
		convertTenantBase,
	)
}

// convertTenantBase handles the basic tenant conversion
func convertTenantBase(record ports.Tenant, helpers *ConversionHelpers) domain.Tenant {
	return domain.Tenant{
		TenantUID:   record.TenantUID,
		Name:        record.Name,
		Description: record.Description,
		Status:      record.Status,
		CreatedAt:   helpers.FormatTime(record.CreatedAt),
	}
}

// NewTenantsService creates a new TenantsService
func NewTenantsService(db ports.Database) TenantsService {
	// Create the base service
	baseService := NewBaseService(db)

	// Create the generic paginated search service
	paginatedSearchService := NewPaginatedSearchService(
		func(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.Tenant, error) {
			return baseService.GetDB().LoadAllTenants(ctx, search, utils.PaginationParams{
				Page:  (offset / limit) + 1,
				Limit: limit,
			})
		},
		baseService.GetDB().GetTenantsCountWithSearch,
		func(ctx context.Context, limit, offset int) ([]ports.Tenant, error) {
			return baseService.GetDB().LoadAllTenants(ctx, utils.SearchParams{}, utils.PaginationParams{
				Page:  (offset / limit) + 1,
				Limit: limit,
			})
		},
		baseService.GetDB().GetTenantsCount,
		convertTenantDBToDomain,
	)

	return &tenantsService{
		BaseService:            baseService,
		paginatedSearchService: paginatedSearchService,
	}
}

// GetTenantsGeneric uses the generic search pattern
func (s *tenantsService) GetTenantsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Tenant, domain.PaginationInfo, error) {
	return s.paginatedSearchService.GetWithPaginationAndSearch(ctx, params)
}

// GetTenants retrieves all tenants from the database with pagination support
func (s *tenantsService) GetTenants(ctx context.Context, pagination utils.PaginationParams) ([]domain.Tenant, domain.PaginationInfo, error) {
	// Convert PaginationParams to PaginationAndSearchParams for generic service
	params := utils.PaginationAndSearchParams{
		PaginationParams: pagination,
		SearchParams:     utils.SearchParams{}, // Empty search params for simple pagination
	}
	return s.paginatedSearchService.GetWithPaginationAndSearch(ctx, params)
}

// GetTenantsCount retrieves the total count of tenants from the database
func (s *tenantsService) GetTenantsCount(ctx context.Context) (int, error) {
	return s.GetDB().GetTenantsCount(ctx)
}

// GetTenantByUID retrieves a specific tenant by its short_uid
func (s *tenantsService) GetTenantByUID(ctx context.Context, tenantUID string) (domain.Tenant, error) {
	dbRecord, err := s.GetDB().GetTenantByUID(ctx, tenantUID)
	if err != nil {
		return domain.Tenant{}, errors.WithDetails(errors.ErrTenantNotFound, "tenant with UID '%s'", tenantUID)
	}

	return domain.Tenant{
		TenantUID:   dbRecord.TenantUID,
		Name:        dbRecord.Name,
		Description: dbRecord.Description,
		Status:      dbRecord.Status,
		CreatedAt:   dbRecord.CreatedAt.Format(time.RFC3339),
	}, nil
}

// SaveTenant saves a tenant to the database
func (s *tenantsService) SaveTenant(ctx context.Context, tenant domain.Tenant) (domain.Tenant, error) {
	// Check if tenant already exists
	existingTenant := false
	if tenant.TenantUID != "" {
		_, err := s.GetDB().GetTenantByUID(ctx, tenant.TenantUID)
		if err == nil {
			existingTenant = true
		}
	}

	if existingTenant {
		// Update existing tenant
		updates := domain.TenantUpdates{
			Name:        &tenant.Name,
			Description: &tenant.Description,
			Status:      &tenant.Status,
		}
		err := s.UpdateTenant(ctx, tenant.TenantUID, updates)
		if err != nil {
			return domain.Tenant{}, err
		}
		// Return the updated tenant
		return s.GetTenantByUID(ctx, tenant.TenantUID)
	}

	// Create new tenant
	randomID, err := utils.GenerateFixedShortUID()
	if err != nil {
		return domain.Tenant{}, errors.WithDetails(errors.ErrInternal, "failed to generate tenant ID: %v", err)
	}
	tenant.TenantUID = randomID

	// Error if tenant name is empty
	if tenant.Name == "" {
		return domain.Tenant{}, errors.WithDetails(errors.ErrInvalidInput, "tenant name cannot be empty")
	}

	// Log the tenant metadata
	log.Info("Saving tenant", "tenant", tenant)

	dbTenant := ports.NewTenant{
		TenantUID:   tenant.TenantUID,
		Name:        tenant.Name,
		Description: tenant.Description,
	}

	err = s.GetDB().CreateTenant(ctx, dbTenant)
	if err != nil {
		return domain.Tenant{}, errors.WithDetails(errors.ErrInternal, "failed to save tenant: %v", err)
	}

	// Get the created tenant to get the internal ID
	createdDbTenant, err := s.GetDB().GetTenantByUID(ctx, tenant.TenantUID)
	if err != nil {
		return domain.Tenant{}, errors.WithDetails(errors.ErrInternal, "failed to retrieve created tenant: %v", err)
	}

	// Create default tenant-specific roles and groups
	if err := s.createDefaultTenantRoles(ctx, createdDbTenant.ID); err != nil {
		log.Error("Failed to create default roles for new tenant", "tenant_uid", tenant.TenantUID, "error", err)
		// Don't fail tenant creation for this
	} else {
		log.Info("Created default roles for new tenant", "tenant_uid", tenant.TenantUID)
	}

	// Create default tenant-specific permissions
	if err := s.createDefaultTenantPermissions(ctx, createdDbTenant.ID); err != nil {
		log.Error("Failed to create default permissions for new tenant", "tenant_uid", tenant.TenantUID, "error", err)
		// Don't fail tenant creation for this
	} else {
		log.Info("Created default permissions for new tenant", "tenant_uid", tenant.TenantUID)
	}

	// Create default email templates if we have an authenticated user context
	// (skip during system initialization to avoid foreign key constraint issues)
	if userID := s.getUserIDFromContext(ctx); userID > 0 {
		if err := s.createDefaultEmailTemplates(ctx, createdDbTenant.ID, userID); err != nil {
			log.Error("Failed to create default email templates for new tenant", "tenant_uid", tenant.TenantUID, "error", err)
			// Don't fail tenant creation for this
		} else {
			log.Info("Created default email templates for new tenant", "tenant_uid", tenant.TenantUID)
		}
	} else {
		log.Info("Skipping email template creation during system initialization", "tenant_uid", tenant.TenantUID)
	}

	return tenant, nil
}

// UpdateTenant updates a tenant in the database
func (s *tenantsService) UpdateTenant(ctx context.Context, tenantUID string, updates domain.TenantUpdates) error {
	if tenantUID == "" {
		return errors.ErrTenantUIDEmpty
	}

	// Verify tenant exists
	_, err := s.GetDB().GetTenantByUID(ctx, tenantUID)
	if err != nil {
		return errors.NewTenantNotFoundError(tenantUID)
	}

	// Convert domain updates to ports updates
	portsUpdates := ports.TenantUpdates{
		Name:        updates.Name,
		Description: updates.Description,
		Status:      updates.Status,
	}

	// Update the tenant in the database
	err = s.GetDB().UpdateTenant(ctx, tenantUID, portsUpdates)
	if err != nil {
		return errors.WithDetails(errors.ErrDatabaseUpdate, "failed to update tenant: %v", err)
	}

	return nil
}

// Domain management methods

// GetTenantDomains retrieves all domains for a tenant
func (s *tenantsService) GetTenantDomains(ctx context.Context, tenantUID string) ([]domain.TenantDomain, error) {
	// First get the tenant to ensure it exists and get its internal ID
	tenant, err := s.GetDB().GetTenantByUID(ctx, tenantUID)
	if err != nil {
		return nil, errors.WithDetails(errors.ErrTenantNotFound, "tenant with UID '%s'", tenantUID)
	}

	// Get domains from database
	dbDomains, err := s.GetDB().GetTenantDomains(ctx, tenant.ID)
	if err != nil {
		return nil, errors.WithDetails(errors.ErrInternal, "failed to get tenant domains: %v", err)
	}

	// Convert to domain models
	domains := make([]domain.TenantDomain, 0, len(dbDomains))
	for _, dbDomain := range dbDomains {
		domains = append(domains, domain.TenantDomain{
			ID:         dbDomain.ID,
			Domain:     dbDomain.Domain,
			IsVerified: dbDomain.IsVerified,
			IsPrimary:  dbDomain.IsPrimary,
			CreatedAt:  dbDomain.CreatedAt.Format(time.RFC3339),
			UpdatedAt:  dbDomain.UpdatedAt.Format(time.RFC3339),
		})
	}

	return domains, nil
}

// AddTenantDomain adds a new domain to a tenant
func (s *tenantsService) AddTenantDomain(ctx context.Context, tenantUID string, request domain.NewTenantDomainRequest) error {
	// First get the tenant to ensure it exists and get its internal ID
	tenant, err := s.GetDB().GetTenantByUID(ctx, tenantUID)
	if err != nil {
		return errors.WithDetails(errors.ErrTenantNotFound, "tenant with UID '%s'", tenantUID)
	}

	// Check if domain already exists
	if exists, err := s.GetDB().DomainExists(ctx, request.Domain); err != nil {
		return errors.WithDetails(errors.ErrInternal, "failed to check domain existence: %v", err)
	} else if exists {
		return errors.WithDetails(errors.ErrAlreadyExists, "domain '%s' is already registered to a tenant", request.Domain)
	}

	// Create domain
	newDomain := ports.NewTenantDomain{
		TenantID:   tenant.ID,
		Domain:     request.Domain,
		IsVerified: false, // New domains start unverified
		IsPrimary:  request.IsPrimary,
	}

	err = s.GetDB().AddTenantDomain(ctx, newDomain)
	if err != nil {
		return errors.WithDetails(errors.ErrInternal, "failed to add tenant domain: %v", err)
	}

	log.Info("Domain added to tenant", "tenant", tenantUID, "domain", request.Domain, "isPrimary", request.IsPrimary)
	return nil
}

// UpdateTenantDomain updates domain verification or primary status
func (s *tenantsService) UpdateTenantDomain(ctx context.Context, domainID int, updates domain.TenantDomainUpdates) error {
	// Convert domain updates to ports updates
	portsUpdates := ports.TenantDomainUpdates{
		IsVerified: updates.IsVerified,
		IsPrimary:  updates.IsPrimary,
	}

	// Update the domain in the database
	err := s.GetDB().UpdateTenantDomain(ctx, domainID, portsUpdates)
	if err != nil {
		return errors.WithDetails(errors.ErrInternal, "failed to update tenant domain: %v", err)
	}

	log.Info("Tenant domain updated", "domainID", domainID)
	return nil
}

// RemoveTenantDomain removes a domain from a tenant
func (s *tenantsService) RemoveTenantDomain(ctx context.Context, domainID int) error {
	err := s.GetDB().RemoveTenantDomain(ctx, domainID)
	if err != nil {
		return errors.WithDetails(errors.ErrInternal, "failed to remove tenant domain: %v", err)
	}

	log.Info("Tenant domain removed", "domainID", domainID)
	return nil
}

// DeleteTenant deletes a tenant from the database
func (s *tenantsService) DeleteTenant(ctx context.Context, tenantUID string) error {
	err := s.GetDB().DeleteTenant(ctx, tenantUID)
	if err != nil {
		return errors.WithDetails(errors.ErrInternal, "failed to delete tenant: %v", err)
	}

	log.Info("Tenant deleted", "tenantUID", tenantUID)
	return nil
}

// Close releases all resources held by the service
func (s *tenantsService) Close() {
	// no-op
}

// createDefaultTenantRoles creates default roles for a new tenant
func (s *tenantsService) createDefaultTenantRoles(ctx context.Context, tenantID int) error {
	// Define default tenant-specific roles
	defaultRoles := []struct {
		name        string
		description string
	}{
		{"admin", "Tenant administrator with full access to tenant resources"},
		{"researcher", "Researcher with access to studies, cases, and annotations"},
		{"viewer", "Read-only access to tenant content"},
	}

	// Create each role for this tenant
	for _, role := range defaultRoles {
		roleID, err := s.GetDB().CreateTenantRoleIfNotExists(ctx, tenantID, role.name, role.description)
		if err != nil {
			log.Error("Failed to create tenant role", "tenant_id", tenantID, "role", role.name, "error", err)
			// Continue with other roles even if one fails
			continue
		}

		// Assign appropriate permissions to each role
		if err := s.assignPermissionsToTenantRole(ctx, roleID, role.name); err != nil {
			log.Error("Failed to assign permissions to tenant role", "tenant_id", tenantID, "role", role.name, "error", err)
		}

		log.Debug("Created tenant role", "tenant_id", tenantID, "role", role.name, "role_id", roleID)
	}

	return nil
}

// assignPermissionsToTenantRole assigns appropriate permissions to a tenant role
func (s *tenantsService) assignPermissionsToTenantRole(ctx context.Context, roleID int, roleName string) error {
	// Get all permissions
	allPermissions, err := s.GetDB().GetAllPermissions(ctx)
	if err != nil {
		return err
	}

	// Define permission sets for each role
	var permissionNames []string
	switch roleName {
	case "admin":
		// Admin gets all permissions except system-level ones
		for _, perm := range allPermissions {
			if !strings.HasPrefix(perm.Name, "system.") && !strings.HasPrefix(perm.Name, "tenants.") {
				permissionNames = append(permissionNames, perm.Name)
			}
		}
	case "researcher":
		// Researcher gets read/write access to studies, cases, slides, annotations
		permissionNames = []string{
			"studies.view", "studies.create", "studies.edit", "studies.add_case", "studies.modify_case", "studies.annotate_case",
			"cases.view", "cases.create", "cases.edit", "cases.add_slide", "cases.modify_slide",
			"slides.view", "slides.create", "slides.edit", "slides.annotate",
			"annotations.view", "annotations.create", "annotations.edit",
			"users.view",
		}
	case "viewer":
		// Viewer gets read-only access
		permissionNames = []string{
			"studies.view",
			"cases.view",
			"slides.view",
			"annotations.view",
		}
	}

	// Assign permissions to the role
	for _, permName := range permissionNames {
		for _, perm := range allPermissions {
			if perm.Name == permName {
				err := s.GetDB().AssignPermissionToRole(ctx, roleID, perm.ID)
				if err != nil {
					log.Warn("Failed to assign permission to role", "role_id", roleID, "permission", permName, "error", err)
				}
				break
			}
		}
	}

	return nil
}

// createDefaultTenantPermissions creates default permissions for a new tenant
func (s *tenantsService) createDefaultTenantPermissions(ctx context.Context, tenantID int) error {
	// Define default tenant-specific permissions
	defaultPermissions := []struct {
		name        string
		description string
	}{
		{"tenant.admin", "Tenant administrator with full access to tenant resources"},
		{"tenant.researcher", "Researcher with access to studies, cases, and annotations"},
		{"tenant.viewer", "Read-only access to tenant content"},
	}

	// Create each permission for this tenant
	for _, perm := range defaultPermissions {
		permissionID, err := s.GetDB().CreateTenantPermissionIfNotExists(ctx, tenantID, perm.name, perm.description)
		if err != nil {
			log.Error("Failed to create tenant permission", "tenant_id", tenantID, "permission", perm.name, "error", err)
			// Continue with other permissions even if one fails
			continue
		}

		log.Debug("Created tenant permission", "tenant_id", tenantID, "permission", perm.name, "permission_id", permissionID)
	}

	return nil
}

// getUserIDFromContext extracts the user ID from the context if available
func (s *tenantsService) getUserIDFromContext(ctx context.Context) int {
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		// No authentication context available (e.g., during system initialization)
		return 0
	}
	return authCtx.CreatorID
}

// createDefaultEmailTemplates creates default email templates for a new tenant
func (s *tenantsService) createDefaultEmailTemplates(ctx context.Context, tenantID int, createdByUserID int) error {
	err := s.GetDB().CreateDefaultTemplates(ctx, tenantID, createdByUserID)
	if err != nil {
		return fmt.Errorf("failed to create default email templates: %w", err)
	}

	log.Info("Created default email templates for tenant", "tenant_id", tenantID, "created_by", createdByUserID)
	return nil
}
