// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package services

import (
	"fmt"
	"image"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"aifo.dev/aifo/slidescope/internal/slides"
)

// PyramidCache is a cache for whole slide image and mask resources and their tile generators with LRU eviction
type PyramidCache struct {
	mu              sync.RWMutex
	resources       map[string]resourceEntry            // Maps resourceURI -> resource
	tileGenerators  map[string]*slides.XYZTileGenerator // Maps "resourceURI|bgColor|precise" -> generator
	generatorAccess map[string]time.Time                // For LRU tracking of generators

	maxResources     int
	maxGenerators    int
	cleanupThreshold float64 // trigger cleanup when len(resources) >= maxResources * cleanupThreshold
	cleanupTarget    float64 // shrink down to maxResources * cleanupTarget

	isSlide bool   // Whether this cache holds slides (true) or masks (false)
	bgColor string // Background color for tile generation
	precise bool   // Precise tile generation flag

	// metrics
	hits    atomic.Uint64
	misses  atomic.Uint64
	evicted atomic.Uint64
	created atomic.Uint64
}

// resourceEntry holds a slide/mask handle and its last access timestamp.
type resourceEntry struct {
	resource   slides.Slide
	lastAccess time.Time
}

// NewPyramidCache creates a new cache for slide/mask resources and their tile generators
func NewPyramidCache(maxGenerators int, maxResources int, isSlide bool, bgColor string, precise bool) *PyramidCache {
	return &PyramidCache{
		resources:        make(map[string]resourceEntry),
		tileGenerators:   make(map[string]*slides.XYZTileGenerator),
		generatorAccess:  make(map[string]time.Time),
		maxResources:     maxResources,
		maxGenerators:    maxGenerators,
		cleanupThreshold: 0.90,
		cleanupTarget:    0.70,
		isSlide:          isSlide,
		bgColor:          bgColor,
		precise:          precise,
	}
}

// Get checks if a tile generator exists in the cache
func (c *PyramidCache) Get(wsiId string) (*slides.XYZTileGenerator, bool) {
	// Generate the key for the tile generator
	key := fmt.Sprintf("%s|%s|%v", wsiId, c.bgColor, c.precise)

	c.mu.RLock()
	generator, found := c.tileGenerators[key]
	c.mu.RUnlock()

	if found {
		c.hits.Add(1)
		// Update access time
		c.mu.Lock()
		c.generatorAccess[wsiId] = time.Now()
		c.mu.Unlock()
	} else {
		c.misses.Add(1)
	}

	return generator, found
}

// GetResource retrieves a cached resource (slide or mask) or opens a new one.
// Implements ResourceCache.GetResource.
func (c *PyramidCache) GetResource(resourceURI string, isSlide bool) (slides.Slide, error) {
	// Handle calling with explicit isSlide flag that might differ from cache setting
	if isSlide != c.isSlide {
		return nil, fmt.Errorf("resource type mismatch: cache is for %s but requested %s",
			resourceType(c.isSlide), resourceType(isSlide))
	}

	return c.getResource(resourceURI)
}

// GetResourceTileGenerator gets or creates a tile generator for a resource.
// Implements ResourceCache.GetResourceTileGenerator.
func (c *PyramidCache) GetResourceTileGenerator(resourceURI string, isSlide bool,
	backgroundColor string, precise bool,
) (*slides.XYZTileGenerator, error) {
	// Handle calling with explicit flags that might differ from cache settings
	if isSlide != c.isSlide || backgroundColor != c.bgColor || precise != c.precise {
		return nil, fmt.Errorf("requested generator with different settings than cache supports")
	}

	return c.getTileGenerator(resourceURI)
}

// Internal method: getResource retrieves a cached resource (slide or mask) or opens a new one.
func (c *PyramidCache) getResource(resourceURI string) (slides.Slide, error) {
	// Fast path: read lock
	c.mu.RLock()
	entry, exists := c.resources[resourceURI]
	c.mu.RUnlock()

	if exists {
		c.updateLastAccess(resourceURI)
		return entry.resource, nil
	}

	// Slow path: upgrade to write lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check under write lock
	if entry, exists = c.resources[resourceURI]; exists {
		entry.lastAccess = time.Now()
		c.resources[resourceURI] = entry
		return entry.resource, nil
	}

	// Evict if we're above the threshold
	if len(c.resources) >= int(float64(c.maxResources)*c.cleanupThreshold) {
		c.cleanupResources()
	}

	// Still full?
	if len(c.resources) >= c.maxResources {
		return nil, fmt.Errorf("resource cache is full (max size: %d)", c.maxResources)
	}

	var resource slides.Slide
	var err error

	// Open the resource using appropriate method
	if c.isSlide {
		resource, err = slides.OpenSlide(resourceURI, "")
	} else {
		resource, err = slides.NewTiffAdapter(resourceURI)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open resource %s: %w", resourceURI, err)
	}

	// Insert into cache
	c.resources[resourceURI] = resourceEntry{
		resource:   resource,
		lastAccess: time.Now(),
	}
	c.created.Add(1)

	return resource, nil
}

// resourceType returns a string description of the resource type
func resourceType(isSlide bool) string {
	if isSlide {
		return "slides"
	}
	return "masks"
}

// CloseResource closes and removes a resource from the cache.
// Implements ResourceCache.CloseResource.
func (c *PyramidCache) CloseResource(resourceURI string) {
	c.Remove(resourceURI)
}

// getTileGenerator retrieves or creates a tile generator for a resource
func (c *PyramidCache) getTileGenerator(resourceURI string) (*slides.XYZTileGenerator, error) {
	// Use '|' as a safe delimiter in the key
	key := fmt.Sprintf("%s|%s|%v", resourceURI, c.bgColor, c.precise)

	// Fast path
	c.mu.RLock()
	gen, exists := c.tileGenerators[key]
	c.mu.RUnlock()

	if exists {
		c.updateLastAccess(resourceURI)
		// Also update generator access time
		c.mu.Lock()
		c.generatorAccess[resourceURI] = time.Now()
		c.mu.Unlock()
		return gen, nil
	}

	// Ensure resource is open
	resource, err := c.getResource(resourceURI)
	if err != nil {
		return nil, err
	}

	// Slow path: write lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check
	if gen, exists = c.tileGenerators[key]; exists {
		c.generatorAccess[resourceURI] = time.Now()
		return gen, nil
	}

	// Check if we need to evict a generator
	if len(c.tileGenerators) >= c.maxGenerators {
		c.evictOldestGenerator()
	}

	// Create new generator
	gen, err = slides.NewXYZTileGenerator(resource, c.bgColor, c.precise)
	if err != nil {
		return nil, fmt.Errorf("failed to create tile generator: %w", err)
	}

	c.tileGenerators[key] = gen
	c.generatorAccess[resourceURI] = time.Now()
	c.created.Add(1)

	return gen, nil
}

// Put adds a tile generator to the cache
func (c *PyramidCache) Put(wsiId string, generator *slides.XYZTileGenerator) {
	key := fmt.Sprintf("%s|%s|%v", wsiId, c.bgColor, c.precise)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict an entry
	if len(c.tileGenerators) >= c.maxGenerators && c.tileGenerators[key] == nil {
		c.evictOldestGenerator()
	}

	// Add/update the entry
	c.tileGenerators[key] = generator
	c.generatorAccess[wsiId] = time.Now()
}

// GetTile retrieves a tile using a cached tile generator or creates a new one if needed
func (c *PyramidCache) GetTile(wsiId string, resourceURI string, z, x, y int) (image.Image, error) {
	// Try to get from cache first
	generator, found := c.Get(wsiId)

	if !found {
		// Need to create a new generator
		var err error
		generator, err = c.getTileGenerator(resourceURI)
		if err != nil {
			return nil, fmt.Errorf("error creating tile generator for %s: %w", wsiId, err)
		}

		// Add to cache
		c.Put(wsiId, generator)
	}

	// Generate the tile
	tile, err := generator.GetTile(z, x, y)
	if err != nil {
		return nil, fmt.Errorf("error generating tile at z=%d, x=%d, y=%d for %s: %w", z, x, y, wsiId, err)
	}

	return tile, nil
}

// updateLastAccess refreshes the lastAccess timestamp for a resource.
func (c *PyramidCache) updateLastAccess(resourceURI string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.resources[resourceURI]; exists {
		entry.lastAccess = time.Now()
		c.resources[resourceURI] = entry
	}
}

// cleanupResources evicts the least-recently-used resources down to cleanupTarget * maxResources.
func (c *PyramidCache) cleanupResources() {
	// Gather all access times
	type access struct {
		uri  string
		when time.Time
	}
	list := make([]access, 0, len(c.resources))
	for uri, entry := range c.resources {
		list = append(list, access{uri: uri, when: entry.lastAccess})
	}

	// Sort by oldest first
	sort.Slice(list, func(i, j int) bool {
		return list[i].when.Before(list[j].when)
	})

	target := int(float64(c.maxResources) * c.cleanupTarget)
	for _, a := range list {
		if len(c.resources) <= target {
			break
		}
		// Evict this resource
		if entry, exists := c.resources[a.uri]; exists {
			entry.resource.Close()
			delete(c.resources, a.uri)
			c.evicted.Add(1)
		}
		// Evict related generators
		for k := range c.tileGenerators {
			if strings.HasPrefix(k, a.uri+"|") {
				delete(c.tileGenerators, k)
			}
		}
	}
}

// Remove removes a resource from the cache
func (c *PyramidCache) Remove(wsiId string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Find and delete any generators for this ID
	generatorKey := fmt.Sprintf("%s|%s|%v", wsiId, c.bgColor, c.precise)
	delete(c.tileGenerators, generatorKey)
	delete(c.generatorAccess, wsiId)

	// If this is also a resource URI, close and remove the resource
	if entry, exists := c.resources[wsiId]; exists {
		entry.resource.Close()
		delete(c.resources, wsiId)
	}
}

// evictOldestGenerator removes the least recently accessed generator
func (c *PyramidCache) evictOldestGenerator() {
	if len(c.generatorAccess) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time

	// Find the oldest entry
	for key, accessTime := range c.generatorAccess {
		if oldestTime.IsZero() || accessTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = accessTime
		}
	}

	// Remove the generator
	generatorKey := fmt.Sprintf("%s|%s|%v", oldestKey, c.bgColor, c.precise)
	delete(c.tileGenerators, generatorKey)
	delete(c.generatorAccess, oldestKey)
}

// Stats returns cache statistics
func (c *PyramidCache) Stats() map[string]uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]uint64{
		"resources":     uint64(len(c.resources)),
		"generators":    uint64(len(c.tileGenerators)),
		"hits":          c.hits.Load(),
		"misses":        c.misses.Load(),
		"evicted":       c.evicted.Load(),
		"created":       c.created.Load(),
		"maxResources":  uint64(c.maxResources),
		"maxGenerators": uint64(c.maxGenerators),
	}
}

// Clear empties the cache
func (c *PyramidCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close all resources
	for _, entry := range c.resources {
		entry.resource.Close()
		c.evicted.Add(1)
	}

	c.tileGenerators = make(map[string]*slides.XYZTileGenerator)
	c.generatorAccess = make(map[string]time.Time)
	c.resources = make(map[string]resourceEntry)
}

// Close shuts down all resources and clears the cache.
func (c *PyramidCache) Close() {
	c.Clear()
}
