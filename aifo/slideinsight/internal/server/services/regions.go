// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"
	"encoding/json"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/google/uuid"
)

// RegionsService interface defines region-related operations
type RegionsService interface {
	// Basic CRUD operations
	GetRegions(ctx context.Context, params utils.PaginationAndSearchParams, searchParams ports.RegionSearchParams) ([]domain.Region, domain.PaginationInfo, error)
	GetRegionsBySlideUID(ctx context.Context, slideUID string) ([]domain.Region, error)
	GetRegionByID(ctx context.Context, regionID string) (domain.Region, error)
	CreateRegion(ctx context.Context, slideUID string, request domain.CreateRegionRequest) (domain.Region, error)
	UpdateRegion(ctx context.Context, regionID string, request domain.UpdateRegionRequest) error
	DeleteRegion(ctx context.Context, regionID string) error

	// Bulk operations
	BulkCreateRegions(ctx context.Context, request domain.BulkCreateRegionsRequest) ([]domain.Region, error)
	BulkUpdateRegions(ctx context.Context, request domain.BulkUpdateRegionsRequest) error
	BulkDeleteRegions(ctx context.Context, request domain.BulkDeleteRegionsRequest) error

	// Soft delete management
	GetDeletedRegions(ctx context.Context) ([]domain.Region, error)
	RestoreRegion(ctx context.Context, regionID string) error

	// Statistics
	GetRegionStatistics(ctx context.Context, slideUID string) (domain.RegionStatistics, error)

	Close()
}

type regionsService struct {
	*BaseService
	db ports.RegionsRepository
	// paginatedSearchService *PaginatedSearchService[ports.Region, domain.Region] // TODO: Implement when database layer is ready
}

// NewRegionsService creates a new RegionsService
func NewRegionsService(db ports.RegionsRepository, baseService *BaseService) RegionsService {
	return &regionsService{
		BaseService: baseService,
		db:          db,
		// TODO: Initialize paginatedSearchService when database layer is ready
	}
}

// convertRegionDBToDomain converts a database Region record to a domain Region model
func convertRegionDBToDomain(record ports.Region) domain.Region {
	// Convert the database region to domain region
	region := domain.Region{
		RegionID:         record.ID,
		RegionName:       record.Name,
		SlideUID:         record.SlideUID,
		RegionType:       record.RegionType,
		CoordinateSystem: record.CoordinateSystem,
		AreaPixels:       record.AreaPixels,
		AreaPhysical:     record.AreaPhysical,
		ActorType:        record.ActorType,
		ActorID:          record.ActorID,
		Mutable:          record.Mutable,
		Visible:          record.Visible,
		CreatedAt:        record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        record.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Handle soft deletion timestamps
	if record.DeletedAt != nil {
		deletedAtStr := record.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		region.DeletedAt = &deletedAtStr
	}
	if record.DeletedBy != nil {
		region.DeletedBy = record.DeletedBy
	}

	// Parse JSON fields
	if record.GeometryData != "" {
		var geometry domain.RegionGeometry
		if err := json.Unmarshal([]byte(record.GeometryData), &geometry); err == nil {
			region.Geometry = geometry
		} else {
			// Fallback for malformed JSON
			region.Geometry = domain.RegionGeometry{
				Type:        "Polygon",
				Coordinates: record.GeometryData,
			}
		}
	}

	// TODO: Parse JSON fields properly when implementing full database layer
	// For now, leaving these as nil/empty since they're optional pointers
	// Labels, Metadata, and StyleConfig will be nil by default

	return region
}

// convertDomainToDBRegion converts a domain Region creation request to database NewRegion
func convertDomainToDBRegion(slideUID string, request domain.CreateRegionRequest, authCtx *AuthContext, regionID string) ports.NewRegion {
	// Convert geometry to JSON string
	geometryJSON := ""
	if request.Geometry.Coordinates != nil {
		if geometryBytes, err := json.Marshal(request.Geometry); err == nil {
			geometryJSON = string(geometryBytes)
		}
	}

	// Convert labels to JSON string
	labelsJSON := "{}"
	if request.Labels != nil {
		if labelsBytes, err := json.Marshal(request.Labels); err == nil {
			labelsJSON = string(labelsBytes)
		}
	}

	// Convert metadata to JSON string
	metadataJSON := "{}"
	if request.Metadata != nil {
		if metadataBytes, err := json.Marshal(request.Metadata); err == nil {
			metadataJSON = string(metadataBytes)
		}
	}

	// Convert style config to JSON string
	styleConfigJSON := "{}"
	if request.StyleConfig != nil {
		if styleBytes, err := json.Marshal(request.StyleConfig); err == nil {
			styleConfigJSON = string(styleBytes)
		}
	}

	return ports.NewRegion{
		ID:               regionID,
		TenantID:         authCtx.TenantID,
		ActorType:        "user", // Default to user, could be parameterized
		ActorID:          authCtx.CreatorID,
		CreatorID:        authCtx.CreatorID,
		SlideUID:         slideUID,
		Version:          1,
		Name:             request.RegionName,
		RegionType:       request.RegionType,
		GeometryData:     geometryJSON,
		CoordinateSystem: request.CoordinateSystem,
		AreaPixels:       request.AreaPixels,
		AreaPhysical:     request.AreaPhysical,
		Labels:           labelsJSON,
		Metadata:         metadataJSON,
		Mutable:          request.Mutable != nil && *request.Mutable,
		Visible:          request.Visible == nil || *request.Visible, // Default to true
		StyleConfig:      styleConfigJSON,
	}
}

// GetRegions retrieves regions with pagination and search
func (s *regionsService) GetRegions(ctx context.Context, params utils.PaginationAndSearchParams, searchParams ports.RegionSearchParams) ([]domain.Region, domain.PaginationInfo, error) {
	// TODO: Implement when database layer is ready
	regions, paginationInfo, err := s.db.GetRegionsGeneric(ctx, params, searchParams)
	if err != nil {
		return nil, domain.PaginationInfo{}, err
	}

	domainRegions := make([]domain.Region, len(regions))
	for i, region := range regions {
		domainRegions[i] = convertRegionDBToDomain(region)
	}

	return domainRegions, paginationInfo, nil
}

// GetRegionsBySlideUID retrieves all regions for a specific slide
func (s *regionsService) GetRegionsBySlideUID(ctx context.Context, slideUID string) ([]domain.Region, error) {
	regions, err := s.db.GetRegionsBySlideUID(ctx, slideUID)
	if err != nil {
		return nil, err
	}

	domainRegions := make([]domain.Region, len(regions))
	for i, region := range regions {
		domainRegions[i] = convertRegionDBToDomain(region)
	}

	return domainRegions, nil
}

// GetRegionByID retrieves a specific region by its ID
func (s *regionsService) GetRegionByID(ctx context.Context, regionID string) (domain.Region, error) {
	region, err := s.db.GetRegionByID(ctx, regionID)
	if err != nil {
		return domain.Region{}, err
	}

	return convertRegionDBToDomain(region), nil
}

// CreateRegion creates a new region
func (s *regionsService) CreateRegion(ctx context.Context, slideUID string, request domain.CreateRegionRequest) (domain.Region, error) {
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return domain.Region{}, err
	}

	// Generate UUID for the region
	regionID := uuid.New().String()

	newRegion := convertDomainToDBRegion(slideUID, request, authCtx, regionID)

	err = s.db.CreateRegion(ctx, newRegion)
	if err != nil {
		return domain.Region{}, err
	}

	// Return the created region
	return s.GetRegionByID(ctx, regionID)
}

// UpdateRegion updates an existing region
func (s *regionsService) UpdateRegion(ctx context.Context, regionID string, request domain.UpdateRegionRequest) error {
	updates := ports.UpdateRegion{}

	if request.RegionName != nil {
		updates.Name = request.RegionName
	}
	if request.RegionType != nil {
		updates.RegionType = request.RegionType
	}
	if request.CoordinateSystem != nil {
		updates.CoordinateSystem = request.CoordinateSystem
	}
	if request.AreaPixels != nil {
		updates.AreaPixels = request.AreaPixels
	}
	if request.AreaPhysical != nil {
		updates.AreaPhysical = request.AreaPhysical
	}
	if request.Mutable != nil {
		updates.Mutable = request.Mutable
	}
	if request.Visible != nil {
		updates.Visible = request.Visible
	}

	// Handle geometry update
	if request.Geometry != nil {
		// Convert geometry to JSON string
		geometryJSON := "" // Should properly serialize the geometry
		updates.GeometryData = &geometryJSON
	}

	// Handle JSON field updates (labels, metadata, style config)
	// These would need proper JSON serialization in a real implementation

	return s.db.UpdateRegion(ctx, regionID, updates)
}

// DeleteRegion soft-deletes a region
func (s *regionsService) DeleteRegion(ctx context.Context, regionID string) error {
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return err
	}

	return s.db.SoftDeleteRegion(ctx, regionID, authCtx.CreatorID)
}

// BulkCreateRegions creates multiple regions
func (s *regionsService) BulkCreateRegions(ctx context.Context, request domain.BulkCreateRegionsRequest) ([]domain.Region, error) {
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	newRegions := make([]ports.NewRegion, len(request.Regions))
	regionIDs := make([]string, len(request.Regions))

	for i, regionRequest := range request.Regions {
		regionID := uuid.New().String()
		regionIDs[i] = regionID
		newRegions[i] = convertDomainToDBRegion(request.SlideUID, regionRequest, authCtx, regionID)
	}

	err = s.db.BulkCreateRegions(ctx, newRegions)
	if err != nil {
		return nil, err
	}

	// Return the created regions
	createdRegions := make([]domain.Region, len(regionIDs))
	for i, regionID := range regionIDs {
		region, err := s.GetRegionByID(ctx, regionID)
		if err != nil {
			return nil, err
		}
		createdRegions[i] = region
	}

	return createdRegions, nil
}

// BulkUpdateRegions updates multiple regions
func (s *regionsService) BulkUpdateRegions(ctx context.Context, request domain.BulkUpdateRegionsRequest) error {
	updates := make(map[string]ports.UpdateRegion)

	for regionID, updateRequest := range request.Updates {
		update := ports.UpdateRegion{}

		if updateRequest.RegionName != nil {
			update.Name = updateRequest.RegionName
		}
		if updateRequest.RegionType != nil {
			update.RegionType = updateRequest.RegionType
		}
		if updateRequest.CoordinateSystem != nil {
			update.CoordinateSystem = updateRequest.CoordinateSystem
		}
		if updateRequest.AreaPixels != nil {
			update.AreaPixels = updateRequest.AreaPixels
		}
		if updateRequest.AreaPhysical != nil {
			update.AreaPhysical = updateRequest.AreaPhysical
		}
		if updateRequest.Mutable != nil {
			update.Mutable = updateRequest.Mutable
		}
		if updateRequest.Visible != nil {
			update.Visible = updateRequest.Visible
		}

		// Handle geometry and JSON field updates
		// These would need proper serialization in a real implementation

		updates[regionID] = update
	}

	return s.db.BulkUpdateRegions(ctx, updates)
}

// BulkDeleteRegions soft-deletes multiple regions
func (s *regionsService) BulkDeleteRegions(ctx context.Context, request domain.BulkDeleteRegionsRequest) error {
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return err
	}

	return s.db.BulkDeleteRegions(ctx, request.RegionIDs, authCtx.CreatorID)
}

// GetDeletedRegions retrieves all soft-deleted regions
func (s *regionsService) GetDeletedRegions(ctx context.Context) ([]domain.Region, error) {
	regions, err := s.db.GetDeletedRegions(ctx)
	if err != nil {
		return nil, err
	}

	domainRegions := make([]domain.Region, len(regions))
	for i, region := range regions {
		domainRegions[i] = convertRegionDBToDomain(region)
	}

	return domainRegions, nil
}

// RestoreRegion restores a soft-deleted region
func (s *regionsService) RestoreRegion(ctx context.Context, regionID string) error {
	return s.db.RestoreRegion(ctx, regionID)
}

// GetRegionStatistics returns statistics for regions on a slide
func (s *regionsService) GetRegionStatistics(ctx context.Context, slideUID string) (domain.RegionStatistics, error) {
	regions, err := s.GetRegionsBySlideUID(ctx, slideUID)
	if err != nil {
		return domain.RegionStatistics{}, err
	}

	stats := domain.RegionStatistics{
		SlideUID:       slideUID,
		TotalRegions:   len(regions),
		RegionsByType:  make(map[string]int),
		RegionsByActor: make(map[string]int),
	}

	visibleCount := 0
	totalAreaPixels := 0
	totalAreaPhysical := 0.0
	pixelCount := 0
	physicalCount := 0

	for _, region := range regions {
		if region.Visible {
			visibleCount++
		}

		// Count by type
		stats.RegionsByType[region.RegionType]++

		// Count by actor
		stats.RegionsByActor[region.ActorType]++

		// Calculate area statistics
		if region.AreaPixels != nil {
			totalAreaPixels += *region.AreaPixels
			pixelCount++
		}
		if region.AreaPhysical != nil {
			totalAreaPhysical += *region.AreaPhysical
			physicalCount++
		}
	}

	stats.VisibleRegions = visibleCount

	if pixelCount > 0 {
		stats.TotalAreaPixels = &totalAreaPixels
		avgPixels := float64(totalAreaPixels) / float64(pixelCount)
		stats.AverageAreaPixels = &avgPixels
	}

	if physicalCount > 0 {
		stats.TotalAreaPhysical = &totalAreaPhysical
		avgPhysical := totalAreaPhysical / float64(physicalCount)
		stats.AverageAreaPhysical = &avgPhysical
	}

	return stats, nil
}

// Close closes the service and its resources
func (s *regionsService) Close() {
	// Close any resources if needed
}
