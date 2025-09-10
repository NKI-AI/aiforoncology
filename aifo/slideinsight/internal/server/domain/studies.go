// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.

package domain

import "encoding/json"

type Study struct {
	ID          int             `json:"-"`
	TenantID    int             `json:"-"`
	TenantUID   string          `json:"tenantUid,omitempty"`
	StudyUID    string          `json:"studyUid"`
	CreatorID   int             `json:"-"`
	CreatorUID  string          `json:"creatorUid,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	IsPublished bool            `json:"isPublished"`
	CaseCount   int             `json:"caseCount"`
	SlideCount  int             `json:"slideCount"`
	DeletedAt   *string         `json:"deletedAt,omitempty"`
	DeletedBy   *int            `json:"deletedBy,omitempty"`
	CreatedAt   string          `json:"createdAt"`
}

type StudyUpdates struct {
	Name        *string          `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	Metadata    *json.RawMessage `json:"metadata,omitempty"`
	IsPublished *bool            `json:"isPublished,omitempty"`
}

type StudyMetadata struct {
	StudyUID  string `json:"studyUid"`
	CaseCount int    `json:"caseCount"`
	// Future metadata fields can be added here
}

type StudiesResponse struct {
	Studies    []Study        `json:"studies"`
	Pagination PaginationInfo `json:"pagination"`
}
