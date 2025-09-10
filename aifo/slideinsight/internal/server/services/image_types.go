// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"
	"encoding/binary"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
)

// ImageTypesService is an interface that defines the methods for the image types service.
type ImageTypesService interface {
	GetImageTypes(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.ImageType, domain.PaginationInfo, error)
	GetImageTypesCount(ctx context.Context, search utils.SearchParams) (int, error)
	GetImageTypeByID(ctx context.Context, id string) (domain.ImageType, error)
	CreateImageType(ctx context.Context, imageType domain.ImageType) (domain.ImageType, error)
	UpdateImageType(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteImageType(ctx context.Context, id string) error
	Close()
}

// SlideHistogramsService is an interface that defines the methods for the slide histograms service.
type SlideHistogramsService interface {
	GetHistogramsBySlideUID(ctx context.Context, slideUID string) ([]domain.SlideHistogram, error)
	GetHistogramByID(ctx context.Context, id string) (domain.SlideHistogram, error)
	CreateHistogram(ctx context.Context, slideUID string, histogram domain.SlideHistogram) (domain.SlideHistogram, error)
	UpdateHistogram(ctx context.Context, id string, histogram domain.SlideHistogram) error
	DeleteHistogram(ctx context.Context, id string) error
	DeleteHistogramsBySlideUID(ctx context.Context, slideUID string) error
	Close()
}

// StainingProtocolsService is an interface that defines the methods for the staining protocols service.
type StainingProtocolsService interface {
	GetProtocolsBySlideUID(ctx context.Context, slideUID string) ([]domain.StainingProtocol, error)
	GetProtocolByID(ctx context.Context, id string) (domain.StainingProtocol, error)
	CreateProtocol(ctx context.Context, slideUID string, protocol domain.StainingProtocol) (domain.StainingProtocol, error)
	UpdateProtocol(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteProtocol(ctx context.Context, id string) error
	Close()
}

// imageTypesService implements ImageTypesService
type imageTypesService struct {
	*BaseService
	db                     ports.Database
	paginatedSearchService *PaginatedSearchService[ports.ImageType, domain.ImageType]
}

// slideHistogramsService implements SlideHistogramsService
type slideHistogramsService struct {
	*BaseService
	db ports.Database
}

// stainingProtocolsService implements StainingProtocolsService
type stainingProtocolsService struct {
	*BaseService
	db ports.Database
}

// Conversion helpers
var imageTypeConversionHelpers = DefaultConversionHelpers()

// convertImageTypeDBToDomain converts a database ImageType record to a domain ImageType model
func convertImageTypeDBToDomain(record ports.ImageType, helpers *ConversionHelpers) domain.ImageType {
	return domain.ImageType{
		ID:                record.ID,
		TypeUID:           record.TypeUID,
		Name:              record.Name,
		Description:       record.Description,
		Category:          record.Category,
		RequiresHistogram: record.RequiresHistogram,
		MetadataSchema:    record.MetadataSchema,
		IsActive:          record.IsActive,
		CreatedAt:         helpers.FormatTime(record.CreatedAt),
		UpdatedAt:         helpers.FormatTime(record.UpdatedAt),
	}
}

// convertSlideHistogramDBToDomain converts a database SlideHistogram record to a domain SlideHistogram model
func convertSlideHistogramDBToDomain(record ports.SlideHistogram) domain.SlideHistogram {
	// Convert binary histogram data to int slice for API response
	var counts []int
	if len(record.HistogramData) > 0 {
		counts = make([]int, record.BinCount)
		for i := 0; i < record.BinCount && i*4 < len(record.HistogramData); i++ {
			counts[i] = int(binary.LittleEndian.Uint32(record.HistogramData[i*4 : (i+1)*4]))
		}
	}

	return domain.SlideHistogram{
		ID:            record.ID,
		SlideUID:      record.SlideUID,
		ChannelIndex:  record.ChannelIndex,
		ChannelName:   record.ChannelName,
		BinCount:      record.BinCount,
		MinValue:      record.MinValue,
		MaxValue:      record.MaxValue,
		HistogramData: record.HistogramData,
		Counts:        counts,
		Metadata:      record.Metadata,
		CreatedAt:     imageTypeConversionHelpers.FormatTime(record.CreatedAt),
		UpdatedAt:     imageTypeConversionHelpers.FormatTime(record.UpdatedAt),
	}
}

// convertStainingProtocolDBToDomain converts a database StainingProtocol record to a domain StainingProtocol model
func convertStainingProtocolDBToDomain(record ports.StainingProtocol) domain.StainingProtocol {
	return domain.StainingProtocol{
		ID:             record.ID,
		SlideUID:       record.SlideUID,
		StainName:      record.StainName,
		StainType:      record.StainType,
		Concentration:  record.Concentration,
		IncubationTime: record.IncubationTime,
		AntibodyInfo:   record.AntibodyInfo,
		ExcitationNm:   record.ExcitationNm,
		EmissionNm:     record.EmissionNm,
		Metadata:       record.Metadata,
		CreatedAt:      imageTypeConversionHelpers.FormatTime(record.CreatedAt),
		UpdatedAt:      imageTypeConversionHelpers.FormatTime(record.UpdatedAt),
	}
}

// NewImageTypesService creates a new ImageTypesService
func NewImageTypesService(db ports.Database) ImageTypesService {
	baseService := NewBaseService(db)

	service := &imageTypesService{
		BaseService: baseService,
		db:          db,
	}

	// Set up paginated search service
	service.paginatedSearchService = NewPaginatedSearchService(
		func(ctx context.Context, search utils.SearchParams, limit, offset int) ([]ports.ImageType, error) {
			pagination := utils.PaginationParams{Page: (offset / limit) + 1, Limit: limit}
			return db.LoadAllImageTypes(ctx, search, pagination)
		},
		func(ctx context.Context, search utils.SearchParams) (int, error) {
			return db.GetImageTypesCount(ctx, search)
		},
		nil, // No plain pagination needed for image types
		nil, // No plain count needed for image types
		func(record ports.ImageType) domain.ImageType {
			return convertImageTypeDBToDomain(record, imageTypeConversionHelpers)
		},
	)

	return service
}

// NewSlideHistogramsService creates a new SlideHistogramsService
func NewSlideHistogramsService(db ports.Database) SlideHistogramsService {
	return &slideHistogramsService{
		BaseService: NewBaseService(db),
		db:          db,
	}
}

// NewStainingProtocolsService creates a new StainingProtocolsService
func NewStainingProtocolsService(db ports.Database) StainingProtocolsService {
	return &stainingProtocolsService{
		BaseService: NewBaseService(db),
		db:          db,
	}
}

// ImageTypesService implementation

func (s *imageTypesService) GetImageTypes(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.ImageType, domain.PaginationInfo, error) {
	return s.paginatedSearchService.GetWithPaginationAndSearch(ctx, params)
}

func (s *imageTypesService) GetImageTypesCount(ctx context.Context, search utils.SearchParams) (int, error) {
	return s.db.GetImageTypesCount(ctx, search)
}

func (s *imageTypesService) GetImageTypeByID(ctx context.Context, id string) (domain.ImageType, error) {
	record, err := s.db.GetImageTypeByID(ctx, id)
	if err != nil {
		log.Error("Failed to get image type", "error", err, "id", id)
		return domain.ImageType{}, err
	}
	return convertImageTypeDBToDomain(record, imageTypeConversionHelpers), nil
}

func (s *imageTypesService) CreateImageType(ctx context.Context, imageType domain.ImageType) (domain.ImageType, error) {
	if imageType.Name == "" {
		return domain.ImageType{}, errors.WithDetails(errors.ErrInvalidInput, "image type name cannot be empty")
	}

	if imageType.ID == "" {
		uid, err := s.GenerateShortUID()
		if err != nil {
			return domain.ImageType{}, err
		}
		imageType.ID = "img_type_" + uid
	}

	if imageType.TypeUID == "" {
		return domain.ImageType{}, errors.WithDetails(errors.ErrInvalidInput, "type UID cannot be empty")
	}

	// Get auth context for tenant ID (but image types are usually system-level with tenant_id = 0)
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return domain.ImageType{}, err
	}

	newImageType := ports.NewImageType{
		ID:                imageType.ID,
		TenantID:          0, // System-level by default, unless specified
		TypeUID:           imageType.TypeUID,
		Name:              imageType.Name,
		Description:       imageType.Description,
		Category:          imageType.Category,
		RequiresHistogram: imageType.RequiresHistogram,
		MetadataSchema:    imageType.MetadataSchema,
		IsActive:          true, // Default to active
	}

	// Only superadmins can create system-level image types
	if !authCtx.IsSuperAdmin {
		newImageType.TenantID = authCtx.TenantID
	}

	err = s.db.CreateImageType(ctx, newImageType)
	if err != nil {
		log.Error("Failed to create image type", "error", err)
		return domain.ImageType{}, err
	}

	return s.GetImageTypeByID(ctx, imageType.ID)
}

func (s *imageTypesService) UpdateImageType(ctx context.Context, id string, updates map[string]interface{}) error {
	// Convert map to structured updates
	updateStruct := ports.ImageTypeUpdates{}

	if name, ok := updates["name"].(string); ok {
		updateStruct.Name = &name
	}
	if description, ok := updates["description"].(string); ok {
		updateStruct.Description = &description
	}
	if category, ok := updates["category"].(string); ok {
		updateStruct.Category = &category
	}
	if requiresHistogram, ok := updates["requiresHistogram"].(bool); ok {
		updateStruct.RequiresHistogram = &requiresHistogram
	}
	if metadataSchema, ok := updates["metadataSchema"].(string); ok {
		updateStruct.MetadataSchema = &metadataSchema
	}
	if isActive, ok := updates["isActive"].(bool); ok {
		updateStruct.IsActive = &isActive
	}

	return s.db.UpdateImageType(ctx, id, updateStruct)
}

func (s *imageTypesService) DeleteImageType(ctx context.Context, id string) error {
	return s.db.DeleteImageType(ctx, id)
}

func (s *imageTypesService) Close() {
	// no-op
}

// SlideHistogramsService implementation

func (s *slideHistogramsService) GetHistogramsBySlideUID(ctx context.Context, slideUID string) ([]domain.SlideHistogram, error) {
	records, err := s.db.GetHistogramsBySlideUID(ctx, slideUID)
	if err != nil {
		log.Error("Failed to get histograms for slide", "error", err, "slideUID", slideUID)
		return nil, err
	}

	histograms := make([]domain.SlideHistogram, len(records))
	for i, record := range records {
		histograms[i] = convertSlideHistogramDBToDomain(record)
	}

	return histograms, nil
}

func (s *slideHistogramsService) GetHistogramByID(ctx context.Context, id string) (domain.SlideHistogram, error) {
	record, err := s.db.GetHistogramByID(ctx, id)
	if err != nil {
		log.Error("Failed to get histogram", "error", err, "id", id)
		return domain.SlideHistogram{}, err
	}
	return convertSlideHistogramDBToDomain(record), nil
}

func (s *slideHistogramsService) CreateHistogram(ctx context.Context, slideUID string, histogram domain.SlideHistogram) (domain.SlideHistogram, error) {
	if histogram.ID == "" {
		uid, err := s.GenerateShortUID()
		if err != nil {
			return domain.SlideHistogram{}, err
		}
		histogram.ID = uid
	}

	// Convert counts to binary data if provided
	var histogramData []byte
	if len(histogram.Counts) > 0 {
		histogramData = make([]byte, len(histogram.Counts)*4)
		for i, count := range histogram.Counts {
			binary.LittleEndian.PutUint32(histogramData[i*4:(i+1)*4], uint32(count))
		}
	} else {
		histogramData = histogram.HistogramData
	}

	newHistogram := ports.NewSlideHistogram{
		ID:            histogram.ID,
		SlideUID:      slideUID,
		ChannelIndex:  histogram.ChannelIndex,
		ChannelName:   histogram.ChannelName,
		BinCount:      histogram.BinCount,
		MinValue:      histogram.MinValue,
		MaxValue:      histogram.MaxValue,
		HistogramData: histogramData,
		Metadata:      histogram.Metadata,
	}

	err := s.db.CreateHistogram(ctx, newHistogram)
	if err != nil {
		log.Error("Failed to create histogram", "error", err)
		return domain.SlideHistogram{}, err
	}

	return s.GetHistogramByID(ctx, histogram.ID)
}

func (s *slideHistogramsService) UpdateHistogram(ctx context.Context, id string, histogram domain.SlideHistogram) error {
	// Convert counts to binary data if provided
	var histogramData []byte
	if len(histogram.Counts) > 0 {
		histogramData = make([]byte, len(histogram.Counts)*4)
		for i, count := range histogram.Counts {
			binary.LittleEndian.PutUint32(histogramData[i*4:(i+1)*4], uint32(count))
		}
	} else {
		histogramData = histogram.HistogramData
	}

	updateHistogram := ports.NewSlideHistogram{
		ID:            histogram.ID,
		SlideUID:      histogram.SlideUID,
		ChannelIndex:  histogram.ChannelIndex,
		ChannelName:   histogram.ChannelName,
		BinCount:      histogram.BinCount,
		MinValue:      histogram.MinValue,
		MaxValue:      histogram.MaxValue,
		HistogramData: histogramData,
		Metadata:      histogram.Metadata,
	}

	return s.db.UpdateHistogram(ctx, id, updateHistogram)
}

func (s *slideHistogramsService) DeleteHistogram(ctx context.Context, id string) error {
	return s.db.DeleteHistogram(ctx, id)
}

func (s *slideHistogramsService) DeleteHistogramsBySlideUID(ctx context.Context, slideUID string) error {
	return s.db.DeleteHistogramsBySlideUID(ctx, slideUID)
}

func (s *slideHistogramsService) Close() {
	// no-op
}

// StainingProtocolsService implementation

func (s *stainingProtocolsService) GetProtocolsBySlideUID(ctx context.Context, slideUID string) ([]domain.StainingProtocol, error) {
	records, err := s.db.GetProtocolsBySlideUID(ctx, slideUID)
	if err != nil {
		log.Error("Failed to get protocols for slide", "error", err, "slideUID", slideUID)
		return nil, err
	}

	protocols := make([]domain.StainingProtocol, len(records))
	for i, record := range records {
		protocols[i] = convertStainingProtocolDBToDomain(record)
	}

	return protocols, nil
}

func (s *stainingProtocolsService) GetProtocolByID(ctx context.Context, id string) (domain.StainingProtocol, error) {
	record, err := s.db.GetProtocolByID(ctx, id)
	if err != nil {
		log.Error("Failed to get protocol", "error", err, "id", id)
		return domain.StainingProtocol{}, err
	}
	return convertStainingProtocolDBToDomain(record), nil
}

func (s *stainingProtocolsService) CreateProtocol(ctx context.Context, slideUID string, protocol domain.StainingProtocol) (domain.StainingProtocol, error) {
	if protocol.ID == "" {
		uid, err := s.GenerateShortUID()
		if err != nil {
			return domain.StainingProtocol{}, err
		}
		protocol.ID = uid
	}

	if protocol.StainName == "" {
		return domain.StainingProtocol{}, errors.WithDetails(errors.ErrInvalidInput, "stain name cannot be empty")
	}

	newProtocol := ports.NewStainingProtocol{
		ID:             protocol.ID,
		SlideUID:       slideUID,
		StainName:      protocol.StainName,
		StainType:      protocol.StainType,
		Concentration:  protocol.Concentration,
		IncubationTime: protocol.IncubationTime,
		AntibodyInfo:   protocol.AntibodyInfo,
		ExcitationNm:   protocol.ExcitationNm,
		EmissionNm:     protocol.EmissionNm,
		Metadata:       protocol.Metadata,
	}

	err := s.db.CreateProtocol(ctx, newProtocol)
	if err != nil {
		log.Error("Failed to create protocol", "error", err)
		return domain.StainingProtocol{}, err
	}

	return s.GetProtocolByID(ctx, protocol.ID)
}

func (s *stainingProtocolsService) UpdateProtocol(ctx context.Context, id string, updates map[string]interface{}) error {
	// Convert map to structured updates
	updateStruct := ports.StainingProtocolUpdates{}

	if stainName, ok := updates["stainName"].(string); ok {
		updateStruct.StainName = &stainName
	}
	if stainType, ok := updates["stainType"].(string); ok {
		updateStruct.StainType = &stainType
	}
	if concentration, ok := updates["concentration"].(string); ok {
		updateStruct.Concentration = &concentration
	}
	if incubationTime, ok := updates["incubationTime"].(string); ok {
		updateStruct.IncubationTime = &incubationTime
	}
	if antibodyInfo, ok := updates["antibodyInfo"].(string); ok {
		updateStruct.AntibodyInfo = &antibodyInfo
	}
	if excitationNm, ok := updates["excitationNm"].(int); ok {
		updateStruct.ExcitationNm = &excitationNm
	}
	if emissionNm, ok := updates["emissionNm"].(int); ok {
		updateStruct.EmissionNm = &emissionNm
	}
	if metadata, ok := updates["metadata"].(string); ok {
		updateStruct.Metadata = &metadata
	}

	return s.db.UpdateProtocol(ctx, id, updateStruct)
}

func (s *stainingProtocolsService) DeleteProtocol(ctx context.Context, id string) error {
	return s.db.DeleteProtocol(ctx, id)
}

func (s *stainingProtocolsService) Close() {
	// no-op
}
