// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package domain

// Mask represents a segmentation mask annotation for a slide
type Mask struct {
	MaskID     string  `json:"maskId"`               // Unique identifier for the mask
	MaskName   string  `json:"maskName"`             // Display name for the mask
	MaskURI    string  `json:"maskUri"`              // URI to the mask file
	TilesURL   string  `json:"tilesUrl"`             // URL to fetch mask tiles
	SlideID    string  `json:"slideId,omitempty"`    // Direct reference to slide ID
	MaskWidth  int     `json:"maskWidth,omitempty"`  // Width of the mask in pixels
	MaskHeight int     `json:"maskHeight,omitempty"` // Height of the mask in pixels
	MaskMpp    float64 `json:"maskMpp,omitempty"`    // MPP of the mask
}

// MaskList represents a list of masks for a slide
type MaskList struct {
	SlideID string `json:"slide_id"`
	Masks   []Mask `json:"masks"`
}

// MaskTile represents a tile from a mask annotation
type MaskTile struct {
	Image       []byte `json:"-"` // Raw image data
	Format      string `json:"-"` // Image format (png for masks)
	ContentType string `json:"-"` // HTTP content type
}
