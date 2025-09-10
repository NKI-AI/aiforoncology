// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package sqlite

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Slide-related methods that delegate to the slides adapter

// LoadAllSlides retrieves slides from the database with search/filter and pagination support
func (db *DB) LoadAllSlides(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.Slide, error) {
	return db.slides.LoadAllSlides(ctx, search, pagination)
}

// GetSlidesCount returns the total count of slides matching search criteria
func (db *DB) GetSlidesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	return db.slides.GetSlidesCount(ctx, search)
}

// GetSlidesByCaseUID retrieves slides by case ID
func (db *DB) GetSlidesByCaseUID(ctx context.Context, caseUID string) ([]ports.Slide, error) {
	return db.slides.GetSlidesByCaseUID(ctx, caseUID)
}

// CreateSlide adds a new slide to the database
func (db *DB) CreateSlide(ctx context.Context, newSlide ports.NewSlide) error {
	return db.slides.CreateSlide(ctx, newSlide)
}

// SlideExists checks if a slide with the given ID already exists
func (db *DB) SlideExists(ctx context.Context, slideUID string) (bool, error) {
	return db.slides.SlideExists(ctx, slideUID)
}

// GetSlideByUID retrieves a specific slide by its slide_uid
func (db *DB) GetSlideByUID(ctx context.Context, slideUID string) (ports.Slide, error) {
	return db.slides.GetSlideByUID(ctx, slideUID)
}

// SoftDeleteSlide marks a slide as deleted without removing it from the database
func (db *DB) SoftDeleteSlide(ctx context.Context, slideUID string, deletedBy int) error {
	return db.slides.SoftDeleteSlide(ctx, slideUID, deletedBy)
}

// GetDeletedSlides retrieves all soft-deleted slides
func (db *DB) GetDeletedSlides(ctx context.Context) ([]ports.Slide, error) {
	return db.slides.GetDeletedSlides(ctx)
}

// RestoreSlide restores a soft-deleted slide
func (db *DB) RestoreSlide(ctx context.Context, slideUID string) error {
	return db.slides.RestoreSlide(ctx, slideUID)
}
