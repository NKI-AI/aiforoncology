// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package middleware

import (
	"context"
	"fmt"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/gofiber/fiber/v2"

	// "github.com/gofiber/fiber/v2/log"
	"go.uber.org/zap"
)

// RequirePermission creates middleware that checks if the user has the specified permission
// on the resource identified by the route parameters, with inheritance support.
//
// Inheritance logic:
// - For slides: check slide permission, then parent case, then parent study
// - For cases: check case permission, then parent study(ies)
// - For studies: check study permission only
//
// Parameters:
// - permission: the permission to check (e.g., "studies.view", "cases.view", "slides.view")
// - resourceType: the primary resource type ("study", "case", "slide")
// - resourceParam: the route parameter name containing the resource UID (e.g., "studyUID", "caseUID", "slideUID")
func RequirePermission(db ports.Database, permission, resourceType, resourceParam string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := zap.L()

		// Get the principal from the JWT middleware
		principal := c.Locals("principal")
		if principal == nil {
			logger.Warn("Permission check attempted without authentication")
			return SendError(c, fiber.StatusUnauthorized, "authentication required")
		}

		p, ok := principal.(*Principal)
		if !ok || p == nil {
			logger.Warn("Invalid principal object in permission check")
			return SendError(c, fiber.StatusUnauthorized, "invalid authentication")
		}

		// Get the user ID - we need to fetch it from the database since it's not in the JWT
		userID, err := getUserIDFromPrincipal(c.UserContext(), db, p)
		if err != nil {
			logger.Error("Failed to get user ID for permission check",
				zap.Error(err),
				zap.String("userUID", p.UserUID))
			return SendError(c, fiber.StatusInternalServerError, "permission check failed")
		}

		// Get the resource UID from route parameters
		resourceUID := c.Params(resourceParam)
		if resourceUID == "" {
			logger.Warn("Permission check attempted without resource identifier",
				zap.String("param", resourceParam))
			return SendError(c, fiber.StatusBadRequest, fmt.Sprintf("%s is required", resourceParam))
		}

		logger.Info("Starting permission check cascade",
			zap.Int("userID", userID),
			zap.String("userUID", p.UserUID),
			zap.String("permission", permission),
			zap.String("resourceType", resourceType),
			zap.String("resourceUID", resourceUID),
			zap.String("resourceParam", resourceParam))

		// Check permission with inheritance
		hasPermission, err := checkPermissionWithInheritance(c.UserContext(), db, userID, permission, resourceType, resourceUID)
		if err != nil {
			logger.Error("Permission check cascade failed with error",
				zap.Error(err),
				zap.Int("userID", userID),
				zap.String("permission", permission),
				zap.String("resourceType", resourceType),
				zap.String("resourceUID", resourceUID))
			return SendError(c, fiber.StatusInternalServerError, "permission check failed")
		}

		if !hasPermission {
			logger.Warn("Permission check cascade completed - all checks failed, access denied",
				zap.Int("userID", userID),
				zap.String("userUID", p.UserUID),
				zap.String("permission", permission),
				zap.String("resourceType", resourceType),
				zap.String("resourceUID", resourceUID))
			return SendError(c, fiber.StatusForbidden, "insufficient permissions")
		}

		logger.Info("Permission check cascade completed - access granted",
			zap.Int("userID", userID),
			zap.String("userUID", p.UserUID),
			zap.String("permission", permission),
			zap.String("resourceType", resourceType),
			zap.String("resourceUID", resourceUID))

		return c.Next()
	}
}

// getUserIDFromPrincipal gets the internal user ID from the database using the principal's UserUID
func getUserIDFromPrincipal(ctx context.Context, db ports.Database, p *Principal) (int, error) {
	// Use the database method to get user ID by UID
	return db.GetUserIDByUID(ctx, p.UserUID)
}

// checkPermissionWithInheritance checks permission with inheritance logic
func checkPermissionWithInheritance(ctx context.Context, db ports.Database, userID int, permission, resourceType, resourceUID string) (bool, error) {
	logger := zap.L()

	logger.Info("Beginning permission inheritance check",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.String("resourceType", resourceType),
		zap.String("resourceUID", resourceUID))

	switch resourceType {
	case "study":
		logger.Info("Permission check cascade: role-based -> study-only (no inheritance)",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.String("resourceUID", resourceUID))
		return checkStudyPermission(ctx, db, userID, permission, resourceUID)
	case "case":
		logger.Info("Permission check cascade: role-based -> case -> study inheritance",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.String("resourceUID", resourceUID))
		return checkCasePermission(ctx, db, userID, permission, resourceUID)
	case "slide":
		logger.Info("Permission check cascade: role-based -> slide -> case -> study inheritance",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.String("resourceUID", resourceUID))
		return checkSlidePermission(ctx, db, userID, permission, resourceUID)
	default:
		logger.Error("Unsupported resource type in permission check",
			zap.String("resourceType", resourceType),
			zap.Int("userID", userID),
			zap.String("permission", permission))
		return false, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// checkStudyPermission checks permission for a study
func checkStudyPermission(ctx context.Context, db ports.Database, userID int, permission, studyUID string) (bool, error) {
	logger := zap.L()

	logger.Info("Checking permission cascade step 1/2: role-based permission (global precedence)",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.String("studyUID", studyUID))

	// Step 1: Check role-based permission first (takes precedence)
	hasRolePermission, err := db.UserHasRolePermission(ctx, userID, permission)
	if err != nil {
		logger.Error("Failed to check role-based permission",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.Error(err))
		return false, fmt.Errorf("failed to check role permission: %w", err)
	}

	if hasRolePermission {
		logger.Info("Permission check cascade completed: role-based permission granted - skipping remaining checks as this overrides all object-specific permissions",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.String("studyUID", studyUID),
			zap.String("grantType", "role_based_grant"))
		return true, nil
	}

	logger.Info("Permission check cascade step 1/2 failed: no role-based permission - proceeding to step 2/2: checking object-specific permissions",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.String("studyUID", studyUID))

	// Step 2: Check direct study permission
	logger.Info("Checking permission cascade step 2/2: direct study permission",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.String("studyUID", studyUID))

	studyID, err := db.GetStudyIDByUID(ctx, studyUID)
	if err != nil {
		logger.Error("Failed to resolve study UID to ID",
			zap.String("studyUID", studyUID),
			zap.Error(err))
		return false, fmt.Errorf("failed to get study ID: %w", err)
	}

	logger.Debug("Resolved study UID to ID",
		zap.String("studyUID", studyUID),
		zap.Int("studyID", studyID))

	hasPermission, err := db.HasObjectGrant(ctx, userID, permission, "study", studyID)
	if err != nil {
		logger.Error("Failed to check direct study permission",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.Int("studyID", studyID),
			zap.Error(err))
		return false, err
	}

	if hasPermission {
		logger.Info("Permission check cascade completed: direct study permission granted",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.String("studyUID", studyUID),
			zap.Int("studyID", studyID),
			zap.String("grantType", "direct_study_grant"))
	} else {
		logger.Info("Permission check cascade step 2/2 failed: no direct study permission found - access denied",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.String("studyUID", studyUID),
			zap.Int("studyID", studyID))
	}

	return hasPermission, nil
}

// checkCasePermission checks permission for a case with study inheritance
func checkCasePermission(ctx context.Context, db ports.Database, userID int, permission, caseUID string) (bool, error) {
	logger := zap.L()

	logger.Info("Checking permission cascade step 1/3: role-based permission (global precedence)",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.String("caseUID", caseUID))

	// Step 1: Check role-based permission first (takes precedence)
	hasRolePermission, err := db.UserHasRolePermission(ctx, userID, permission)
	if err != nil {
		logger.Error("Failed to check role-based permission",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.Error(err))
		return false, fmt.Errorf("failed to check role permission: %w", err)
	}

	if hasRolePermission {
		logger.Info("Permission check cascade completed: role-based permission granted - skipping remaining checks as this overrides all object-specific permissions",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.String("caseUID", caseUID),
			zap.String("grantType", "role_based_grant"))
		return true, nil
	}

	logger.Info("Permission check cascade step 1/3 failed: no role-based permission - proceeding to step 2/3: checking object-specific permissions",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.String("caseUID", caseUID))

	// Step 2: Check direct case permission
	logger.Info("Checking permission cascade step 2/3: direct case permission",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.String("caseUID", caseUID))

	caseID, err := db.GetCaseIDByUID(ctx, caseUID)
	if err != nil {
		logger.Error("Failed to resolve case UID to ID",
			zap.String("caseUID", caseUID),
			zap.Error(err))
		return false, fmt.Errorf("failed to get case ID: %w", err)
	}

	logger.Debug("Resolved case UID to ID",
		zap.String("caseUID", caseUID),
		zap.Int("caseID", caseID))

	// First check direct permission on case
	hasPermission, err := db.HasObjectGrant(ctx, userID, permission, "case", caseID)
	if err != nil {
		logger.Error("Failed to check direct case permission",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.Int("caseID", caseID),
			zap.Error(err))
		return false, fmt.Errorf("failed to check case permission: %w", err)
	}
	if hasPermission {
		logger.Info("Permission check cascade completed: direct case permission granted - skipping remaining checks as this overrides inheritance",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.String("caseUID", caseUID),
			zap.Int("caseID", caseID),
			zap.String("grantType", "direct_case_grant"))
		return true, nil
	}

	logger.Info("Permission check cascade step 2/3 failed: no direct case permission - proceeding to step 3/3: checking parent study inheritance",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.Int("caseID", caseID))

	// Step 3: If no direct permission, check parent study(ies)
	studyIDs, err := db.GetCaseStudyRelations(ctx, caseID)
	if err != nil {
		logger.Error("Failed to get case study relations",
			zap.Int("caseID", caseID),
			zap.Error(err))
		return false, fmt.Errorf("failed to get case study relations: %w", err)
	}

	logger.Info("Permission check cascade step 3/3: checking inherited study permissions",
		zap.Int("caseID", caseID),
		zap.Ints("studyIDs", studyIDs),
		zap.Int("studyCount", len(studyIDs)))

	// Check permission on each parent study
	for i, studyID := range studyIDs {
		logger.Info("Checking inherited permission on parent study",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.Int("studyID", studyID),
			zap.Int("caseID", caseID),
			zap.Int("studyIndex", i+1),
			zap.Int("totalStudies", len(studyIDs)))

		hasPermission, err := db.HasObjectGrant(ctx, userID, permission, "study", studyID)
		if err != nil {
			logger.Error("Failed to check inherited study permission",
				zap.Int("userID", userID),
				zap.String("permission", permission),
				zap.Int("studyID", studyID),
				zap.Int("caseID", caseID),
				zap.Error(err))
			return false, fmt.Errorf("failed to check study permission: %w", err)
		}
		if hasPermission {
			logger.Info("Permission check cascade completed: inherited study permission granted - skipping remaining study checks as this overrides further inheritance",
				zap.Int("userID", userID),
				zap.String("permission", permission),
				zap.String("caseUID", caseUID),
				zap.Int("caseID", caseID),
				zap.Int("parentStudyID", studyID),
				zap.String("grantType", "inherited_study_grant"),
				zap.String("inheritancePath", "case->study"))
			return true, nil
		}

		logger.Info("Inherited permission check failed on parent study - continuing to next study if available",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.Int("studyID", studyID),
			zap.Int("caseID", caseID),
			zap.Int("studyIndex", i+1),
			zap.Int("totalStudies", len(studyIDs)))
	}

	logger.Info("Permission check cascade step 3/3 failed: no inherited study permissions found - access denied",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.String("caseUID", caseUID),
		zap.Int("caseID", caseID),
		zap.Ints("checkedStudyIDs", studyIDs))

	return false, nil
}

// checkSlidePermission checks permission for a slide with case and study inheritance
func checkSlidePermission(ctx context.Context, db ports.Database, userID int, permission, slideUID string) (bool, error) {
	logger := zap.L()

	logger.Info("Checking permission cascade step 1/4: role-based permission (global precedence)",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.String("slideUID", slideUID))

	// Step 1: Check role-based permission first (takes precedence)
	hasRolePermission, err := db.UserHasRolePermission(ctx, userID, permission)
	if err != nil {
		logger.Error("Failed to check role-based permission",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.Error(err))
		return false, fmt.Errorf("failed to check role permission: %w", err)
	}

	if hasRolePermission {
		logger.Info("Permission check cascade completed: role-based permission granted - skipping remaining checks as this overrides all object-specific permissions",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.String("slideUID", slideUID),
			zap.String("grantType", "role_based_grant"))
		return true, nil
	}

	logger.Info("Permission check cascade step 1/4 failed: no role-based permission - proceeding to step 2/4: checking object-specific permissions",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.String("slideUID", slideUID))

	// Step 2: Check direct slide permission
	logger.Info("Checking permission cascade step 2/4: direct slide permission",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.String("slideUID", slideUID))

	slideID, err := db.GetSlideIDByUID(ctx, slideUID)
	if err != nil {
		logger.Error("Failed to resolve slide UID to ID",
			zap.String("slideUID", slideUID),
			zap.Error(err))
		return false, fmt.Errorf("failed to get slide ID: %w", err)
	}

	logger.Debug("Resolved slide UID to ID",
		zap.String("slideUID", slideUID),
		zap.Int("slideID", slideID))

	// First check direct permission on slide
	hasPermission, err := db.HasObjectGrant(ctx, userID, permission, "slide", slideID)
	if err != nil {
		logger.Error("Failed to check direct slide permission",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.Int("slideID", slideID),
			zap.Error(err))
		return false, fmt.Errorf("failed to check slide permission: %w", err)
	}
	if hasPermission {
		logger.Info("Permission check cascade completed: direct slide permission granted - skipping remaining checks as this overrides inheritance",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.String("slideUID", slideUID),
			zap.Int("slideID", slideID),
			zap.String("grantType", "direct_slide_grant"))
		return true, nil
	}

	logger.Info("Permission check cascade step 2/4 failed: no direct slide permission - proceeding to step 3/4: checking parent case inheritance",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.Int("slideID", slideID))

	// Step 3: If no direct permission, check parent case
	caseID, err := db.GetSlideCaseRelation(ctx, slideID)
	if err != nil {
		logger.Error("Failed to get slide case relation",
			zap.Int("slideID", slideID),
			zap.Error(err))
		return false, fmt.Errorf("failed to get slide case relation: %w", err)
	}

	logger.Info("Permission check cascade step 3/4: checking inherited case permission",
		zap.Int("slideID", slideID),
		zap.Int("caseID", caseID))

	hasPermission, err = db.HasObjectGrant(ctx, userID, permission, "case", caseID)
	if err != nil {
		logger.Error("Failed to check inherited case permission",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.Int("caseID", caseID),
			zap.Int("slideID", slideID),
			zap.Error(err))
		return false, fmt.Errorf("failed to check case permission: %w", err)
	}
	if hasPermission {
		logger.Info("Permission check cascade completed: inherited case permission granted - skipping remaining checks as this overrides further inheritance",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.String("slideUID", slideUID),
			zap.Int("slideID", slideID),
			zap.Int("parentCaseID", caseID),
			zap.String("grantType", "inherited_case_grant"),
			zap.String("inheritancePath", "slide->case"))
		return true, nil
	}

	logger.Info("Permission check cascade step 3/4 failed: no inherited case permission - proceeding to step 4/4: checking parent study inheritance",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.Int("caseID", caseID),
		zap.Int("slideID", slideID))

	// Step 4: If no permission on case, check parent study(ies)
	studyIDs, err := db.GetCaseStudyRelations(ctx, caseID)
	if err != nil {
		logger.Error("Failed to get case study relations",
			zap.Int("caseID", caseID),
			zap.Int("slideID", slideID),
			zap.Error(err))
		return false, fmt.Errorf("failed to get case study relations: %w", err)
	}

	logger.Info("Permission check cascade step 4/4: checking inherited study permissions",
		zap.Int("caseID", caseID),
		zap.Int("slideID", slideID),
		zap.Ints("studyIDs", studyIDs),
		zap.Int("studyCount", len(studyIDs)))

	// Check permission on each parent study
	for i, studyID := range studyIDs {
		logger.Info("Checking inherited permission on parent study",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.Int("studyID", studyID),
			zap.Int("caseID", caseID),
			zap.Int("slideID", slideID),
			zap.Int("studyIndex", i+1),
			zap.Int("totalStudies", len(studyIDs)))

		hasPermission, err := db.HasObjectGrant(ctx, userID, permission, "study", studyID)
		if err != nil {
			logger.Error("Failed to check inherited study permission",
				zap.Int("userID", userID),
				zap.String("permission", permission),
				zap.Int("studyID", studyID),
				zap.Int("caseID", caseID),
				zap.Int("slideID", slideID),
				zap.Error(err))
			return false, fmt.Errorf("failed to check study permission: %w", err)
		}
		if hasPermission {
			logger.Info("Permission check cascade completed: inherited study permission granted - skipping remaining study checks as this overrides further inheritance",
				zap.Int("userID", userID),
				zap.String("permission", permission),
				zap.String("slideUID", slideUID),
				zap.Int("slideID", slideID),
				zap.Int("parentCaseID", caseID),
				zap.Int("parentStudyID", studyID),
				zap.String("grantType", "inherited_study_grant"),
				zap.String("inheritancePath", "slide->case->study"))
			return true, nil
		}

		logger.Info("Inherited permission check failed on parent study - continuing to next study if available",
			zap.Int("userID", userID),
			zap.String("permission", permission),
			zap.Int("studyID", studyID),
			zap.Int("caseID", caseID),
			zap.Int("slideID", slideID),
			zap.Int("studyIndex", i+1),
			zap.Int("totalStudies", len(studyIDs)))
	}

	logger.Info("Permission check cascade step 4/4 failed: no inherited study permissions found - access denied",
		zap.Int("userID", userID),
		zap.String("permission", permission),
		zap.String("slideUID", slideUID),
		zap.Int("slideID", slideID),
		zap.Int("caseID", caseID),
		zap.Ints("checkedStudyIDs", studyIDs))

	return false, nil
}

// Convenience functions for common permission checks

// RequireStudyView creates middleware that requires "studies.view" permission
func RequireStudyView(db ports.Database) fiber.Handler {
	return RequirePermission(db, "studies.view", "study", "studyUID")
}

// RequireCaseView creates middleware that requires "cases.view" permission
func RequireCaseView(db ports.Database) fiber.Handler {
	return RequirePermission(db, "cases.view", "case", "caseUID")
}

// RequireSlideView creates middleware that requires "slides.view" permission
func RequireSlideView(db ports.Database) fiber.Handler {
	return RequirePermission(db, "slides.view", "slide", "slideUID")
}

// RequireStudyEdit creates middleware that requires "studies.edit" permission
func RequireStudyEdit(db ports.Database) fiber.Handler {
	return RequirePermission(db, "studies.edit", "study", "studyUID")
}

// RequireCaseEdit creates middleware that requires "cases.edit" permission
func RequireCaseEdit(db ports.Database) fiber.Handler {
	return RequirePermission(db, "cases.edit", "case", "caseUID")
}

// RequireSlideEdit creates middleware that requires "slides.edit" permission
func RequireSlideEdit(db ports.Database) fiber.Handler {
	return RequirePermission(db, "slides.edit", "slide", "slideUID")
}
