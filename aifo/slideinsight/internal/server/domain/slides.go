// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

package domain

import (
	"encoding/json"
)

// Slide represents a slide
type Slide struct {
	CaseID      int             `json:"-"`
	CaseUID     string          `json:"caseUid,omitempty"`
	SlideUID    string          `json:"slideUid"`
	SlideID     int             `json:"-"`
	SlideName   string          `json:"slideName,omitempty"`
	SlideURI    string          `json:"-"` // TODO: For admins this field should be fine.
	SlideWidth  int             `json:"slideWidth,omitempty"`
	SlideHeight int             `json:"slideHeight,omitempty"`
	SlideMpp    float64         `json:"slideMpp,omitempty"`
	ImageTypeId string          `json:"imageTypeId,omitempty"` // Reference to image type
	Metadata    json.RawMessage `json:"metadata,omitempty"`    // JSON metadata including channel info for spectral slides
	CreatorUID  string          `json:"creatorUid,omitempty"`
	DeletedAt   *string         `json:"deletedAt,omitempty"`
	DeletedBy   *int            `json:"deletedBy,omitempty"`
	CreatedAt   string          `json:"createdAt,omitempty"`
	UpdatedAt   string          `json:"updatedAt,omitempty"`
}

// SlideMetadata represents detailed metadata for a slide
type SlideMetadata struct {
	SlideUID      string  `json:"slideUid"`
	SlideName     string  `json:"slideName"`
	MinLevel      int     `json:"minLevel"`
	MaxLevel      int     `json:"maxLevel"`
	TileSize      int     `json:"tileSize"`
	Format        string  `json:"format"`
	SlideMpp      float64 `json:"slideMpp"`
	SlideWidth    int     `json:"slideWidth"`
	SlideHeight   int     `json:"slideHeight"`
	Vendor        string  `json:"vendor"`
	Magnification string  `json:"magnification"`
}

// ImageType represents an image type (brightfield H&E, fluorescence, etc.)
type ImageType struct {
	ID                string `json:"id"`                       // e.g. 'img_type_bf_he'
	TypeUID           string `json:"typeUid"`                  // e.g. 'brightfield_he'
	Name              string `json:"name"`                     // e.g. 'Brightfield H&E'
	Description       string `json:"description,omitempty"`    // Optional longer description
	Category          string `json:"category"`                 // brightfield, fluorescence, other
	RequiresHistogram bool   `json:"requiresHistogram"`        // Whether this type needs histogram data
	MetadataSchema    string `json:"metadataSchema,omitempty"` // JSON schema for per-type metadata
	IsActive          bool   `json:"isActive"`                 // Whether this type is active
	CreatedAt         string `json:"createdAt,omitempty"`
	UpdatedAt         string `json:"updatedAt,omitempty"`
}

// ImageTypesResponse represents a response containing a list of image types
type ImageTypesResponse struct {
	ImageTypes []ImageType    `json:"imageTypes"`
	Pagination PaginationInfo `json:"pagination"`
}

// SlideHistogram represents histogram data for a slide
type SlideHistogram struct {
	ID            string  `json:"id"`                    // UUID
	SlideUID      string  `json:"slideUid"`              // Reference to slide
	ChannelIndex  int     `json:"channelIndex"`          // Channel index (0-based)
	ChannelName   string  `json:"channelName,omitempty"` // Optional channel name
	BinCount      int     `json:"binCount"`              // Number of histogram bins
	MinValue      float64 `json:"minValue"`              // Minimum value in histogram
	MaxValue      float64 `json:"maxValue"`              // Maximum value in histogram
	HistogramData []byte  `json:"-"`                     // Raw histogram data (binary)
	Counts        []int   `json:"counts,omitempty"`      // Processed histogram counts for API
	Metadata      string  `json:"metadata,omitempty"`    // Additional metadata as JSON
	CreatedAt     string  `json:"createdAt,omitempty"`
	UpdatedAt     string  `json:"updatedAt,omitempty"`
}

// SlideHistogramResponse represents histogram data for a slide with multiple channels
type SlideHistogramResponse struct {
	SlideUID   string           `json:"slideUid"`
	Histograms []SlideHistogram `json:"histograms"`
}

// StainingProtocol represents a staining protocol used for a slide
type StainingProtocol struct {
	ID             string `json:"id"`                       // UUID
	SlideUID       string `json:"slideUid"`                 // Reference to slide
	StainName      string `json:"stainName"`                // e.g. 'Hematoxylin', 'DAPI'
	StainType      string `json:"stainType"`                // primary, counterstain, fluorophore, other
	Concentration  string `json:"concentration,omitempty"`  // e.g. '1:1000'
	IncubationTime string `json:"incubationTime,omitempty"` // e.g. '30m'
	AntibodyInfo   string `json:"antibodyInfo,omitempty"`   // JSON with clone, supplier, etc.
	ExcitationNm   *int   `json:"excitationNm,omitempty"`   // For fluorophores
	EmissionNm     *int   `json:"emissionNm,omitempty"`     // For fluorophores
	Metadata       string `json:"metadata,omitempty"`       // Additional metadata as JSON
	CreatedAt      string `json:"createdAt,omitempty"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
}

// StainingProtocolsResponse represents a list of staining protocols for a slide
type StainingProtocolsResponse struct {
	SlideUID          string             `json:"slideUid"`
	StainingProtocols []StainingProtocol `json:"stainingProtocols"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"totalPages"`
	HasNext    bool `json:"hasNext"`
	HasPrev    bool `json:"hasPrev"`
}

// SlidesResponse represents a response containing a list of slides with pagination
type SlidesResponse struct {
	Slides     []Slide        `json:"slides"`
	Pagination PaginationInfo `json:"pagination"`
}

// SlideTile represents a tile image from a slide
type SlideTile struct {
	Image       []byte `json:"-"` // Raw image data
	Format      string `json:"-"` // Image format (jpg, png)
	ContentType string `json:"-"` // HTTP content type
}

// SlideAnnotationsOverview represents an overview of annotations for a slide
type SlideAnnotationsOverview struct {
	SlideUID    string `json:"slideUid"`    // Slide identifier
	RasterURL   string `json:"rasterUrl"`   // URL to get raster annotations
	VectorURL   string `json:"vectorUrl"`   // URL to get vector annotations
	RasterCount int    `json:"rasterCount"` // Number of raster annotations
	VectorCount int    `json:"vectorCount"` // Number of vector annotations
}
