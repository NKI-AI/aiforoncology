// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package ports

import "context"

// Database defines the interface for interacting with the slide database.
// Using this interface allows changing the implementation without affecting the rest of the code.
type Database interface {
	SlidesRepository
	ImageTypesRepository
	SlideHistogramsRepository
	StainingProtocolsRepository
	RasterAnnotationsRepository
	VectorAnnotationsRepository
	RegionsRepository
	UsersRepository
	TenantsRepository
	StudiesRepository
	CasesRepository
	RBACRepository
	NotificationsRepository
	EmailTemplateRepository
	AlgorithmsRepository
	SettingsRepository

	// CloseConnections closes all open connections to the database.
	CloseConnections()

	// Permission-related helper methods for middleware
	GetCaseIDByUID(ctx context.Context, caseUID string) (int, error)
	GetStudyIDByUID(ctx context.Context, studyUID string) (int, error)
	GetSlideIDByUID(ctx context.Context, slideUID string) (int, error)
	GetCaseStudyRelations(ctx context.Context, caseID int) ([]int, error)
	GetSlideCaseRelation(ctx context.Context, slideID int) (int, error)
	GetUserIDByUID(ctx context.Context, userUID string) (int, error)

	// SQL-based permission filtering methods for efficient bulk operations
	GetAccessibleStudyIDs(ctx context.Context, userID int, permission string) ([]int, error)
	GetAccessibleCaseIDs(ctx context.Context, userID int, permission string) ([]int, error)
	GetAccessibleSlideIDs(ctx context.Context, userID int, permission string) ([]int, error)
}
