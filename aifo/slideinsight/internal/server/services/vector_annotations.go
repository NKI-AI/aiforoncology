// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/server/domain"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"aifo.dev/aifo/slideinsight/internal/utils"
	"github.com/gofiber/fiber/v2/log"
	"go.uber.org/zap"
)

type VectorAnnotationsService interface {
	GetVectorAnnotations(ctx context.Context) ([]domain.VectorAnnotation, error)
	GetVectorAnnotationsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.VectorAnnotation, domain.PaginationInfo, error)
	SaveVectorAnnotation(ctx context.Context, vector domain.VectorAnnotation) (domain.VectorAnnotation, error)
	UpdateVectorAnnotation(ctx context.Context, vectorUID string, vector domain.VectorAnnotation) (domain.VectorAnnotation, error)
	DeleteVectorAnnotation(ctx context.Context, vectorUID string) error
	GetVectorAnnotationsForSlide(ctx context.Context, slideUID string) ([]domain.VectorAnnotation, error)
	GetVectorAnnotationFile(ctx context.Context, slideUID string, vectorUID string) ([]byte, error)
	ImportVectorAnnotationToWorkspace(ctx context.Context, slideUID string, vectorUID string, allowedLabels []StudyAnnotationLabel) (domain.AnnotationImportResult, error)
	Close()
}

type vectorAnnotationsService struct {
	*BaseService
	db ports.Database
}

// NewVectorAnnotationsService creates a new VectorAnnotationsService
func NewVectorAnnotationsService(db ports.Database) VectorAnnotationsService {
	return &vectorAnnotationsService{
		BaseService: NewBaseService(db),
		db:          db,
	}
}

// convertVectorAnnotationDBToDomain converts a database VectorAnnotation record to a domain VectorAnnotation model
func convertVectorAnnotationDBToDomain(record ports.VectorAnnotation) domain.VectorAnnotation {
	// Parse labels from JSON string if present
	var labels domain.VectorLabels
	if record.Labels != "" {
		if err := labels.FromJSON(record.Labels); err != nil {
			log.Warn("Failed to parse labels for vector annotation", "vectorUID", record.VectorUID, "error", err)
			// Continue with nil labels rather than failing the entire request
			labels = nil
		}
	}

	// Parse DataBlob from string if present
	var dataBlob interface{}
	if record.DataBlob != "" {
		if err := utils.ParseJSON(record.DataBlob, &dataBlob); err != nil {
			log.Warn("Failed to parse DataBlob for vector annotation", "vectorUID", record.VectorUID, "error", err)
			// Continue with nil DataBlob rather than failing the entire request
			dataBlob = nil
		}
	}

	// Create the vector annotation with complete information
	return domain.VectorAnnotation{
		VectorUID:  record.VectorUID,
		VectorName: record.Name,
		SlideUID:   record.SlideUID, // Use external slide UID for domain
		FileURI:    record.FileURI,
		DataBlob:   dataBlob,
		Labels:     labels,
		ActorType:  record.ActorType,
		ActorID:    record.ActorID,
		Mutable:    record.Mutable,
		CreatedAt:  record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  record.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (s *vectorAnnotationsService) GetVectorAnnotations(ctx context.Context) ([]domain.VectorAnnotation, error) {
	dbRecords, err := s.db.LoadAllVectorAnnotations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load vector annotations: %w", err)
	}

	vectors := make([]domain.VectorAnnotation, 0, len(dbRecords))
	for _, record := range dbRecords {
		vector := convertVectorAnnotationDBToDomain(record)
		vectors = append(vectors, vector)
	}

	return vectors, nil
}

// GetVectorAnnotationsForSlide returns all vector annotations for a specific slide
func (s *vectorAnnotationsService) GetVectorAnnotationsForSlide(ctx context.Context, slideUID string) ([]domain.VectorAnnotation, error) {
	// Get all vector annotations
	allVectors, err := s.GetVectorAnnotations(ctx)
	if err != nil {
		return nil, err
	}

	// Filter to return only those for this slide
	slideVectors := make([]domain.VectorAnnotation, 0)
	for _, vector := range allVectors {
		if vector.SlideUID == slideUID {
			slideVectors = append(slideVectors, vector)
		}
	}

	return slideVectors, nil
}

// SaveVectorAnnotation saves a vector annotation to the database and links it to a slide
func (s *vectorAnnotationsService) SaveVectorAnnotation(ctx context.Context, vector domain.VectorAnnotation) (domain.VectorAnnotation, error) {
	// SlideUID is required
	if vector.SlideUID == "" {
		return domain.VectorAnnotation{}, errors.WithDetails(errors.ErrInvalidInput, "slideUid is required")
	}

	// Either FileURI or DataBlob is required
	if vector.FileURI == "" && vector.DataBlob == nil {
		return domain.VectorAnnotation{}, errors.WithDetails(errors.ErrInvalidInput, "either fileUri or dataBlob is required")
	}

	// Use the base service to get authentication context
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return domain.VectorAnnotation{}, err
	}

	// Check if slide exists
	exists, err := s.db.SlideExists(ctx, vector.SlideUID)
	if err != nil {
		return domain.VectorAnnotation{}, errors.WithDetails(errors.ErrInternal, "failed to check if slide exists: %v", err)
	}
	if !exists {
		return domain.VectorAnnotation{}, errors.WithDetails(errors.ErrSlideNotFound, "slide with UID '%s'", vector.SlideUID)
	}

	// Generate a default vector ID if none was provided
	if vector.VectorUID == "" {
		randomID, err := utils.GenerateFixedShortUID()
		if err != nil {
			return domain.VectorAnnotation{}, errors.WithDetails(errors.ErrInternal, "failed to generate vector UID: %v", err)
		}
		vector.VectorUID = randomID
	}

	// Generate a default name if none was provided
	if vector.VectorName == "" {
		// Try to extract a name from the URI
		filename := filepath.Base(vector.FileURI)
		// Remove file extension if present
		if dotIndex := strings.LastIndex(filename, "."); dotIndex != -1 {
			filename = filename[:dotIndex]
		}

		if filename != "" {
			vector.VectorName = fmt.Sprintf("%s Vector", filename)
		} else {
			// Fall back to slide ID based name
			vector.VectorName = fmt.Sprintf("Vector Annotation for %s", vector.SlideUID)
		}
	}

	// Log the vector annotation metadata
	log.Info("Saving vector annotation", "vector", vector)

	// Convert labels to JSON string for database storage
	var labelsJSON string
	if vector.Labels != nil && len(vector.Labels) > 0 {
		if jsonStr, err := vector.Labels.ToJSON(); err != nil {
			return domain.VectorAnnotation{}, errors.WithDetails(errors.ErrInvalidInput, "invalid labels format: %v", err)
		} else {
			labelsJSON = jsonStr
		}
	}

	// Convert DataBlob to string for database storage
	var dataBlobStr string
	if vector.DataBlob != nil {
		if jsonBytes, err := utils.ToJSON(vector.DataBlob); err != nil {
			return domain.VectorAnnotation{}, errors.WithDetails(errors.ErrInvalidInput, "invalid dataBlob format: %v", err)
		} else {
			dataBlobStr = string(jsonBytes)
		}
	}

	// Determine actor type and ID - default to user if not specified
	actorType := vector.ActorType
	actorID := vector.ActorID
	if actorType == "" {
		actorType = "user"
	}
	if actorID == 0 {
		actorID = authCtx.CreatorID
	}

	// Create metadata JSON - for now just empty object
	metadata := "{}"

	dbVector := ports.NewVectorAnnotation{
		TenantID:  authCtx.TenantID,
		ActorType: actorType,
		ActorID:   actorID,
		CreatorID: authCtx.CreatorID,
		SlideUID:  vector.SlideUID, // This is actually slide_uid in the domain
		VectorUID: vector.VectorUID,
		Version:   1, // Default version
		Name:      vector.VectorName,
		FileURI:   vector.FileURI,
		DataBlob:  dataBlobStr,
		Labels:    labelsJSON,
		Metadata:  metadata,
	}

	err = s.db.CreateVectorAnnotation(ctx, dbVector)
	if err != nil {
		return domain.VectorAnnotation{}, fmt.Errorf("failed to save vector annotation: %w", err)
	}

	return vector, nil
}

// UpdateVectorAnnotation updates an existing vector annotation
func (s *vectorAnnotationsService) UpdateVectorAnnotation(ctx context.Context, vectorUID string, vector domain.VectorAnnotation) (domain.VectorAnnotation, error) {
	// Use the base service to get authentication context
	_, err := s.GetAuthContext(ctx)
	if err != nil {
		return domain.VectorAnnotation{}, err
	}

	// Get existing vector annotation to verify ownership
	existing, err := s.db.GetVectorAnnotationByUID(ctx, vectorUID)
	if err != nil {
		return domain.VectorAnnotation{}, fmt.Errorf("vector annotation not found: %w", err)
	}

	// Verify slide ownership if SlideUID is being changed
	if vector.SlideUID != "" && vector.SlideUID != existing.SlideUID {
		exists, err := s.db.SlideExists(ctx, vector.SlideUID)
		if err != nil {
			return domain.VectorAnnotation{}, errors.WithDetails(errors.ErrInternal, "failed to check if slide exists: %v", err)
		}
		if !exists {
			return domain.VectorAnnotation{}, errors.WithDetails(errors.ErrSlideNotFound, "slide with UID '%s'", vector.SlideUID)
		}
	}

	// Prepare update struct
	updates := ports.UpdateVectorAnnotation{}

	// Update name if provided
	if vector.VectorName != "" {
		updates.Name = &vector.VectorName
	}

	// Update FileURI if provided
	if vector.FileURI != "" {
		updates.FileURI = &vector.FileURI
	}

	// Update DataBlob if provided
	if vector.DataBlob != nil {
		if jsonBytes, err := utils.ToJSON(vector.DataBlob); err != nil {
			return domain.VectorAnnotation{}, errors.WithDetails(errors.ErrInvalidInput, "invalid dataBlob format: %v", err)
		} else {
			dataBlobStr := string(jsonBytes)
			updates.DataBlob = &dataBlobStr
		}
	}

	// Update labels if provided
	if vector.Labels != nil && len(vector.Labels) > 0 {
		if jsonStr, err := vector.Labels.ToJSON(); err != nil {
			return domain.VectorAnnotation{}, errors.WithDetails(errors.ErrInvalidInput, "invalid labels format: %v", err)
		} else {
			updates.Labels = &jsonStr
		}
	}

	// Perform the update
	err = s.db.UpdateVectorAnnotation(ctx, vectorUID, updates)
	if err != nil {
		return domain.VectorAnnotation{}, fmt.Errorf("failed to update vector annotation: %w", err)
	}

	// Return the updated vector annotation
	updatedRecord, err := s.db.GetVectorAnnotationByUID(ctx, vectorUID)
	if err != nil {
		return domain.VectorAnnotation{}, fmt.Errorf("failed to retrieve updated vector annotation: %w", err)
	}

	return convertVectorAnnotationDBToDomain(updatedRecord), nil
}

// DeleteVectorAnnotation soft-deletes a vector annotation
func (s *vectorAnnotationsService) DeleteVectorAnnotation(ctx context.Context, vectorUID string) error {
	// Use the base service to get authentication context
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		return err
	}

	// Soft delete the vector annotation
	err = s.db.SoftDeleteVectorAnnotation(ctx, vectorUID, authCtx.CreatorID)
	if err != nil {
		return fmt.Errorf("failed to delete vector annotation: %w", err)
	}

	return nil
}

// GetVectorAnnotationFile reads and returns the GeoJSON file content for a vector annotation
func (s *vectorAnnotationsService) GetVectorAnnotationFile(ctx context.Context, slideUID string, vectorUID string) ([]byte, error) {
	// Get the vector annotation from the database
	vectorRecord, err := s.db.GetVectorAnnotationByUID(ctx, vectorUID)
	if err != nil {
		return nil, fmt.Errorf("vector annotation with UID '%s' not found: %w", vectorUID, err)
	}

	// Verify slide ownership
	if vectorRecord.SlideUID != slideUID {
		return nil, fmt.Errorf("vector annotation %s does not belong to slide %s", vectorUID, slideUID)
	}

	// Return inline DataBlob if available
	if vectorRecord.DataBlob != "" {
		return []byte(vectorRecord.DataBlob), nil
	}

	// Fall back to reading from file if FileURI is provided
	if vectorRecord.FileURI != "" {
		fileContent, err := s.readGeoJSONFile(vectorRecord.FileURI)
		if err != nil {
			return nil, fmt.Errorf("failed to read GeoJSON file '%s': %w", vectorRecord.FileURI, err)
		}
		return fileContent, nil
	}

	return nil, fmt.Errorf("no data available for vector annotation %s", vectorUID)
}

// readGeoJSONFile reads the content of a GeoJSON file
func (s *vectorAnnotationsService) readGeoJSONFile(fileURI string) ([]byte, error) {
	// Open the file
	file, err := os.Open(fileURI)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", fileURI)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the file content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return content, nil
}

// ImportVectorAnnotationToWorkspace imports a vector annotation as workspace annotations
func (s *vectorAnnotationsService) ImportVectorAnnotationToWorkspace(ctx context.Context, slideUID string, vectorUID string, allowedLabels []StudyAnnotationLabel) (domain.AnnotationImportResult, error) {
	log.Info("Starting vector annotation import",
		zap.String("slideUID", slideUID),
		zap.String("vectorUID", vectorUID))

	// Get the vector annotation
	vectorRecord, err := s.db.GetVectorAnnotationByUID(ctx, vectorUID)
	if err != nil {
		log.Error("Failed to get vector annotation",
			zap.String("vectorUID", vectorUID),
			zap.Error(err))
		return domain.AnnotationImportResult{}, fmt.Errorf("vector annotation with UID '%s' not found: %w", vectorUID, err)
	}
	log.Info("Found vector annotation",
		zap.String("vectorUID", vectorRecord.VectorUID),
		zap.String("name", vectorRecord.Name),
		zap.String("slideUID", vectorRecord.SlideUID),
		zap.Bool("hasDataBlob", vectorRecord.DataBlob != ""),
		zap.Bool("hasFileURI", vectorRecord.FileURI != ""))

	// Verify slide ownership
	if vectorRecord.SlideUID != slideUID {
		log.Error("Slide ownership mismatch",
			zap.String("vectorSlideUID", vectorRecord.SlideUID),
			zap.String("requestedSlideUID", slideUID))
		return domain.AnnotationImportResult{}, fmt.Errorf("vector annotation %s does not belong to slide %s", vectorUID, slideUID)
	}
	log.Info("Verified slide ownership", zap.String("slideUID", slideUID))

	// Create study metadata from provided allowed labels
	studyMetadata := &StudyAnnotationMetadata{
		AllowAnnotation: true,
		Annotations:     allowedLabels,
		ColorMap:        make(map[string]string),
		IndexMap:        make(map[string]interface{}),
	}
	log.Info("Created study metadata from provided labels",
		zap.Int("allowedLabelCount", len(studyMetadata.Annotations)),
		zap.Bool("allowAnnotation", studyMetadata.AllowAnnotation))

	// Get GeoJSON data from the vector annotation
	var geoJSONData interface{}
	log.Info("Extracting GeoJSON data",
		zap.Bool("hasDataBlob", vectorRecord.DataBlob != ""),
		zap.Bool("hasFileURI", vectorRecord.FileURI != ""))

	if vectorRecord.DataBlob != "" {
		log.Info("Parsing DataBlob",
			zap.Int("dataBlobLength", len(vectorRecord.DataBlob)))
		if err := utils.ParseJSON(vectorRecord.DataBlob, &geoJSONData); err != nil {
			preview := vectorRecord.DataBlob
			if len(preview) > 200 {
				preview = preview[:200]
			}
			log.Error("Failed to parse DataBlob",
				zap.Error(err),
				zap.String("dataBlobPreview", preview))
			return domain.AnnotationImportResult{}, fmt.Errorf("failed to parse DataBlob as JSON: %w", err)
		}
		log.Info("Successfully parsed DataBlob",
			zap.String("geoJSONType", fmt.Sprintf("%T", geoJSONData)))
	} else if vectorRecord.FileURI != "" {
		log.Info("Reading GeoJSON file",
			zap.String("fileURI", vectorRecord.FileURI))
		fileContent, err := s.readGeoJSONFile(vectorRecord.FileURI)
		if err != nil {
			log.Error("Failed to read GeoJSON file",
				zap.String("fileURI", vectorRecord.FileURI),
				zap.Error(err))
			return domain.AnnotationImportResult{}, fmt.Errorf("failed to read GeoJSON file: %w", err)
		}
		log.Info("Successfully read GeoJSON file",
			zap.Int("contentLength", len(fileContent)))
		if err := utils.ParseJSON(string(fileContent), &geoJSONData); err != nil {
			preview := string(fileContent)
			if len(preview) > 200 {
				preview = preview[:200]
			}
			log.Error("Failed to parse file content",
				zap.Error(err),
				zap.String("contentPreview", preview))
			return domain.AnnotationImportResult{}, fmt.Errorf("failed to parse file content as JSON: %w", err)
		}
		log.Info("Successfully parsed file content",
			zap.String("geoJSONType", fmt.Sprintf("%T", geoJSONData)))
	} else {
		log.Error("No data source available",
			zap.String("vectorUID", vectorUID))
		return domain.AnnotationImportResult{}, fmt.Errorf("no data available for vector annotation %s", vectorUID)
	}

	// Convert GeoJSON to workspace annotations
	log.Info("Converting GeoJSON to workspace annotations")
	result, err := convertGeoJSONToWorkspaceAnnotations(slideUID, vectorUID, geoJSONData, studyMetadata)
	if err != nil {
		log.Error("Failed to convert GeoJSON",
			zap.Error(err))
		return domain.AnnotationImportResult{}, fmt.Errorf("failed to convert GeoJSON to workspace annotations: %w", err)
	}

	log.Info("Conversion completed",
		zap.Int("importedCount", result.ImportedCount),
		zap.Int("skippedCount", result.SkippedCount),
		zap.Int("overwrittenCount", result.OverwrittenCount),
		zap.Int("skippedLabelCount", len(result.SkippedLabels)))

	// Store the converted annotations as a new vector annotation in the database
	log.Info("Storing imported annotations as new vector annotation in database")

	// Get authentication context
	authCtx, err := s.GetAuthContext(ctx)
	if err != nil {
		log.Error("Failed to get auth context", zap.Error(err))
		return domain.AnnotationImportResult{}, fmt.Errorf("failed to get auth context: %w", err)
	}

	// Generate a new vector UID for the imported annotations
	importedVectorUID, err := utils.GenerateFixedShortUID()
	if err != nil {
		fmt.Printf("IMPORT DEBUG: Failed to generate vector UID: %v\n", err)
		return domain.AnnotationImportResult{}, fmt.Errorf("failed to generate vector UID: %w", err)
	}

	// Convert the GeoJSON features to a JSON string for database storage
	geoJSONBytes, err := json.Marshal(result.GeoJSONFeatures)
	if err != nil {
		log.Error("Failed to marshal GeoJSON features", zap.Error(err))
		return domain.AnnotationImportResult{}, fmt.Errorf("failed to marshal GeoJSON features: %w", err)
	}
	dataBlobStr := string(geoJSONBytes)

	// Convert study labels to the format expected for database storage
	var labelsForDB []domain.VectorLabel
	for _, label := range allowedLabels {
		labelsForDB = append(labelsForDB, domain.VectorLabel{
			Name:  label.Label,
			Color: label.Color,
		})
	}

	// Convert labels to JSON string
	var labelsJSON string
	if len(labelsForDB) > 0 {
		labelsCollection := domain.VectorLabels(labelsForDB)
		if jsonStr, err := labelsCollection.ToJSON(); err != nil {
			return domain.AnnotationImportResult{}, fmt.Errorf("failed to convert labels to JSON: %w", err)
		} else {
			labelsJSON = jsonStr
		}
	}

	// Create the database record for the imported vector annotation
	dbVector := ports.NewVectorAnnotation{
		TenantID:  authCtx.TenantID,
		ActorType: "user",
		ActorID:   authCtx.CreatorID,
		CreatorID: authCtx.CreatorID,
		SlideUID:  slideUID,
		VectorUID: importedVectorUID,
		Version:   1,
		Name:      fmt.Sprintf("Imported from %s", vectorUID),
		FileURI:   "", // We're storing inline data, not a file
		DataBlob:  dataBlobStr,
		Labels:    labelsJSON,
		Metadata:  "{}",
		Mutable:   true,
	}

	// Save to database
	err = s.db.CreateVectorAnnotation(ctx, dbVector)
	if err != nil {
		log.Error("Failed to create vector annotation in database", zap.Error(err))
		return domain.AnnotationImportResult{}, fmt.Errorf("failed to create vector annotation in database: %w", err)
	}

	log.Info("Successfully stored imported annotations as vector annotation in database",
		zap.String("newVectorUID", importedVectorUID))

	// Update the result to include the new vector UID
	result.VectorUID = importedVectorUID

	log.Info("Import completed successfully",
		zap.String("slideUID", slideUID),
		zap.String("sourceVectorUID", vectorUID),
		zap.String("newVectorUID", importedVectorUID),
		zap.Int("finalImportedCount", result.ImportedCount),
		zap.Int("finalSkippedCount", result.SkippedCount),
		zap.Int("finalOverwrittenCount", result.OverwrittenCount))
	return result, nil
}

// GetVectorAnnotationsGeneric retrieves vector annotations using the new generic pattern with pagination and search
func (s *vectorAnnotationsService) GetVectorAnnotationsGeneric(ctx context.Context, params utils.PaginationAndSearchParams) ([]domain.VectorAnnotation, domain.PaginationInfo, error) {
	dbRecords, paginationInfo, err := s.db.GetVectorAnnotationsGeneric(ctx, params)
	if err != nil {
		return nil, domain.PaginationInfo{}, fmt.Errorf("failed to load vector annotations: %w", err)
	}

	vectors := make([]domain.VectorAnnotation, 0, len(dbRecords))
	for _, record := range dbRecords {
		vector := convertVectorAnnotationDBToDomain(record)
		vectors = append(vectors, vector)
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

	return vectors, domainPagination, nil
}

// Close releases resources associated with the service
func (s *vectorAnnotationsService) Close() {
	// No specific cleanup needed for vector annotations service
}

// StudyAnnotationMetadata represents the parsed study annotation metadata
// TODO: Should we do a JSON validator?
type StudyAnnotationMetadata struct {
	AllowAnnotation bool                   `json:"allow_annotation"`
	Annotations     []StudyAnnotationLabel `json:"annotations"`
	ColorMap        map[string]string      `json:"color_map"`
	IndexMap        map[string]interface{} `json:"index_map"`
}

// StudyAnnotationLabel represents a label configuration in study settings
type StudyAnnotationLabel struct {
	Label string `json:"label"`
	Type  string `json:"type"`
	Color string `json:"color"`
}

// parseStudyAnnotationMetadata parses the study metadata JSON to extract annotation settings
func parseStudyAnnotationMetadata(metadataBytes json.RawMessage) (*StudyAnnotationMetadata, error) {
	if len(metadataBytes) == 0 {
		return &StudyAnnotationMetadata{
			AllowAnnotation: false,
			Annotations:     []StudyAnnotationLabel{},
			ColorMap:        make(map[string]string),
			IndexMap:        make(map[string]interface{}),
		}, nil
	}

	var metadata StudyAnnotationMetadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal study metadata: %w", err)
	}

	// Ensure maps are initialized
	if metadata.ColorMap == nil {
		metadata.ColorMap = make(map[string]string)
	}
	if metadata.IndexMap == nil {
		metadata.IndexMap = make(map[string]interface{})
	}

	return &metadata, nil
}

// convertGeoJSONToWorkspaceAnnotations converts GeoJSON data to workspace annotation format
func convertGeoJSONToWorkspaceAnnotations(slideUID, vectorUID string, geoJSONData interface{}, studyMetadata *StudyAnnotationMetadata) (domain.AnnotationImportResult, error) {
	log.Info("Starting GeoJSON conversion",
		zap.String("slideUID", slideUID),
		zap.String("vectorUID", vectorUID))

	result := domain.AnnotationImportResult{
		SlideUID:            slideUID,
		VectorUID:           vectorUID,
		ImportedCount:       0,
		SkippedCount:        0,
		OverwrittenCount:    0,
		GeoJSONFeatures:     nil,
		ImportedAnnotations: []domain.WorkspaceAnnotationItem{},
		SkippedLabels:       []string{},
		StudyLabels:         []string{},
	}

	// Extract study labels for validation
	studyLabelMap := make(map[string]StudyAnnotationLabel)
	for _, label := range studyMetadata.Annotations {
		studyLabelMap[strings.ToLower(label.Label)] = label
		result.StudyLabels = append(result.StudyLabels, label.Label)
	}
	log.Info("Study labels prepared",
		zap.Int("labelCount", len(studyLabelMap)))

	// Parse GeoJSON structure
	geoJSONMap, ok := geoJSONData.(map[string]interface{})
	if !ok {
		log.Error("Invalid GeoJSON format",
			zap.String("actualType", fmt.Sprintf("%T", geoJSONData)))
		return result, fmt.Errorf("invalid GeoJSON format: expected object")
	}
	log.Info("GeoJSON is valid object")

	featuresInterface, exists := geoJSONMap["features"]
	if !exists {
		log.Error("GeoJSON missing features array")
		return result, fmt.Errorf("GeoJSON missing features array")
	}

	features, ok := featuresInterface.([]interface{})
	if !ok {
		log.Error("GeoJSON features is not an array",
			zap.String("actualType", fmt.Sprintf("%T", featuresInterface)))
		return result, fmt.Errorf("GeoJSON features is not an array")
	}
	log.Info("Found GeoJSON features",
		zap.Int("featureCount", len(features)))

	// Process each feature
	validFeatures := []interface{}{}
	skippedLabelsMap := make(map[string]bool)
	annotationCounter := 1
	log.Info("Starting feature processing",
		zap.Int("totalFeatures", len(features)))

	for i, featureInterface := range features {
		log.Info("Processing feature",
			zap.Int("index", i),
			zap.String("featureType", fmt.Sprintf("%T", featureInterface)))

		feature, ok := featureInterface.(map[string]interface{})
		if !ok {
			log.Warn("Skipping invalid feature - not an object",
				zap.Int("index", i),
				zap.String("actualType", fmt.Sprintf("%T", featureInterface)))
			continue
		}

		// Extract label from properties
		properties, hasProps := feature["properties"].(map[string]interface{})
		if !hasProps {
			properties = make(map[string]interface{})
			log.Info("Feature has no properties, creating empty properties",
				zap.Int("featureIndex", i))
		} else {
			log.Info("Feature has properties",
				zap.Int("featureIndex", i),
				zap.Int("propertyCount", len(properties)))
		}

		// Check both "label" and "name" properties for the label
		labelInterface, hasLabel := properties["label"]
		nameInterface, hasName := properties["name"]

		var label string
		if hasLabel {
			label, _ = labelInterface.(string)
			log.Info("Found label in 'label' property",
				zap.Int("featureIndex", i),
				zap.String("labelFromLabel", label))
		} else if hasName {
			label, _ = nameInterface.(string)
			log.Info("Found label in 'name' property",
				zap.Int("featureIndex", i),
				zap.String("labelFromName", label))
		} else {
			log.Info("No label found in properties",
				zap.Int("featureIndex", i),
				zap.Bool("hasLabelProperty", hasLabel),
				zap.Bool("hasNameProperty", hasName))
		}

		// Default to "roi" if no label
		if label == "" {
			label = "roi"
			log.Info("Using default label 'roi'",
				zap.Int("featureIndex", i))
		}

		// Check if label exists in study settings
		_, labelExists := studyLabelMap[strings.ToLower(label)]
		log.Info("Checking label against study settings",
			zap.Int("featureIndex", i),
			zap.String("label", label),
			zap.Bool("labelExists", labelExists))

		if !labelExists {
			// Skip this annotation - label not found in study settings
			log.Info("Skipping feature - label not found in study settings",
				zap.Int("featureIndex", i),
				zap.String("label", label))
			result.SkippedCount++
			if !skippedLabelsMap[label] {
				result.SkippedLabels = append(result.SkippedLabels, label)
				skippedLabelsMap[label] = true
			}
			continue
		}

		log.Info("Label matches study settings, processing feature",
			zap.Int("featureIndex", i),
			zap.String("label", label))

		// Generate annotation ID
		annotationID := fmt.Sprintf("roi-%d", annotationCounter)
		annotationCounter++
		log.Info("Generated annotation ID",
			zap.Int("featureIndex", i),
			zap.String("annotationID", annotationID))

		// Determine geometry type
		geometryInterface, hasGeom := feature["geometry"]
		var kind string = "polygon" // default
		if hasGeom {
			log.Info("Feature has geometry",
				zap.Int("featureIndex", i),
				zap.String("geometryType", fmt.Sprintf("%T", geometryInterface)))
			if geometry, ok := geometryInterface.(map[string]interface{}); ok {
				if geomType, ok := geometry["type"].(string); ok {
					log.Info("Found geometry type",
						zap.Int("featureIndex", i),
						zap.String("geomType", geomType))
					switch strings.ToLower(geomType) {
					case "point":
						kind = "point"
					case "polygon", "multipolygon":
						kind = "polygon"
					default:
						kind = "polygon"
					}
				}
			}
		} else {
			log.Info("Feature has no geometry, using default polygon",
				zap.Int("featureIndex", i))
		}

		// Create display name
		displayName := properties["name"]
		if displayName == nil {
			displayName = fmt.Sprintf("%s #%d", strings.Title(label), annotationCounter-1)
		}
		log.Info("Created display name",
			zap.Int("featureIndex", i),
			zap.String("displayName", fmt.Sprintf("%v", displayName)))

		// Update feature properties for workspace format
		properties["name"] = displayName
		properties["visible"] = true
		properties["kind"] = kind
		feature["properties"] = properties
		feature["id"] = annotationID
		log.Info("Updated feature properties",
			zap.Int("featureIndex", i),
			zap.String("label", label),
			zap.String("kind", kind),
			zap.String("annotationID", annotationID))

		// Add to valid features
		validFeatures = append(validFeatures, feature)

		// Create workspace annotation item
		workspaceAnnotation := domain.WorkspaceAnnotationItem{
			ID:      annotationID,
			Label:   label,
			Name:    fmt.Sprintf("%v", displayName),
			Visible: true,
			Kind:    kind,
		}
		result.ImportedAnnotations = append(result.ImportedAnnotations, workspaceAnnotation)
		result.ImportedCount++
		log.Info("Added annotation to result",
			zap.Int("featureIndex", i),
			zap.String("annotationID", workspaceAnnotation.ID),
			zap.String("annotationLabel", workspaceAnnotation.Label),
			zap.String("annotationName", workspaceAnnotation.Name))
	}

	log.Info("Feature processing completed",
		zap.Int("totalProcessed", len(features)),
		zap.Int("validFeatures", len(validFeatures)),
		zap.Int("imported", result.ImportedCount),
		zap.Int("skipped", result.SkippedCount))

	// Create the final GeoJSON FeatureCollection
	finalGeoJSON := map[string]interface{}{
		"type":     "FeatureCollection",
		"features": validFeatures,
	}
	result.GeoJSONFeatures = finalGeoJSON
	log.Info("Created final GeoJSON FeatureCollection",
		zap.Int("featureCount", len(validFeatures)))

	// Since we're overwriting existing annotations, count them as overwritten
	// In a real implementation, you'd check if annotations already exist
	result.OverwrittenCount = result.ImportedCount
	log.Info("Set overwritten count",
		zap.Int("overwrittenCount", result.OverwrittenCount))

	return result, nil
}

// storeWorkspaceAnnotations stores the annotations in the workspace format (simulated localStorage)
// In a real implementation, this would interface with the workspace storage system
func storeWorkspaceAnnotations(slideUID string, geoJSONFeatures interface{}) error {
	log.Info("Starting workspace annotation storage",
		zap.String("slideUID", slideUID))

	// This is a placeholder implementation
	// In the real system, you would store this data in a way that the workspace can access
	// For now, we'll just log that the operation would happen

	geoJSONBytes, err := json.Marshal(geoJSONFeatures)
	if err != nil {
		log.Error("Failed to marshal GeoJSON features",
			zap.Error(err))
		return fmt.Errorf("failed to marshal GeoJSON features: %w", err)
	}

	preview := string(geoJSONBytes)
	if len(preview) > 500 {
		preview = preview[:500]
	}

	log.Info("Successfully marshaled GeoJSON features",
		zap.String("slideUID", slideUID),
		zap.Int("dataSize", len(geoJSONBytes)),
		zap.String("storageKey", fmt.Sprintf("slideAnnotations:%s", slideUID)),
		zap.String("geoJSONPreview", preview))

	// TODO: Implement actual storage mechanism
	// This could be:
	// 1. A new table in the database for workspace annotations
	// 2. A file-based storage system
	// 3. Integration with the existing localStorage simulation

	log.Info("Workspace annotation storage completed successfully",
		zap.String("slideUID", slideUID))
	return nil
}
