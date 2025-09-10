// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"strconv"

	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"aifo.dev/aifo/slideinsight/internal/server/validation"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// Object grant input structures
type CreateObjectGrantInput struct {
	GranteeType  string `json:"grantee_type" validate:"required,oneof=user group role"`
	GranteeID    *int   `json:"grantee_id,omitempty" validate:"omitempty,min=1"`
	GranteeUID   string `json:"grantee_uid,omitempty"`
	Permission   string `json:"permission" validate:"required"`
	ResourceType string `json:"resource_type" validate:"required,oneof=study case slide"`
	ResourceID   *int   `json:"resource_id,omitempty" validate:"omitempty,min=1"`
	ResourceUID  string `json:"resource_uid,omitempty"`
}

type ObjectGrantParams struct {
	ResourceType string `params:"resource_type" validate:"required,oneof=study case slide"`
	ResourceID   string `params:"resource_id" validate:"required"`
}

type DeleteObjectGrantInput struct {
	GranteeType string `json:"grantee_type" validate:"required,oneof=user group role"`
	GranteeID   int    `json:"grantee_id" validate:"required,min=1"`
	Permission  string `json:"permission" validate:"required"`
}

// CreateObjectGrant creates a new object grant
// @Summary Create object grant
// @Description Create a new permission grant for a specific resource
// @Tags object-grants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body CreateObjectGrantInput true "Object grant data"
// @Success 201 {object} fiber.Map "Grant created successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/object-grants [post]
func CreateObjectGrant(service services.ObjectGrantService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input CreateObjectGrantInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("CreateObjectGrant request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		err := service.CreateObjectGrant(c.UserContext(), input.GranteeType, input.GranteeID, input.GranteeUID, input.Permission, input.ResourceType, input.ResourceID, input.ResourceUID)
		if err != nil {
			log.Error("CreateObjectGrant failed", "error", err)
			return middleware.SendError(c, fiber.StatusInternalServerError, "failed to create object grant")
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(fiber.Map{
			"message": "Object grant created successfully",
		})
	}
}

// GetObjectGrants retrieves all grants for a specific resource
// @Summary Get object grants for resource
// @Description Retrieve all permission grants for a specific resource
// @Tags object-grants
// @Produce json
// @Security BearerAuth
// @Param resource_type path string true "Resource type" Enums(study, case, slide)
// @Param resource_id path string true "Resource ID or UID"
// @Success 200 {array} domain.ObjectGrant "List of object grants"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/object-grants/{resource_type}/{resource_id} [get]
func GetObjectGrants(service services.ObjectGrantService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params ObjectGrantParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetObjectGrants request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		// Try to parse as integer first, if it fails assume it's a UID
		resourceID, err := strconv.Atoi(params.ResourceID)
		if err != nil {
			// Assume it's a UID and use the new service method
			grants, err := service.GetObjectGrantsByUID(c.UserContext(), params.ResourceType, params.ResourceID)
			if err != nil {
				log.Error("GetObjectGrantsByUID failed", "error", err)
				return middleware.SendError(c, fiber.StatusInternalServerError, "failed to retrieve object grants")
			}
			return c.JSON(grants)
		}

		// Use the existing method for integer IDs
		grants, err := service.GetObjectGrants(c.UserContext(), params.ResourceType, resourceID)
		if err != nil {
			log.Error("GetObjectGrants failed", "error", err)
			return middleware.SendError(c, fiber.StatusInternalServerError, "failed to retrieve object grants")
		}

		return c.JSON(grants)
	}
}

// DeleteObjectGrant removes an object grant
// @Summary Delete object grant
// @Description Remove a permission grant for a specific resource
// @Tags object-grants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param resource_type path string true "Resource type" Enums(study, case, slide)
// @Param resource_id path int true "Resource ID"
// @Param input body DeleteObjectGrantInput true "Grant deletion data"
// @Success 200 {object} fiber.Map "Grant deleted successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/object-grants/{resource_type}/{resource_id} [delete]
func DeleteObjectGrant(service services.ObjectGrantService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params ObjectGrantParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("DeleteObjectGrant params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		resourceID, err := strconv.Atoi(params.ResourceID)
		if err != nil {
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid resource ID")
		}

		var input DeleteObjectGrantInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("DeleteObjectGrant body parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		err = service.DeleteObjectGrant(c.UserContext(), input.GranteeType, input.GranteeID, input.Permission, params.ResourceType, resourceID)
		if err != nil {
			log.Error("DeleteObjectGrant failed", "error", err)
			return middleware.SendError(c, fiber.StatusInternalServerError, "failed to delete object grant")
		}

		return c.JSON(fiber.Map{
			"message": "Object grant deleted successfully",
		})
	}
}
