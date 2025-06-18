// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package slides

import (
	"image"
	"testing"
)

// Define a mock implementation of the Slide interface for testing
type mockSlide struct {
	levelCount          int
	level0Dims          [2]int
	levelDims           [][2]int
	downsamples         []float64
	properties          map[string]string
	readRegionCallCount int
	readRegionFunc      func(x, y, level, w, h int) (image.Image, error)
	closed              bool
}

func (m *mockSlide) Close() {
	m.closed = true
}

func (m *mockSlide) LevelCount() (int, error) {
	return m.levelCount, nil
}

func (m *mockSlide) LargestLevelDimensions() ([2]int, error) {
	return m.level0Dims, nil
}

func (m *mockSlide) LevelDimensions(level int) ([2]int, error) {
	if level < 0 || level >= len(m.levelDims) {
		return [2]int{0, 0}, nil
	}
	return m.levelDims[level], nil
}

func (m *mockSlide) LevelDownsample(level int) (float64, error) {
	if level < 0 || level >= len(m.downsamples) {
		return 1.0, nil
	}
	return m.downsamples[level], nil
}

func (m *mockSlide) LevelDownsamples() ([]float64, error) {
	result := make([]float64, len(m.downsamples))
	copy(result, m.downsamples)
	return result, nil
}

func (m *mockSlide) BestLevelForDownsample(downsample float64) (int, error) {
	// Simple implementation for testing
	for i, ds := range m.downsamples {
		if ds >= downsample {
			return i, nil
		}
	}
	return len(m.downsamples) - 1, nil
}

func (m *mockSlide) ReadRegion(x, y, level, w, h int) (image.Image, error) {
	m.readRegionCallCount++
	if m.readRegionFunc != nil {
		return m.readRegionFunc(x, y, level, w, h)
	}
	// Default implementation returns a blank image
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	return img, nil
}

func (m *mockSlide) PropertyValue(name string) (string, error) {
	return m.properties[name], nil
}

func (m *mockSlide) Properties() (map[string]string, error) {
	result := make(map[string]string, len(m.properties))
	for k, v := range m.properties {
		result[k] = v
	}
	return result, nil
}

// createStandardMockSlide creates a mock slide with standard test data
func createStandardMockSlide() *mockSlide {
	return &mockSlide{
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
			"openslide.vendor":  "Test Vendor",
			"openslide.mpp-x":   "0.25",
			"openslide.mpp-y":   "0.25",
		},
		readRegionFunc: func(x, y, level, w, h int) (image.Image, error) {
			return image.NewRGBA(image.Rect(0, 0, w, h)), nil
		},
	}
}

// Test the mock Slide implementation
func TestMockSlide(t *testing.T) {
	mock := createStandardMockSlide()

	// Test basic methods
	levelCount, err := mock.LevelCount()
	if err != nil {
		t.Errorf("Unexpected error from LevelCount: %v", err)
	}
	if levelCount != 3 {
		t.Errorf("Expected level count 3, got %d", levelCount)
	}

	dims, err := mock.LargestLevelDimensions()
	if err != nil {
		t.Errorf("Unexpected error from LargestLevelDimensions: %v", err)
	}
	if dims != [2]int{10000, 8000} {
		t.Errorf("Expected dimensions [10000, 8000], got %v", dims)
	}

	ds, err := mock.LevelDownsample(1)
	if err != nil {
		t.Errorf("Unexpected error from LevelDownsample: %v", err)
	}
	if ds != 2.0 {
		t.Errorf("Expected downsample 2.0, got %f", ds)
	}

	// Test ReadRegion
	img, err := mock.ReadRegion(0, 0, 0, 100, 100)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if img == nil {
		t.Error("Expected non-nil image")
	}
	if mock.readRegionCallCount != 1 {
		t.Errorf("Expected 1 call to ReadRegion, got %d", mock.readRegionCallCount)
	}

	// Test property access
	val, err := mock.PropertyValue(PropObjectivePower)
	if err != nil {
		t.Errorf("Unexpected error from PropertyValue: %v", err)
	}
	if val != "40" {
		t.Errorf("Expected objective power 40, got %s", val)
	}

	// Test close
	mock.Close()
	if !mock.closed {
		t.Error("Expected Close() to set closed flag")
	}
}

// Create a mock OpenSlideAdapter for testing
type mockOpenSlideAdapter struct {
	mockSlide *mockSlide
}

func newMockOpenSlideAdapter() *mockOpenSlideAdapter {
	return &mockOpenSlideAdapter{
		mockSlide: createStandardMockSlide(),
	}
}

func (m *mockOpenSlideAdapter) Close() {
	m.mockSlide.Close()
}

func (m *mockOpenSlideAdapter) LevelCount() (int, error) {
	return m.mockSlide.LevelCount()
}

func (m *mockOpenSlideAdapter) LargestLevelDimensions() ([2]int, error) {
	return m.mockSlide.LargestLevelDimensions()
}

func (m *mockOpenSlideAdapter) LevelDimensions(level int) ([2]int, error) {
	return m.mockSlide.LevelDimensions(level)
}

func (m *mockOpenSlideAdapter) LevelDownsample(level int) (float64, error) {
	return m.mockSlide.LevelDownsample(level)
}

func (m *mockOpenSlideAdapter) LevelDownsamples() ([]float64, error) {
	return m.mockSlide.LevelDownsamples()
}

func (m *mockOpenSlideAdapter) BestLevelForDownsample(downsample float64) (int, error) {
	return m.mockSlide.BestLevelForDownsample(downsample)
}

func (m *mockOpenSlideAdapter) ReadRegion(x, y, level, w, h int) (image.Image, error) {
	return m.mockSlide.ReadRegion(x, y, level, w, h)
}

func (m *mockOpenSlideAdapter) PropertyValue(name string) (string, error) {
	return m.mockSlide.PropertyValue(name)
}

func (m *mockOpenSlideAdapter) Properties() (map[string]string, error) {
	return m.mockSlide.Properties()
}

// Test the OpenSlideAdapter via a mock implementation
func TestOpenSlideAdapterInterface(t *testing.T) {
	mockAdapter := newMockOpenSlideAdapter()

	// Verify adapter fully implements the Slide interface by using it as a Slide
	var slide Slide = mockAdapter

	// Basic property checks
	levelCount, err := slide.LevelCount()
	if err != nil {
		t.Errorf("Unexpected error from LevelCount: %v", err)
	}
	if levelCount != 3 {
		t.Errorf("Expected level count 3, got %d", levelCount)
	}

	dims, err := slide.LargestLevelDimensions()
	if err != nil {
		t.Errorf("Unexpected error from LargestLevelDimensions: %v", err)
	}
	expectedDims := [2]int{10000, 8000}
	if dims != expectedDims {
		t.Errorf("Expected dimensions %v, got %v", expectedDims, dims)
	}

	levelDims, err := slide.LevelDimensions(1)
	if err != nil {
		t.Errorf("Unexpected error from LevelDimensions: %v", err)
	}
	expectedLevelDims := [2]int{5000, 4000}
	if levelDims != expectedLevelDims {
		t.Errorf("Expected level 1 dimensions %v, got %v", expectedLevelDims, levelDims)
	}

	// Check downsamples
	downsamples, err := slide.LevelDownsamples()
	if err != nil {
		t.Errorf("Unexpected error from LevelDownsamples: %v", err)
	}
	expectedDownsamples := []float64{1.0, 2.0, 4.0}
	if len(downsamples) != len(expectedDownsamples) {
		t.Errorf("Expected %d downsamples, got %d", len(expectedDownsamples), len(downsamples))
	} else {
		for i, ds := range downsamples {
			if ds != expectedDownsamples[i] {
				t.Errorf("Expected downsamples[%d] = %f, got %f", i, expectedDownsamples[i], ds)
			}
		}
	}

	// Test read region with expected dimensions
	img, err := slide.ReadRegion(0, 0, 0, 100, 100)
	if err != nil {
		t.Errorf("ReadRegion failed: %v", err)
	}
	if img.Bounds().Dx() != 100 || img.Bounds().Dy() != 100 {
		t.Errorf("Expected 100x100 image, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

// Create a mock TiffAdapter for testing
type mockTiffAdapter struct {
	mockSlide *mockSlide
}

func newMockTiffAdapter() *mockTiffAdapter {
	return &mockTiffAdapter{
		mockSlide: createStandardMockSlide(),
	}
}

func (m *mockTiffAdapter) Close() {
	m.mockSlide.Close()
}

func (m *mockTiffAdapter) LevelCount() (int, error) {
	return m.mockSlide.LevelCount()
}

func (m *mockTiffAdapter) LargestLevelDimensions() ([2]int, error) {
	return m.mockSlide.LargestLevelDimensions()
}

func (m *mockTiffAdapter) LevelDimensions(level int) ([2]int, error) {
	return m.mockSlide.LevelDimensions(level)
}

func (m *mockTiffAdapter) LevelDownsample(level int) (float64, error) {
	return m.mockSlide.LevelDownsample(level)
}

func (m *mockTiffAdapter) LevelDownsamples() ([]float64, error) {
	return m.mockSlide.LevelDownsamples()
}

func (m *mockTiffAdapter) BestLevelForDownsample(downsample float64) (int, error) {
	return m.mockSlide.BestLevelForDownsample(downsample)
}

func (m *mockTiffAdapter) ReadRegion(x, y, level, w, h int) (image.Image, error) {
	return m.mockSlide.ReadRegion(x, y, level, w, h)
}

func (m *mockTiffAdapter) PropertyValue(name string) (string, error) {
	return m.mockSlide.PropertyValue(name)
}

func (m *mockTiffAdapter) Properties() (map[string]string, error) {
	return m.mockSlide.Properties()
}

// Test the TiffAdapter via a mock implementation
func TestTiffAdapterInterface(t *testing.T) {
	mockAdapter := newMockTiffAdapter()

	// Verify adapter fully implements the Slide interface by using it as a Slide
	var slide Slide = mockAdapter

	// Basic property checks
	levelCount, err := slide.LevelCount()
	if err != nil {
		t.Errorf("Unexpected error from LevelCount: %v", err)
	}
	if levelCount != 3 {
		t.Errorf("Expected level count 3, got %d", levelCount)
	}

	dims, err := slide.LargestLevelDimensions()
	if err != nil {
		t.Errorf("Unexpected error from LargestLevelDimensions: %v", err)
	}
	expectedDims := [2]int{10000, 8000}
	if dims != expectedDims {
		t.Errorf("Expected dimensions %v, got %v", expectedDims, dims)
	}

	// Test property access specific to TIFF adapter
	mppX, err := slide.PropertyValue("openslide.mpp-x")
	if err != nil {
		t.Errorf("Unexpected error from PropertyValue: %v", err)
	}
	if mppX != "0.25" {
		t.Errorf("Expected mpp-x 0.25, got %s", mppX)
	}
}
