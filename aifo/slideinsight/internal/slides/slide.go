// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package slides

import (
	"image"

	"aifo.dev/aifo/fastslide_go/fastslide"
)

// ChannelMetadata represents metadata for a single channel in a spectral image.
type ChannelMetadata struct {
	Name      string   // Channel name
	Biomarker string   // Associated biomarker
	Color     [3]uint8 // RGB color for visualization (optional)
}

// DisplayRange represents the min/max values for displaying a channel.
type DisplayRange struct {
	Min float64
	Max float64
}

// SpectralConversionOptions contains parameters for converting spectral images to RGB.
type SpectralConversionOptions struct {
	ChannelMetadata []ChannelMetadata
	DisplayRanges   []DisplayRange
	Saturation      float64 // Saturation value for histogram-based display range computation
}

// Slide defines the common interface that all slide implementations must satisfy.
// This allows the XYZ tile generator to work with different slide backends.
type Slide interface {
	// Close releases resources associated with the slide.
	Close()

	// LevelCount returns the number of levels in the pyramid.
	LevelCount() (int, error)

	// LargestLevelDimensions returns the dimensions of level 0 (most detailed).
	LargestLevelDimensions() ([2]int, error)

	// LevelDimensions returns the dimensions for the specified level.
	LevelDimensions(level int) ([2]int, error)

	// LevelDownsample returns the downsample factor of a level.
	LevelDownsample(level int) (float64, error)

	// LevelDownsamples returns all downsample factors.
	LevelDownsamples() ([]float64, error)

	// BestLevelForDownsample returns the closest level for the requested downsample.
	BestLevelForDownsample(downsample float64) (int, error)

	// ReadRegion reads a region at the specified level and location.
	ReadRegion(x, y, level, w, h int) (image.Image, error)

	// ReadRegionAsFastslideImage reads a region with optional spectral conversion parameters.
	// This method allows spectral slides to use advanced RGB conversion.
	ReadRegionAsFastslideImage(x, y, level, w, h int) (*fastslide.Image, error)

	// Mpp returns the microns per pixel for X and Y directions.
	Mpp() (mppX, mppY float64, err error)

	// PropertyValue returns the value of a specific property.
	PropertyValue(name string) (string, error)

	// Properties returns all properties as a map.
	Properties() (map[string]string, error)

	// IsSpectral returns true if this slide contains spectral/multiplex data.
	IsSpectral() bool

	// GetChannelMetadata returns channel metadata for spectral slides.
	// Returns nil for non-spectral slides.
	GetChannelMetadata() ([]ChannelMetadata, error)
}

// SlideOpener is a function type that can open a slide file.
type SlideOpener func(path string) (Slide, error)
