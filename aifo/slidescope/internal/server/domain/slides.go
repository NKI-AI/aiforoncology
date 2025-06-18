// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.

package domain

// Slide represents a slide
type Slide struct {
	SlideID     string  `json:"slideId"`
	SlideName   string  `json:"slideName,omitempty"`
	SlideURI    string  `json:"slideUri"`
	SlideWidth  int     `json:"slideWidth,omitempty"`
	SlideHeight int     `json:"slideHeight,omitempty"`
	SlideMpp    float64 `json:"slideMpp,omitempty"`
}

// PhysicalDimensions represents the physical dimensions of a slide
type PhysicalDimensions struct {
	WidthMm  float64 `json:"widthMm"`
	HeightMm float64 `json:"heightMm"`
	WidthPx  int     `json:"widthPx"`
	HeightPx int     `json:"heightPx"`
}

// SlideMetadata represents detailed metadata for a slide
type SlideMetadata struct {
	SlideID            string             `json:"slideId"`
	SlideName          string             `json:"slideName"`
	MinLevel           int                `json:"minLevel"`
	MaxLevel           int                `json:"maxLevel"`
	TileSize           int                `json:"tileSize"`
	Format             string             `json:"format"`
	SlideMpp           float64            `json:"slideMpp"`
	SlideWidth         int                `json:"slideWidth"`
	SlideHeight        int                `json:"slideHeight"`
	Vendor             string             `json:"vendor"`
	Magnification      string             `json:"magnification"`
	PhysicalDimensions PhysicalDimensions `json:"physicalDimensions"`
}

// SlidesResponse represents a response containing a list of slides
type SlidesResponse struct {
	Slides []Slide `json:"slides"`
}

// SlideTile represents a tile image from a slide
type SlideTile struct {
	Image       []byte `json:"-"` // Raw image data
	Format      string `json:"-"` // Image format (jpeg, png)
	ContentType string `json:"-"` // HTTP content type
}
