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

	"aifo.dev/aifo/slideinsight/internal/server/ports"

	"github.com/gofiber/fiber/v2/log"
)

// ObjectGrant represents an object grant in the domain
type ObjectGrant struct {
	ID           int                    `json:"id"`
	GranteeType  string                 `json:"grantee_type"`
	GranteeID    int                    `json:"grantee_id"`
	GranteeName  string                 `json:"grantee_name,omitempty"` // For display purposes
	Permission   string                 `json:"permission"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   int                    `json:"resource_id"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
	GranteeInfo  map[string]interface{} `json:"grantee_info,omitempty"` // Additional grantee information
}

// ObjectGrantService interface defines methods for managing object grants
type ObjectGrantService interface {
	CreateObjectGrant(ctx context.Context, granteeType string, granteeID *int, granteeUID, permission, resourceType string, resourceID *int, resourceUID string) error
	GetObjectGrants(ctx context.Context, resourceType string, resourceID int) ([]ObjectGrant, error)
	GetObjectGrantsByUID(ctx context.Context, resourceType string, resourceUID string) ([]ObjectGrant, error)
	DeleteObjectGrant(ctx context.Context, granteeType string, granteeID int, permission, resourceType string, resourceID int) error
}

type objectGrantService struct {
	*BaseService
	db ports.Database
}

// NewObjectGrantService creates a new object grant service
func NewObjectGrantService(db ports.Database) ObjectGrantService {
	return &objectGrantService{
		BaseService: NewBaseService(db),
		db:          db,
	}
}

// CreateObjectGrant creates a new object grant
func (s *objectGrantService) CreateObjectGrant(ctx context.Context, granteeType string, granteeID *int, granteeUID, permission, resourceType string, resourceID *int, resourceUID string) error {
	// Resolve resource ID if UID is provided
	resolvedResourceID := 0
	if resourceID != nil {
		resolvedResourceID = *resourceID
	} else if resourceUID != "" {
		var err error
		resolvedResourceID, err = s.resolveResourceUID(ctx, resourceType, resourceUID)
		if err != nil {
			return fmt.Errorf("failed to resolve resource UID: %w", err)
		}
	} else {
		return fmt.Errorf("either resource_id or resource_uid must be provided")
	}

	// Check if the resource exists first
	if err := s.validateResource(ctx, resourceType, resolvedResourceID); err != nil {
		return fmt.Errorf("resource validation failed: %w", err)
	}

	// Resolve grantee ID if UID is provided
	resolvedGranteeID := 0
	if granteeID != nil {
		resolvedGranteeID = *granteeID
	} else if granteeUID != "" {
		var err error
		resolvedGranteeID, err = s.resolveGranteeUID(ctx, granteeType, granteeUID)
		if err != nil {
			return fmt.Errorf("failed to resolve grantee UID: %w", err)
		}
	} else {
		return fmt.Errorf("either grantee_id or grantee_uid must be provided")
	}

	// Check if the grantee exists
	if err := s.validateGrantee(ctx, granteeType, resolvedGranteeID); err != nil {
		return fmt.Errorf("grantee validation failed: %w", err)
	}

	// Create the grant
	err := s.db.CreateObjectGrant(ctx, granteeType, resolvedGranteeID, permission, resourceType, resolvedResourceID)
	if err != nil {
		log.Error("Failed to create object grant", "error", err)
		return fmt.Errorf("failed to create object grant: %w", err)
	}

	log.Info("Object grant created", "granteeType", granteeType, "granteeID", resolvedGranteeID, "permission", permission, "resourceType", resourceType, "resourceID", resolvedResourceID)
	return nil
}

// GetObjectGrants retrieves all grants for a specific resource
func (s *objectGrantService) GetObjectGrants(ctx context.Context, resourceType string, resourceID int) ([]ObjectGrant, error) {
	// Check if the resource exists first
	if err := s.validateResource(ctx, resourceType, resourceID); err != nil {
		return nil, fmt.Errorf("resource validation failed: %w", err)
	}

	dbGrants, err := s.db.GetObjectGrants(ctx, resourceType, resourceID)
	if err != nil {
		log.Error("Failed to get object grants", "error", err)
		return nil, fmt.Errorf("failed to get object grants: %w", err)
	}

	// Convert database grants to domain grants with additional information
	grants := make([]ObjectGrant, len(dbGrants))
	for i, dbGrant := range dbGrants {
		grants[i] = ObjectGrant{
			ID:           dbGrant.ID,
			GranteeType:  dbGrant.GranteeType,
			GranteeID:    dbGrant.GranteeID,
			Permission:   dbGrant.Permission,
			ResourceType: dbGrant.ResourceType,
			ResourceID:   dbGrant.ResourceID,
			CreatedAt:    dbGrant.CreatedAt,
			UpdatedAt:    dbGrant.UpdatedAt,
		}

		// Add grantee information for display purposes
		if granteeInfo, err := s.getGranteeInfo(ctx, dbGrant.GranteeType, dbGrant.GranteeID); err == nil {
			grants[i].GranteeName = granteeInfo["name"].(string)
			grants[i].GranteeInfo = granteeInfo
		}
	}

	return grants, nil
}

// GetObjectGrantsByUID retrieves all grants for a specific resource by UID
func (s *objectGrantService) GetObjectGrantsByUID(ctx context.Context, resourceType string, resourceUID string) ([]ObjectGrant, error) {
	// Resolve resource UID to internal ID
	resourceID, err := s.resolveResourceUID(ctx, resourceType, resourceUID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve resource UID: %w", err)
	}

	// Use the existing GetObjectGrants method
	return s.GetObjectGrants(ctx, resourceType, resourceID)
}

// DeleteObjectGrant removes an object grant
func (s *objectGrantService) DeleteObjectGrant(ctx context.Context, granteeType string, granteeID int, permission, resourceType string, resourceID int) error {
	err := s.db.DeleteObjectGrant(ctx, granteeType, granteeID, permission, resourceType, resourceID)
	if err != nil {
		log.Error("Failed to delete object grant", "error", err)
		return fmt.Errorf("failed to delete object grant: %w", err)
	}

	log.Info("Object grant deleted", "granteeType", granteeType, "granteeID", granteeID, "permission", permission, "resourceType", resourceType, "resourceID", resourceID)
	return nil
}

// validateResource checks if the resource exists
func (s *objectGrantService) validateResource(ctx context.Context, resourceType string, resourceID int) error {
	switch resourceType {
	case "study":
		// Convert internal ID to study UID for validation
		study, err := s.db.GetStudyByInternalID(ctx, resourceID)
		if err != nil {
			return fmt.Errorf("study with ID %d not found", resourceID)
		}
		log.Debug("Validated study resource", "studyID", resourceID, "studyUID", study.StudyUID)
	case "case":
		// Add case validation if needed
		log.Debug("Case validation not implemented yet", "caseID", resourceID)
	case "slide":
		// Add slide validation if needed
		log.Debug("Slide validation not implemented yet", "slideID", resourceID)
	default:
		return fmt.Errorf("invalid resource type: %s", resourceType)
	}
	return nil
}

// validateGrantee checks if the grantee exists
func (s *objectGrantService) validateGrantee(ctx context.Context, granteeType string, granteeID int) error {
	switch granteeType {
	case "user":
		user, err := s.db.GetUserByInternalID(ctx, granteeID)
		if err != nil {
			return fmt.Errorf("user with ID %d not found", granteeID)
		}
		log.Debug("Validated user grantee", "userID", granteeID, "email", user.Email)
	case "group":
		// Add group validation if needed
		log.Debug("Group validation not implemented yet", "groupID", granteeID)
	case "role":
		// Add role validation if needed
		log.Debug("Role validation not implemented yet", "roleID", granteeID)
	default:
		return fmt.Errorf("invalid grantee type: %s", granteeType)
	}
	return nil
}

// getGranteeInfo retrieves display information for a grantee
func (s *objectGrantService) getGranteeInfo(ctx context.Context, granteeType string, granteeID int) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	switch granteeType {
	case "user":
		user, err := s.db.GetUserByInternalID(ctx, granteeID)
		if err != nil {
			log.Warn("Failed to get user info for grantee", "userID", granteeID, "error", err)
			info["name"] = fmt.Sprintf("User %d", granteeID)
			return info, nil
		}
		info["name"] = user.Email
		info["email"] = user.Email
		info["first_name"] = user.FirstName
		info["last_name"] = user.LastName
	case "group":
		// Add group info retrieval if needed
		info["name"] = fmt.Sprintf("Group %d", granteeID)
	case "role":
		// Add role info retrieval if needed
		info["name"] = fmt.Sprintf("Role %d", granteeID)
	}

	return info, nil
}

// resolveGranteeUID converts a grantee UID to internal ID
func (s *objectGrantService) resolveGranteeUID(ctx context.Context, granteeType, granteeUID string) (int, error) {
	switch granteeType {
	case "user":
		user, err := s.db.GetUserByUID(ctx, granteeUID)
		if err != nil {
			return 0, fmt.Errorf("user with UID %s not found", granteeUID)
		}
		log.Debug("Resolved user UID to internal ID", "userUID", granteeUID, "userID", user.ID)
		return user.ID, nil
	case "group":
		// Add group UID resolution if needed
		return 0, fmt.Errorf("group UID resolution not implemented yet")
	case "role":
		// Add role UID resolution if needed
		return 0, fmt.Errorf("role UID resolution not implemented yet")
	default:
		return 0, fmt.Errorf("invalid grantee type: %s", granteeType)
	}
}

// resolveResourceUID converts a resource UID to internal ID
func (s *objectGrantService) resolveResourceUID(ctx context.Context, resourceType, resourceUID string) (int, error) {
	switch resourceType {
	case "study":
		study, err := s.db.GetStudyByUID(ctx, resourceUID)
		if err != nil {
			return 0, fmt.Errorf("study with UID %s not found", resourceUID)
		}
		log.Debug("Resolved study UID to internal ID", "studyUID", resourceUID, "studyID", study.ID)
		return study.ID, nil
	case "case":
		// Add case UID resolution if needed
		return 0, fmt.Errorf("case UID resolution not implemented yet")
	case "slide":
		// Add slide UID resolution if needed
		return 0, fmt.Errorf("slide UID resolution not implemented yet")
	default:
		return 0, fmt.Errorf("invalid resource type: %s", resourceType)
	}
}
