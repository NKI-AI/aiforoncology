// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
)

// PermissionDB defines the minimal interface needed for permission checking
// This allows for easier testing while the real implementation uses ports.Database
type PermissionDB interface {
	UserHasRolePermission(ctx context.Context, userID int, permission string) (bool, error)
	HasObjectGrant(ctx context.Context, userID int, permission, resourceType string, resourceID int) (bool, error)
	GetCaseStudyRelations(ctx context.Context, caseID int) ([]int, error)
	GetStudyByInternalID(ctx context.Context, studyID int) (ports.Study, error)
}

// SharedPermissionChecker provides centralized permission checking logic
// to ensure consistency across all entity types (studies, cases, slides)
type SharedPermissionChecker struct {
	db PermissionDB
}

// NewSharedPermissionChecker creates a new shared permission checker
func NewSharedPermissionChecker(db PermissionDB) *SharedPermissionChecker {
	return &SharedPermissionChecker{db: db}
}

// CanViewStudy checks if a user can view a specific study
// This is the canonical implementation that should be used everywhere
func (p *SharedPermissionChecker) CanViewStudy(ctx context.Context, userID int, study ports.Study) (bool, error) {
	// First check if user has general role permission to view studies
	allowed, err := p.db.UserHasRolePermission(ctx, userID, "studies.view")
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}

	// Then check if user has specific object grant for this study
	allowed, err = p.db.HasObjectGrant(ctx, userID, "studies.view", "study", study.ID)
	if err != nil {
		return false, err
	}
	return allowed, nil
}

// CanViewCase checks if a user can view a specific case
// This follows the permission inheritance pattern: role -> direct case -> parent study
func (p *SharedPermissionChecker) CanViewCase(ctx context.Context, userID int, case_ ports.Case) (bool, error) {
	// Step 1: Check role permission for cases first
	allowed, err := p.db.UserHasRolePermission(ctx, userID, "cases.view")
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}

	// Step 2: Check direct case permission
	allowed, err = p.db.HasObjectGrant(ctx, userID, "cases.view", "case", case_.ID)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}

	// Step 3: Check inherited permission from parent studies
	// This is the key fix - we need to check for BOTH cases.view AND studies.view on parent studies
	studyIDs, err := p.db.GetCaseStudyRelations(ctx, case_.ID)
	if err != nil {
		return false, err
	}

	for _, studyID := range studyIDs {
		// Check for cases.view permission on parent study (allows case-specific access)
		allowed, err = p.db.HasObjectGrant(ctx, userID, "cases.view", "study", studyID)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}

		// Check for studies.view permission on parent study (allows general study access)
		// This is the critical missing piece that causes the discrepancy
		allowed, err = p.db.HasObjectGrant(ctx, userID, "studies.view", "study", studyID)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}

	return false, nil
}

// CanViewCaseByStudyAccess is a convenience method that checks if a user can view cases
// in a study based on their study access permissions. This is useful for bulk operations.
func (p *SharedPermissionChecker) CanViewCaseByStudyAccess(ctx context.Context, userID int, case_ ports.Case) (bool, error) {
	// First check direct case permissions (same as CanViewCase)
	allowed, err := p.db.UserHasRolePermission(ctx, userID, "cases.view")
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}

	allowed, err = p.db.HasObjectGrant(ctx, userID, "cases.view", "case", case_.ID)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}

	// Then check if user has access to any parent study
	studyIDs, err := p.db.GetCaseStudyRelations(ctx, case_.ID)
	if err != nil {
		return false, err
	}

	for _, studyID := range studyIDs {
		// Get the study record to use the canonical study permission check
		study, err := p.db.GetStudyByInternalID(ctx, studyID)
		if err != nil {
			continue // Skip this study if we can't load it
		}

		// Use the canonical study permission check
		allowed, err = p.CanViewStudy(ctx, userID, study)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}

	return false, nil
}
