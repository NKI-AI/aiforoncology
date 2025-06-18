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
	"image/png"
	"log/slog"
	"path/filepath"
	"strings"

	"aifo.dev/aifo/slidescope/internal/datasources/database"
	"aifo.dev/aifo/slidescope/internal/server/domain"
	"aifo.dev/aifo/slidescope/internal/server/errors"
	"aifo.dev/aifo/slidescope/internal/tiff"
	"aifo.dev/aifo/slidescope/internal/utils"
)

type MasksService interface {
	GetMasks(ctx context.Context) ([]domain.Mask, error)
	SaveMask(ctx context.Context, mask domain.Mask) (domain.Mask, error)
	GetMaskTile(ctx context.Context, slideID string, maskID string, z, x, y int) (domain.MaskTile, error)
	GetMasksForSlide(ctx context.Context, slideID string) ([]domain.Mask, error)
	Close()
}

type masksService struct {
	db            database.Database
	metadataCache *utils.MetadataCache
	pyramidCache  *PyramidCache
}

// MaskInfo holds mask metadata needed for efficient tile retrieval
type MaskInfo struct {
	MaskURI string
	SlideID string
}

// MaskInfo serializes to string for storage in MetadataCache
func (m MaskInfo) String() string {
	return fmt.Sprintf("%s|%s", m.MaskURI, m.SlideID)
}

// ParseMaskInfo parses a string back into MaskInfo
func ParseMaskInfo(s string) MaskInfo {
	parts := strings.Split(s, "|")
	if len(parts) < 2 {
		return MaskInfo{}
	}
	return MaskInfo{
		MaskURI: parts[0],
		SlideID: parts[1],
	}
}

// NewMasksService creates a new MasksService
func NewMasksService(db database.Database) MasksService {
	// Create a metadata cache for masks (using generic cache from utils)
	metadataCache := utils.NewMetadataCache(1000) // Cache up to 1000 mask metadata entries

	// Create a PyramidCache for mask tile generators
	// Using 'false' for isSlide since this is for masks
	// "transparent" is typically used for masks background
	// true for precise tile generation
	pyramidCache := NewPyramidCache(100, 100, false, "transparent", true)

	return &masksService{
		db:            db,
		metadataCache: metadataCache,
		pyramidCache:  pyramidCache,
	}
}

func (s *masksService) GetMasks(ctx context.Context) ([]domain.Mask, error) {
	dbRecords, err := s.db.LoadAllMasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load masks: %w", err)
	}

	masks := make([]domain.Mask, 0, len(dbRecords))
	for _, record := range dbRecords {
		// Create the mask with complete information but without detailed slide info
		mask := domain.Mask{
			MaskID:     record.MaskID,
			MaskName:   record.Name,
			MaskURI:    record.MaskURI,
			TilesURL:   record.TilesURL,
			SlideID:    record.SlideID,
			MaskWidth:  record.MaskWidth,
			MaskHeight: record.MaskHeight,
			MaskMpp:    record.MaskMpp,
		}

		// We're not including the detailed slide object as requested by the user
		// Just the slideId field above is sufficient

		masks = append(masks, mask)
	}

	return masks, nil
}

// GetMasksForSlide returns all masks for a specific slide
func (s *masksService) GetMasksForSlide(ctx context.Context, slideID string) ([]domain.Mask, error) {
	// Get all masks
	allMasks, err := s.GetMasks(ctx)
	if err != nil {
		return nil, err
	}

	// Filter to return only those for this slide
	var slideMasks []domain.Mask
	for _, mask := range allMasks {
		if mask.SlideID == slideID {
			slideMasks = append(slideMasks, mask)
		}
	}

	return slideMasks, nil
}

// SaveMask saves a mask to the database and links it to a slide
func (s *masksService) SaveMask(ctx context.Context, mask domain.Mask) (domain.Mask, error) {
	// SlideID is required
	if mask.SlideID == "" {
		return domain.Mask{}, errors.WithDetails(errors.ErrInvalidInput, "slideId is required")
	}

	// Check if mask already exists (for cache invalidation)
	existingMask := false
	if mask.MaskID != "" {
		_, err := s.db.GetMaskByID(ctx, mask.MaskID)
		if err == nil {
			existingMask = true
		}
	}

	// Check if slide exists
	exists, err := s.db.SlideExists(ctx, mask.SlideID)
	if err != nil {
		return domain.Mask{}, errors.WithDetails(errors.ErrInternal, "failed to check if slide exists: %v", err)
	}
	if !exists {
		return domain.Mask{}, errors.WithDetails(errors.ErrSlideNotFound, "slide with ID '%s'", mask.SlideID)
	}

	// We need to get the metadata from the mask file if not already populated
	if mask.MaskWidth == 0 || mask.MaskHeight == 0 {
		err = populateMaskMetadata(&mask)
		if err != nil {
			return domain.Mask{}, errors.WithDetails(errors.ErrInvalidInput, "failed to populate mask metadata: %v", err)
		}
	}

	// Log the mask metadata
	slog.Info("Saving mask", "mask", mask)

	// Generate a default mask ID if none was provided
	if mask.MaskID == "" {
		randomID, err := utils.GenerateFixedShortID()
		if err != nil {
			return domain.Mask{}, errors.WithDetails(errors.ErrInternal, "failed to generate mask ID: %v", err)
		}
		mask.MaskID = randomID
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
			mask.MaskName = fmt.Sprintf("Mask for %s", mask.SlideID)
		}
	}

	dbMask := database.NewMask{
		SlideID:    mask.SlideID,
		MaskID:     mask.MaskID,
		Name:       mask.MaskName,
		MaskURI:    mask.MaskURI,
		TilesURL:   mask.TilesURL,
		MaskWidth:  mask.MaskWidth,
		MaskHeight: mask.MaskHeight,
		MaskMpp:    mask.MaskMpp,
	}

	err = s.db.CreateMask(ctx, dbMask)
	if err != nil {
		return domain.Mask{}, fmt.Errorf("failed to save mask: %w", err)
	}

	// If we updated an existing mask, invalidate its cache entries
	if existingMask {
		s.metadataCache.Remove(mask.MaskID)
		s.pyramidCache.Remove(mask.MaskID)
	}

	return mask, nil
}

// GetMaskTile returns a specific tile of a mask
func (s *masksService) GetMaskTile(ctx context.Context, slideID, maskID string, z, x, y int) (domain.MaskTile, error) {
	// Try to get mask info from the metadata cache first
	infoStr, found := s.metadataCache.Get(maskID)
	var maskInfo MaskInfo

	if !found {
		// Cache miss - fetch from database
		slog.Debug("Metadata cache miss", "maskID", maskID)
		maskRecord, err := s.db.GetMaskByID(ctx, maskID)
		if err != nil {
			return domain.MaskTile{}, errors.WithDetails(errors.ErrMaskNotFound, "mask ID: %s", maskID)
		}

		// Store in cache for future use
		maskInfo = MaskInfo{
			MaskURI: maskRecord.MaskURI,
			SlideID: maskRecord.SlideID,
		}
		s.metadataCache.Put(maskID, maskInfo.String())
		slog.Debug("Added mask to metadata cache", "maskID", maskID)
	} else {
		slog.Debug("Metadata cache hit", "maskID", maskID)
		maskInfo = ParseMaskInfo(infoStr)
	}

	// Verify slide ownership
	if maskInfo.SlideID != slideID {
		return domain.MaskTile{}, errors.WithDetails(
			errors.ErrInvalidInput,
			"mask %s does not belong to slide %s",
			maskID,
			slideID,
		)
	}

	// Use the tile cache to get the tile
	img, err := s.pyramidCache.GetTile(maskID, maskInfo.MaskURI, z, x, y)
	if err != nil {
		// Check for out of bounds error patterns
		if strings.Contains(err.Error(), "out of bounds") ||
			strings.Contains(err.Error(), "outside slide bounds") {
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
func (s *masksService) Close() {
	// Clear the caches
	s.metadataCache.Clear()
	s.pyramidCache.Clear()

	// Close underlying resource cache through pyramid cache
	// Assuming pyramid cache has a reference to the resource cache
}

// populateMaskMetadata reads the mask file to populate width, height, and other properties
func populateMaskMetadata(mask *domain.Mask) error {
	if mask.MaskURI == "" {
		return fmt.Errorf("mask URI is empty")
	}

	// Use the tiff package to open the mask file
	maskObj, err := tiff.Open(mask.MaskURI)
	if err != nil {
		return fmt.Errorf("failed to open mask: %w", err)
	}
	defer maskObj.Close()

	// Get dimensions from level 0
	width, height, err := maskObj.LevelSize(0)
	if err != nil {
		return fmt.Errorf("failed to get mask dimensions: %w", err)
	}
	mask.MaskWidth = int(width)
	mask.MaskHeight = int(height)

	// Try to get MPP if available
	mppX, mppY, err := maskObj.BaseResolution()
	if err != nil {
		// If we can't get MPP from the mask file, we'll log it but not fail
		slog.Warn("Could not get MPP from mask file", "error", err)
		// We'll use a default MPP or leave it at 0, which will be handled elsewhere
		return nil
	}

	mask.MaskMpp = (mppX + mppY) / 2.0

	return nil
}
