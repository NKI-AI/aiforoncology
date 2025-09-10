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

// Image types related methods that delegate to the image types adapters

// ImageTypes methods

// LoadAllImageTypes retrieves image types from the database with search/filter and pagination support
func (db *DB) LoadAllImageTypes(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ports.ImageType, error) {
	return db.imageTypes.LoadAllImageTypes(ctx, search, pagination)
}

// GetImageTypesCount returns the total count of image types matching search criteria
func (db *DB) GetImageTypesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	return db.imageTypes.GetImageTypesCount(ctx, search)
}

// GetImageTypeByID retrieves a specific image type by its ID
func (db *DB) GetImageTypeByID(ctx context.Context, id string) (ports.ImageType, error) {
	return db.imageTypes.GetImageTypeByID(ctx, id)
}

// CreateImageType adds a new image type to the database
func (db *DB) CreateImageType(ctx context.Context, imageType ports.NewImageType) error {
	return db.imageTypes.CreateImageType(ctx, imageType)
}

// UpdateImageType updates an existing image type
func (db *DB) UpdateImageType(ctx context.Context, id string, updates ports.ImageTypeUpdates) error {
	return db.imageTypes.UpdateImageType(ctx, id, updates)
}

// DeleteImageType removes an image type from the database
func (db *DB) DeleteImageType(ctx context.Context, id string) error {
	return db.imageTypes.DeleteImageType(ctx, id)
}

// ImageTypeExists checks if an image type with the given ID already exists
func (db *DB) ImageTypeExists(ctx context.Context, id string) (bool, error) {
	return db.imageTypes.ImageTypeExists(ctx, id)
}

// Slide Histograms methods

// GetHistogramsBySlideUID retrieves all histograms for a given slide
func (db *DB) GetHistogramsBySlideUID(ctx context.Context, slideUID string) ([]ports.SlideHistogram, error) {
	return db.slideHistograms.GetHistogramsBySlideUID(ctx, slideUID)
}

// GetHistogramByID retrieves a specific histogram by its ID
func (db *DB) GetHistogramByID(ctx context.Context, id string) (ports.SlideHistogram, error) {
	return db.slideHistograms.GetHistogramByID(ctx, id)
}

// CreateHistogram adds a new histogram to the database
func (db *DB) CreateHistogram(ctx context.Context, histogram ports.NewSlideHistogram) error {
	return db.slideHistograms.CreateHistogram(ctx, histogram)
}

// UpdateHistogram updates an existing histogram
func (db *DB) UpdateHistogram(ctx context.Context, id string, histogram ports.NewSlideHistogram) error {
	return db.slideHistograms.UpdateHistogram(ctx, id, histogram)
}

// DeleteHistogram removes a histogram from the database
func (db *DB) DeleteHistogram(ctx context.Context, id string) error {
	return db.slideHistograms.DeleteHistogram(ctx, id)
}

// DeleteHistogramsBySlideUID removes all histograms for a given slide
func (db *DB) DeleteHistogramsBySlideUID(ctx context.Context, slideUID string) error {
	return db.slideHistograms.DeleteHistogramsBySlideUID(ctx, slideUID)
}

// Staining Protocols methods

// GetProtocolsBySlideUID retrieves all staining protocols for a given slide
func (db *DB) GetProtocolsBySlideUID(ctx context.Context, slideUID string) ([]ports.StainingProtocol, error) {
	return db.stainingProtocols.GetProtocolsBySlideUID(ctx, slideUID)
}

// GetProtocolByID retrieves a specific staining protocol by its ID
func (db *DB) GetProtocolByID(ctx context.Context, id string) (ports.StainingProtocol, error) {
	return db.stainingProtocols.GetProtocolByID(ctx, id)
}

// CreateProtocol adds a new staining protocol to the database
func (db *DB) CreateProtocol(ctx context.Context, protocol ports.NewStainingProtocol) error {
	return db.stainingProtocols.CreateProtocol(ctx, protocol)
}

// UpdateProtocol updates an existing staining protocol
func (db *DB) UpdateProtocol(ctx context.Context, id string, updates ports.StainingProtocolUpdates) error {
	return db.stainingProtocols.UpdateProtocol(ctx, id, updates)
}

// DeleteProtocol removes a staining protocol from the database
func (db *DB) DeleteProtocol(ctx context.Context, id string) error {
	return db.stainingProtocols.DeleteProtocol(ctx, id)
}
