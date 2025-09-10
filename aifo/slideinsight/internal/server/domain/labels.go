// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// RasterLabel represents a single label for raster annotations with name, index, and color
type RasterLabel struct {
	Name  string `json:"name"`  // Display name of the label (e.g., "tumor", "normal")
	Index int    `json:"index"` // Numeric index for the label (e.g., 1, 2, 3) - maps to pixel values
	Color string `json:"color"` // Hex color code (e.g., "#FF0000")
}

// VectorLabel represents a single label for vector annotations with just name and color
type VectorLabel struct {
	Name  string `json:"name"`  // Display name of the label (e.g., "tumor", "normal")
	Color string `json:"color"` // Hex color code (e.g., "#FF0000")
}

// RasterLabels represents the collection of labels for a raster annotation
type RasterLabels []RasterLabel

// VectorLabels represents the collection of labels for a vector annotation
type VectorLabels []VectorLabel

// hexColorRegex validates hex color codes (both #RGB and #RRGGBB formats)
var hexColorRegex = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

// Validate ensures the raster label data is valid
func (l *RasterLabel) Validate() error {
	if l.Name == "" {
		return fmt.Errorf("label name cannot be empty")
	}
	if l.Index < 0 {
		return fmt.Errorf("label index must be non-negative, got %d", l.Index)
	}
	if !hexColorRegex.MatchString(l.Color) {
		return fmt.Errorf("invalid color format '%s', expected hex color (e.g., #FF0000 or #F00)", l.Color)
	}
	return nil
}

// Validate ensures the vector label data is valid
func (l *VectorLabel) Validate() error {
	if l.Name == "" {
		return fmt.Errorf("label name cannot be empty")
	}
	if !hexColorRegex.MatchString(l.Color) {
		return fmt.Errorf("invalid color format '%s', expected hex color (e.g., #FF0000 or #F00)", l.Color)
	}
	return nil
}

// Validate ensures the raster labels collection is valid
func (ls *RasterLabels) Validate() error {
	if len(*ls) == 0 {
		return nil // Empty labels are allowed
	}

	// Track used indices and names to ensure uniqueness
	usedIndices := make(map[int]bool)
	usedNames := make(map[string]bool)

	for i, label := range *ls {
		if err := label.Validate(); err != nil {
			return fmt.Errorf("label at index %d: %w", i, err)
		}

		// Check for duplicate indices
		if usedIndices[label.Index] {
			return fmt.Errorf("duplicate label index %d", label.Index)
		}
		usedIndices[label.Index] = true

		// Check for duplicate names (case-insensitive)
		nameLower := strings.ToLower(label.Name)
		if usedNames[nameLower] {
			return fmt.Errorf("duplicate label name '%s'", label.Name)
		}
		usedNames[nameLower] = true
	}

	return nil
}

// Validate ensures the vector labels collection is valid
func (ls *VectorLabels) Validate() error {
	if len(*ls) == 0 {
		return nil // Empty labels are allowed
	}

	// Track used names to ensure uniqueness (no indices for vector labels)
	usedNames := make(map[string]bool)

	for i, label := range *ls {
		if err := label.Validate(); err != nil {
			return fmt.Errorf("label at index %d: %w", i, err)
		}

		// Check for duplicate names (case-insensitive)
		nameLower := strings.ToLower(label.Name)
		if usedNames[nameLower] {
			return fmt.Errorf("duplicate label name '%s'", label.Name)
		}
		usedNames[nameLower] = true
	}

	return nil
}

// ToJSON converts raster labels to JSON string for database storage
func (ls *RasterLabels) ToJSON() (string, error) {
	if err := ls.Validate(); err != nil {
		return "", err
	}

	data, err := json.Marshal(ls)
	if err != nil {
		return "", fmt.Errorf("failed to marshal raster labels to JSON: %w", err)
	}

	return string(data), nil
}

// ToJSON converts vector labels to JSON string for database storage
func (ls *VectorLabels) ToJSON() (string, error) {
	if err := ls.Validate(); err != nil {
		return "", err
	}

	data, err := json.Marshal(ls)
	if err != nil {
		return "", fmt.Errorf("failed to marshal vector labels to JSON: %w", err)
	}

	return string(data), nil
}

// FromJSON parses raster labels from JSON string
func (ls *RasterLabels) FromJSON(jsonStr string) error {
	if jsonStr == "" {
		*ls = nil
		return nil
	}

	if err := json.Unmarshal([]byte(jsonStr), ls); err != nil {
		return fmt.Errorf("failed to unmarshal raster labels from JSON: %w", err)
	}

	return ls.Validate()
}

// FromJSON parses vector labels from JSON string
func (ls *VectorLabels) FromJSON(jsonStr string) error {
	if jsonStr == "" {
		*ls = nil
		return nil
	}

	if err := json.Unmarshal([]byte(jsonStr), ls); err != nil {
		return fmt.Errorf("failed to unmarshal vector labels from JSON: %w", err)
	}

	return ls.Validate()
}
