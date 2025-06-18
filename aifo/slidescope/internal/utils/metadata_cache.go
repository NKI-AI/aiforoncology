// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package utils

import (
	"sync"
	"time"
)

// MetadataCache is a generic LRU cache for storing resource metadata
type MetadataCache struct {
	mu      sync.RWMutex
	values  map[string]string // id -> value
	maxSize int
	access  map[string]time.Time // For LRU tracking
}

// NewMetadataCache creates a new metadata cache with the specified size
func NewMetadataCache(size int) *MetadataCache {
	return &MetadataCache{
		values:  make(map[string]string),
		maxSize: size,
		access:  make(map[string]time.Time),
	}
}

// Get retrieves a value from the cache
func (c *MetadataCache) Get(id string) (string, bool) {
	c.mu.RLock()
	value, found := c.values[id]
	c.mu.RUnlock()

	if found {
		// Update access time
		c.mu.Lock()
		c.access[id] = time.Now()
		c.mu.Unlock()
	}

	return value, found
}

// Put adds or updates a value in the cache
func (c *MetadataCache) Put(id string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict an entry
	if len(c.values) >= c.maxSize && c.values[id] == "" {
		c.evictOldest()
	}

	// Add/update the entry
	c.values[id] = value
	c.access[id] = time.Now()
}

// Remove removes an ID from the cache
func (c *MetadataCache) Remove(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.values, id)
	delete(c.access, id)
}

// evictOldest removes the least recently accessed entry
func (c *MetadataCache) evictOldest() {
	if len(c.values) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time

	// Find the oldest entry
	for key, accessTime := range c.access {
		if oldestTime.IsZero() || accessTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = accessTime
		}
	}

	// Remove it
	delete(c.values, oldestKey)
	delete(c.access, oldestKey)
}

// Clear empties the cache
func (c *MetadataCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values = make(map[string]string)
	c.access = make(map[string]time.Time)
}

// Size returns the number of entries in the cache
func (c *MetadataCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.values)
}
