// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package schema

import (
	"database/sql"
)

// SchemaInitializer defines the interface for schema components
type SchemaInitializer interface {
	CreateTables(db *sql.DB) error
	CreateTriggers(db *sql.DB) error
	CreateIndexes(db *sql.DB) error
}

// InitializeSchema initializes all schema components in the correct order
func InitializeSchema(db *sql.DB) error {
	// Enable foreign key constraints
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return err
	}

	// Initialize schemas in dependency order
	schemas := []SchemaInitializer{
		&TenantsSchema{},
		&SettingsSchema{},
		&UsersSchema{},
		&StudiesSchema{},
		&CasesSchema{},
		&ImageTypesSchema{}, // Must come before SlidesSchema
		&SlidesSchema{},
		&AnnotationsSchema{},
		&RegionsSchema{}, // Must come after SlidesSchema
		&RBACSchema{},
		&NotificationsSchema{},
		&EmailTemplatesSchema{},
		&AlgorithmsSchema{},
	}

	// Create all tables first
	for _, schema := range schemas {
		if err := schema.CreateTables(db); err != nil {
			return err
		}
	}

	// Then create all triggers
	for _, schema := range schemas {
		if err := schema.CreateTriggers(db); err != nil {
			return err
		}
	}

	// Finally create all indexes
	for _, schema := range schemas {
		if err := schema.CreateIndexes(db); err != nil {
			return err
		}
	}

	return nil
}
