// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package sqlite

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Tenant-related methods that delegate to the tenants adapter

// CreateTenant adds a new tenant to the database
func (db *DB) CreateTenant(ctx context.Context, newTenant ports.NewTenant) error {
	return db.tenants.CreateTenant(ctx, newTenant)
}

// TenantExists checks if a tenant with the given ID already exists
func (db *DB) TenantExists(ctx context.Context, tenantUID string) (bool, error) {
	return db.tenants.TenantExists(ctx, tenantUID)
}

// GetTenantByUID retrieves a specific tenant by its ID
func (db *DB) GetTenantByUID(ctx context.Context, tenantUID string) (ports.Tenant, error) {
	return db.tenants.GetTenantByUID(ctx, tenantUID)
}

// LoadAllTenants retrieves tenants from the database with search/filter and pagination support
func (db *DB) LoadAllTenants(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Tenant, error) {
	return db.tenants.LoadAllTenants(ctx, search, pagination)
}

// GetTenantsCount retrieves the total count of tenants in the database
func (db *DB) GetTenantsCount(ctx context.Context) (int, error) {
	return db.tenants.GetTenantsCount(ctx)
}

// GetTenantsCountWithSearch returns the total count of tenants matching search criteria
func (db *DB) GetTenantsCountWithSearch(ctx context.Context, search utils.SearchParams) (int, error) {
	return db.tenants.GetTenantsCountWithSearch(ctx, search)
}

// UpdateTenant updates tenant information for a tenant with the specified UID
func (db *DB) UpdateTenant(ctx context.Context, tenantUID string, updates ports.TenantUpdates) error {
	return db.tenants.UpdateTenant(ctx, tenantUID, updates)
}

// DeleteTenant deletes a tenant by UID
func (db *DB) DeleteTenant(ctx context.Context, tenantUID string) error {
	return db.tenants.DeleteTenant(ctx, tenantUID)
}

// Domain-related methods

// GetTenantByDomain retrieves a tenant by email domain
func (db *DB) GetTenantByDomain(ctx context.Context, domain string) (ports.Tenant, error) {
	return db.tenants.GetTenantByDomain(ctx, domain)
}

// AddTenantDomain adds a domain to a tenant
func (db *DB) AddTenantDomain(ctx context.Context, domain ports.NewTenantDomain) error {
	return db.tenants.AddTenantDomain(ctx, domain)
}

// GetTenantDomains retrieves all domains for a tenant
func (db *DB) GetTenantDomains(ctx context.Context, tenantID int) ([]ports.TenantDomain, error) {
	return db.tenants.GetTenantDomains(ctx, tenantID)
}

// UpdateTenantDomain updates a tenant domain
func (db *DB) UpdateTenantDomain(ctx context.Context, domainID int, updates ports.TenantDomainUpdates) error {
	return db.tenants.UpdateTenantDomain(ctx, domainID, updates)
}

// RemoveTenantDomain removes a domain from a tenant
func (db *DB) RemoveTenantDomain(ctx context.Context, domainID int) error {
	return db.tenants.RemoveTenantDomain(ctx, domainID)
}

// DomainExists checks if a domain is already claimed by any tenant
func (db *DB) DomainExists(ctx context.Context, domain string) (bool, error) {
	return db.tenants.DomainExists(ctx, domain)
}
