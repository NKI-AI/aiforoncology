// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package domain

// Mask represents a segmentation mask annotation for a slide
type Mask struct {
	MaskUID    string       `json:"maskUid"`                                          // Unique identifier for the mask
	MaskName   string       `json:"maskName"`                                         // Display name for the mask
	MaskURI    string       `json:"maskUri" validate:"required"`                      // URI to the mask file, should only be present for admins.
	FileHash   *string      `json:"fileHash,omitempty"`                               // Hash for deduplication, nullable and unique
	TilesURL   string       `json:"tilesUrl"`                                         // URL to fetch mask tiles
	SlideUID   string       `json:"slideUid,omitempty" validate:"required,slide_uid"` // Direct reference to slide UID
	Labels     RasterLabels `json:"labels,omitempty"`                                 // Structured label data with colors and indices
	MaskWidth  int          `json:"maskWidth,omitempty" validate:"min=1"`             // Width of the mask in pixels
	MaskHeight int          `json:"maskHeight,omitempty" validate:"min=1"`            // Height of the mask in pixels
	MaskMpp    float64      `json:"maskMpp,omitempty" validate:"min=0"`               // MPP of the mask
	ActorType  string       `json:"actorType,omitempty"`                              // Actor type (user, model)
	ActorID    int          `json:"actorId,omitempty"`                                // Actor ID
	Mutable    bool         `json:"mutable"`                                          // Whether the annotation can be modified
	DeletedAt  *string      `json:"deletedAt,omitempty"`                              // Soft deletion timestamp
	DeletedBy  *int         `json:"deletedBy,omitempty"`                              // User who deleted the mask
	CreatedAt  string       `json:"createdAt,omitempty"`                              // Creation timestamp
}

// MaskList represents a list of masks for a slide
type MaskList struct {
	SlideUID string `json:"slideUid"`
	Masks    []Mask `json:"masks"`
}

// RasterAnnotationsResponse represents a paginated list of raster annotations
type RasterAnnotationsResponse struct {
	Annotations []Mask         `json:"annotations"`
	Pagination  PaginationInfo `json:"pagination"`
}

// MaskTile represents a tile from a mask annotation
type MaskTile struct {
	Image       []byte `json:"-"` // Raw image data
	Format      string `json:"-"` // Image format (png for masks)
	ContentType string `json:"-"` // HTTP content type
}
