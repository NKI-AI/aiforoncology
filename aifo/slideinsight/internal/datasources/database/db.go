// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package database

import (
	"context"
	"strings"

	"aifo.dev/aifo/slideinsight/internal/datasources/database/sqlite"
	"aifo.dev/aifo/slideinsight/internal/server/errors"
	"aifo.dev/aifo/slideinsight/internal/server/ports"
	"github.com/gofiber/fiber/v2/log"
)

// NewDatabase creates a new Database instance
func NewDatabase(ctx context.Context, databaseURL string) (ports.Database, error) {
	// Handle SQLite database
	if strings.HasPrefix(databaseURL, "sqlite://") || strings.HasSuffix(databaseURL, ".db") || strings.HasSuffix(databaseURL, ".sqlite") {
		db, err := sqlite.NewDB(databaseURL)
		if err != nil {
			return nil, errors.WithDetails(errors.ErrDatabaseConnect, "SQLite initialization failed: %v", err)
		}
		log.Info("Using SQLite database", "databaseURL", databaseURL)
		return db, nil
	}

	return nil, errors.WithDetails(errors.ErrInvalidInput, "unsupported database URL scheme: %s", databaseURL)
}
