// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package slides

import (
	"encoding/hex"
	"fmt"
	"image/color"
	"math"
	"math/bits"
	"time"

	"aifo.dev/aifo/fastslide_go/fastslide"
)

// levelMapping stores pre-computed mapping information for each XYZ zoom level.
type levelMapping struct {
	SlideLevel    int  // which slide level to read from
	IsDirectMatch bool // whether this is a direct power-of-2 match
	ShiftDiff     int  // = (maxZoom-z) - slideLevelShifts[slideLevel]
}

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

	// Power-of-2 optimization for tile size
	tileSizeShift uint // log2(tileSize) since tileSize is always power of 2

	// Background color for empty regions
	backgroundColor color.Color

	// If the slide is a mask (affects resampling)
	isMask bool

	// XYZ pyramid properties
	maxZoomLevel    int
	zoomLevelsCount int
	zoomMaps        []levelMapping // Pre-computed mapping for each XYZ zoom level

	// Pre-computed: slide level -> downsample exponent (i.e. LevelDownsample = 1<<slideLevelShifts[i])
	slideLevelShifts []uint
}

// NewXYZTileGenerator creates a new XYZ tile generator for the given slide.
func NewXYZTileGenerator(slide Slide, backgroundColor string, isMask bool) (*XYZTileGenerator, error) {
	// Parse background color
	bg := parseHexColor(backgroundColor)
	if bg == nil {
		// Default to white if color parsing fails
		bg = color.RGBA{255, 255, 255, 255}
	}

	// Fixed tile size - must be power of 2
	tileSize := 256

	// Check if tile size is a power of 2
	if tileSize <= 0 || (tileSize&(tileSize-1)) != 0 {
		return nil, fmt.Errorf("tile size must be a power of 2, got %d", tileSize)
	}

	// Calculate log2(tileSize) using standard library
	tileSizeShift := uint(bits.TrailingZeros(uint(tileSize)))

	// Create the generator
	generator := &XYZTileGenerator{
		slide:           slide,
		tileSize:        tileSize,
		tileSizeShift:   tileSizeShift,
		backgroundColor: bg,
		isMask:          isMask,
	}

	// Initialize XYZ pyramid properties
	var err error

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
	dims, err := g.slide.LargestLevelDimensions()
	if err != nil {
		return 10
	}
	// pick the larger side
	maxDim := dims[0]
	if dims[1] > maxDim {
		maxDim = dims[1]
	}
	// how many tiles at z=0, rounded up:
	n := (maxDim + g.tileSize - 1) >> g.tileSizeShift
	if n <= 1 {
		return 0
	}
	// ceil(log2(n)) == bits.Len(n-1)
	return bits.Len(uint(n - 1))
}

// getXYZLevelDimensions calculates the theoretical dimensions at the given XYZ zoom level.
func (g *XYZTileGenerator) getXYZLevelDimensions(z int) (int, int, error) {
	level0Dims, err := g.slide.LargestLevelDimensions()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get largest level dimensions: %w", err)
	}
	xyzDownsample := 1 << uint(g.maxZoomLevel-z)
	// Use integer ceiling division: ceil(a/b) = (a + b - 1) / b
	levelWidth := (level0Dims[0] + xyzDownsample - 1) / xyzDownsample
	levelHeight := (level0Dims[1] + xyzDownsample - 1) / xyzDownsample
	return levelWidth, levelHeight, nil
}

// initLevelMappings initializes optimized mappings between XYZ zoom levels and slide levels.
func (g *XYZTileGenerator) initLevelMappings() error {
	// Get slide level count and downsamples to find direct mappings
	levelCount, err := g.slide.LevelCount()
	if err != nil {
		return fmt.Errorf("failed to get slide level count: %w", err)
	}

	// Pre-compute downsample exponents for each slide level
	g.slideLevelShifts = make([]uint, levelCount)
	for lvl := 0; lvl < levelCount; lvl++ {
		ds, err := g.slide.LevelDownsample(lvl)
		if err != nil {
			continue
		}

		// Round to nearest integer
		dsInt := uint(math.Round(ds))
		if dsInt == 0 {
			continue
		}

		// Check if dsInt is a power of 2 (has exactly one bit set)
		if dsInt&(dsInt-1) != 0 {
			// Not a power of 2, skip this level
			continue
		}

		// Check if original downsample is within 1% of the rounded power of 2
		diff := math.Abs(ds - float64(dsInt))
		tolerance := ds * 0.01 // 1%
		if diff > tolerance {
			// Skip levels that aren't close enough to a power of 2
			continue
		}

		// Use bits.Len to get the exponent directly: since dsInt == 1<<exp
		g.slideLevelShifts[lvl] = uint(bits.Len(dsInt) - 1)
	}

	// Create unified mapping with pre-computed shift differences and direct match flags
	g.zoomMaps = make([]levelMapping, g.zoomLevelsCount)

	for z := 0; z < g.zoomLevelsCount; z++ {
		downsampleInt := 1 << uint(g.maxZoomLevel-z)
		xyzDownsample := float64(downsampleInt)
		var slideLevel int
		isDirectMatch := false

		// Check each slide level for direct mapping opportunity
		for level := 0; level < levelCount; level++ {
			// Skip levels that don't have valid power-of-2 exponents
			if g.slideLevelShifts[level] == 0 && level != 0 {
				continue // Level 0 should have shift 0, others shouldn't
			}

			slideDownsample, err := g.slide.LevelDownsample(level)
			if err != nil {
				continue // Skip this level on error
			}

			// Check if slide downsample is within 1% of XYZ downsample
			diff := math.Abs(slideDownsample-xyzDownsample) / xyzDownsample
			if diff <= 0.01 { // Within 1%
				// fmt.Printf("Direct match found: zoom %d (downsample %.1f) -> slide level %d (downsample %.1f)\n",
				// 	z, xyzDownsample, level, slideDownsample)
				slideLevel = level
				isDirectMatch = true
				break
			}
		}

		// If no direct mapping found, use BestLevelForDownsample as fallback
		if !isDirectMatch {
			fallbackLevel, err := g.slide.BestLevelForDownsample(xyzDownsample)
			if err != nil {
				return fmt.Errorf("failed to find best level for downsample (zoom %d): %w", z, err)
			}

			// fallbackDownsample, _ := g.slide.LevelDownsample(fallbackLevel)
			// fmt.Printf("No direct match for zoom %d (downsample %.1f), using fallback level %d (downsample %.1f)\n",
			// 	z, xyzDownsample, fallbackLevel, fallbackDownsample)

			// Validate that the fallback level has a valid power-of-2 exponent
			if g.slideLevelShifts[fallbackLevel] == 0 && fallbackLevel != 0 {
				// Find the closest valid level
				bestLevel := 0
				downsampleInt := 1 << uint(g.maxZoomLevel-z)
				xyzDownsampleFloat := float64(downsampleInt)
				bestDiff := math.Abs(xyzDownsampleFloat - 1.0) // Level 0 has downsample 1

				for level := 1; level < levelCount; level++ {
					if g.slideLevelShifts[level] == 0 {
						continue // Skip invalid levels
					}

					levelDownsample := float64(uint(1) << g.slideLevelShifts[level])
					diff := math.Abs(xyzDownsampleFloat - levelDownsample)
					if diff < bestDiff {
						bestDiff = diff
						bestLevel = level
					}
				}
				bestDownsample, _ := g.slide.LevelDownsample(bestLevel)
				fmt.Printf("Fallback level %d invalid, using closest valid level %d (downsample %.1f)\n",
					fallbackLevel, bestLevel, bestDownsample)
				fallbackLevel = bestLevel
			}

			slideLevel = fallbackLevel
		}

		// Pre-compute shift difference for direct use in GetTile
		xyzExp := uint(g.maxZoomLevel - z)
		slideExp := g.slideLevelShifts[slideLevel]
		shiftDiff := int(xyzExp) - int(slideExp)

		g.zoomMaps[z] = levelMapping{
			SlideLevel:    slideLevel,
			IsDirectMatch: isDirectMatch,
			ShiftDiff:     shiftDiff,
		}
	}

	// Sanity check: validate that bit-shift operations work correctly for all zoom levels
	for z, m := range g.zoomMaps {
		var readSize int
		if m.ShiftDiff >= 0 {
			// Zooming in: should be able to shift left then right to get back to tileSize
			readSize = g.tileSize << m.ShiftDiff
			if readSize>>m.ShiftDiff != g.tileSize {
				panic(fmt.Sprintf("zoom mapping[%d] shiftDiff=%d: bit-shift validation failed for positive shift", z, m.ShiftDiff))
			}
		} else {
			// Downsampling: check that we don't lose too much precision
			negShift := -m.ShiftDiff
			readSize = g.tileSize >> negShift
			// For downsampling, we expect Lanczos to restore to tileSize, so we mainly check for reasonable bounds
			if readSize <= 0 || readSize > g.tileSize {
				panic(fmt.Sprintf("zoom mapping[%d] shiftDiff=%d: downsampled size %d is out of reasonable bounds", z, m.ShiftDiff, readSize))
			}
		}
	}

	// Log the final mapping summary
	// fmt.Printf("\nXYZ Tile Generator initialized: max zoom %d, %d zoom levels\n", g.maxZoomLevel, g.zoomLevelsCount)
	// fmt.Printf("Zoom level mappings:\n")
	// for z, m := range g.zoomMaps {
	// 	xyzDownsample := 1 << uint(g.maxZoomLevel-z)
	// 	slideDownsample, _ := g.slide.LevelDownsample(m.SlideLevel)
	// 	directStr := ""
	// 	if !m.IsDirectMatch {
	// 		directStr = " (resampling)"
	// 	}
	// 	fmt.Printf("  z=%d -> slide level %d (downsample %.1f->%.1f, shift_diff=%d)%s\n",
	// 		z, m.SlideLevel, float64(xyzDownsample), slideDownsample, m.ShiftDiff, directStr)
	// }
	// fmt.Println()

	return nil
}

// GetTile retrieves a tile for the specified XYZ coordinates.
func (g *XYZTileGenerator) GetTile(z, x, y int) (*fastslide.Image, error) {
	// Validate zoom level
	if z < 0 || z > g.maxZoomLevel {
		return nil, fmt.Errorf("invalid zoom level: %d (valid range: 0-%d)", z, g.maxZoomLevel)
	}

	// Calculate theoretical dimensions at this zoom level
	levelWidth, levelHeight, err := g.getXYZLevelDimensions(z)
	if err != nil {
		return nil, fmt.Errorf("failed to get XYZ level dimensions: %w", err)
	}

	// Calculate max tile indices at this level for bounds checking
	// Use bit shift for power-of-2 ceiling division: ceil(a/b) = (a + b - 1) >> shift
	maxX := (levelWidth+g.tileSize-1)>>g.tileSizeShift - 1
	maxY := (levelHeight+g.tileSize-1)>>g.tileSizeShift - 1

	// Return an error if tile coordinates are out of bounds
	if x < 0 || y < 0 || x > maxX || y > maxY {
		return nil, fmt.Errorf("tile coordinates out of bounds: x=%d, y=%d, z=%d (valid range: x=0-%d, y=0-%d)",
			x, y, z, maxX, maxY)
	}

	// Get pre-computed mapping for this zoom level
	m := g.zoomMaps[z]
	slideLevel := m.SlideLevel

	// Safety check: ensure we have a valid slide level with power-of-2 downsample
	slideExp := g.slideLevelShifts[slideLevel]
	if slideExp == 0 && slideLevel != 0 {
		return nil, fmt.Errorf("invalid slide level %d: not a power-of-2 downsample", slideLevel)
	}

	// Base pixel offset in the slide's native coords at level 0
	baseOffsetX := x * g.tileSize
	baseOffsetY := y * g.tileSize

	// Use pre-computed shift difference - no runtime calculations
	var levelNativeX, levelNativeY, readWidth, readHeight int
	if m.ShiftDiff >= 0 {
		// Zooming in: multiply by 2^ShiftDiff
		levelNativeX = baseOffsetX << m.ShiftDiff
		levelNativeY = baseOffsetY << m.ShiftDiff
		// How many pixels we need from the slide to cover one browser-tile
		readWidth = g.tileSize << m.ShiftDiff
		readHeight = g.tileSize << m.ShiftDiff
	} else {
		// Downsampling: divide by 2^(-ShiftDiff)
		negShift := -m.ShiftDiff
		levelNativeX = baseOffsetX >> negShift
		levelNativeY = baseOffsetY >> negShift
		// Likewise for size
		readWidth = g.tileSize >> negShift
		readHeight = g.tileSize >> negShift
	}

	// Get slide level dimensions and clamp the read region
	slideLevelDims, err := g.slide.LevelDimensions(slideLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to get slide level dimensions: %w", err)
	}

	// Use clampRegion to handle bounds checking and clamping in one step
	clampedWidth, clampedHeight, ok := clampRegion(levelNativeX, levelNativeY, readWidth, readHeight, slideLevelDims[0], slideLevelDims[1])
	if !ok {
		return nil, fmt.Errorf("tile region is outside slide bounds")
	}

	readWidth = clampedWidth
	readHeight = clampedHeight

	// Determine if this is a border tile that needs padding
	var expectedReadWidth, expectedReadHeight int
	if m.ShiftDiff >= 0 {
		// Zooming in: we expect to read more pixels
		expectedReadWidth = g.tileSize << m.ShiftDiff
		expectedReadHeight = g.tileSize << m.ShiftDiff
	} else {
		// Downsampling: we expect to read fewer pixels
		negShift := -m.ShiftDiff
		expectedReadWidth = g.tileSize >> negShift
		expectedReadHeight = g.tileSize >> negShift
	}
	isBorderTile := (readWidth < expectedReadWidth || readHeight < expectedReadHeight)

	// Determine target size after resampling
	var targetWidth, targetHeight int
	if m.ShiftDiff >= 0 {
		// Zooming in: target is always tileSize
		targetWidth = g.tileSize
		targetHeight = g.tileSize
	} else {
		// Downsampling: target might be smaller
		negShift := -m.ShiftDiff
		targetWidth = g.tileSize >> negShift
		targetHeight = g.tileSize >> negShift
	}

	// Log tile request details for debugging
	// fmt.Printf("Tile request z=%d, x=%d, y=%d: using slide level %d (direct_match=%v, shift_diff=%d)\n",
	// z, x, y, slideLevel, m.IsDirectMatch, m.ShiftDiff)
	// fmt.Printf("  Expected read: %dx%d, actual read: %dx%d, target: %dx%d\n",
	// expectedReadWidth, expectedReadHeight, readWidth, readHeight, targetWidth, targetHeight)
	// if isBorderTile {
	// 	fmt.Printf("  Border tile detected: read region smaller than expected\n")
	// }
	if !m.IsDirectMatch {
		// Show what downsample we need vs what the slide level provides
		// xyzDownsample := 1 << uint(g.maxZoomLevel-z)
		// slideDownsample, _ := g.slide.LevelDownsample(slideLevel)
		// fmt.Printf("  Level resampling needed: zoom %d wants downsample %.1f, slide level %d provides %.1f\n",
		// z, float64(xyzDownsample), slideLevel, slideDownsample)

		// Show available slide levels for reference
		// g.logAvailableLevels()
	}

	// Read the region from the slide
	var region *fastslide.Image

	// readStart := time.Now()
	region, err = g.readRegion(
		levelNativeX, levelNativeY,
		slideLevel,
		readWidth, readHeight,
		expectedReadWidth, expectedReadHeight,
		isBorderTile,
		!m.IsDirectMatch,
		targetWidth, targetHeight,
	)
	// readDuration := time.Since(readStart)
	// fmt.Printf("Total readRegion took: %v\n", readDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to read region: %w", err)
	}

	return region, nil
}

// GetTileBounds returns the maximum tile coordinates (columns and rows) at zoom level z.
// Returns (maxX, maxY) representing the highest valid tile indices.
func (g *XYZTileGenerator) GetTileBounds(z int) (int, int) {
	// Calculate theoretical dimensions at this zoom level
	levelWidth, levelHeight, err := g.getXYZLevelDimensions(z)
	if err != nil {
		// Return 0,0 as a fallback on error
		return 0, 0
	}

	// Calculate max tile indices at this level
	// Use bit shift for power-of-2 ceiling division: ceil(a/b) = (a + b - 1) >> shift
	maxX := (levelWidth+g.tileSize-1)>>g.tileSizeShift - 1
	maxY := (levelHeight+g.tileSize-1)>>g.tileSizeShift - 1

	return maxX, maxY
}

// GetMetadata returns metadata about the slide and tile pyramid.
func (g *XYZTileGenerator) GetMetadata() (map[string]interface{}, error) {
	properties, err := g.slide.Properties()
	if err != nil {
		return nil, fmt.Errorf("failed to get slide properties: %w", err)
	}

	level0Dims, err := g.slide.LargestLevelDimensions()
	if err != nil {
		return nil, fmt.Errorf("failed to get largest level dimensions: %w", err)
	}

	return map[string]interface{}{
		"width":      level0Dims[0],
		"height":     level0Dims[1],
		"tileSize":   g.tileSize,
		"minLevel":   0,
		"maxLevel":   g.maxZoomLevel,
		"properties": properties,
	}, nil
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

// readRegion reads a region from the slide and returns a *fastslide.Image.
// It handles border tile padding and level resampling as separate, clearly defined steps.
func (g *XYZTileGenerator) readRegion(levelNativeX, levelNativeY, slideLevel, readWidth, readHeight, expectedWidth, expectedHeight int,
	isBorderTile, needsResampling bool, targetWidth, targetHeight int,
) (*fastslide.Image, error) {
	// Step 1: Read the raw region from the slide
	// readStart := time.Now()
	region, err := g.slide.ReadRegionAsFastslideImage(levelNativeX, levelNativeY, slideLevel, readWidth, readHeight)
	// readDuration := time.Since(readStart)
	// fmt.Printf("ReadRegionAsFastslideImage took: %v\n", readDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to read region from slide: %w", err)
	}

	// Step 2: Handle border tile padding if needed
	if isBorderTile {
		// fmt.Printf("    Padding border tile: from %dx%d to %dx%d\n",
		// readWidth, readHeight, expectedWidth, expectedHeight)
		paddedRegion, err := g.padRegionForBorderTile(region, readWidth, readHeight, expectedWidth, expectedHeight)
		if err != nil {
			region.Close()
			return nil, fmt.Errorf("failed to pad border tile: %w", err)
		}
		region.Close() // Clean up original
		region = paddedRegion
	}

	// Step 3: Handle level resampling if needed
	if needsResampling {
		// currentWidth := int(region.Width())
		// currentHeight := int(region.Height())
		// fmt.Printf("    Resampling: from %dx%d to %dx%d (level %d -> target)\n",
		// currentWidth, currentHeight, targetWidth, targetHeight, slideLevel)
		resampledRegion, err := g.resampleRegionForLevel(region, targetWidth, targetHeight, slideLevel)
		if err != nil {
			region.Close()
			return nil, fmt.Errorf("failed to resample region: %w", err)
		}
		region.Close() // Clean up original
		region = resampledRegion
	}

	return region, nil
}

// padRegionForBorderTile pads a region that was read smaller than expected due to slide boundaries.
func (g *XYZTileGenerator) padRegionForBorderTile(region *fastslide.Image, actualWidth, actualHeight, expectedWidth, expectedHeight int) (*fastslide.Image, error) {
	// If the region is already the expected size, no padding needed
	if actualWidth == expectedWidth && actualHeight == expectedHeight {
		return region, nil
	}

	// Create a blank canvas at the expected size
	paddedRegion, err := fastslide.CreateBlankImage(fastslide.ImageDimensions{
		Width:  expectedWidth,
		Height: expectedHeight,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create padded image canvas: %w", err)
	}

	// TODO: Fill with background color when fastslide.Image supports it
	// For now, the blank image defaults to black/transparent

	// Paste the actual region at the top-left
	if err := paddedRegion.PasteSimple(region, 0, 0); err != nil {
		paddedRegion.Close()
		return nil, fmt.Errorf("failed to paste region into padded canvas: %w", err)
	}

	return paddedRegion, nil
}

// resampleRegionForLevel resamples a region to the target size for level matching.
func (g *XYZTileGenerator) resampleRegionForLevel(region *fastslide.Image, targetWidth, targetHeight, slideLevel int) (*fastslide.Image, error) {
	currentWidth := int(region.Width())
	currentHeight := int(region.Height())

	// If already at target size, no resampling needed
	if currentWidth == targetWidth && currentHeight == targetHeight {
		return region, nil
	}

	resampleStart := time.Now()

	// Case 1: Power-of-2 downsampling - use efficient AverageResample
	if currentWidth > targetWidth && currentHeight > targetHeight {
		widthFactor := currentWidth / targetWidth
		heightFactor := currentHeight / targetHeight

		// Use the smaller factor to ensure we don't oversample
		factor := min(widthFactor, heightFactor)
		if factor > 1 && currentWidth%factor == 0 && currentHeight%factor == 0 {
			// fmt.Printf("      Using AverageResample with factor %d (downsampling)\n", factor)
			resampled, err := region.AverageResample(factor)
			if err != nil {
				return nil, fmt.Errorf("failed to average resample image: %w", err)
			}

			// resampleDuration := time.Since(resampleStart)
			// fmt.Printf("      AverageResample(factor=%d) completed in %v\n", factor, resampleDuration)
			return resampled, nil
		}
	}

	// Case 2: General resampling using Paste with scaling
	// fmt.Printf("      Using Paste scaling (general resampling)\n")
	target, err := fastslide.CreateBlankImage(fastslide.ImageDimensions{
		Width:  targetWidth,
		Height: targetHeight,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create resampling target: %w", err)
	}

	// TODO: Fill with background color when fastslide.Image supports it

	// Scale the entire source region to fit the target
	err = target.Paste(region, 0, 0, 0, 0, region.Width(), region.Height())
	if err != nil {
		target.Close()
		return nil, fmt.Errorf("failed to paste and scale region: %w", err)
	}

	resampleDuration := time.Since(resampleStart)
	fmt.Printf("      Paste scaling completed in %v\n", resampleDuration)
	return target, nil
}

// clampRegion clamps a read region to fit within the given dimensions and validates bounds.
// Returns the clamped width and height, and whether the region is valid (not completely outside bounds).
func clampRegion(x, y, width, height int, maxWidth, maxHeight int) (clampedWidth, clampedHeight int, ok bool) {
	// Check if we're completely outside the bounds
	if x >= maxWidth || y >= maxHeight || x < 0 || y < 0 {
		return 0, 0, false
	}

	// Calculate available space from the starting position
	availableWidth := maxWidth - x
	availableHeight := maxHeight - y

	// If no space available, region is outside bounds
	if availableWidth <= 0 || availableHeight <= 0 {
		return 0, 0, false
	}

	// Clamp to available space
	clampedWidth = min(width, availableWidth)
	clampedHeight = min(height, availableHeight)

	// Final validation - ensure we have a valid region
	if clampedWidth <= 0 || clampedHeight <= 0 {
		return 0, 0, false
	}

	return clampedWidth, clampedHeight, true
}

// logAvailableLevels logs information about available slide levels for debugging.
func (g *XYZTileGenerator) logAvailableLevels() {
	levelCount, err := g.slide.LevelCount()
	if err != nil {
		// fmt.Printf("  Available levels: error getting count - %v\n", err)
		return
	}

	// fmt.Printf("  Available slide levels (%d total):\n", levelCount)
	for level := 0; level < levelCount; level++ {
		// downsample, err := g.slide.LevelDownsample(level)
		// if err != nil {
		// 	fmt.Printf("    Level %d: error getting downsample - %v\n", level, err)
		// 	continue
		// }

		// dims, err := g.slide.LevelDimensions(level)
		// if err != nil {
		// fmt.Printf("    Level %d: downsample %.1f, error getting dimensions - %v\n", level, downsample, err)
		// continue
		// }

		// Show if this level has a valid power-of-2 shift
		// shiftValid := ""
		// if g.slideLevelShifts[level] == 0 && level != 0 {
		// 	shiftValid = " (invalid - not power-of-2)"
		// } else {
		// 	// shiftValid = fmt.Sprintf(" (shift=%d)", g.slideLevelShifts[level])
		// }

		// fmt.Printf("    Level %d: downsample %.1f, %dx%d%s\n",
		// 	level, downsample, dims[0], dims[1], shiftValid)
	}
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
