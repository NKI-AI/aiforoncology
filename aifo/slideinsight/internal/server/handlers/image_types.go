// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/middleware"
	"aifo.dev/aifo/slideinsight/internal/server/services"
	"aifo.dev/aifo/slideinsight/internal/server/validation"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// Input structs for image types
type ImageTypeInput struct {
	ID                string `json:"id,omitempty"`
	TypeUID           string `json:"typeUid" validate:"required"`
	Name              string `json:"name" validate:"required"`
	Description       string `json:"description,omitempty"`
	Category          string `json:"category" validate:"required,oneof=brightfield fluorescence other"`
	RequiresHistogram bool   `json:"requiresHistogram"`
	MetadataSchema    string `json:"metadataSchema,omitempty"`
	IsActive          bool   `json:"isActive"`
}

type ImageTypeUpdateInput struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	Category          *string `json:"category,omitempty" validate:"omitempty,oneof=brightfield fluorescence other"`
	RequiresHistogram *bool   `json:"requiresHistogram,omitempty"`
	MetadataSchema    *string `json:"metadataSchema,omitempty"`
	IsActive          *bool   `json:"isActive,omitempty"`
}

// Input structs for histograms
type SlideHistogramInput struct {
	ChannelIndex int     `json:"channelIndex"`
	ChannelName  string  `json:"channelName,omitempty"`
	BinCount     int     `json:"binCount" validate:"required,min=1"`
	MinValue     float64 `json:"minValue"`
	MaxValue     float64 `json:"maxValue"`
	Counts       []int   `json:"counts,omitempty"`
	Metadata     string  `json:"metadata,omitempty"`
}

// Input structs for staining protocols
type StainingProtocolInput struct {
	StainName      string `json:"stainName" validate:"required"`
	StainType      string `json:"stainType" validate:"required,oneof=primary counterstain fluorophore other"`
	Concentration  string `json:"concentration,omitempty"`
	IncubationTime string `json:"incubationTime,omitempty"`
	AntibodyInfo   string `json:"antibodyInfo,omitempty"`
	ExcitationNm   *int   `json:"excitationNm,omitempty"`
	EmissionNm     *int   `json:"emissionNm,omitempty"`
	Metadata       string `json:"metadata,omitempty"`
}

type StainingProtocolUpdateInput struct {
	StainName      *string `json:"stainName,omitempty"`
	StainType      *string `json:"stainType,omitempty" validate:"omitempty,oneof=primary counterstain fluorophore other"`
	Concentration  *string `json:"concentration,omitempty"`
	IncubationTime *string `json:"incubationTime,omitempty"`
	AntibodyInfo   *string `json:"antibodyInfo,omitempty"`
	ExcitationNm   *int    `json:"excitationNm,omitempty"`
	EmissionNm     *int    `json:"emissionNm,omitempty"`
	Metadata       *string `json:"metadata,omitempty"`
}

// Parameter structs
type ImageTypeIDParams struct {
	ID string `params:"id" validate:"required"`
}

type HistogramIDParams struct {
	ID string `params:"id" validate:"required"`
}

type ProtocolIDParams struct {
	ID string `params:"id" validate:"required"`
}

// ImageTypes handlers

// GetImageTypes returns a handler function that retrieves all image types
func GetImageTypes(service services.ImageTypesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		params, err := utils.ParsePaginationAndSearchParams(c)
		if err != nil {
			return err
		}

		imageTypes, paginationInfo, err := service.GetImageTypes(c.UserContext(), params)
		if err != nil {
			log.Error("GetImageTypes failed", "error", err)
			return middleware.HandleError(c, err)
		}

		return c.JSON(domain.ImageTypesResponse{
			ImageTypes: imageTypes,
			Pagination: paginationInfo,
		})
	}
}

// GetImageType returns a handler function that retrieves a specific image type by ID
func GetImageType(service services.ImageTypesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params ImageTypeIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetImageType request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err
		}

		imageType, err := service.GetImageTypeByID(c.UserContext(), params.ID)
		if err != nil {
			log.Error("GetImageType failed", "error", err, "id", params.ID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(imageType)
	}
}

// CreateImageType returns a handler function that creates a new image type
func CreateImageType(service services.ImageTypesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input ImageTypeInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("CreateImageType request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err
		}

		imageType := domain.ImageType{
			ID:                input.ID,
			TypeUID:           input.TypeUID,
			Name:              input.Name,
			Description:       input.Description,
			Category:          input.Category,
			RequiresHistogram: input.RequiresHistogram,
			MetadataSchema:    input.MetadataSchema,
			IsActive:          input.IsActive,
		}

		createdImageType, err := service.CreateImageType(c.UserContext(), imageType)
		if err != nil {
			log.Error("CreateImageType failed", "error", err)
			return middleware.HandleError(c, err)
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(createdImageType)
	}
}

// UpdateImageType returns a handler function that updates an image type
func UpdateImageType(service services.ImageTypesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params ImageTypeIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("UpdateImageType request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err
		}

		var input ImageTypeUpdateInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("UpdateImageType request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err
		}

		// Convert input to map for the service
		updates := make(map[string]interface{})
		if input.Name != nil {
			updates["name"] = *input.Name
		}
		if input.Description != nil {
			updates["description"] = *input.Description
		}
		if input.Category != nil {
			updates["category"] = *input.Category
		}
		if input.RequiresHistogram != nil {
			updates["requiresHistogram"] = *input.RequiresHistogram
		}
		if input.MetadataSchema != nil {
			updates["metadataSchema"] = *input.MetadataSchema
		}
		if input.IsActive != nil {
			updates["isActive"] = *input.IsActive
		}

		err := service.UpdateImageType(c.UserContext(), params.ID, updates)
		if err != nil {
			log.Error("UpdateImageType failed", "error", err, "id", params.ID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{"message": "Image type updated successfully"})
	}
}

// DeleteImageType returns a handler function that deletes an image type
func DeleteImageType(service services.ImageTypesService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params ImageTypeIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("DeleteImageType request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err
		}

		err := service.DeleteImageType(c.UserContext(), params.ID)
		if err != nil {
			log.Error("DeleteImageType failed", "error", err, "id", params.ID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{"message": "Image type deleted successfully"})
	}
}

// Slide Histograms handlers

// GetSlideHistogram returns a handler function that retrieves histogram data for a slide
func GetSlideHistogram(service services.SlideHistogramsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params SlideUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetSlideHistogram request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err
		}

		histograms, err := service.GetHistogramsBySlideUID(c.UserContext(), params.SlideUID)
		if err != nil {
			log.Error("GetSlideHistogram failed", "error", err, "slideUID", params.SlideUID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(domain.SlideHistogramResponse{
			SlideUID:   params.SlideUID,
			Histograms: histograms,
		})
	}
}

// CreateSlideHistogram returns a handler function that creates histogram data for a slide
func CreateSlideHistogram(service services.SlideHistogramsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params SlideUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("CreateSlideHistogram request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err
		}

		var input SlideHistogramInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("CreateSlideHistogram request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err
		}

		histogram := domain.SlideHistogram{
			ChannelIndex: input.ChannelIndex,
			ChannelName:  input.ChannelName,
			BinCount:     input.BinCount,
			MinValue:     input.MinValue,
			MaxValue:     input.MaxValue,
			Counts:       input.Counts,
			Metadata:     input.Metadata,
		}

		createdHistogram, err := service.CreateHistogram(c.UserContext(), params.SlideUID, histogram)
		if err != nil {
			log.Error("CreateSlideHistogram failed", "error", err)
			return middleware.HandleError(c, err)
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(createdHistogram)
	}
}

// DeleteSlideHistogram returns a handler function that deletes histogram data for a slide
func DeleteSlideHistogram(service services.SlideHistogramsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params SlideUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("DeleteSlideHistogram request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err
		}

		err := service.DeleteHistogramsBySlideUID(c.UserContext(), params.SlideUID)
		if err != nil {
			log.Error("DeleteSlideHistogram failed", "error", err, "slideUID", params.SlideUID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{"message": "Slide histograms deleted successfully"})
	}
}

// Staining Protocols handlers

// GetSlideStainingProtocols returns a handler function that retrieves staining protocols for a slide
func GetSlideStainingProtocols(service services.StainingProtocolsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params SlideUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetSlideStainingProtocols request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err
		}

		protocols, err := service.GetProtocolsBySlideUID(c.UserContext(), params.SlideUID)
		if err != nil {
			log.Error("GetSlideStainingProtocols failed", "error", err, "slideUID", params.SlideUID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(domain.StainingProtocolsResponse{
			SlideUID:          params.SlideUID,
			StainingProtocols: protocols,
		})
	}
}

// GetStainingProtocol returns a handler function that retrieves a specific staining protocol
func GetStainingProtocol(service services.StainingProtocolsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params ProtocolIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("GetStainingProtocol request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err
		}

		protocol, err := service.GetProtocolByID(c.UserContext(), params.ID)
		if err != nil {
			log.Error("GetStainingProtocol failed", "error", err, "id", params.ID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(protocol)
	}
}

// CreateStainingProtocol returns a handler function that creates a staining protocol for a slide
func CreateStainingProtocol(service services.StainingProtocolsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params SlideUIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("CreateStainingProtocol request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err
		}

		var input StainingProtocolInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("CreateStainingProtocol request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err
		}

		protocol := domain.StainingProtocol{
			StainName:      input.StainName,
			StainType:      input.StainType,
			Concentration:  input.Concentration,
			IncubationTime: input.IncubationTime,
			AntibodyInfo:   input.AntibodyInfo,
			ExcitationNm:   input.ExcitationNm,
			EmissionNm:     input.EmissionNm,
			Metadata:       input.Metadata,
		}

		createdProtocol, err := service.CreateProtocol(c.UserContext(), params.SlideUID, protocol)
		if err != nil {
			log.Error("CreateStainingProtocol failed", "error", err)
			return middleware.HandleError(c, err)
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(createdProtocol)
	}
}

// UpdateStainingProtocol returns a handler function that updates a staining protocol
func UpdateStainingProtocol(service services.StainingProtocolsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params ProtocolIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("UpdateStainingProtocol request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err
		}

		var input StainingProtocolUpdateInput
		if err := c.BodyParser(&input); err != nil {
			log.Warn("UpdateStainingProtocol request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid request")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, input); err != nil {
			return err
		}

		// Convert input to map for the service
		updates := make(map[string]interface{})
		if input.StainName != nil {
			updates["stainName"] = *input.StainName
		}
		if input.StainType != nil {
			updates["stainType"] = *input.StainType
		}
		if input.Concentration != nil {
			updates["concentration"] = *input.Concentration
		}
		if input.IncubationTime != nil {
			updates["incubationTime"] = *input.IncubationTime
		}
		if input.AntibodyInfo != nil {
			updates["antibodyInfo"] = *input.AntibodyInfo
		}
		if input.ExcitationNm != nil {
			updates["excitationNm"] = *input.ExcitationNm
		}
		if input.EmissionNm != nil {
			updates["emissionNm"] = *input.EmissionNm
		}
		if input.Metadata != nil {
			updates["metadata"] = *input.Metadata
		}

		err := service.UpdateProtocol(c.UserContext(), params.ID, updates)
		if err != nil {
			log.Error("UpdateStainingProtocol failed", "error", err, "id", params.ID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{"message": "Staining protocol updated successfully"})
	}
}

// DeleteStainingProtocol returns a handler function that deletes a staining protocol
func DeleteStainingProtocol(service services.StainingProtocolsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var params ProtocolIDParams
		if err := c.ParamsParser(&params); err != nil {
			log.Warn("DeleteStainingProtocol request parsing failed", "error", err)
			return middleware.SendError(c, fiber.StatusBadRequest, "invalid parameters")
		}

		if err := validation.GlobalValidator.ValidateStruct(c, params); err != nil {
			return err
		}

		err := service.DeleteProtocol(c.UserContext(), params.ID)
		if err != nil {
			log.Error("DeleteStainingProtocol failed", "error", err, "id", params.ID)
			return middleware.HandleError(c, err)
		}

		return c.JSON(fiber.Map{"message": "Staining protocol deleted successfully"})
	}
}
