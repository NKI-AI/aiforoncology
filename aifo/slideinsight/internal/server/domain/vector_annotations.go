// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package domain

// VectorAnnotation represents a vector annotation for a slide
type VectorAnnotation struct {
	VectorUID  string       `json:"vectorUid"`                                        // Unique identifier for the vector annotation
	VectorName string       `json:"vectorName"`                                       // Display name for the vector annotation
	SlideUID   string       `json:"slideUid,omitempty" validate:"required,slide_uid"` // Direct reference to slide UID
	FileURI    string       `json:"fileUri,omitempty"`                                // URI to the GeoJSON file (optional if DataBlob is provided)
	DataBlob   interface{}  `json:"dataBlob,omitempty"`                               // Inline data (JSON, GeoJSON, etc.) as alternative to FileURI
	Labels     VectorLabels `json:"labels,omitempty"`                                 // Structured label data with colors (no indices)
	ActorType  string       `json:"actorType,omitempty"`                              // Actor type (user, model)
	ActorID    int          `json:"actorId,omitempty"`                                // Actor ID
	Mutable    bool         `json:"mutable"`                                          // Whether the annotation can be modified
	DeletedAt  *string      `json:"deletedAt,omitempty"`                              // Soft deletion timestamp
	DeletedBy  *int         `json:"deletedBy,omitempty"`                              // User who deleted the vector annotation
	CreatedAt  string       `json:"createdAt,omitempty"`                              // Creation timestamp
	UpdatedAt  string       `json:"updatedAt,omitempty"`                              // Last update timestamp
}

// VectorAnnotationList represents a list of vector annotations for a slide
type VectorAnnotationList struct {
	SlideUID    string             `json:"slideUid"`
	Annotations []VectorAnnotation `json:"annotations"`
}

// VectorAnnotationsResponse represents a paginated list of vector annotations
type VectorAnnotationsResponse struct {
	Annotations []VectorAnnotation `json:"annotations"`
	Pagination  PaginationInfo     `json:"pagination"`
}

// WorkspaceAnnotationItem represents an annotation item in the workspace format
type WorkspaceAnnotationItem struct {
	ID      string `json:"id"`      // Unique identifier for the annotation
	Label   string `json:"label"`   // Label name (e.g., "tumor", "stroma")
	Name    string `json:"name"`    // Display name (e.g., "Tumor #1")
	Visible bool   `json:"visible"` // Whether the annotation is visible
	Kind    string `json:"kind"`    // Geometry type: "point", "box", "polygon"
}

// AnnotationImportResult represents the result of importing a vector annotation
type AnnotationImportResult struct {
	SlideUID            string                    `json:"slideUid"`            // Slide UID
	VectorUID           string                    `json:"vectorUid"`           // Source vector annotation UID
	ImportedCount       int                       `json:"importedCount"`       // Number of annotations successfully imported
	SkippedCount        int                       `json:"skippedCount"`        // Number of annotations skipped (label not found in study)
	OverwrittenCount    int                       `json:"overwrittenCount"`    // Number of existing annotations overwritten
	GeoJSONFeatures     interface{}               `json:"geoJsonFeatures"`     // The GeoJSON FeatureCollection that was imported
	ImportedAnnotations []WorkspaceAnnotationItem `json:"importedAnnotations"` // List of imported annotations in workspace format
	SkippedLabels       []string                  `json:"skippedLabels"`       // Labels that were skipped (not found in study settings)
	StudyLabels         []string                  `json:"studyLabels"`         // Available labels in the study settings
}
