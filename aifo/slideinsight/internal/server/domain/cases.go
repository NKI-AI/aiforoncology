// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

package domain

import "encoding/json"

type Case struct {
	ID                    int             `json:"-"`
	TenantID              int             `json:"-"`
	TenantUID             string          `json:"tenantUid,omitempty"`
	CaseUID               string          `json:"caseUid"`
	CreatorID             int             `json:"-"`
	CreatorUID            string          `json:"creatorUid,omitempty"`
	Name                  string          `json:"name"`
	Metadata              json.RawMessage `json:"metadata"`
	SlideCount            int             `json:"slideCount,omitempty"`            // Number of slides in this case
	AnnotationCount       int             `json:"annotationCount,omitempty"`       // Total annotations across all slides
	SlidesWithAnnotations int             `json:"slidesWithAnnotations,omitempty"` // Number of slides that have annotations
	DeletedAt             *string         `json:"deletedAt,omitempty"`
	DeletedBy             *int            `json:"deletedBy,omitempty"`
	CreatedAt             string          `json:"createdAt"`
	UpdatedAt             string          `json:"updatedAt"`
}

type CasesResponse struct {
	Cases      []Case         `json:"cases"`
	Pagination PaginationInfo `json:"pagination"`
}

// CaseNeighbor represents a simplified case structure for navigation
type CaseNeighbor struct {
	CaseUID string `json:"caseUid"`
	Name    string `json:"name"`
}

// CaseNeighborsResponse represents the response for case neighbors navigation
type CaseNeighborsResponse struct {
	Prev   *CaseNeighbor `json:"prev"`
	Next   *CaseNeighbor `json:"next"`
	Number int           `json:"number"`
	Count  int           `json:"count"`
}
