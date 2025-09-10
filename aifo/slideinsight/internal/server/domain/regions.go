// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package domain

// RegionGeometry represents the GeoJSON geometry for a region
type RegionGeometry struct {
	Type        string      `json:"type" validate:"required,oneof=Polygon Rectangle Point LineString MultiPolygon"`
	Coordinates interface{} `json:"coordinates" validate:"required"`
}

// RegionLabels represents structured labels for regions
type RegionLabels struct {
	Primary   *string           `json:"primary,omitempty"`   // Primary classification
	Secondary *string           `json:"secondary,omitempty"` // Secondary classification
	Tags      []string          `json:"tags,omitempty"`      // Additional tags
	Custom    map[string]string `json:"custom,omitempty"`    // Custom key-value labels
}

// RegionStyleConfig represents UI styling configuration for regions
type RegionStyleConfig struct {
	StrokeColor   *string  `json:"strokeColor,omitempty"`   // Hex color for border
	FillColor     *string  `json:"fillColor,omitempty"`     // Hex color for fill
	StrokeWidth   *int     `json:"strokeWidth,omitempty"`   // Border width in pixels
	FillOpacity   *float64 `json:"fillOpacity,omitempty"`   // Fill opacity (0.0-1.0)
	StrokeOpacity *float64 `json:"strokeOpacity,omitempty"` // Stroke opacity (0.0-1.0)
	ZIndex        *int     `json:"zIndex,omitempty"`        // Layer ordering
}

// RegionMetadata represents additional metadata for regions
type RegionMetadata struct {
	PatientID    *string                `json:"patientId,omitempty"`    // Associated patient ID
	SectionID    *string                `json:"sectionId,omitempty"`    // Tissue section identifier
	TissueType   *string                `json:"tissueType,omitempty"`   // Type of tissue
	Diagnosis    *string                `json:"diagnosis,omitempty"`    // Medical diagnosis
	Notes        *string                `json:"notes,omitempty"`        // Free-form notes
	QualityScore *float64               `json:"qualityScore,omitempty"` // Quality assessment score
	Custom       map[string]interface{} `json:"custom,omitempty"`       // Custom metadata fields
}

// Region represents a region of interest for a slide
type Region struct {
	RegionID         string             `json:"regionId"`                                                                          // UUID identifier for the region
	RegionName       string             `json:"regionName" validate:"required"`                                                    // Display name for the region
	SlideUID         string             `json:"slideUid,omitempty" validate:"required,slide_uid"`                                  // Direct reference to slide UID
	RegionType       string             `json:"regionType" validate:"required,oneof=roi patient tissue artifact background other"` // Type of region
	Geometry         RegionGeometry     `json:"geometry" validate:"required"`                                                      // GeoJSON geometry
	CoordinateSystem string             `json:"coordinateSystem" validate:"required,oneof=pixel physical"`                         // Coordinate system used
	AreaPixels       *int               `json:"areaPixels,omitempty"`                                                              // Cached area in pixels
	AreaPhysical     *float64           `json:"areaPhysical,omitempty"`                                                            // Cached area in physical units (μm²)
	Labels           *RegionLabels      `json:"labels,omitempty"`                                                                  // Structured label data
	Metadata         string             `json:"metadata,omitempty"`                                                                // Additional metadata
	StyleConfig      *RegionStyleConfig `json:"styleConfig,omitempty"`                                                             // UI styling configuration
	ActorType        string             `json:"actorType,omitempty"`                                                               // Actor type (user, model)
	ActorID          int                `json:"actorId,omitempty"`                                                                 // Actor ID
	Mutable          bool               `json:"mutable"`                                                                           // Whether the region can be modified
	Visible          bool               `json:"visible"`                                                                           // Whether region is visible in UI
	DeletedAt        *string            `json:"deletedAt,omitempty"`                                                               // Soft deletion timestamp
	DeletedBy        *int               `json:"deletedBy,omitempty"`                                                               // User who deleted the region
	CreatedAt        string             `json:"createdAt,omitempty"`                                                               // Creation timestamp
	UpdatedAt        string             `json:"updatedAt,omitempty"`                                                               // Last update timestamp
}

// RegionList represents a list of regions for a slide
type RegionList struct {
	SlideUID   string   `json:"slideUid"`
	Regions    []Region `json:"regions"`
	TotalCount int      `json:"totalCount"`
	HasMore    bool     `json:"hasMore"`
	NextCursor *string  `json:"nextCursor,omitempty"`
}

// CreateRegionRequest represents a request to create a new region
type CreateRegionRequest struct {
	RegionName       string             `json:"regionName" validate:"required"`
	SlideUID         string             `json:"slideUid" validate:"required,slide_uid"`
	RegionType       string             `json:"regionType" validate:"required,oneof=roi patient tissue artifact background other"`
	Geometry         RegionGeometry     `json:"geometry" validate:"required"`
	CoordinateSystem string             `json:"coordinateSystem" validate:"required,oneof=pixel physical"`
	AreaPixels       *int               `json:"areaPixels,omitempty"`
	AreaPhysical     *float64           `json:"areaPhysical,omitempty"`
	Labels           *RegionLabels      `json:"labels,omitempty"`
	Metadata         *RegionMetadata    `json:"metadata,omitempty"`
	StyleConfig      *RegionStyleConfig `json:"styleConfig,omitempty"`
	Mutable          *bool              `json:"mutable,omitempty"` // Defaults to true
	Visible          *bool              `json:"visible,omitempty"` // Defaults to true
}

// UpdateRegionRequest represents a request to update an existing region
type UpdateRegionRequest struct {
	RegionName       *string            `json:"regionName,omitempty"`
	RegionType       *string            `json:"regionType,omitempty" validate:"omitempty,oneof=roi patient tissue artifact background other"`
	Geometry         *RegionGeometry    `json:"geometry,omitempty"`
	CoordinateSystem *string            `json:"coordinateSystem,omitempty" validate:"omitempty,oneof=pixel physical"`
	AreaPixels       *int               `json:"areaPixels,omitempty"`
	AreaPhysical     *float64           `json:"areaPhysical,omitempty"`
	Labels           *RegionLabels      `json:"labels,omitempty"`
	Metadata         *RegionMetadata    `json:"metadata,omitempty"`
	StyleConfig      *RegionStyleConfig `json:"styleConfig,omitempty"`
	Mutable          *bool              `json:"mutable,omitempty"`
	Visible          *bool              `json:"visible,omitempty"`
}

// BulkCreateRegionsRequest represents a request to create multiple regions
type BulkCreateRegionsRequest struct {
	SlideUID string                `json:"slideUid" validate:"required,slide_uid"`
	Regions  []CreateRegionRequest `json:"regions" validate:"required,min=1,max=100,dive"`
}

// BulkUpdateRegionsRequest represents a request to update multiple regions
type BulkUpdateRegionsRequest struct {
	Updates map[string]UpdateRegionRequest `json:"updates" validate:"required,min=1,max=100"`
}

// BulkDeleteRegionsRequest represents a request to delete multiple regions
type BulkDeleteRegionsRequest struct {
	RegionIDs []string `json:"regionIds" validate:"required,min=1,max=100,dive,required"`
}

// RegionSearchRequest represents search parameters for regions
type RegionSearchRequest struct {
	SlideUID         *string  `json:"slideUid,omitempty"`
	RegionType       *string  `json:"regionType,omitempty" validate:"omitempty,oneof=roi patient tissue artifact background other"`
	ActorType        *string  `json:"actorType,omitempty" validate:"omitempty,oneof=user model"`
	CreatorID        *int     `json:"creatorId,omitempty"`
	Visible          *bool    `json:"visible,omitempty"`
	MinAreaPixels    *int     `json:"minAreaPixels,omitempty"`
	MaxAreaPixels    *int     `json:"maxAreaPixels,omitempty"`
	MinAreaPhysical  *float64 `json:"minAreaPhysical,omitempty"`
	MaxAreaPhysical  *float64 `json:"maxAreaPhysical,omitempty"`
	CoordinateSystem *string  `json:"coordinateSystem,omitempty" validate:"omitempty,oneof=pixel physical"`
	Query            *string  `json:"query,omitempty"` // Text search in name, labels, metadata
	Tags             []string `json:"tags,omitempty"`  // Filter by label tags
}

// RegionStatistics represents statistics for regions on a slide
type RegionStatistics struct {
	SlideUID            string         `json:"slideUid"`
	TotalRegions        int            `json:"totalRegions"`
	VisibleRegions      int            `json:"visibleRegions"`
	RegionsByType       map[string]int `json:"regionsByType"`
	RegionsByActor      map[string]int `json:"regionsByActor"`
	TotalAreaPixels     *int           `json:"totalAreaPixels,omitempty"`
	TotalAreaPhysical   *float64       `json:"totalAreaPhysical,omitempty"`
	AverageAreaPixels   *float64       `json:"averageAreaPixels,omitempty"`
	AverageAreaPhysical *float64       `json:"averageAreaPhysical,omitempty"`
}
