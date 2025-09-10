// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package tenants

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
)

// Adapter provides a unified interface for all tenant operations
type Adapter struct {
	db *sql.DB
}

// NewAdapter creates a new tenants adapter
func NewAdapter(db *sql.DB) *Adapter {
	return &Adapter{db: db}
}

// buildTenantsSearchWhereClause builds WHERE conditions and arguments for tenants search queries
func (a *Adapter) buildTenantsSearchWhereClause(search utils.SearchParams, searchableFields []string) ([]string, []interface{}) {
	var whereConditions []string
	var args []interface{}

	// General search across specified searchable fields
	if search.Query != "" {
		// Build LIKE conditions for each searchable field
		var fieldConditions []string
		for _, field := range searchableFields {
			switch field {
			case "name":
				fieldConditions = append(fieldConditions, "name LIKE ?")
				args = append(args, "%"+search.Query+"%")
			case "short_uid":
				fieldConditions = append(fieldConditions, "short_uid LIKE ?")
				args = append(args, "%"+search.Query+"%")
			case "description":
				fieldConditions = append(fieldConditions, "description LIKE ?")
				args = append(args, "%"+search.Query+"%")
			}
		}
		if len(fieldConditions) > 0 {
			whereConditions = append(whereConditions, "("+strings.Join(fieldConditions, " OR ")+")")
		}
	}

	// Add specific field filters
	for field, value := range search.Filters {
		switch field {
		case "name":
			whereConditions = append(whereConditions, "name LIKE ?")
			args = append(args, "%"+value+"%")
		case "description":
			whereConditions = append(whereConditions, "description LIKE ?")
			args = append(args, "%"+value+"%")
		case "status":
			whereConditions = append(whereConditions, "status = ?")
			args = append(args, value)
		default:
			// Skip unknown filter fields
			continue
		}
	}

	return whereConditions, args
}

// validateSortDir ensures sort direction is valid to prevent SQL injection
func (a *Adapter) validateSortDir(sortDir string) string {
	dir := strings.ToUpper(sortDir)
	if dir != "ASC" && dir != "DESC" {
		return "ASC"
	}
	return dir
}

// CreateTenant adds a new tenant to the database
func (a *Adapter) CreateTenant(ctx context.Context, newTenant ports.NewTenant) error {
	_, err := a.db.Exec("INSERT INTO tenants (short_uid, name, description) VALUES (?, ?, ?)",
		newTenant.TenantUID, newTenant.Name, newTenant.Description)
	if err != nil {
		return errors.NewDatabaseInsertError("tenant", err)
	}
	return nil
}

// TenantExists checks if a tenant with the given ID already exists
func (a *Adapter) TenantExists(ctx context.Context, tenantUID string) (bool, error) {
	var exists bool
	err := a.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tenants WHERE short_uid = ?)", tenantUID).Scan(&exists)
	if err != nil {
		return false, errors.NewDatabaseQueryError("tenant existence check", err)
	}
	return exists, nil
}

// GetTenantByUID retrieves a specific tenant by its ID
func (a *Adapter) GetTenantByUID(ctx context.Context, tenantUID string) (ports.Tenant, error) {
	var tenant ports.Tenant
	var createdAtStr string
	err := a.db.QueryRow("SELECT id, short_uid, name, description, status, created_at FROM tenants WHERE short_uid = ?", tenantUID).Scan(
		&tenant.ID, &tenant.TenantUID, &tenant.Name, &tenant.Description, &tenant.Status, &createdAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.Tenant{}, errors.NewTenantNotFoundError(tenantUID)
		}
		return ports.Tenant{}, errors.NewDatabaseQueryError("tenant", err)
	}

	// Parse the created_at timestamp
	if createdAtStr != "" {
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			log.Error("failed to parse created_at timestamp", "error", err)
		} else {
			tenant.CreatedAt = createdAt
		}
	}

	return tenant, nil
}

// LoadAllTenants retrieves tenants from the database with search/filter and pagination support
func (a *Adapter) LoadAllTenants(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Tenant, error) {
	baseQuery := "SELECT id, short_uid, name, description, status, created_at FROM tenants"

	// Build WHERE clause based on search parameters
	searchableFields := []string{"name", "short_uid", "description"} // Include description in searchable fields
	whereConditions, args := a.buildTenantsSearchWhereClause(search, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Add ordering
	orderBy := "created_at DESC" // Default ordering
	if search.HasSort() {
		safeSortDir := a.validateSortDir(search.SortDir)
		switch search.SortBy {
		case "name":
			orderBy = "name " + safeSortDir
		case "created_at", "createdAt":
			orderBy = "created_at " + safeSortDir
		case "short_uid", "shortId":
			orderBy = "short_uid " + safeSortDir
		case "description":
			orderBy = "description " + safeSortDir
		default:
			// Keep default ordering for unknown sort fields
		}
	}
	baseQuery += " ORDER BY " + orderBy

	// Add pagination
	if pagination.Limit > 0 {
		offset := pagination.CalculateOffset()
		baseQuery += " LIMIT ? OFFSET ?"
		args = append(args, pagination.Limit, offset)
	}

	rows, err := a.db.Query(baseQuery, args...)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("tenants", err)
	}
	defer rows.Close()

	var tenants []ports.Tenant
	for rows.Next() {
		var tenant ports.Tenant
		var createdAtStr string
		if err := rows.Scan(&tenant.ID, &tenant.TenantUID, &tenant.Name, &tenant.Description, &tenant.Status, &createdAtStr); err != nil {
			return nil, errors.NewDatabaseScanError("tenant", err)
		}

		// Parse the created_at timestamp
		if createdAtStr != "" {
			createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
			if err != nil {
				// Try alternative format if the first one fails
				createdAt, err = time.Parse(time.RFC3339, createdAtStr)
				if err != nil {
					log.Error("failed to parse created_at timestamp", "error", err)
				} else {
					tenant.CreatedAt = createdAt
				}
			} else {
				tenant.CreatedAt = createdAt
			}
		}

		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("tenants", err)
	}

	return tenants, nil
}

// GetTenantsCount retrieves the total count of tenants in the database
func (a *Adapter) GetTenantsCount(ctx context.Context) (int, error) {
	var count int
	err := a.db.QueryRow("SELECT COUNT(*) FROM tenants").Scan(&count)
	if err != nil {
		return 0, errors.NewDatabaseQueryError("tenant count", err)
	}
	return count, nil
}

// GetTenantsCountWithSearch returns the total count of tenants matching search criteria
func (a *Adapter) GetTenantsCountWithSearch(ctx context.Context, search utils.SearchParams) (int, error) {
	baseQuery := "SELECT COUNT(*) FROM tenants"

	// Build WHERE clause based on search parameters (same logic as LoadAllTenants)
	searchableFields := []string{"name", "short_uid", "description"} // Include description in searchable fields
	whereConditions, args := a.buildTenantsSearchWhereClause(search, searchableFields)

	// Add WHERE conditions to query
	if len(whereConditions) > 0 {
		baseQuery += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	var count int
	err := a.db.QueryRow(baseQuery, args...).Scan(&count)
	if err != nil {
		return 0, errors.NewDatabaseQueryError("tenant count with search", err)
	}
	return count, nil
}

// UpdateTenant updates tenant information for a tenant with the specified UID
func (a *Adapter) UpdateTenant(ctx context.Context, tenantUID string, updates ports.TenantUpdates) error {
	var setParts []string
	var args []interface{}

	// Build SET clause dynamically based on provided updates
	if updates.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *updates.Name)
	}

	if updates.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *updates.Description)
	}

	if updates.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *updates.Status)
	}

	if len(setParts) == 0 {
		return errors.ErrNoFieldsToUpdate
	}

	// Add tenantUID to args for WHERE clause
	args = append(args, tenantUID)

	query := fmt.Sprintf("UPDATE tenants SET %s WHERE short_uid = ?", strings.Join(setParts, ", "))
	result, err := a.db.Exec(query, args...)
	if err != nil {
		return errors.NewDatabaseUpdateError("tenant", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.NewTenantNotFoundError(tenantUID)
	}

	return nil
}

// DeleteTenant deletes a tenant by UID
func (a *Adapter) DeleteTenant(ctx context.Context, tenantUID string) error {
	result, err := a.db.Exec("DELETE FROM tenants WHERE short_uid = ?", tenantUID)
	if err != nil {
		return errors.NewDatabaseDeleteError("tenant", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.NewTenantNotFoundError(tenantUID)
	}

	return nil
}

// Domain-related methods

// GetTenantByDomain retrieves a tenant by email domain
func (a *Adapter) GetTenantByDomain(ctx context.Context, domain string) (ports.Tenant, error) {
	var tenant ports.Tenant
	var createdAtStr string

	// Join tenants with tenant_domains to find the tenant for this domain
	query := `
		SELECT t.id, t.short_uid, t.name, t.description, t.status, t.created_at 
		FROM tenants t 
		INNER JOIN tenant_domains td ON t.id = td.tenant_id 
		WHERE td.domain = ? AND td.is_verified = 1
	`

	err := a.db.QueryRow(query, domain).Scan(
		&tenant.ID, &tenant.TenantUID, &tenant.Name, &tenant.Description, &tenant.Status, &createdAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return ports.Tenant{}, errors.WithDetails(errors.ErrTenantNotFound, "no verified tenant found for domain '%s'", domain)
		}
		return ports.Tenant{}, errors.NewDatabaseQueryError("tenant by domain", err)
	}

	// Parse the created_at timestamp
	if createdAtStr != "" {
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			log.Error("failed to parse created_at timestamp", "error", err)
		} else {
			tenant.CreatedAt = createdAt
		}
	}

	return tenant, nil
}

// AddTenantDomain adds a domain to a tenant
func (a *Adapter) AddTenantDomain(ctx context.Context, domain ports.NewTenantDomain) error {
	_, err := a.db.Exec(
		"INSERT INTO tenant_domains (tenant_id, domain, is_verified, is_primary) VALUES (?, ?, ?, ?)",
		domain.TenantID, domain.Domain, domain.IsVerified, domain.IsPrimary)
	if err != nil {
		return errors.NewDatabaseInsertError("tenant domain", err)
	}
	return nil
}

// GetTenantDomains retrieves all domains for a tenant
func (a *Adapter) GetTenantDomains(ctx context.Context, tenantID int) ([]ports.TenantDomain, error) {
	rows, err := a.db.Query(
		"SELECT id, tenant_id, domain, is_verified, is_primary, created_at, updated_at FROM tenant_domains WHERE tenant_id = ? ORDER BY is_primary DESC, domain ASC",
		tenantID)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("tenant domains", err)
	}
	defer rows.Close()

	var domains []ports.TenantDomain
	for rows.Next() {
		var domain ports.TenantDomain
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(&domain.ID, &domain.TenantID, &domain.Domain, &domain.IsVerified,
			&domain.IsPrimary, &createdAtStr, &updatedAtStr); err != nil {
			return nil, errors.NewDatabaseScanError("tenant domain", err)
		}

		// Parse timestamps
		if createdAtStr != "" {
			if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
				domain.CreatedAt = createdAt
			}
		}
		if updatedAtStr != "" {
			if updatedAt, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
				domain.UpdatedAt = updatedAt
			}
		}

		domains = append(domains, domain)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("tenant domains", err)
	}

	return domains, nil
}

// UpdateTenantDomain updates a tenant domain
func (a *Adapter) UpdateTenantDomain(ctx context.Context, domainID int, updates ports.TenantDomainUpdates) error {
	var setParts []string
	var args []interface{}

	// Build SET clause dynamically based on provided updates
	if updates.IsVerified != nil {
		setParts = append(setParts, "is_verified = ?")
		args = append(args, *updates.IsVerified)
	}

	if updates.IsPrimary != nil {
		setParts = append(setParts, "is_primary = ?")
		args = append(args, *updates.IsPrimary)
	}

	if len(setParts) == 0 {
		return errors.ErrNoFieldsToUpdate
	}

	// Add domainID to args for WHERE clause
	args = append(args, domainID)

	query := fmt.Sprintf("UPDATE tenant_domains SET %s WHERE id = ?", strings.Join(setParts, ", "))
	result, err := a.db.Exec(query, args...)
	if err != nil {
		return errors.NewDatabaseUpdateError("tenant domain", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.NewDomainNotFoundError(domainID)
	}

	return nil
}

// RemoveTenantDomain removes a domain from a tenant
func (a *Adapter) RemoveTenantDomain(ctx context.Context, domainID int) error {
	result, err := a.db.Exec("DELETE FROM tenant_domains WHERE id = ?", domainID)
	if err != nil {
		return errors.NewDatabaseDeleteError("tenant domain", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseCheckRowsError(err)
	}

	if rowsAffected == 0 {
		return errors.NewDomainNotFoundError(domainID)
	}

	return nil
}

// DomainExists checks if a domain is already claimed by any tenant
func (a *Adapter) DomainExists(ctx context.Context, domain string) (bool, error) {
	var exists bool
	err := a.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tenant_domains WHERE domain = ?)", domain).Scan(&exists)
	if err != nil {
		return false, errors.NewDatabaseQueryError("domain existence check", err)
	}
	return exists, nil
}
