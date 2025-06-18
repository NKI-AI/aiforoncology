// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package slides

import (
	"path/filepath"
	"strings"
)

// Supported file formats
const (
	FormatOpenSlide = "openslide" // For formats supported by OpenSlide (SVS, NDPI, etc.)
	FormatTIFF      = "tiff"      // For single-channel TIFF files
)

// OpenSlide supported extensions
var openslideExtensions = map[string]bool{
	".svs":  true, // Aperio
	".ndpi": true, // Hamamatsu
	".scn":  true, // Leica
	".mrxs": true, // MIRAX
	".tif":  true, // BigTIFF
	".vms":  true, // Hamamatsu
	".vmu":  true, // Hamamatsu
	".bif":  true, // Ventana
}

// OpenSlide creates a slide using the appropriate implementation based on file format.
// It tries to determine the format from the file extension or format hint if provided.
func OpenSlide(path string, formatHint string) (Slide, error) {
	if formatHint == "" {
		// Try to determine format from file extension
		ext := strings.ToLower(filepath.Ext(path))
		if openslideExtensions[ext] {
			formatHint = FormatOpenSlide
		} else if ext == ".tif" || ext == ".tiff" {
			formatHint = FormatTIFF
		}
	}

	// Create the appropriate slide implementation
	switch formatHint {
	case FormatOpenSlide:
		return NewOpenSlideAdapter(path)
	case FormatTIFF:
		return NewTiffAdapter(path)
	default:
		// If format is still unknown, try OpenSlide first, then TIFF
		slide, err := NewOpenSlideAdapter(path)
		if err == nil {
			return slide, nil
		}

		// If OpenSlide fails, try TIFF
		return NewTiffAdapter(path)
	}
}
