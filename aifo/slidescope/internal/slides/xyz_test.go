// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package slides

import (
	"image"
	"image/color"
	"testing"
)

// MockSlide implements the Slide interface for testing purposes.
type MockSlide struct {
	levelCount          int
	level0Dims          [2]int
	levelDims           [][2]int
	downsamples         []float64
	properties          map[string]string
	readRegionCallCount int
	readRegionFunc      func(x, y, level, w, h int) (image.Image, error)
}

func (m *MockSlide) Close() {}

func (m *MockSlide) LevelCount() (int, error) {
	return m.levelCount, nil
}

func (m *MockSlide) LargestLevelDimensions() ([2]int, error) {
	return m.level0Dims, nil
}

func (m *MockSlide) LevelDimensions(level int) ([2]int, error) {
	if level < 0 || level >= len(m.levelDims) {
		return [2]int{0, 0}, nil
	}
	return m.levelDims[level], nil
}

func (m *MockSlide) LevelDownsample(level int) (float64, error) {
	if level < 0 || level >= len(m.downsamples) {
		return 1.0, nil
	}
	return m.downsamples[level], nil
}

func (m *MockSlide) LevelDownsamples() ([]float64, error) {
	result := make([]float64, len(m.downsamples))
	copy(result, m.downsamples)
	return result, nil
}

func (m *MockSlide) BestLevelForDownsample(downsample float64) (int, error) {
	// Simple implementation for testing
	for i, ds := range m.downsamples {
		if ds >= downsample {
			return i, nil
		}
	}
	return len(m.downsamples) - 1, nil
}

func (m *MockSlide) ReadRegion(x, y, level, w, h int) (image.Image, error) {
	m.readRegionCallCount++
	if m.readRegionFunc != nil {
		return m.readRegionFunc(x, y, level, w, h)
	}
	// Default implementation returns a blank image
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	return img, nil
}

func (m *MockSlide) PropertyValue(name string) (string, error) {
	return m.properties[name], nil
}

func (m *MockSlide) Properties() (map[string]string, error) {
	result := make(map[string]string, len(m.properties))
	for k, v := range m.properties {
		result[k] = v
	}
	return result, nil
}

// Helper function to create a mock slide for testing
func createMockSlide() *MockSlide {
	return &MockSlide{
		levelCount: 3,
		level0Dims: [2]int{10000, 8000},
		levelDims: [][2]int{
			{10000, 8000},
			{5000, 4000},
			{2500, 2000},
		},
		downsamples: []float64{1.0, 2.0, 4.0},
		properties: map[string]string{
			PropObjectivePower:  "40",
			PropBackgroundColor: "#FFFFFF",
		},
	}
}

func TestNewXYZTileGenerator(t *testing.T) {
	slide := createMockSlide()

	// Test with default background color
	generator, err := NewXYZTileGenerator(slide, "", false)
	if err != nil {
		t.Fatalf("Failed to create XYZTileGenerator: %v", err)
	}

	// Check initialization
	if generator.slide != slide {
		t.Error("Slide not properly set")
	}

	if generator.tileSize != 512 {
		t.Errorf("Expected tile size 512, got %d", generator.tileSize)
	}

	// Check level calculations
	expectedMaxZoom := 5 // log2(10000/512) = 4.28, ceil to 5
	if generator.maxZoomLevel != expectedMaxZoom {
		t.Errorf("Expected max zoom level %d, got %d", expectedMaxZoom, generator.maxZoomLevel)
	}

	// Test with custom background color
	customColor := "#FF0000"
	generator, err = NewXYZTileGenerator(slide, customColor, false)
	if err != nil {
		t.Fatalf("Failed to create XYZTileGenerator with custom color: %v", err)
	}

	// The generator uses the background color from the slide properties,
	// not the custom color provided, unless overridden explicitly
	// So in this case it should still be white
	// Check if the slide's background color property was used
	propColor, err := slide.PropertyValue(PropBackgroundColor)
	if err != nil {
		t.Fatalf("Error retrieving background color property: %v", err)
	}
	if propColor != "#FFFFFF" {
		t.Errorf("Slide background color property should be #FFFFFF, got %s", propColor)
	}

	// Note: In the actual implementation, the slide property has priority
	// Let's explicitly verify that behavior
	expectedColor := color.RGBA{255, 255, 255, 255} // White from slide property
	if bg, ok := generator.backgroundColor.(color.RGBA); !ok || bg != expectedColor {
		t.Errorf("Expected background color %v, got %v", expectedColor, bg)
	}

	// Create a slide without background color to verify that the custom color is used
	slideNoColor := createMockSlide()
	delete(slideNoColor.properties, PropBackgroundColor)

	generatorCustomColor, err := NewXYZTileGenerator(slideNoColor, customColor, false)
	if err != nil {
		t.Fatalf("Failed to create XYZTileGenerator with custom color: %v", err)
	}

	// Now the custom color should be used
	expectedCustomColor := color.RGBA{255, 0, 0, 255}
	if bg, ok := generatorCustomColor.backgroundColor.(color.RGBA); !ok || bg != expectedCustomColor {
		t.Errorf("Expected custom background color %v, got %v", expectedCustomColor, bg)
	}
}

func TestGetTileBounds(t *testing.T) {
	slide := createMockSlide()
	generator, _ := NewXYZTileGenerator(slide, "", false)

	// Calculate expected values using the same logic as the implementation
	testCases := []struct {
		zoomLevel    int
		expectedMaxX int
		expectedMaxY int
	}{
		{0, 0, 0},   // Most zoomed out - just 1 tile
		{2, 2, 1},   // Middle zoom level
		{5, 19, 15}, // Most detailed level
	}

	for _, tc := range testCases {
		maxX, maxY := generator.GetTileBounds(tc.zoomLevel)

		if maxX != tc.expectedMaxX || maxY != tc.expectedMaxY {
			t.Errorf("For zoom level %d, expected bounds (%d,%d), got (%d,%d)",
				tc.zoomLevel, tc.expectedMaxX, tc.expectedMaxY, maxX, maxY)
		}
	}
}

func TestGetTile(t *testing.T) {
	slide := createMockSlide()
	generator, _ := NewXYZTileGenerator(slide, "", false)

	// Test invalid zoom level
	_, err := generator.GetTile(-1, 0, 0)
	if err == nil {
		t.Error("Expected error for invalid zoom level, got nil")
	}

	// Test invalid tile coordinates
	_, err = generator.GetTile(5, 100, 100)
	if err == nil {
		t.Error("Expected error for out-of-bounds coordinates, got nil")
	}

	// Test valid tile retrieval
	slide.readRegionFunc = func(x, y, level, w, h int) (image.Image, error) {
		// Return a colored image to verify it's properly processed
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		return img, nil
	}

	img, err := generator.GetTile(5, 10, 8)
	if err != nil {
		t.Fatalf("Failed to get valid tile: %v", err)
	}

	// Check that we got a correctly sized tile
	if img.Bounds().Dx() != generator.tileSize || img.Bounds().Dy() != generator.tileSize {
		t.Errorf("Expected tile size %dx%d, got %dx%d",
			generator.tileSize, generator.tileSize,
			img.Bounds().Dx(), img.Bounds().Dy())
	}

	// Verify read region was called once
	if slide.readRegionCallCount != 1 {
		t.Errorf("Expected 1 call to ReadRegion, got %d", slide.readRegionCallCount)
	}
}

func TestParseHexColor(t *testing.T) {
	testCases := []struct {
		input    string
		expected color.RGBA
		valid    bool
	}{
		{"#FFFFFF", color.RGBA{255, 255, 255, 255}, true},
		{"FFFFFF", color.RGBA{255, 255, 255, 255}, true},
		{"#FF0000", color.RGBA{255, 0, 0, 255}, true},
		{"00FF00", color.RGBA{0, 255, 0, 255}, true},
		{"#0000FF", color.RGBA{0, 0, 255, 255}, true},
		{"", color.RGBA{}, false},
		{"#FFF", color.RGBA{}, false},   // Invalid length
		{"GGGGGG", color.RGBA{}, false}, // Invalid characters
	}

	for _, tc := range testCases {
		result := parseHexColor(tc.input)

		if tc.valid {
			if result == nil {
				t.Errorf("For input %s, expected valid color, got nil", tc.input)
				continue
			}

			rgba, ok := result.(color.RGBA)
			if !ok {
				t.Errorf("For input %s, expected color.RGBA, got %T", tc.input, result)
				continue
			}

			if rgba != tc.expected {
				t.Errorf("For input %s, expected %v, got %v", tc.input, tc.expected, rgba)
			}
		} else {
			if result != nil {
				t.Errorf("For input %s, expected nil, got %v", tc.input, result)
			}
		}
	}
}
