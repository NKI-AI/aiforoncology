// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"path/filepath"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/slides"
	"aifo.dev/aifo/slideinsight/internal/tiff"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
)

type RasterAnnotationsService interface {
	GetRasterAnnotations(ctx context.Context) ([]domain.Mask, error)
	GetRasterAnnotationsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Mask, domain.PaginationInfo, error)
	SaveMask(ctx context.Context, mask domain.Mask) (domain.Mask, error)
	GetMaskTile(ctx context.Context, slideUID string, maskUID string, z, x, y int) (domain.MaskTile, error)
	GetRasterAnnotationsForSlide(ctx context.Context, slideUID string) ([]domain.Mask, error)
	Close()
}

type rasterAnnotationsService struct {
	*BaseService
	db ports.Database
}

// MaskInfo holds mask metadata needed for efficient tile retrieval
type MaskInfo struct {
	MaskURI  string
	SlideUID string
}

// MaskInfo serializes to string for storage in MetadataCache
func (m MaskInfo) String() string {
	return fmt.Sprintf("%s|%s", m.MaskURI, m.SlideUID)
}

// ParseMaskInfo parses a string back into MaskInfo
func ParseMaskInfo(s string) MaskInfo {
	parts := strings.Split(s, "|")
	if len(parts) < 2 {
		return MaskInfo{}
	}
	return MaskInfo{
		MaskURI:  parts[0],
		SlideUID: parts[1],
	}
}

// NewRasterAnnotationsService creates a new RasterAnnotationsService
func NewRasterAnnotationsService(db ports.Database) RasterAnnotationsService {
	return &rasterAnnotationsService{
		BaseService: NewBaseService(db),
		db:          db,
	}
}

// convertRasterAnnotationDBToDomain converts a database Mask record to a domain Mask model
func convertRasterAnnotationDBToDomain(record ports.Mask) domain.Mask {
	// Parse labels from JSON string if present
	var labels domain.RasterLabels
	if record.Labels != "" {
		if err := labels.FromJSON(record.Labels); err != nil {
			log.Warn("Failed to parse labels for mask", "maskUID", record.MaskUID, "error", err)
			// Continue with nil labels rather than failing the entire request
			labels = nil
		}
	}

	// Create the mask with complete information
	return domain.Mask{
		MaskUID:    record.MaskUID,
		MaskName:   record.Name,
		MaskURI:    record.MaskURI,
		TilesURL:   record.TilesURL,
		SlideUID:   record.SlideUID, // Use external slide UID for domain
		Labels:     labels,
		MaskWidth:  record.MaskWidth,
		MaskHeight: record.MaskHeight,
		MaskMpp:    record.MaskMpp,
		ActorType:  record.ActorType,
		ActorID:    record.ActorID,
		Mutable:    record.Mutable,
		CreatedAt:  record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (s *rasterAnnotationsService) GetRasterAnnotations(ctx context.Context) ([]domain.Mask, error) {
	dbRecords, err := s.db.LoadAllMasks(ctx)
	if err != nil {
		return nil, errors.WithDetails(errors.ErrDatabaseQuery, "failed to load masks: %v", err)
	}

	masks := make([]domain.Mask, 0, len(dbRecords))
	for _, record := range dbRecords {
		mask := convertRasterAnnotationDBToDomain(record)
		masks = append(masks, mask)
	}

	return masks, nil
}

// GetRasterAnnotationsForSlide returns all masks for a specific slide
func (s *rasterAnnotationsService) GetRasterAnnotationsForSlide(ctx context.Context, slideUID string) ([]domain.Mask, error) {
	// Get all masks
	allMasks, err := s.GetRasterAnnotations(ctx)
	if err != nil {
		return nil, err
	}

	// Filter to return only those for this slide
	var slideMasks []domain.Mask
	for _, mask := range allMasks {
		if mask.SlideUID == slideUID {
			slideMasks = append(slideMasks, mask)
		}
	}

	return slideMasks, nil
}

// SaveMask saves a mask to the database and links it to a slide
func (s *rasterAnnotationsService) SaveMask(ctx context.Context, mask domain.Mask) (domain.Mask, error) {
	// SlideUID is required
	if mask.SlideUID == "" {
		return domain.Mask{}, errors.WithDetails(errors.ErrInvalidInput, "slideUid is required")
	}

	// Use the base service to get authentication context
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return domain.Mask{}, err
	}

	// Check if mask already exists (for cache invalidation)
	existingMask := false
	if mask.MaskUID != "" {
		_, err := s.db.GetMaskByUID(ctx, mask.MaskUID)
		if err == nil {
			existingMask = true
		}
	}

	// Check if slide exists
	exists, err := s.db.SlideExists(ctx, mask.SlideUID)
	if err != nil {
		return domain.Mask{}, errors.WithDetails(errors.ErrInternal, "failed to check if slide exists: %v", err)
	}
	if !exists {
		return domain.Mask{}, errors.WithDetails(errors.ErrSlideNotFound, "slide with ID '%s'", mask.SlideUID)
	}

	// We need to get the metadata from the mask file if not already populated
	if mask.MaskWidth == 0 || mask.MaskHeight == 0 {
		err = populateMaskMetadata(&mask)
		if err != nil {
			return domain.Mask{}, errors.WithDetails(errors.ErrInvalidInput, "failed to populate mask metadata: %v", err)
		}
	}

	// Log the mask metadata
	log.Info("Saving mask", "mask", mask)

	// Generate a default mask ID if none was provided
	if mask.MaskUID == "" {
		randomID, err := utils.GenerateFixedShortUID()
		if err != nil {
			return domain.Mask{}, errors.WithDetails(errors.ErrInternal, "failed to generate mask ID: %v", err)
		}
		mask.MaskUID = randomID
	}

	// Generate a default name if none was provided
	if mask.MaskName == "" {
		// Try to extract a name from the URI
		filename := filepath.Base(mask.MaskURI)
		// Remove file extension if present
		if dotIndex := strings.LastIndex(filename, "."); dotIndex != -1 {
			filename = filename[:dotIndex]
		}

		if filename != "" {
			mask.MaskName = fmt.Sprintf("%s Mask", filename)
		} else {
			// Fall back to slide ID based name
			mask.MaskName = fmt.Sprintf("Mask for %s", mask.SlideUID)
		}
	}

	// Determine format from URI if not provided
	format := "tiff" // default
	if mask.MaskURI != "" {
		if strings.HasSuffix(strings.ToLower(mask.MaskURI), ".png") {
			format = "png"
		}
	}

	// Create metadata JSON - for now just empty object
	metadata := "{}"

	// Convert labels to JSON string for database storage
	var labelsJSON string
	if mask.Labels != nil && len(mask.Labels) > 0 {
		if jsonStr, err := mask.Labels.ToJSON(); err != nil {
			return domain.Mask{}, errors.WithDetails(errors.ErrInvalidInput, "invalid labels format: %v", err)
		} else {
			labelsJSON = jsonStr
		}
	}

	dbMask := ports.NewMask{
		TenantID:   authCtx.TenantID,
		ActorType:  "model",
		ActorID:    1,
		CreatorID:  authCtx.CreatorID,
		SlideUID:   mask.SlideUID,
		MaskUID:    mask.MaskUID,
		Version:    1,
		Name:       mask.MaskName,
		MaskURI:    mask.MaskURI,
		TilesURL:   mask.TilesURL,
		Format:     format,
		MaskWidth:  mask.MaskWidth,
		MaskHeight: mask.MaskHeight,
		MaskMpp:    mask.MaskMpp,
		Labels:     labelsJSON,
		Metadata:   metadata,
	}

	err = s.db.CreateMask(ctx, dbMask)
	if err != nil {
		return domain.Mask{}, errors.WithDetails(errors.ErrDatabaseInsert, "failed to save mask: %v", err)
	}

	// If we updated an existing mask, invalidate its cache entries
	if existingMask {
		// No cache to invalidate as we removed the pyramidCache
	}

	return mask, nil
}

// GetMaskTile returns a specific tile of a mask
func (s *rasterAnnotationsService) GetMaskTile(ctx context.Context, slideUID, maskUID string, z, x, y int) (domain.MaskTile, error) {
	// Fetch from database
	maskRecord, err := s.db.GetMaskByUID(ctx, maskUID)
	if err != nil {
		return domain.MaskTile{}, errors.WithDetails(errors.ErrMaskNotFound, "mask ID: %s", maskUID)
	}

	// Create mask info
	maskInfo := MaskInfo{
		MaskURI:  maskRecord.MaskURI,
		SlideUID: maskRecord.SlideUID, // Use external slide UID
	}

	// Verify slide ownership
	if maskInfo.SlideUID != slideUID {
		return domain.MaskTile{}, errors.WithDetails(
			errors.ErrInvalidInput,
			"mask %s does not belong to slide %s",
			maskUID,
			slideUID,
		)
	}

	// Get tile fresh for each request
	img, err := getTileFresh(maskUID, maskInfo.MaskURI, z, x, y, false) // false = mask, not slide
	if err != nil {
		// Check for out of bounds error using proper error types
		if errors.IsOutOfBounds(err) {
			return domain.MaskTile{}, errors.ErrTileOutOfBounds
		}
		return domain.MaskTile{}, errors.WithDetails(errors.ErrInternal, "error generating mask tile: %v", err)
	}

	// Encode the image to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return domain.MaskTile{}, errors.WithDetails(errors.ErrInternal, "error encoding mask tile to PNG: %v", err)
	}

	return domain.MaskTile{
		Image:       buf.Bytes(),
		Format:      "png",
		ContentType: "image/png",
	}, nil
}

// Close releases resources associated with the service
func (s *rasterAnnotationsService) Close() {
	// No resources to close since we don't cache anything
}

// populateMaskMetadata reads the mask file to populate width, height, and other properties
func populateMaskMetadata(mask *domain.Mask) error {
	if mask.MaskURI == "" {
		return errors.ErrMaskUriInvalid
	}

	// Use the tiff package to open the mask file
	maskObj, err := tiff.Open(mask.MaskURI)
	if err != nil {
		return errors.WithDetails(errors.ErrInternal, "failed to open mask: %v", err)
	}
	defer maskObj.Close()

	// Get dimensions from level 0
	width, height, err := maskObj.LevelSize(0)
	if err != nil {
		return errors.WithDetails(errors.ErrMaskMetadataInvalid, "failed to get mask dimensions: %v", err)
	}
	mask.MaskWidth = int(width)
	mask.MaskHeight = int(height)

	// Try to get MPP if available
	mppX, mppY, err := maskObj.BaseResolution()
	if err != nil {
		// If we can't get MPP from the mask file, we'll log it but not fail
		log.Warn("Could not get MPP from mask file", "error", err)
		// We'll use a default MPP or leave it at 0, which will be handled elsewhere
		return nil
	}

	mask.MaskMpp = (mppX + mppY) / 2.0

	return nil
}

// GetRasterAnnotationsGeneric retrieves raster annotations using the generic pattern with pagination and search
func (s *rasterAnnotationsService) GetRasterAnnotationsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.Mask, domain.PaginationInfo, error) {
	dbRecords, paginationInfo, err := s.db.GetRasterAnnotationsGeneric(ctx, params)
	if err != nil {
		return nil, domain.PaginationInfo{}, errors.WithDetails(errors.ErrDatabaseQuery, "failed to load raster annotations: %v", err)
	}

	masks := make([]domain.Mask, 0, len(dbRecords))
	for _, record := range dbRecords {
		mask := convertRasterAnnotationDBToDomain(record)
		masks = append(masks, mask)
	}

	// Convert utils.PaginationInfo to domain.PaginationInfo
	domainPagination := domain.PaginationInfo{
		Page:       paginationInfo.Page,
		Limit:      paginationInfo.Limit,
		Total:      paginationInfo.Total,
		TotalPages: paginationInfo.TotalPages,
		HasNext:    paginationInfo.HasNext,
		HasPrev:    paginationInfo.HasPrev,
	}

	return masks, domainPagination, nil
}

// getTileFresh generates a tile fresh for each request without any caching
func getTileFresh(resourceUID string, uri string, z, x, y int, isSlide bool) (image.Image, error) {
	// Open resource fresh for each request
	var slideObj slides.Slide
	var err error
	if isSlide {
		slideObj, err = slides.OpenSlide(uri, "")
	} else {
		slideObj, err = slides.NewTiffAdapter(uri)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open resource %s: %w", uri, err)
	}
	defer slideObj.Close() // Always close after use

	// Create tile generator fresh
	bgColor := "#FFFFFF" // White background for slides
	if !isSlide {
		bgColor = "transparent" // Transparent for masks
	}

	generator, err := slides.NewXYZTileGenerator(slideObj, bgColor, !isSlide) // precise=true for masks
	if err != nil {
		return nil, fmt.Errorf("failed to create tile generator: %w", err)
	}

	// Generate the tile
	fastslideImg, err := generator.GetTile(z, x, y)
	if err != nil {
		return nil, fmt.Errorf("failed to get tile: %w", err)
	}

	// Convert fastslide.Image to image.Image
	return fastslideImg.ToGoImage()
}
