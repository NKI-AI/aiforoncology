// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package ports

import (
	"context"
	"time"

	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Tenant represents a tenant in the database
type Tenant struct {
	ID          int    // Internal database ID - not exposed in API
	TenantUID   string // Public identifier
	Name        string
	Description string
	Status      string
	CreatedAt   time.Time
}

// NewTenant represents a new tenant to be created to the database
type NewTenant struct {
	TenantUID   string
	Name        string
	Description string
}

// TenantUpdates represents fields that can be updated for an existing tenant
type TenantUpdates struct {
	Name        *string
	Description *string
	Status      *string
}

// TenantDomain represents a domain associated with a tenant
type TenantDomain struct {
	ID         int    // Internal database ID - not exposed in API
	TenantID   int    // Internal reference to tenant - not exposed in API
	Domain     string // Changed from DomainUID to Domain to match adapter expectation
	IsVerified bool
	IsPrimary  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewTenantDomain represents a new domain to be associated with a tenant
type NewTenantDomain struct {
	TenantID   int // Internal reference - used for database operations
	Domain     string
	IsVerified bool
	IsPrimary  bool
}

// TenantDomainUpdates represents fields that can be updated for a tenant domain
type TenantDomainUpdates struct {
	IsVerified *bool
	IsPrimary  *bool
}

// TenantsRepository defines the interface for tenant-related database operations
type TenantsRepository interface {
	// LoadAllTenants retrieves tenants from the database with search/filter and pagination support
	LoadAllTenants(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]Tenant, error)

	// GetTenantsCount retrieves the total count of tenants in the database
	GetTenantsCount(ctx context.Context) (int, error)

	// GetTenantsCountWithSearch returns the total count of tenants matching search criteria
	GetTenantsCountWithSearch(ctx context.Context, search utils.SearchParams) (int, error)

	// CreateTenant adds a new tenant to the database.
	CreateTenant(ctx context.Context, newTenant NewTenant) error

	// TenantExists checks if a tenant with the given ID already exists.
	TenantExists(ctx context.Context, tenantUID string) (bool, error)

	// GetTenantByUID retrieves a specific tenant by its ID.
	GetTenantByUID(ctx context.Context, tenantUID string) (Tenant, error)

	// UpdateTenant updates tenant information for a tenant with the specified UID.
	UpdateTenant(ctx context.Context, tenantUID string, updates TenantUpdates) error

	// DeleteTenant deletes a tenant by UID.
	DeleteTenant(ctx context.Context, tenantUID string) error

	// Domain-related methods

	// GetTenantByDomain retrieves a tenant by email domain
	GetTenantByDomain(ctx context.Context, domain string) (Tenant, error)

	// AddTenantDomain adds a domain to a tenant
	AddTenantDomain(ctx context.Context, domain NewTenantDomain) error

	// GetTenantDomains retrieves all domains for a tenant
	GetTenantDomains(ctx context.Context, tenantID int) ([]TenantDomain, error)

	// UpdateTenantDomain updates a tenant domain
	UpdateTenantDomain(ctx context.Context, domainID int, updates TenantDomainUpdates) error

	// RemoveTenantDomain removes a domain from a tenant
	RemoveTenantDomain(ctx context.Context, domainID int) error

	// DomainExists checks if a domain is already claimed by any tenant
	DomainExists(ctx context.Context, domain string) (bool, error)
}
