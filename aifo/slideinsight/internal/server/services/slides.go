// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	slidesService "aifo.dev/aifo/slideinsight/internal/server/services/slides"
	"aifo.dev/aifo/slideinsight/internal/utils"
	// "go.uber.org/zap"
)

// SlidesService re-exports the slides service interface for compatibility
type SlidesService = slidesService.SlidesService

// slidesServiceWrapper wraps the slides service
type slidesServiceWrapper struct {
	SlidesService
	db ports.Database
}

func (w *slidesServiceWrapper) canViewSlide(ctx context.Context, userID int, s ports.Slide) (bool, error) {
	allowed, err := w.db.UserHasRolePermission(ctx, userID, "slides.view")
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}
	allowed, err = w.db.HasObjectGrant(ctx, userID, "slides.view", "slide", s.ID)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}
	caseID, err := w.db.GetSlideCaseRelation(ctx, s.ID)
	if err != nil {
		return false, err
	}
	allowed, err = w.db.HasObjectGrant(ctx, userID, "slides.view", "case", caseID)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}
	studyIDs, err := w.db.GetCaseStudyRelations(ctx, caseID)
	if err != nil {
		return false, err
	}
	for _, sid := range studyIDs {
		allowed, err = w.db.HasObjectGrant(ctx, userID, "slides.view", "study", sid)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}

// authServiceAdapter adapts the main services BaseService to the slides package BaseService interface
type authServiceAdapter struct {
	*BaseService
}

// GetAuthContext implements the slides package BaseService interface
func (a *authServiceAdapter) GetAuthContext(ctx context.Context) (*slidesService.AuthContext, error) {
	authCtx, err := a.BaseService.GetAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	// Convert from main services AuthContext to slides package AuthContext
	return &slidesService.AuthContext{
		TenantID:     authCtx.TenantID,
		CreatorID:    authCtx.CreatorID,
		Email:        authCtx.Email,
		IsSuperAdmin: authCtx.IsSuperAdmin,
	}, nil
}

// NewSlidesService creates a new SlidesService using the slides package
func NewSlidesService(db ports.Database) SlidesService {
	base := NewBaseService(db)
	// Create adapter for the slides package
	adapter := &authServiceAdapter{BaseService: base}

	// Create the generic paginated search service
	paginatedSearchService := NewPaginatedSearchService(
		func(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.Slide, error) {
			pagination := utils.PaginationParams{Page: (offset / limit) + 1, Limit: limit}
			records, err := db.LoadAllSlides(ctx, search, pagination)
			if err != nil {
				return nil, err
			}
			return FilterByPermission(ctx, base, records, func(c context.Context, uid int, r ports.Slide) (bool, error) {
				wrapper := &slidesServiceWrapper{db: db}
				return wrapper.canViewSlide(c, uid, r)
			})
		},
		func(ctx context.Context, search utils.SearchParams) (int, error) {
			total, err := db.GetSlidesCount(ctx, search)
			if err != nil {
				return 0, err
			}
			pagination := utils.PaginationParams{Page: 1, Limit: total}
			records, err := db.LoadAllSlides(ctx, search, pagination)
			if err != nil {
				return 0, err
			}
			return CountByPermission(ctx, base, records, func(c context.Context, uid int, r ports.Slide) (bool, error) {
				wrapper := &slidesServiceWrapper{db: db}
				return wrapper.canViewSlide(c, uid, r)
			})
		},
		func(ctx context.Context, limit, offset int) ([]ports.Slide, error) {
			pagination := utils.PaginationParams{Page: (offset / limit) + 1, Limit: limit}
			records, err := db.LoadAllSlides(ctx, utils.SearchParams{}, pagination)
			if err != nil {
				return nil, err
			}
			return FilterByPermission(ctx, base, records, func(c context.Context, uid int, r ports.Slide) (bool, error) {
				wrapper := &slidesServiceWrapper{db: db}
				return wrapper.canViewSlide(c, uid, r)
			})
		},
		func(ctx context.Context) (int, error) {
			total, err := db.GetSlidesCount(ctx, utils.SearchParams{})
			if err != nil {
				return 0, err
			}
			pagination := utils.PaginationParams{Page: 1, Limit: total}
			records, err := db.LoadAllSlides(ctx, utils.SearchParams{}, pagination)
			if err != nil {
				return 0, err
			}
			return CountByPermission(ctx, base, records, func(c context.Context, uid int, r ports.Slide) (bool, error) {
				wrapper := &slidesServiceWrapper{db: db}
				return wrapper.canViewSlide(c, uid, r)
			})
		},
		func(record ports.Slide) domain.Slide {
			// Use a background context for image type fetching in batch operations
			// This is acceptable because image type data is not user-specific
			return convertSlideDBToDomainWithImageType(context.Background(), db, record)
		},
	)

	service := slidesService.NewSlidesService(db, paginatedSearchService, adapter)

	// Return wrapped service
	return &slidesServiceWrapper{
		SlidesService: service,
		db:            db,
	}
}

// convertSlideDBToDomain converts a database Slide record to a domain Slide model using conversion helpers
func convertSlideDBToDomain(record ports.Slide) domain.Slide {
	return ConvertDBToDomain(
		record,
		DefaultConversionHelpers(),
		convertSlideBase,
	)
}

// convertSlideDBToDomainWithImageType converts a database Slide record to a domain Slide model including image type ID
func convertSlideDBToDomainWithImageType(ctx context.Context, db ports.Database, record ports.Slide) domain.Slide {
	helpers := DefaultConversionHelpers()

	// Create the slide with the image type ID directly
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
		CreatorUID:  record.CreatorUID,
		CreatedAt:   helpers.FormatTime(record.CreatedAt),
		UpdatedAt:   helpers.FormatTime(record.UpdatedAt),
		DeletedAt:   helpers.FormatOptionalTime(record.DeletedAt),
		DeletedBy:   record.DeletedBy,
	}

	return slide
}

// convertSlideBase handles the basic slide conversion
func convertSlideBase(record ports.Slide, helpers *ConversionHelpers) domain.Slide {
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
		CreatorUID:  record.CreatorUID,
		CreatedAt:   helpers.FormatTime(record.CreatedAt),
		UpdatedAt:   helpers.FormatTime(record.UpdatedAt),
		DeletedAt:   helpers.FormatOptionalTime(record.DeletedAt),
		DeletedBy:   record.DeletedBy,
	}
}
