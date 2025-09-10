// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package slides

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"path/filepath"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/slides"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
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

// PyramidCache is a placeholder for the PyramidCache type - kept for interface compatibility
type PyramidCache interface {
	GetTile(slideUID string, uri string, z, x, y int) (image.Image, error)
	Remove(slideUID string)
	Clear()
	Stats() map[string]uint64
}

// PaginatedSearchService is a placeholder for the generic search service
type PaginatedSearchService[TDB any, TDomain any] interface {
	GetWithPaginationAndSearch(ctx context.Context, params utils.PaginationAndSearchParams) ([]TDomain, domain.PaginationInfo, error)
}

// BaseService and AuthContext need to be provided by the main services package
// These interfaces allow the slides service to get authentication context without circular imports
type BaseService interface {
	GetAuthContext(ctx context.Context) (*AuthContext, error)
}

type AuthContext struct {
	TenantID     int
	CreatorID    int
	Email        string
	IsSuperAdmin bool
}

// SlidesService is an interface that defines the methods for the slides service.
// Interface is needed for mocking in tests.
type SlidesService interface {
	GetSlides(ctx context.Context, pagination utils.PaginationParams) ([]domain.Slide, domain.PaginationInfo, error)
	GetSlidesGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Slide, domain.PaginationInfo, error)
	GetSlidesCount(ctx context.Context) (int, error)
	GetSlidesByCaseUID(ctx context.Context, caseUID string) ([]domain.Slide, error)
	GetSlideByUID(ctx context.Context, slideUID string) (domain.Slide, error)
	GetSlideMetadata(ctx context.Context, slideUID string) (domain.SlideMetadata, error)
	GetSlideTile(ctx context.Context, slideUID string, z, x, y int, format string, quality int) (domain.SlideTile, error)
	SaveSlide(ctx context.Context, newSlide domain.Slide) (domain.Slide, error)
	SoftDeleteSlide(ctx context.Context, slideUID string, deletedBy int) error
	Close()
}

type slidesService struct {
	db                     ports.Database
	paginatedSearchService PaginatedSearchService[ports.Slide, domain.Slide]
	baseService            BaseService
}

// convertSlideDBToDomain converts a database Slide record to a domain Slide model
func convertSlideDBToDomain(record ports.Slide) domain.Slide {
	return domain.Slide{
		CaseID:      record.CaseID,
		CaseUID:     record.CaseUID,
		SlideID:     record.ID,       // Internal database ID
		SlideUID:    record.SlideUID, // External identifier
		SlideName:   record.SlideName,
		SlideURI:    record.SlideURI,
		SlideWidth:  record.SlideWidth,
		SlideHeight: record.SlideHeight,
		SlideMpp:    record.SlideMpp,
		ImageTypeId: record.ImageTypeID, // Include the image type ID reference
		Metadata:    record.Metadata,    // Include the metadata field
	}
}

// convertSlideWithImageType converts a database Slide record to a domain Slide model including image type ID
func (s *slidesService) convertSlideWithImageType(ctx context.Context, record ports.Slide) domain.Slide {
	slide := domain.Slide{
		CaseID:      record.CaseID,
		CaseUID:     record.CaseUID,
		SlideID:     record.ID,       // Internal database ID
		SlideUID:    record.SlideUID, // External identifier
		SlideName:   record.SlideName,
		SlideURI:    record.SlideURI,
		SlideWidth:  record.SlideWidth,
		SlideHeight: record.SlideHeight,
		SlideMpp:    record.SlideMpp,
		ImageTypeId: record.ImageTypeID, // Just include the ID reference
		Metadata:    record.Metadata,    // Include the metadata field
	}

	return slide
}

// NewSlidesService creates a new SlidesService
// NOTE: This function is not directly callable due to interface dependencies.
// It should be called from the main services package which provides the concrete implementations.
func NewSlidesService(db ports.Database, paginatedSearchService PaginatedSearchService[ports.Slide, domain.Slide], baseService BaseService) SlidesService {
	return &slidesService{
		db:                     db,
		paginatedSearchService: paginatedSearchService,
		baseService:            baseService,
	}
}

// GetSlides retrieves all slides from the database with pagination support
func (s *slidesService) GetSlides(ctx context.Context, pagination utils.PaginationParams) ([]domain.Slide, domain.PaginationInfo, error) {
	// Convert PaginationParams to PaginationAndSearchParams for generic service
	params := utils.PaginationAndSearchParams{
		PaginationParams: pagination,
		SearchParams:     utils.SearchParams{}, // Empty search params for simple pagination
	}
	return s.paginatedSearchService.GetWithPaginationAndSearch(ctx, params)
}

// GetSlideByUID retrieves a specific slide by its slide_uid
func (s *slidesService) GetSlideByUID(ctx context.Context, slideUID string) (domain.Slide, error) {
	dbRecord, err := s.db.GetSlideByUID(ctx, slideUID)
	if err != nil {
		return domain.Slide{}, errors.WithDetails(errors.ErrSlideNotFound, "slide with ID '%s'", slideUID)
	}

	return s.convertSlideWithImageType(ctx, dbRecord), nil
}

func (s *slidesService) GetSlidesByCaseUID(ctx context.Context, caseUID string) ([]domain.Slide, error) {
	dbRecords, err := s.db.GetSlidesByCaseUID(ctx, caseUID)
	if err != nil {
		log.Error("GetSlidesByCaseUID failed", "error", err, "caseUID", caseUID)
		return nil, fmt.Errorf("failed to load slides: %w", err)
	}

	slides := make([]domain.Slide, 0, len(dbRecords))
	for _, record := range dbRecords {
		slides = append(slides, s.convertSlideWithImageType(ctx, record))
	}

	return slides, nil
}

// SaveSlide saves a slide to the database
func (s *slidesService) SaveSlide(ctx context.Context, slide domain.Slide) (domain.Slide, error) {
	// Check if slide already exists (for cache invalidation)
	existingSlide := false
	if slide.SlideUID != "" {
		_, err := s.db.GetSlideByUID(ctx, slide.SlideUID)
		if err == nil {
			existingSlide = true
		}
	}

	// Generate a random UID if none is provided
	if slide.SlideUID == "" {
		// Generate a unique slide UID using the utility function
		randomUID, err := utils.GenerateFixedShortUID()
		if err != nil {
			return domain.Slide{}, errors.WithDetails(errors.ErrInternal, "failed to generate slide UID: %v", err)
		}
		slide.SlideUID = randomUID
	}

	// Generate a default name if none is provided
	if slide.SlideName == "" {
		// Extract a name from the URI if possible
		filename := filepath.Base(slide.SlideURI)
		if filename != "" && filename != "." && filename != "/" {
			// Remove file extension if present
			slide.SlideName = strings.TrimSuffix(filename, filepath.Ext(filename))
		} else {
			// Fall back to UID-based name
			slide.SlideName = fmt.Sprintf("Slide %s", slide.SlideUID)
		}
	}

	// Check if slide with the same UID already exists
	if !existingSlide {
		exists, err := s.db.SlideExists(ctx, slide.SlideUID)
		if err != nil {
			return domain.Slide{}, errors.WithDetails(errors.ErrInternal, "failed to check if slide exists: %v", err)
		}
		if exists {
			return domain.Slide{}, errors.WithDetails(errors.ErrAlreadyExists, "slide with UID '%s'", slide.SlideUID)
		}
	}

	// We need to get the metadata from the slide file
	err := populateSlideMetadata(&slide)
	if err != nil {
		return domain.Slide{}, errors.WithDetails(errors.ErrInvalidInput, "failed to populate slide metadata: %v", err)
	}

	// Log the slide metadata
	log.Info("Saving slide", "slide", slide)

	// Get authentication context properly
	authCtx, err := s.baseService.GetAuthContext(ctx)
	if err != nil {
		return domain.Slide{}, errors.WithDetails(errors.ErrInvalidInput, "not authenticated: %v", err)
	}
	log.Info("Context", "context", authCtx)

	dbSlide := ports.NewSlide{
		CaseID:      slide.CaseID,
		SlideID:     slide.SlideUID, // Note: NewSlide.SlideID maps to the external UID
		SlideName:   slide.SlideName,
		SlideURI:    slide.SlideURI,
		SlideWidth:  slide.SlideWidth,
		SlideHeight: slide.SlideHeight,
		SlideMpp:    slide.SlideMpp,
		ImageTypeID: slide.ImageTypeId,
		Metadata:    slide.Metadata,    // Include the metadata field
		CreatorID:   authCtx.CreatorID, // Set the creator ID from auth context
		TenantID:    authCtx.TenantID,  // Set the tenant ID from auth context
	}

	err = s.db.CreateSlide(ctx, dbSlide)
	if err != nil {
		return domain.Slide{}, errors.WithDetails(errors.ErrInternal, "failed to save slide: %v", err)
	}

	// If we updated an existing slide, invalidate its cache entries
	if existingSlide {
		// s.metadataCache.Remove(slide.SlideUID) // Removed cache dependency
		// s.pyramidCache.Remove(slide.SlideURI) // Use SlideURI instead of SlideUID
	}

	return slide, nil
}

// Close releases all resources held by the service
func (s *slidesService) Close() {
	// Clear the caches
	// s.metadataCache.Clear() // Removed cache dependency
	// s.pyramidCache.Clear()
}

// GetSlidesCount retrieves the total count of slides from the database
func (s *slidesService) GetSlidesCount(ctx context.Context) (int, error) {
	return s.db.GetSlidesCount(ctx, utils.SearchParams{})
}

// GetSlidesGeneric retrieves slides using the new generic pattern
func (s *slidesService) GetSlidesGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Slide, domain.PaginationInfo, error) {
	return s.paginatedSearchService.GetWithPaginationAndSearch(ctx, params)
}

// SoftDeleteSlide soft deletes a slide
func (s *slidesService) SoftDeleteSlide(ctx context.Context, slideUID string, deletedBy int) error {
	err := s.db.SoftDeleteSlide(ctx, slideUID, deletedBy)
	if err != nil {
		return fmt.Errorf("failed to soft delete slide: %w", err)
	}
	return nil
}

// ChannelInfo represents channel metadata for JSON serialization
type ChannelInfo struct {
	Name      string `json:"name"`
	Biomarker string `json:"biomarker,omitempty"`
	Color     string `json:"color,omitempty"` // Hex color string like "#ff00ff"
}

// SlideMetadataJSON represents the JSON structure for slide metadata
type SlideMetadataJSON struct {
	Channels map[string]ChannelInfo `json:"channels,omitempty"`
}

// populateChannelMetadata extracts and populates channel metadata for spectral slides
func populateChannelMetadata(slide *domain.Slide, slideObj slides.Slide) error {
	log.Info("Populating channel metadata", "slideURI", slide.SlideURI)
	// Check if the slide is spectral
	if !slideObj.IsSpectral() {
		log.Info("Not a spectral slide", "slideURI", slide.SlideURI)
		// Set default image type for non-spectral slides if not already set
		if slide.ImageTypeId == "" {
			slide.ImageTypeId = "img_type_unspec"
		}
		return nil // Not an error, just not a spectral slide
	}

	log.Info("Processing spectral slide", "slideURI", slide.SlideURI)

	// Set image type for spectral slides
	slide.ImageTypeId = "img_type_fluor" // Spectral/multiplex slides are fluorescence

	// Get channel metadata
	channelMetadata, err := slideObj.GetChannelMetadata()
	if err != nil {
		return fmt.Errorf("failed to get channel metadata: %w", err)
	}

	if len(channelMetadata) == 0 {
		log.Info("No channel metadata found for spectral slide", "slideURI", slide.SlideURI)
		return nil // No channels to process
	}

	log.Info("Found channel metadata", "slideURI", slide.SlideURI, "channelCount", len(channelMetadata))

	// Build the metadata JSON structure
	metadata := SlideMetadataJSON{
		Channels: make(map[string]ChannelInfo),
	}

	for i, ch := range channelMetadata {
		channelKey := fmt.Sprintf("%d", i) // Use index as key

		// Convert RGB color to hex string
		colorHex := fmt.Sprintf("#%02x%02x%02x", ch.Color[0], ch.Color[1], ch.Color[2])

		metadata.Channels[channelKey] = ChannelInfo{
			Name:      ch.Name,
			Biomarker: ch.Biomarker,
			Color:     colorHex,
		}
	}

	// Marshall to JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal channel metadata to JSON: %w", err)
	}

	// Store in the slide as raw JSON bytes
	slide.Metadata = json.RawMessage(metadataJSON)

	log.Info("Successfully populated channel metadata", "slideURI", slide.SlideURI, "metadataLength", len(slide.Metadata))

	return nil
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

	// Get MPP (microns per pixel) using the new direct method
	mppX, mppY, err := slideObj.Mpp()
	if err != nil {
		return fmt.Errorf("failed to get MPP values: %w", err)
	}

	// Calculate average MPP
	slide.SlideMpp = (mppX + mppY) / 2.0

	// Extract and populate channel metadata for spectral slides
	if err := populateChannelMetadata(slide, slideObj); err != nil {
		log.Error("Failed to populate channel metadata", "error", err, "slideURI", slide.SlideURI)
		// Don't fail the whole operation for channel metadata errors
	}

	// Ensure we always have a valid image type ID set
	if slide.ImageTypeId == "" {
		log.Info("No image type set, using default", "slideURI", slide.SlideURI)
		slide.ImageTypeId = "img_type_unspec"
	}

	return nil
}
