// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

package domain

// Role represents a role in the domain layer
type Role struct {
	Name        string `json:"name"`
	TenantID    int    `json:"tenant_id"` // 0 = system tenant, >0 = regular tenant
	ShortUID    string `json:"short_uid"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// RolesResponse represents the response for paginated roles
type RolesResponse struct {
	Roles      []Role         `json:"roles"`
	Pagination PaginationInfo `json:"pagination"`
}
