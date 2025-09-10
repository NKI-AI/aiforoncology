// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package ports

import (
	"context"
	"time"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Region represents a region of interest in the database
type Region struct {
	ID               string // UUID stored as TEXT (primary key)
	ActorType        string // 'user' or 'model'
	ActorID          int    // ID of the user or model
	CreatorID        int
	SlideID          int    // Internal slide ID (foreign key to slides.id)
	SlideUID         string // External slide UID for API responses
	Version          int
	Name             string   // Required name (e.g., "Patient 1", "Tissue Section A")
	RegionType       string   // 'roi', 'patient', 'tissue', 'artifact', 'background', 'other'
	GeometryData     string   // GeoJSON geometry (polygon, rectangle, etc.)
	CoordinateSystem string   // 'pixel' or 'physical'
	AreaPixels       *int     // Cached area in pixels (nullable)
	AreaPhysical     *float64 // Cached area in physical units (μm², nullable)
	Labels           string   // JSON labels data
	Metadata         string   // JSON metadata
	Mutable          bool     // Whether the region can be modified
	Visible          bool     // Whether region is visible in UI
	StyleConfig      string   // JSON UI styling configuration
	DeletedAt        *time.Time
	DeletedBy        *int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewRegion represents a new region to be created in the database
type NewRegion struct {
	ID               string // UUID for the new region
	TenantID         int
	ActorType        string // 'user' or 'model'
	ActorID          int    // ID of the user or model
	CreatorID        int
	SlideUID         string // External slide UID, will be converted to internal slide_id
	Version          int
	Name             string   // Required name
	RegionType       string   // 'roi', 'patient', 'tissue', 'artifact', 'background', 'other'
	GeometryData     string   // GeoJSON geometry
	CoordinateSystem string   // 'pixel' or 'physical'
	AreaPixels       *int     // Optional cached area in pixels
	AreaPhysical     *float64 // Optional cached area in physical units
	Labels           string   // JSON labels data
	Metadata         string   // JSON metadata
	Mutable          bool     // Whether the region can be modified
	Visible          bool     // Whether region is visible in UI
	StyleConfig      string   // JSON UI styling configuration
}

// UpdateRegion represents updates to an existing region
type UpdateRegion struct {
	Name             *string  // Optional name update
	RegionType       *string  // Optional region type update
	GeometryData     *string  // Optional geometry update
	CoordinateSystem *string  // Optional coordinate system update
	AreaPixels       *int     // Optional area in pixels update
	AreaPhysical     *float64 // Optional area in physical units update
	Labels           *string  // Optional labels update
	Metadata         *string  // Optional metadata update
	Mutable          *bool    // Optional mutable flag update
	Visible          *bool    // Optional visibility update
	StyleConfig      *string  // Optional style configuration update
}

// RegionSearchParams represents search parameters for regions
type RegionSearchParams struct {
	SlideUID         *string  // Filter by slide UID
	RegionType       *string  // Filter by region type
	ActorType        *string  // Filter by actor type
	CreatorID        *int     // Filter by creator
	Visible          *bool    // Filter by visibility
	MinAreaPixels    *int     // Filter by minimum area in pixels
	MaxAreaPixels    *int     // Filter by maximum area in pixels
	MinAreaPhysical  *float64 // Filter by minimum area in physical units
	MaxAreaPhysical  *float64 // Filter by maximum area in physical units
	CoordinateSystem *string  // Filter by coordinate system
}

// RegionsRepository defines the interface for region-related database operations
type RegionsRepository interface {
	// LoadAllRegions retrieves all regions from the database.
	LoadAllRegions(ctx context.Context) ([]Region, error)

	// GetRegionsGeneric retrieves regions with pagination and search support.
	GetRegionsGeneric(ctx context.Context, params utils.PaginationAndSearchParams, searchParams RegionSearchParams) ([]Region, domain.PaginationInfo, error)

	// GetRegionsBySlideUID retrieves all regions for a specific slide.
	GetRegionsBySlideUID(ctx context.Context, slideUID string) ([]Region, error)

	// CreateRegion adds a new region to the database.
	CreateRegion(ctx context.Context, newRegion NewRegion) error

	// GetRegionByID retrieves a specific region by its ID (UUID).
	GetRegionByID(ctx context.Context, regionID string) (Region, error)

	// UpdateRegion updates an existing region.
	UpdateRegion(ctx context.Context, regionID string, updates UpdateRegion) error

	// SoftDeleteRegion marks a region as deleted without removing it from the database.
	SoftDeleteRegion(ctx context.Context, regionID string, deletedBy int) error

	// GetDeletedRegions retrieves all soft-deleted regions.
	GetDeletedRegions(ctx context.Context) ([]Region, error)

	// RestoreRegion restores a soft-deleted region.
	RestoreRegion(ctx context.Context, regionID string) error

	// BulkCreateRegions creates multiple regions in a single transaction.
	BulkCreateRegions(ctx context.Context, newRegions []NewRegion) error

	// BulkUpdateRegions updates multiple regions in a single transaction.
	BulkUpdateRegions(ctx context.Context, updates map[string]UpdateRegion) error

	// BulkDeleteRegions soft-deletes multiple regions in a single transaction.
	BulkDeleteRegions(ctx context.Context, regionIDs []string, deletedBy int) error
}
