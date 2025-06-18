// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package slides

import (
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"math"

	"golang.org/x/image/draw"
)

// XYZTileGenerator generates XYZ format tiles for deep zoom viewing with OpenLayers.
//
// XYZ format uses a coordinate system where:
// - z: zoom level (0 is the most zoomed out, max_level is most detailed)
// - x, y: tile coordinates at the given zoom level
//
// This implementation adapts OpenSlide's level mapping strategy to handle
// missing levels in the original slide.
type XYZTileGenerator struct {
	// The slide object implementing our Slide interface
	slide Slide

	// Fixed tile size for optimal performance with browsers
	tileSize int

	// Background color for empty regions
	backgroundColor color.Color

	// If the slide is a mask (affects resampling)
	isMask bool

	// Basic slide properties
	level0Dims          [2]int
	slideLevelsCount    int
	slideDownsamples    []float64
	maxZoomLevel        int
	zoomLevelsCount     int
	slideLevelForXYZ    []int
	relativeDownsamples []float64
	xyzDownsamples      []float64
}

// NewXYZTileGenerator creates a new XYZ tile generator for the given slide.
func NewXYZTileGenerator(slide Slide, backgroundColor string, isMask bool) (*XYZTileGenerator, error) {
	// Parse background color
	bg := parseHexColor(backgroundColor)
	if bg == nil {
		// Default to white if color parsing fails
		bg = color.RGBA{255, 255, 255, 255}
	}

	// Create the generator
	generator := &XYZTileGenerator{
		slide:           slide,
		tileSize:        512, // Fixed tile size
		backgroundColor: bg,
		isMask:          isMask,
	}

	// Initialize basic properties
	var err error
	generator.level0Dims, err = slide.LargestLevelDimensions()
	if err != nil {
		return nil, fmt.Errorf("failed to get largest level dimensions: %w", err)
	}

	generator.slideLevelsCount, err = slide.LevelCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get level count: %w", err)
	}

	generator.slideDownsamples, err = slide.LevelDownsamples()
	if err != nil {
		return nil, fmt.Errorf("failed to get level downsamples: %w", err)
	}

	// Calculate max zoom level and initialize level mappings
	generator.maxZoomLevel = generator.calculateMaxZoom()
	generator.zoomLevelsCount = generator.maxZoomLevel + 1
	err = generator.initLevelMappings()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize level mappings: %w", err)
	}

	// Check for background color property in slide
	bgColorProp, err := slide.PropertyValue(PropBackgroundColor)
	if err != nil {
		return nil, fmt.Errorf("failed to get background color property: %w", err)
	}

	if bgColorProp != "" {
		if parsedColor := parseHexColor(bgColorProp); parsedColor != nil {
			generator.backgroundColor = parsedColor
		}
	}

	return generator, nil
}

// calculateMaxZoom determines the maximum zoom level based on slide dimensions.
func (g *XYZTileGenerator) calculateMaxZoom() int {
	maxDim := math.Max(float64(g.level0Dims[0]), float64(g.level0Dims[1]))
	return int(math.Ceil(math.Log2(maxDim / float64(g.tileSize))))
}

// initLevelMappings initializes mappings between XYZ zoom levels and slide levels.
func (g *XYZTileGenerator) initLevelMappings() error {
	// Calculate downsampling factor for each XYZ zoom level
	g.xyzDownsamples = make([]float64, g.zoomLevelsCount)
	for z := 0; z < g.zoomLevelsCount; z++ {
		g.xyzDownsamples[z] = math.Pow(2.0, float64(g.maxZoomLevel-z))
	}

	// Find the best slide level to use for each XYZ zoom level
	g.slideLevelForXYZ = make([]int, g.zoomLevelsCount)
	g.relativeDownsamples = make([]float64, g.zoomLevelsCount)

	for z := 0; z < g.zoomLevelsCount; z++ {
		var err error
		g.slideLevelForXYZ[z], err = g.slide.BestLevelForDownsample(g.xyzDownsamples[z])
		if err != nil {
			return fmt.Errorf("failed to find best level for downsample (zoom %d): %w", z, err)
		}
		g.relativeDownsamples[z] = g.xyzDownsamples[z] / g.slideDownsamples[g.slideLevelForXYZ[z]]
	}

	return nil
}

// GetTile retrieves a tile for the specified XYZ coordinates.
func (g *XYZTileGenerator) GetTile(z, x, y int) (image.Image, error) {
	// Validate zoom level
	if z < 0 || z > g.maxZoomLevel {
		return nil, fmt.Errorf("invalid zoom level: %d (valid range: 0-%d)", z, g.maxZoomLevel)
	}

	// Calculate theoretical dimensions at this zoom level
	levelWidth := int(math.Ceil(float64(g.level0Dims[0]) / g.xyzDownsamples[z]))
	levelHeight := int(math.Ceil(float64(g.level0Dims[1]) / g.xyzDownsamples[z]))

	// Calculate max tile indices at this level
	maxX := int(math.Ceil(float64(levelWidth)/float64(g.tileSize))) - 1
	maxY := int(math.Ceil(float64(levelHeight)/float64(g.tileSize))) - 1

	// Return an error if coordinates are out of bounds
	if x < 0 || y < 0 || x > maxX || y > maxY {
		return nil, fmt.Errorf("tile coordinates out of bounds: x=%d, y=%d, z=%d (valid range: x=0-%d, y=0-%d)",
			x, y, z, maxX, maxY)
	}

	// Translate from xyz tile coordinates to level 0 pixel coordinates
	l0X := int(float64(x) * float64(g.tileSize) * g.xyzDownsamples[z])
	l0Y := int(float64(y) * float64(g.tileSize) * g.xyzDownsamples[z])

	// Get the best slide level to use
	slideLevel := g.slideLevelForXYZ[z]
	slideDownsample := g.slideDownsamples[slideLevel]
	relativeDownsample := g.relativeDownsamples[z]

	// Calculate size to read in slide level coordinates
	readSizeAtSlideLevel := int(math.Ceil(float64(g.tileSize) * relativeDownsample))

	// Check if we're reading past the edge of the slide
	remainingWidth := max(0, g.level0Dims[0]-l0X)
	remainingHeight := max(0, g.level0Dims[1]-l0Y)

	// Convert remaining dimensions to slide level coordinates
	remainingWidthAtLevel := int(math.Ceil(float64(remainingWidth) / slideDownsample))
	remainingHeightAtLevel := int(math.Ceil(float64(remainingHeight) / slideDownsample))

	// Final read size (handling edge cases)
	readWidth := min(readSizeAtSlideLevel, remainingWidthAtLevel)
	readHeight := min(readSizeAtSlideLevel, remainingHeightAtLevel)

	// If we're completely off the slide, return an error
	if readWidth <= 0 || readHeight <= 0 {
		return nil, fmt.Errorf("tile region is outside slide bounds")
	}

	// Read the region from the slide
	region, err := g.slide.ReadRegion(l0X, l0Y, slideLevel, readWidth, readHeight)
	if err != nil {
		return nil, fmt.Errorf("failed to read region: %w", err)
	}

	// If we're at the edge, compose with background color
	if readWidth < readSizeAtSlideLevel || readHeight < readSizeAtSlideLevel {
		// Create background image
		bg := image.NewRGBA(image.Rect(0, 0, readSizeAtSlideLevel, readSizeAtSlideLevel))
		draw.Draw(bg, bg.Bounds(), &image.Uniform{g.backgroundColor}, image.Point{}, draw.Src)

		// Paste the region onto the background
		draw.Draw(bg, image.Rect(0, 0, readWidth, readHeight), region, image.Point{}, draw.Over)
		region = bg
	}

	// Resize to the exact tile size expected by OpenLayers
	if region.Bounds().Dx() != g.tileSize || region.Bounds().Dy() != g.tileSize {
		tile := image.NewRGBA(image.Rect(0, 0, g.tileSize, g.tileSize))

		// Choose appropriate interpolation method
		var scaler draw.Scaler
		if g.isMask {
			scaler = draw.NearestNeighbor
		} else {
			scaler = draw.BiLinear
		}

		scaler.Scale(tile, tile.Bounds(), region, region.Bounds(), draw.Over, nil)
		return tile, nil
	}

	return region, nil
}

// GetTileBounds returns the maximum tile coordinates (columns and rows) at zoom level z.
// Returns (maxX, maxY) representing the highest valid tile indices.
func (g *XYZTileGenerator) GetTileBounds(z int) (int, int) {
	// Calculate theoretical dimensions at this zoom level
	levelWidth := int(math.Ceil(float64(g.level0Dims[0]) / g.xyzDownsamples[z]))
	levelHeight := int(math.Ceil(float64(g.level0Dims[1]) / g.xyzDownsamples[z]))

	// Calculate max tile indices at this level
	maxX := int(math.Ceil(float64(levelWidth)/float64(g.tileSize))) - 1
	maxY := int(math.Ceil(float64(levelHeight)/float64(g.tileSize))) - 1

	return maxX, maxY
}

// GetMetadata returns metadata about the slide and tile pyramid.
func (g *XYZTileGenerator) GetMetadata() (map[string]interface{}, error) {
	properties, err := g.slide.Properties()
	if err != nil {
		return nil, fmt.Errorf("failed to get slide properties: %w", err)
	}

	return map[string]interface{}{
		"width":      g.level0Dims[0],
		"height":     g.level0Dims[1],
		"tileSize":   g.tileSize,
		"minLevel":   0,
		"maxLevel":   g.maxZoomLevel,
		"properties": properties,
	}, nil
}

// createBlankTile creates an empty tile with the background color.
func (g *XYZTileGenerator) createBlankTile() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, g.tileSize, g.tileSize))
	draw.Draw(img, img.Bounds(), &image.Uniform{g.backgroundColor}, image.Point{}, draw.Src)
	return img
}

// parseHexColor parses a hex color string (like "#FFFFFF") to a color.RGBA.
func parseHexColor(hexColor string) color.Color {
	// TODO: Make an error when it fails
	if len(hexColor) == 0 {
		return nil
	}

	// Remove the leading '#' if present
	if hexColor[0] == '#' {
		hexColor = hexColor[1:]
	}

	// Check for valid length
	if len(hexColor) != 6 {
		return nil
	}

	decoded, err := hex.DecodeString(hexColor)
	if err != nil || len(decoded) != 3 {
		return nil
	}

	return color.RGBA{decoded[0], decoded[1], decoded[2], 255}
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
