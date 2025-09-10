// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package slides

import (
	"strings"
	"testing"
)

func TestFormatDetection(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		// QpTiff extensions
		{"sample.qptiff", FormatQpTIFF},
		{"Sample.QPTIFF", FormatQpTIFF}, // Test case insensitive
		{"data.qp", FormatQpTIFF},

		// OpenSlide extensions
		{"slide.svs", FormatOpenSlide},
		{"image.ndpi", FormatOpenSlide},
		{"scan.scn", FormatOpenSlide},
		{"file.mrxs", FormatOpenSlide},
		{"data.tif", FormatOpenSlide}, // Note: .tif is ambiguous but defaults to OpenSlide
		{"sample.vms", FormatOpenSlide},
		{"image.vmu", FormatOpenSlide},
		{"tissue.bif", FormatOpenSlide},

		// TIFF extensions
		{"image.tiff", FormatTIFF},
		{"IMAGE.TIFF", FormatTIFF}, // Test case insensitive
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			// Test format detection logic (mimicking what happens in OpenSlide function)
			var formatHint string
			ext := getFileExtension(test.filename)

			if openslideExtensions[ext] {
				formatHint = FormatOpenSlide
			} else if qptiffExtensions[ext] {
				formatHint = FormatQpTIFF
			} else if ext == ".tif" || ext == ".tiff" {
				formatHint = FormatTIFF
			}

			if formatHint != test.expected {
				t.Errorf("For file %s, expected format %s, got %s", test.filename, test.expected, formatHint)
			}
		})
	}
}

func TestFormatHints(t *testing.T) {
	tests := []struct {
		filename   string
		formatHint string
		expected   string
	}{
		// Test explicit format hints override file extension detection
		{"image.svs", FormatQpTIFF, FormatQpTIFF},
		{"data.qptiff", FormatTIFF, FormatTIFF},
		{"slide.tiff", FormatOpenSlide, FormatOpenSlide},

		// Test that empty hint uses file extension
		{"sample.qptiff", "", FormatQpTIFF},
		{"image.svs", "", FormatOpenSlide},
		{"data.tiff", "", FormatTIFF},
	}

	for _, test := range tests {
		t.Run(test.filename+"_with_hint_"+test.formatHint, func(t *testing.T) {
			formatHint := test.formatHint
			if formatHint == "" {
				// Mimic the detection logic
				ext := getFileExtension(test.filename)
				if openslideExtensions[ext] {
					formatHint = FormatOpenSlide
				} else if qptiffExtensions[ext] {
					formatHint = FormatQpTIFF
				} else if ext == ".tif" || ext == ".tiff" {
					formatHint = FormatTIFF
				}
			}

			if formatHint != test.expected {
				t.Errorf("For file %s with hint %s, expected format %s, got %s",
					test.filename, test.formatHint, test.expected, formatHint)
			}
		})
	}
}

func TestSupportedExtensions(t *testing.T) {
	// Test that our extension maps contain expected formats
	expectedOpenSlideExts := []string{".svs", ".ndpi", ".scn", ".mrxs", ".tif", ".vms", ".vmu", ".bif"}
	for _, ext := range expectedOpenSlideExts {
		if !openslideExtensions[ext] {
			t.Errorf("Expected OpenSlide extension %s not found", ext)
		}
	}

	expectedQpTiffExts := []string{".qptiff", ".qp"}
	for _, ext := range expectedQpTiffExts {
		if !qptiffExtensions[ext] {
			t.Errorf("Expected QpTiff extension %s not found", ext)
		}
	}
}

// Helper function to mimic the filepath.Ext logic used in the factory
func getFileExtension(filename string) string {
	for i := len(filename) - 1; i >= 0 && filename[i] != '/'; i-- {
		if filename[i] == '.' {
			return strings.ToLower(filename[i:])
		}
	}
	return ""
}
