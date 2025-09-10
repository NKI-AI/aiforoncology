// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"strconv"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"aifo.dev/aifo/slideinsight/internal/server/validation"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// RegionIDParams represents path parameters for region ID requests
type RegionIDParams struct {
	RegionID string `params:"regionID" validate:"required,uuid"`
}

// SlideRegionParams represents path parameters for slide UID + region ID requests
type SlideRegionParams struct {
	SlideUID string `params:"slideUID" validate:"required"`
	RegionID string `params:"regionID" validate:"required,uuid"`
}

// CreateRegionInput represents the input for creating a region
type CreateRegionInput struct {
	RegionName       string                    `json:"regionName" validate:"required"`
	RegionType       string                    `json:"regionType" validate:"required,oneof=roi patient tissue artifact background other"`
	Geometry         domain.RegionGeometry     `json:"geometry" validate:"required"`
	CoordinateSystem string                    `json:"coordinateSystem" validate:"required,oneof=pixel physical"`
	AreaPixels       *int                      `json:"areaPixels,omitempty"`
	AreaPhysical     *float64                  `json:"areaPhysical,omitempty"`
	Labels           *domain.RegionLabels      `json:"labels,omitempty"`
	Metadata         *domain.RegionMetadata    `json:"metadata,omitempty"`
	StyleConfig      *domain.RegionStyleConfig `json:"styleConfig,omitempty"`
	Mutable          *bool                     `json:"mutable,omitempty"`
	Visible          *bool                     `json:"visible,omitempty"`
}

// UpdateRegionInput represents the input for updating a region
type UpdateRegionInput struct {
	RegionName       *string                   `json:"regionName,omitempty"`
	RegionType       *string                   `json:"regionType,omitempty" validate:"omitempty,oneof=roi patient tissue artifact background other"`
	Geometry         *domain.RegionGeometry    `json:"geometry,omitempty"`
	CoordinateSystem *string                   `json:"coordinateSystem,omitempty" validate:"omitempty,oneof=pixel physical"`
	AreaPixels       *int                      `json:"areaPixels,omitempty"`
	AreaPhysical     *float64                  `json:"areaPhysical,omitempty"`
	Labels           *domain.RegionLabels      `json:"labels,omitempty"`
	Metadata         *domain.RegionMetadata    `json:"metadata,omitempty"`
	StyleConfig      *domain.RegionStyleConfig `json:"styleConfig,omitempty"`
	Mutable          *bool                     `json:"mutable,omitempty"`
	Visible          *bool                     `json:"visible,omitempty"`
}

// BulkCreateRegionsInput represents the input for bulk creating regions
type BulkCreateRegionsInput struct {
	Regions []CreateRegionInput `json:"regions" validate:"required,min=1,max=100,dive"`
}

// BulkUpdateRegionsInput represents the input for bulk updating regions
type BulkUpdateRegionsInput struct {
	Updates map[string]UpdateRegionInput `json:"updates" validate:"required,min=1,max=100"`
}

// BulkDeleteRegionsInput represents the input for bulk deleting regions
type BulkDeleteRegionsInput struct {
	RegionIDs []string `json:"regionIds" validate:"required,min=1,max=100,dive,required,uuid"`
}

// GetRegionsBySlideUID retrieves all regions for a specific slide
// @Summary Get regions for a slide
// @Description Retrieve all regions associated with a specific slide
// @Tags regions
// @Produce json
// @Security BearerAuth
// @Param slideUID path string true "Slide UID"
// @Success 200 {object} domain.RegionList "List of regions for the slide"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid slide UID"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUID}/regions [get]
func GetRegionsBySlideUID(service services.RegionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params SlideUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetRegionsBySlideUID params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid slide UID")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		regions, err := service.GetRegionsBySlideUID(c.UserContext(), params.SlideUID)
		if err != nil {
			log.Error("GetRegionsBySlideUID failed", "error", err, "slideUID", params.SlideUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		response := domain.RegionList{
			SlideUID:   params.SlideUID,
			Regions:    regions,
			TotalCount: len(regions),
			HasMore:    false,
			NextCursor: nil,
		}

		return c.JSON(response)
	}
}

// CreateRegion creates a new region on a slide
// @Summary Create a new region
// @Description Create a new region of interest on a specific slide
// @Tags regions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slideUID path string true "Slide UID"
// @Param region body CreateRegionInput true "Region data"
// @Success 201 {object} domain.Region "Created region"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUID}/regions [post]
func CreateRegion(service services.RegionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params SlideUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("CreateRegion params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid slide UID")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		var input CreateRegionInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("CreateRegion request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// Convert input to domain model
		request := domain.CreateRegionRequest{
			RegionName:       input.RegionName,
			SlideUID:         params.SlideUID,
			RegionType:       input.RegionType,
			Geometry:         input.Geometry,
			CoordinateSystem: input.CoordinateSystem,
			AreaPixels:       input.AreaPixels,
			AreaPhysical:     input.AreaPhysical,
			Labels:           input.Labels,
			Metadata:         input.Metadata,
			StyleConfig:      input.StyleConfig,
			Mutable:          input.Mutable,
			Visible:          input.Visible,
		}

		createdRegion, err := service.CreateRegion(c.UserContext(), params.SlideUID, request)
		if err != nil {
			log.Error("CreateRegion failed", "error", err, "slideUID", params.SlideUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(createdRegion)
	}
}

// BulkCreateRegions creates multiple regions on a slide
// @Summary Bulk create regions
// @Description Create multiple regions of interest on a specific slide
// @Tags regions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slideUID path string true "Slide UID"
// @Param regions body BulkCreateRegionsInput true "Regions data"
// @Success 201 {array} domain.Region "Created regions"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUID}/regions/bulk [post]
func BulkCreateRegions(service services.RegionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params SlideUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("BulkCreateRegions params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid slide UID")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		var input BulkCreateRegionsInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("BulkCreateRegions request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// Convert input to domain model
		regions := make([]domain.CreateRegionRequest, len(input.Regions))
		for i, regionInput := range input.Regions {
			regions[i] = domain.CreateRegionRequest{
				RegionName:       regionInput.RegionName,
				SlideUID:         params.SlideUID,
				RegionType:       regionInput.RegionType,
				Geometry:         regionInput.Geometry,
				CoordinateSystem: regionInput.CoordinateSystem,
				AreaPixels:       regionInput.AreaPixels,
				AreaPhysical:     regionInput.AreaPhysical,
				Labels:           regionInput.Labels,
				Metadata:         regionInput.Metadata,
				StyleConfig:      regionInput.StyleConfig,
				Mutable:          regionInput.Mutable,
				Visible:          regionInput.Visible,
			}
		}

		request := domain.BulkCreateRegionsRequest{
			SlideUID: params.SlideUID,
			Regions:  regions,
		}

		createdRegions, err := service.BulkCreateRegions(c.UserContext(), request)
		if err != nil {
			log.Error("BulkCreateRegions failed", "error", err, "slideUID", params.SlideUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(createdRegions)
	}
}

// GetRegionByID retrieves a specific region by its ID
// @Summary Get region by ID
// @Description Retrieve a specific region by its UUID
// @Tags regions
// @Produce json
// @Security BearerAuth
// @Param regionID path string true "Region ID (UUID)"
// @Success 200 {object} domain.Region "Region details"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid region ID"
// @Failure 404 {object} domain.ErrorResponse "Region not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/regions/{regionID} [get]
func GetRegionByID(service services.RegionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params RegionIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetRegionByID params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid region ID")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		region, err := service.GetRegionByID(c.UserContext(), params.RegionID)
		if err != nil {
			log.Error("GetRegionByID failed", "error", err, "regionID", params.RegionID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.JSON(region)
	}
}

// UpdateRegion updates an existing region
// @Summary Update region
// @Description Update an existing region by its UUID
// @Tags regions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param regionID path string true "Region ID (UUID)"
// @Param region body UpdateRegionInput true "Updated region data"
// @Success 200 {object} domain.Region "Updated region"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 404 {object} domain.ErrorResponse "Region not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/regions/{regionID} [put]
func UpdateRegion(service services.RegionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params RegionIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("UpdateRegion params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid region ID")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		var input UpdateRegionInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("UpdateRegion request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// Convert input to domain model
		request := domain.UpdateRegionRequest{
			RegionName:       input.RegionName,
			RegionType:       input.RegionType,
			Geometry:         input.Geometry,
			CoordinateSystem: input.CoordinateSystem,
			AreaPixels:       input.AreaPixels,
			AreaPhysical:     input.AreaPhysical,
			Labels:           input.Labels,
			Metadata:         input.Metadata,
			StyleConfig:      input.StyleConfig,
			Mutable:          input.Mutable,
			Visible:          input.Visible,
		}

		err := service.UpdateRegion(c.UserContext(), params.RegionID, request)
		if err != nil {
			log.Error("UpdateRegion failed", "error", err, "regionID", params.RegionID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		// Return the updated region
		updatedRegion, err := service.GetRegionByID(c.UserContext(), params.RegionID)
		if err != nil {
			log.Error("UpdateRegion - failed to retrieve updated region", "error", err, "regionID", params.RegionID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.JSON(updatedRegion)
	}
}

// DeleteRegion soft-deletes a region
// @Summary Delete region
// @Description Soft delete a region by its UUID
// @Tags regions
// @Security BearerAuth
// @Param regionID path string true "Region ID (UUID)"
// @Success 204 "Region deleted successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid region ID"
// @Failure 404 {object} domain.ErrorResponse "Region not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/regions/{regionID} [delete]
func DeleteRegion(service services.RegionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params RegionIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("DeleteRegion params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid region ID")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		err := service.DeleteRegion(c.UserContext(), params.RegionID)
		if err != nil {
			log.Error("DeleteRegion failed", "error", err, "regionID", params.RegionID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

// BulkUpdateRegions updates multiple regions
// @Summary Bulk update regions
// @Description Update multiple regions in a single request
// @Tags regions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param regions body BulkUpdateRegionsInput true "Regions update data"
// @Success 200 "Regions updated successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/regions/bulk [put]
func BulkUpdateRegions(service services.RegionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input BulkUpdateRegionsInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("BulkUpdateRegions request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		// Convert input to domain model
		updates := make(map[string]domain.UpdateRegionRequest)
		for regionID, updateInput := range input.Updates {
			updates[regionID] = domain.UpdateRegionRequest{
				RegionName:       updateInput.RegionName,
				RegionType:       updateInput.RegionType,
				Geometry:         updateInput.Geometry,
				CoordinateSystem: updateInput.CoordinateSystem,
				AreaPixels:       updateInput.AreaPixels,
				AreaPhysical:     updateInput.AreaPhysical,
				Labels:           updateInput.Labels,
				Metadata:         updateInput.Metadata,
				StyleConfig:      updateInput.StyleConfig,
				Mutable:          updateInput.Mutable,
				Visible:          updateInput.Visible,
			}
		}

		request := domain.BulkUpdateRegionsRequest{
			Updates: updates,
		}

		err := service.BulkUpdateRegions(c.UserContext(), request)
		if err != nil {
			log.Error("BulkUpdateRegions failed", "error", err)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.SendStatus(fiber.StatusOK)
	}
}

// BulkDeleteRegions soft-deletes multiple regions
// @Summary Bulk delete regions
// @Description Soft delete multiple regions in a single request
// @Tags regions
// @Accept json
// @Security BearerAuth
// @Param regions body BulkDeleteRegionsInput true "Region IDs to delete"
// @Success 204 "Regions deleted successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid input"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/regions/bulk [delete]
func BulkDeleteRegions(service services.RegionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input BulkDeleteRegionsInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("BulkDeleteRegions request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		// Validate the parsed input
		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err // Error already formatted and sent
		}

		request := domain.BulkDeleteRegionsRequest{
			RegionIDs: input.RegionIDs,
		}

		err := service.BulkDeleteRegions(c.UserContext(), request)
		if err != nil {
			log.Error("BulkDeleteRegions failed", "error", err)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

// GetRegions retrieves regions with pagination and search
// @Summary Get regions with search
// @Description Retrieve regions with pagination, filtering, and search capabilities
// @Tags regions
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" minimum(1) example(1)
// @Param limit query int false "Items per page (default: 100)" minimum(1) maximum(1000) example(100)
// @Param q query string false "General search across region name and metadata" example("patient")
// @Param slideUID query string false "Filter by slide UID" example("slide-123")
// @Param regionType query string false "Filter by region type" Enums(roi,patient,tissue,artifact,background,other) example("patient")
// @Param visible query bool false "Filter by visibility" example(true)
// @Success 200 {object} domain.RegionList "List of regions with pagination info"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid parameters"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/regions [get]
func GetRegions(service services.RegionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse pagination and search parameters
		params, err := utils.ParsePaginationAndSearchParamsWithConfig(c, utils.DefaultPaginationOptions(), utils.SearchConfig{
			SearchableFields: []string{"name", "labels", "metadata"},
			SortableFields:   []string{"name", "created_at", "updated_at", "region_type"},
		})
		if err != nil {
			return err // Error already formatted and sent
		}

		// Parse region-specific search parameters
		regionSearchParams := ports.RegionSearchParams{}

		if slideUID := c.Query("slideUID"); slideUID != "" {
			regionSearchParams.SlideUID = &slideUID
		}

		if regionType := c.Query("regionType"); regionType != "" {
			regionSearchParams.RegionType = &regionType
		}

		if visibleStr := c.Query("visible"); visibleStr != "" {
			visible, err := strconv.ParseBool(visibleStr)
			if err != nil {
				return middleware.SendError(c, fiber.StatusBadRequest, "invalid visible parameter")
			}
			regionSearchParams.Visible = &visible
		}

		if actorType := c.Query("actorType"); actorType != "" {
			regionSearchParams.ActorType = &actorType
		}

		if coordinateSystem := c.Query("coordinateSystem"); coordinateSystem != "" {
			regionSearchParams.CoordinateSystem = &coordinateSystem
		}

		regions, paginationInfo, err := service.GetRegions(c.UserContext(), params, regionSearchParams)
		if err != nil {
			log.Error("GetRegions failed", "error", err)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		slideUID := ""
		if regionSearchParams.SlideUID != nil {
			slideUID = *regionSearchParams.SlideUID
		}

		response := domain.RegionList{
			SlideUID:   slideUID,
			Regions:    regions,
			TotalCount: paginationInfo.Total,
			HasMore:    paginationInfo.HasNext,
			NextCursor: nil, // Could implement cursor-based pagination if needed
		}

		return c.JSON(response)
	}
}

// GetRegionStatistics returns statistics for regions on a slide
// @Summary Get region statistics
// @Description Get statistics for all regions on a specific slide
// @Tags regions
// @Produce json
// @Security BearerAuth
// @Param slideUID path string true "Slide UID"
// @Success 200 {object} domain.RegionStatistics "Region statistics for the slide"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid slide UID"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUID}/regions/statistics [get]
func GetRegionStatistics(service services.RegionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params SlideUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetRegionStatistics params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid slide UID")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		statistics, err := service.GetRegionStatistics(c.UserContext(), params.SlideUID)
		if err != nil {
			log.Error("GetRegionStatistics failed", "error", err, "slideUID", params.SlideUID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.JSON(statistics)
	}
}

// GetDeletedRegions retrieves all soft-deleted regions
// @Summary Get deleted regions
// @Description Retrieve all soft-deleted regions for administrative purposes
// @Tags regions
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.Region "List of deleted regions"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/regions/deleted [get]
func GetDeletedRegions(service services.RegionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		regions, err := service.GetDeletedRegions(c.UserContext())
		if err != nil {
			log.Error("GetDeletedRegions failed", "error", err)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.JSON(regions)
	}
}

// RestoreRegion restores a soft-deleted region
// @Summary Restore region
// @Description Restore a soft-deleted region by its UUID
// @Tags regions
// @Security BearerAuth
// @Param regionID path string true "Region ID (UUID)"
// @Success 200 {object} domain.Region "Restored region"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid region ID"
// @Failure 404 {object} domain.ErrorResponse "Region not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/regions/{regionID}/restore [post]
func RestoreRegion(service services.RegionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params RegionIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("RestoreRegion params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid region ID")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		err := service.RestoreRegion(c.UserContext(), params.RegionID)
		if err != nil {
			log.Error("RestoreRegion failed", "error", err, "regionID", params.RegionID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		// Return the restored region
		restoredRegion, err := service.GetRegionByID(c.UserContext(), params.RegionID)
		if err != nil {
			log.Error("RestoreRegion - failed to retrieve restored region", "error", err, "regionID", params.RegionID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.JSON(restoredRegion)
	}
}

// DeleteRegionBySlideUID soft-deletes a region by slide UID and region ID
// @Summary Delete region by slide UID
// @Description Soft delete a region by its UUID within a specific slide context
// @Tags regions
// @Security BearerAuth
// @Param slideUID path string true "Slide UID"
// @Param regionID path string true "Region ID (UUID)"
// @Success 204 "Region deleted successfully"
// @Failure 400 {object} domain.ErrorResponse "Bad request - invalid slide UID or region ID"
// @Failure 404 {object} domain.ErrorResponse "Region not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/slides/{slideUID}/regions/{regionID} [delete]
func DeleteRegionBySlideUID(service services.RegionsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params SlideRegionParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("DeleteRegionBySlideUID params parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid slide UID or region ID")
		}

		// Validate the parsed parameters
		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err // Error already formatted and sent
		}

		// First verify that the region belongs to the specified slide
		region, err := service.GetRegionByID(c.UserContext(), params.RegionID)
		if err != nil {
			log.Error("DeleteRegionBySlideUID - failed to get region", "error", err, "regionID", params.RegionID)
			return middleware.SendError(c, fiber.StatusNotFound, "region not found")
		}

		// Verify the region belongs to the specified slide
		if region.SlideUID != params.SlideUID {
			log.Warn("DeleteRegionBySlideUID - region does not belong to slide", "regionID", params.RegionID, "expectedSlideUID", params.SlideUID, "actualSlideUID", region.SlideUID)
			return middleware.SendError(c, fiber.StatusNotFound, "region not found in specified slide")
		}

		// Delete the region
		err = service.DeleteRegion(c.UserContext(), params.RegionID)
		if err != nil {
			log.Error("DeleteRegionBySlideUID failed", "error", err, "slideUID", params.SlideUID, "regionID", params.RegionID)
			return middleware.SendError(c, fiber.StatusInternalServerError, "internal error")
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}
