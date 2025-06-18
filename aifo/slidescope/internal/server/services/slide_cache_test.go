// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package services

import (
	"image"
	"image/color"
	"testing"
	"time"

	"aifo.dev/aifo/slidescope/internal/slides"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSlide implements slides.Slide interface for testing
type MockSlide struct {
	mock.Mock
}

func (m *MockSlide) Close() {
	m.Called()
}

func (m *MockSlide) LevelCount() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func (m *MockSlide) LargestLevelDimensions() ([2]int, error) {
	args := m.Called()
	return args.Get(0).([2]int), args.Error(1)
}

func (m *MockSlide) LevelDimensions(level int) ([2]int, error) {
	args := m.Called(level)
	return args.Get(0).([2]int), args.Error(1)
}

func (m *MockSlide) LevelDownsample(level int) (float64, error) {
	args := m.Called(level)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockSlide) LevelDownsamples() ([]float64, error) {
	args := m.Called()
	return args.Get(0).([]float64), args.Error(1)
}

func (m *MockSlide) BestLevelForDownsample(downsample float64) (int, error) {
	args := m.Called(downsample)
	return args.Int(0), args.Error(1)
}

func (m *MockSlide) ReadRegion(x, y, level, w, h int) (image.Image, error) {
	args := m.Called(x, y, level, w, h)
	return args.Get(0).(image.Image), args.Error(1)
}

func (m *MockSlide) PropertyValue(name string) (string, error) {
	args := m.Called(name)
	return args.String(0), args.Error(1)
}

func (m *MockSlide) Properties() (map[string]string, error) {
	args := m.Called()
	return args.Get(0).(map[string]string), args.Error(1)
}

// TestNewPyramidCache tests the creation of a new cache
func TestNewPyramidCache(t *testing.T) {
	cache := NewPyramidCache(100, 200, true, "white", false)

	assert.NotNil(t, cache)
	assert.Equal(t, 100, cache.maxGenerators)
	assert.Equal(t, 200, cache.maxResources)
	assert.Equal(t, true, cache.isSlide)
	assert.Equal(t, "white", cache.bgColor)
	assert.Equal(t, false, cache.precise)
}

// TestPyramidCacheStats tests the stats collection
func TestPyramidCacheStats(t *testing.T) {
	cache := NewPyramidCache(10, 20, true, "white", false)

	stats := cache.Stats()
	assert.Equal(t, uint64(0), stats["hits"])
	assert.Equal(t, uint64(0), stats["misses"])
	assert.Equal(t, uint64(0), stats["evicted"])
	assert.Equal(t, uint64(0), stats["created"])
	assert.Equal(t, uint64(0), stats["resources"])
	assert.Equal(t, uint64(0), stats["generators"])
	assert.Equal(t, uint64(10), stats["maxGenerators"])
	assert.Equal(t, uint64(20), stats["maxResources"])
}

// createSimpleGenerator creates a test tile generator for cache testing
func createSimpleGenerator(t *testing.T) *slides.XYZTileGenerator {
	img := createTestImage(512, 512)

	// Create a mock slide
	mockSlide := new(MockSlide)
	mockSlide.On("LevelCount").Return(3, nil)
	mockSlide.On("LargestLevelDimensions").Return([2]int{1024, 1024}, nil)
	mockSlide.On("LevelDimensions", 0).Return([2]int{1024, 1024}, nil)
	mockSlide.On("LevelDimensions", 1).Return([2]int{512, 512}, nil)
	mockSlide.On("LevelDimensions", 2).Return([2]int{256, 256}, nil)
	mockSlide.On("LevelDownsample", 0).Return(1.0, nil)
	mockSlide.On("LevelDownsample", 1).Return(2.0, nil)
	mockSlide.On("LevelDownsample", 2).Return(4.0, nil)
	mockSlide.On("LevelDownsamples").Return([]float64{1.0, 2.0, 4.0}, nil)
	mockSlide.On("BestLevelForDownsample", mock.Anything).Return(0, nil)
	mockSlide.On("ReadRegion", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(img, nil)
	mockSlide.On("PropertyValue", mock.Anything).Return("", nil)
	mockSlide.On("Properties").Return(map[string]string{}, nil)
	mockSlide.On("Close").Return()

	// Create a real tile generator with the mock slide
	generator, err := slides.NewXYZTileGenerator(mockSlide, "white", false)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	return generator
}

// TestPyramidCachePutAndGet tests adding and retrieving a tile generator
func TestPyramidCachePutAndGet(t *testing.T) {
	cache := NewPyramidCache(10, 20, true, "white", false)

	// Create a real generator
	generator := createSimpleGenerator(t)

	// Put it in the cache
	cache.Put("test-slide", generator)

	// Retrieve it
	gen, found := cache.Get("test-slide")

	assert.True(t, found)
	assert.Equal(t, generator, gen)

	// Check stats
	stats := cache.Stats()
	assert.Equal(t, uint64(1), stats["hits"])
	assert.Equal(t, uint64(0), stats["misses"])
}

// TestPyramidCacheGetMiss tests cache miss behavior
func TestPyramidCacheGetMiss(t *testing.T) {
	cache := NewPyramidCache(10, 20, true, "white", false)

	// Try to get a non-existent generator
	gen, found := cache.Get("non-existent")

	assert.False(t, found)
	assert.Nil(t, gen)

	// Check stats
	stats := cache.Stats()
	assert.Equal(t, uint64(0), stats["hits"])
	assert.Equal(t, uint64(1), stats["misses"])
}

// TestPyramidCacheRemove tests removing items from cache
func TestPyramidCacheRemove(t *testing.T) {
	cache := NewPyramidCache(10, 20, true, "white", false)
	generator := createSimpleGenerator(t)

	// Put it in the cache
	cache.Put("test-slide", generator)

	// Verify it's there
	_, found := cache.Get("test-slide")
	assert.True(t, found)

	// Remove it
	cache.Remove("test-slide")

	// Verify it's gone
	_, found = cache.Get("test-slide")
	assert.False(t, found)
}

// TestPyramidCacheClear tests clearing the cache
func TestPyramidCacheClear(t *testing.T) {
	cache := NewPyramidCache(10, 20, true, "white", false)
	generator := createSimpleGenerator(t)

	// Put some items in the cache
	cache.Put("slide1", generator)
	cache.Put("slide2", generator)

	// Clear the cache
	cache.Clear()

	// Check stats
	stats := cache.Stats()
	assert.Equal(t, uint64(0), stats["resources"])
	assert.Equal(t, uint64(0), stats["generators"])

	// Verify items are gone
	_, found := cache.Get("slide1")
	assert.False(t, found)
	_, found = cache.Get("slide2")
	assert.False(t, found)
}

// TestPyramidCacheEviction tests the LRU eviction policy
func TestPyramidCacheEviction(t *testing.T) {
	// Create a tiny cache with only 2 generator spots
	cache := NewPyramidCache(2, 5, true, "white", false)

	// Create generators
	generator := createSimpleGenerator(t)

	// Add first two with different keys
	cache.Put("slide1", generator)
	cache.Put("slide2", generator)

	// Access slide1 to make it more recently used than slide2
	cache.Get("slide1")
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	// Add third generator, should evict slide2 (least recently used)
	cache.Put("slide3", generator)

	// Check what's still in cache
	_, found1 := cache.Get("slide1")
	_, found2 := cache.Get("slide2")
	_, found3 := cache.Get("slide3")

	assert.True(t, found1, "slide1 should still be in cache")
	assert.False(t, found2, "slide2 should have been evicted")
	assert.True(t, found3, "slide3 should be in cache")
}

// TestGetTile tests the GetTile function with mocked slides
func TestGetTile(t *testing.T) {
	// Create a cache
	cache := NewPyramidCache(10, 20, true, "white", false)

	// Create a generator with a mock slide
	generator := createSimpleGenerator(t)

	// Put the generator in the cache
	cache.Put("test-slide", generator)

	// Get a tile - this will use the cached generator
	img, err := cache.GetTile("test-slide", "test-uri", 0, 0, 0)

	// Assertions
	assert.Nil(t, err)
	assert.NotNil(t, img)
	assert.Equal(t, 512, img.Bounds().Dx(), "Tile width should match the mock image")
	assert.Equal(t, 512, img.Bounds().Dy(), "Tile height should match the mock image")
}

// TestMaskInfoParsing tests MaskInfo serialization
func TestMaskInfoParsing(t *testing.T) {
	info := MaskInfo{
		MaskURI: "path/to/mask.tiff",
		SlideID: "slide123",
	}

	serialized := info.String()

	// Parse back
	parsed := ParseMaskInfo(serialized)

	assert.Equal(t, info.MaskURI, parsed.MaskURI)
	assert.Equal(t, info.SlideID, parsed.SlideID)
}

// Helper function to create a simple test image
func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 200, B: 150, A: 255})
		}
	}
	return img
}
