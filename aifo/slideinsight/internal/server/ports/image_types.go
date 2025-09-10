// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package ports

import (
	"context"
	"time"

	"aifo.dev/aifo/slideinsight/internal/utils"
)

// ImageType represents an image type in the database
type ImageType struct {
	ID                string
	TenantID          int
	TypeUID           string
	Name              string
	Description       string
	Category          string
	RequiresHistogram bool
	MetadataSchema    string
	IsActive          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewImageType represents a new image type to be created in the database
type NewImageType struct {
	ID                string
	TenantID          int
	TypeUID           string
	Name              string
	Description       string
	Category          string
	RequiresHistogram bool
	MetadataSchema    string
	IsActive          bool
}

// ImageTypeUpdates represents fields that can be updated for an existing image type
type ImageTypeUpdates struct {
	Name              *string
	Description       *string
	Category          *string
	RequiresHistogram *bool
	MetadataSchema    *string
	IsActive          *bool
}

// SlideHistogram represents a slide histogram in the database
type SlideHistogram struct {
	ID            string
	SlideID       int
	SlideUID      string
	ChannelIndex  int
	ChannelName   string
	BinCount      int
	MinValue      float64
	MaxValue      float64
	HistogramData []byte
	Metadata      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewSlideHistogram represents a new slide histogram to be created in the database
type NewSlideHistogram struct {
	ID            string
	SlideUID      string
	ChannelIndex  int
	ChannelName   string
	BinCount      int
	MinValue      float64
	MaxValue      float64
	HistogramData []byte
	Metadata      string
}

// StainingProtocol represents a staining protocol in the database
type StainingProtocol struct {
	ID             string
	SlideID        int
	SlideUID       string
	StainName      string
	StainType      string
	Concentration  string
	IncubationTime string
	AntibodyInfo   string
	ExcitationNm   *int
	EmissionNm     *int
	Metadata       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewStainingProtocol represents a new staining protocol to be created in the database
type NewStainingProtocol struct {
	ID             string
	SlideUID       string
	StainName      string
	StainType      string
	Concentration  string
	IncubationTime string
	AntibodyInfo   string
	ExcitationNm   *int
	EmissionNm     *int
	Metadata       string
}

// StainingProtocolUpdates represents fields that can be updated for an existing staining protocol
type StainingProtocolUpdates struct {
	StainName      *string
	StainType      *string
	Concentration  *string
	IncubationTime *string
	AntibodyInfo   *string
	ExcitationNm   *int
	EmissionNm     *int
	Metadata       *string
}

// ImageTypesRepository defines the interface for image type-related database operations
type ImageTypesRepository interface {
	// LoadAllImageTypes retrieves image types from the database with search/filter and pagination support.
	LoadAllImageTypes(ctx context.Context, search utils.SearchParams, pagination utils.PaginationParams) ([]ImageType, error)

	// GetImageTypesCount returns the total count of image types matching search criteria.
	GetImageTypesCount(ctx context.Context, search utils.SearchParams) (int, error)

	// GetImageTypeByID retrieves a specific image type by its ID.
	GetImageTypeByID(ctx context.Context, id string) (ImageType, error)

	// CreateImageType adds a new image type to the database.
	CreateImageType(ctx context.Context, imageType NewImageType) error

	// UpdateImageType updates an existing image type.
	UpdateImageType(ctx context.Context, id string, updates ImageTypeUpdates) error

	// DeleteImageType removes an image type from the database.
	DeleteImageType(ctx context.Context, id string) error

	// ImageTypeExists checks if an image type with the given ID already exists.
	ImageTypeExists(ctx context.Context, id string) (bool, error)
}

// SlideHistogramsRepository defines the interface for slide histogram-related database operations
type SlideHistogramsRepository interface {
	// GetHistogramsBySlideUID retrieves all histograms for a given slide.
	GetHistogramsBySlideUID(ctx context.Context, slideUID string) ([]SlideHistogram, error)

	// GetHistogramByID retrieves a specific histogram by its ID.
	GetHistogramByID(ctx context.Context, id string) (SlideHistogram, error)

	// CreateHistogram adds a new histogram to the database.
	CreateHistogram(ctx context.Context, histogram NewSlideHistogram) error

	// UpdateHistogram updates an existing histogram.
	UpdateHistogram(ctx context.Context, id string, histogram NewSlideHistogram) error

	// DeleteHistogram removes a histogram from the database.
	DeleteHistogram(ctx context.Context, id string) error

	// DeleteHistogramsBySlideUID removes all histograms for a given slide.
	DeleteHistogramsBySlideUID(ctx context.Context, slideUID string) error
}

// StainingProtocolsRepository defines the interface for staining protocol-related database operations
type StainingProtocolsRepository interface {
	// GetProtocolsBySlideUID retrieves all staining protocols for a given slide.
	GetProtocolsBySlideUID(ctx context.Context, slideUID string) ([]StainingProtocol, error)

	// GetProtocolByID retrieves a specific staining protocol by its ID.
	GetProtocolByID(ctx context.Context, id string) (StainingProtocol, error)

	// CreateProtocol adds a new staining protocol to the database.
	CreateProtocol(ctx context.Context, protocol NewStainingProtocol) error

	// UpdateProtocol updates an existing staining protocol.
	UpdateProtocol(ctx context.Context, id string, updates StainingProtocolUpdates) error

	// DeleteProtocol removes a staining protocol from the database.
	DeleteProtocol(ctx context.Context, id string) error
}
