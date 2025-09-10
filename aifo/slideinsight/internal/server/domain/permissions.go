// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

package domain

// Permission represents a permission in the domain layer
type Permission struct {
	TenantID    int    `json:"tenantId"` // 0 = system tenant, >0 = regular tenant
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// PermissionsResponse represents the response for paginated permissions
type PermissionsResponse struct {
	Permissions []Permission   `json:"permissions"`
	Pagination  PaginationInfo `json:"pagination"`
}

// PermissionExplanation represents a detailed explanation of why a user has access to a resource
type PermissionExplanation struct {
	UserUID          string                   `json:"userUid"`
	Permission       string                   `json:"permission"`
	ResourceType     string                   `json:"resourceType"`
	ResourceUID      string                   `json:"resourceUid"`
	HasAccess        bool                     `json:"hasAccess"`
	GrantType        string                   `json:"grantType,omitempty"`        // "role_based_grant", "direct_object_grant", "inherited_grant", "access_denied"
	InheritancePath  string                   `json:"inheritancePath,omitempty"`  // e.g., "slide->case->study" for inherited access
	GrantingResource *PermissionGrantResource `json:"grantingResource,omitempty"` // The resource that actually grants the permission
	ChecksPerformed  []PermissionCheck        `json:"checksPerformed"`            // All permission checks that were performed
	Message          string                   `json:"message"`                    // Human-readable explanation
}

// PermissionGrantResource represents the resource that grants the permission
type PermissionGrantResource struct {
	ResourceType string `json:"resourceType"` // "role", "study", "case", "slide"
	ResourceUID  string `json:"resourceUid,omitempty"`
	ResourceName string `json:"resourceName,omitempty"`
}

// PermissionCheck represents an individual permission check that was performed
type PermissionCheck struct {
	CheckType      string                   `json:"checkType"` // "role_based", "direct_object", "inherited_object"
	ResourceType   string                   `json:"resourceType,omitempty"`
	ResourceUID    string                   `json:"resourceUid,omitempty"`
	ResourceName   string                   `json:"resourceName,omitempty"`
	Result         bool                     `json:"result"`                   // Whether this specific check passed
	Description    string                   `json:"description"`              // Human-readable description of what was checked
	GrantingEntity *PermissionGrantResource `json:"grantingEntity,omitempty"` // What granted this permission (role, direct grant, etc.)
}
