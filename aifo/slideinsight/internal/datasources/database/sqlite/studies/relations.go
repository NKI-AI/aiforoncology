// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package studies

import (
	"context"
	"database/sql"

	"aifo.dev/aifo/slideinsight/internal/server/errors"
)

// RelationsService handles study relationship operations
type RelationsService struct {
	db *sql.DB
}

// NewRelationsService creates a new relations service instance
func NewRelationsService(db *sql.DB) *RelationsService {
	return &RelationsService{db: db}
}

// GetStudyIDByShortUID retrieves the internal study ID by its short UID
func (s *RelationsService) GetStudyIDByShortUID(ctx context.Context, studyUID string) (int, error) {
	var id int
	err := s.db.QueryRow("SELECT id FROM studies WHERE short_uid = ? AND deleted_at IS NULL", studyUID).Scan(&id)
	if err != nil {
		return 0, errors.NewDatabaseQueryError("study ID lookup", err)
	}
	return id, nil
}

// GetStudyCaseCounts retrieves case counts for all studies
func (s *RelationsService) GetStudyCaseCounts(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT s.short_uid, COUNT(sc.case_id) as case_count
		FROM studies s
		LEFT JOIN study_cases sc ON s.id = sc.study_id
		LEFT JOIN cases c ON sc.case_id = c.id
		WHERE s.deleted_at IS NULL AND (c.id IS NULL OR c.deleted_at IS NULL)
		GROUP BY s.id, s.short_uid`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("study case counts", err)
	}
	defer rows.Close()

	caseCounts := make(map[string]int)
	for rows.Next() {
		var shortUID string
		var count int
		if err := rows.Scan(&shortUID, &count); err != nil {
			return nil, errors.NewDatabaseScanError("case count", err)
		}
		caseCounts[shortUID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("case counts", err)
	}

	return caseCounts, nil
}

// GetStudySlideCounts retrieves slide counts for all studies
func (s *RelationsService) GetStudySlideCounts(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT s.short_uid, COUNT(sl.id) as slide_count
		FROM studies s
		LEFT JOIN study_cases sc ON s.id = sc.study_id
		LEFT JOIN cases c ON sc.case_id = c.id AND c.deleted_at IS NULL
		LEFT JOIN slides sl ON c.id = sl.case_id AND sl.deleted_at IS NULL
		WHERE s.deleted_at IS NULL
		GROUP BY s.id, s.short_uid`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, errors.NewDatabaseQueryError("study slide counts", err)
	}
	defer rows.Close()

	slideCounts := make(map[string]int)
	for rows.Next() {
		var shortUID string
		var count int
		if err := rows.Scan(&shortUID, &count); err != nil {
			return nil, errors.NewDatabaseScanError("slide count", err)
		}
		slideCounts[shortUID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseIterateRowsError("slide counts", err)
	}

	return slideCounts, nil
}

// GetStudyCaseAndSlideCounts retrieves both case and slide counts for all studies efficiently
func (s *RelationsService) GetStudyCaseAndSlideCounts(ctx context.Context) (map[string]int, map[string]int, error) {
	query := `
		SELECT 
			s.short_uid, 
			COUNT(DISTINCT c.id) as case_count,
			COUNT(sl.id) as slide_count
		FROM studies s
		LEFT JOIN study_cases sc ON s.id = sc.study_id
		LEFT JOIN cases c ON sc.case_id = c.id AND c.deleted_at IS NULL
		LEFT JOIN slides sl ON c.id = sl.case_id AND sl.deleted_at IS NULL
		WHERE s.deleted_at IS NULL
		GROUP BY s.id, s.short_uid`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, nil, errors.NewDatabaseQueryError("study case and slide counts", err)
	}
	defer rows.Close()

	caseCounts := make(map[string]int)
	slideCounts := make(map[string]int)

	for rows.Next() {
		var shortUID string
		var caseCount, slideCount int
		if err := rows.Scan(&shortUID, &caseCount, &slideCount); err != nil {
			return nil, nil, errors.NewDatabaseScanError("case and slide counts", err)
		}
		caseCounts[shortUID] = caseCount
		slideCounts[shortUID] = slideCount
	}

	if err := rows.Err(); err != nil {
		return nil, nil, errors.NewDatabaseIterateRowsError("case and slide counts", err)
	}

	return caseCounts, slideCounts, nil
}
