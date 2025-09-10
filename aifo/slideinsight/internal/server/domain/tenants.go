// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

package domain

type Tenant struct {
	TenantUID   string `json:"tenantUid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
}

type TenantUpdates struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type TenantsResponse struct {
	Tenants    []Tenant       `json:"tenants"`
	Pagination PaginationInfo `json:"pagination"`
}

// TenantDomain represents a domain associated with a tenant
type TenantDomain struct {
	ID         int    `json:"id"`
	Domain     string `json:"domain"`
	IsVerified bool   `json:"isVerified"`
	IsPrimary  bool   `json:"isPrimary"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

// TenantDomainsResponse represents the response for tenant domains
type TenantDomainsResponse struct {
	Domains []TenantDomain `json:"domains"`
}

// NewTenantDomainRequest represents a request to add a domain to a tenant
type NewTenantDomainRequest struct {
	Domain    string `json:"domain" validate:"required"`
	IsPrimary bool   `json:"isPrimary"`
}

// TenantDomainUpdates represents updates to a tenant domain
type TenantDomainUpdates struct {
	IsVerified *bool `json:"isVerified,omitempty"`
	IsPrimary  *bool `json:"isPrimary,omitempty"`
}
