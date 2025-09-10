// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package slides

// Common property key constants used across different slide formats
// Based on OpenSlide property names for compatibility
const (
	PropBackgroundColor = "openslide.background-color"
	PropBoundsHeight    = "openslide.bounds-height"
	PropBoundsWidth     = "openslide.bounds-width"
	PropBoundsX         = "openslide.bounds-x"
	PropBoundsY         = "openslide.bounds-y"
	PropObjectivePower  = "openslide.objective-power"

	// Generalized property names (format-agnostic)
	PropMppX = "mpp-x" // Microns per pixel in X direction
	PropMppY = "mpp-y" // Microns per pixel in Y direction

	// OpenSlide-specific property names (for backwards compatibility)
	PropOpenSlideMppX = "openslide.mpp-x"
	PropOpenSlideMppY = "openslide.mpp-y"
)
