// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package algorithms

import (
	"fmt"
	"time"
)

// parseTimestamp flexibly parses timestamps in multiple formats
func parseTimestamp(timestamp string) (time.Time, error) {
	if timestamp == "" {
		return time.Time{}, nil
	}

	// Try RFC3339 format first (2006-01-02T15:04:05Z)
	if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
		return t, nil
	}

	// Try SQLite datetime format (2006-01-02 15:04:05)
	if t, err := time.Parse("2006-01-02 15:04:05", timestamp); err == nil {
		return t, nil
	}

	// Try ISO8601 with timezone offset (2006-01-02T15:04:05Z07:00)
	if t, err := time.Parse(time.RFC3339Nano, timestamp); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp '%s'", timestamp)
}
