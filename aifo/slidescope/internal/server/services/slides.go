// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideScope.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideScope project root.
package services

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"image/png"
	"log/slog"
	"strings"

	"aifo.dev/aifo/openslide_go/openslide"
	"aifo.dev/aifo/slidescope/internal/datasources/database"
	"aifo.dev/aifo/slidescope/internal/server/domain"
	"aifo.dev/aifo/slidescope/internal/server/errors"
	"aifo.dev/aifo/slidescope/internal/slides"
	"aifo.dev/aifo/slidescope/internal/utils"
)

// ResourceCache is an interface for resource caching functionality
type ResourceCache interface {
	// Resource access methods
	GetResource(resourceURI string, isSlide bool) (slides.Slide, error)
	GetResourceTileGenerator(resourceURI string, isSlide bool, backgroundColor string, precise bool) (*slides.XYZTileGenerator, error)

	// Future tile generation and caching methods
	// GetResourceTile gets a raw tile image
	// GetResourceTile(resourceURI string, isSlide bool, z, x, y int) (image.Image, error)

	// GetEncodedTile gets an encoded tile in the specified format
	// GetEncodedTile(resourceURI string, isSlide bool, z, x, y int, format string) ([]byte, string, error)

	// Resource management
	CloseResource(resourceURI string)
	Close()
}

// SlidesService is an interface that defines the methods for the slides service.
// Interface is needed for mocking in tests.
type SlidesService interface {
	GetSlides(ctx context.Context) ([]domain.Slide, error)
	GetSlideByID(ctx context.Context, slideID string) (domain.Slide, error)
	GetSlideMetadata(ctx context.Context, slideID string) (domain.SlideMetadata, error)
	GetSlideTile(ctx context.Context, slideID string, z, x, y int, format string) (domain.SlideTile, error)
	SaveSlide(ctx context.Context, newSlide domain.Slide) (domain.Slide, error)
	Close()
}

type slidesService struct {
	db            database.Database
	pyramidCache  *PyramidCache
	metadataCache *utils.MetadataCache
}

// NewSlidesService creates a new SlidesService
func NewSlidesService(db database.Database) SlidesService {
	// Create a PyramidCache for slide tile generators
	// Using 'true' for isSlide since this is for slides
	// "#FFFFFF" is typically used for slide background
	// false for precise tile generation (most slides don't need precise)
	pyramidCache := NewPyramidCache(100, 100, true, "#FFFFFF", false)

	// Create a metadata cache for slides
	metadataCache := utils.NewMetadataCache(1000) // Cache up to 1000 slide metadata entries

	return &slidesService{
		db:            db,
		pyramidCache:  pyramidCache,
		metadataCache: metadataCache,
	}
}

// GetSlides retrieves all slides from the database
func (s *slidesService) GetSlides(ctx context.Context) ([]domain.Slide, error) {
	dbRecords, err := s.db.LoadAllSlides(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load slides: %w", err)
	}

	slides := make([]domain.Slide, 0, len(dbRecords))
	for _, record := range dbRecords {
		slides = append(slides, domain.Slide{
			SlideID:     record.SlideID,
			SlideName:   record.SlideName,
			SlideURI:    record.SlideURI,
			SlideWidth:  record.SlideWidth,
			SlideHeight: record.SlideHeight,
			SlideMpp:    record.SlideMpp,
		})
	}

	return slides, nil
}

// GetSlideByID retrieves a specific slide by its slide_id
func (s *slidesService) GetSlideByID(ctx context.Context, slideID string) (domain.Slide, error) {
	dbRecord, err := s.db.GetSlideByID(ctx, slideID)
	if err != nil {
		return domain.Slide{}, errors.WithDetails(errors.ErrSlideNotFound, "slide with ID '%s'", slideID)
	}

	return domain.Slide{
		SlideID:     dbRecord.SlideID,
		SlideName:   dbRecord.SlideName,
		SlideURI:    dbRecord.SlideURI,
		SlideWidth:  dbRecord.SlideWidth,
		SlideHeight: dbRecord.SlideHeight,
		SlideMpp:    dbRecord.SlideMpp,
	}, nil
}

// GetSlideMetadata retrieves detailed metadata for a specific slide
func (s *slidesService) GetSlideMetadata(ctx context.Context, slideID string) (domain.SlideMetadata, error) {
	// First get the basic slide info
	slide, err := s.GetSlideByID(ctx, slideID)
	if err != nil {
		return domain.SlideMetadata{}, err
	}

	// Get cached slide or open it
	slideObj, err := openslide.Open(slide.SlideURI)
	if err != nil {
		return domain.SlideMetadata{}, err
	}

	// Get vendor information
	vendor := "Unknown"
	properties, err := slideObj.Properties()
	if err != nil {
		return domain.SlideMetadata{}, fmt.Errorf("failed to get slide properties: %w", err)
	}

	if v, ok := properties["openslide.vendor"]; ok {
		vendor = v
	}

	// Get magnification
	magnification := "Unknown"
	if m, ok := properties["openslide.objective-power"]; ok {
		magnification = m
	}

	// Calculate physical dimensions
	widthMm := 0.0
	heightMm := 0.0
	if slide.SlideMpp > 0 {
		widthMm = float64(slide.SlideWidth) * slide.SlideMpp / 1000.0
		heightMm = float64(slide.SlideHeight) * slide.SlideMpp / 1000.0
	}

	// Determine zoom levels
	// This is a simplification - you may need to adjust this based on actual slide dimensions
	maxLevel := 0
	if slide.SlideWidth > 0 && slide.SlideHeight > 0 {
		maxDimension := slide.SlideWidth
		if slide.SlideHeight > maxDimension {
			maxDimension = slide.SlideHeight
		}

		// TODO: Make tile size configurable
		// Calculate how many times we can divide by 2 until we reach a reasonable minimum size
		for size := maxDimension; size > 512; size /= 2 {
			maxLevel++
		}
	}

	return domain.SlideMetadata{
		SlideID:       slideID,
		SlideName:     slide.SlideName,
		MinLevel:      0,
		MaxLevel:      maxLevel,
		TileSize:      512, // Standard tile size, TODO: Make configurable
		Format:        "jpeg",
		SlideMpp:      slide.SlideMpp,
		SlideWidth:    slide.SlideWidth,
		SlideHeight:   slide.SlideHeight,
		Vendor:        vendor,
		Magnification: magnification,
		PhysicalDimensions: domain.PhysicalDimensions{
			WidthMm:  widthMm,
			HeightMm: heightMm,
			WidthPx:  slide.SlideWidth,
			HeightPx: slide.SlideHeight,
		},
	}, nil
}

// GetSlideTile retrieves a specific tile from a slide
func (s *slidesService) GetSlideTile(ctx context.Context, slideID string, z, x, y int, format string) (domain.SlideTile, error) {
	// Try to get slide URI from the metadata cache first
	uri, found := s.metadataCache.Get(slideID)

	if !found {
		// Cache miss - fetch from database
		slog.Debug("Slide metadata cache miss", "slideID", slideID)
		slide, err := s.GetSlideByID(ctx, slideID)
		if err != nil {
			return domain.SlideTile{}, err
		}

		// Store in cache for future use
		uri = slide.SlideURI
		s.metadataCache.Put(slideID, uri)
		slog.Debug("Added slide to metadata cache", "slideID", slideID)
	} else {
		slog.Debug("Slide metadata cache hit", "slideID", slideID)
	}

	// Validate format
	validFormat := false
	contentType := ""
	switch format {
	case "jpeg", "jpg":
		validFormat = true
		contentType = "image/jpeg"
	case "png":
		validFormat = true
		contentType = "image/png"
	}

	if !validFormat {
		return domain.SlideTile{}, errors.WithDetails(errors.ErrInvalidFormat, "format: %s", format)
	}

	// Get the tile image using the pyramid cache
	img, err := s.pyramidCache.GetTile(slideID, uri, z, x, y)
	if err != nil {
		if strings.Contains(err.Error(), "out of bounds") ||
			strings.Contains(err.Error(), "outside slide bounds") {
			return domain.SlideTile{}, errors.ErrTileOutOfBounds
		}
		return domain.SlideTile{}, errors.WithDetails(errors.ErrInternal, "failed to generate tile: %v", err)
	}

	// Encode the image to the requested format
	buf := &bytes.Buffer{}

	if format == "png" {
		err = png.Encode(buf, img)
	} else { // jpeg
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 75})
	}

	if err != nil {
		return domain.SlideTile{}, errors.WithDetails(errors.ErrInternal, "failed to encode tile: %v", err)
	}

	return domain.SlideTile{
		Image:       buf.Bytes(),
		Format:      format,
		ContentType: contentType,
	}, nil
}

// SaveSlide saves a slide to the database
func (s *slidesService) SaveSlide(ctx context.Context, slide domain.Slide) (domain.Slide, error) {
	// Check if slide already exists (for cache invalidation)
	existingSlide := false
	if slide.SlideID != "" {
		_, err := s.db.GetSlideByID(ctx, slide.SlideID)
		if err == nil {
			existingSlide = true
		}
	}

	// Generate a random ID if none is provided
	if slide.SlideID == "" {
		// Generate a unique slide ID using the utility function
		randomID, err := utils.GenerateFixedShortID()
		if err != nil {
			return domain.Slide{}, errors.WithDetails(errors.ErrInternal, "failed to generate slide ID: %v", err)
		}
		slide.SlideID = randomID
	}

	// Generate a default name if none is provided
	if slide.SlideName == "" {
		// Extract a name from the URI if possible
		parts := strings.Split(slide.SlideURI, "/")
		if len(parts) > 0 {
			filename := parts[len(parts)-1]
			// Remove file extension if present
			if dotIndex := strings.LastIndex(filename, "."); dotIndex != -1 {
				filename = filename[:dotIndex]
			}
			slide.SlideName = filename
		} else {
			// Fall back to ID-based name
			slide.SlideName = fmt.Sprintf("Slide %s", slide.SlideID)
		}
	}

	// Check if slide with the same ID already exists
	if !existingSlide {
		exists, err := s.db.SlideExists(ctx, slide.SlideID)
		if err != nil {
			return domain.Slide{}, errors.WithDetails(errors.ErrInternal, "failed to check if slide exists: %v", err)
		}
		if exists {
			return domain.Slide{}, errors.WithDetails(errors.ErrAlreadyExists, "slide with ID '%s'", slide.SlideID)
		}
	}

	// We need to get the metadata from the slide file
	err := populateSlideMetadata(&slide)
	if err != nil {
		return domain.Slide{}, errors.WithDetails(errors.ErrInvalidInput, "failed to populate slide metadata: %v", err)
	}

	// Log the slide metadata
	slog.Info("Saving slide", "slide", slide)

	dbSlide := database.NewSlide{
		SlideID:     slide.SlideID,
		SlideName:   slide.SlideName,
		SlideURI:    slide.SlideURI,
		SlideWidth:  slide.SlideWidth,
		SlideHeight: slide.SlideHeight,
		SlideMpp:    slide.SlideMpp,
	}

	err = s.db.CreateSlide(ctx, dbSlide)
	if err != nil {
		return domain.Slide{}, errors.WithDetails(errors.ErrInternal, "failed to save slide: %v", err)
	}

	// If we updated an existing slide, invalidate its cache entries
	if existingSlide {
		s.metadataCache.Remove(slide.SlideID)
		s.pyramidCache.Remove(slide.SlideID)
	}

	return slide, nil
}

// Close releases all resources held by the service
func (s *slidesService) Close() {
	// Clear the caches
	s.metadataCache.Clear()
	s.pyramidCache.Clear()
}

func populateSlideMetadata(slide *domain.Slide) error {
	if slide.SlideURI == "" {
		return fmt.Errorf("slide URI is empty")
	}

	// Use our slide factory instead of directly using openslide
	slideObj, err := slides.OpenSlide(slide.SlideURI, "")
	if err != nil {
		return fmt.Errorf("failed to open slide: %w", err)
	}
	defer slideObj.Close()

	// Get dimensions
	dims, err := slideObj.LargestLevelDimensions()
	if err != nil {
		return fmt.Errorf("failed to get slide dimensions: %w", err)
	}
	slide.SlideWidth = dims[0]
	slide.SlideHeight = dims[1]

	// Get MPP (microns per pixel)
	mppX, err := slideObj.PropertyValue("openslide.mpp-x")
	if err != nil {
		return fmt.Errorf("failed to get mpp-x property: %w", err)
	}

	mppY, err := slideObj.PropertyValue("openslide.mpp-y")
	if err != nil {
		return fmt.Errorf("failed to get mpp-y property: %w", err)
	}

	if mppX != "" && mppY != "" {
		var x, y float64
		fmt.Sscanf(mppX, "%f", &x)
		fmt.Sscanf(mppY, "%f", &y)
		slide.SlideMpp = (x + y) / 2.0
	}

	return nil
}
