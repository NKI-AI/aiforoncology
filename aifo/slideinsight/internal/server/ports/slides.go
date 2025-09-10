// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package ports

import (
	"context"
	"encoding/json"
	"time"

	"aifo.dev/aifo/slideinsight/internal/utils"
)

// Slide represents a slide in the database
type Slide struct {
	ID          int
	CaseID      int
	CaseUID     string
	SlideID     string
	SlideUID    string
	SlideName   string
	SlideURI    string
	SlideHash   string
	SlideWidth  int
	SlideHeight int
	SlideMpp    float64
	ImageTypeID string          // Foreign key to image_types table
	Metadata    json.RawMessage // JSON metadata including channel info for spectral slides
	CreatorID   int
	CreatorUID  string
	DeletedAt   *time.Time
	DeletedBy   *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewSlide represents a new slide to be created to the database
type NewSlide struct {
	CaseID      int
	SlideID     string
	SlideName   string
	SlideURI    string
	SlideHash   string
	SlideWidth  int
	SlideHeight int
	SlideMpp    float64         // TODO: Make SlideMpp
	ImageTypeID string          // Foreign key to image_types table
	Metadata    json.RawMessage // JSON metadata including channel info for spectral slides
	CreatorID   int
	TenantID    int
}

// SlidesRepository defines the interface for slide-related database operations
type SlidesRepository interface {
	// LoadAllSlides retrieves slides from the database with search/filter and pagination support.
	LoadAllSlides(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]Slide, error)

	// GetSlidesCount returns the total count of slides matching search criteria.
	GetSlidesCount(ctx context.Context, search utils.SearchParams) (int, error)

	// GetSlidesByCaseUID retrieves all slides for a given case.
	GetSlidesByCaseUID(_ context.Context, caseUID string) ([]Slide, error)

	// CreateSlide adds a new slide to the database.
	CreateSlide(ctx context.Context, newSlide NewSlide) error

	// SlideExists checks if a slide with the given ID already exists.
	SlideExists(ctx context.Context, slideUID string) (bool, error)

	// GetSlideByUID retrieves a specific slide by its slide_uid.
	GetSlideByUID(ctx context.Context, slideUID string) (Slide, error)

	// SoftDeleteSlide marks a slide as deleted without removing it from the database.
	SoftDeleteSlide(ctx context.Context, slideUID string, deletedBy int) error

	// GetDeletedSlides retrieves all soft-deleted slides.
	GetDeletedSlides(ctx context.Context) ([]Slide, error)

	// RestoreSlide restores a soft-deleted slide.
	RestoreSlide(ctx context.Context, slideUID string) error
}
